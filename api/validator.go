package api

import (
	"github.com/Oliver-Zen/simplebank/util"
	"github.com/go-playground/validator/v10"
)

// `validator.Func` is a type representing a custom validation function.
var validCurrency validator.Func = func(fieldLevel validator.FieldLevel) bool {
	if currency, ok := fieldLevel.Field().Interface().(string); ok {
		return util.IsSupportedCurrency(currency)
	}
	return false
}
