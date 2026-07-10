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
	ErrCodeExpiredOrNotFound = func() *Error { return New("code not found or expired", fasthttp.StatusBadRequest, ErrInternal) }

	ErrInvalidEmailError    = func() *Error { return New("invalid email", fasthttp.StatusBadRequest, ErrInvalidRequest) }
	ErrUserNotFoundError    = func() *Error { return New("user not found", fasthttp.StatusNotFound, ErrNotFound) }
	ErrInvalidPasswordError = func() *Error { return New("invalid email or password", fasthttp.StatusUnauthorized, ErrUnauthorized) }
	ErrEmailExistsError     = func() *Error { return New("email already registered", fasthttp.StatusConflict, ErrConflict) }
	ErrInvalidCodeError     = func() *Error { return New("invalid verification code", fasthttp.StatusBadRequest, ErrInvalidRequest) }
)

const (
	ErrInternal         = "monetization.errors.auth.internalError"    // Внутренняя ошибка
	ErrBadRequest       = "monetization.errors.auth.badRequest"       // Плохой запрос
	ErrMethodNotAllowed = "monetization.errors.auth.methodNotAllowed" // Метод не поддерживается
	ErrForbidden        = "monetization.errors.auth.forbidden"        // Доступ запрещен
	ErrInvalidRequest   = "monetization.errors.auth.invalidRequest"   // Неправильный запрос
	ErrAccessDenied     = "monetization.errors.auth.accessDenied"     // Отказано в доступе
	ErrNotFound         = "monetization.errors.auth.notFound"         // Не найдено
	ErrUnauthorized     = "monetization.errors.auth.unauthorized"     // Не авторизован
	ErrConflict         = "monetization.errors.auth.conflict"         // Конфликт данных
)
