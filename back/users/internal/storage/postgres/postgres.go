package postgres

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrClientNotFound   = errors.New("client not found")
	ErrFavoriteNotFound = errors.New("favorite not found")
	ErrEmailTaken       = errors.New("email already taken")
	ErrPhoneTaken       = errors.New("phone already taken")
	ErrProductNotFound  = errors.New("product not found")
	ErrInvalidSort      = errors.New("invalid sort")
)

const (
	uniqueViolationCode     = "23505"
	foreignKeyViolationCode = "23503"
)

//go:embed sql/*.sql
var queries embed.FS

func query(name string) string {
	value, err := queries.ReadFile("sql/" + name)
	if err != nil {
		panic(err)
	}
	return string(value)
}

var (
	sqlGetUserByID           = query("getUserByID.sql")
	sqlGetUserByEmail        = query("getUserByEmail.sql")
	sqlCreateUser            = query("createUser.sql")
	sqlCreateClient          = query("createClient.sql")
	sqlClientExists          = query("clientExists.sql")
	sqlLinkUserClient        = query("linkUserClient.sql")
	sqlSetInitialClient      = query("setInitialClient.sql")
	sqlListUsers             = query("listUsers.sql")
	sqlCountUsers            = query("countUsers.sql")
	sqlUpdateUser            = query("updateUser.sql")
	sqlUpdateClient          = query("updateClient.sql")
	sqlSoftDeleteUser        = query("softDeleteUser.sql")
	sqlSetUserActive         = query("setUserActive.sql")
	sqlUpdateProfile         = query("updateProfile.sql")
	sqlListUserClients       = query("listUserClients.sql")
	sqlUserHasClient         = query("userHasClient.sql")
	sqlGetClientDetails      = query("getClientDetails.sql")
	sqlGetClientConditions   = query("getClientConditions.sql")
	sqlListCategoryDiscounts = query("listCategoryDiscounts.sql")
	sqlGetActiveClient       = query("getActiveClient.sql")
	sqlLockUserClient        = query("lockUserClient.sql")
	sqlSetActiveClient       = query("setActiveClient.sql")
	sqlListFavorites         = query("listFavorites.sql")
	sqlAddFavorite           = query("addFavorite.sql")
	sqlGetFavorite           = query("getFavorite.sql")
	sqlDeleteFavorites       = query("deleteFavorites.sql")
)

type UserRow struct {
	ID             uuid.UUID
	Email          string
	Phone          string
	FirstName      string
	LastName       string
	MiddleName     string
	Role           string
	Status         string
	ClientID       uuid.NullUUID
	CompanyName    string
	INN            string
	IsActive       bool
	ActiveClientID uuid.NullUUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      sql.NullTime
}

type ClientRow struct {
	ID          uuid.UUID
	CompanyName string
	CompanyType string
	INN         string
	OGRN        string
	Address     string
	ContactName string
	Phone       string
	Email       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	IsActive    bool
}

type CategoryDiscountRow struct {
	CategoryID      uuid.UUID
	CategoryName    string
	DiscountPercent float64
	ValidFrom       sql.NullTime
	ValidTo         sql.NullTime
}

type ClientConditionsRow struct {
	ClientID          uuid.UUID
	PriceGroup        string
	CreditLimit       float64
	CreditUsed        float64
	PaymentTerms      string
	SalesContact      string
	ContactChannel    string
	CategoryDiscounts []CategoryDiscountRow
}

type FavoriteRow struct {
	UserID    uuid.UUID
	ClientID  uuid.UUID
	ProductID uuid.UUID
	CreatedAt time.Time
}

type CreateUserParams struct {
	Email       string
	Phone       string
	FirstName   string
	LastName    string
	MiddleName  string
	Status      string
	IsActive    bool
	ClientID    *uuid.UUID
	CompanyName string
	INN         string
}

type UpdateUserParams struct {
	UserID      uuid.UUID
	Email       string
	Phone       string
	FirstName   string
	LastName    string
	MiddleName  string
	Status      string
	IsActive    *bool
	ClientID    *uuid.UUID
	CompanyName string
	INN         string
}

type ListUsersParams struct {
	Q        string
	Role     string
	Status   string
	ClientID *uuid.UUID
	IsActive *bool
	Limit    int
	Offset   int
	Sort     string
}

