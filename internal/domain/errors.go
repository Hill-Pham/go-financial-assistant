package domain

import "errors"

var (
	ErrInvalidAmount        = errors.New("The value must be greater than zero.")
	ErrInvalidPaymentMethod = errors.New("Invalid payment method.")
	ErrPurchaseNotFound     = errors.New("Purchase not found.")
)
