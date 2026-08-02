package models

type CreateProductResponse struct {
	Product Product `json:"product"`
}

type GetProductResponse struct {
	Product Product `json:"product"`
}

type ListProductsResponse struct {
	Items      []ProductListItem `json:"items"`
	Pagination Pagination        `json:"pagination"`
}

type UpdateProductResponse struct {
	Product Product `json:"product"`
}

type DeleteProductResponse struct {
	Deleted bool `json:"deleted"`
}

type ListCategoriesResponse struct {
	Items      []Category `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type ListBrandsResponse struct {
	Items      []Brand    `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type UploadProductImageResponse struct {
	Image ProductImage `json:"image"`
}

type DeleteProductImageResponse struct {
	Deleted bool `json:"deleted"`
}

type ProductImageBatchResult struct {
	FileName  string        `json:"fileName"`
	SKU       string        `json:"sku"`
	Success   bool          `json:"success"`
	Image     *ProductImage `json:"image,omitempty"`
	ErrorText string        `json:"errorText,omitempty"`
}

type UploadProductImagesBatchResponse struct {
	Items []ProductImageBatchResult `json:"items"`
}

type CreatePromotionResponse struct {
	Promotion Promotion `json:"promotion"`
}

type GetPromotionResponse struct {
	Promotion Promotion `json:"promotion"`
}

type ListPromotionsResponse struct {
	Items      []Promotion `json:"items"`
	Pagination Pagination  `json:"pagination"`
}

type UpdatePromotionResponse struct {
	Promotion Promotion `json:"promotion"`
}

type DeletePromotionResponse struct {
	Deleted bool `json:"deleted"`
}
