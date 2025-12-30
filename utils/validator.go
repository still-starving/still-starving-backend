package utils

import (
	"github.com/go-playground/validator/v10"
)

type CustomValidator struct {
	validator *validator.Validate
}

func NewValidator() *CustomValidator {
	return &CustomValidator{
		validator: validator.New(),
	}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func FormatValidationErrors(err error) map[string]string {
	errors := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()
			switch e.Tag() {
			case "required":
				errors[field] = field + " is required"
			case "email":
				errors[field] = "Invalid email format"
			case "min":
				errors[field] = field + " must be at least " + e.Param() + " characters"
			case "max":
				errors[field] = field + " must be at most " + e.Param() + " characters"
			case "oneof":
				errors[field] = field + " must be one of: " + e.Param()
			default:
				errors[field] = field + " is invalid"
			}
		}
	}

	return errors
}
