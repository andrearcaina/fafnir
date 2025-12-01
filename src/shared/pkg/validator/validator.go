package validator

import (
	"fafnir/shared/pkg/errors"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	validate := validator.New()

	return &Validator{
		validate: validate,
	}
}

func (v *Validator) ValidateRequest(i interface{}) error {
	if err := v.validate.Struct(i); err != nil {
		return errors.BadRequestError("Validation error. Missing required field.").
			WithDetails(formatValidationError(err))
	}

	return nil
}

func formatValidationError(err error) string {
	if _, ok := err.(*validator.InvalidValidationError); ok {
		return err.Error()
	}

	var errorMsg string
	for _, err := range err.(validator.ValidationErrors) {
		errorMsg += "Field '" + err.Field() + "' failed on the '" + err.Tag() + "' tag"
	}

	return errorMsg
}
