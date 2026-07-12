package validation

import (
	"fmt"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (v *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field `%s`: %s", v.Field, v.Message)
}

type ValidationErrors []*ValidationError

func (v ValidationErrors) Error() string {
	msgs := make([]string, len(v))
	for i, e := range v {
		msgs[i] = e.Error()
	}

	return strings.Join(msgs, "; ")
}

// * ===== VALIDATORS =====

func Required(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{
			Field:   field,
			Message: "is required",
		}
	}
	return nil
}

func IsValidEmail(field, value string) error {
	if !emailRegex.MatchString(value) {
		return &ValidationError{
			Field:   field,
			Message: "is not a valid email address",
		}
	}
	return nil
}

func MinLength(field, value string, min int) error {
	if value == "" {
		return nil
	}
	if len(value) < min {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be at least %d characters long", min),
		}
	}
	return nil
}

func ValidateAddressID(id int) error {
	if id <= 0 {
		return &ValidationError{Field: "address_id", Message: "is required"}
	}

	return nil
}

// * ===== COMBINE =====

func Validate(validations ...error) error {
	var errs ValidationErrors

	for _, err := range validations {
		if err == nil {
			continue
		}
		var ve *ValidationError
		if ok := AsValidationError(err, &ve); ok {
			errs = append(errs, ve)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}

func AsValidationError(err error, target **ValidationError) bool {
	if ve, ok := err.(*ValidationError); ok {
		*target = ve
		return true
	}
	return false
}
