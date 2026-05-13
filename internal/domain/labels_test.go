package domain

import "testing"

func TestCategoryLabel(t *testing.T) {
	cases := []struct {
		category Category
		expected string
	}{
		{CategoryFood, "Food"},
		{CategoryTransport, "Transport"},
		{CategoryHealth, "SaĂºde"},
		{CategoryEntertainment, "Entertainment"},
		{CategoryShopping, "Shopping"},
		{CategoryMarket, "Market"},
		{CategoryOther, "Others"},
		{Category("UNKNOWN"), "Others"},
	}
	for _, c := range cases {
		got := c.category.Label()
		if got != c.expected {
			t.Errorf("Category(%q).Label() = %q, expected %q", c.category, got, c.expected)
		}
	}
}

func TestPaymentMethodLabel(t *testing.T) {
	cases := []struct {
		method   PaymentMethod
		expected string
	}{
		{PaymentMethodCash, "Cash"},
		{PaymentMethodCreditCard, "Credit Card"},
		{PaymentMethodDebitCard, "Debit Card"},
		{PaymentMethodPix, "Pix"},
		{PaymentMethodOther, "Other"},
		{PaymentMethod("UNKNOWN"), "Other"},
	}
	for _, c := range cases {
		got := c.method.Label()
		if got != c.expected {
			t.Errorf("PaymentMethod(%q).Label() = %q, expected %q", c.method, got, c.expected)
		}
	}
}


