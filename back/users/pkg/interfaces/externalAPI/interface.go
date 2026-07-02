// Package externalapi describes the public users API contract.
// @tg version=0.0.1
// @tg backend=users
// @tg title=`users`
// @tg servers=
//
//go:generate tg transport --services . --out ../../../internal/transport/jsonRPC/externalapi --outSwagger ../../../swaggers/externalapi/swagger.yaml
package externalapi

import (
	"context"

	"github.com/google/uuid"
	"github.com/mbatimel/AMC/users/pkg/models"
)

// UsersAPI
// @tg http-server metrics log
// @tg http-prefix=/api
// @tg 200=github.com/mbatimel/AMC/users/swaggers/externalapi/models:Resp200
// @tg 400=github.com/mbatimel/AMC/users/swaggers/externalapi/models:Err400
// @tg 401=github.com/mbatimel/AMC/users/swaggers/externalapi/models:Err401
// @tg 403=github.com/mbatimel/AMC/users/swaggers/externalapi/models:Err403
// @tg 500=github.com/mbatimel/AMC/users/swaggers/externalapi/models:Err500
type UsersAPI interface {
	// CreateUser ...
	// @tg http-method=POST
	// @tg http-path=/v1/users
	// @tg http-args=email|email
	// @tg http-args=phone|phone
	// @tg http-args=firstName|firstName
	// @tg http-args=lastName|lastName
	// @tg http-args=middleName|middleName
	// @tg http-args=role|role
	// @tg http-args=status|status
	// @tg http-args=clientID|clientID
	// @tg http-args=companyName|companyName
	// @tg http-args=inn|inn
	// @tg http-args=isActive|isActive
	// @tg http-response=github.com/mbatimel/AMC/users/internal/transport/custom-handlers:CreateUser
	// @tg summary=`Создание пользователя`
	// @tg desc=`Создание пользователя личного кабинета`
	CreateUser(ctx context.Context, email string, phone string, firstName string, lastName string, middleName string, role string, status string, clientID string, companyName string, inn string, isActive bool) (response models.CreateUserResponse, err error)

	// GetUser ...
	// @tg http-method=GET
	// @tg http-path=/v1/users/{userID}
	// @tg http-args=userID|userID
	// @tg http-response=github.com/mbatimel/AMC/users/internal/transport/custom-handlers:GetUser
	// @tg summary=`Получение пользователя`
	// @tg desc=`Получение пользователя по идентификатору`
	// @tg uuidPackage=github.com/google/uuid
	GetUser(ctx context.Context, userID uuid.UUID) (response models.GetUserResponse, err error)

	// ListUsers ...
	// @tg http-method=GET
	// @tg http-path=/v1/users
	// @tg http-args=q|q
	// @tg http-args=role|role
	// @tg http-args=status|status
	// @tg http-args=clientID|clientID
	// @tg http-args=isActive|isActive
	// @tg http-args=limit|limit
	// @tg http-args=offset|offset
	// @tg http-args=sort|sort
	// @tg http-response=github.com/mbatimel/AMC/users/internal/transport/custom-handlers:ListUsers
	// @tg summary=`Список пользователей`
	// @tg desc=`Получение списка пользователей с фильтрами и пагинацией`
	ListUsers(ctx context.Context, q string, role string, status string, clientID string, isActive bool, limit int, offset int, sort string) (response models.ListUsersResponse, err error)

	// UpdateUser ...
	// @tg http-method=PATCH
	// @tg http-path=/v1/users/{userID}
	// @tg http-args=userID|userID
	// @tg http-args=email|email
	// @tg http-args=phone|phone
	// @tg http-args=firstName|firstName
	// @tg http-args=lastName|lastName
	// @tg http-args=middleName|middleName
	// @tg http-args=role|role
	// @tg http-args=status|status
	// @tg http-args=clientID|clientID
	// @tg http-args=companyName|companyName
	// @tg http-args=inn|inn
	// @tg http-args=isActive|isActive
	// @tg http-response=github.com/mbatimel/AMC/users/internal/transport/custom-handlers:UpdateUser
	// @tg summary=`Обновление пользователя`
	// @tg desc=`Обновление данных пользователя`
	// @tg uuidPackage=github.com/google/uuid
	UpdateUser(ctx context.Context, userID uuid.UUID, email string, phone string, firstName string, lastName string, middleName string, role string, status string, clientID string, companyName string, inn string, isActive bool) (response models.UpdateUserResponse, err error)

	// DeleteUser ...
	// @tg http-method=DELETE
	// @tg http-path=/v1/users/{userID}
	// @tg http-args=userID|userID
	// @tg http-response=github.com/mbatimel/AMC/users/internal/transport/custom-handlers:DeleteUser
	// @tg summary=`Удаление пользователя`
	// @tg desc=`Удаление или скрытие пользователя`
	// @tg uuidPackage=github.com/google/uuid
	DeleteUser(ctx context.Context, userID uuid.UUID) (response models.DeleteUserResponse, err error)

	// ActivateUser ...
	// @tg http-method=POST
	// @tg http-path=/v1/users/{userID}/activate
	// @tg http-args=userID|userID
	// @tg http-response=github.com/mbatimel/AMC/users/internal/transport/custom-handlers:ActivateUser
	// @tg summary=`Активация пользователя`
	// @tg desc=`Активация пользователя`
	// @tg uuidPackage=github.com/google/uuid
	ActivateUser(ctx context.Context, userID uuid.UUID) (response models.ActivateUserResponse, err error)

	// DeactivateUser ...
	// @tg http-method=POST
	// @tg http-path=/v1/users/{userID}/deactivate
	// @tg http-args=userID|userID
	// @tg http-response=github.com/mbatimel/AMC/users/internal/transport/custom-handlers:DeactivateUser
	// @tg summary=`Деактивация пользователя`
	// @tg desc=`Деактивация пользователя`
	// @tg uuidPackage=github.com/google/uuid
	DeactivateUser(ctx context.Context, userID uuid.UUID) (response models.DeactivateUserResponse, err error)
}
