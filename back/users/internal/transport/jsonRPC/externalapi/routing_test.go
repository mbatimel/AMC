package externalapi

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	externalAPI "github.com/mbatimel/AMC/users/pkg/interfaces/externalAPI"
	"github.com/mbatimel/AMC/users/pkg/models"
)

// routingFake implements externalAPI.UsersAPI, recording which method the
// router actually dispatched to. Everything but GetUser/GetProfile is a
// bare no-op — this test only cares about routing, not business logic.
type routingFake struct {
	getUserCalled    bool
	getProfileCalled bool
}

func (f *routingFake) CreateUser(context.Context, uuid.UUID, string, string, string, string, string, string, string, string, string, string, bool) (models.CreateUserResponse, error) {
	return models.CreateUserResponse{}, nil
}
func (f *routingFake) GetProfile(context.Context, uuid.UUID) (models.GetProfileResponse, error) {
	f.getProfileCalled = true
	return models.GetProfileResponse{}, nil
}
func (f *routingFake) UpdateProfile(context.Context, uuid.UUID, string, string, string, string, string) (models.UpdateProfileResponse, error) {
	return models.UpdateProfileResponse{}, nil
}
func (f *routingFake) ListUserClients(context.Context, uuid.UUID) (models.ListUserClientsResponse, error) {
	return models.ListUserClientsResponse{}, nil
}
func (f *routingFake) GetClientDetails(context.Context, uuid.UUID, uuid.UUID) (models.GetClientDetailsResponse, error) {
	return models.GetClientDetailsResponse{}, nil
}
func (f *routingFake) GetClientConditions(context.Context, uuid.UUID, uuid.UUID) (models.GetClientConditionsResponse, error) {
	return models.GetClientConditionsResponse{}, nil
}
func (f *routingFake) SwitchActiveClient(context.Context, uuid.UUID, uuid.UUID) (models.SwitchActiveClientResponse, error) {
	return models.SwitchActiveClientResponse{}, nil
}
func (f *routingFake) ListFavorites(context.Context, uuid.UUID) (models.ListFavoritesResponse, error) {
	return models.ListFavoritesResponse{}, nil
}
func (f *routingFake) AddFavorite(context.Context, uuid.UUID, string) (models.AddFavoriteResponse, error) {
	return models.AddFavoriteResponse{}, nil
}
func (f *routingFake) DeleteFavorites(context.Context, uuid.UUID, []string) (models.DeleteFavoritesResponse, error) {
	return models.DeleteFavoritesResponse{}, nil
}
func (f *routingFake) GetUser(context.Context, uuid.UUID) (models.GetUserResponse, error) {
	f.getUserCalled = true
	return models.GetUserResponse{}, nil
}
func (f *routingFake) ListUsers(context.Context, string, string, string, string, *bool, int, int, string) (models.ListUsersResponse, error) {
	return models.ListUsersResponse{}, nil
}
func (f *routingFake) UpdateUser(context.Context, uuid.UUID, uuid.UUID, string, string, string, string, string, string, string, string, string, string, *bool) (models.UpdateUserResponse, error) {
	return models.UpdateUserResponse{}, nil
}
func (f *routingFake) DeleteUser(context.Context, uuid.UUID) (models.DeleteUserResponse, error) {
	return models.DeleteUserResponse{}, nil
}
func (f *routingFake) DeleteUserByEmail(context.Context, string) (models.DeleteUserResponse, error) {
	return models.DeleteUserResponse{}, nil
}
func (f *routingFake) ActivateUser(context.Context, uuid.UUID) (models.ActivateUserResponse, error) {
	return models.ActivateUserResponse{}, nil
}
func (f *routingFake) DeactivateUser(context.Context, uuid.UUID, string, string, string, string) (models.DeactivateUserResponse, error) {
	return models.DeactivateUserResponse{}, nil
}

var _ externalAPI.UsersAPI = (*routingFake)(nil)

// TestProfileRouteDoesNotShadowUserIDRoute is a regression test for a real
// production incident: /api/v1/users/:userID was registered before
// /api/v1/users/profile, so Fiber matched "profile" as the :userID param
// and GetProfile never ran. See usersapi-http.go's SetRoutes comment.
func TestProfileRouteDoesNotShadowUserIDRoute(t *testing.T) {
	fake := &routingFake{}
	srv := NewUsersAPI(fake)
	app := fiber.New()
	srv.SetRoutes(app)

	req := httptest.NewRequest("GET", "/api/v1/users/profile", nil)
	req.Header.Set("X-User-Id", uuid.New().String())

	if _, err := app.Test(req); err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}

	if !fake.getProfileCalled {
		t.Fatal("GET /api/v1/users/profile did not reach GetProfile")
	}
	if fake.getUserCalled {
		t.Fatal("GET /api/v1/users/profile was shadowed by /api/v1/users/:userID (GetUser ran instead)")
	}
}
