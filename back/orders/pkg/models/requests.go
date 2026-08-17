package models

type GetCartRequest struct {
	UserID   string `json:"user_id"`
	ClientID string `json:"client_id"`
}

type AddCartItemRequest struct {
	UserID    string `json:"user_id"`
	ClientID  string `json:"client_id"`
	ProductID string `json:"product_id"`
	Qty       int    `json:"qty"`
}

type UpdateCartItemRequest struct {
	UserID     string `json:"user_id"`
	ClientID   string `json:"client_id"`
	CartItemID string `json:"cart_item_id"`
	Qty        int    `json:"qty"`
}

type DeleteCartItemRequest struct {
	UserID     string `json:"user_id"`
	ClientID   string `json:"client_id"`
	CartItemID string `json:"cart_item_id"`
}

type ClearCartRequest struct {
	UserID   string `json:"user_id"`
	ClientID string `json:"client_id"`
}

type CreateOrderRequest struct {
	UserID          string `json:"user_id"`
	ClientID        string `json:"client_id"`
	DeliveryType    string `json:"delivery_type"`
	DeliveryAddress string `json:"delivery_address"`
	ContactName     string `json:"contact_name"`
	Phone           string `json:"phone"`
	Email           string `json:"email"`
	Comment         string `json:"comment"`
}

type GetOrderRequest struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id,omitempty"`
	ClientID string `json:"client_id,omitempty"`
}

type ListOrdersRequest struct {
	UserID        string `json:"user_id,omitempty"`
	ClientID      string `json:"client_id,omitempty"`
	Status        string `json:"status,omitempty"`
	PaymentStatus string `json:"payment_status,omitempty"`
	Limit         int    `json:"limit,omitempty"`
	Offset        int    `json:"offset,omitempty"`
	Sort          string `json:"sort,omitempty"`
}

type ListPreviouslyOrderedProductsRequest struct {
	UserID   string `json:"user_id,omitempty"`
	ClientID string `json:"client_id,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Offset   int    `json:"offset,omitempty"`
}

type CancelOrderRequest struct {
	ID      string `json:"id"`
	UserID  string `json:"user_id,omitempty"`
	Comment string `json:"comment,omitempty"`
}

type RepeatOrderRequest struct {
	ID     string `json:"id"`
	UserID string `json:"user_id,omitempty"`
}

type GetOrderDocumentsRequest struct {
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id,omitempty"`
}

type GetOrderHistoryRequest struct {
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id,omitempty"`
}

type UpdateOrderStatusRequest struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	PaymentStatus string `json:"payment_status,omitempty"`
	Comment       string `json:"comment,omitempty"`
	ChangedBy     string `json:"changed_by,omitempty"`
}
