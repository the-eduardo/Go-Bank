package util

// Set supported currencies
const (
	USD = "USD"
	EUR = "EUR"
	BRL = "BRL"
)

func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, EUR, BRL:
		return true
	}
	return false
}
