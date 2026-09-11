package service

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	customErrors "github.com/mbatimel/AMC/auth/internal/errors"
	"github.com/mbatimel/AMC/auth/internal/storage/postgres"
)

type stubStorage struct {
	createIPUserCalled                  bool
	surename                            string
	name                                string
	middleName                          string
	phone                               *string
	fileURL                             string
	fileName                            string
	innInUse                            bool
	innInUseErr                         error
	createIPUserErr                     error
	email                               string
	inn                                 *string
	getUserByEmailFn                    func(context.Context, string) (postgres.User, error)
	createPasswordResetTokenFn          func(context.Context, uuid.UUID, string, time.Time) error
	invalidateUserPasswordResetTokensFn func(context.Context, uuid.UUID) error
	getPasswordResetTokenFn             func(context.Context, string) (postgres.PasswordResetToken, error)
	updateUserPasswordFn                func(context.Context, uuid.UUID, string) error
}

func (s *stubStorage) GetUserByEmail(ctx context.Context, email string) (postgres.User, error) {
	if s.getUserByEmailFn != nil {
		return s.getUserByEmailFn(ctx, email)
	}
	return postgres.User{}, postgres.ErrUserNotFound
}

func (s *stubStorage) GetUserByID(ctx context.Context, userID uuid.UUID) (postgres.User, error) {
	return postgres.User{}, postgres.ErrUserNotFound
}

func (s *stubStorage) CounterpartyINNInUse(ctx context.Context, inn string) (bool, error) {
	return s.innInUse, s.innInUseErr
}

func (s *stubStorage) CreateIPUser(
	ctx context.Context,
	email, passwordHash string,
	surename, name, middleName string,
	fullName, shortName, inn, kpp, ogrn, okved, taxSystem, legalAddress, actualAddress,
	directorFullName, directorPosition, phone, additionalPhone, website,
	bankAccount, bankName, bankBik, correspondentAccount *string,
	requisitesFileURL, requisitesFileName string,
	roleCode int,
) (uuid.UUID, error) {
	s.createIPUserCalled = true
	s.email = email
	s.inn = inn
	s.surename, s.name, s.middleName, s.phone = surename, name, middleName, phone
	s.fileURL, s.fileName = requisitesFileURL, requisitesFileName
	if s.createIPUserErr != nil {
		return uuid.Nil, s.createIPUserErr
	}
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
	if s.updateUserPasswordFn != nil {
		return s.updateUserPasswordFn(ctx, userID, passwordHash)
	}
	return nil
}

func (s *stubStorage) CreatePasswordResetToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	if s.createPasswordResetTokenFn != nil {
		return s.createPasswordResetTokenFn(ctx, userID, tokenHash, expiresAt)
	}
	return nil
}

func (s *stubStorage) InvalidateUserPasswordResetTokens(ctx context.Context, userID uuid.UUID) error {
	if s.invalidateUserPasswordResetTokensFn != nil {
		return s.invalidateUserPasswordResetTokensFn(ctx, userID)
	}
	return nil
}

func (s *stubStorage) GetPasswordResetToken(ctx context.Context, tokenHash string) (postgres.PasswordResetToken, error) {
	if s.getPasswordResetTokenFn != nil {
		return s.getPasswordResetTokenFn(ctx, tokenHash)
	}
	return postgres.PasswordResetToken{}, postgres.ErrTokenNotFound
}

type stubObjectStorage struct {
	uploadCalled bool
	deleteCalled bool
	uploadErr    error
}

func (o *stubObjectStorage) Upload(ctx context.Context, objectKey string, body io.Reader, size int64, contentType string) error {
	o.uploadCalled = true
	return o.uploadErr
}

func (o *stubObjectStorage) Delete(ctx context.Context, objectKey string) error {
	o.deleteCalled = true
	return nil
}

