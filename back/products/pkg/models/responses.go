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
