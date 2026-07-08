// Package internalapi describes the internal access API contract.
// @tg version=0.0.1
// @tg backend=access
// @tg title=`access`
// @tg servers=
//
//go:generate tg transport --services . --out ../../../internal/transport/jsonRPC/internalapi --outSwagger ../../../swaggers/internalapi/swagger.yaml
//go:generate tg client --services . --outPath ../../client/transport -go
package internalAPI

import (
	"context"

	"github.com/google/uuid"
)

// AccessAPI
// @tg http-server metrics log
// @tg http-prefix=/api
// @tg 200=github.com/mbatimel/AMC/access/swaggers/internalapi/models:Resp200
// @tg 400=github.com/mbatimel/AMC/access/swaggers/internalapi/models:Err400
// @tg 401=github.com/mbatimel/AMC/access/swaggers/internalapi/models:Err401
// @tg 403=github.com/mbatimel/AMC/access/swaggers/internalapi/models:Err403
// @tg 500=github.com/mbatimel/AMC/access/swaggers/internalapi/models:Err500
type AccessAPI interface {
	// CheckAccess ...
	// @tg http-method=POST
	// @tg http-path=/v1/access/check
	// @tg http-args=userID|userID
	// @tg http-args=role|role
	// @tg uuidPackage=github.com/google/uuid
	// @tg userID.format=uuid
	// @tg role.type=int
	// @tg summary=`Проверка доступа пользователя`
	// @tg desc=`Проверяет, что фактическая роль пользователя не ниже ожидаемой роли`
	CheckAccess(
		ctx context.Context,
		userID uuid.UUID,
		role int,
	) (allowed bool, err error)

	// AddRole ...
	// @tg http-method=POST
	// @tg http-path=/v1/access/role/add
	// @tg http-headers=adminUserID|X-Admin-User-Id
	// @tg http-args=userID|userID
	// @tg http-args=role|role
	// @tg uuidPackage=github.com/google/uuid
	// @tg adminUserID.format=uuid
	// @tg userID.format=uuid
	// @tg role.type=int
	// @tg summary=`Добавление роли пользователю`
	// @tg desc=`Добавляет пользователю роль с указанным кодом после проверки прав администратора`
	AddRole(
		ctx context.Context,
		adminUserID uuid.UUID,
		userID uuid.UUID,
		role int,
	) (success bool, err error)

	// UpdateRole ...
	// @tg http-method=POST
	// @tg http-path=/v1/access/role/update
	// @tg http-headers=adminUserID|X-Admin-User-Id
	// @tg http-args=userID|userID
	// @tg http-args=role|role
	// @tg uuidPackage=github.com/google/uuid
	// @tg adminUserID.format=uuid
	// @tg userID.format=uuid
	// @tg role.type=int
	// @tg summary=`Обновление роли пользователя`
	// @tg desc=`Обновляет роль пользователя после проверки прав администратора`
	UpdateRole(
		ctx context.Context,
		adminUserID uuid.UUID,
		userID uuid.UUID,
		role int,
	) (success bool, err error)

	// DeleteRole ...
	// @tg http-method=POST
	// @tg http-path=/v1/access/role/delete
	// @tg http-headers=adminUserID|X-Admin-User-Id
	// @tg http-args=userID|userID
	// @tg http-args=role|role
	// @tg uuidPackage=github.com/google/uuid
	// @tg adminUserID.format=uuid
	// @tg userID.format=uuid
	// @tg role.type=int
	// @tg summary=`Удаление роли пользователя`
	// @tg desc=`Удаляет указанную роль пользователя после проверки прав администратора`
	DeleteRole(
		ctx context.Context,
		adminUserID uuid.UUID,
		userID uuid.UUID,
		role int,
	) (success bool, err error)
}
