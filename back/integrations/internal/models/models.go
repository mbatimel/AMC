package models

import "github.com/google/uuid"

type SyncLogLevel string

const (
	SyncLogInfo  SyncLogLevel = "info"
	SyncLogWarn  SyncLogLevel = "warn"
	SyncLogError SyncLogLevel = "error"
)

type CategoryInput struct {
	OneCGUID uuid.UUID
	Name     string
}

type WarehouseInput struct {
	OneCGUID uuid.UUID
	Name     string
}

type ProductInput struct {
	OneCGUID   uuid.UUID
	CategoryID *uuid.UUID
	SKU        string
	Name       string
}

type PriceInput struct {
	ProductID uuid.UUID
	PriceType string
	Price     float64
}

type StockInput struct {
	ProductID   uuid.UUID
	WarehouseID uuid.UUID
	Quantity    float64
}