func (o *stubObjectStorage) URL(objectKey string) string {
	return "https://s3.example.com/" + objectKey
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

func validRequisitesFile() RequisitesFile {
	return RequisitesFile{FileName: "requisites.pdf", Content: []byte("%PDF-1.4 test")}
}

func newTestService(storage Storage, fnsClient FnsClient) *service {
	return &service{
		logger:        zerolog.Nop(),
		storage:       storage,
		fnsClient:     fnsClient,
		objectStorage: &stubObjectStorage{},
		maxFileSize:   10 << 20,
	}
}

func TestRegisterIP_FnsValid_CreatesUser(t *testing.T) {
	storage := &stubStorage{}
	fnsClient := &stubFnsClient{valid: true}
	svc := newTestService(storage, fnsClient)

	_, err := svc.RegisterIP(context.Background(), "user@example.com", "password",
		"ИП Иванов", validINN, "Иванов Иван Иванович", "+79990000000", validRequisitesFile())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !storage.createIPUserCalled {
		t.Fatal("expected CreateIPUser to be called")
	}
	if fnsClient.calls != 1 || fnsClient.lastINN != validINN {
		t.Fatalf("expected fns client called once with %q, got calls=%d inn=%q", validINN, fnsClient.calls, fnsClient.lastINN)
	}
	if storage.fileURL == "" || storage.fileName != "requisites.pdf" {
		t.Fatalf("expected requisites file to be persisted, got url=%q name=%q", storage.fileURL, storage.fileName)
	}
}

func TestRegisterIP_PersistsDirectorNameAndPhone(t *testing.T) {
	storage := &stubStorage{}
	svc := newTestService(storage, &stubFnsClient{valid: true})

	director := "Иванов Иван Иванович"
	phone := "+7 999 000-00-00"

	_, err := svc.RegisterIP(context.Background(), "user@example.com", "password",
		"ИП Иванов", validINN, director, phone, validRequisitesFile())
	if err != nil {
		t.Fatalf("RegisterIP() error = %v", err)
	}
	if storage.surename != "Иванов" || storage.name != "Иван" || storage.middleName != "Иванович" {
		t.Fatalf("stored name = %q %q %q", storage.surename, storage.name, storage.middleName)
	}
	if storage.phone == nil || *storage.phone != phone {
		t.Fatalf("stored phone = %v, want %q", storage.phone, phone)
	}
}

func TestRegisterIP_NormalizesRegistrationIdentifiers(t *testing.T) {
	storage := &stubStorage{}
	fnsClient := &stubFnsClient{valid: true}
	svc := newTestService(storage, fnsClient)

	_, err := svc.RegisterIP(context.Background(), " User@Example.COM ", "password",
		"ИП Иванов", "773 208-978609", "Иванов Иван Иванович", " +79990000000 ", validRequisitesFile())
	if err != nil {
		t.Fatalf("RegisterIP() error = %v", err)
	}
	if storage.email != "user@example.com" {
		t.Fatalf("stored email = %q", storage.email)
	}
	if storage.inn == nil || *storage.inn != validINN || fnsClient.lastINN != validINN {
		t.Fatalf("stored INN = %v, FNS INN = %q, want %q", storage.inn, fnsClient.lastINN, validINN)
	}
	if storage.phone == nil || *storage.phone != "+79990000000" {
		t.Fatalf("stored phone = %v", storage.phone)
	}
}

func TestRegisterIP_InvalidEmailReturnsValidation(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{name: "malformed", email: "not-an-email"},
		{name: "too long", email: strings.Repeat("a", maxEmailLength) + "@example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &stubStorage{}
			fnsClient := &stubFnsClient{valid: true}
			svc := newTestService(storage, fnsClient)

			_, err := svc.RegisterIP(context.Background(), tt.email, "password",
				"ИП Иванов", validINN, "Иванов Иван Иванович", "+79990000000", validRequisitesFile())
			var validationErr *customErrors.Error
			if !errors.As(err, &validationErr) || validationErr.Code() != 400 || validationErr.Cause["field"] != "email" {
				t.Fatalf("RegisterIP() error = %#v, want email validation HTTP 400", err)
			}
			if storage.createIPUserCalled || fnsClient.calls != 0 {
				t.Fatal("invalid email must fail before FNS and storage")
			}
		})
	}
}

func TestRegisterIP_InvalidINNReturnsValidation(t *testing.T) {
	tests := []struct {
		name string
		inn  string
	}{
		{name: "empty", inn: " - "},
		{name: "invalid length", inn: "123"},
		{name: "letters", inn: "77320897860A"},
		{name: "checksum", inn: "773208978608"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &stubStorage{}
			fnsClient := &stubFnsClient{valid: true}
			svc := newTestService(storage, fnsClient)

			_, err := svc.RegisterIP(context.Background(), "user@example.com", "password",
				"ИП Иванов", tt.inn, "Иванов Иван Иванович", "+79990000000", validRequisitesFile())
			var validationErr *customErrors.Error
			if !errors.As(err, &validationErr) || validationErr.Code() != 400 {
				t.Fatalf("RegisterIP() error = %#v, want INN validation HTTP 400", err)
			}
			if storage.createIPUserCalled || fnsClient.calls != 0 {
				t.Fatal("invalid INN must fail before FNS and storage")
			}
		})
	}
}

func TestRegisterIP_MapsStorageConflicts(t *testing.T) {
	tests := []struct {
		name       string
		storageErr error
		status     int
		errorText  string
	}{
		{name: "email", storageErr: postgres.ErrEmailTaken, status: 409, errorText: "email already registered"},
		{name: "phone", storageErr: postgres.ErrPhoneTaken, status: 409, errorText: "phone already registered"},
		{name: "inn", storageErr: postgres.ErrInnTaken, status: 409, errorText: "inn already registered"},
		{name: "database", storageErr: errors.New("database unavailable"), status: 500, errorText: "internal server error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &stubStorage{createIPUserErr: tt.storageErr}
			objectStorage := &stubObjectStorage{}
			svc := newTestService(storage, &stubFnsClient{valid: true})
			svc.objectStorage = objectStorage

			_, err := svc.RegisterIP(context.Background(), "user@example.com", "password",
				"ИП Иванов", validINN, "Иванов Иван Иванович", "+79990000000", validRequisitesFile())
			var mapped *customErrors.Error
			if !errors.As(err, &mapped) || mapped.Code() != tt.status || mapped.ErrorText != tt.errorText {
				t.Fatalf("RegisterIP() error = %#v, want status=%d text=%q", err, tt.status, tt.errorText)
			}
			if !objectStorage.deleteCalled {
				t.Fatal("failed transaction must compensate the uploaded S3 object")
			}
		})
	}
}

