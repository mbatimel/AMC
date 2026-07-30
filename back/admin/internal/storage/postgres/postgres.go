package postgres

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

//go:embed sql/insertAuditLogEntry.sql
var sqlInsertAuditLogEntry string

//go:embed sql/listAuditLogEntries.sql
var sqlListAuditLogEntries string

//go:embed sql/countAuditLogEntries.sql
var sqlCountAuditLogEntries string

type AuditLogEntry struct {
	ID          uuid.UUID
	ActorUserID uuid.UUID
	ActorLabel  string
	Action      string
	CreatedAt   time.Time
}

type Storage struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Storage {
	return &Storage{pool: pool}
}

func (s *Storage) InsertAuditLogEntry(ctx context.Context, actorUserID uuid.UUID, actorLabel string, action string) error {
	if _, err := s.pool.Exec(ctx, sqlInsertAuditLogEntry, actorUserID, actorLabel, action); err != nil {
		return fmt.Errorf("insert audit log entry: %w", err)
	}
	return nil
}

func (s *Storage) ListAuditLogEntries(ctx context.Context, limit int, offset int) ([]AuditLogEntry, error) {
	rows, err := s.pool.Query(ctx, sqlListAuditLogEntries, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list audit log entries: %w", err)
	}
	defer rows.Close()

	entries := make([]AuditLogEntry, 0)
	for rows.Next() {
		var e AuditLogEntry
		if err = rows.Scan(&e.ID, &e.ActorUserID, &e.ActorLabel, &e.Action, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan audit log entry: %w", err)
		}
		entries = append(entries, e)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit log entries: %w", err)
	}

	return entries, nil
}

func (s *Storage) CountAuditLogEntries(ctx context.Context) (int, error) {
	var total int
	if err := s.pool.QueryRow(ctx, sqlCountAuditLogEntries).Scan(&total); err != nil {
		return 0, fmt.Errorf("count audit log entries: %w", err)
	}
	return total, nil
}
