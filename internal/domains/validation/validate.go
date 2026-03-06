package validation

import (
	"errors"
	"reflect"
	"strings"
	"sync"

	playground "github.com/go-playground/validator/v10"
)

var (
	initOnce sync.Once
	v        *playground.Validate
)

func validatorInstance() *playground.Validate {
	initOnce.Do(func() {
		instance := playground.New()
		instance.RegisterTagNameFunc(func(field reflect.StructField) string {
			if label := strings.TrimSpace(field.Tag.Get("label")); label != "" {
				return label
			}
			return strings.ToLower(field.Name)
		})
		_ = instance.RegisterValidation("notblank", func(fl playground.FieldLevel) bool {
			value := fl.Field()
			if value.Kind() != reflect.String {
				return false
			}
			return strings.TrimSpace(value.String()) != ""
		})
		v = instance
	})
	return v
}

// Struct validates a struct using validation tags and returns typed field errors.
func Struct(input any) error {
	err := validatorInstance().Struct(input)
	if err == nil {
		return nil
	}
	var validationErrs playground.ValidationErrors
	if !errors.As(err, &validationErrs) {
		return err
	}
	fields := make([]FieldError, 0, len(validationErrs))
	for _, validationErr := range validationErrs {
		fields = append(fields, FieldError{
			Field: validationErr.Field(),
			Rule:  validationErr.Tag(),
		})
	}
	return &Error{Fields: fields}
}
