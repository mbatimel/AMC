package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	customErrors "github.com/mbatimel/AMC/auth/internal/errors"
	"github.com/mbatimel/AMC/auth/internal/storage/postgres"
)

type loginStorageStub struct {
	getUserByEmailCalled bool
}

func (s *loginStorageStub) GetUserByEmail(context.Context, string) (postgres.User, error) {
	s.getUserByEmailCalled = true
	return postgres.User{}, postgres.ErrUserNotFound
}

func (*loginStorageStub) GetUserByID(context.Context, uuid.UUID) (postgres.User, error) {
	return postgres.User{}, postgres.ErrUserNotFound
}

func (*loginStorageStub) CreateIPUser(
	context.Context,
	string, string,
	*string, *string, *string, *string, *string, *string, *string, *string, *string,
	*string, *string, *string, *string, *string,
	*string, *string, *string, *string,
	int,
) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (*loginStorageStub) CreateIndividualUser(
	context.Context,
	string, string, string, string, string, string, string, string,
	*string,
	int,
) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (*loginStorageStub) UpdateUserPassword(context.Context, uuid.UUID, string) error {
	return nil
}

type accessClientStub struct{}

func (accessClientStub) CheckAccess(context.Context, uuid.UUID, int) (bool, error) {
	return true, nil
}

func TestLoginUserRejectsMissingCredentialsBeforeStorageLookup(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		email    string
		password string
	}{
		{name: "empty email", password: "password"},
		{name: "blank email", email: "   ", password: "password"},
		{name: "empty password", email: "user@example.com"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			storage := &loginStorageStub{}
			svc := NewAuthApiService(zerolog.Nop(), storage, accessClientStub{})

			userID, err := svc.LoginUser(context.Background(), testCase.email, testCase.password)

			if userID != uuid.Nil {
				t.Fatalf("userID = %s", userID)
			}
			if err == nil || !customErrors.Is(err, customErrors.InvalidCredentialsError()) {
				t.Fatalf("error = %v", err)
			}
			if storage.getUserByEmailCalled {
				t.Fatal("storage lookup must not happen for missing credentials")
			}
		})
	}
}
