package wildberries

import "encoding/json"

// CreateCardParams — параметры создания отдельной (не объединённой) карточки
// товара Wildberries: один subjectID и один вариант.
type CreateCardParams struct {
	SubjectID        int
	VendorCode       string
	Title            string
	Description      string
	Brand            string
	KizMarked        bool
	WholesaleEnabled bool
	WholesaleQuantum int
	Dimensions       Dimensions
	Sizes            []Size
	Characteristics  []Characteristic
	Documents        Documents
}

type Dimensions struct {
	LengthCM       int
	WidthCM        int
	HeightCM       int
	WeightBruttoKG float64
}

type Size struct {
	TechSize string
	WBSize   string
	Price    float64
	SKUs     []string
}

// Characteristic — значение характеристики предмета. ValueJSON — валидный
// JSON ровно в том виде, которого ожидает WB для конкретной характеристики
// (число, строка или массив строк).
type Characteristic struct {
	ID        int
	ValueJSON string
}

type DocumentItem struct {
	Type          int
	Number        string
	ProductNumber string
	TradeName     string
	Applicant     string
	StartDate     string
	EndDate       string
	IsEndless     bool
}

type Documents struct {
	Items            []DocumentItem
	ExcludeDocuments bool
}

// CreateCardResult — создание карточки в WB асинхронно, поэтому в результате
// нет идентификатора карточки: только факт того, что запрос принят.
type CreateCardResult struct {
	Accepted bool
}

// --- wire-формат запроса/ответа Wildberries Content API v2 ---

type uploadCardGroup struct {
	SubjectID int             `json:"subjectID"`
	Variants  []uploadVariant `json:"variants"`
}

type uploadVariant struct {
	Brand           string                 `json:"brand,omitempty"`
	Title           string                 `json:"title,omitempty"`
	Description     string                 `json:"description,omitempty"`
	VendorCode      string                 `json:"vendorCode"`
	KizMarked       bool                   `json:"kizMarked"`
	Wholesale       *uploadWholesale       `json:"wholesale,omitempty"`
	Dimensions      *uploadDimensions      `json:"dimensions,omitempty"`
	Sizes           []uploadSize           `json:"sizes,omitempty"`
	Characteristics []uploadCharacteristic `json:"characteristics,omitempty"`
	Documents       *uploadDocuments       `json:"documents,omitempty"`
}

type uploadWholesale struct {
	Enabled bool `json:"enabled"`
	Quantum int  `json:"quantum,omitempty"`
}

type uploadDimensions struct {
	Length       int     `json:"length"`
	Width        int     `json:"width"`
	Height       int     `json:"height"`
	WeightBrutto float64 `json:"weightBrutto"`
}

type uploadSize struct {
	TechSize string   `json:"techSize,omitempty"`
	WBSize   string   `json:"wbSize,omitempty"`
	Price    float64  `json:"price,omitempty"`
	SKUs     []string `json:"skus,omitempty"`
}

type uploadCharacteristic struct {
	ID    int             `json:"id"`
	Value json.RawMessage `json:"value"`
}

type uploadDocumentItem struct {
	Type          int    `json:"type"`
	Number        string `json:"number,omitempty"`
	ProductNumber string `json:"productNumber,omitempty"`
	TradeName     string `json:"tradeName,omitempty"`
	Applicant     string `json:"applicant,omitempty"`
	StartDate     string `json:"startDate,omitempty"`
	EndDate       string `json:"endDate,omitempty"`
	IsEndless     bool   `json:"isEndless"`
}

type uploadDocuments struct {
	Items            []uploadDocumentItem `json:"items,omitempty"`
	ExcludeDocuments bool                 `json:"excludeDocuments"`
}

type uploadResponse struct {
	Data             json.RawMessage        `json:"data"`
	Error            bool                   `json:"error"`
	ErrorText        string                 `json:"errorText"`
	AdditionalErrors map[string]interface{} `json:"additionalErrors"`
}

// uploadErrorEnvelope — формат ошибок шлюза WB (401/402/403/413/429): поля
// title/detail несут человекочитаемое сообщение, остальное — служебное.
type uploadErrorEnvelope struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Status int    `json:"status"`
}
