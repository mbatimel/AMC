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

func legalDocPDFBase64() string {
	return base64.StdEncoding.EncodeToString([]byte("%PDF-1.4\n%test"))
}

func newLegalDocService(storage *fakeStorage, objects *fakeBannerObjectStorage) *service {
	return NewAdminApiService(zerolog.Nop(), storage, &fakeAuthClient{}, &fakeAccessClient{allowed: true}, nil, nil, WithObjectStorage(objects, 4096))
}

func TestCreateLegalDoc(t *testing.T) {
	storage := &fakeStorage{legalDocs: map[string]postgres.LegalDoc{}}
	objects := &fakeBannerObjectStorage{}
	result, err := newLegalDocService(storage, objects).CreateLegalDoc(context.Background(), uuid.New(),
		"oferta", "Договор оферты", "", "", "offer.pdf", legalDocPDFBase64())
	if err != nil {
		t.Fatalf("CreateLegalDoc() error = %v", err)
	}
	if result.ID != "oferta" || result.CurrentVersion != defaultLegalDocVersion {
		t.Fatalf("result = %+v", result)
	}
	if len(objects.uploaded) != 1 {
		t.Fatalf("uploaded = %v", objects.uploaded)
	}
}

func TestCreateLegalDoc_InvalidID(t *testing.T) {
	storage := &fakeStorage{legalDocs: map[string]postgres.LegalDoc{}}
	_, err := newLegalDocService(storage, &fakeBannerObjectStorage{}).CreateLegalDoc(context.Background(), uuid.New(),
		"Not Valid!", "Договор оферты", "", "", "offer.pdf", legalDocPDFBase64())
	var customErr *customErrors.Error
	if !errors.As(err, &customErr) || customErr.GetStatusCode() != 400 {
		t.Fatalf("error = %v, want bad request", err)
	}
}

func TestCreateLegalDoc_RequiresFile(t *testing.T) {
	storage := &fakeStorage{legalDocs: map[string]postgres.LegalDoc{}}
	_, err := newLegalDocService(storage, &fakeBannerObjectStorage{}).CreateLegalDoc(context.Background(), uuid.New(),
		"oferta", "Договор оферты", "", "", "", "")
	if err == nil {
		t.Fatal("CreateLegalDoc() error = nil, want bad request when file is missing")
	}
}

func TestCreateLegalDoc_DuplicateIDConflict(t *testing.T) {
	storage := &fakeStorage{createLegalDocFn: func(context.Context, postgres.LegalDoc, postgres.LegalDocVersion) (postgres.LegalDoc, error) {
		return postgres.LegalDoc{}, postgres.ErrLegalDocAlreadyExists
	}}
	objects := &fakeBannerObjectStorage{}
	_, err := newLegalDocService(storage, objects).CreateLegalDoc(context.Background(), uuid.New(),
		"oferta", "Договор оферты", "", "", "offer.pdf", legalDocPDFBase64())
	var customErr *customErrors.Error
	if !errors.As(err, &customErr) || customErr.GetStatusCode() != 409 {
		t.Fatalf("error = %v, want conflict", err)
	}
	if len(objects.deleted) != 1 {
		t.Fatalf("deleted = %v, want compensating delete", objects.deleted)
	}
}

func TestReplaceLegalDocFile_KeepsHistory(t *testing.T) {
	docID := "oferta"
	storage := &fakeStorage{
		getLegalDocFn: func(context.Context, string) (postgres.LegalDoc, error) {
			return postgres.LegalDoc{ID: docID, Name: "Договор оферты", CurrentVersion: "1.0"}, nil
		},
	}
	objects := &fakeBannerObjectStorage{}
	result, err := newLegalDocService(storage, objects).ReplaceLegalDocFile(context.Background(), uuid.New(), docID,
		"1.1", "", "offer.pdf", legalDocPDFBase64())
	if err != nil {
		t.Fatalf("ReplaceLegalDocFile() error = %v", err)
	}
	if result.CurrentVersion != "1.1" {
		t.Fatalf("result = %+v", result)
	}
	if len(objects.deleted) != 0 {
		t.Fatalf("deleted = %v, want no deletions (history must be kept)", objects.deleted)
	}
}

func TestReplaceLegalDocFile_SameVersionConflict(t *testing.T) {
	docID := "oferta"
	storage := &fakeStorage{
		getLegalDocFn: func(context.Context, string) (postgres.LegalDoc, error) {
			return postgres.LegalDoc{ID: docID, CurrentVersion: "1.0"}, nil
		},
	}
	_, err := newLegalDocService(storage, &fakeBannerObjectStorage{}).ReplaceLegalDocFile(context.Background(), uuid.New(), docID,
		"1.0", "", "offer.pdf", legalDocPDFBase64())
	var customErr *customErrors.Error
	if !errors.As(err, &customErr) || customErr.GetStatusCode() != 409 {
		t.Fatalf("error = %v, want conflict for repeated version", err)
	}
}

func TestDeleteLegalDoc_RemovesEveryVersionFile(t *testing.T) {
	docID := "oferta"
	storage := &fakeStorage{
		getLegalDocFn: func(context.Context, string) (postgres.LegalDoc, error) {
			return postgres.LegalDoc{ID: docID, Name: "Договор оферты"}, nil
		},
		legalDocVersions: map[string][]postgres.LegalDocVersion{
			docID: {{Version: "1.0", ObjectKey: "legal-docs/oferta/a.pdf"}, {Version: "1.1", ObjectKey: "legal-docs/oferta/b.pdf"}},
		},
	}
	objects := &fakeBannerObjectStorage{}
	result, err := newLegalDocService(storage, objects).DeleteLegalDoc(context.Background(), uuid.New(), docID)
	if err != nil || !result.Deleted {
		t.Fatalf("result=%+v error=%v", result, err)
	}
	if len(objects.deleted) != 2 {
		t.Fatalf("deleted = %v, want both version files removed", objects.deleted)
	}
}

func TestListLegalDocs_RequiresAdmin(t *testing.T) {
	storage := &fakeStorage{legalDocs: map[string]postgres.LegalDoc{"oferta": {ID: "oferta"}}}
	svc := NewAdminApiService(zerolog.Nop(), storage, &fakeAuthClient{}, &fakeAccessClient{allowed: false}, nil, nil)
	_, err := svc.ListLegalDocs(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("ListLegalDocs() error = nil, want forbidden for non-admin")
	}
}
