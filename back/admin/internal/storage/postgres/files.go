// back/admin/internal/storage/postgres/files.go
package postgres

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

//go:embed sql/insertFile.sql
var sqlInsertFile string

//go:embed sql/deleteFiles.sql
var sqlDeleteFiles string

// insertFile stores generic file metadata shared by banners, legal docs and
// certificates.
func insertFile(ctx context.Context, tx pgx.Tx, objectKey string, originalName string, contentType string, sizeBytes int64) (uuid.UUID, error) {
	var fileID uuid.UUID
	if err := tx.QueryRow(ctx, sqlInsertFile, objectKey, originalName, contentType, sizeBytes).Scan(&fileID); err != nil {
		return uuid.Nil, fmt.Errorf("insert file: %w", err)
	}
	return fileID, nil
}

// deleteFiles removes file metadata rows for the given ids. Deleting an
// empty slice is a no-op.
func deleteFiles(ctx context.Context, tx pgx.Tx, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	if _, err := tx.Exec(ctx, sqlDeleteFiles, ids); err != nil {
		return fmt.Errorf("delete files: %w", err)
	}
	return nil
}
