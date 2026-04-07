package apperr

import (
	"errors"
	"net/http"
)

type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string { return e.Message }

func New(status int, message string) error {
	return &Error{Status: status, Message: message}
}

func Status(err error) int {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Status
	}
	return http.StatusInternalServerError
}

func Message(err error) string {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Message
	}
	return "Internal server error"
}
