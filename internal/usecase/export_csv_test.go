package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Hill-Pham/go-financial-assistant/internal/domain/ports"
)

func TestExportCSV_EmptyMonth(t *testing.T) {
	repo := &mockPurchaseRepo{
		findPaymentDetailsByMonthFn: func(_ context.Context, _ time.Time) ([]ports.PaymentDetail, error) {
			return nil, nil
		},
	}

	uc := NewExportCSV(repo)
	data, filename, _, err := uc.Execute(context.Background(), time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC))

	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if data != nil {
		t.Error("expected data == nil for month sem expenses")
	}
	if filename != "" {
		t.Errorf("expected filename empty, got: %q", filename)
	}
}

func TestExportCSV_RepoError(t *testing.T) {
	repo := &mockPurchaseRepo{
		findPaymentDetailsByMonthFn: func(_ context.Context, _ time.Time) ([]ports.PaymentDetail, error) {
			return nil, errSentinel
		},
	}

	uc := NewExportCSV(repo)
	_, _, _, err := uc.Execute(context.Background(), time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC))

	if err == nil {
		t.Fatal("expected error from repositĂ³rio")
	}
}

func TestExportCSV_Filename(t *testing.T) {
	repo := repoWithOneDetail(singleDetail())

	uc := NewExportCSV(repo)
	_, filename, _, err := uc.Execute(context.Background(), time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filename != "expenses_march_2025.csv" {
		t.Errorf("unexpected filename: %q", filename)
	}
}

func TestExportCSV_HasBOM(t *testing.T) {
	repo := repoWithOneDetail(singleDetail())

	uc := NewExportCSV(repo)
	data, _, _, _ := uc.Execute(context.Background(), time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC))

	bom := []byte{0xEF, 0xBB, 0xBF}
	if !bytes.HasPrefix(data, bom) {
		t.Error("CSV should start with UTF-8 BOM")
	}
}

func TestExportCSV_Header(t *testing.T) {
	repo := repoWithOneDetail(singleDetail())

	uc := NewExportCSV(repo)
	data, _, _, _ := uc.Execute(context.Background(), time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC))

	records := parseCSV(t, data)
	if len(records) < 1 {
		t.Fatal("CSV has no rows")
	}
	header := records[0]
	expected := []string{"Date", "Description", "Category", "Payment Method", "Type", "Installment", "Amount (R$)"}
	for i, col := range expected {
		if i >= len(header) || header[i] != col {
			t.Errorf("column %d: expected %q, got %q", i, col, header[i])
		}
	}
}

func TestExportCSV_SingleRow(t *testing.T) {
	due := time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)
	detail := ports.PaymentDetail{
		Description:   strPtr("Lunch"),
		Category:      "FOOD",
		PaymentMethod: "PIX",
		Amount:        45.50,
		Status:        "PAID",
		PurchaseType:  "SINGLE",
		DueDate:       &due,
	}

	repo := repoWithOneDetail(detail)
	uc := NewExportCSV(repo)
	data, _, _, _ := uc.Execute(context.Background(), time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC))

	records := parseCSV(t, data)
	if len(records) < 2 {
		t.Fatal("CSV sem linha de rows")
	}

	row := records[1]
	if row[0] != "15/03/2025" {
		t.Errorf("date: expected 15/03/2025, got %q", row[0])
	}
	if row[1] != "Lunch" {
		t.Errorf("description: expected Lunch, got %q", row[1])
	}
	if row[2] != "Food" {
		t.Errorf("category: expected Food, got %q", row[2])
	}
	if row[3] != "PIX" {
		t.Errorf("payment method: expected PIX, got %q", row[3])
	}
	if row[4] != "Single" {
		t.Errorf("type: expected Single, got %q", row[4])
	}
	if row[5] != "-" {
		t.Errorf("installment: expected -, got %q", row[5])
	}
	if row[6] != "45.50" {
		t.Errorf("amount: expected 45.50, got %q", row[6])
	}
}

