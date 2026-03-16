package validation

import (
	"college-graduation-project-backend/internal/errs"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func Validate(s any) error {
	if err := validate.Struct(s); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return errs.BadRequest(
				"Validation failed!",
				e.Error(),
			)
		}
	}
	return nil
}