type UpdateProfileParams struct {
	UserID     uuid.UUID
	Email      string
	Phone      string
	FirstName  string
	LastName   string
	MiddleName string
}

type Storage struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Storage {
	return &Storage{pool: pool}
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanUser(row rowScanner) (UserRow, error) {
	var user UserRow
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.FirstName,
		&user.LastName,
		&user.MiddleName,
		&user.Role,
		&user.Status,
		&user.ClientID,
		&user.CompanyName,
		&user.INN,
		&user.IsActive,
		&user.ActiveClientID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	return user, err
}

func scanClient(row rowScanner, withActive bool) (ClientRow, error) {
	var client ClientRow
	dest := []interface{}{
		&client.ID,
		&client.CompanyName,
		&client.CompanyType,
		&client.INN,
		&client.OGRN,
		&client.Address,
		&client.ContactName,
		&client.Phone,
		&client.Email,
		&client.CreatedAt,
		&client.UpdatedAt,
	}
	if withActive {
		dest = append(dest, &client.IsActive)
	}
	return client, row.Scan(dest...)
}

func classifyWriteError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	if pgErr.Code == uniqueViolationCode {
		switch {
		case strings.Contains(pgErr.ConstraintName, "email"):
			return ErrEmailTaken
		case strings.Contains(pgErr.ConstraintName, "phone"):
			return ErrPhoneTaken
		}
	}
	if pgErr.Code == foreignKeyViolationCode && strings.Contains(pgErr.ConstraintName, "product") {
		return ErrProductNotFound
	}
	return err
}

