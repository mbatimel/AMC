package service

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	customErrors "github.com/mbatimel/AMC/admin/internal/errors"
	"github.com/mbatimel/AMC/admin/internal/storage/postgres"
	"github.com/mbatimel/AMC/admin/pkg/models"
)

type fakeBannerObjectStorage struct {
	uploadErr error
	deleteErr error
	uploaded  []string
	deleted   []string
}

func (f *fakeBannerObjectStorage) Upload(_ context.Context, key string, _ io.Reader, _ int64, _ string) error {
	if f.uploadErr != nil {
		return f.uploadErr
	}
	f.uploaded = append(f.uploaded, key)
	return nil
}

func (f *fakeBannerObjectStorage) Delete(_ context.Context, key string) error {
	f.deleted = append(f.deleted, key)
	return f.deleteErr
}

func (f *fakeBannerObjectStorage) URL(key string) string {
	return "https://cdn.example/amc-images/" + key
}

func bannerPNG() *models.ImageFile {
	return &models.ImageFile{FileName: "banner.png", Content: []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")}
}

func validBannerInput() models.BannerInput {
	return models.BannerInput{
		Title: "Banner", Link: "/catalog", SortOrder: 1, IsActive: true, Image: bannerPNG(),
	}
}

func newBannerService(storage *fakeStorage, objects *fakeBannerObjectStorage) *service {
	return NewAdminApiService(zerolog.Nop(), storage, &fakeAuthClient{}, &fakeAccessClient{allowed: true}, WithObjectStorage(objects, 1024))
}

func TestCreateBannerWithImage(t *testing.T) {
	var saved postgres.Banner
	storage := &fakeStorage{createBannerFn: func(_ context.Context, banner postgres.Banner) (postgres.Banner, error) {
		saved = banner
		return banner, nil
	}}
	objects := &fakeBannerObjectStorage{}
	result, err := newBannerService(storage, objects).CreateBanner(context.Background(), uuid.New(), validBannerInput())
	if err != nil {
		t.Fatalf("CreateBanner() error = %v", err)
	}
	if len(objects.uploaded) != 1 || !strings.HasPrefix(objects.uploaded[0], "banners/"+result.ID.String()+"/") {
		t.Fatalf("uploaded = %v", objects.uploaded)
	}
	if saved.ObjectKey != objects.uploaded[0] || saved.ContentType != "image/png" || result.Image == "" {
		t.Fatalf("saved=%+v result=%+v", saved, result)
	}
}

func TestCreateBannerUploadError(t *testing.T) {
	objects := &fakeBannerObjectStorage{uploadErr: errors.New("s3 unavailable")}
	_, err := newBannerService(&fakeStorage{}, objects).CreateBanner(context.Background(), uuid.New(), validBannerInput())
	var customErr *customErrors.Error
	if !errors.As(err, &customErr) || customErr.GetStatusCode() != 500 {
		t.Fatalf("error = %v, want internal", err)
	}
}

func TestCreateBannerCompensatesDatabaseError(t *testing.T) {
	storage := &fakeStorage{createBannerFn: func(context.Context, postgres.Banner) (postgres.Banner, error) {
		return postgres.Banner{}, errors.New("database unavailable")
	}}
	objects := &fakeBannerObjectStorage{}
	_, err := newBannerService(storage, objects).CreateBanner(context.Background(), uuid.New(), validBannerInput())
	if err == nil || len(objects.deleted) != 1 || objects.deleted[0] != objects.uploaded[0] {
		t.Fatalf("error=%v uploaded=%v deleted=%v", err, objects.uploaded, objects.deleted)
	}
}

func TestListBannersReturnsImageURL(t *testing.T) {
	bannerID := uuid.New()
	storage := &fakeStorage{listBannersFn: func(_ context.Context, visibleOnly bool) ([]postgres.Banner, int, error) {
		if visibleOnly {
			t.Fatal("admin list unexpectedly requested visible-only banners")
		}
		return []postgres.Banner{{ID: bannerID, Title: "Banner", ImageURL: "https://cdn.example/banner.png"}}, 7, nil
	}}
	result, err := newBannerService(storage, &fakeBannerObjectStorage{}).ListBanners(context.Background(), uuid.New())
	if err != nil || result.DelaySec != 7 || len(result.Items) != 1 || result.Items[0].Image == "" {
		t.Fatalf("result=%+v error=%v", result, err)
	}
}

func TestUpdateBannerReplacesImageAndDeletesOldObject(t *testing.T) {
	bannerID := uuid.New()
	oldKey := "banners/" + bannerID.String() + "/old.png"
	storage := &fakeStorage{
		getBannerFn: func(context.Context, uuid.UUID) (postgres.Banner, error) {
			return postgres.Banner{ID: bannerID, FileID: uuid.New(), ObjectKey: oldKey}, nil
		},
		updateBannerFn: func(_ context.Context, banner postgres.Banner, replaceImage bool) (postgres.Banner, error) {
			if !replaceImage || banner.ObjectKey == oldKey {
				t.Fatalf("replaceImage=%v banner=%+v", replaceImage, banner)
			}
			return banner, nil
		},
	}
	objects := &fakeBannerObjectStorage{}
	result, err := newBannerService(storage, objects).UpdateBanner(context.Background(), uuid.New(), bannerID, validBannerInput())
	if err != nil || result.Image == "" {
		t.Fatalf("result=%+v error=%v", result, err)
	}
	if len(objects.deleted) != 1 || objects.deleted[0] != oldKey {
		t.Fatalf("deleted = %v", objects.deleted)
	}
}

func TestUpdateBannerCompensatesDatabaseError(t *testing.T) {
	bannerID := uuid.New()
	storage := &fakeStorage{
		getBannerFn: func(context.Context, uuid.UUID) (postgres.Banner, error) {
			return postgres.Banner{ID: bannerID}, nil
		},
		updateBannerFn: func(context.Context, postgres.Banner, bool) (postgres.Banner, error) {
			return postgres.Banner{}, errors.New("database unavailable")
		},
	}
	objects := &fakeBannerObjectStorage{}
	_, err := newBannerService(storage, objects).UpdateBanner(context.Background(), uuid.New(), bannerID, validBannerInput())
	if err == nil || len(objects.deleted) != 1 || objects.deleted[0] != objects.uploaded[0] {
		t.Fatalf("error=%v uploaded=%v deleted=%v", err, objects.uploaded, objects.deleted)
	}
}

func TestDeleteBanner(t *testing.T) {
	bannerID := uuid.New()
	objectKey := "banners/" + bannerID.String() + "/image.png"
	storage := &fakeStorage{getBannerFn: func(context.Context, uuid.UUID) (postgres.Banner, error) {
		return postgres.Banner{ID: bannerID, ObjectKey: objectKey}, nil
	}}
	objects := &fakeBannerObjectStorage{}
	result, err := newBannerService(storage, objects).DeleteBanner(context.Background(), uuid.New(), bannerID)
	if err != nil || !result.Deleted || len(objects.deleted) != 1 || objects.deleted[0] != objectKey {
		t.Fatalf("result=%+v error=%v deleted=%v", result, err, objects.deleted)
	}
}
