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
	InvalidCredentialsError = func() *Error {
		return New("invalid email or password", fasthttp.StatusUnauthorized, ErrInvalidCredentials)
	}
	EmailTakenError = func() *Error { return New("email already registered", fasthttp.StatusConflict, ErrEmailTaken) }
	NotFoundError   = func() *Error { return New("not found", fasthttp.StatusNotFound, ErrNotFound) }
	ValidationError = func(field string) *Error {
		return New("validation failed", fasthttp.StatusBadRequest, ErrInvalidRequest).AddCause("field", field)
	}
)

const (
	ErrInternal           = "auth.errors.internalError"      // Внутренняя ошибка
	ErrBadRequest         = "auth.errors.badRequest"         // Плохой запрос
	ErrMethodNotAllowed   = "auth.errors.methodNotAllowed"   // Метод не поддерживается
	ErrForbidden          = "auth.errors.forbidden"          // Доступ запрещен
	ErrInvalidRequest     = "auth.errors.invalidRequest"     // Неправильный запрос
	ErrAccessDenied       = "auth.errors.accessDenied"       // Отказано в доступе
	ErrInvalidCredentials = "auth.errors.invalidCredentials" // Неверный email или пароль
	ErrEmailTaken         = "auth.errors.emailTaken"         // Email уже зарегистрирован
	ErrNotFound           = "auth.errors.notFound"           // Не найдено
)
