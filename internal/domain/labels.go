package domain

func (c Category) Label() string {
	switch c {
	case CategoryFood:
		return "Food"
	case CategoryTransport:
		return "Transport"
	case CategoryHealth:
		return "Health"
	case CategoryEntertainment:
		return "Entertainment"
	case CategoryShopping:
		return "Shopping"
	case CategoryMarket:
		return "Market"
	case CategoryInvestment:
		return "Investment"
	case CategorySalary:
		return "Salary/Income"
	default:
		return "Others"
	}
}

func (p PaymentMethod) Label() string {
	switch p {
	case PaymentMethodCash:
		return "Cash"
	case PaymentMethodCreditCard:
		return "Credit Card"
	case PaymentMethodDebitCard:
		return "Debit Card"
	case PaymentMethodPix:
		return "Pix"
	default:
		return "Other"
	}
}
