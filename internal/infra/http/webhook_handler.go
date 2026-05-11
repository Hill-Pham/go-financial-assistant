package httpserver

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Hill-Pham/go-financial-assistant/internal/domain/ports"
	"github.com/Hill-Pham/go-financial-assistant/internal/usecase"
)

const pendingImportTTL = 30 * time.Minute

type pendingImportSession struct {
	items     []usecase.PendingTransaction
	index     int
	total     int
	expiresAt time.Time
}

type webhookHandler struct {
	analyzeExpense usecase.ExpenseAnalyzer
	csvExporter    usecase.CSVExporter
	messenger      ports.Messenger
	ownerPhone     string
	logger         *slog.Logger
	allowedNumbers map[string]struct{}
	sentIDs        sync.Map
	processedIDs   sync.Map
	pendingImports sync.Map // key: ownerPhone, value: *pendingImportSession
}

func newWebhookHandler(cfg ServerConfig, analyzeExpense usecase.ExpenseAnalyzer, csvExporter usecase.CSVExporter, messenger ports.Messenger, logger *slog.Logger) *webhookHandler {
	allowed := make(map[string]struct{}, len(cfg.AllowedNumbers)+1)
	for k := range cfg.AllowedNumbers {
		allowed[k] = struct{}{}
	}
	allowed[cfg.OwnerPhone+"@s.whatsapp.net"] = struct{}{}

	return &webhookHandler{
		analyzeExpense: analyzeExpense,
		csvExporter:    csvExporter,
		messenger:      messenger,
		ownerPhone:     cfg.OwnerPhone,
		logger:         logger,
		allowedNumbers: allowed,
	}
}

func (h *webhookHandler) startCleanup(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			cutoff := now.Add(-time.Hour)
			h.processedIDs.Range(func(key, value any) bool {
				if t, ok := value.(time.Time); ok && t.Before(cutoff) {
					h.processedIDs.Delete(key)
				}
				return true
			})
			h.sentIDs.Range(func(key, value any) bool {
				if t, ok := value.(time.Time); ok && t.Before(cutoff) {
					h.sentIDs.Delete(key)
				}
				return true
			})
			h.pendingImports.Range(func(key, value any) bool {
				if s, ok := value.(*pendingImportSession); ok && s.expiresAt.Before(now) {
					h.pendingImports.Delete(key)
				}
				return true
			})
		}
	}
}

