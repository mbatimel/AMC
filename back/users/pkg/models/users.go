package models

import "time"

type User struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	MiddleName  string    `json:"middle_name"`
	Role        string    `json:"role"`
	Status      string    `json:"status"`
	ClientID    string    `json:"client_id"`
	CompanyName string    `json:"company_name"`
	INN         string    `json:"inn"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserListItem struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	MiddleName  string    `json:"middle_name"`
	Role        string    `json:"role"`
	Status      string    `json:"status"`
	ClientID    string    `json:"client_id"`
	CompanyName string    `json:"company_name"`
	INN         string    `json:"inn"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
