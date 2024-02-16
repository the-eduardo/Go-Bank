package util

// All supported currencies
const (
	USD = "USD"
	EUR = "EUR"
	BRL = "BRL"
)

// IsSupportedCurrency checks if the currency is valid
func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, EUR, BRL:
		return true
	}
	return false
}
