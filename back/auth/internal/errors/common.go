package errors

import (
	"github.com/valyala/fasthttp"
)

var (
	AccessDeniedError     = func() *Error { return New("access denied", fasthttp.StatusForbidden, ErrAccessDenied) }
	ForbiddenError        = func() *Error { return New("forbidden", fasthttp.StatusForbidden, ErrForbidden) }
	MethodNotAllowedError = func() *Error { return New("method not allowed", fasthttp.StatusBadRequest, ErrMethodNotAllowed) }
	InternalServerError   = func() *Error {
		return New("internal server error", fasthttp.StatusInternalServerError, ErrInternal)
	}
)

const (
	ErrInternal         = "monetization.errors.auth.internalError"    // Внутренняя ошибка
	ErrBadRequest       = "monetization.errors.auth.badRequest"       // Плохой запрос
	ErrMethodNotAllowed = "monetization.errors.auth.methodNotAllowed" // Метод не поддерживается
	ErrForbidden        = "monetization.errors.auth.forbidden"        // Доступ запрещен
	ErrInvalidRequest   = "monetization.errors.auth.invalidRequest"   // Неправильный запрос
	ErrAccessDenied     = "monetization.errors.auth.accessDenied"     // Отказано в доступе
)
