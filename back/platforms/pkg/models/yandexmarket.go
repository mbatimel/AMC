package models

// YandexMarketDimensions — габариты упаковки (см) и вес (кг) товара.
type YandexMarketDimensions struct {
	LengthCM float64 `json:"length_cm"`
	WidthCM  float64 `json:"width_cm"`
	HeightCM float64 `json:"height_cm"`
	WeightKG float64 `json:"weight_kg"`
}

// YandexMarketParameterValue — значение категорийной характеристики
// (см. POST v2/category/{categoryId}/parameters).
type YandexMarketParameterValue struct {
	ParameterID int64  `json:"parameter_id"`
	ValueID     int64  `json:"value_id,omitempty"`
	UnitID      int64  `json:"unit_id,omitempty"`
	Value       string `json:"value,omitempty"`
}
