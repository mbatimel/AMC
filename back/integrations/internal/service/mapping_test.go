package service

import (
	"testing"

	"github.com/google/uuid"

	"github.com/mbatimel/AMC/integrations/internal/onec"
)

func TestMapCategory_WithParent(t *testing.T) {
	dto := onec.CategoryDTO{
		RefKey:      "11111111-1111-1111-1111-111111111111",
		ParentKey:   "22222222-2222-2222-2222-222222222222",
		Description: "Дрели",
	}
	in, parent, err := mapCategory(dto)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in.OneCGUID.String() != dto.RefKey || in.Name != "Дрели" {
		t.Fatalf("unexpected input: %+v", in)
	}
	if parent == nil || parent.String() != dto.ParentKey {
		t.Fatalf("expected parent %s, got %v", dto.ParentKey, parent)
	}
}

func TestMapCategory_ZeroParentIsRoot(t *testing.T) {
	dto := onec.CategoryDTO{
		RefKey:      "11111111-1111-1111-1111-111111111111",
		ParentKey:   zeroGUID,
		Description: "Инструмент",
	}
	_, parent, err := mapCategory(dto)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parent != nil {
		t.Fatalf("expected nil parent for zero guid, got %v", parent)
	}
}

func TestMapCategory_InvalidRef(t *testing.T) {
	dto := onec.CategoryDTO{RefKey: "not-a-guid", Description: "x"}
	_, _, err := mapCategory(dto)
	if err == nil {
		t.Fatal("expected error for invalid ref")
	}
}

func TestMapProduct_ResolvesCategory(t *testing.T) {
	categoryRef := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	categoryID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	categoryIDs := map[uuid.UUID]uuid.UUID{categoryRef: categoryID}

	dto := onec.ProductDTO{
		RefKey:      "11111111-1111-1111-1111-111111111111",
		CategoryKey: categoryRef.String(),
		Code:        "SKU-1",
		Description: "Дрель",
	}
	in, err := mapProduct(dto, categoryIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in.CategoryID == nil || *in.CategoryID != categoryID {
		t.Fatalf("expected category %s, got %v", categoryID, in.CategoryID)
	}
	if in.SKU != "SKU-1" || in.Name != "Дрель" {
		t.Fatalf("unexpected input: %+v", in)
	}
}

func TestMapProduct_UnknownCategory_LeavesNil(t *testing.T) {
	dto := onec.ProductDTO{
		RefKey:      "11111111-1111-1111-1111-111111111111",
		CategoryKey: "99999999-9999-9999-9999-999999999999",
		Code:        "SKU-1",
		Description: "Дрель",
	}
	in, err := mapProduct(dto, map[uuid.UUID]uuid.UUID{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in.CategoryID != nil {
		t.Fatalf("expected nil category, got %v", in.CategoryID)
	}
}

func TestMapProduct_InvalidRef(t *testing.T) {
	dto := onec.ProductDTO{RefKey: "not-a-guid"}
	_, err := mapProduct(dto, map[uuid.UUID]uuid.UUID{})
	if err == nil {
		t.Fatal("expected error for invalid ref")
	}
}

func TestMapPrice_SkipsUnknownProduct(t *testing.T) {
	dto := onec.PriceDTO{ProductKey: "11111111-1111-1111-1111-111111111111", PriceTypeKey: "x", Price: 10}
	_, ok, err := mapPrice(dto, map[uuid.UUID]uuid.UUID{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected ok=false for unknown product")
	}
}

func TestMapPrice_ResolvesProduct(t *testing.T) {
	productRef := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	productID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	dto := onec.PriceDTO{ProductKey: productRef.String(), PriceTypeKey: "type-a", Price: 99.5}
	in, ok, err := mapPrice(dto, map[uuid.UUID]uuid.UUID{productRef: productID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok || in.ProductID != productID || in.PriceType != "type-a" || in.Price != 99.5 {
		t.Fatalf("unexpected result: %+v", in)
	}
}

func TestMapStock_ResolvesProductAndWarehouse(t *testing.T) {
	productRef := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	productID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	warehouseRef := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	warehouseID := uuid.MustParse("66666666-6666-6666-6666-666666666666")

	dto := onec.StockDTO{ProductKey: productRef.String(), WarehouseKey: warehouseRef.String(), Quantity: 7}
	in, ok, err := mapStock(dto,
		map[uuid.UUID]uuid.UUID{productRef: productID},
		map[uuid.UUID]uuid.UUID{warehouseRef: warehouseID},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok || in.ProductID != productID || in.WarehouseID != warehouseID || in.Quantity != 7 {
		t.Fatalf("unexpected result: %+v", in)
	}
}

func TestMapStock_SkipsUnknownWarehouse(t *testing.T) {
	productRef := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	productID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	dto := onec.StockDTO{ProductKey: productRef.String(), WarehouseKey: "77777777-7777-7777-7777-777777777777", Quantity: 3}
	_, ok, err := mapStock(dto, map[uuid.UUID]uuid.UUID{productRef: productID}, map[uuid.UUID]uuid.UUID{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected ok=false for unknown warehouse")
	}
}
