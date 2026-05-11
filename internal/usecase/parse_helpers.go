package usecase

import (
	"strings"

	"github.com/Hill-Pham/go-financial-assistant/internal/domain"
)

func resolvePaymentMethod(aiSuggestion *string, fallback string) string {
	if aiSuggestion != nil && *aiSuggestion != "" {
		return *aiSuggestion
	}
	return fallback
}

func inferPaymentMethod(text string) string {
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "credit"), strings.Contains(lower, "credit"):
		return "CREDIT_CARD"
	case strings.Contains(lower, "debit"), strings.Contains(lower, "debit"):
		return "DEBIT_CARD"
	case strings.Contains(lower, "money"), strings.Contains(lower, "species"), strings.Contains(lower, "species"):
		return "CASH"
	default:
		return "OTHER"
	}
}

func parsePaymentMethod(s string) (domain.PaymentMethod, error) {
	switch strings.ToUpper(s) {
	case "CASH":
		return domain.PaymentMethodCash, nil
	case "CREDIT_CARD":
		return domain.PaymentMethodCreditCard, nil
	case "DEBIT_CARD":
		return domain.PaymentMethodDebitCard, nil
	case "OTHER", "":
		return domain.PaymentMethodOther, nil
	default:
		return "", domain.ErrInvalidPaymentMethod
	}
}

func extractDescription(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	return s
}

func descriptionOrFallback(d *string, fallback string) string {
	if d != nil && *d != "" {
		return *d
	}
	return fallback
}

func parseCategory(s *string) domain.Category {
	if s == nil {
		return domain.CategoryOther
	}
	switch strings.ToUpper(*s) {
	case "FOOD":
		return domain.CategoryFood
	case "TRANSPORT":
		return domain.CategoryTransport
	case "HEALTH":
		return domain.CategoryHealth
	case "ENTERTAINMENT":
		return domain.CategoryEntertainment
	case "SHOPPING":
		return domain.CategoryShopping
	case "MARKET":
		return domain.CategoryMarket
	case "SALARY":
		return domain.CategorySalary
	default:
		return domain.CategoryOther
	}
}
