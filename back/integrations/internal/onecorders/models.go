package onecorders

import "github.com/google/uuid"

type CounterpartyDTO struct {
	OneCGUID string `json:"onec_guid"`
	INN      string `json:"inn"`
	Name     string `json:"name"`
}

type DeliveryDTO struct {
	Type        string `json:"type"`
	Address     string `json:"address"`
	ContactName string `json:"contact_name"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
}

type ItemDTO struct {
	OneCGUID string  `json:"onec_guid"`
	SKU      string  `json:"sku"`
	Name     string  `json:"name"`
	Qty      int     `json:"qty"`
	Price    float64 `json:"price"`
	VATRate  float64 `json:"vat_rate"`
}

type PushOrderRequest struct {
	ClientOrderID uuid.UUID       `json:"client_order_id"`
	OrderNumber   string          `json:"order_number"`
	Counterparty  CounterpartyDTO `json:"counterparty"`
	Delivery      DeliveryDTO     `json:"delivery"`
	Comment       string          `json:"comment"`
	Items         []ItemDTO       `json:"items"`
}

type pushOrderSuccessResponse struct {
	OnecDocumentGUID   uuid.UUID `json:"onec_document_guid"`
	OnecDocumentNumber string    `json:"onec_document_number"`
}

type PushOrderResult struct {
	OnecDocumentGUID   uuid.UUID
	OnecDocumentNumber string
}
