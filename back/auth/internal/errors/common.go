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
	InnEmptyErr = func(field string) *Error {
		return New("inn is empty", fasthttp.StatusBadRequest, ErrInvalidRequest).AddCause("field", field)
	}
	InnInvalidError = func(inn string) *Error {
		return New("inn is invalid", fasthttp.StatusBadRequest, ErrInvalidRequest).AddCause("inn", inn)
	}
	InnTakenError = func(inn string) *Error {
		return New("inn already registered", fasthttp.StatusConflict, ErrInnTaken).AddCause("inn", inn)
	}
	RequisitesFileRequiredError = func() *Error {
		return New("requisitesFile required", fasthttp.StatusBadRequest, ErrInvalidRequest).AddCause("field", "requisitesFile")
	}
	RequisitesFileTooLargeError = func() *Error {
		return New("requisitesFile too large", fasthttp.StatusBadRequest, ErrInvalidRequest).AddCause("field", "requisitesFile")
	}
	RequisitesFileInvalidTypeError = func() *Error {
		return New("requisitesFile invalid type", fasthttp.StatusBadRequest, ErrInvalidRequest).AddCause("field", "requisitesFile")
	}
	UserBlockedError = func(reason, contactName, contactPhone, contactEmail string) *Error {
		err := New("user is blocked", fasthttp.StatusForbidden, ErrUserBlocked)
		cause := make([]string, 0, 8)
		if reason != "" {
			cause = append(cause, "reason", reason)
		}
		if contactName != "" {
			cause = append(cause, "contactName", contactName)
		}
		if contactPhone != "" {
			cause = append(cause, "contactPhone", contactPhone)
		}
		if contactEmail != "" {
			cause = append(cause, "contactEmail", contactEmail)
		}
		if len(cause) > 0 {
			err = err.AddCause(cause...)
		}
		return err
	}
	TokenInvalidError = func() *Error {
		return New("token is invalid", fasthttp.StatusBadRequest, ErrTokenInvalid)
	}
	TokenExpiredError = func() *Error {
		return New("token is expired", fasthttp.StatusBadRequest, ErrTokenExpired)
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
	ErrInnTaken           = "auth.errors.innTaken"           // ИНН уже зарегистрирован
	ErrNotFound           = "auth.errors.notFound"           // Не найдено
	ErrUserBlocked        = "auth.errors.userBlocked"        // Пользователь заблокирован
	ErrTokenInvalid       = "auth.errors.tokenInvalid"       // Токен сброса пароля недействителен
	ErrTokenExpired       = "auth.errors.tokenExpired"       // Токен сброса пароля просрочен
)
