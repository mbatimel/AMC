package errors

import (
	"errors"
	"fmt"
	"net/http"
)

type Error struct {
	ErrorText  string                 `json:"message"`
	Cause      map[string]interface{} `json:"cause,omitempty"`
	statusCode int
}

func New(message string, statusCode int) *Error {
	return &Error{ErrorText: message, statusCode: statusCode}
}

func (e *Error) Error() string {
	if len(e.Cause) == 0 {
		return e.ErrorText
	}
	return fmt.Sprintf("%s. Causes: %v", e.ErrorText, e.Cause)
}

func (e *Error) Is(target error) bool {
	var targetError *Error
	return errors.As(target, &targetError) && e.ErrorText == targetError.ErrorText
}

func (e *Error) AddCause(args ...interface{}) *Error {
	result := *e
	result.Cause = make(map[string]interface{}, len(e.Cause)+len(args)/2)
	for key, value := range e.Cause {
		result.Cause[key] = value
	}
	for i := 0; i < len(args); i += 2 {
		key := fmt.Sprint(args[i])
		var value interface{} = ""
		if i+1 < len(args) {
			value = args[i+1]
		}
		result.Cause[key] = value
	}
	return &result
}

func (e *Error) GetStatusCode() int {
	return e.statusCode
}

func (e *Error) Code() int {
	return e.statusCode
}

var (
	ErrBadRequest = New("users.errors.badRequest", http.StatusBadRequest)
	ErrValidation = New("users.errors.validation", http.StatusBadRequest)
	ErrNotFound   = New("users.errors.notFound", http.StatusNotFound)
	ErrForbidden  = New("users.errors.forbidden", http.StatusForbidden)
	ErrConflict   = New("users.errors.conflict", http.StatusConflict)
	ErrInternal   = New("users.errors.internalError", http.StatusInternalServerError)
)
