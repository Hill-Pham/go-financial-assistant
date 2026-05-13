package usecase

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/Hill-Pham/go-financial-assistant/internal/domain"
	"github.com/google/uuid"
)

func recurringPurchase(dayOfMonth int) domain.Purchase {
	day := dayOfMonth
	desc := "Netflix"
	return domain.Purchase{
		ID:            uuid.New(),
		Description:   &desc,
		TotalAmount:   55.0,
		Category:      domain.CategoryEntertainment,
		PaymentMethod: domain.PaymentMethodCreditCard,
		Type:          domain.PurchaseTypeRecurring,
		IsActive:      true,
		DayOfMonth:    &day,
		RawInput:      "Netflix todo month",
	}
}

func TestGenerateRecurringExpenses_SkipsWhenNotTargetDay(t *testing.T) {
	today := time.Now().UTC().Day()
	differentDay := today%28 + 1
	if differentDay == today {
		differentDay = differentDay%28 + 1
	}

	paymentSaved := false
	repo := &mockPurchaseRepo{
		findActiveRecurringFn: func(_ context.Context) ([]domain.Purchase, error) {
			return []domain.Purchase{recurringPurchase(differentDay)}, nil
		},
		savePaymentFn: func(_ context.Context, _ *domain.Payment) error {
			paymentSaved = true
			return nil
		},
	}

	uc := NewAnalyzeExpense(repo, &mockAnalyzer{}, slog.Default())
	if err := uc.GenerateRecurringExpenses(context.Background()); err != nil {
		t.Fatalf("erro inexpected: %v", err)
	}
	if paymentSaved {
		t.Error("not should gerar pagamento fora do dia configurado")
	}
}

func TestGenerateRecurringExpenses_GeneratesOnTargetDay(t *testing.T) {
	today := time.Now().UTC().Day()

	paymentSaved := false
	repo := &mockPurchaseRepo{
		findActiveRecurringFn: func(_ context.Context) ([]domain.Purchase, error) {
			return []domain.Purchase{recurringPurchase(today)}, nil
		},
		hasPaymentForMonthFn: func(_ context.Context, _ uuid.UUID, _ time.Time) (bool, error) {
			return false, nil
		},
		savePaymentFn: func(_ context.Context, _ *domain.Payment) error {
			paymentSaved = true
			return nil
		},
	}

	uc := NewAnalyzeExpense(repo, &mockAnalyzer{}, slog.Default())
	if err := uc.GenerateRecurringExpenses(context.Background()); err != nil {
		t.Fatalf("erro inexpected: %v", err)
	}
	if !paymentSaved {
		t.Error("should ter gerado o pagamento no dia configurado")
	}
}

func TestGenerateRecurringExpenses_SkipsAlreadyPaidMonth(t *testing.T) {
	today := time.Now().UTC().Day()

	paymentSaved := false
	repo := &mockPurchaseRepo{
		findActiveRecurringFn: func(_ context.Context) ([]domain.Purchase, error) {
			return []domain.Purchase{recurringPurchase(today)}, nil
		},
		hasPaymentForMonthFn: func(_ context.Context, _ uuid.UUID, _ time.Time) (bool, error) {
			return true, nil
		},
		savePaymentFn: func(_ context.Context, _ *domain.Payment) error {
			paymentSaved = true
			return nil
		},
	}

	uc := NewAnalyzeExpense(repo, &mockAnalyzer{}, slog.Default())
	if err := uc.GenerateRecurringExpenses(context.Background()); err != nil {
		t.Fatalf("erro inexpected: %v", err)
	}
	if paymentSaved {
		t.Error("not should gerar pagamento duplicado no monthmo month")
	}
}

func TestLastValidDay_NormalDay(t *testing.T) {
	got := lastValidDay(2025, time.March, 15)
	if got != 15 {
		t.Errorf("expected 15, got %d", got)
	}
}

func TestLastValidDay_DayExceedsMonth(t *testing.T) {
	got := lastValidDay(2025, time.February, 31)
	if got != 28 {
		t.Errorf("expected 28 (Ăºltimo dia de fev/2025), got %d", got)
	}
}

func TestLastValidDay_LeapYear(t *testing.T) {
	got := lastValidDay(2024, time.February, 30)
	if got != 29 {
		t.Errorf("expected 29 (Ăºltimo dia de fev/2024 leap-year), got %d", got)
	}
}

func TestLastValidDay_Day31InMonth30(t *testing.T) {
	got := lastValidDay(2025, time.April, 31)
	if got != 30 {
		t.Errorf("expected 30 (Ăºltimo dia de april), got %d", got)
	}
}


