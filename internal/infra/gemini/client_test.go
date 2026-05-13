package gemini

import (
	"testing"

	"google.golang.org/genai"
)

func makeResponse(text string) *genai.GenerateContentResponse {
	return &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						{Text: text},
					},
				},
			},
		},
	}
}

func TestParseResponse_Success(t *testing.T) {
	raw := `{"amount": 49.90, "description": "Lunch", "category": "FOOD", "confidence": 0.95}`
	analysis, err := parseResponse(makeResponse(raw))
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if analysis.Amount == nil || *analysis.Amount != 49.90 {
		t.Errorf("expected amount 49.90, got %v", analysis.Amount)
	}
	if analysis.Description == nil || *analysis.Description != "Lunch" {
		t.Errorf("incorrect description: %v", analysis.Description)
	}
	if analysis.Category == nil || *analysis.Category != "FOOD" {
		t.Errorf("incorrect category: %v", analysis.Category)
	}
	if analysis.Confidence != 0.95 {
		t.Errorf("expected confidence 0.95, got %v", analysis.Confidence)
	}
	if analysis.RawResponse != raw {
		t.Errorf("rawResponse was not prebeved")
	}
}

func TestParseResponse_AmountNull(t *testing.T) {
	raw := `{"amount": null, "description": "?", "category": "OTHER", "confidence": 0.1}`
	analysis, err := parseResponse(makeResponse(raw))
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if analysis.Amount != nil {
		t.Errorf("amount should be nil, got %v", *analysis.Amount)
	}
}

func TestParseResponse_NilResponse(t *testing.T) {
	_, err := parseResponse(nil)
	if err == nil {
		t.Fatal("expected error for nil response")
	}
}

func TestParseResponse_EmptyCandidates(t *testing.T) {
	_, err := parseResponse(&genai.GenerateContentResponse{Candidates: nil})
	if err == nil {
		t.Fatal("expected error for empty candidates")
	}
}

func TestParseResponse_NilContent(t *testing.T) {
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{Content: nil},
		},
	}
	_, err := parseResponse(resp)
	if err == nil {
		t.Fatal("expected error for nil content")
	}
}

func TestParseResponse_EmptyParts(t *testing.T) {
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{Content: &genai.Content{Parts: []*genai.Part{}}},
		},
	}
	_, err := parseResponse(resp)
	if err == nil {
		t.Fatal("expected error for empty parts")
	}
}

func TestParseResponse_InvalidJSON(t *testing.T) {
	_, err := parseResponse(makeResponse("not json"))
	if err == nil {
		t.Fatal("expected invalid JSON error")
	}
}

func TestToAnalysis(t *testing.T) {
	amount := 100.0
	desc := "Pharmacy"
	cat := "HEALTH"
	raw := `{"amount":100,"description":"Pharmacy","category":"HEALTH","confidence":0.8}`

	g := &geminiResponse{
		Amount:      &amount,
		Description: &desc,
		Category:    &cat,
		Confidence:  0.8,
	}

	analysis := g.toAnalysis(raw)

	if analysis.Amount != &amount {
		t.Error("amount was not prebeved")
	}
	if analysis.Description != &desc {
		t.Error("description was not prebeved")
	}
	if analysis.Category != &cat {
		t.Error("category was not prebeved")
	}
	if analysis.Confidence != 0.8 {
		t.Errorf("expected confidence 0.8, got %v", analysis.Confidence)
	}
	if analysis.RawResponse != raw {
		t.Errorf("rawResponse was not prebeved")
	}
}

func TestParseResponse_Installment(t *testing.T) {
	raw := `{
		"type": "INSTALLMENT",
		"amount": 1200.0,
		"description": "iPhone 15",
		"category": "SHOPPING",
		"payment_method": "CREDIT_CARD",
		"confidence": 0.97,
		"installments": {"total": 12, "amount_per_installment": 100.0}
	}`
	analysis, err := parseResponse(makeResponse(raw))
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if analysis.Type != "INSTALLMENT" {
		t.Errorf("expected type INSTALLMENT, got %s", analysis.Type)
	}
	if analysis.Installments == nil {
		t.Fatal("installments should not be nil")
	}
	if analysis.Installments.Total != 12 {
		t.Errorf("expected total 12, got %d", analysis.Installments.Total)
	}
	if analysis.Installments.AmountPerInstallment != 100.0 {
		t.Errorf("expected amount_per_installment 100.0, got %v", analysis.Installments.AmountPerInstallment)
	}
	if analysis.PaymentMethod == nil || *analysis.PaymentMethod != "CREDIT_CARD" {
		t.Errorf("expected payment_method CREDIT_CARD, got %v", analysis.PaymentMethod)
	}
}

func TestParseResponse_Recurring(t *testing.T) {
	raw := `{
		"type": "RECURRING",
		"amount": 55.0,
		"description": "Netflix",
		"category": "ENTERTAINMENT",
		"payment_method": "CREDIT_CARD",
		"confidence": 0.99,
		"recurring": {"day_of_month": 15}
	}`
	analysis, err := parseResponse(makeResponse(raw))
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if analysis.Type != "RECURRING" {
		t.Errorf("expected type RECURRING, got %s", analysis.Type)
	}
	if analysis.RecurringInfo == nil {
		t.Fatal("recurring_info should not be nil")
	}
	if analysis.RecurringInfo.DayOfMonth != 15 {
		t.Errorf("expected day_of_month 15, got %d", analysis.RecurringInfo.DayOfMonth)
	}
}

func TestParseResponse_CancelRecurring(t *testing.T) {
	raw := `{
		"type": "CANCEL_RECURRING",
		"confidence": 0.95,
		"cancel_recurring": {"description": "Netflix"}
	}`
	analysis, err := parseResponse(makeResponse(raw))
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if analysis.Type != "CANCEL_RECURRING" {
		t.Errorf("expected type CANCEL_RECURRING, got %s", analysis.Type)
	}
	if analysis.CancelInfo == nil {
		t.Fatal("cancel_info should not be nil")
	}
	if analysis.CancelInfo.Description != "Netflix" {
		t.Errorf("expected description Netflix, got %s", analysis.CancelInfo.Description)
	}
}

func TestToExpenseType(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"SINGLE", "SINGLE"},
		{"INSTALLMENT", "INSTALLMENT"},
		{"RECURRING", "RECURRING"},
		{"CANCEL_RECURRING", "CANCEL_RECURRING"},
		{"", "SINGLE"},
		{"UNKNOWN", "SINGLE"},
	}
	for _, c := range cases {
		got := toExpenseType(c.input)
		if string(got) != c.expected {
			t.Errorf("toExpenseType(%q) = %q, expected %q", c.input, got, c.expected)
		}
	}
}
