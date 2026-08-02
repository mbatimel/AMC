package errors

import (
	"errors"
	"fmt"
)

const (
	noneValue = "None"
	Errors    = "errors"

	defaultErrorCode = -32603 // internalError

	CauseErrDescription = "description"
)

type TrParams struct {
	TrKey  string                 `json:"trKey"`
	Params map[string]interface{} `json:"params,omitempty"`
}

type Error struct {
	ErrorText  string
	Cause      map[string]interface{}
	statusCode int
	errorCode  *int
	trKey      string

	internalError error
}

func (e *Error) Error() string {
	var cause string
	if e.Cause != nil {
		cause = fmt.Sprintf(". Causes: %v", e.Cause)
	}
	return e.ErrorText + cause
}

func Is(errOne, errTwo error) bool {
	custErrOne, custErrTwo, ok := getCustomError(errOne, errTwo)

	if ok {
		return custErrOne.ErrorText == custErrTwo.ErrorText
	} else {
		return errOne.Error() == errTwo.Error()
	}
}

func HasCause(e error, cause string) bool {
	var custErr *Error
	ok := errors.As(e, &custErr)
	if !ok {
		return false
	}

	for key, value := range custErr.Cause {
		if key == cause || value == cause {
			return true
		}
	}

	return false
}

func New(msg string, statusCode int, outerError string) *Error {
	return &Error{
		ErrorText:  msg,
		statusCode: statusCode,
		trKey:      outerError,
	}
}

func getCustomError(err1 error, err2 error) (*Error, *Error, bool) {
	var e1, e2 *Error

	var jsonErr1 *JsonRPCError
	ok := errors.As(err1, &jsonErr1)
	if ok {
		e1 = jsonErr1.Data
	} else {
		ok = errors.As(err1, &e1)
		if !ok {
			return e1, e2, false
		}
	}

	var jsonErr2 *JsonRPCError
	ok = errors.As(err2, &jsonErr2)
	if ok {
		e2 = jsonErr2.Data
	} else {
		ok = errors.As(err2, &e2)
		if !ok {
			return e1, e2, false
		}
	}

	return e1, e2, true
}

func (e *Error) SetStatusCode(code int) *Error {
	e.statusCode = code
	return e
}

func (e *Error) GetStatusCode() int {
	return e.statusCode
}

func (e *Error) SetOuterError(err error) *Error {
	e.internalError = err
	return e
}

func (e *Error) GetOuterError() error {
	return e.internalError
}

func (e *Error) GetTranslationKey() string {
	return e.trKey
}

func (e *Error) Unwrap() error {
	return e.internalError
}

func (e *Error) AddTrErrors(trError TrParams) *Error {
	if e.Cause == nil {
		e.Cause = make(map[string]interface{}, 1)
	}

	errors, ok := e.Cause[Errors].([]TrParams)
	if !ok {
		e.Cause[Errors] = []TrParams{{
			TrKey:  trError.TrKey,
			Params: trError.Params,
		}}

		return e
	}

	e.Cause[Errors] = append(errors, trError)

	return e
}

func (e *Error) AddCause(args ...string) *Error {
	if e.Cause == nil {
		e.Cause = make(map[string]interface{})
	}

	for i := 0; i < len(args); i += 2 {
		strKey := args[i]
		e.Cause[strKey] = noneValue
		if i+1 < len(args) {
			e.Cause[strKey] = args[i+1]
		}
	}

	return e
}

// Code returns the HTTP status code. The generated transport layer
// (jsonRPC/externalapi) type-asserts errors against a `Code() int` method
// to set the response status, so this must stay in sync with statusCode
// rather than the JSON-RPC errorCode.
func (e *Error) Code() int {
	return e.statusCode
}

func (e *Error) SetErrorCode(errorCode int) *Error {
	e.errorCode = &errorCode
	return e
}

type BadRequestTypeError struct {
	StatusCode int
	Body       []byte
}

func (err *BadRequestTypeError) Error() string {
	return fmt.Sprintf("status code %d, data '%s'", err.StatusCode, err.Body)
}

func HTTPStatusCode(err error) (int, bool) {
	var badRequestErr *BadRequestTypeError
	if !errors.As(err, &badRequestErr) {
		return 0, false
	}

	return badRequestErr.StatusCode, true
}

func IsHTTPClientError(err error) bool {
	statusCode, ok := HTTPStatusCode(err)
	return ok && statusCode >= 400 && statusCode < 500
}
func EqualErrorCode(err error, targetError int) bool {
	customError, _, _ := getCustomError(err, nil)
	if customError == nil {
		return false
	}

	if customError.errorCode != nil && *customError.errorCode == targetError {
		return true
	}
	return false
}