func (h *webhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 10<<20))
	if err != nil {
		h.writeError(w, "error reading body", http.StatusBadRequest)
		return
	}

	var envelope evolutionEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		h.writeError(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if envelope.Event != "" && envelope.Event != "messages.upsert" {
		h.logger.Info("ignored event", "event", envelope.Event)
		w.WriteHeader(http.StatusOK)
		return
	}

	var payload evolutionPayload
	payload.Instance = envelope.Instance
	if err := json.Unmarshal(envelope.Data, &payload.Data); err != nil {
		h.writeError(w, "invalid payload", http.StatusBadRequest)
		return
	}

	msgID := payload.Data.Key.ID
	from := payload.Data.Key.RemoteJID

	h.logger.Info("webhook received", "instance", payload.Instance, "from", maskPhone(from), "id", msgID)

	if _, isSentByBot := h.sentIDs.LoadAndDelete(msgID); isSentByBot {
		w.WriteHeader(http.StatusOK)
		return
	}

	if msgID != "" {
		if _, already := h.processedIDs.LoadOrStore(msgID, time.Now()); already {
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	if _, isAllowed := h.allowedNumbers[from]; !isAllowed {
		h.logger.Info("message ignored", "from", maskPhone(from))
		w.WriteHeader(http.StatusOK)
		return
	}

	// Documents (PDF/statement) have their own flow
	if payload.Data.Message.DocumentMessage != nil {
		h.handleDocumentImport(r.Context(), w, payload)
		return
	}

	// Check if there is a pending import awaiting confirmation
	text := extractText(payload)
	if h.tryHandlePendingConfirmation(r.Context(), w, text) {
		return
	}

	output, err := h.route(r.Context(), payload)
	if err != nil {
		h.handleError(w, err)
		h.notifyError(r.Context(), err)
		return
	}

	if output.Type == "EXPORT_CSV" {
		h.handleExportCommand(r.Context(), w, output.ExportMonthTime)
		return
	}

	reply := formatReply(output)

	if sentID, msgErr := h.messenger.SendText(r.Context(), h.ownerPhone, reply); msgErr != nil {
		h.logger.Error("error sending reply", "error", msgErr)
	} else if sentID != "" {
		h.sentIDs.Store(sentID, time.Now())
	}

	h.writeJSON(w, http.StatusCreated, output)
}

func extractText(payload evolutionPayload) string {
	if payload.Data.Message.ExtendedTextMessage != nil {
		return payload.Data.Message.ExtendedTextMessage.Text
	}
	return payload.Data.Message.Conversation
}

func (h *webhookHandler) handleDocumentImport(ctx context.Context, w http.ResponseWriter, payload evolutionPayload) {
	doc := payload.Data.Message.DocumentMessage

	var docData []byte
	var err error

	base64Data := payload.Data.Base64
	if base64Data == "" {
		h.logger.Info("base64 missing in document webhook, fetching via API")
		key := payload.Data.Key
		base64Data, err = h.messenger.FetchImageBase64(ctx, key.RemoteJID, key.FromMe, key.ID)
		if err != nil {
			h.logger.Error("failed to fetch document base64", "error", err)
			h.sendText(ctx, "❌ Could not read the document. Please try sending it again.")
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	docData, err = decodeBase64Image(base64Data)
	if err != nil {
		h.logger.Error("failed to decode document base64", "error", err)
		h.sendText(ctx, "❌ Could not process the document. Invalid format.")
		w.WriteHeader(http.StatusOK)
		return
	}

	mimeType := doc.Mimetype
	if mimeType == "" {
		mimeType = "application/pdf"
	}

	h.sendText(ctx, "⏳ Analyzing the statement, please wait...")

	result, err := h.analyzeExpense.ExecuteDocument(ctx, usecase.DocumentInput{
		Data:     docData,
		MimeType: mimeType,
		Caption:  doc.Caption,
	})
	if err != nil {
		h.logger.Error("error processing statement", "error", err)
		h.sendText(ctx, "❌ Could not process the statement. Please try again.")
		w.WriteHeader(http.StatusOK)
		return
	}

	summary := formatStatementSummary(result)
	h.sendText(ctx, summary)

	if len(result.Pending) > 0 {
		session := &pendingImportSession{
			items:     result.Pending,
			index:     0,
			total:     len(result.Pending),
			expiresAt: time.Now().Add(pendingImportTTL),
		}
		h.pendingImports.Store(h.ownerPhone, session)
		h.sendText(ctx, formatConfirmationQuestion(result.Pending[0], 1, len(result.Pending)))
	}

	w.WriteHeader(http.StatusOK)
}

func (h *webhookHandler) tryHandlePendingConfirmation(ctx context.Context, w http.ResponseWriter, text string) bool {
	val, ok := h.pendingImports.Load(h.ownerPhone)
	if !ok {
		return false
	}

	session, ok := val.(*pendingImportSession)
	if !ok || time.Now().After(session.expiresAt) {
		h.pendingImports.Delete(h.ownerPhone)
		return false
	}

	answer := strings.ToLower(strings.TrimSpace(text))
	if answer != "yes" && answer != "y" && answer != "no" && answer != "n" {
		return false
	}

	current := session.items[session.index]

	if answer == "yes" || answer == "y" {
		h.saveConfirmedTransaction(ctx, current)
	}

	session.index++

	if session.index >= session.total {
		h.pendingImports.Delete(h.ownerPhone)
		h.sendText(ctx, "✅ Import completed!")
	} else {
		session.expiresAt = time.Now().Add(pendingImportTTL)
		next := session.items[session.index]
		h.sendText(ctx, formatConfirmationQuestion(next, session.index+1, session.total))
	}

	w.WriteHeader(http.StatusOK)
	return true
}

func (h *webhookHandler) saveConfirmedTransaction(ctx context.Context, tx usecase.PendingTransaction) {
	if err := h.analyzeExpense.SavePendingTransaction(ctx, tx); err != nil {
		h.logger.Error("error saving confirmed transaction", "description", tx.Description, "error", err)
		h.sendText(ctx, "⚠️ Could not save: "+tx.Description)
	}
}

func (h *webhookHandler) notifyError(ctx context.Context, err error) {
	msg := "Could not register the expense: " + err.Error()
	if sentID, msgErr := h.messenger.SendText(ctx, h.ownerPhone, msg); msgErr != nil {
		h.logger.Error("error sending error notification", "error", msgErr)
	} else if sentID != "" {
		h.sentIDs.Store(sentID, time.Now())
	}
}

func (h *webhookHandler) sendText(ctx context.Context, msg string) {
	if sentID, err := h.messenger.SendText(ctx, h.ownerPhone, msg); err != nil {
		h.logger.Error("error sending message", "error", err)
	} else if sentID != "" {
		h.sentIDs.Store(sentID, time.Now())
	}
}

func (h *webhookHandler) route(ctx context.Context, payload evolutionPayload) (*usecase.ExpenseOutput, error) {
	msg := payload.Data.Message

	if msg.ImageMessage != nil {
		return h.handleImage(ctx, payload)
	}

	text := msg.Conversation
	if msg.ExtendedTextMessage != nil {
		text = msg.ExtendedTextMessage.Text
	}

	if text == "" {
		return nil, errUnsupportedMessage
	}

	return h.handleText(ctx, text)
}

func (h *webhookHandler) handleText(ctx context.Context, text string) (*usecase.ExpenseOutput, error) {
	return h.analyzeExpense.ExecuteText(ctx, usecase.TextInput{Text: text})
}

func (h *webhookHandler) handleImage(ctx context.Context, payload evolutionPayload) (*usecase.ExpenseOutput, error) {
	img := payload.Data.Message.ImageMessage

	var imageData []byte
	var err error

	base64Data := payload.Data.Base64
	if base64Data == "" {
		h.logger.Info("base64 missing in webhook, fetching via API")
		key := payload.Data.Key
		base64Data, err = h.messenger.FetchImageBase64(ctx, key.RemoteJID, key.FromMe, key.ID)
		if err != nil {
			h.logger.Error("failed to fetch base64 via API", "error", err)
			return nil, errInvalidImage
		}
	}

	imageData, err = decodeBase64Image(base64Data)
	if err != nil {
		h.logger.Error("failed to decode base64", "error", err)
		return nil, errInvalidImage
	}

	return h.analyzeExpense.ExecuteImage(ctx, usecase.ImageInput{
		ImageData: imageData,
		MimeType:  img.Mimetype,
		Caption:   img.Caption,
	})
}
