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
			return fmt.Errorf("%s: %s", e.Field(), e.Tag())
		}
	}
	return nil
}
