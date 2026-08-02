package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID              uuid.UUID
	SKU             string
	Name            string
	Description     string
	CategoryID      uuid.UUID
	CategoryName    string
	BrandID         uuid.UUID
	BrandName       string
	GOST            string
	Material        string
	Size            string
	PackageQty      int
	StockQty        int
	BasePrice       float64
	ClientPrice     float64
	DiscountPercent float64
	Images          []ProductImage
	IsPublished     bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ProductImage struct {
	ID           uuid.UUID
	ProductID    uuid.UUID
	FileID       uuid.UUID
	URL          string
	ObjectKey    string
	OriginalName string
	ContentType  string
	SizeBytes    int64
	Alt          string
	SortOrder    int
	IsPrimary    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Category struct {
	ID        uuid.UUID
	Name      string
	Slug      string
	ParentID  uuid.NullUUID
	SortOrder int
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Brand struct {
	ID        uuid.UUID
	Name      string
	Slug      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateProductParams struct {
	SKU             string
	Name            string
	Description     string
	CategoryID      uuid.UUID
	BrandID         uuid.UUID
	GOST            string
	Material        string
	Size            string
	PackageQty      int
	StockQty        int
	BasePrice       float64
	DiscountPercent float64
	Images          []ProductImage
	IsPublished     bool
}

type ListProductsParams struct {
	Q          *string
	CategoryID *uuid.UUID
	BrandID    *uuid.UUID
	Material   *string
	Size       *string
	GOST       *string
	InStock    *bool
	Limit      int
	Offset     int
	Sort       string
}

type UpdateProductParams struct {
	ProductID       uuid.UUID
	SKU             *string
	Name            *string
	Description     *string
	CategoryID      *uuid.UUID
	BrandID         *uuid.UUID
	GOST            *string
	Material        *string
	Size            *string
	PackageQty      *int
	StockQty        *int
	BasePrice       *float64
	DiscountPercent *float64
	Images          *[]ProductImage
	IsPublished     *bool
}

type ProductPrice struct {
	ID              uuid.UUID
	ProductID       uuid.UUID
	BasePrice       float64
	ClientPrice     float64
	DiscountPercent float64
	Currency        string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ProductStock struct {
	ID        uuid.UUID
	ProductID uuid.UUID
	StockQty  int
	Warehouse sql.NullString
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Promotion struct {
	ID              uuid.UUID
	Name            string
	DiscountPercent float64
	StartsAt        time.Time
	EndsAt          time.Time
	Products        []PromotionProduct
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type PromotionProduct struct {
	ProductID uuid.UUID
	MinQty    int
}

type CreatePromotionParams struct {
	Name            string
	DiscountPercent float64
	StartsAt        time.Time
	EndsAt          time.Time
	Products        []PromotionProduct
}

type UpdatePromotionParams struct {
	PromotionID     uuid.UUID
	Name            string
	DiscountPercent float64
	StartsAt        time.Time
	EndsAt          time.Time
	Products        []PromotionProduct
}
