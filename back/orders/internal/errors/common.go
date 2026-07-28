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
	NotFoundError       = func() *Error { return New("not found", fasthttp.StatusNotFound, ErrNotFound) }
	NotImplementedError = func() *Error {
		return New("not implemented", fasthttp.StatusNotImplemented, ErrNotImplemented)
	}
)

const (
	ErrInternal         = "orders.errors.internalError"    // Внутренняя ошибка
	ErrBadRequest       = "orders.errors.badRequest"       // Плохой запрос
	ErrMethodNotAllowed = "orders.errors.methodNotAllowed" // Метод не поддерживается
	ErrForbidden        = "orders.errors.forbidden"        // Доступ запрещен
	ErrInvalidRequest   = "orders.errors.invalidRequest"   // Неправильный запрос
	ErrAccessDenied     = "orders.errors.accessDenied"     // Отказано в доступе
	ErrNotFound         = "orders.errors.notFound"         // Не найдено
	ErrNotImplemented   = "orders.errors.notImplemented"   // Метод не реализован
)
