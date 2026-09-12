package models

// OzonDimensions — объёмно-вес товара. Ozon требует реальные значения:
// нулевые depth/width/height/weight приводят к отклонению товара.
type OzonDimensions struct {
	DepthMM       int    `json:"depth_mm"`
	WidthMM       int    `json:"width_mm"`
	HeightMM      int    `json:"height_mm"`
	DimensionUnit string `json:"dimension_unit"`
	Weight        int    `json:"weight"`
	WeightUnit    string `json:"weight_unit"`
}

// OzonAttributeValue — одно значение характеристики: либо значение справочника
// (DictionaryValueID), либо произвольный текст (Value), либо оба.
type OzonAttributeValue struct {
	DictionaryValueID int64  `json:"dictionary_value_id,omitempty"`
	Value             string `json:"value,omitempty"`
}

// OzonAttribute — характеристика товара (см. /v1/description-category/attribute).
type OzonAttribute struct {
	ComplexID int                  `json:"complex_id"`
	ID        int                  `json:"id"`
	Values    []OzonAttributeValue `json:"values"`
}
