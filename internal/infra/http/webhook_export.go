package httpserver

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/MarcosAAlbanoJunior/go-financial-assistant/internal/usecase"
)

func (h *webhookHandler) handleExportCommand(ctx context.Context, w http.ResponseWriter, month time.Time) {
	data, filename, summary, err := h.csvExporter.Execute(ctx, month)
	if err != nil {
		h.logger.Error("error generating CSV", "error", err)
		h.sendText(ctx, "❌ I was unable to generate the spreadsheet. Please try again.")
		h.writeError(w, "error generating CSV", http.StatusInternalServerError)
		return
	}

	if data == nil {
		h.sendText(ctx, fmt.Sprintf("📊 No entries recorded in %s.", month.Format("01/2006")))
		w.WriteHeader(http.StatusOK)
		return
	}

	caption := usecase.BuildExportCaption(month, summary)
	base64Data := base64.StdEncoding.EncodeToString(data)

	if sentID, err := h.messenger.SendDocument(ctx, h.ownerPhone, filename, base64Data, caption); err != nil {
		h.logger.Error("error sending document", "error", err)
	} else if sentID != "" {
		h.sentIDs.Store(sentID, time.Now())
	}

	w.WriteHeader(http.StatusOK)
}
