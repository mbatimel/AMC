package yandexmarket

// CreateCardParams — параметры создания/обновления одного товара (один
// элемент offerMappings[] запроса /v2/businesses/{businessId}/offer-mappings/update).
type CreateCardParams struct {
	OfferID          string
	Name             string
	MarketCategoryID int64
	Description      string
	Vendor           string
	VendorCode       string
	Pictures         []string
	Barcodes         []string
	Dimensions       Dimensions
	ParameterValues  []ParameterValue
}

type Dimensions struct {
	LengthCM float64
	WidthCM  float64
	HeightCM float64
	WeightKG float64
}

type ParameterValue struct {
	ParameterID int64
	ValueID     int64
	UnitID      int64
	Value       string
}

// CreateCardResult — offer-mappings/update обрабатывается синхронно: успех
// известен сразу, без опроса статуса отдельным методом.
type CreateCardResult struct {
	Accepted bool
	Warnings []string
}

// --- wire-формат запроса/ответа Yandex Market Partner API v2 ---

type updateRequest struct {
	OfferMappings           []offerMapping `json:"offerMappings"`
	OnlyPartnerMediaContent bool           `json:"onlyPartnerMediaContent"`
}

type offerMapping struct {
	Offer offerDTO `json:"offer"`
}

type offerDTO struct {
	OfferID          string              `json:"offerId"`
	Name             string              `json:"name"`
	MarketCategoryID int64               `json:"marketCategoryId"`
	Description      string              `json:"description"`
	Vendor           string              `json:"vendor"`
	VendorCode       string              `json:"vendorCode,omitempty"`
	Pictures         []string            `json:"pictures,omitempty"`
	Barcodes         []string            `json:"barcodes,omitempty"`
	WeightDimensions *weightDimensions   `json:"weightDimensions,omitempty"`
	ParameterValues  []parameterValueDTO `json:"parameterValues,omitempty"`
}

type weightDimensions struct {
	Length float64 `json:"length,omitempty"`
	Width  float64 `json:"width,omitempty"`
	Height float64 `json:"height,omitempty"`
	Weight float64 `json:"weight,omitempty"`
}

type parameterValueDTO struct {
	ParameterID int64  `json:"parameterId"`
	ValueID     int64  `json:"valueId,omitempty"`
	UnitID      int64  `json:"unitId,omitempty"`
	Value       string `json:"value,omitempty"`
}

type offerMappingError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type updateResult struct {
	OfferID  string              `json:"offerId"`
	Errors   []offerMappingError `json:"errors"`
	Warnings []offerMappingError `json:"warnings"`
}

type updateResponse struct {
	Status  string         `json:"status"`
	Results []updateResult `json:"results"`
}

// errorEnvelope — единый формат ошибок Yandex Market Partner API (400/401/403/404/420/423/500).
type errorEnvelope struct {
	Errors []struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
}
