package errors

import (
	"github.com/valyala/fasthttp"
)

var (
	BadRequestError = func() *Error { return New("bad request", fasthttp.StatusBadRequest, ErrBadRequest) }
	UnauthorizedError = func() *Error { return New("unauthorized", fasthttp.StatusUnauthorized, ErrUnauthorized) }
	NotFoundError = func() *Error { return New("not found", fasthttp.StatusNotFound, ErrNotFound) }
	ConflictError = func() *Error { return New("conflict", fasthttp.StatusConflict, ErrConflict) }
	InternalServerError = func() *Error {
		return New("internal server error", fasthttp.StatusInternalServerError, ErrInternal)
	}
)

const (
	ErrInternal    = "integrations.errors.internalError"
	ErrBadRequest  = "integrations.errors.badRequest"
	ErrUnauthorized = "integrations.errors.unauthorized"
	ErrNotFound    = "integrations.errors.notFound"
	ErrConflict    = "integrations.errors.conflict"
)
