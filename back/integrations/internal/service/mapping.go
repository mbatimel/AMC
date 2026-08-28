package service

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/mbatimel/AMC/integrations/internal/models"
	"github.com/mbatimel/AMC/integrations/internal/onec"
)

const zeroGUID = "00000000-0000-0000-0000-000000000000"

func mapCategory(dto onec.CategoryDTO) (models.CategoryInput, *uuid.UUID, error) {
	ref, err := uuid.Parse(dto.RefKey)
	if err != nil {
		return models.CategoryInput{}, nil, fmt.Errorf("parse category ref %q: %w", dto.RefKey, err)
	}
	var parent *uuid.UUID
	if dto.ParentKey != "" && dto.ParentKey != zeroGUID {
		parentRef, parseErr := uuid.Parse(dto.ParentKey)
		if parseErr != nil {
			return models.CategoryInput{}, nil, fmt.Errorf("parse category parent %q: %w", dto.ParentKey, parseErr)
		}
		parent = &parentRef
	}
	return models.CategoryInput{OneCGUID: ref, Name: dto.Description}, parent, nil
}

func mapWarehouse(dto onec.WarehouseDTO) (models.WarehouseInput, error) {
	ref, err := uuid.Parse(dto.RefKey)
	if err != nil {
		return models.WarehouseInput{}, fmt.Errorf("parse warehouse ref %q: %w", dto.RefKey, err)
	}
	return models.WarehouseInput{OneCGUID: ref, Name: dto.Description}, nil
}

func mapProduct(dto onec.ProductDTO, categoryIDs map[uuid.UUID]uuid.UUID) (models.ProductInput, error) {
	ref, err := uuid.Parse(dto.RefKey)
	if err != nil {
		return models.ProductInput{}, fmt.Errorf("parse product ref %q: %w", dto.RefKey, err)
	}
	var categoryID *uuid.UUID
	if dto.CategoryKey != "" && dto.CategoryKey != zeroGUID {
		categoryRef, parseErr := uuid.Parse(dto.CategoryKey)
		if parseErr != nil {
			return models.ProductInput{}, fmt.Errorf("parse product category %q: %w", dto.CategoryKey, parseErr)
		}
		if id, ok := categoryIDs[categoryRef]; ok {
			categoryID = &id
		}
	}
	return models.ProductInput{OneCGUID: ref, CategoryID: categoryID, SKU: dto.Code, Name: dto.Description}, nil
}

func mapPrice(dto onec.PriceDTO, productIDs map[uuid.UUID]uuid.UUID) (models.PriceInput, bool, error) {
	productRef, err := uuid.Parse(dto.ProductKey)
	if err != nil {
		return models.PriceInput{}, false, fmt.Errorf("parse price product ref %q: %w", dto.ProductKey, err)
	}
	productID, ok := productIDs[productRef]
	if !ok {
		return models.PriceInput{}, false, nil
	}
	return models.PriceInput{ProductID: productID, PriceType: dto.PriceTypeKey, Price: dto.Price}, true, nil
}

func mapStock(dto onec.StockDTO, productIDs, warehouseIDs map[uuid.UUID]uuid.UUID) (models.StockInput, bool, error) {
	productRef, err := uuid.Parse(dto.ProductKey)
	if err != nil {
		return models.StockInput{}, false, fmt.Errorf("parse stock product ref %q: %w", dto.ProductKey, err)
	}
	warehouseRef, err := uuid.Parse(dto.WarehouseKey)
	if err != nil {
		return models.StockInput{}, false, fmt.Errorf("parse stock warehouse ref %q: %w", dto.WarehouseKey, err)
	}
	productID, okP := productIDs[productRef]
	warehouseID, okW := warehouseIDs[warehouseRef]
	if !okP || !okW {
		return models.StockInput{}, false, nil
	}
	return models.StockInput{ProductID: productID, WarehouseID: warehouseID, Quantity: dto.Quantity}, true, nil
}
