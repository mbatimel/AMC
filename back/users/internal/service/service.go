package service

import (
	"context"
	"errors"
	"net/mail"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/mbatimel/AMC/users/internal/clients"
	customErrors "github.com/mbatimel/AMC/users/internal/errors"
	internalModels "github.com/mbatimel/AMC/users/internal/models"
	"github.com/mbatimel/AMC/users/internal/storage/postgres"
	"github.com/mbatimel/AMC/users/pkg/models"
)

const (
	defaultLimit         = 20
	maxLimit             = 100
	maxUserNameLength    = 40
	maxCompanyNameLength = 255
)

type Storage interface {
	CreateUser(ctx context.Context, params internalModels.CreateUserParams) (internalModels.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (internalModels.User, error)
	ListUsers(ctx context.Context, params internalModels.ListUsersParams) ([]internalModels.User, error)
	CountUsers(ctx context.Context, params internalModels.ListUsersParams) (int, error)
	UpdateUser(ctx context.Context, params internalModels.UpdateUserParams) (internalModels.User, error)
	SoftDeleteUser(ctx context.Context, userID uuid.UUID) error
	SetUserActive(ctx context.Context, userID uuid.UUID, active bool) (internalModels.User, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (internalModels.User, *internalModels.Client, error)
	UpdateProfile(ctx context.Context, params internalModels.UpdateProfileParams) (internalModels.User, *internalModels.Client, error)
	ListUserClients(ctx context.Context, userID uuid.UUID) ([]internalModels.Client, error)
	UserHasClient(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (bool, error)
	GetClientDetails(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (internalModels.Client, error)
	GetClientConditions(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (internalModels.ClientConditions, error)
	SetActiveClient(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (internalModels.Client, error)
	GetActiveClient(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
	ListFavorites(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) ([]internalModels.Favorite, error)
	AddFavorite(ctx context.Context, userID uuid.UUID, clientID uuid.UUID, productID uuid.UUID) (internalModels.Favorite, bool, error)
	DeleteFavorites(ctx context.Context, userID uuid.UUID, clientID uuid.UUID, productIDs []uuid.UUID) (int, error)
}

type Service struct {
	logger       zerolog.Logger
	storage      Storage
	accessClient clients.AccessClient
}

func New(logger zerolog.Logger, storage Storage, accessClient clients.AccessClient) *Service {
	return &Service{logger: logger, storage: storage, accessClient: accessClient}
}

func validation(field string) error {
	return customErrors.ErrValidation.AddCause("field", field)
}

func normalizeEmail(value string, required bool) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		if required {
			return "", validation("email")
		}
		return "", nil
	}
	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value {
		return "", validation("email")
	}
	return value, nil
}

func normalizePhone(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	digits := 0
	for _, r := range value {
		switch {
		case r >= '0' && r <= '9':
			digits++
		case strings.ContainsRune("+()- ", r):
		default:
			return "", validation("phone")
		}
	}
	if digits < 7 || digits > 15 {
		return "", validation("phone")
	}
	return value, nil
}

func validateLength(field, value string, maxLength int) error {
	if utf8.RuneCountInString(value) > maxLength {
		return validation(field)
	}
	return nil
}

func parseRole(value string) (int, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	roleCodes := map[string]int{"admin": 0, "support": 1, "supplier": 2, "buyer": 3}
	if code, ok := roleCodes[value]; ok {
		return code, nil
	}
	code, err := strconv.Atoi(value)
	if err != nil || code < 0 || code > 3 {
		return 0, validation("role")
	}
	return code, nil
}

func parseOptionalUUID(value, field string) (*uuid.UUID, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	id, err := uuid.Parse(value)
	if err != nil || id == uuid.Nil {
		return nil, validation(field)
	}
	return &id, nil
}

func mapStorageError(err error) error {
	switch {
	case errors.Is(err, postgres.ErrUserNotFound):
		return customErrors.ErrNotFound.AddCause("entity", "user")
	case errors.Is(err, postgres.ErrClientNotFound):
		return customErrors.ErrNotFound.AddCause("entity", "client")
	case errors.Is(err, postgres.ErrFavoriteNotFound):
		return customErrors.ErrNotFound.AddCause("entity", "favorite")
	case errors.Is(err, postgres.ErrEmailTaken):
		return customErrors.ErrConflict.AddCause("field", "email")
	case errors.Is(err, postgres.ErrPhoneTaken):
		return customErrors.ErrConflict.AddCause("field", "phone")
	case errors.Is(err, postgres.ErrProductNotFound):
		return customErrors.ErrNotFound.AddCause("entity", "product")
	case errors.Is(err, postgres.ErrInvalidSort):
		return validation("sort")
	default:
		return customErrors.ErrInternal
	}
}

func modelUser(row internalModels.User) models.User {
	user := models.User{
		ID:          row.ID.String(),
		Email:       row.Email,
		Phone:       row.Phone,
		FirstName:   row.FirstName,
		LastName:    row.LastName,
		MiddleName:  row.MiddleName,
		Role:        row.Role,
		Status:      row.Status,
		CompanyName: row.CompanyName,
		INN:         row.INN,
		IsActive:    row.IsActive,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
	if row.ClientID.Valid {
		user.ClientID = row.ClientID.UUID.String()
	}
	if row.ActiveClientID.Valid {
		user.ActiveClientID = row.ActiveClientID.UUID.String()
	}
	if row.DeletedAt.Valid {
		deletedAt := row.DeletedAt.Time
		user.DeletedAt = &deletedAt
	}
	return user
}

func modelClient(row internalModels.Client) models.Client {
	return models.Client{
		ID:          row.ID.String(),
		CompanyName: row.CompanyName,
		CompanyType: row.CompanyType,
		INN:         row.INN,
		OGRN:        row.OGRN,
		Address:     row.Address,
		ContactName: row.ContactName,
		Phone:       row.Phone,
		Email:       row.Email,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func modelProfile(user internalModels.User, client *internalModels.Client) models.Profile {
	profile := models.Profile{
		UserID:     user.ID.String(),
		Email:      user.Email,
		Phone:      user.Phone,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
		Status:     user.Status,
		IsActive:   user.IsActive,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}
	if user.ActiveClientID.Valid {
		profile.ActiveClientID = user.ActiveClientID.UUID.String()
	}
	if client != nil {
		activeClient := modelClient(*client)
		profile.ActiveClient = &activeClient
	}
	return profile
}

func modelFavorite(row internalModels.Favorite) models.Favorite {
	return models.Favorite{
		UserID:    row.UserID.String(),
		ClientID:  row.ClientID.String(),
		ProductID: row.ProductID.String(),
		CreatedAt: row.CreatedAt,
	}
}

func requireUserID(userID uuid.UUID) error {
	if userID == uuid.Nil {
		return validation("X-User-Id")
	}
	return nil
}

func (s *Service) checkBuyerAccess(ctx context.Context, userID uuid.UUID) error {
	if err := requireUserID(userID); err != nil {
		return err
	}
	if s.accessClient == nil {
		return customErrors.ErrInternal
	}
	allowed, err := s.accessClient.CheckAccess(ctx, userID, RoleCodeBuyer)
	if err != nil {
		return customErrors.ErrInternal
	}
	if !allowed {
		return customErrors.ErrForbidden
	}
	return nil
}

func (s *Service) CreateUser(
	ctx context.Context,
	adminUserID uuid.UUID,
	email string,
	phone string,
	firstName string,
	lastName string,
	middleName string,
	role string,
	status string,
	clientID string,
	companyName string,
	inn string,
	isActive bool,
) (response models.CreateUserResponse, err error) {
	email, err = normalizeEmail(email, true)
	if err != nil {
		return response, err
	}
	phone, err = normalizePhone(phone)
	if err != nil {
		return response, err
	}
	for field, value := range map[string]string{
		"firstName": firstName, "lastName": lastName, "middleName": middleName,
	} {
		if err = validateLength(field, value, maxUserNameLength); err != nil {
			return response, err
		}
	}
	if err = validateLength("companyName", companyName, maxCompanyNameLength); err != nil {
		return response, err
	}
	parsedClientID, err := parseOptionalUUID(clientID, "clientID")
	if err != nil {
		return response, err
	}
	status = strings.TrimSpace(status)
	if status == "" {
		status = "active"
		isActive = true
	}
	var roleCode int
	if strings.TrimSpace(role) != "" {
		if adminUserID == uuid.Nil {
			return response, validation("X-Admin-User-Id")
		}
		roleCode, err = parseRole(role)
		if err != nil {
			return response, err
		}
		if s.accessClient == nil {
			return response, customErrors.ErrInternal
		}
	}
	row, err := s.storage.CreateUser(ctx, internalModels.CreateUserParams{
		Email: email, Phone: phone, FirstName: strings.TrimSpace(firstName),
		LastName: strings.TrimSpace(lastName), MiddleName: strings.TrimSpace(middleName),
		Status: status, IsActive: isActive, ClientID: parsedClientID,
		CompanyName: strings.TrimSpace(companyName), INN: strings.TrimSpace(inn),
	})
	if err != nil {
		return response, mapStorageError(err)
	}
	if strings.TrimSpace(role) != "" {
		success, accessErr := s.accessClient.AddRole(ctx, adminUserID, row.ID, roleCode)
		if accessErr != nil || !success {
			s.logger.Error().Err(accessErr).Str("userID", row.ID.String()).Msg("failed to assign user role")
			return response, customErrors.ErrInternal
		}
		if row, err = s.storage.GetUserByID(ctx, row.ID); err != nil {
			return response, mapStorageError(err)
		}
	}
	return models.CreateUserResponse{User: modelUser(row)}, nil
}

func (s *Service) GetUser(ctx context.Context, userID uuid.UUID) (response models.GetUserResponse, err error) {
	if err = requireUserID(userID); err != nil {
		return response, err
	}
	row, err := s.storage.GetUserByID(ctx, userID)
	if err != nil {
		return response, mapStorageError(err)
	}
	return models.GetUserResponse{User: modelUser(row)}, nil
}

func (s *Service) ListUsers(
	ctx context.Context,
	q string,
	role string,
	status string,
	clientID string,
	isActive *bool,
	limit int,
	offset int,
	sort string,
) (response models.ListUsersResponse, err error) {
	if limit == 0 {
		limit = defaultLimit
	}
	if limit < 0 || limit > maxLimit {
		return response, validation("limit")
	}
	if offset < 0 {
		return response, validation("offset")
	}
	if role != "" {
		if _, err = parseRole(role); err != nil {
			return response, err
		}
	}
	parsedClientID, err := parseOptionalUUID(clientID, "clientID")
	if err != nil {
		return response, err
	}
	params := internalModels.ListUsersParams{
		Q: strings.TrimSpace(q), Role: strings.ToLower(strings.TrimSpace(role)),
		Status: strings.TrimSpace(status), ClientID: parsedClientID, IsActive: isActive,
		Limit: limit, Offset: offset, Sort: strings.TrimSpace(sort),
	}
	rows, err := s.storage.ListUsers(ctx, params)
	if err != nil {
		return response, mapStorageError(err)
	}
	total, err := s.storage.CountUsers(ctx, params)
	if err != nil {
		return response, mapStorageError(err)
	}
	items := make([]models.UserListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, modelUser(row))
	}
	return models.ListUsersResponse{
		Items:      items,
		Pagination: models.Pagination{Limit: limit, Offset: offset, Total: total},
	}, nil
}

func (s *Service) UpdateUser(
	ctx context.Context,
	adminUserID uuid.UUID,
	userID uuid.UUID,
	email string,
	phone string,
	firstName string,
	lastName string,
	middleName string,
	role string,
	status string,
	clientID string,
	companyName string,
	inn string,
	isActive *bool,
) (response models.UpdateUserResponse, err error) {
	if err = requireUserID(userID); err != nil {
		return response, err
	}
	if _, err = s.storage.GetUserByID(ctx, userID); err != nil {
		return response, mapStorageError(err)
	}
	email, err = normalizeEmail(email, false)
	if err != nil {
		return response, err
	}
	phone, err = normalizePhone(phone)
	if err != nil {
		return response, err
	}
	for field, value := range map[string]string{
		"firstName": firstName, "lastName": lastName, "middleName": middleName,
	} {
		if err = validateLength(field, value, maxUserNameLength); err != nil {
			return response, err
		}
	}
	if err = validateLength("companyName", companyName, maxCompanyNameLength); err != nil {
		return response, err
	}
	parsedClientID, err := parseOptionalUUID(clientID, "clientID")
	if err != nil {
		return response, err
	}
	var roleCode int
	if strings.TrimSpace(role) != "" {
		if adminUserID == uuid.Nil {
			return response, validation("X-Admin-User-Id")
		}
		roleCode, err = parseRole(role)
		if err != nil {
			return response, err
		}
		if s.accessClient == nil {
			return response, customErrors.ErrInternal
		}
	}
	row, err := s.storage.UpdateUser(ctx, internalModels.UpdateUserParams{
		UserID: userID, Email: email, Phone: phone,
		FirstName: strings.TrimSpace(firstName), LastName: strings.TrimSpace(lastName),
		MiddleName: strings.TrimSpace(middleName), Status: strings.TrimSpace(status),
		IsActive: isActive, ClientID: parsedClientID,
		CompanyName: strings.TrimSpace(companyName), INN: strings.TrimSpace(inn),
	})
	if err != nil {
		return response, mapStorageError(err)
	}
	if strings.TrimSpace(role) != "" {
		success, accessErr := s.accessClient.UpdateRole(ctx, adminUserID, userID, roleCode)
		if accessErr != nil || !success {
			s.logger.Error().Err(accessErr).Str("userID", userID.String()).Msg("failed to update user role")
			return response, customErrors.ErrInternal
		}
		if row, err = s.storage.GetUserByID(ctx, userID); err != nil {
			return response, mapStorageError(err)
		}
	}
	return models.UpdateUserResponse{User: modelUser(row)}, nil
}

func (s *Service) DeleteUser(ctx context.Context, userID uuid.UUID) (response models.DeleteUserResponse, err error) {
	if err = requireUserID(userID); err != nil {
		return response, err
	}
	if err = s.storage.SoftDeleteUser(ctx, userID); err != nil {
		return response, mapStorageError(err)
	}
	return models.DeleteUserResponse{Deleted: true}, nil
}

func (s *Service) ActivateUser(ctx context.Context, userID uuid.UUID) (response models.ActivateUserResponse, err error) {
	row, err := s.setUserActive(ctx, userID, true)
	if err != nil {
		return response, err
	}
	return models.ActivateUserResponse{User: modelUser(row)}, nil
}

func (s *Service) DeactivateUser(ctx context.Context, userID uuid.UUID) (response models.DeactivateUserResponse, err error) {
	row, err := s.setUserActive(ctx, userID, false)
	if err != nil {
		return response, err
	}
	return models.DeactivateUserResponse{User: modelUser(row)}, nil
}

func (s *Service) setUserActive(ctx context.Context, userID uuid.UUID, active bool) (internalModels.User, error) {
	if err := requireUserID(userID); err != nil {
		return internalModels.User{}, err
	}
	row, err := s.storage.SetUserActive(ctx, userID, active)
	if err != nil {
		return internalModels.User{}, mapStorageError(err)
	}
	return row, nil
}

func (s *Service) GetProfile(ctx context.Context, userID uuid.UUID) (response models.GetProfileResponse, err error) {
	if err = s.checkBuyerAccess(ctx, userID); err != nil {
		return response, err
	}
	user, client, err := s.storage.GetProfile(ctx, userID)
	if err != nil {
		return response, mapStorageError(err)
	}
	return models.GetProfileResponse{Profile: modelProfile(user, client)}, nil
}

func (s *Service) UpdateProfile(
	ctx context.Context,
	userID uuid.UUID,
	email string,
	phone string,
	firstName string,
	lastName string,
	middleName string,
) (response models.UpdateProfileResponse, err error) {
	if err = s.checkBuyerAccess(ctx, userID); err != nil {
		return response, err
	}
	email, err = normalizeEmail(email, false)
	if err != nil {
		return response, err
	}
	phone, err = normalizePhone(phone)
	if err != nil {
		return response, err
	}
	for field, value := range map[string]string{
		"firstName": firstName, "lastName": lastName, "middleName": middleName,
	} {
		if err = validateLength(field, value, maxUserNameLength); err != nil {
			return response, err
		}
	}
	user, client, err := s.storage.UpdateProfile(ctx, internalModels.UpdateProfileParams{
		UserID: userID, Email: email, Phone: phone,
		FirstName: strings.TrimSpace(firstName), LastName: strings.TrimSpace(lastName),
		MiddleName: strings.TrimSpace(middleName),
	})
	if err != nil {
		return response, mapStorageError(err)
	}
	return models.UpdateProfileResponse{Profile: modelProfile(user, client)}, nil
}

func (s *Service) ListUserClients(ctx context.Context, userID uuid.UUID) (response models.ListUserClientsResponse, err error) {
	if err = s.checkBuyerAccess(ctx, userID); err != nil {
		return response, err
	}
	if _, err = s.storage.GetUserByID(ctx, userID); err != nil {
		return response, mapStorageError(err)
	}
	rows, err := s.storage.ListUserClients(ctx, userID)
	if err != nil {
		return response, mapStorageError(err)
	}
	items := make([]models.UserClientListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, models.UserClientListItem{Client: modelClient(row), IsActive: row.IsActive})
	}
	return models.ListUserClientsResponse{Items: items}, nil
}

func (s *Service) ensureClientAccess(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) error {
	if err := s.checkBuyerAccess(ctx, userID); err != nil {
		return err
	}
	if clientID == uuid.Nil {
		return validation("clientID")
	}
	allowed, err := s.storage.UserHasClient(ctx, userID, clientID)
	if err != nil {
		return mapStorageError(err)
	}
	if !allowed {
		return customErrors.ErrForbidden.AddCause("clientID", clientID.String())
	}
	return nil
}

func (s *Service) GetClientDetails(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (response models.GetClientDetailsResponse, err error) {
	if err = s.ensureClientAccess(ctx, userID, clientID); err != nil {
		return response, err
	}
	row, err := s.storage.GetClientDetails(ctx, userID, clientID)
	if err != nil {
		return response, mapStorageError(err)
	}
	return models.GetClientDetailsResponse{Details: models.ClientDetails{Client: modelClient(row)}}, nil
}

func (s *Service) GetClientConditions(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (response models.GetClientConditionsResponse, err error) {
	if err = s.ensureClientAccess(ctx, userID, clientID); err != nil {
		return response, err
	}
	row, err := s.storage.GetClientConditions(ctx, userID, clientID)
	if err != nil {
		return response, mapStorageError(err)
	}
	discounts := make([]models.CategoryDiscount, 0, len(row.CategoryDiscounts))
	for _, discount := range row.CategoryDiscounts {
		item := models.CategoryDiscount{
			CategoryID: discount.CategoryID.String(), CategoryName: discount.CategoryName,
			DiscountPercent: discount.DiscountPercent,
		}
		if discount.ValidFrom.Valid {
			item.ValidFrom = discount.ValidFrom.Time.Format("2006-01-02")
		}
		if discount.ValidTo.Valid {
			item.ValidTo = discount.ValidTo.Time.Format("2006-01-02")
		}
		discounts = append(discounts, item)
	}
	return models.GetClientConditionsResponse{Conditions: models.ClientConditions{
		ClientID: row.ClientID.String(), PriceGroup: row.PriceGroup,
		CreditLimit: row.CreditLimit, CreditUsed: row.CreditUsed,
		PaymentTerms: row.PaymentTerms, CategoryDiscounts: discounts,
		SalesContact: row.SalesContact, ContactChannel: row.ContactChannel,
	}}, nil
}

func (s *Service) SwitchActiveClient(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (response models.SwitchActiveClientResponse, err error) {
	if err = s.ensureClientAccess(ctx, userID, clientID); err != nil {
		return response, err
	}
	row, err := s.storage.SetActiveClient(ctx, userID, clientID)
	if err != nil {
		return response, mapStorageError(err)
	}
	return models.SwitchActiveClientResponse{ActiveClient: models.UserClientListItem{
		Client: modelClient(row), IsActive: true,
	}}, nil
}

func (s *Service) activeClient(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	if err := s.checkBuyerAccess(ctx, userID); err != nil {
		return uuid.Nil, err
	}
	clientID, err := s.storage.GetActiveClient(ctx, userID)
	if errors.Is(err, postgres.ErrClientNotFound) {
		return uuid.Nil, customErrors.ErrValidation.AddCause("field", "activeClientID")
	}
	if err != nil {
		return uuid.Nil, mapStorageError(err)
	}
	return clientID, nil
}

func (s *Service) ListFavorites(ctx context.Context, userID uuid.UUID) (response models.ListFavoritesResponse, err error) {
	clientID, err := s.activeClient(ctx, userID)
	if err != nil {
		return response, err
	}
	rows, err := s.storage.ListFavorites(ctx, userID, clientID)
	if err != nil {
		return response, mapStorageError(err)
	}
	items := make([]models.Favorite, 0, len(rows))
	for _, row := range rows {
		items = append(items, modelFavorite(row))
	}
	return models.ListFavoritesResponse{Items: items}, nil
}

func (s *Service) AddFavorite(ctx context.Context, userID uuid.UUID, productID string) (response models.AddFavoriteResponse, err error) {
	productUUID, err := uuid.Parse(strings.TrimSpace(productID))
	if err != nil || productUUID == uuid.Nil {
		return response, validation("productID")
	}
	clientID, err := s.activeClient(ctx, userID)
	if err != nil {
		return response, err
	}
	row, _, err := s.storage.AddFavorite(ctx, userID, clientID, productUUID)
	if err != nil {
		return response, mapStorageError(err)
	}
	return models.AddFavoriteResponse{Favorite: modelFavorite(row)}, nil
}

func (s *Service) DeleteFavorites(ctx context.Context, userID uuid.UUID, productIDs []string) (response models.DeleteFavoritesResponse, err error) {
	if len(productIDs) == 0 {
		return response, validation("productIDs")
	}
	ids := make([]uuid.UUID, 0, len(productIDs))
	seen := make(map[uuid.UUID]struct{}, len(productIDs))
	for _, rawID := range productIDs {
		id, parseErr := uuid.Parse(strings.TrimSpace(rawID))
		if parseErr != nil || id == uuid.Nil {
			return response, validation("productIDs")
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	clientID, err := s.activeClient(ctx, userID)
	if err != nil {
		return response, err
	}
	deleted, err := s.storage.DeleteFavorites(ctx, userID, clientID, ids)
	if err != nil {
		return response, mapStorageError(err)
	}
	return models.DeleteFavoritesResponse{Deleted: deleted}, nil
}
