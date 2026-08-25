package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/mbatimel/AMC/integrations/internal/models"
	"github.com/mbatimel/AMC/integrations/internal/onec"
)

type fakeOnecClient struct {
	categories []onec.CategoryDTO
	warehouses []onec.WarehouseDTO
	products   []onec.ProductDTO
	prices     []onec.PriceDTO
	stock      []onec.StockDTO

	categoriesErr error
}

func (f *fakeOnecClient) FetchCategories(context.Context) ([]onec.CategoryDTO, error) {
	return f.categories, f.categoriesErr
}
func (f *fakeOnecClient) FetchWarehouses(context.Context) ([]onec.WarehouseDTO, error) {
	return f.warehouses, nil
}
func (f *fakeOnecClient) FetchProducts(context.Context) ([]onec.ProductDTO, error) {
	return f.products, nil
}
func (f *fakeOnecClient) FetchPrices(context.Context) ([]onec.PriceDTO, error) {
	return f.prices, nil
}
func (f *fakeOnecClient) FetchStock(context.Context) ([]onec.StockDTO, error) {
	return f.stock, nil
}

type fakeStorage struct {
	nextID       int
	categoryIDs  map[uuid.UUID]uuid.UUID
	warehouseIDs map[uuid.UUID]uuid.UUID
	productIDs   map[uuid.UUID]uuid.UUID
	parents      map[uuid.UUID]uuid.UUID
	prices       []models.PriceInput
	stocks       []models.StockInput
	logs         []string
	finalStatus  string
	failSKU      string
}

func newFakeStorage() *fakeStorage {
	return &fakeStorage{
		categoryIDs:  map[uuid.UUID]uuid.UUID{},
		warehouseIDs: map[uuid.UUID]uuid.UUID{},
		productIDs:   map[uuid.UUID]uuid.UUID{},
		parents:      map[uuid.UUID]uuid.UUID{},
	}
}

func (f *fakeStorage) newID() uuid.UUID {
	f.nextID++
	return uuid.MustParse(fmt.Sprintf("99999999-9999-9999-9999-%012d", f.nextID))
}

func (f *fakeStorage) UpsertIntegrationSystem(context.Context, string, string) (uuid.UUID, error) {
	return uuid.MustParse("11111111-1111-1111-1111-111111111111"), nil
}
func (f *fakeStorage) CreateSyncJob(context.Context, uuid.UUID) (uuid.UUID, error) {
	return uuid.MustParse("22222222-2222-2222-2222-222222222222"), nil
}
func (f *fakeStorage) FinishSyncJob(_ context.Context, _ uuid.UUID, status string) error {
	f.finalStatus = status
	return nil
}
func (f *fakeStorage) AddSyncLog(_ context.Context, _, _ uuid.UUID, level models.SyncLogLevel, message string) error {
	f.logs = append(f.logs, string(level)+": "+message)
	return nil
}
func (f *fakeStorage) UpsertCategory(_ context.Context, in models.CategoryInput) (uuid.UUID, error) {
	id := f.newID()
	f.categoryIDs[in.OneCGUID] = id
	return id, nil
}
func (f *fakeStorage) SetCategoryParent(_ context.Context, id, parentID uuid.UUID) error {
	f.parents[id] = parentID
	return nil
}
func (f *fakeStorage) UpsertWarehouse(_ context.Context, in models.WarehouseInput) (uuid.UUID, error) {
	id := f.newID()
	f.warehouseIDs[in.OneCGUID] = id
	return id, nil
}
func (f *fakeStorage) UpsertProduct(_ context.Context, in models.ProductInput) (uuid.UUID, error) {
	if in.SKU == f.failSKU {
		return uuid.Nil, errors.New("duplicate sku")
	}
	id := f.newID()
	f.productIDs[in.OneCGUID] = id
	return id, nil
}
func (f *fakeStorage) UpsertProductPrice(_ context.Context, in models.PriceInput) error {
	f.prices = append(f.prices, in)
	return nil
}
func (f *fakeStorage) UpsertStockBalance(_ context.Context, in models.StockInput) error {
	f.stocks = append(f.stocks, in)
	return nil
}

const (
	parentGUID    = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	childGUID     = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	productGUID   = "cccccccc-cccc-cccc-cccc-cccccccccccc"
	warehouseGUID = "dddddddd-dddd-dddd-dddd-dddddddddddd"
)

