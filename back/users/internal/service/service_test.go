package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	customErrors "github.com/mbatimel/AMC/users/internal/errors"
	internalModels "github.com/mbatimel/AMC/users/internal/models"
	"github.com/mbatimel/AMC/users/internal/storage/postgres"
)

type fakeStorage struct {
	createUserFn          func(context.Context, internalModels.CreateUserParams) (internalModels.User, error)
	getUserByIDFn         func(context.Context, uuid.UUID) (internalModels.User, error)
	listUsersFn           func(context.Context, internalModels.ListUsersParams) ([]internalModels.User, error)
	countUsersFn          func(context.Context, internalModels.ListUsersParams) (int, error)
	updateUserFn          func(context.Context, internalModels.UpdateUserParams) (internalModels.User, error)
	softDeleteUserFn      func(context.Context, uuid.UUID) error
	setUserActiveFn       func(context.Context, uuid.UUID, bool) (internalModels.User, error)
	getProfileFn          func(context.Context, uuid.UUID) (internalModels.User, *internalModels.Client, error)
	updateProfileFn       func(context.Context, internalModels.UpdateProfileParams) (internalModels.User, *internalModels.Client, error)
	listUserClientsFn     func(context.Context, uuid.UUID) ([]internalModels.Client, error)
	userHasClientFn       func(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	getClientDetailsFn    func(context.Context, uuid.UUID, uuid.UUID) (internalModels.Client, error)
	getClientConditionsFn func(context.Context, uuid.UUID, uuid.UUID) (internalModels.ClientConditions, error)
	setActiveClientFn     func(context.Context, uuid.UUID, uuid.UUID) (internalModels.Client, error)
	getActiveClientFn     func(context.Context, uuid.UUID) (uuid.UUID, error)
	listFavoritesFn       func(context.Context, uuid.UUID, uuid.UUID) ([]internalModels.Favorite, error)
	addFavoriteFn         func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (internalModels.Favorite, bool, error)
	deleteFavoritesFn     func(context.Context, uuid.UUID, uuid.UUID, []uuid.UUID) (int, error)
}

func (f *fakeStorage) CreateUser(ctx context.Context, params internalModels.CreateUserParams) (internalModels.User, error) {
	return f.createUserFn(ctx, params)
}
func (f *fakeStorage) GetUserByID(ctx context.Context, userID uuid.UUID) (internalModels.User, error) {
	if f.getUserByIDFn == nil {
		return internalModels.User{}, nil
	}
	return f.getUserByIDFn(ctx, userID)
}
func (f *fakeStorage) ListUsers(ctx context.Context, params internalModels.ListUsersParams) ([]internalModels.User, error) {
	return f.listUsersFn(ctx, params)
}
func (f *fakeStorage) CountUsers(ctx context.Context, params internalModels.ListUsersParams) (int, error) {
	return f.countUsersFn(ctx, params)
}
func (f *fakeStorage) UpdateUser(ctx context.Context, params internalModels.UpdateUserParams) (internalModels.User, error) {
	return f.updateUserFn(ctx, params)
}
func (f *fakeStorage) SoftDeleteUser(ctx context.Context, userID uuid.UUID) error {
	return f.softDeleteUserFn(ctx, userID)
}
func (f *fakeStorage) SetUserActive(ctx context.Context, userID uuid.UUID, active bool) (internalModels.User, error) {
	return f.setUserActiveFn(ctx, userID, active)
}
func (f *fakeStorage) GetProfile(ctx context.Context, userID uuid.UUID) (internalModels.User, *internalModels.Client, error) {
	return f.getProfileFn(ctx, userID)
}
func (f *fakeStorage) UpdateProfile(ctx context.Context, params internalModels.UpdateProfileParams) (internalModels.User, *internalModels.Client, error) {
	return f.updateProfileFn(ctx, params)
}
func (f *fakeStorage) ListUserClients(ctx context.Context, userID uuid.UUID) ([]internalModels.Client, error) {
	return f.listUserClientsFn(ctx, userID)
}
func (f *fakeStorage) UserHasClient(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (bool, error) {
	return f.userHasClientFn(ctx, userID, clientID)
}
func (f *fakeStorage) GetClientDetails(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (internalModels.Client, error) {
	return f.getClientDetailsFn(ctx, userID, clientID)
}
func (f *fakeStorage) GetClientConditions(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (internalModels.ClientConditions, error) {
	return f.getClientConditionsFn(ctx, userID, clientID)
}
func (f *fakeStorage) SetActiveClient(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (internalModels.Client, error) {
	return f.setActiveClientFn(ctx, userID, clientID)
}
func (f *fakeStorage) GetActiveClient(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	return f.getActiveClientFn(ctx, userID)
}
func (f *fakeStorage) ListFavorites(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) ([]internalModels.Favorite, error) {
	return f.listFavoritesFn(ctx, userID, clientID)
}
func (f *fakeStorage) AddFavorite(ctx context.Context, userID uuid.UUID, clientID uuid.UUID, productID uuid.UUID) (internalModels.Favorite, bool, error) {
	return f.addFavoriteFn(ctx, userID, clientID, productID)
}
func (f *fakeStorage) DeleteFavorites(ctx context.Context, userID uuid.UUID, clientID uuid.UUID, productIDs []uuid.UUID) (int, error) {
	return f.deleteFavoritesFn(ctx, userID, clientID, productIDs)
}

func testUser(userID uuid.UUID, clientID uuid.UUID) internalModels.User {
	now := time.Date(2026, 7, 29, 10, 0, 0, 0, time.UTC)
	return internalModels.User{
		ID: userID, Email: "user@example.com", Phone: "+79990000000",
		FirstName: "Иван", LastName: "Иванов", Status: "active",
		IsActive: true, ActiveClientID: uuid.NullUUID{UUID: clientID, Valid: clientID != uuid.Nil},
		CreatedAt: now, UpdatedAt: now,
	}
}

type fakeAccessClient struct {
	checkAccessFn func(context.Context, uuid.UUID, int) (bool, error)
	addRoleFn     func(context.Context, uuid.UUID, uuid.UUID, int) (bool, error)
	updateRoleFn  func(context.Context, uuid.UUID, uuid.UUID, int) (bool, error)
}

func (f *fakeAccessClient) CheckAccess(ctx context.Context, userID uuid.UUID, role int) (bool, error) {
	if f.checkAccessFn == nil {
		return true, nil
	}
	return f.checkAccessFn(ctx, userID, role)
}

func (f *fakeAccessClient) AddRole(ctx context.Context, adminUserID uuid.UUID, userID uuid.UUID, role int) (bool, error) {
	if f.addRoleFn == nil {
		return true, nil
	}
	return f.addRoleFn(ctx, adminUserID, userID, role)
}

func (f *fakeAccessClient) UpdateRole(ctx context.Context, adminUserID uuid.UUID, userID uuid.UUID, role int) (bool, error) {
	if f.updateRoleFn == nil {
		return true, nil
	}
	return f.updateRoleFn(ctx, adminUserID, userID, role)
}

func testService(storage Storage) *Service {
	return New(zerolog.Nop(), storage, &fakeAccessClient{})
}

func TestCheckBuyerAccess(t *testing.T) {
	userID := uuid.New()
	access := &fakeAccessClient{checkAccessFn: func(_ context.Context, gotUserID uuid.UUID, role int) (bool, error) {
		if gotUserID != userID {
			t.Fatalf("unexpected user id: %s", gotUserID)
		}
		if role != RoleCodeBuyer {
			t.Fatalf("role = %d, want %d", role, RoleCodeBuyer)
		}
		return true, nil
	}}

	if err := New(zerolog.Nop(), &fakeStorage{}, access).checkBuyerAccess(context.Background(), userID); err != nil {
		t.Fatalf("checkBuyerAccess() error = %v", err)
	}
}

func TestGetProfileRejectsNonBuyerBeforeStorage(t *testing.T) {
	storageCalled := false
	storage := &fakeStorage{getProfileFn: func(context.Context, uuid.UUID) (internalModels.User, *internalModels.Client, error) {
		storageCalled = true
		return internalModels.User{}, nil, nil
	}}
	access := &fakeAccessClient{checkAccessFn: func(context.Context, uuid.UUID, int) (bool, error) {
		return false, nil
	}}

	_, err := New(zerolog.Nop(), storage, access).GetProfile(context.Background(), uuid.New())
	if err == nil || !errors.Is(err, customErrors.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
	if storageCalled {
		t.Fatal("storage was called for a non-buyer")
	}
}

func TestGetProfileOwnProfile(t *testing.T) {
	userID, clientID := uuid.New(), uuid.New()
	user := testUser(userID, clientID)
	client := internalModels.Client{ID: clientID, CompanyName: "ООО Тест"}
	repo := &fakeStorage{getProfileFn: func(_ context.Context, gotUserID uuid.UUID) (internalModels.User, *internalModels.Client, error) {
		if gotUserID != userID {
			t.Fatalf("unexpected user id: %s", gotUserID)
		}
		return user, &client, nil
	}}

	response, err := testService(repo).GetProfile(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}
	if response.Profile.UserID != userID.String() || response.Profile.ActiveClientID != clientID.String() {
		t.Fatalf("unexpected profile: %+v", response.Profile)
	}
	if response.Profile.ActiveClient == nil || response.Profile.ActiveClient.CompanyName != "ООО Тест" {
		t.Fatalf("active client was not returned: %+v", response.Profile.ActiveClient)
	}
}

func TestGetProfileWithoutActiveClientStillReturnsUserFields(t *testing.T) {
	userID := uuid.New()
	user := testUser(userID, uuid.Nil)
	repo := &fakeStorage{getProfileFn: func(context.Context, uuid.UUID) (internalModels.User, *internalModels.Client, error) {
		return user, nil, nil
	}}

	response, err := testService(repo).GetProfile(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}
	if response.Profile.UserID != userID.String() || response.Profile.FirstName != "Иван" || response.Profile.Phone != "+79990000000" {
		t.Fatalf("profile = %+v", response.Profile)
	}
	if response.Profile.ActiveClient != nil || response.Profile.ActiveClientID != "" {
		t.Fatalf("unexpected active client: %+v", response.Profile.ActiveClient)
	}
}

func TestUpdateProfile(t *testing.T) {
	userID := uuid.New()
	var captured internalModels.UpdateProfileParams
	repo := &fakeStorage{updateProfileFn: func(_ context.Context, params internalModels.UpdateProfileParams) (internalModels.User, *internalModels.Client, error) {
		captured = params
		user := testUser(userID, uuid.Nil)
		user.Email = params.Email
		user.FirstName = params.FirstName
		return user, nil, nil
	}}

	response, err := testService(repo).UpdateProfile(
		context.Background(), userID, " New@Example.com ", "", "Пётр", "", "",
	)
	if err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}
	if captured.UserID != userID || captured.Email != "new@example.com" || captured.FirstName != "Пётр" {
		t.Fatalf("unexpected update params: %+v", captured)
	}
	if response.Profile.Email != "new@example.com" || response.Profile.FirstName != "Пётр" {
		t.Fatalf("unexpected response: %+v", response.Profile)
	}
}

func TestUpdateProfileRejectsNameLongerThanDatabaseColumn(t *testing.T) {
	repo := &fakeStorage{}

	_, err := testService(repo).UpdateProfile(
		context.Background(), uuid.New(), "", "", strings.Repeat("я", maxUserNameLength+1), "", "",
	)
	if err == nil || !errors.Is(err, customErrors.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestCreateUserRejectsNameLongerThanDatabaseColumn(t *testing.T) {
	repo := &fakeStorage{}

	_, err := testService(repo).CreateUser(
		context.Background(), uuid.Nil, "user@example.com", "",
		strings.Repeat("я", maxUserNameLength+1), "", "", "", "", "", "", "", true,
	)
	if err == nil || !errors.Is(err, customErrors.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestSwitchActiveClientAllowed(t *testing.T) {
	userID, clientID := uuid.New(), uuid.New()
	var switched bool
	repo := &fakeStorage{
		userHasClientFn: func(_ context.Context, gotUserID, gotClientID uuid.UUID) (bool, error) {
			return gotUserID == userID && gotClientID == clientID, nil
		},
		setActiveClientFn: func(_ context.Context, gotUserID, gotClientID uuid.UUID) (internalModels.Client, error) {
			switched = gotUserID == userID && gotClientID == clientID
			return internalModels.Client{ID: clientID, CompanyName: "ИП Тест"}, nil
		},
	}

	response, err := testService(repo).SwitchActiveClient(context.Background(), userID, clientID)
	if err != nil {
		t.Fatalf("SwitchActiveClient() error = %v", err)
	}
	if !switched || !response.ActiveClient.IsActive || response.ActiveClient.Client.ID != clientID.String() {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestSwitchActiveClientForeignForbidden(t *testing.T) {
	repo := &fakeStorage{userHasClientFn: func(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
		return false, nil
	}}

	_, err := testService(repo).SwitchActiveClient(context.Background(), uuid.New(), uuid.New())
	if err == nil || !errors.Is(err, customErrors.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestListUserClients(t *testing.T) {
	userID, firstClientID, secondClientID := uuid.New(), uuid.New(), uuid.New()
	repo := &fakeStorage{
		getUserByIDFn: func(context.Context, uuid.UUID) (internalModels.User, error) {
			return testUser(userID, firstClientID), nil
		},
		listUserClientsFn: func(_ context.Context, gotUserID uuid.UUID) ([]internalModels.Client, error) {
			if gotUserID != userID {
				t.Fatalf("unexpected user id: %s", gotUserID)
			}
			return []internalModels.Client{
				{ID: firstClientID, CompanyName: "ООО", IsActive: true},
				{ID: secondClientID, CompanyName: "ИП"},
			}, nil
		},
	}

	response, err := testService(repo).ListUserClients(context.Background(), userID)
	if err != nil {
		t.Fatalf("ListUserClients() error = %v", err)
	}
	if len(response.Items) != 2 || !response.Items[0].IsActive || response.Items[1].IsActive {
		t.Fatalf("unexpected clients: %+v", response.Items)
	}
}

func TestAddFavoriteIsIdempotent(t *testing.T) {
	userID, clientID, productID := uuid.New(), uuid.New(), uuid.New()
	createdAt := time.Now()
	calls := 0
	repo := &fakeStorage{
		getActiveClientFn: func(context.Context, uuid.UUID) (uuid.UUID, error) {
			return clientID, nil
		},
		addFavoriteFn: func(_ context.Context, gotUserID, gotClientID, gotProductID uuid.UUID) (internalModels.Favorite, bool, error) {
			calls++
			return internalModels.Favorite{
				UserID: gotUserID, ClientID: gotClientID, ProductID: gotProductID, CreatedAt: createdAt,
			}, calls == 1, nil
		},
	}
	svc := testService(repo)

	first, err := svc.AddFavorite(context.Background(), userID, productID.String())
	if err != nil {
		t.Fatalf("first AddFavorite() error = %v", err)
	}
	second, err := svc.AddFavorite(context.Background(), userID, productID.String())
	if err != nil {
		t.Fatalf("second AddFavorite() error = %v", err)
	}
	if calls != 2 || first.Favorite != second.Favorite {
		t.Fatalf("favorite is not idempotent: first=%+v second=%+v", first, second)
	}
}

func TestListFavoritesEmpty(t *testing.T) {
	userID, clientID := uuid.New(), uuid.New()
	repo := &fakeStorage{
		getActiveClientFn: func(context.Context, uuid.UUID) (uuid.UUID, error) {
			return clientID, nil
		},
		listFavoritesFn: func(_ context.Context, gotUserID, gotClientID uuid.UUID) ([]internalModels.Favorite, error) {
			if gotUserID != userID || gotClientID != clientID {
				t.Fatalf("unexpected scope: user=%s client=%s", gotUserID, gotClientID)
			}
			return nil, nil
		},
	}

	response, err := testService(repo).ListFavorites(context.Background(), userID)
	if err != nil {
		t.Fatalf("ListFavorites() error = %v", err)
	}
	if response.Items == nil || len(response.Items) != 0 {
		t.Fatalf("items = %#v, want non-nil empty list", response.Items)
	}
}

func TestListFavoritesWithData(t *testing.T) {
	userID, clientID, productID := uuid.New(), uuid.New(), uuid.New()
	createdAt := time.Now()
	repo := &fakeStorage{
		getActiveClientFn: func(context.Context, uuid.UUID) (uuid.UUID, error) {
			return clientID, nil
		},
		listFavoritesFn: func(context.Context, uuid.UUID, uuid.UUID) ([]internalModels.Favorite, error) {
			return []internalModels.Favorite{{
				UserID: userID, ClientID: clientID, ProductID: productID, CreatedAt: createdAt,
			}}, nil
		},
	}

	response, err := testService(repo).ListFavorites(context.Background(), userID)
	if err != nil {
		t.Fatalf("ListFavorites() error = %v", err)
	}
	if len(response.Items) != 1 || response.Items[0].ProductID != productID.String() || response.Items[0].CreatedAt != createdAt {
		t.Fatalf("items = %#v", response.Items)
	}
}

func TestListFavoritesStorageError(t *testing.T) {
	repo := &fakeStorage{
		getActiveClientFn: func(context.Context, uuid.UUID) (uuid.UUID, error) {
			return uuid.New(), nil
		},
		listFavoritesFn: func(context.Context, uuid.UUID, uuid.UUID) ([]internalModels.Favorite, error) {
			return nil, errors.New("database unavailable")
		},
	}

	_, err := testService(repo).ListFavorites(context.Background(), uuid.New())
	if err == nil || !errors.Is(err, customErrors.ErrInternal) {
		t.Fatalf("expected internal error, got %v", err)
	}
}

func TestDeleteFavoritesBulk(t *testing.T) {
	userID, clientID := uuid.New(), uuid.New()
	productIDs := []uuid.UUID{uuid.New(), uuid.New()}
	repo := &fakeStorage{
		getActiveClientFn: func(context.Context, uuid.UUID) (uuid.UUID, error) {
			return clientID, nil
		},
		deleteFavoritesFn: func(_ context.Context, gotUserID, gotClientID uuid.UUID, gotProductIDs []uuid.UUID) (int, error) {
			if gotUserID != userID || gotClientID != clientID {
				t.Fatalf("unexpected scope: user=%s client=%s", gotUserID, gotClientID)
			}
			if len(gotProductIDs) != 2 || gotProductIDs[0] != productIDs[0] || gotProductIDs[1] != productIDs[1] {
				t.Fatalf("unexpected product ids: %v", gotProductIDs)
			}
			return 2, nil
		},
	}

	response, err := testService(repo).DeleteFavorites(
		context.Background(), userID, []string{productIDs[0].String(), productIDs[1].String()},
	)
	if err != nil {
		t.Fatalf("DeleteFavorites() error = %v", err)
	}
	if response.Deleted != 2 {
		t.Fatalf("deleted = %d, want 2", response.Deleted)
	}
}

func TestDeleteFavoriteSingle(t *testing.T) {
	userID, clientID, productID := uuid.New(), uuid.New(), uuid.New()
	repo := &fakeStorage{
		getActiveClientFn: func(context.Context, uuid.UUID) (uuid.UUID, error) {
			return clientID, nil
		},
		deleteFavoritesFn: func(_ context.Context, gotUserID, gotClientID uuid.UUID, gotProductIDs []uuid.UUID) (int, error) {
			if gotUserID != userID || gotClientID != clientID || len(gotProductIDs) != 1 || gotProductIDs[0] != productID {
				t.Fatalf("unexpected delete: user=%s client=%s products=%v", gotUserID, gotClientID, gotProductIDs)
			}
			return 1, nil
		},
	}

	response, err := testService(repo).DeleteFavorites(context.Background(), userID, []string{productID.String()})
	if err != nil {
		t.Fatalf("DeleteFavorites() error = %v", err)
	}
	if response.Deleted != 1 {
		t.Fatalf("deleted = %d, want 1", response.Deleted)
	}
}

func TestDeleteMissingFavoriteIsIdempotent(t *testing.T) {
	repo := &fakeStorage{
		getActiveClientFn: func(context.Context, uuid.UUID) (uuid.UUID, error) {
			return uuid.New(), nil
		},
		deleteFavoritesFn: func(context.Context, uuid.UUID, uuid.UUID, []uuid.UUID) (int, error) {
			return 0, nil
		},
	}

	response, err := testService(repo).DeleteFavorites(context.Background(), uuid.New(), []string{uuid.NewString()})
	if err != nil {
		t.Fatalf("DeleteFavorites() error = %v", err)
	}
	if response.Deleted != 0 {
		t.Fatalf("deleted = %d, want 0", response.Deleted)
	}
}

func TestDeleteFavoritesCannotDeleteForeignScope(t *testing.T) {
	userID, activeClientID, productID := uuid.New(), uuid.New(), uuid.New()
	repo := &fakeStorage{
		getActiveClientFn: func(context.Context, uuid.UUID) (uuid.UUID, error) {
			return activeClientID, nil
		},
		deleteFavoritesFn: func(_ context.Context, gotUserID, gotClientID uuid.UUID, _ []uuid.UUID) (int, error) {
			if gotUserID != userID || gotClientID != activeClientID {
				t.Fatal("delete was attempted outside the current user/client scope")
			}
			return 0, nil
		},
	}

	response, err := testService(repo).DeleteFavorites(context.Background(), userID, []string{productID.String()})
	if err != nil {
		t.Fatalf("DeleteFavorites() error = %v", err)
	}
	if response.Deleted != 0 {
		t.Fatalf("foreign favorite was deleted: %+v", response)
	}
}

func TestGetUserNotFound(t *testing.T) {
	repo := &fakeStorage{getUserByIDFn: func(context.Context, uuid.UUID) (internalModels.User, error) {
		return internalModels.User{}, postgres.ErrUserNotFound
	}}

	_, err := testService(repo).GetUser(context.Background(), uuid.New())
	if err == nil || !errors.Is(err, customErrors.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestListUsersFiltersAndPagination(t *testing.T) {
	clientID := uuid.New()
	active := false
	var listParams, countParams internalModels.ListUsersParams
	repo := &fakeStorage{
		listUsersFn: func(_ context.Context, params internalModels.ListUsersParams) ([]internalModels.User, error) {
			listParams = params
			return []internalModels.User{testUser(uuid.New(), clientID)}, nil
		},
		countUsersFn: func(_ context.Context, params internalModels.ListUsersParams) (int, error) {
			countParams = params
			return 7, nil
		},
	}

	response, err := testService(repo).ListUsers(
		context.Background(), "иван", "buyer", "inactive", clientID.String(), &active, 10, 20, "-email",
	)
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	if listParams.Q != "иван" || listParams.Role != "buyer" || listParams.Status != "inactive" ||
		listParams.ClientID == nil || *listParams.ClientID != clientID || listParams.IsActive == nil ||
		*listParams.IsActive || listParams.Limit != 10 || listParams.Offset != 20 || listParams.Sort != "-email" {
		t.Fatalf("unexpected list params: %+v", listParams)
	}
	if countParams.Q != listParams.Q || response.Pagination.Total != 7 ||
		response.Pagination.Limit != 10 || response.Pagination.Offset != 20 || len(response.Items) != 1 {
		t.Fatalf("unexpected response/count params: response=%+v count=%+v", response, countParams)
	}
}

func TestActivateAndDeactivateUser(t *testing.T) {
	userID := uuid.New()
	repo := &fakeStorage{setUserActiveFn: func(_ context.Context, gotUserID uuid.UUID, active bool) (internalModels.User, error) {
		user := testUser(gotUserID, uuid.Nil)
		user.IsActive = active
		if active {
			user.Status = "active"
		} else {
			user.Status = "inactive"
		}
		return user, nil
	}}
	svc := testService(repo)

	activated, err := svc.ActivateUser(context.Background(), userID)
	if err != nil || !activated.User.IsActive || activated.User.Status != "active" {
		t.Fatalf("ActivateUser() response=%+v error=%v", activated, err)
	}
	deactivated, err := svc.DeactivateUser(context.Background(), userID)
	if err != nil || deactivated.User.IsActive || deactivated.User.Status != "inactive" {
		t.Fatalf("DeactivateUser() response=%+v error=%v", deactivated, err)
	}
}
