package service

import (
	"errors"
	"strings"
)

var (
	ErrNotFound           = errors.New("not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrForbidden          = errors.New("forbidden")
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return "validation failed"
	}
	parts := make([]string, len(v))
	for i, err := range v {
		if err.Field != "" {
			parts[i] = err.Field + ": " + err.Message
		} else {
			parts[i] = err.Message
		}
	}
	return strings.Join(parts, ", ")
}
