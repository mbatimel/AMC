// Package externalapi describes the public users API contract.
// @tg version=0.0.1
// @tg backend=users
// @tg title=`users`
// @tg servers=
//
//go:generate tg transport --services . --out ../../../internal/transport/jsonRPC/externalapi --outSwagger ../../../swaggers/externalapi/swagger.yaml
package externalAPI

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
// @tg 404=github.com/mbatimel/AMC/users/swaggers/externalapi/models:Err400
// @tg 409=github.com/mbatimel/AMC/users/swaggers/externalapi/models:Err400
// @tg 500=github.com/mbatimel/AMC/users/swaggers/externalapi/models:Err500
type UsersAPI interface {
	// CreateUser creates an administrative user account.
	// @tg http-method=POST
	// @tg http-path=/v1/users
	// @tg http-headers=adminUserID|X-Admin-User-Id
	// @tg http-response=github.com/mbatimel/AMC/users/internal/transport/custom-handlers:CreateUser
	// @tg summary=`Создание пользователя`
	// @tg desc=`Создаёт пользователя и при необходимости связывает его с клиентским кабинетом`
	// @tg uuidPackage=github.com/google/uuid
	// @tg adminUserID.format=uuid
	CreateUser(ctx context.Context, adminUserID uuid.UUID, email string, phone string, firstName string, lastName string, middleName string, role string, status string, clientID string, companyName string, inn string, isActive bool) (response models.CreateUserResponse, err error)

	// GetUser returns a user by ID.
	// @tg http-method=GET
	// @tg http-path=/v1/users/:userID
	// @tg http-response=github.com/mbatimel/AMC/users/internal/transport/custom-handlers:GetUser
	// @tg summary=`Получение пользователя`
	// @tg desc=`Возвращает пользователя по идентификатору`
	// @tg uuidPackage=github.com/google/uuid
	// @tg userID.format=uuid
	GetUser(ctx context.Context, userID uuid.UUID) (response models.GetUserResponse, err error)

	// ListUsers returns filtered users.
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
	// @tg desc=`Возвращает пользователей с фильтрами, безопасной сортировкой и пагинацией`
	ListUsers(ctx context.Context, q string, role string, status string, clientID string, isActive *bool, limit int, offset int, sort string) (response models.ListUsersResponse, err error)

	// UpdateUser updates allowed administrative user fields.
	// @tg http-method=PATCH
	// @tg http-path=/v1/users/:userID
	// @tg http-headers=adminUserID|X-Admin-User-Id
	// @tg http-response=github.com/mbatimel/AMC/users/internal/transport/custom-handlers:UpdateUser
	// @tg summary=`Обновление пользователя`
	// @tg desc=`Обновляет разрешённые поля пользователя и роль через access-сервис`
	// @tg uuidPackage=github.com/google/uuid
	// @tg adminUserID.format=uuid
	// @tg userID.format=uuid
	UpdateUser(ctx context.Context, adminUserID uuid.UUID, userID uuid.UUID, email string, phone string, firstName string, lastName string, middleName string, role string, status string, clientID string, companyName string, inn string, isActive *bool) (response models.UpdateUserResponse, err error)

	// DeleteUser soft-deletes a user.
	// @tg http-method=DELETE
	// @tg http-path=/v1/users/:userID
	// @tg http-response=github.com/mbatimel/AMC/users/internal/transport/custom-handlers:DeleteUser
	// @tg summary=`Удаление пользователя`
	// @tg desc=`Помечает пользователя удалённым и запрещает дальнейшее использование users API`
	// @tg uuidPackage=github.com/google/uuid
	// @tg userID.format=uuid
	DeleteUser(ctx context.Context, userID uuid.UUID) (response models.DeleteUserResponse, err error)

	// ActivateUser activates a user.
	// @tg http-method=POST
	// @tg http-path=/v1/users/:userID/activate
	// @tg http-response=github.com/mbatimel/AMC/users/internal/transport/custom-handlers:ActivateUser
	// @tg summary=`Активация пользователя`
	// @tg desc=`Устанавливает активный статус пользователя`
	// @tg uuidPackage=github.com/google/uuid
	// @tg userID.format=uuid
	ActivateUser(ctx context.Context, userID uuid.UUID) (response models.ActivateUserResponse, err error)

	// DeactivateUser deactivates a user.
	// @tg http-method=POST
	// @tg http-path=/v1/users/:userID/deactivate
	// @tg http-response=github.com/mbatimel/AMC/users/internal/transport/custom-handlers:DeactivateUser
	// @tg summary=`Деактивация пользователя`
	// @tg desc=`Устанавливает неактивный статус пользователя`
	// @tg uuidPackage=github.com/google/uuid
	// @tg userID.format=uuid
	DeactivateUser(ctx context.Context, userID uuid.UUID) (response models.DeactivateUserResponse, err error)

	// GetProfile returns the current user's profile.
	// @tg http-method=GET
	// @tg http-path=/v1/profile
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/users/internal/transport/custom-handlers:GetProfile
	// @tg summary=`Профиль пользователя`
	// @tg desc=`Возвращает профиль текущего пользователя и активный кабинет`
	// @tg uuidPackage=github.com/google/uuid
	// @tg userID.format=uuid
	GetProfile(ctx context.Context, userID uuid.UUID) (response models.GetProfileResponse, err error)

	// UpdateProfile updates the current user's contact fields.
	// @tg http-method=PATCH
	// @tg http-path=/v1/profile
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/users/internal/transport/custom-handlers:UpdateProfile
	// @tg summary=`Обновление профиля`
	// @tg desc=`Обновляет только контактные данные текущего пользователя`
	// @tg uuidPackage=github.com/google/uuid
	// @tg userID.format=uuid
	UpdateProfile(ctx context.Context, userID uuid.UUID, email string, phone string, firstName string, lastName string, middleName string) (response models.UpdateProfileResponse, err error)

	// ListUserClients lists the current user's client cabinets.
	// @tg http-method=GET
	// @tg http-path=/v1/profile/clients
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/users/internal/transport/custom-handlers:ListUserClients
	// @tg summary=`Кабинеты пользователя`
	// @tg desc=`Возвращает только клиентские кабинеты, доступные текущему пользователю`
	// @tg uuidPackage=github.com/google/uuid
	// @tg userID.format=uuid
	ListUserClients(ctx context.Context, userID uuid.UUID) (response models.ListUserClientsResponse, err error)

	// GetClientDetails returns linked client details.
	// @tg http-method=GET
	// @tg http-path=/v1/profile/clients/:clientID
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/users/internal/transport/custom-handlers:GetClientDetails
	// @tg summary=`Реквизиты кабинета`
	// @tg desc=`Возвращает реквизиты доступного пользователю клиентского кабинета`
	// @tg uuidPackage=github.com/google/uuid
	// @tg userID.format=uuid
	// @tg clientID.format=uuid
	GetClientDetails(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (response models.GetClientDetailsResponse, err error)

	// GetClientConditions returns linked client conditions.
	// @tg http-method=GET
	// @tg http-path=/v1/profile/clients/:clientID/conditions
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/users/internal/transport/custom-handlers:GetClientConditions
	// @tg summary=`Индивидуальные условия`
	// @tg desc=`Возвращает ценовую группу, кредитные условия и скидки доступного кабинета`
	// @tg uuidPackage=github.com/google/uuid
	// @tg userID.format=uuid
	// @tg clientID.format=uuid
	GetClientConditions(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (response models.GetClientConditionsResponse, err error)

	// SwitchActiveClient changes the current active client.
	// @tg http-method=POST
	// @tg http-path=/v1/profile/clients/:clientID/activate
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/users/internal/transport/custom-handlers:SwitchActiveClient
	// @tg summary=`Переключение кабинета`
	// @tg desc=`Транзакционно сохраняет доступный клиентский кабинет как активный`
	// @tg uuidPackage=github.com/google/uuid
	// @tg userID.format=uuid
	// @tg clientID.format=uuid
	SwitchActiveClient(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (response models.SwitchActiveClientResponse, err error)

	// ListFavorites returns favorites for the active client.
	// @tg http-method=GET
	// @tg http-path=/v1/favorites
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/users/internal/transport/custom-handlers:ListFavorites
	// @tg summary=`Список избранного`
	// @tg desc=`Возвращает идентификаторы избранных товаров текущего пользователя и активного кабинета`
	// @tg uuidPackage=github.com/google/uuid
	// @tg userID.format=uuid
	ListFavorites(ctx context.Context, userID uuid.UUID) (response models.ListFavoritesResponse, err error)

	// AddFavorite adds a product to favorites.
	// @tg http-method=POST
	// @tg http-path=/v1/favorites
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/users/internal/transport/custom-handlers:AddFavorite
	// @tg summary=`Добавление в избранное`
	// @tg desc=`Идемпотентно добавляет товар в избранное активного кабинета`
	// @tg uuidPackage=github.com/google/uuid
	// @tg userID.format=uuid
	AddFavorite(ctx context.Context, userID uuid.UUID, productID string) (response models.AddFavoriteResponse, err error)

	// DeleteFavorites deletes several favorites at once.
	// @tg http-method=DELETE
	// @tg http-path=/v1/favorites
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/users/internal/transport/custom-handlers:DeleteFavorites
	// @tg summary=`Массовое удаление из избранного`
	// @tg desc=`Удаляет массив идентификаторов товаров одним параметризованным SQL-запросом`
	// @tg uuidPackage=github.com/google/uuid
	// @tg userID.format=uuid
	DeleteFavorites(ctx context.Context, userID uuid.UUID, productIDs []string) (response models.DeleteFavoritesResponse, err error)
}
