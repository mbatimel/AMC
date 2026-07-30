package models

import "time"

type User struct {
	ID             string     `json:"id"`
	Email          string     `json:"email"`
	Phone          string     `json:"phone"`
	FirstName      string     `json:"first_name"`
	LastName       string     `json:"last_name"`
	MiddleName     string     `json:"middle_name"`
	Role           string     `json:"role"`
	Status         string     `json:"status"`
	ClientID       string     `json:"client_id"`
	CompanyName    string     `json:"company_name"`
	INN            string     `json:"inn"`
	IsActive       bool       `json:"is_active"`
	ActiveClientID string     `json:"active_client_id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

type UserListItem = User

type Profile struct {
	UserID         string    `json:"user_id"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	MiddleName     string    `json:"middle_name"`
	Status         string    `json:"status"`
	IsActive       bool      `json:"is_active"`
	ActiveClientID string    `json:"active_client_id"`
	ActiveClient   *Client   `json:"active_client,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Client struct {
	ID          string    `json:"id"`
	CompanyName string    `json:"company_name"`
	CompanyType string    `json:"company_type"`
	INN         string    `json:"inn"`
	OGRN        string    `json:"ogrn"`
	Address     string    `json:"address"`
	ContactName string    `json:"contact_name"`
	Phone       string    `json:"phone"`
	Email       string    `json:"email"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ClientDetails struct {
	Client Client `json:"client"`
}

type ClientConditions struct {
	ClientID          string             `json:"client_id"`
	PriceGroup        string             `json:"price_group"`
	CreditLimit       float64            `json:"credit_limit"`
	CreditUsed        float64            `json:"credit_used"`
	PaymentTerms      string             `json:"payment_terms"`
	CategoryDiscounts []CategoryDiscount `json:"category_discounts"`
	SalesContact      string             `json:"sales_contact"`
	ContactChannel    string             `json:"contact_channel"`
}

type CategoryDiscount struct {
	CategoryID      string  `json:"category_id"`
	CategoryName    string  `json:"category_name"`
	DiscountPercent float64 `json:"discount_percent"`
	ValidFrom       string  `json:"valid_from,omitempty"`
	ValidTo         string  `json:"valid_to,omitempty"`
}

type UserClientListItem struct {
	Client   Client `json:"client"`
	IsActive bool   `json:"is_active"`
}

type Favorite struct {
	UserID    string    `json:"user_id"`
	ClientID  string    `json:"client_id"`
	ProductID string    `json:"product_id"`
	CreatedAt time.Time `json:"created_at"`
}
