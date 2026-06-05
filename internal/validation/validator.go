package validation

import (
	"college-graduation-project-backend/internal/errs"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func Validate(s any) error {
	if err := validate.Struct(s); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return errs.BadRequest("Validation failed", "invalid request payload")
		}

		reasons := make([]string, 0, len(validationErrors))
		for _, e := range validationErrors {
			reasons = append(reasons, humanizeValidationError(s, e))
		}

		return errs.BadRequest("Validation failed", strings.Join(reasons, "; "))
	}
	return nil
}

func humanizeValidationError(payload any, e validator.FieldError) string {
	field := jsonFieldName(payload, e.StructField())
	if field == "" {
		field = strings.ToLower(e.Field())
	}

	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", field, e.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, e.Param())
	default:
		return fmt.Sprintf("%s is invalid (%s)", field, e.Tag())
	}
}

func jsonFieldName(payload any, structField string) string {
	t := reflect.TypeOf(payload)
	if t == nil {
		return ""
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return ""
	}

	field, ok := t.FieldByName(structField)
	if !ok {
		return ""
	}

	tag := field.Tag.Get("json")
	if tag == "" || tag == "-" {
		return strings.ToLower(structField)
	}

	name := strings.Split(tag, ",")[0]
	if name == "" {
		return strings.ToLower(structField)
	}

	return name
}
