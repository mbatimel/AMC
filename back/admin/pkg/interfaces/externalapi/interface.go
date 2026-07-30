// Package externalapi describes the public admin panel API contract.
// @tg version=0.0.1
// @tg backend=admin
// @tg title=`admin`
// @tg servers=
//
//go:generate tg transport --services . --out ../../../internal/transport/jsonRPC/externalapi --outSwagger ../../../swaggers/externalapi/swagger.yaml
package externalapi

import (
	"context"

	"github.com/google/uuid"
	"github.com/mbatimel/AMC/admin/pkg/models"
)

// AdminAPI
// @tg http-server metrics log
// @tg http-prefix=/api
// @tg 200=github.com/mbatimel/AMC/admin/swaggers/externalapi/models:Resp200
// @tg 400=github.com/mbatimel/AMC/admin/swaggers/externalapi/models:Err400
// @tg 401=github.com/mbatimel/AMC/admin/swaggers/externalapi/models:Err401
// @tg 403=github.com/mbatimel/AMC/admin/swaggers/externalapi/models:Err403
// @tg 500=github.com/mbatimel/AMC/admin/swaggers/externalapi/models:Err500
type AdminAPI interface {
	// Login ...
	// @tg http-method=POST
	// @tg http-path=/v1/admin/auth/login
	// @tg http-args=email|email
	// @tg http-args=password|password
	// @tg http-response=github.com/mbatimel/AMC/admin/internal/transport/custom-handlers:Login
	// @tg summary=`Вход в админ-панель`
	// @tg desc=`Авторизация сотрудника портала по email и паролю`
	// @tg uuidPackage=github.com/google/uuid
	Login(ctx context.Context, email string, password string) (response models.LoginResponse, err error)

	// Logout ...
	// @tg http-method=POST
	// @tg http-path=/v1/admin/auth/logout
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/admin/internal/transport/custom-handlers:Logout
	// @tg summary=`Выход из админ-панели`
	// @tg desc=`Завершение сессии сотрудника портала`
	// @tg uuidPackage=github.com/google/uuid
	Logout(ctx context.Context, userID uuid.UUID) (err error)

	// GetSession ...
	// @tg http-method=GET
	// @tg http-path=/v1/admin/auth/session
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/admin/internal/transport/custom-handlers:GetSession
	// @tg summary=`Проверка сессии`
	// @tg desc=`Проверка, что пользователь авторизован и имеет права администратора`
	// @tg uuidPackage=github.com/google/uuid
	GetSession(ctx context.Context, userID uuid.UUID) (response models.SessionResponse, err error)

	// ListAuditLog ...
	// @tg http-method=GET
	// @tg http-path=/v1/admin/audit-log
	// @tg http-headers=userID|X-User-Id
	// @tg http-args=limit|limit
	// @tg http-args=offset|offset
	// @tg http-response=github.com/mbatimel/AMC/admin/internal/transport/custom-handlers:ListAuditLog
	// @tg summary=`Журнал действий`
	// @tg desc=`Список действий администраторов портала без возможности удаления`
	// @tg uuidPackage=github.com/google/uuid
	ListAuditLog(ctx context.Context, userID uuid.UUID, limit int, offset int) (response models.ListAuditLogResponse, err error)
}