func TestRegisterIP_FnsInvalid_RejectsRegistration(t *testing.T) {
	storage := &stubStorage{}
	fnsClient := &stubFnsClient{valid: false}
	svc := newTestService(storage, fnsClient)

	_, err := svc.RegisterIP(context.Background(), "user@example.com", "password",
		"ИП Иванов", validINN, "Иванов Иван Иванович", "+79990000000", validRequisitesFile())
	var mapped *customErrors.Error
	if !errors.As(err, &mapped) || mapped.Code() != 400 || mapped.ErrorText != "inn is invalid" {
		t.Fatalf("RegisterIP() error = %#v, want invalid INN HTTP 400", err)
	}
	if storage.createIPUserCalled {
		t.Fatal("CreateIPUser must not be called when fns marks inn invalid")
	}
}

func TestRegisterIP_FnsError_FailsClosed(t *testing.T) {
	storage := &stubStorage{}
	fnsClient := &stubFnsClient{err: errors.New("timeout")}
	svc := newTestService(storage, fnsClient)

	_, err := svc.RegisterIP(context.Background(), "user@example.com", "password",
		"ИП Иванов", validINN, "Иванов Иван Иванович", "+79990000000", validRequisitesFile())
	var mapped *customErrors.Error
	if !errors.As(err, &mapped) || mapped.Code() != 500 || mapped.ErrorText != "internal server error" {
		t.Fatalf("RegisterIP() error = %#v, want external service HTTP 500", err)
	}
	if storage.createIPUserCalled {
		t.Fatal("CreateIPUser must not be called when fns check errors")
	}
}

func TestRegisterIP_InnAlreadyTaken_RejectsRegistration(t *testing.T) {
	storage := &stubStorage{innInUse: true}
	fnsClient := &stubFnsClient{valid: true}
	svc := newTestService(storage, fnsClient)

	_, err := svc.RegisterIP(context.Background(), "user@example.com", "password",
		"ИП Иванов", validINN, "Иванов Иван Иванович", "+79990000000", validRequisitesFile())
	if err == nil {
		t.Fatal("expected error")
	}
	if storage.createIPUserCalled {
		t.Fatal("CreateIPUser must not be called when inn already exists")
	}
	if fnsClient.calls != 0 {
		t.Fatal("fns client must not be called when inn already exists in storage")
	}
}

func TestRegisterIP_LocalChecksumInvalid_NeverCallsFns(t *testing.T) {
	storage := &stubStorage{}
	fnsClient := &stubFnsClient{valid: true}
	svc := newTestService(storage, fnsClient)

	badINN := "773208978608" // validINN with last digit flipped — fails validate12 checksum

	_, err := svc.RegisterIP(context.Background(), "user@example.com", "password",
		"ИП Иванов", badINN, "Иванов Иван Иванович", "+79990000000", validRequisitesFile())
	if err == nil {
		t.Fatal("expected error")
	}
	if fnsClient.calls != 0 {
		t.Fatal("fns client must not be called when local checksum is already invalid")
	}
}

func TestRegisterIP_MissingFile_RejectsRegistration(t *testing.T) {
	storage := &stubStorage{}
	fnsClient := &stubFnsClient{valid: true}
	svc := newTestService(storage, fnsClient)

	_, err := svc.RegisterIP(context.Background(), "user@example.com", "password",
		"ИП Иванов", validINN, "Иванов Иван Иванович", "+79990000000", RequisitesFile{})
	if err == nil {
		t.Fatal("expected error")
	}
	if storage.createIPUserCalled {
		t.Fatal("CreateIPUser must not be called when requisites file is missing")
	}
}

func TestRegisterIP_InvalidFileType_RejectsRegistration(t *testing.T) {
	storage := &stubStorage{}
	fnsClient := &stubFnsClient{valid: true}
	svc := newTestService(storage, fnsClient)

	badFile := RequisitesFile{FileName: "requisites.exe", Content: []byte("MZ")}
	_, err := svc.RegisterIP(context.Background(), "user@example.com", "password",
		"ИП Иванов", validINN, "Иванов Иван Иванович", "+79990000000", badFile)
	if err == nil {
		t.Fatal("expected error")
	}
	if storage.createIPUserCalled {
		t.Fatal("CreateIPUser must not be called when file type is invalid")
	}
}
