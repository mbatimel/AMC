// Package onec — HTTP OData-клиент к 1С:УТ 10.3.
//
// Имена entity set'ов и полей сверены с реальными метаданными
// ($metadata) сервера 46.0.206.50:4002/UT. Два момента, не совпадающие с
// «наивным» именованием по документам:
//
//  1. InformationRegister_ЦеныНоменклатуры в этой базе подчинён
//     регистратору (Key = Recorder+Recorder_Type) — верхнеуровневый JSON
//     несёт только Recorder/RecordSet/Recorder_Type, а сами строки цен
//     лежат во вложенном RecordSet. Плоского PriceDTO из value[] не
//     получить, отсюда priceRecorderDTO ниже и ручной flatten в client.go.
//  2. Виртуальная сущность AccumulationRegister_...Balance НЕ публикуется
//     как отдельный EntitySet — баланс запрашивается через OData-функцию
//     .../AccumulationRegister_ТоварыНаСкладах/Balance(), которая уже
//     возвращает плоские строки (в отличие от базового EntitySet, который
//     отдаёт сырые движения, тоже обёрнутые в Recorder/RecordSet).
package onec

type odataEnvelope[T any] struct {
	Value []T `json:"value"`
}

type CategoryDTO struct {
	RefKey      string `json:"Ref_Key"`
	ParentKey   string `json:"Parent_Key"`
	Description string `json:"Description"`
}

type WarehouseDTO struct {
	RefKey      string `json:"Ref_Key"`
	Description string `json:"Description"`
}

type ProductDTO struct {
	RefKey      string `json:"Ref_Key"`
	CategoryKey string `json:"НоменклатурнаяГруппа_Key"`
	Code        string `json:"Code"`
	Description string `json:"Description"`
}

// PriceDTO — плоская строка цены после flatten'а priceRecorderDTO.RecordSet
// (см. fetchPrices в client.go). Из 1С напрямую в таком виде не приходит.
type PriceDTO struct {
	ProductKey   string
	PriceTypeKey string
	Price        float64
}

// priceRecorderDTO — реальная форма ответа
// InformationRegister_ЦеныНоменклатуры: строки цен лежат в RecordSet,
// сгруппированные по документу-регистратору (Recorder).
type priceRecorderDTO struct {
	RecordSet []priceRecordDTO `json:"RecordSet"`
}

type priceRecordDTO struct {
	Active       bool    `json:"Active"`
	ProductKey   string  `json:"Номенклатура_Key"`
	PriceTypeKey string  `json:"ТипЦен_Key"`
	Price        float64 `json:"Цена"`
}

type StockDTO struct {
	ProductKey   string  `json:"Номенклатура_Key"`
	WarehouseKey string  `json:"Склад_Key"`
	Quantity     float64 `json:"КоличествоBalance"`
}
