package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/mbatimel/AMC/auth/internal/storage/postgres"
)

type stubStorage struct {
	createIPUserCalled bool
}

func (s *stubStorage) GetUserByEmail(ctx context.Context, email string) (postgres.User, error) {
	return postgres.User{}, postgres.ErrUserNotFound
}

func (s *stubStorage) GetUserByID(ctx context.Context, userID uuid.UUID) (postgres.User, error) {
	return postgres.User{}, postgres.ErrUserNotFound
}

func (s *stubStorage) CreateIPUser(
	ctx context.Context,
	email, passwordHash string,
	fullName, shortName, inn, kpp, ogrn, okved, taxSystem, legalAddress, actualAddress,
	directorFullName, directorPosition, phone, additionalPhone, website,
	bankAccount, bankName, bankBik, correspondentAccount *string,
	roleCode int,
) (uuid.UUID, error) {
	s.createIPUserCalled = true
	return uuid.New(), nil
}

func (s *stubStorage) CreateIndividualUser(
	ctx context.Context,
	email, passwordHash, surename, name, middleName, phone, city, deliveryAddress string,
	inn *string,
	roleCode int,
) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (s *stubStorage) UpdateUserPassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	return nil
}

type stubFnsClient struct {
	valid   bool
	err     error
	calls   int
	lastINN string
}

func (f *stubFnsClient) CheckIndividual(ctx context.Context, inn string) (bool, error) {
	f.calls++
	f.lastINN = inn
	return f.valid, f.err
}

const validINN = "773208978609" // passes local checksum (validate12)

func registerIPArgs(inn *string) []*string {
	// fullName, shortName, inn, kpp, ogrn, okved, taxSystem, legalAddress, actualAddress,
	// directorFullName, directorPosition, phone, additionalPhone, website,
	// bankAccount, bankName, bankBik, correspondentAccount
	empty := ""
	return []*string{
		&empty, &empty, inn, &empty, &empty, &empty, &empty, &empty, &empty,
		&empty, &empty, &empty, &empty, &empty,
		&empty, &empty, &empty, &empty,
	}
}

func TestRegisterIP_FnsValid_CreatesUser(t *testing.T) {
	storage := &stubStorage{}
	fnsClient := &stubFnsClient{valid: true}
	svc := &service{logger: zerolog.Nop(), storage: storage, fnsClient: fnsClient}

	inn := validINN
	args := registerIPArgs(&inn)
	_, err := svc.RegisterIP(context.Background(), "user@example.com", "password",
		args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8],
		args[9], args[10], args[11], args[12], args[13],
		args[14], args[15], args[16], args[17],
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !storage.createIPUserCalled {
		t.Fatal("expected CreateIPUser to be called")
	}
	if fnsClient.calls != 1 || fnsClient.lastINN != validINN {
		t.Fatalf("expected fns client called once with %q, got calls=%d inn=%q", validINN, fnsClient.calls, fnsClient.lastINN)
	}
}

func TestRegisterIP_FnsInvalid_RejectsRegistration(t *testing.T) {
	storage := &stubStorage{}
	fnsClient := &stubFnsClient{valid: false}
	svc := &service{logger: zerolog.Nop(), storage: storage, fnsClient: fnsClient}

	inn := validINN
	args := registerIPArgs(&inn)
	_, err := svc.RegisterIP(context.Background(), "user@example.com", "password",
		args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8],
		args[9], args[10], args[11], args[12], args[13],
		args[14], args[15], args[16], args[17],
	)
	if err == nil {
		t.Fatal("expected error")
	}
	if storage.createIPUserCalled {
		t.Fatal("CreateIPUser must not be called when fns marks inn invalid")
	}
}

func TestRegisterIP_FnsError_FailsClosed(t *testing.T) {
	storage := &stubStorage{}
	fnsClient := &stubFnsClient{err: errors.New("timeout")}
	svc := &service{logger: zerolog.Nop(), storage: storage, fnsClient: fnsClient}

	inn := validINN
	args := registerIPArgs(&inn)
	_, err := svc.RegisterIP(context.Background(), "user@example.com", "password",
		args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8],
		args[9], args[10], args[11], args[12], args[13],
		args[14], args[15], args[16], args[17],
	)
	if err == nil {
		t.Fatal("expected error")
	}
	if storage.createIPUserCalled {
		t.Fatal("CreateIPUser must not be called when fns check errors")
	}
}

func TestRegisterIP_LocalChecksumInvalid_NeverCallsFns(t *testing.T) {
	storage := &stubStorage{}
	fnsClient := &stubFnsClient{valid: true}
	svc := &service{logger: zerolog.Nop(), storage: storage, fnsClient: fnsClient}

	badINN := "773208978608" // validINN with last digit flipped — fails validate12 checksum
	args := registerIPArgs(&badINN)
	_, err := svc.RegisterIP(context.Background(), "user@example.com", "password",
		args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8],
		args[9], args[10], args[11], args[12], args[13],
		args[14], args[15], args[16], args[17],
	)
	if err == nil {
		t.Fatal("expected error")
	}
	if fnsClient.calls != 0 {
		t.Fatal("fns client must not be called when local checksum is already invalid")
	}
}
