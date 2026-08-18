package service

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	customErrors "github.com/mbatimel/AMC/admin/internal/errors"
	"github.com/mbatimel/AMC/admin/internal/storage/postgres"
)

func certificatePDFBase64() string {
	return base64.StdEncoding.EncodeToString([]byte("%PDF-1.4\n%test"))
}

func newCertificateService(storage *fakeStorage, objects *fakeBannerObjectStorage) *service {
	return NewAdminApiService(zerolog.Nop(), storage, &fakeAuthClient{}, &fakeAccessClient{allowed: true}, nil, nil, WithObjectStorage(objects, 4096))
}

func TestCreateCertificateWithFile(t *testing.T) {
	var saved postgres.Certificate
	storage := &fakeStorage{createCertificateFn: func(_ context.Context, cert postgres.Certificate) (postgres.Certificate, error) {
		saved = cert
		return cert, nil
	}}
	objects := &fakeBannerObjectStorage{}
	result, err := newCertificateService(storage, objects).CreateCertificate(context.Background(), uuid.New(),
		"ISO 9001", 1, true, "iso9001.pdf", certificatePDFBase64())
	if err != nil {
		t.Fatalf("CreateCertificate() error = %v", err)
	}
	if len(objects.uploaded) != 1 || saved.ObjectKey != objects.uploaded[0] {
		t.Fatalf("uploaded=%v saved=%+v", objects.uploaded, saved)
	}
	if result.FileURL == "" {
		t.Fatalf("result = %+v, want a file URL", result)
	}
}

func TestCreateCertificate_RequiresFile(t *testing.T) {
	storage := &fakeStorage{}
	_, err := newCertificateService(storage, &fakeBannerObjectStorage{}).CreateCertificate(context.Background(), uuid.New(),
		"ISO 9001", 0, true, "", "")
	if err == nil {
		t.Fatal("CreateCertificate() error = nil, want bad request when file is missing")
	}
}

func TestUpdateCertificate_ReplacesFileAndDeletesOld(t *testing.T) {
	certID := uuid.New()
	oldKey := "certificates/" + certID.String() + "/old.pdf"
	storage := &fakeStorage{
		getCertificateFn: func(context.Context, uuid.UUID) (postgres.Certificate, error) {
			return postgres.Certificate{ID: certID, FileID: uuid.New(), ObjectKey: oldKey, Title: "ISO 9001"}, nil
		},
		updateCertificateFn: func(_ context.Context, cert postgres.Certificate, replaceFile bool) (postgres.Certificate, error) {
			if !replaceFile || cert.ObjectKey == oldKey {
				t.Fatalf("replaceFile=%v cert=%+v", replaceFile, cert)
			}
			return cert, nil
		},
	}
	objects := &fakeBannerObjectStorage{}
	result, err := newCertificateService(storage, objects).UpdateCertificate(context.Background(), uuid.New(), certID,
		"ISO 9001:2015", 0, true, "iso9001.pdf", certificatePDFBase64())
	if err != nil || result.FileURL == "" {
		t.Fatalf("result=%+v error=%v", result, err)
	}
	if len(objects.deleted) != 1 || objects.deleted[0] != oldKey {
		t.Fatalf("deleted = %v", objects.deleted)
	}
}

func TestUpdateCertificate_WithoutFileKeepsCurrent(t *testing.T) {
	certID := uuid.New()
	currentKey := "certificates/" + certID.String() + "/current.pdf"
	storage := &fakeStorage{
		getCertificateFn: func(context.Context, uuid.UUID) (postgres.Certificate, error) {
			return postgres.Certificate{ID: certID, ObjectKey: currentKey, Title: "ISO 9001"}, nil
		},
		updateCertificateFn: func(_ context.Context, cert postgres.Certificate, replaceFile bool) (postgres.Certificate, error) {
			if replaceFile {
				t.Fatalf("replaceFile=%v, want false when no file supplied", replaceFile)
			}
			cert.ObjectKey = currentKey
			return cert, nil
		},
	}
	objects := &fakeBannerObjectStorage{}
	_, err := newCertificateService(storage, objects).UpdateCertificate(context.Background(), uuid.New(), certID,
		"ISO 9001 (rename)", 2, true, "", "")
	if err != nil || len(objects.deleted) != 0 {
		t.Fatalf("error=%v deleted=%v", err, objects.deleted)
	}
}

func TestDeleteCertificate(t *testing.T) {
	certID := uuid.New()
	objectKey := "certificates/" + certID.String() + "/cert.pdf"
	storage := &fakeStorage{getCertificateFn: func(context.Context, uuid.UUID) (postgres.Certificate, error) {
		return postgres.Certificate{ID: certID, ObjectKey: objectKey}, nil
	}}
	objects := &fakeBannerObjectStorage{}
	result, err := newCertificateService(storage, objects).DeleteCertificate(context.Background(), uuid.New(), certID)
	if err != nil || !result.Deleted || len(objects.deleted) != 1 || objects.deleted[0] != objectKey {
		t.Fatalf("result=%+v error=%v deleted=%v", result, err, objects.deleted)
	}
}

func TestListPublicCertificates_OnlyVisible(t *testing.T) {
	storage := &fakeStorage{listCertificatesFn: func(_ context.Context, visibleOnly bool) ([]postgres.Certificate, error) {
		if !visibleOnly {
			t.Fatal("public list unexpectedly requested all certificates")
		}
		return []postgres.Certificate{{ID: uuid.New(), Title: "ISO 9001"}}, nil
	}}
	result, err := newCertificateService(storage, &fakeBannerObjectStorage{}).ListPublicCertificates(context.Background())
	if err != nil || len(result.Items) != 1 {
		t.Fatalf("result=%+v error=%v", result, err)
	}
}

func TestUpdateCertificate_NotFoundOnMissing(t *testing.T) {
	storage := &fakeStorage{getCertificateFn: func(context.Context, uuid.UUID) (postgres.Certificate, error) {
		return postgres.Certificate{}, postgres.ErrCertificateNotFound
	}}
	_, err := newCertificateService(storage, &fakeBannerObjectStorage{}).UpdateCertificate(context.Background(), uuid.New(), uuid.New(),
		"X", 0, true, "", "")
	var customErr *customErrors.Error
	if !errors.As(err, &customErr) || customErr.GetStatusCode() != 404 {
		t.Fatalf("error = %v, want not found", err)
	}
}
