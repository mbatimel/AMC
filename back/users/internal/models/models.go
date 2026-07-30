package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID
	Email          string
	Phone          string
	FirstName      string
	LastName       string
	MiddleName     string
	Role           string
	Status         string
	ClientID       uuid.NullUUID
	CompanyName    string
	INN            string
	IsActive       bool
	ActiveClientID uuid.NullUUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      sql.NullTime
}

type Client struct {
	ID          uuid.UUID
	CompanyName string
	CompanyType string
	INN         string
	OGRN        string
	Address     string
	ContactName string
	Phone       string
	Email       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	IsActive    bool
}

type CategoryDiscount struct {
	CategoryID      uuid.UUID
	CategoryName    string
	DiscountPercent float64
	ValidFrom       sql.NullTime
	ValidTo         sql.NullTime
}

type ClientConditions struct {
	ClientID          uuid.UUID
	PriceGroup        string
	CreditLimit       float64
	CreditUsed        float64
	PaymentTerms      string
	SalesContact      string
	ContactChannel    string
	CategoryDiscounts []CategoryDiscount
}

type Favorite struct {
	UserID    uuid.UUID
	ClientID  uuid.UUID
	ProductID uuid.UUID
	CreatedAt time.Time
}

type CreateUserParams struct {
	Email       string
	Phone       string
	FirstName   string
	LastName    string
	MiddleName  string
	Status      string
	IsActive    bool
	ClientID    *uuid.UUID
	CompanyName string
	INN         string
}

type UpdateUserParams struct {
	UserID      uuid.UUID
	Email       string
	Phone       string
	FirstName   string
	LastName    string
	MiddleName  string
	Status      string
	IsActive    *bool
	ClientID    *uuid.UUID
	CompanyName string
	INN         string
}

type ListUsersParams struct {
	Q        string
	Role     string
	Status   string
	ClientID *uuid.UUID
	IsActive *bool
	Limit    int
	Offset   int
	Sort     string
}

type UpdateProfileParams struct {
	UserID     uuid.UUID
	Email      string
	Phone      string
	FirstName  string
	LastName   string
	MiddleName string
}
