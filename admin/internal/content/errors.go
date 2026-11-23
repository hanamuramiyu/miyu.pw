package content

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound   = errors.New("content: work not found")
	ErrValidation = errors.New("content: validation failed")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func (e *ValidationError) Unwrap() error {
	return ErrValidation
}

func invalid(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}
