package models

// WildberriesDimensions — габариты и вес товара с упаковкой.
// Габариты в сантиметрах, вес брутто — в килограммах.
type WildberriesDimensions struct {
	LengthCM       int     `json:"length_cm"`
	WidthCM        int     `json:"width_cm"`
	HeightCM       int     `json:"height_cm"`
	WeightBruttoKG float64 `json:"weight_brutto_kg"`
}

// WildberriesSize — размер товара с ценой и баркодами.
type WildberriesSize struct {
	TechSize string   `json:"tech_size"`
	WBSize   string   `json:"wb_size"`
	Price    float64  `json:"price"`
	SKUs     []string `json:"skus"`
}

// WildberriesCharacteristic — характеристика предмета (справочник WB).
// ValueJSON — значение в виде валидного JSON, ровно в том формате, которого
// ожидает WB для конкретной характеристики: число ("1200"), строка
// ("\"Turkish flag\"") или массив строк ("[\"red\"]").
type WildberriesCharacteristic struct {
	ID        int    `json:"id"`
	ValueJSON string `json:"value_json"`
}

// WildberriesDocumentItem — документ соответствия (сертификат/декларация).
type WildberriesDocumentItem struct {
	Type          int    `json:"type"`
	Number        string `json:"number"`
	ProductNumber string `json:"product_number"`
	TradeName     string `json:"trade_name"`
	Applicant     string `json:"applicant"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	IsEndless     bool   `json:"is_endless"`
}

// WildberriesDocuments — блок документов карточки товара.
type WildberriesDocuments struct {
	Items            []WildberriesDocumentItem `json:"items"`
	ExcludeDocuments bool                      `json:"exclude_documents"`
}
