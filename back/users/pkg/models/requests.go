package models

type CreateUserRequest struct {
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	MiddleName  string `json:"middle_name"`
	Role        string `json:"role"`
	Status      string `json:"status"`
	ClientID    string `json:"client_id"`
	CompanyName string `json:"company_name"`
	INN         string `json:"inn"`
	IsActive    bool   `json:"is_active"`
}

type UpdateUserRequest struct {
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	MiddleName  string `json:"middle_name,omitempty"`
	Role        string `json:"role,omitempty"`
	Status      string `json:"status,omitempty"`
	ClientID    string `json:"client_id,omitempty"`
	CompanyName string `json:"company_name,omitempty"`
	INN         string `json:"inn,omitempty"`
	IsActive    *bool  `json:"is_active,omitempty"`
}

type UpdateProfileRequest struct {
	Email      string `json:"email,omitempty"`
	Phone      string `json:"phone,omitempty"`
	FirstName  string `json:"first_name,omitempty"`
	LastName   string `json:"last_name,omitempty"`
	MiddleName string `json:"middle_name,omitempty"`
}

type AddFavoriteRequest struct {
	ProductID string `json:"product_id"`
}

type DeleteFavoritesRequest struct {
	ProductIDs []string `json:"product_ids"`
}
