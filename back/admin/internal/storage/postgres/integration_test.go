//go:build integration

package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost port=5432 dbname=AMC sslmode=disable user=mbatimel password=mbatimel"
	}

	pool, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		t.Skipf("postgres not reachable, skipping integration test: %v", err)
	}
	return pool
}

func TestAuditLogRoundTrip(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()

	ctx := context.Background()
	storage := New(pool)

	actorID := uuid.New()

	if err := storage.InsertAuditLogEntry(ctx, actorID, "Админ портала", "Выполнен вход в систему"); err != nil {
		t.Fatalf("InsertAuditLogEntry (login) failed: %v", err)
	}
	if err := storage.InsertAuditLogEntry(ctx, actorID, "Админ портала", "Выполнен выход из системы"); err != nil {
		t.Fatalf("InsertAuditLogEntry (logout) failed: %v", err)
	}

	entries, err := storage.ListAuditLogEntries(ctx, 10, 0)
	if err != nil {
		t.Fatalf("ListAuditLogEntries failed: %v", err)
	}
	if len(entries) < 2 {
		t.Fatalf("ListAuditLogEntries returned %d entries, want at least 2", len(entries))
	}
	if entries[0].Action != "Выполнен выход из системы" {
		t.Fatalf("ListAuditLogEntries[0].Action = %q, want the most recently inserted entry first", entries[0].Action)
	}
	if entries[0].ActorUserID != actorID {
		t.Fatalf("ListAuditLogEntries[0].ActorUserID = %v, want %v", entries[0].ActorUserID, actorID)
	}

	total, err := storage.CountAuditLogEntries(ctx)
	if err != nil {
		t.Fatalf("CountAuditLogEntries failed: %v", err)
	}
	if total < 2 {
		t.Fatalf("CountAuditLogEntries = %d, want at least 2", total)
	}
}
