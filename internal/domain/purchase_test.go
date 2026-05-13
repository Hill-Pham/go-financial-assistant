package domain

import (
	"testing"
)

func TestNewPurchase_Valid(t *testing.T) {
	desc := "Supermercado"
	p, err := NewPurchase(150.0, &desc, CategoryFood, PaymentMethodPix, PurchaseTypeSingle, "purchase via pix")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if p.TotalAmount != 150.0 {
		t.Errorf("expected amount 150.0, got %v", p.TotalAmount)
	}
	if p.Description == nil || *p.Description != "Supermercado" {
		t.Errorf("expected description 'Supermercado', got %v", p.Description)
	}
	if !p.IsActive {
		t.Error("new purchase should be active")
	}
	if p.ID.String() == "" {
		t.Error("ID should not be empty")
	}
}

func TestNewPurchase_NilDescription(t *testing.T) {
	p, err := NewPurchase(50.0, nil, CategoryOther, PaymentMethodCash, PurchaseTypeSingle, "raw")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if p.Description != nil {
		t.Errorf("description should be nil, got %v", p.Description)
	}
}

func TestNewPurchase_InvalidAmount(t *testing.T) {
	_, err := NewPurchase(0, nil, CategoryOther, PaymentMethodCash, PurchaseTypeSingle, "raw")
	if err == nil {
		t.Fatal("expected error for amount zero")
	}
	if err != ErrInvalidAmount {
		t.Errorf("expected ErrInvalidAmount, got: %v", err)
	}

	_, err = NewPurchase(-10, nil, CategoryOther, PaymentMethodCash, PurchaseTypeSingle, "raw")
	if err == nil {
		t.Fatal("expected error for negative amount")
	}
}

func TestPurchase_Cancel(t *testing.T) {
	p, _ := NewPurchase(100.0, nil, CategoryOther, PaymentMethodPix, PurchaseTypeSingle, "raw")

	p.Cancel("cancellation test")

	if p.IsActive {
		t.Error("purchase should be inactive after cancellation")
	}
	if p.CancelledAt == nil {
		t.Error("CancelledAt should be set")
	}
	if p.CancellationReason == nil || *p.CancellationReason != "cancellation test" {
		t.Errorf("expected CancellationReason 'cancellation test', got %v", p.CancellationReason)
	}
}

func TestNewPurchase_Recurring_Types(t *testing.T) {
	p, err := NewPurchase(80.0, nil, CategoryHealth, PaymentMethodCreditCard, PurchaseTypeRecurring, "raw")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if p.Type != PurchaseTypeRecurring {
		t.Errorf("type expected RECURRING, got %s", p.Type)
	}
}