func TestRunSync_HappyPath(t *testing.T) {
	onecClient := &fakeOnecClient{
		categories: []onec.CategoryDTO{
			{RefKey: parentGUID, ParentKey: zeroGUID, Description: "Инструмент"},
			{RefKey: childGUID, ParentKey: parentGUID, Description: "Дрели"},
		},
		warehouses: []onec.WarehouseDTO{{RefKey: warehouseGUID, Description: "Склад №1"}},
		products: []onec.ProductDTO{
			{RefKey: productGUID, CategoryKey: childGUID, Code: "SKU-1", Description: "Дрель"},
		},
		prices: []onec.PriceDTO{{ProductKey: productGUID, PriceTypeKey: "type-a", Price: 100}},
		stock:  []onec.StockDTO{{ProductKey: productGUID, WarehouseKey: warehouseGUID, Quantity: 5}},
	}
	storage := newFakeStorage()
	svc := New(zerolog.Nop(), onecClient, storage)

	if err := svc.RunSync(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if storage.finalStatus != "success" {
		t.Fatalf("expected success, got %s", storage.finalStatus)
	}
	if len(storage.prices) != 1 || storage.prices[0].Price != 100 {
		t.Fatalf("unexpected prices: %+v", storage.prices)
	}
	if len(storage.stocks) != 1 || storage.stocks[0].Quantity != 5 {
		t.Fatalf("unexpected stocks: %+v", storage.stocks)
	}
	childID := storage.categoryIDs[uuid.MustParse(childGUID)]
	parentID := storage.categoryIDs[uuid.MustParse(parentGUID)]
	if storage.parents[childID] != parentID {
		t.Fatalf("expected child category parent to be set to parent id")
	}
}

func TestRunSync_DuplicateSKU_SkippedButOthersContinue(t *testing.T) {
	onecClient := &fakeOnecClient{
		products: []onec.ProductDTO{
			{RefKey: "cccccccc-cccc-cccc-cccc-cccccccccccc", Code: "SKU-DUP", Description: "A"},
			{RefKey: "ffffffff-ffff-ffff-ffff-ffffffffffff", Code: "SKU-OK", Description: "B"},
		},
	}
	storage := newFakeStorage()
	storage.failSKU = "SKU-DUP"
	svc := New(zerolog.Nop(), onecClient, storage)

	if err := svc.RunSync(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if storage.finalStatus != "partial" {
		t.Fatalf("expected partial, got %s", storage.finalStatus)
	}
	if len(storage.productIDs) != 1 {
		t.Fatalf("expected 1 product synced, got %d", len(storage.productIDs))
	}
	found := false
	for _, l := range storage.logs {
		if strings.Contains(l, "SKU-DUP") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected sync log about duplicate sku, got %v", storage.logs)
	}
}

func TestRunSync_BadCategory_SkippedButOthersContinue(t *testing.T) {
	onecClient := &fakeOnecClient{
		categories: []onec.CategoryDTO{
			{RefKey: "not-a-guid", Description: "Broken"},
			{RefKey: parentGUID, ParentKey: zeroGUID, Description: "Инструмент"},
		},
	}
	storage := newFakeStorage()
	svc := New(zerolog.Nop(), onecClient, storage)

	if err := svc.RunSync(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if storage.finalStatus != "partial" {
		t.Fatalf("expected partial, got %s", storage.finalStatus)
	}
	if len(storage.categoryIDs) != 1 {
		t.Fatalf("expected 1 category synced, got %d", len(storage.categoryIDs))
	}
	if _, ok := storage.categoryIDs[uuid.MustParse(parentGUID)]; !ok {
		t.Fatalf("expected valid category to still be synced, got %v", storage.categoryIDs)
	}
	found := false
	for _, l := range storage.logs {
		if strings.Contains(l, "not-a-guid") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected sync log about bad category ref, got %v", storage.logs)
	}
}

func TestRunSync_BadWarehouse_SkippedButOthersContinue(t *testing.T) {
	onecClient := &fakeOnecClient{
		warehouses: []onec.WarehouseDTO{
			{RefKey: "not-a-guid", Description: "Broken"},
			{RefKey: warehouseGUID, Description: "Склад №1"},
		},
	}
	storage := newFakeStorage()
	svc := New(zerolog.Nop(), onecClient, storage)

	if err := svc.RunSync(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if storage.finalStatus != "partial" {
		t.Fatalf("expected partial, got %s", storage.finalStatus)
	}
	if len(storage.warehouseIDs) != 1 {
		t.Fatalf("expected 1 warehouse synced, got %d", len(storage.warehouseIDs))
	}
	if _, ok := storage.warehouseIDs[uuid.MustParse(warehouseGUID)]; !ok {
		t.Fatalf("expected valid warehouse to still be synced, got %v", storage.warehouseIDs)
	}
	found := false
	for _, l := range storage.logs {
		if strings.Contains(l, "not-a-guid") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected sync log about bad warehouse ref, got %v", storage.logs)
	}
}

func TestRunSync_CategoriesFetchError_ContinuesOtherSteps(t *testing.T) {
	onecClient := &fakeOnecClient{
		categoriesErr: errors.New("network down"),
		warehouses:    []onec.WarehouseDTO{{RefKey: warehouseGUID, Description: "Склад №1"}},
	}
	storage := newFakeStorage()
	svc := New(zerolog.Nop(), onecClient, storage)

	if err := svc.RunSync(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if storage.finalStatus != "partial" {
		t.Fatalf("expected partial, got %s", storage.finalStatus)
	}
	if len(storage.warehouseIDs) != 1 {
		t.Fatalf("expected warehouses step to still run, got %d", len(storage.warehouseIDs))
	}
	found := false
	for _, l := range storage.logs {
		if strings.Contains(l, "network down") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected sync log about categories error, got %v", storage.logs)
	}
}
