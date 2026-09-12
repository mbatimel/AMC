package models

// CreateWildberriesCardResponse — WB создаёт карточку асинхронно: успешный
// ответ 200 означает только то, что запрос принят в обработку (до 30 минут
// на синхронизацию), а не что карточка уже создана.
type CreateWildberriesCardResponse struct {
	Accepted   bool   `json:"accepted"`
	VendorCode string `json:"vendor_code"`
	SubjectID  int    `json:"subject_id"`
}

// CreateOzonCardResponse — Ozon тоже обрабатывает создание/обновление товара
// асинхронно: успешный ответ содержит task_id задания, статус которого
// проверяется отдельным методом /v1/product/import/info (не в рамках этой ручки).
type CreateOzonCardResponse struct {
	Accepted bool   `json:"accepted"`
	OfferID  string `json:"offer_id"`
	TaskID   int64  `json:"task_id"`
}

// CreateYandexMarketCardResponse — в отличие от WB/Ozon, Яндекс Маркет
// обрабатывает offer-mappings/update синхронно: статус и список
// ошибок/предупреждений по товару приходят сразу в ответе.
type CreateYandexMarketCardResponse struct {
	Accepted bool     `json:"accepted"`
	OfferID  string   `json:"offer_id"`
	Warnings []string `json:"warnings,omitempty"`
}
