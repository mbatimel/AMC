package service

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	customErrors "github.com/mbatimel/AMC/products/internal/errors"
	internalModels "github.com/mbatimel/AMC/products/internal/models"
	"github.com/mbatimel/AMC/products/internal/storage/postgres"
	"github.com/mbatimel/AMC/products/pkg/models"
)

type fakeObjectStorage struct {
	uploadErr error
	deleteErr error
	uploaded  []string
	deleted   []string
	content   []byte
}

func (f *fakeObjectStorage) Upload(_ context.Context, objectKey string, body io.Reader, _ int64, _ string) error {
	if f.uploadErr != nil {
		return f.uploadErr
	}
	f.uploaded = append(f.uploaded, objectKey)
	f.content, _ = io.ReadAll(body)
	return nil
}

func (f *fakeObjectStorage) Delete(_ context.Context, objectKey string) error {
	f.deleted = append(f.deleted, objectKey)
	return f.deleteErr
}

func (f *fakeObjectStorage) URL(objectKey string) string {
	return "https://cdn.example/amc-images/" + objectKey
}

func pngFile(name string) models.ImageFile {
	return models.ImageFile{FileName: name, Content: []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")}
}

func newImageService(storage *fakeStorage, objectStorage *fakeObjectStorage, maxSize int64) *Service {
	return New(zerolog.Nop(), storage, &fakeAccess{allowed: true}, WithObjectStorage(objectStorage, maxSize))
}

func TestUploadProductImageSuccess(t *testing.T) {
	productID := sampleProduct().ID
	var saved internalModels.ProductImage
	storage := &fakeStorage{addUploadedProductImageFn: func(_ context.Context, gotProductID uuid.UUID, image internalModels.ProductImage) (internalModels.ProductImage, error) {
		if gotProductID != productID {
			t.Fatalf("productID = %s, want %s", gotProductID, productID)
		}
		saved = image
		image.ID = uuid.New()
		image.ProductID = gotProductID
		return image, nil
	}}
	objects := &fakeObjectStorage{}

	response, err := newImageService(storage, objects, 1024).UploadProductImage(context.Background(), uuid.New(), productID, pngFile("product.png"))
	if err != nil {
		t.Fatalf("UploadProductImage() error = %v", err)
	}
	if len(objects.uploaded) != 1 || !strings.HasPrefix(objects.uploaded[0], "products/"+productID.String()+"/") {
		t.Fatalf("uploaded keys = %v", objects.uploaded)
	}
	if saved.ProductID != productID || saved.ObjectKey != objects.uploaded[0] || saved.ContentType != "image/png" {
		t.Fatalf("saved image = %+v", saved)
	}
	if response.Image.ProductID != productID.String() || response.Image.URL == "" {
		t.Fatalf("response = %+v", response)
	}
}

func TestUploadProductImageValidation(t *testing.T) {
	tests := []struct {
		name string
		file models.ImageFile
		max  int64
	}{
		{name: "unsupported MIME", file: models.ImageFile{FileName: "payload.jpg", Content: []byte("not an image")}, max: 1024},
		{name: "too large", file: pngFile("product.png"), max: 4},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			objects := &fakeObjectStorage{}
			_, err := newImageService(&fakeStorage{}, objects, test.max).UploadProductImage(context.Background(), uuid.New(), sampleProduct().ID, test.file)
			if !errors.Is(err, customErrors.ErrValidation) {
				t.Fatalf("error = %v, want validation", err)
			}
			if len(objects.uploaded) != 0 {
				t.Fatalf("uploaded keys = %v, want none", objects.uploaded)
			}
		})
	}
}

