// Package externalapi
// @tg version=0.0.1
// @tg backend=auth
// @tg title=`auth`
// @tg servers=
//
//go:generate tg transport --services . --out ../../../internal/transport/jsonRPC/externalapi --outSwagger ../../../swaggers/externalapi/swagger.yaml
package externalAPI

import (
	"context"

	"github.com/google/uuid"
)

// AuthAPI
// @tg http-server metrics log
// @tg http-prefix=/api
// @tg 200=github.com/mbatimel/AMC/auth/swaggers/externalapi/models:Resp200
// @tg 400=github.com/mbatimel/AMC/auth/swaggers/externalapi/models:Err400
// @tg 401=github.com/mbatimel/AMC/auth/swaggers/externalapi/models:Err401
// @tg 403=github.com/mbatimel/AMC/auth/swaggers/externalapi/models:Err403
// @tg 500=github.com/mbatimel/AMC/auth/swaggers/externalapi/models:Err500
type AuthAPI interface {

	// LoginUser ...
	// @tg http-method=POST
	// @tg http-path=/v1/auth/login
	// @tg uuidPackage=github.com/google/uuid
	// @tg summary=`Авторизация пользователя`
	// @tg desc=`Авторизация пользователя по email и паролю`
	LoginUser(
		ctx context.Context,
		email string,
		password string,
	) (userID uuid.UUID, err error)

	// RegisterIP ...
	// @tg http-method=POST
	// @tg http-path=/v1/auth/register/ip
	// @tg uuidPackage=github.com/google/uuid
	// @tg summary=`Регистрация ИП/организации`
	// @tg desc=`Создание нового пользователя с реквизитами ИП/организации. Все поля кроме email и пароля необязательны.`
	RegisterIP(
		ctx context.Context,
		email string,
		password string,
		fullName *string,
		shortName *string,
		inn *string,
		kpp *string,
		ogrn *string,
		okved *string,
		taxSystem *string,
		legalAddress *string,
		actualAddress *string,
		directorFullName *string,
		directorPosition *string,
		phone *string,
		additionalPhone *string,
		website *string,
		bankAccount *string,
		bankName *string,
		bankBik *string,
		correspondentAccount *string,
	) (userID uuid.UUID, err error)

	// RegisterIndividual ...
	// @tg http-method=POST
	// @tg http-path=/v1/auth/register/individual
	// @tg uuidPackage=github.com/google/uuid
	// @tg summary=`Регистрация физического лица`
	// @tg desc=`Создание нового пользователя-физлица. ИНН необязателен, остальные поля обязательны.`
	RegisterIndividual(
		ctx context.Context,
		fio string,
		phone string,
		email string,
		deliveryAddress string,
		password string,
		city string,
		inn *string,
	) (userID uuid.UUID, err error)

	// LogoutUser ...
	// @tg http-method=POST
	// @tg http-path=/v1/auth/logout
	// @tg http-headers=userID|X-User-Id
	// @tg uuidPackage=github.com/google/uuid
	// @tg summary=`Выход пользователя`
	// @tg desc=`Завершение пользовательской сессии`
	LogoutUser(
		ctx context.Context,
		userID uuid.UUID,
	) (err error)

	// ChangePassword ...
	// @tg http-method=POST
	// @tg http-path=/v1/auth/change-password
	// @tg http-headers=userID|X-User-Id
	// @tg http-args=oldPassword|oldPassword
	// @tg http-args=newPassword|newPassword
	// @tg uuidPackage=github.com/google/uuid
	// @tg summary=`Изменение пароля`
	// @tg desc=`Изменение пароля пользователя`
	ChangePassword(
		ctx context.Context,
		userID uuid.UUID,
		oldPassword string,
		newPassword string,
	) (err error)

	// VerifyEmailCode ...
	// @tg http-method=POST
	// @tg http-path=/v1/auth/verify-email
	// @tg http-headers=userID|X-User-Id
	// @tg http-args=code|code
	// @tg uuidPackage=github.com/google/uuid
	// @tg summary=`Подтверждение email`
	// @tg desc=`Подтверждение электронной почты пользователя`
	VerifyEmailCode(
		ctx context.Context,
		userID uuid.UUID,
		code int64,
	) (err error)

	// SendEmailVerification ...
	// @tg http-method=POST
	// @tg http-path=/v1/auth/send-verification
	// @tg http-headers=userID|X-User-Id
	// @tg uuidPackage=github.com/google/uuid
	// @tg summary=`Отправить код подтверждения`
	// @tg desc=`Отправка письма с кодом подтверждения email`
	SendEmailVerification(
		ctx context.Context,
		userID uuid.UUID,
	) (err error)
}
