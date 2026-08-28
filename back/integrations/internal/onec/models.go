// Package onec — HTTP OData-клиент к 1С:УТ 10.3.
//
// Имена entity set'ов и полей ниже — лучшее приближение по стандартному
// именованию 1С OData (Ref_Key/Parent_Key/Description для справочников,
// суффикс Balance для регистров накопления). На момент написания сервер
// PVISERVER ещё не публикует OData (см. back/integrations/README.md) —
// сверить с реальными метаданными после публикации.
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

type PriceDTO struct {
	ProductKey   string  `json:"Номенклатура_Key"`
	PriceTypeKey string  `json:"ТипЦен_Key"`
	Price        float64 `json:"Цена"`
}

type StockDTO struct {
	ProductKey   string  `json:"Номенклатура_Key"`
	WarehouseKey string  `json:"Склад_Key"`
	Quantity     float64 `json:"КоличествоBalance"`
}
