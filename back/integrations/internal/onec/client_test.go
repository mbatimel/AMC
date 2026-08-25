package onec

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
)

func testLogger() zerolog.Logger { return zerolog.Nop() }

func TestFetchCategories_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Catalog_НоменклатурныеГруппы" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Fatal("expected Authorization header")
		}
		w.Write([]byte(`{"value":[{"Ref_Key":"11111111-1111-1111-1111-111111111111","Parent_Key":"00000000-0000-0000-0000-000000000000","Description":"Инструмент"}]}`))
	}))
	defer server.Close()

	c := New(server.URL, "user", "pass", testLogger())
	got, err := c.FetchCategories(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Description != "Инструмент" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestFetchCategories_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	c := New(server.URL, "user", "pass", testLogger())
	_, err := c.FetchCategories(context.Background())
	if err == nil {
		t.Fatal("expected error on non-200 status")
	}
}

func TestFetchCategories_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`not json`))
	}))
	defer server.Close()

	c := New(server.URL, "user", "pass", testLogger())
	_, err := c.FetchCategories(context.Background())
	if err == nil {
		t.Fatal("expected error on malformed json")
	}
}

func TestFetchWarehouses_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"value":[{"Ref_Key":"44444444-4444-4444-4444-444444444444","Description":"Склад №1"}]}`))
	}))
	defer server.Close()

	c := New(server.URL, "user", "pass", testLogger())
	got, err := c.FetchWarehouses(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Description != "Склад №1" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestFetchProducts_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"value":[{"Ref_Key":"22222222-2222-2222-2222-222222222222","НоменклатурнаяГруппа_Key":"11111111-1111-1111-1111-111111111111","Code":"SKU-1","Description":"Дрель"}]}`))
	}))
	defer server.Close()

	c := New(server.URL, "user", "pass", testLogger())
	got, err := c.FetchProducts(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Code != "SKU-1" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestFetchPrices_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"value":[{"Номенклатура_Key":"22222222-2222-2222-2222-222222222222","ТипЦен_Key":"33333333-3333-3333-3333-333333333333","Цена":150.5}]}`))
	}))
	defer server.Close()

	c := New(server.URL, "user", "pass", testLogger())
	got, err := c.FetchPrices(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Price != 150.5 {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestFetchStock_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"value":[{"Номенклатура_Key":"22222222-2222-2222-2222-222222222222","Склад_Key":"44444444-4444-4444-4444-444444444444","КоличествоBalance":7}]}`))
	}))
	defer server.Close()

	c := New(server.URL, "user", "pass", testLogger())
	got, err := c.FetchStock(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Quantity != 7 {
		t.Fatalf("unexpected result: %+v", got)
	}
}