func TestUploadProductImageS3AndCompensationErrors(t *testing.T) {
	t.Run("S3 upload error", func(t *testing.T) {
		objects := &fakeObjectStorage{uploadErr: errors.New("s3 unavailable")}
		_, err := newImageService(&fakeStorage{}, objects, 1024).UploadProductImage(context.Background(), uuid.New(), sampleProduct().ID, pngFile("product.png"))
		if !errors.Is(err, customErrors.ErrInternal) {
			t.Fatalf("error = %v, want internal", err)
		}
	})

	t.Run("database error deletes uploaded object", func(t *testing.T) {
		storage := &fakeStorage{addUploadedProductImageFn: func(context.Context, uuid.UUID, internalModels.ProductImage) (internalModels.ProductImage, error) {
			return internalModels.ProductImage{}, errors.New("database unavailable")
		}}
		objects := &fakeObjectStorage{}
		_, err := newImageService(storage, objects, 1024).UploadProductImage(context.Background(), uuid.New(), sampleProduct().ID, pngFile("product.png"))
		if !errors.Is(err, customErrors.ErrInternal) {
			t.Fatalf("error = %v, want internal", err)
		}
		if len(objects.deleted) != 1 || objects.deleted[0] != objects.uploaded[0] {
			t.Fatalf("uploaded=%v deleted=%v", objects.uploaded, objects.deleted)
		}
	})
}

func TestDeleteProductImage(t *testing.T) {
	productID := sampleProduct().ID
	imageID := uuid.New()
	objectKey := "products/" + productID.String() + "/image.png"
	storage := &fakeStorage{getProductImageFn: func(_ context.Context, gotProductID uuid.UUID, gotImageID uuid.UUID) (internalModels.ProductImage, error) {
		if gotProductID != productID || gotImageID != imageID {
			return internalModels.ProductImage{}, postgres.ErrProductImageNotFound
		}
		return internalModels.ProductImage{ID: imageID, ProductID: productID, ObjectKey: objectKey}, nil
	}}
	objects := &fakeObjectStorage{} // Missing S3 objects are also a successful idempotent delete.
	response, err := newImageService(storage, objects, 1024).DeleteProductImage(context.Background(), uuid.New(), productID, imageID)
	if err != nil || !response.Deleted {
		t.Fatalf("DeleteProductImage() response=%+v error=%v", response, err)
	}
	if len(objects.deleted) != 1 || objects.deleted[0] != objectKey {
		t.Fatalf("deleted keys = %v", objects.deleted)
	}
}

func TestDeleteProductImageRejectsAnotherProduct(t *testing.T) {
	storage := &fakeStorage{getProductImageFn: func(context.Context, uuid.UUID, uuid.UUID) (internalModels.ProductImage, error) {
		return internalModels.ProductImage{}, postgres.ErrProductImageNotFound
	}}
	objects := &fakeObjectStorage{}
	_, err := newImageService(storage, objects, 1024).DeleteProductImage(context.Background(), uuid.New(), uuid.New(), uuid.New())
	if !errors.Is(err, customErrors.ErrNotFound) {
		t.Fatalf("error = %v, want not found", err)
	}
	if len(objects.deleted) != 0 {
		t.Fatalf("deleted keys = %v, want none", objects.deleted)
	}
}

func TestUploadProductImagesBatchPartialResultAndDuplicateNames(t *testing.T) {
	product := sampleProduct()
	storage := &fakeStorage{getProductBySKUFn: func(_ context.Context, sku string) (internalModels.Product, error) {
		if sku == product.SKU {
			return product, nil
		}
		return internalModels.Product{}, postgres.ErrProductNotFound
	}}
	objects := &fakeObjectStorage{}
	files := []models.ImageFile{
		pngFile(product.SKU + ".png"),
		pngFile("missing.png"),
		{FileName: "invalid", Content: pngFile("x.png").Content},
		pngFile(product.SKU + ".png"),
	}
	response, err := newImageService(storage, objects, 1024).UploadProductImagesBatch(context.Background(), uuid.New(), files)
	if err != nil {
		t.Fatalf("UploadProductImagesBatch() error = %v", err)
	}
	if len(response.Items) != 4 || !response.Items[0].Success || response.Items[1].Success || response.Items[2].Success || !response.Items[3].Success {
		t.Fatalf("items = %+v", response.Items)
	}
	if response.Items[0].SKU != product.SKU || len(objects.uploaded) != 2 || objects.uploaded[0] == objects.uploaded[1] {
		t.Fatalf("items=%+v uploaded=%v", response.Items, objects.uploaded)
	}
}
