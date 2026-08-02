package models

type ImageFile struct {
	FileName string `json:"fileName"`
	Content  []byte `json:"-"`
}

type CreateProductRequest struct {
	SKU             string         `json:"sku"`
	Name            string         `json:"name"`
	Description     string         `json:"description"`
	CategoryID      string         `json:"category_id"`
	BrandID         string         `json:"brand_id"`
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
}

type GetProductRequest struct {
	ID string `json:"id"`
}

type ListProductsRequest struct {
	Q          *string `json:"q,omitempty"`
	CategoryID *string `json:"category_id,omitempty"`
	BrandID    *string `json:"brand_id,omitempty"`
	Material   *string `json:"material,omitempty"`
	Size       *string `json:"size,omitempty"`
	GOST       *string `json:"gost,omitempty"`
	InStock    *bool   `json:"in_stock,omitempty"`
	Limit      *int    `json:"limit,omitempty"`
	Offset     *int    `json:"offset,omitempty"`
	Sort       *string `json:"sort,omitempty"`
}

type UpdateProductRequest struct {
	ID              string          `json:"id"`
	SKU             *string         `json:"sku,omitempty"`
	Name            *string         `json:"name,omitempty"`
	Description     *string         `json:"description,omitempty"`
	CategoryID      *string         `json:"category_id,omitempty"`
	BrandID         *string         `json:"brand_id,omitempty"`
	GOST            *string         `json:"gost,omitempty"`
	Material        *string         `json:"material,omitempty"`
	Size            *string         `json:"size,omitempty"`
	PackageQty      *int            `json:"package_qty,omitempty"`
	StockQty        *int            `json:"stock_qty,omitempty"`
	BasePrice       *float64        `json:"base_price,omitempty"`
	ClientPrice     *float64        `json:"client_price,omitempty"`
	DiscountPercent *float64        `json:"discount_percent,omitempty"`
	Images          *[]ProductImage `json:"images,omitempty"`
	IsPublished     *bool           `json:"is_published,omitempty"`
}

type DeleteProductRequest struct {
	ID string `json:"id"`
}

type ListCategoriesRequest struct {
	Limit  *int `json:"limit,omitempty"`
	Offset *int `json:"offset,omitempty"`
}

type ListBrandsRequest struct {
	Limit  *int `json:"limit,omitempty"`
	Offset *int `json:"offset,omitempty"`
}

type CreatePromotionRequest struct {
	Name            string             `json:"name"`
	DiscountPercent float64            `json:"discount_percent"`
	StartsAt        string             `json:"starts_at"`
	EndsAt          string             `json:"ends_at"`
	Products        []PromotionProduct `json:"products"`
}

type GetPromotionRequest struct {
	ID string `json:"id"`
}

type ListPromotionsRequest struct {
	Limit  *int `json:"limit,omitempty"`
	Offset *int `json:"offset,omitempty"`
}

type UpdatePromotionRequest struct {
	ID              string             `json:"id"`
	Name            string             `json:"name"`
	DiscountPercent float64            `json:"discount_percent"`
	StartsAt        string             `json:"starts_at"`
	EndsAt          string             `json:"ends_at"`
	Products        []PromotionProduct `json:"products"`
}

type DeletePromotionRequest struct {
	ID string `json:"id"`
}
