package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/the-eduardo/Go-Bank/util"
)

var validCurrency validator.Func = func(fl validator.FieldLevel) bool {
	if currency, ok := fl.Field().Interface().(string); ok {
		// Check if currency is valid
		return util.IsSupportedCurrency(currency)
	}
	return false
}
