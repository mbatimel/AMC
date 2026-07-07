package errors

import (
	"errors"
	"fmt"
	"net/http"
)

const (
	defaultErrorCode = -32603
)

type Error struct {
	ErrorText string                 `json:"message"`
	Cause     map[string]interface{} `json:"cause,omitempty"`

	statusCode int
	errorCode  int
}

func New(message string, statusCode int) *Error {
	return &Error{
		ErrorText:  message,
		statusCode: statusCode,
		errorCode:  defaultErrorCode,
	}
}

func (e *Error) Error() string {
	if len(e.Cause) == 0 {
		return e.ErrorText
	}
	return fmt.Sprintf("%s. Causes: %v", e.ErrorText, e.Cause)
}

func (e *Error) Is(target error) bool {
	var targetError *Error
	if !errors.As(target, &targetError) {
		return false
	}
	return e.ErrorText == targetError.ErrorText
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

func (e *Error) Code() int {
	return e.statusCode
}

func (e *Error) JSONRPCCode() int {
	return e.errorCode
}

var (
	ErrRoleNotFound     = New("role not found", http.StatusBadRequest)
	ErrUserRoleNotFound = New("user role not found", http.StatusBadRequest)
	ErrForbidden        = New("access denied", http.StatusForbidden)
)
