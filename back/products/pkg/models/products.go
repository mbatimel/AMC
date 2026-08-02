package models

import "time"

type Product struct {
	ID              string         `json:"id"`
	SKU             string         `json:"sku"`
	Name            string         `json:"name"`
	Description     string         `json:"description"`
	CategoryID      string         `json:"category_id"`
	CategoryName    string         `json:"category_name"`
	BrandID         string         `json:"brand_id"`
	BrandName       string         `json:"brand_name"`
	GOST            string         `json:"gost"`
	Material        string         `json:"material"`
	Size            string         `json:"size"`
	PackageQty      int            `json:"package_qty"`
	StockQty        int            `json:"stock_qty"`
	BasePrice       float64        `json:"base_price"`
	ClientPrice     float64        `json:"client_price"`
	DiscountPercent float64        `json:"discount_percent"`
	Images          []ProductImage `json:"images"`
	IsPublished     bool           `json:"is_published"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

type ProductListItem struct {
	ID              string         `json:"id"`
	SKU             string         `json:"sku"`
	Name            string         `json:"name"`
	CategoryID      string         `json:"category_id"`
	CategoryName    string         `json:"category_name"`
	BrandID         string         `json:"brand_id"`
	BrandName       string         `json:"brand_name"`
	GOST            string         `json:"gost"`
	Material        string         `json:"material"`
	Size            string         `json:"size"`
	PackageQty      int            `json:"package_qty"`
	StockQty        int            `json:"stock_qty"`
	BasePrice       float64        `json:"base_price"`
	ClientPrice     float64        `json:"client_price"`
	DiscountPercent float64        `json:"discount_percent"`
	Images          []ProductImage `json:"images"`
	IsPublished     bool           `json:"is_published"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

type Category struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug,omitempty"`
	ParentID  string    `json:"parent_id,omitempty"`
	SortOrder int       `json:"sort_order"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Brand struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProductImage struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	URL       string    `json:"url"`
	Alt       string    `json:"alt,omitempty"`
	SortOrder int       `json:"sort_order"`
	IsPrimary bool      `json:"is_primary"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProductPrice struct {
	ID              string    `json:"id"`
	ProductID       string    `json:"product_id"`
	BasePrice       float64   `json:"base_price"`
	ClientPrice     float64   `json:"client_price"`
	DiscountPercent float64   `json:"discount_percent"`
	Currency        string    `json:"currency"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ProductStock struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	StockQty  int       `json:"stock_qty"`
	Warehouse string    `json:"warehouse,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Promotion struct {
	ID              string             `json:"id"`
	Name            string             `json:"name"`
	DiscountPercent float64            `json:"discount_percent"`
	StartsAt        time.Time          `json:"starts_at"`
	EndsAt          time.Time          `json:"ends_at"`
	Status          string             `json:"status"`
	Products        []PromotionProduct `json:"products"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

type PromotionProduct struct {
	ProductID string `json:"product_id"`
	MinQty    int    `json:"min_qty"`
}
