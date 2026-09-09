package shared

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

func FormatValidationErrors(err error) map[string]string {
	errors := make(map[string]string)

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		errors["request"] = "invalid request body"
		return errors
	}

	for _, fe := range validationErrors {
		field := strings.ToLower(fe.Field())
		switch fe.Tag() {
		case "required":
			errors[field] = fmt.Sprintf("%s is required", field)
		case "email":
			errors[field] = fmt.Sprintf("%s must be a valid email", field)
		case "min":
			errors[field] = fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
		case "len":
			errors[field] = fmt.Sprintf("%s must be exactly %s characters", field, fe.Param())
		default:
			errors[field] = fmt.Sprintf("%s is invalid", field)
		}
	}

	return errors
}
