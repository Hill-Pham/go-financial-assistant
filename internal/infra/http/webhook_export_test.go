package httpserver

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/Hill-Pham/go-financial-assistant/internal/usecase"
)

func exportAnalyzer(month time.Time) *mockAnalyzer {
	return &mockAnalyzer{
		executeTextFn: func(_ context.Context, _ usecase.TextInput) (*usecase.ExpenseOutput, error) {
			return &usecase.ExpenseOutput{Type: "EXPORT_CSV", ExportMonthTime: month}, nil
		},
	}
}

func TestHandleExportCommand_SendsDocument(t *testing.T) {
	documentSent := false
	monthsenger := &mockMessenger{
		sendDocumentFn: func(_ context.Context, _, _, _, _ string) (string, error) {
			documentSent = true
			return "msg-id-123", nil
		},
	}
	exporter := &mockCSVExporter{
		executeFn: func(_ context.Context, _ time.Time) ([]byte, string, *usecase.ExportSummary, error) {
			return []byte("csv data"), "expenses_march_2025.csv", nil, nil
		},
	}

	month := time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC)
	h := newHandler(exportAnalyzer(month), monthsenger, exporter)
	body := buildPayload("inst", "5511888888888@s.whatsapp.net", "MSG-EXP-1", false,
		evolutionMessage{Conversation: "export march 2025"}, "")
	rr := doRequest(h, body)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if !documentSent {
		t.Error("SendDocument was not called")
	}
}

func TestHandleExportCommand_EmptyMonth_SendsText(t *testing.T) {
	textSent := ""
	monthsenger := &mockMessenger{
		sendTextFn: func(_ context.Context, _, text string) (string, error) {
			textSent = text
			return "", nil
		},
	}
	exporter := &mockCSVExporter{
		executeFn: func(_ context.Context, _ time.Time) ([]byte, string, *usecase.ExportSummary, error) {
			return nil, "", nil, nil
		},
	}

	month := time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC)
	h := newHandler(exportAnalyzer(month), monthsenger, exporter)
	body := buildPayload("inst", "5511888888888@s.whatsapp.net", "MSG-EXP-2", false,
		evolutionMessage{Conversation: "export february 2020"}, "")
	rr := doRequest(h, body)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if textSent == "" {
		t.Error("expected text message for empty month")
	}
}

func TestHandleExportCommand_ExporterError_Returns500(t *testing.T) {
	exporter := &mockCSVExporter{
		executeFn: func(_ context.Context, _ time.Time) ([]byte, string, *usecase.ExportSummary, error) {
			return nil, "", nil, errors.New("db error")
		},
	}

	month := time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC)
	h := newHandler(exportAnalyzer(month), &mockMessenger{}, exporter)
	body := buildPayload("inst", "5511888888888@s.whatsapp.net", "MSG-EXP-3", false,
		evolutionMessage{Conversation: "export march"}, "")
	rr := doRequest(h, body)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestHandleExportCommand_UsesMonthFromAnalyzer(t *testing.T) {
	var receivedMonth time.Time
	exporter := &mockCSVExporter{
		executeFn: func(_ context.Context, m time.Time) ([]byte, string, *usecase.ExportSummary, error) {
			receivedMonth = m
			return []byte("csv"), "file.csv", nil, nil
		},
	}

	want := time.Date(2024, time.April, 1, 0, 0, 0, 0, time.UTC)
	h := newHandler(exportAnalyzer(want), &mockMessenger{}, exporter)
	body := buildPayload("inst", "5511888888888@s.whatsapp.net", "MSG-EXP-4", false,
		evolutionMessage{Conversation: "export april 2024"}, "")
	doRequest(h, body)

	if !receivedMonth.Equal(want) {
		t.Errorf("expected month %v, got %v", want, receivedMonth)
	}
}




