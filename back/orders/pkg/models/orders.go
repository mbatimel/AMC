package models

import "time"

type Cart struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	ClientID  string     `json:"client_id"`
	Items     []CartItem `json:"items"`
	Total     float64    `json:"total"`
	VAT       float64    `json:"vat"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type CartItem struct {
	ID          string    `json:"id"`
	CartID      string    `json:"cart_id"`
	ProductID   string    `json:"product_id"`
	SKU         string    `json:"sku"`
	ProductName string    `json:"product_name"`
	Qty         int       `json:"qty"`
	Price       float64   `json:"price"`
	Total       float64   `json:"total"`
	VAT         float64   `json:"vat"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Order struct {
	ID              string             `json:"id"`
	Number          string             `json:"number"`
	UserID          string             `json:"user_id"`
	ClientID        string             `json:"client_id"`
	Items           []OrderItem        `json:"items"`
	Total           float64            `json:"total"`
	VAT             float64            `json:"vat"`
	DeliveryType    string             `json:"delivery_type"`
	DeliveryAddress string             `json:"delivery_address"`
	ContactName     string             `json:"contact_name"`
	Phone           string             `json:"phone"`
	Email           string             `json:"email"`
	Comment         string             `json:"comment"`
	Status          string             `json:"status"`
	PaymentStatus   string             `json:"payment_status"`
	Documents       []OrderDocument    `json:"documents"`
	History         []OrderHistoryItem `json:"history"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

type OrderItem struct {
	ID          string    `json:"id"`
	OrderID     string    `json:"order_id"`
	ProductID   string    `json:"product_id"`
	SKU         string    `json:"sku"`
	ProductName string    `json:"product_name"`
	Qty         int       `json:"qty"`
	Price       float64   `json:"price"`
	Total       float64   `json:"total"`
	VAT         float64   `json:"vat"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type OrderDocument struct {
	ID        string    `json:"id"`
	OrderID   string    `json:"order_id"`
	Type      string    `json:"type"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OrderHistoryItem struct {
	ID            string    `json:"id"`
	OrderID       string    `json:"order_id"`
	Status        string    `json:"status"`
	PaymentStatus string    `json:"payment_status"`
	Comment       string    `json:"comment"`
	ChangedBy     string    `json:"changed_by"`
	CreatedAt     time.Time `json:"created_at"`
}
