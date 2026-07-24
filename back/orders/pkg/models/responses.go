package models

import "github.com/google/uuid"

type GetCartResponse struct {
	Cart Cart `json:"cart"`
}

type AddCartItemResponse struct {
	Cart Cart `json:"cart"`
}

type UpdateCartItemResponse struct {
	Cart Cart `json:"cart"`
}

type DeleteCartItemResponse struct {
	Deleted bool `json:"deleted"`
	Cart    Cart `json:"cart"`
}

type ClearCartResponse struct {
	Cleared bool `json:"cleared"`
	Cart    Cart `json:"cart"`
}

type CreateOrderResponse struct {
	Order Order `json:"order"`
}

type GetOrderResponse struct {
	Order Order `json:"order"`
}

type ListOrdersResponse struct {
	Items      []Order    `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type CancelOrderResponse struct {
	Order Order `json:"order"`
}

type RepeatOrderResponse struct {
	Order Order `json:"order"`
	Cart  Cart  `json:"cart,omitempty"`
}

type GetOrderDocumentsResponse struct {
	Items []OrderDocument `json:"items"`
}

type GetOrderHistoryResponse struct {
	Items []OrderHistoryItem `json:"items"`
}

type UpdateOrderStatusResponse struct {
	Order Order `json:"order"`
}
type GetCities struct {
	ID   uuid.UUID `json:"id"`
	City string    `json:"city"`
}