func TestExportCSV_InstallmentRow(t *testing.T) {
	due := time.Date(2025, 3, 10, 0, 0, 0, 0, time.UTC)
	installNum := 2
	detail := ports.PaymentDetail{
		Description:       strPtr("Notebook"),
		Category:          "SHOPPING",
		PaymentMethod:     "CREDIT_CARD",
		Amount:            500.00,
		PurchaseType:      "INSTALLMENT",
		InstallmentNumber: &installNum,
		DueDate:           &due,
	}

	repo := repoWithOneDetail(detail)
	uc := NewExportCSV(repo)
	data, _, _, _ := uc.Execute(context.Background(), time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC))

	records := parseCSV(t, data)
	row := records[1]

	if row[4] != "Installment" {
		t.Errorf("type: expected Installment, got %q", row[4])
	}
	if row[5] != "2" {
		t.Errorf("installment: expected 2, got %q", row[5])
	}
}

func TestExportCSV_RecurringRow(t *testing.T) {
	ref := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	detail := ports.PaymentDetail{
		Description:    strPtr("Netflix"),
		Category:       "ENTERTAINMENT",
		PaymentMethod:  "CREDIT_CARD",
		Amount:         55.90,
		PurchaseType:   "RECURRING",
		ReferenceMonth: &ref,
	}

	repo := repoWithOneDetail(detail)
	uc := NewExportCSV(repo)
	data, _, _, _ := uc.Execute(context.Background(), time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC))

	records := parseCSV(t, data)
	row := records[1]

	if row[4] != "Recurring" {
		t.Errorf("type: expected Recurring, got %q", row[4])
	}
	if row[0] != "01/03/2025" {
		t.Errorf("data: expected 01/03/2025, got %q", row[0])
	}
}

func TestExportCSV_TotalRow(t *testing.T) {
	due := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	details := []ports.PaymentDetail{
		{Description: strPtr("A"), Category: "FOOD", PaymentMethod: "PIX", Amount: 30.00, PurchaseType: "SINGLE", DueDate: &due},
		{Description: strPtr("B"), Category: "FOOD", PaymentMethod: "PIX", Amount: 20.50, PurchaseType: "SINGLE", DueDate: &due},
	}
	repo := &mockPurchaseRepo{
		findPaymentDetailsByMonthFn: func(_ context.Context, _ time.Time) ([]ports.PaymentDetail, error) {
			return details, nil
		},
	}

	uc := NewExportCSV(repo)
	data, _, _, _ := uc.Execute(context.Background(), time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC))

	records := parseCSV(t, data)
	if len(records) != 4 {
		t.Fatalf("expected 4 rows, got %d", len(records))
	}

	total := records[3]
	if total[1] != "TOTAL EXPENSES" {
		t.Errorf("total label: expected TOTAL EXPENSES, got %q", total[1])
	}
	if total[6] != "50.50" {
		t.Errorf("total value: expected 50.50, got %q", total[6])
	}
}

func TestExportCSV_NilDescription(t *testing.T) {
	due := time.Date(2025, 3, 5, 0, 0, 0, 0, time.UTC)
	detail := ports.PaymentDetail{
		Description:   nil,
		Category:      "OTHER",
		PaymentMethod: "CASH",
		Amount:        10.00,
		PurchaseType:  "SINGLE",
		DueDate:       &due,
	}

	repo := repoWithOneDetail(detail)
	uc := NewExportCSV(repo)
	data, _, _, _ := uc.Execute(context.Background(), time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC))

	records := parseCSV(t, data)
	if records[1][1] != "" {
		t.Errorf("nil description should produce empty string, got %q", records[1][1])
	}
}

