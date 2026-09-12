package ozon

// CreateCardParams — параметры создания/обновления одного товара (один
// элемент items[] запроса /v3/product/import).
type CreateCardParams struct {
	OfferID               string
	Name                  string
	DescriptionCategoryID int64
	TypeID                int64
	CurrencyCode          string
	Price                 string
	OldPrice              string
	VAT                   string
	Barcode               string
	Dimensions            Dimensions
	Images                []string
	PrimaryImage          string
	ColorImage            string
	Attributes            []Attribute
}

type Dimensions struct {
	DepthMM       int
	WidthMM       int
	HeightMM      int
	DimensionUnit string
	Weight        int
	WeightUnit    string
}

type AttributeValue struct {
	DictionaryValueID int64
	Value             string
}

type Attribute struct {
	ComplexID int
	ID        int
	Values    []AttributeValue
}

// CreateCardResult — Ozon обрабатывает импорт асинхронно: сразу возвращается
// только номер задания, статус которого проверяется отдельным методом
// /v1/product/import/info.
type CreateCardResult struct {
	Accepted bool
	TaskID   int64
}

// --- wire-формат запроса/ответа Ozon Seller API v3 ---

type importRequest struct {
	Items []importItem `json:"items"`
}

type importItem struct {
	Attributes            []importAttribute `json:"attributes"`
	Barcode               string            `json:"barcode,omitempty"`
	DescriptionCategoryID int64             `json:"description_category_id"`
	ColorImage            string            `json:"color_image,omitempty"`
	CurrencyCode          string            `json:"currency_code,omitempty"`
	Depth                 int               `json:"depth"`
	DimensionUnit         string            `json:"dimension_unit"`
	Height                int               `json:"height"`
	Images                []string          `json:"images,omitempty"`
	Name                  string            `json:"name"`
	OfferID               string            `json:"offer_id"`
	OldPrice              string            `json:"old_price,omitempty"`
	Price                 string            `json:"price"`
	PrimaryImage          string            `json:"primary_image,omitempty"`
	TypeID                int64             `json:"type_id"`
	VAT                   string            `json:"vat"`
	Weight                int               `json:"weight"`
	WeightUnit            string            `json:"weight_unit"`
	Width                 int               `json:"width"`
}

type importAttributeValue struct {
	DictionaryValueID int64  `json:"dictionary_value_id,omitempty"`
	Value             string `json:"value,omitempty"`
}

type importAttribute struct {
	ComplexID int                    `json:"complex_id,omitempty"`
	ID        int                    `json:"id"`
	Values    []importAttributeValue `json:"values"`
}

type importResponse struct {
	Result struct {
		TaskID int64 `json:"task_id"`
	} `json:"result"`
}

// errorEnvelope — единый формат ошибок Ozon Seller API (400/403/404/409/429/500).
type errorEnvelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
