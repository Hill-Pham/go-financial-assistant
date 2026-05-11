package httpserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/MarcosAAlbanoJunior/go-financial-assistant/internal/domain"
	"github.com/MarcosAAlbanoJunior/go-financial-assistant/internal/usecase"
)

func formatStatementSummary(output *usecase.StatementOutput) string {
	if output.Inserted == 0 && len(output.Pending) == 0 {
		return "📄 No new transactions found in the statement."
	}

	var sb strings.Builder
	sb.WriteString("📄 *Statement processed!*\n")
	if output.Inserted > 0 {
		sb.WriteString(fmt.Sprintf("✅ %d transaction(s) imported automatically.\n", output.Inserted))
	}
	if len(output.Pending) > 0 {
		sb.WriteString(fmt.Sprintf("⚠️ %d transaction(s) already exist in the database — I will ask one by one.", len(output.Pending)))
	}
	return sb.String()
}

func formatConfirmationQuestion(tx usecase.PendingTransaction, current, total int) string {
	return fmt.Sprintf(
		"❓ Transaction %d/%d\n📅 %s\n📝 %s\n💰 VND %.2f\n🏷️ %s\n\nA transaction with this amount already exists on this date. Do you want to insert it anyway?\nReply *yes* or *no*",
		current, total,
		tx.Date.Format("02/01/2006"),
		tx.Description,
		tx.Amount,
		tx.Category,
	)
}

func (h *webhookHandler) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (h *webhookHandler) writeError(w http.ResponseWriter, msg string, status int) {
	h.writeJSON(w, status, map[string]string{"error": msg})
}

func (h *webhookHandler) handleError(w http.ResponseWriter, err error) {
	h.logger.Error("error processing webhook", "error", err)

	switch {
	case errors.Is(err, domain.ErrInvalidAmount):
		h.writeError(w, err.Error(), http.StatusUnprocessableEntity)
	case errors.Is(err, domain.ErrInvalidPaymentMethod):
		h.writeError(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, errUnsupportedMessage):
		h.writeError(w, "unsupported message type", http.StatusBadRequest)
	case errors.Is(err, errInvalidImage):
		h.writeError(w, "invalid or corrupted image", http.StatusBadRequest)
	default:
		h.writeError(w, "internal error", http.StatusInternalServerError)
	}
}

func formatReply(output *usecase.ExpenseOutput) string {
	switch output.Type {
	case "QUERY":
		return formatQueryReply(output)
	case "INSTALLMENT":
		return fmt.Sprintf(
			"✅ Installment purchase registered!\n💰 Total: VND %.2f\n📅 %dx de VND %.2f\n📝 %s\n🏷️ %s\n💳 %s",
			output.Amount, output.TotalInstallments, output.InstallmentAmount,
			output.Description, output.Category, output.Payment,
		)
	case "RECURRING":
		return fmt.Sprintf(
			"✅ Recurring expense recorded!\n💰 VND %.2f/month\n📝 %s\n🏷️ %s\n💳 %s\n📅 Day %d of each month",
			output.Amount, output.Description, output.Category, output.Payment, output.DayOfMonth,
		)
	case "CANCEL_RECURRING":
		return fmt.Sprintf("✅ Recurring expense canceled!\n📝 %s", output.CancelledDescription)
	case "INCOME":
		return fmt.Sprintf("✅ Income recorded!\n💰 VND %.2f\n📝 %s\n🏷️ %s\n💳 %s",
			output.Amount, output.Description, output.Category, output.Payment)
	case "INCOME_RECURRING":
		return fmt.Sprintf("✅ Recurring income recorded!\n💰 VND %.2f/month\n📝 %s\n🏷️ %s\n💳 %s\n📅 Day %d of each month",
			output.Amount, output.Description, output.Category, output.Payment, output.DayOfMonth)
	case "TRANSFER":
		return fmt.Sprintf("↔️ Transfer recorded!\n💰 VND %.2f\n📝 %s\n💳 %s",
			output.Amount, output.Description, output.Payment)
	default:
		return fmt.Sprintf("✅ Expense recorded!\n💰 VND %.2f\n📝 %s\n🏷️ %s\n💳 %s",
			output.Amount, output.Description, output.Category, output.Payment)
	}
}

func formatQueryReply(output *usecase.ExpenseOutput) string {
	if output.QueryEmpty {
		return fmt.Sprintf("📊 No releases recorded in %s.", output.QueryMonth)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📊 Summary of %s\n\n", output.QueryMonth))

	if len(output.QueryCategories) > 0 {
		sb.WriteString(fmt.Sprintf("💸 Expenses: VND %.2f\n", output.QueryTotal))
		for _, c := range output.QueryCategories {
			sb.WriteString(fmt.Sprintf("  • %s: VND %.2f\n", c.Category, c.Total))
		}
	}

	if output.QueryIncome > 0 {
		sb.WriteString(fmt.Sprintf("\n💰 Income: VND %.2f\n", output.QueryIncome))
		sb.WriteString(fmt.Sprintf("📈 Balance: VND %.2f\n", output.QueryBalance))
	}

	if output.QueryApplied > 0 || output.QueryRedeemed > 0 {
		sb.WriteString(fmt.Sprintf("\n🏦 Investments in %s\n", output.QueryMonth))
		sb.WriteString(fmt.Sprintf("  ↓ Applied: VND %.2f\n", output.QueryApplied))
		sb.WriteString(fmt.Sprintf("  ↑ Redeemed: VND %.2f\n", output.QueryRedeemed))
		sb.WriteString(fmt.Sprintf("💵 In account: VND %.2f\n", output.QueryInAccount))
	}

	return sb.String()
}

var (
	errUnsupportedMessage = errors.New("unsupported message type")
	errInvalidImage       = errors.New("invalid image")
)