func TestExportCSV_TransferRows(t *testing.T) {
	due := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	details := []ports.PaymentDetail{
		{Description: strPtr("Expense"), Category: "FOOD", PaymentMethod: "PIX", Amount: 100.00, PurchaseType: "SINGLE", PurchaseKind: "EXPENSE", DueDate: &due},
		{Description: strPtr("Savings Allocation"), Category: "OTHER", PaymentMethod: "PIX", Amount: 500.00, PurchaseType: "SINGLE", PurchaseKind: "TRANSFER", TransferDirection: "OUT", DueDate: &due},
		{Description: strPtr("CDB Redemption"), Category: "OTHER", PaymentMethod: "PIX", Amount: 200.00, PurchaseType: "SINGLE", PurchaseKind: "TRANSFER", TransferDirection: "IN", DueDate: &due},
	}
	repo := &mockPurchaseRepo{
		findPaymentDetailsByMonthFn: func(_ context.Context, _ time.Time) ([]ports.PaymentDetail, error) {
			return details, nil
		},
	}

	uc := NewExportCSV(repo)
	data, _, summary, err := uc.Execute(context.Background(), time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	records := parseCSV(t, data)
	// header + 3 rows + TOTAL EXPENSES + TOTAL APPLIED + TOTAL REDEEMED = 7
	if len(records) != 7 {
		t.Fatalf("expected 7 rows, got %d: %v", len(records), records)
	}

	totalRow := records[4]
	if totalRow[1] != "TOTAL EXPENSES" || totalRow[6] != "100.00" {
		t.Errorf("TOTAL EXPENSES incorrect: %v", totalRow)
	}
	appliedRow := records[5]
	if appliedRow[1] != "TOTAL APPLIED" || appliedRow[6] != "500.00" {
		t.Errorf("TOTAL APPLIED incorrect: %v", appliedRow)
	}
	redeemedRow := records[6]
	if redeemedRow[1] != "TOTAL REDEEMED" || redeemedRow[6] != "200.00" {
		t.Errorf("TOTAL REDEEMED incorrect: %v", redeemedRow)
	}

	if summary.TotalExpenses != 100.00 {
		t.Errorf("summary.TotalExpenses: expected 100.00, got %.2f", summary.TotalExpenses)
	}
	if summary.TotalApplied != 500.00 {
		t.Errorf("summary.TotalApplied: expected 500.00, got %.2f", summary.TotalApplied)
	}
	if summary.TotalRedeemed != 200.00 {
		t.Errorf("summary.TotalRedeemed: expected 200.00, got %.2f", summary.TotalRedeemed)
	}
}

func TestBuildExportCaption_WithTransfers(t *testing.T) {
	month := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	summary := &ExportSummary{
		TotalExpenses: 1000.00,
		TotalIncome:   6000.00,
		Balance:       5000.00,
		TotalApplied:  3000.00,
		TotalRedeemed: 500.00,
		InAccount:     2500.00,
	}

	caption := BuildExportCaption(month, summary)

	if !strings.Contains(caption, "Expenses: R$ 1000.00") {
		t.Errorf("caption sem expenses: %q", caption)
	}
	if !strings.Contains(caption, "Entradas: R$ 6000.00") {
		t.Errorf("caption sem entries: %q", caption)
	}
	if !strings.Contains(caption, "Resultado: R$ 5000.00") {
		t.Errorf("caption sem resultado: %q", caption)
	}
	if !strings.Contains(caption, "Aplicado: R$ 3000.00") {
		t.Errorf("caption sem applied: %q", caption)
	}
	if !strings.Contains(caption, "Resgatado: R$ 500.00") {
		t.Errorf("caption sem redeemed: %q", caption)
	}
	if !strings.Contains(caption, "Em conta: R$ 2500.00") {
		t.Errorf("caption sem em conta: %q", caption)
	}
}

func TestBuildExportCaption_NoTransfers(t *testing.T) {
	month := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	summary := &ExportSummary{TotalExpenses: 500.00}

	caption := BuildExportCaption(month, summary)

	if strings.Contains(caption, "Aplicado") {
		t.Error("caption not should ter linha de investimentos quando not hĂ¡ transferĂªncias")
	}
	if strings.Contains(caption, "Em conta") {
		t.Error("caption not should ter 'Em conta' sem transferĂªncias")
	}
}

var errSentinel = errors.New("repo error")

func strPtr(s string) *string { return &s }

func singleDetail() ports.PaymentDetail {
	due := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	return ports.PaymentDetail{
		Description:   strPtr("Teste"),
		Category:      "FOOD",
		PaymentMethod: "PIX",
		Amount:        10.00,
		PurchaseType:  "SINGLE",
		DueDate:       &due,
	}
}

func repoWithOneDetail(d ports.PaymentDetail) *mockPurchaseRepo {
	return &mockPurchaseRepo{
		findPaymentDetailsByMonthFn: func(_ context.Context, _ time.Time) ([]ports.PaymentDetail, error) {
			return []ports.PaymentDetail{d}, nil
		},
	}
}

func parseCSV(t *testing.T, data []byte) [][]string {
	t.Helper()
	content := strings.TrimPrefix(string(data), "\xEF\xBB\xBF")
	r := csv.NewReader(strings.NewReader(content))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("error parsing CSV: %v", err)
	}
	return records
}



