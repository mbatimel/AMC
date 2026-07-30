package errors

import (
	"github.com/valyala/fasthttp"
)

var (
	AccessDeniedError     = func() *Error { return New("access denied", fasthttp.StatusForbidden, ErrAccessDenied) }
	ForbiddenError        = func() *Error { return New("forbidden", fasthttp.StatusForbidden, ErrForbidden) }
	BadRequestError       = func() *Error { return New("bad request", fasthttp.StatusBadRequest, ErrBadRequest) }
	MethodNotAllowedError = func() *Error { return New("method not allowed", fasthttp.StatusBadRequest, ErrMethodNotAllowed) }
	InternalServerError   = func() *Error {
		return New("internal server error", fasthttp.StatusInternalServerError, ErrInternal)
	}
	InvalidCredentialsError = func() *Error {
		return New("invalid email or password", fasthttp.StatusUnauthorized, ErrInvalidCredentials)
	}
	NotFoundError       = func() *Error { return New("not found", fasthttp.StatusNotFound, ErrNotFound) }
	NotImplementedError = func() *Error {
		return New("not implemented", fasthttp.StatusNotImplemented, ErrNotImplemented)
	}
)

const (
	ErrInternal           = "admin.errors.internalError"      // Внутренняя ошибка
	ErrBadRequest         = "admin.errors.badRequest"         // Плохой запрос
	ErrMethodNotAllowed   = "admin.errors.methodNotAllowed"   // Метод не поддерживается
	ErrForbidden          = "admin.errors.forbidden"          // Доступ запрещен
	ErrInvalidRequest     = "admin.errors.invalidRequest"     // Неправильный запрос
	ErrAccessDenied       = "admin.errors.accessDenied"       // Отказано в доступе
	ErrNotFound           = "admin.errors.notFound"           // Не найдено
	ErrNotImplemented     = "admin.errors.notImplemented"     // Метод не реализован
	ErrInvalidCredentials = "admin.errors.invalidCredentials" // Неверный email или пароль
)
