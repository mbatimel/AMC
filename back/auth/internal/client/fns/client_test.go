package fns

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
)

func testLogger() zerolog.Logger {
	return zerolog.Nop()
}

func TestCheckIndividual_ValidInn(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("inn") != "773208978609" || r.URL.Query().Get("key") != "test-key" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		w.Write([]byte(`{"Корректность":{"КонтрСумма":true,"Недействительный":false}}`))
	}))
	defer server.Close()

	c := New(server.URL, "test-key", testLogger())
	valid, err := c.CheckIndividual(context.Background(), "773208978609")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !valid {
		t.Fatal("expected valid=true")
	}
}

func TestCheckIndividual_InvalidChecksum(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"Корректность":{"КонтрСумма":false,"Недействительный":false}}`))
	}))
	defer server.Close()

	c := New(server.URL, "test-key", testLogger())
	valid, err := c.CheckIndividual(context.Background(), "773208978609")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valid {
		t.Fatal("expected valid=false")
	}
}

func TestCheckIndividual_MarkedInvalid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"Корректность":{"КонтрСумма":true,"Недействительный":true}}`))
	}))
	defer server.Close()

	c := New(server.URL, "test-key", testLogger())
	valid, err := c.CheckIndividual(context.Background(), "773208978609")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valid {
		t.Fatal("expected valid=false when Недействительный=true")
	}
}

func TestCheckIndividual_NullFields_FailClosed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"Корректность":{"КонтрСумма":null,"Недействительный":null}}`))
	}))
	defer server.Close()

	c := New(server.URL, "test-key", testLogger())
	_, err := c.CheckIndividual(context.Background(), "773208978609")
	if err == nil {
		t.Fatal("expected error on null fields")
	}
}

func TestCheckIndividual_NonOKStatus_FailClosed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := New(server.URL, "test-key", testLogger())
	_, err := c.CheckIndividual(context.Background(), "773208978609")
	if err == nil {
		t.Fatal("expected error on non-200 status")
	}
}

func TestCheckIndividual_MalformedJSON_FailClosed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`not json`))
	}))
	defer server.Close()

	c := New(server.URL, "test-key", testLogger())
	_, err := c.CheckIndividual(context.Background(), "773208978609")
	if err == nil {
		t.Fatal("expected error on malformed json")
	}
}
