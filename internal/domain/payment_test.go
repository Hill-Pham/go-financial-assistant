package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewPayment_Paid(t *testing.T) {
	purchaseID := uuid.New()
	p := NewPayment(purchaseID, 100.0, PaymentStatusPaid)

	if p.PurchaseID != purchaseID {
		t.Errorf("PurchaseID incorrect")
	}
	if p.Amount != 100.0 {
		t.Errorf("amount expected 100.0, got %v", p.Amount)
	}
	if p.Status != PaymentStatusPaid {
		t.Errorf("status expected PAID, got %s", p.Status)
	}
	if p.PaidAt == nil {
		t.Error("PaidAt should be preenchido for status PAID")
	}
	if p.ID.String() == "" {
		t.Error("ID not should be empty")
	}
}

func TestNewPayment_Pending(t *testing.T) {
	p := NewPayment(uuid.New(), 50.0, PaymentStatusPending)

	if p.Status != PaymentStatusPending {
		t.Errorf("status expected PENDING, got %s", p.Status)
	}
	if p.PaidAt != nil {
		t.Error("PaidAt should be nil for status PENDING")
	}
}

func TestNewPayment_Cancelled(t *testing.T) {
	p := NewPayment(uuid.New(), 50.0, PaymentStatusCancelled)

	if p.Status != PaymentStatusCancelled {
		t.Errorf("status expected CANCELLED, got %s", p.Status)
	}
	if p.PaidAt != nil {
		t.Error("PaidAt should be nil for status CANCELLED")
	}
}


