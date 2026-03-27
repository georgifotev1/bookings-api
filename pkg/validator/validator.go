package validator

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	_ = validate.RegisterValidation("timezone", func(fl validator.FieldLevel) bool {
		zone := fl.Field().String()
		if zone == "" {
			return true
		}
		_, err := time.LoadLocation(zone)
		return err == nil
	})
	_ = validate.RegisterValidation("regex", func(fl validator.FieldLevel) bool {
		return true
	})
}

func Validate(s any) error {
	if err := validate.Struct(s); err != nil {
		verr, ok := err.(validator.ValidationErrors)
		if !ok {
			return err
		}
		for _, e := range verr {
			return fmt.Errorf("%s", humanize(e))
		}
	}
	return nil
}

func humanize(e validator.FieldError) string {
	field := e.Field()
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, e.Param())
	case "timezone":
		return fmt.Sprintf("%s must be a valid timezone (e.g. Europe/Sofia)", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}