func (s *Storage) CreateUser(ctx context.Context, params CreateUserParams) (UserRow, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return UserRow{}, fmt.Errorf("begin create user: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var userID uuid.UUID
	err = tx.QueryRow(ctx, sqlCreateUser,
		params.Email,
		params.Phone,
		params.FirstName,
		params.LastName,
		params.MiddleName,
		params.Status,
		params.IsActive,
	).Scan(&userID)
	if err != nil {
		return UserRow{}, classifyWriteError(fmt.Errorf("insert user: %w", err))
	}

	clientID := params.ClientID
	if clientID == nil && (params.CompanyName != "" || params.INN != "") {
		var createdClientID uuid.UUID
		if err = tx.QueryRow(ctx, sqlCreateClient, params.CompanyName, params.INN).Scan(&createdClientID); err != nil {
			return UserRow{}, fmt.Errorf("insert client: %w", err)
		}
		clientID = &createdClientID
	}
	if clientID != nil {
		var exists bool
		if err = tx.QueryRow(ctx, sqlClientExists, *clientID).Scan(&exists); err != nil {
			return UserRow{}, fmt.Errorf("check client: %w", err)
		}
		if !exists {
			return UserRow{}, ErrClientNotFound
		}
		if _, err = tx.Exec(ctx, sqlLinkUserClient, userID, *clientID, true); err != nil {
			return UserRow{}, fmt.Errorf("link user client: %w", err)
		}
		if _, err = tx.Exec(ctx, sqlSetInitialClient, userID, *clientID); err != nil {
			return UserRow{}, fmt.Errorf("set initial client: %w", err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return UserRow{}, fmt.Errorf("commit create user: %w", err)
	}
	return s.GetUserByID(ctx, userID)
}

func (s *Storage) GetUserByID(ctx context.Context, userID uuid.UUID) (UserRow, error) {
	user, err := scanUser(s.pool.QueryRow(ctx, sqlGetUserByID, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return UserRow{}, ErrUserNotFound
	}
	if err != nil {
		return UserRow{}, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (s *Storage) GetUserByEmail(ctx context.Context, email string) (UserRow, error) {
	var userID uuid.UUID
	err := s.pool.QueryRow(ctx, sqlGetUserByEmail, email).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return UserRow{}, ErrUserNotFound
	}
	if err != nil {
		return UserRow{}, fmt.Errorf("get user by email: %w", err)
	}
	return s.GetUserByID(ctx, userID)
}

func (s *Storage) userFilter(params ListUsersParams) (string, []interface{}) {
	clauses := []string{"u.deleted_at IS NULL"}
	args := make([]interface{}, 0, 5)
	add := func(clause string, value interface{}) {
		args = append(args, value)
		clauses = append(clauses, fmt.Sprintf(clause, len(args)))
	}
	if params.Q != "" {
		add(`(
			COALESCE(u.email, '') ILIKE $%[1]d OR
			COALESCE(u.phone, '') ILIKE $%[1]d OR
			COALESCE(u.name, '') ILIKE $%[1]d OR
			COALESCE(u.surename, '') ILIKE $%[1]d
		)`, "%"+params.Q+"%")
	}
	if params.Role != "" {
		add(`EXISTS (
			SELECT 1 FROM user_roles filter_ur
			JOIN roles filter_r ON filter_r.id = filter_ur.role_id
			WHERE filter_ur.user_id = u.id
			  AND (filter_r.name = $%[1]d OR filter_r.code::text = $%[1]d)
		)`, params.Role)
	}
	if params.Status != "" {
		add("u.status = $%d", params.Status)
	}
	if params.ClientID != nil {
		add(`EXISTS (
			SELECT 1 FROM user_clients filter_uc
			WHERE filter_uc.user_id = u.id AND filter_uc.client_id = $%d
		)`, *params.ClientID)
	}
	if params.IsActive != nil {
		add("u.is_active = $%d", *params.IsActive)
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func safeSort(sort string) (string, error) {
	sortWhitelist := map[string]string{
		"":            "u.created_at DESC, u.id",
		"createdAt":   "u.created_at ASC, u.id",
		"-createdAt":  "u.created_at DESC, u.id",
		"created_at":  "u.created_at ASC, u.id",
		"-created_at": "u.created_at DESC, u.id",
		"updatedAt":   "u.updated_at ASC, u.id",
		"-updatedAt":  "u.updated_at DESC, u.id",
		"email":       "u.email ASC NULLS LAST, u.id",
		"-email":      "u.email DESC NULLS LAST, u.id",
		"lastName":    "u.surename ASC NULLS LAST, u.id",
		"-lastName":   "u.surename DESC NULLS LAST, u.id",
		"status":      "u.status ASC NULLS LAST, u.id",
		"-status":     "u.status DESC NULLS LAST, u.id",
	}
	order, ok := sortWhitelist[sort]
	if !ok {
		return "", ErrInvalidSort
	}
	return order, nil
}

func (s *Storage) ListUsers(ctx context.Context, params ListUsersParams) ([]UserRow, error) {
	order, err := safeSort(params.Sort)
	if err != nil {
		return nil, err
	}
	where, args := s.userFilter(params)
	args = append(args, params.Limit, params.Offset)
	statement := sqlListUsers + where +
		fmt.Sprintf(" ORDER BY %s LIMIT $%d OFFSET $%d", order, len(args)-1, len(args))
	rows, err := s.pool.Query(ctx, statement, args...)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()
	users := make([]UserRow, 0)
	for rows.Next() {
		user, scanErr := scanUser(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan user: %w", scanErr)
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}
	return users, nil
}

func (s *Storage) CountUsers(ctx context.Context, params ListUsersParams) (int, error) {
	where, args := s.userFilter(params)
	var total int
	if err := s.pool.QueryRow(ctx, sqlCountUsers+where, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return total, nil
}

func (s *Storage) UpdateUser(ctx context.Context, params UpdateUserParams) (UserRow, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return UserRow{}, fmt.Errorf("begin update user: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	middleSet := params.MiddleName != ""
	var updatedID uuid.UUID
	err = tx.QueryRow(ctx, sqlUpdateUser,
		params.UserID,
		params.Email,
		params.Phone,
		params.FirstName,
		params.LastName,
		params.MiddleName,
		middleSet,
		params.Status,
		params.IsActive,
	).Scan(&updatedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return UserRow{}, ErrUserNotFound
	}
	if err != nil {
		return UserRow{}, classifyWriteError(fmt.Errorf("update user: %w", err))
	}
	if params.ClientID != nil {
		var exists bool
		if err = tx.QueryRow(ctx, sqlClientExists, *params.ClientID).Scan(&exists); err != nil {
			return UserRow{}, fmt.Errorf("check client: %w", err)
		}
		if !exists {
			return UserRow{}, ErrClientNotFound
		}
		if _, err = tx.Exec(ctx, sqlLinkUserClient, params.UserID, *params.ClientID, false); err != nil {
			return UserRow{}, fmt.Errorf("link user client: %w", err)
		}
		if params.CompanyName != "" || params.INN != "" {
			if _, err = tx.Exec(ctx, sqlUpdateClient, *params.ClientID, params.CompanyName, params.INN); err != nil {
				return UserRow{}, fmt.Errorf("update client: %w", err)
			}
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return UserRow{}, fmt.Errorf("commit update user: %w", err)
	}
	return s.GetUserByID(ctx, updatedID)
}

func (s *Storage) SoftDeleteUser(ctx context.Context, userID uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, sqlSoftDeleteUser, userID)
	if err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (s *Storage) SetUserActive(ctx context.Context, userID uuid.UUID, active bool) (UserRow, error) {
	status := "inactive"
	if active {
		status = "active"
	}
	tag, err := s.pool.Exec(ctx, sqlSetUserActive, userID, active, status)
	if err != nil {
		return UserRow{}, fmt.Errorf("set user active: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return UserRow{}, ErrUserNotFound
	}
	return s.GetUserByID(ctx, userID)
}

func (s *Storage) GetProfile(ctx context.Context, userID uuid.UUID) (UserRow, *ClientRow, error) {
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return UserRow{}, nil, err
	}
	if !user.ActiveClientID.Valid {
		return user, nil, nil
	}
	client, err := s.GetClientDetails(ctx, userID, user.ActiveClientID.UUID)
	if err != nil {
		return UserRow{}, nil, err
	}
	return user, &client, nil
}

func (s *Storage) UpdateProfile(ctx context.Context, params UpdateProfileParams) (UserRow, *ClientRow, error) {
	middleSet := params.MiddleName != ""
	var userID uuid.UUID
	err := s.pool.QueryRow(ctx, sqlUpdateProfile,
		params.UserID,
		params.Email,
		params.Phone,
		params.FirstName,
		params.LastName,
		params.MiddleName,
		middleSet,
	).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return UserRow{}, nil, ErrUserNotFound
	}
	if err != nil {
		return UserRow{}, nil, classifyWriteError(fmt.Errorf("update profile: %w", err))
	}
	return s.GetProfile(ctx, userID)
}

func (s *Storage) ListUserClients(ctx context.Context, userID uuid.UUID) ([]ClientRow, error) {
	rows, err := s.pool.Query(ctx, sqlListUserClients, userID)
	if err != nil {
		return nil, fmt.Errorf("list user clients: %w", err)
	}
	defer rows.Close()
	clients := make([]ClientRow, 0)
	for rows.Next() {
		client, scanErr := scanClient(rows, true)
		if scanErr != nil {
			return nil, fmt.Errorf("scan user client: %w", scanErr)
		}
		clients = append(clients, client)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user clients: %w", err)
	}
	return clients, nil
}

func (s *Storage) UserHasClient(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (bool, error) {
	var exists bool
	if err := s.pool.QueryRow(ctx, sqlUserHasClient, userID, clientID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check user client: %w", err)
	}
	return exists, nil
}

func (s *Storage) GetClientDetails(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (ClientRow, error) {
	client, err := scanClient(s.pool.QueryRow(ctx, sqlGetClientDetails, userID, clientID), false)
	if errors.Is(err, pgx.ErrNoRows) {
		return ClientRow{}, ErrClientNotFound
	}
	if err != nil {
		return ClientRow{}, fmt.Errorf("get client details: %w", err)
	}
	return client, nil
}

func (s *Storage) GetClientConditions(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (ClientConditionsRow, error) {
	var conditions ClientConditionsRow
	err := s.pool.QueryRow(ctx, sqlGetClientConditions, userID, clientID).Scan(
		&conditions.ClientID,
		&conditions.PriceGroup,
		&conditions.CreditLimit,
		&conditions.CreditUsed,
		&conditions.PaymentTerms,
		&conditions.SalesContact,
		&conditions.ContactChannel,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ClientConditionsRow{}, ErrClientNotFound
	}
	if err != nil {
		return ClientConditionsRow{}, fmt.Errorf("get client conditions: %w", err)
	}
	rows, err := s.pool.Query(ctx, sqlListCategoryDiscounts, clientID)
	if err != nil {
		return ClientConditionsRow{}, fmt.Errorf("list category discounts: %w", err)
	}
	defer rows.Close()
	conditions.CategoryDiscounts = make([]CategoryDiscountRow, 0)
	for rows.Next() {
		var discount CategoryDiscountRow
		if err = rows.Scan(
			&discount.CategoryID,
			&discount.CategoryName,
			&discount.DiscountPercent,
			&discount.ValidFrom,
			&discount.ValidTo,
		); err != nil {
			return ClientConditionsRow{}, fmt.Errorf("scan category discount: %w", err)
		}
		conditions.CategoryDiscounts = append(conditions.CategoryDiscounts, discount)
	}
	if err = rows.Err(); err != nil {
		return ClientConditionsRow{}, fmt.Errorf("iterate category discounts: %w", err)
	}
	return conditions, nil
}

func (s *Storage) SetActiveClient(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (ClientRow, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ClientRow{}, fmt.Errorf("begin set active client: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var linkedClientID uuid.UUID
	err = tx.QueryRow(ctx, sqlLockUserClient, userID, clientID).Scan(&linkedClientID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ClientRow{}, ErrClientNotFound
	}
	if err != nil {
		return ClientRow{}, fmt.Errorf("lock user client: %w", err)
	}
	tag, err := tx.Exec(ctx, sqlSetActiveClient, userID, linkedClientID)
	if err != nil {
		return ClientRow{}, fmt.Errorf("set active client: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ClientRow{}, ErrUserNotFound
	}
	if err = tx.Commit(ctx); err != nil {
		return ClientRow{}, fmt.Errorf("commit set active client: %w", err)
	}
	client, err := s.GetClientDetails(ctx, userID, clientID)
	if err != nil {
		return ClientRow{}, err
	}
	client.IsActive = true
	return client, nil
}

func (s *Storage) GetActiveClient(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	var activeClientID uuid.NullUUID
	err := s.pool.QueryRow(ctx, sqlGetActiveClient, userID).Scan(&activeClientID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrUserNotFound
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("get active client: %w", err)
	}
	if !activeClientID.Valid {
		return uuid.Nil, ErrClientNotFound
	}
	return activeClientID.UUID, nil
}

func scanFavorite(row rowScanner) (FavoriteRow, error) {
	var favorite FavoriteRow
	err := row.Scan(&favorite.UserID, &favorite.ClientID, &favorite.ProductID, &favorite.CreatedAt)
	return favorite, err
}

func (s *Storage) ListFavorites(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) ([]FavoriteRow, error) {
	rows, err := s.pool.Query(ctx, sqlListFavorites, userID, clientID)
	if err != nil {
		return nil, fmt.Errorf("list favorites: %w", err)
	}
	defer rows.Close()
	favorites := make([]FavoriteRow, 0)
	for rows.Next() {
		favorite, scanErr := scanFavorite(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan favorite: %w", scanErr)
		}
		favorites = append(favorites, favorite)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate favorites: %w", err)
	}
	return favorites, nil
}

func (s *Storage) AddFavorite(ctx context.Context, userID uuid.UUID, clientID uuid.UUID, productID uuid.UUID) (FavoriteRow, bool, error) {
	favorite, err := scanFavorite(s.pool.QueryRow(ctx, sqlAddFavorite, userID, clientID, productID))
	if errors.Is(err, pgx.ErrNoRows) {
		existing, getErr := scanFavorite(s.pool.QueryRow(ctx, sqlGetFavorite, userID, clientID, productID))
		if getErr != nil {
			return FavoriteRow{}, false, fmt.Errorf("get existing favorite: %w", getErr)
		}
		return existing, false, nil
	}
	if err != nil {
		return FavoriteRow{}, false, classifyWriteError(fmt.Errorf("add favorite: %w", err))
	}
	return favorite, true, nil
}

func (s *Storage) DeleteFavorites(ctx context.Context, userID uuid.UUID, clientID uuid.UUID, productIDs []uuid.UUID) (int, error) {
	tag, err := s.pool.Exec(ctx, sqlDeleteFavorites, userID, clientID, productIDs)
	if err != nil {
		return 0, fmt.Errorf("delete favorites: %w", err)
	}
	return int(tag.RowsAffected()), nil
}
