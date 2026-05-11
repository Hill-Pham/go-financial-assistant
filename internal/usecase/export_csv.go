package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"github.com/Hill-Pham/go-financial-assistant/internal/domain"
	"github.com/Hill-Pham/go-financial-assistant/internal/domain/ports"
)

type ExportSummary struct {
	TotalExpenses float64
	TotalIncome   float64
	Balance       float64
	TotalApplied  float64
	TotalRedeemed float64
	InAccount     float64
}

type CSVExporter interface {
	Execute(ctx context.Context, month time.Time) (data []byte, filename string, summary *ExportSummary, err error)
}

type ExportCSV struct {
	repo ports.PurchaseRepository
}

func NewExportCSV(repo ports.PurchaseRepository) *ExportCSV {
	return &ExportCSV{repo: repo}
}

func (uc *ExportCSV) Execute(ctx context.Context, month time.Time) ([]byte, string, *ExportSummary, error) {
	details, err := uc.repo.FindPaymentDetailsByMonth(ctx, month)
	if err != nil {
		return nil, "", nil, fmt.Errorf("error when searching for expenses: %w", err)
	}

	if len(details) == 0 {
		return nil, "", nil, nil
	}

	var buf bytes.Buffer
	buf.WriteString("\xEF\xBB\xBF")

	w := csv.NewWriter(&buf)

	if err := w.Write([]string{"Date", "Description", "Category", "Payment Method", "Type", "Installment", "Amount (VND)"}); err != nil {
		return nil, "", nil, fmt.Errorf("error writing header: %w", err)
	}

	var totalExpenses, totalIncome, totalApplied, totalRedeemed float64
	for _, d := range details {
		row := buildCSVRow(d)
		if err := w.Write(row); err != nil {
			return nil, "", nil, fmt.Errorf("error writing row: %w", err)
		}
		switch d.PurchaseKind {
		case "INCOME":
			totalIncome += d.Amount
		case "TRANSFER":
			if d.TransferDirection == "IN" {
				totalRedeemed += d.Amount
			} else {
				totalApplied += d.Amount
			}
		default:
			totalExpenses += d.Amount
		}
	}

	if err := w.Write([]string{"", "TOTAL EXPENSES", "", "", "", "", fmt.Sprintf("%.2f", totalExpenses)}); err != nil {
		return nil, "", nil, fmt.Errorf("error writing total expenses: %w", err)
	}
	if totalIncome > 0 {
		if err := w.Write([]string{"", "TOTAL INCOME", "", "", "", "", fmt.Sprintf("%.2f", totalIncome)}); err != nil {
			return nil, "", nil, fmt.Errorf("error writing total income: %w", err)
		}
		balance := totalIncome - totalExpenses
		if err := w.Write([]string{"", "BALANCE", "", "", "", "", fmt.Sprintf("%.2f", balance)}); err != nil {
			return nil, "", nil, fmt.Errorf("error writing balance: %w", err)
		}
	}
	if totalApplied > 0 {
		if err := w.Write([]string{"", "TOTAL APPLIED", "", "", "", "", fmt.Sprintf("%.2f", totalApplied)}); err != nil {
			return nil, "", nil, fmt.Errorf("error writing total applied: %w", err)
		}
	}
	if totalRedeemed > 0 {
		if err := w.Write([]string{"", "TOTAL REDEEMED", "", "", "", "", fmt.Sprintf("%.2f", totalRedeemed)}); err != nil {
			return nil, "", nil, fmt.Errorf("error writing total redeemed: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, "", nil, fmt.Errorf("error finalizing CSV: %w", err)
	}

	filename := fmt.Sprintf("expenses_%s_%d.csv",
		strings.ToLower(ptMonths[month.Month()-1]),
		month.Year(),
	)

	balance := totalIncome - totalExpenses
	summary := &ExportSummary{
		TotalExpenses: totalExpenses,
		TotalIncome:   totalIncome,
		Balance:       balance,
		TotalApplied:  totalApplied,
		TotalRedeemed: totalRedeemed,
		InAccount:     balance - (totalApplied - totalRedeemed),
	}

	return buf.Bytes(), filename, summary, nil
}

func BuildExportCaption(month time.Time, summary *ExportSummary) string {
	base := fmt.Sprintf("📊 Spreadsheet for %s %d\n", ptMonths[month.Month()-1], month.Year())
	if summary == nil {
		return base
	}
	caption := base
	caption += fmt.Sprintf("💸 Expenses: VND %.2f\n", summary.TotalExpenses)
	if summary.TotalIncome > 0 {
		caption += fmt.Sprintf("💰 Income: VND %.2f\n", summary.TotalIncome)
		caption += fmt.Sprintf("📈 Balance: VND %.2f\n", summary.Balance)
	}
	if summary.TotalApplied > 0 || summary.TotalRedeemed > 0 {
		caption += fmt.Sprintf("🏦 Applied: VND %.2f | Redeemed: VND %.2f\n", summary.TotalApplied, summary.TotalRedeemed)
		caption += fmt.Sprintf("💵 In Account: VND %.2f", summary.InAccount)
	}
	return caption
}

func buildCSVRow(d ports.PaymentDetail) []string {
	date := resolvePaymentDate(d)

	desc := ""
	if d.Description != nil {
		desc = *d.Description
	}

	installment := "-"
	if d.InstallmentNumber != nil {
		installment = fmt.Sprintf("%d", *d.InstallmentNumber)
	}

	return []string{
		date.Format("02/01/2006"),
		desc,
		domain.Category(d.Category).Label(),
		domain.PaymentMethod(d.PaymentMethod).Label(),
		purchaseKindTypeLabel(d.PurchaseKind, d.PurchaseType),
		installment,
		fmt.Sprintf("%.2f", d.Amount),
	}
}

func resolvePaymentDate(d ports.PaymentDetail) time.Time {
	if d.DueDate != nil {
		return *d.DueDate
	}
	if d.ReferenceMonth != nil {
		return *d.ReferenceMonth
	}
	return d.CreatedAt
}

func purchaseKindTypeLabel(kind, t string) string {
	switch kind {
	case "INCOME":
		if t == "RECURRING" {
			return "Recurring Income"
		}
		return "Income"
	case "TRANSFER":
		if t == "RECURRING" {
			return "Recurring Transfer"
		}
		return "Transfer"
	default:
		switch t {
		case "INSTALLMENT":
			return "Installment"
		case "RECURRING":
			return "Recurring"
		default:
			return "One-time"
		}
	}
}
