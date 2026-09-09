package service

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/mbatimel/AMC/auth/internal/storage/postgres"
)

type stubStorage struct {
	createIPUserCalled bool
	surename           string
	name               string
	middleName         string
	phone              *string
	fileURL            string
	fileName           string
	innExists          bool
	innExistsErr       error
}

func (s *stubStorage) GetUserByEmail(ctx context.Context, email string) (postgres.User, error) {
	return postgres.User{}, postgres.ErrUserNotFound
}

func (s *stubStorage) GetUserByID(ctx context.Context, userID uuid.UUID) (postgres.User, error) {
	return postgres.User{}, postgres.ErrUserNotFound
}

func (s *stubStorage) CounterpartyINNExists(ctx context.Context, inn string) (bool, error) {
	return s.innExists, s.innExistsErr
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
	s.surename, s.name, s.middleName, s.phone = surename, name, middleName, phone
	s.fileURL, s.fileName = requisitesFileURL, requisitesFileName
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

func TestRegisterIP_FnsInvalid_RejectsRegistration(t *testing.T) {
	storage := &stubStorage{}
	fnsClient := &stubFnsClient{valid: false}
	svc := newTestService(storage, fnsClient)

	_, err := svc.RegisterIP(context.Background(), "user@example.com", "password",
		"ИП Иванов", validINN, "Иванов Иван Иванович", "+79990000000", validRequisitesFile())
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
	svc := newTestService(storage, fnsClient)

	_, err := svc.RegisterIP(context.Background(), "user@example.com", "password",
		"ИП Иванов", validINN, "Иванов Иван Иванович", "+79990000000", validRequisitesFile())
	if err == nil {
		t.Fatal("expected error")
	}
	if storage.createIPUserCalled {
		t.Fatal("CreateIPUser must not be called when fns check errors")
	}
}

func TestRegisterIP_InnAlreadyTaken_RejectsRegistration(t *testing.T) {
	storage := &stubStorage{innExists: true}
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
