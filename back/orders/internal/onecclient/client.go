package onecclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/mbatimel/AMC/orders/internal/service"
)

const pushPath = "/api/v1/onec-orders/push"

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string, timeout time.Duration) *Client {
	return &Client{baseURL: baseURL, http: &http.Client{Timeout: timeout}}
}

type wireCounterparty struct {
	OneCGUID string `json:"onec_guid"`
	INN      string `json:"inn"`
	Name     string `json:"name"`
}
type wireDelivery struct {
	Type        string `json:"type"`
	Address     string `json:"address"`
	ContactName string `json:"contact_name"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
}
type wireItem struct {
	OneCGUID string  `json:"onec_guid"`
	SKU      string  `json:"sku"`
	Name     string  `json:"name"`
	Qty      int     `json:"qty"`
	Price    float64 `json:"price"`
	VATRate  float64 `json:"vat_rate"`
}
type wireRequest struct {
	ClientOrderID uuid.UUID        `json:"client_order_id"`
	OrderNumber   string           `json:"order_number"`
	Counterparty  wireCounterparty `json:"counterparty"`
	Delivery      wireDelivery     `json:"delivery"`
	Comment       string           `json:"comment"`
	Items         []wireItem       `json:"items"`
}
type wireResponse struct {
	OnecDocumentGUID   uuid.UUID `json:"onec_document_guid"`
	OnecDocumentNumber string    `json:"onec_document_number"`
}

func nullUUIDString(v uuid.NullUUID) string {
	if !v.Valid {
		return ""
	}
	return v.UUID.String()
}

func (c *Client) PushOrder(ctx context.Context, order service.OnecPushOrder) (uuid.UUID, string, error) {
	items := make([]wireItem, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, wireItem{
			OneCGUID: nullUUIDString(item.OneCGUID),
			SKU:      item.SKU,
			Name:     item.Name,
			Qty:      item.Qty,
			Price:    item.Price,
			VATRate:  item.VATRate,
		})
	}
	body, err := json.Marshal(wireRequest{
		ClientOrderID: order.ClientOrderID,
		OrderNumber:   order.OrderNumber,
		Counterparty: wireCounterparty{
			OneCGUID: nullUUIDString(order.CounterpartyGUID),
			INN:      order.CounterpartyINN,
			Name:     order.CounterpartyName,
		},
		Delivery: wireDelivery{
			Type:        order.DeliveryType,
			Address:     order.DeliveryAddress,
			ContactName: order.ContactName,
			Phone:       order.Phone,
			Email:       order.Email,
		},
		Comment: order.Comment,
		Items:   items,
	})
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("push order: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+pushPath, bytes.NewReader(body))
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("push order: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("push order: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return uuid.Nil, "", fmt.Errorf("push order: unexpected status %d", resp.StatusCode)
	}

	var out wireResponse
	if err = json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return uuid.Nil, "", fmt.Errorf("push order: decode response: %w", err)
	}
	return out.OnecDocumentGUID, out.OnecDocumentNumber, nil
}
