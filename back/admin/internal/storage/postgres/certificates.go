// back/admin/internal/storage/postgres/certificates.go
package postgres

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

//go:embed sql/listCertificates.sql
var sqlListCertificates string

//go:embed sql/getCertificate.sql
var sqlGetCertificate string

//go:embed sql/insertCertificate.sql
var sqlInsertCertificate string

//go:embed sql/updateCertificate.sql
var sqlUpdateCertificate string

//go:embed sql/deleteCertificate.sql
var sqlDeleteCertificate string

var ErrCertificateNotFound = errors.New("certificate not found")

type Certificate struct {
	ID           uuid.UUID
	Title        string
	SortOrder    int
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	FileID       uuid.UUID
	ObjectKey    string
	OriginalName string
	ContentType  string
	SizeBytes    int64
}

func scanCertificate(row rowScanner) (Certificate, error) {
	var cert Certificate
	var fileID uuid.NullUUID
	err := row.Scan(
		&cert.ID, &cert.Title, &cert.SortOrder, &cert.IsActive, &cert.CreatedAt, &cert.UpdatedAt,
		&fileID, &cert.ObjectKey, &cert.OriginalName, &cert.ContentType, &cert.SizeBytes,
	)
	if fileID.Valid {
		cert.FileID = fileID.UUID
	}
	return cert, err
}

func (s *Storage) ListCertificates(ctx context.Context, visibleOnly bool) ([]Certificate, error) {
	rows, err := s.pool.Query(ctx, sqlListCertificates, visibleOnly)
	if err != nil {
		return nil, fmt.Errorf("list certificates: %w", err)
	}
	defer rows.Close()
	items := make([]Certificate, 0)
	for rows.Next() {
		cert, scanErr := scanCertificate(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan certificate: %w", scanErr)
		}
		items = append(items, cert)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate certificates: %w", err)
	}
	return items, nil
}

func (s *Storage) GetCertificate(ctx context.Context, certID uuid.UUID) (Certificate, error) {
	cert, err := scanCertificate(s.pool.QueryRow(ctx, sqlGetCertificate, certID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Certificate{}, ErrCertificateNotFound
	}
	if err != nil {
		return Certificate{}, fmt.Errorf("get certificate: %w", err)
	}
	return cert, nil
}

func (s *Storage) CreateCertificate(ctx context.Context, cert Certificate) (Certificate, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Certificate{}, fmt.Errorf("begin create certificate: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	fileID, err := insertFile(ctx, tx, cert.ObjectKey, cert.OriginalName, cert.ContentType, cert.SizeBytes)
	if err != nil {
		return Certificate{}, err
	}
	if _, err = tx.Exec(ctx, sqlInsertCertificate, cert.ID, cert.Title, cert.SortOrder, cert.IsActive, fileID); err != nil {
		return Certificate{}, fmt.Errorf("insert certificate: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return Certificate{}, fmt.Errorf("commit create certificate: %w", err)
	}
	return s.GetCertificate(ctx, cert.ID)
}

func (s *Storage) UpdateCertificate(ctx context.Context, cert Certificate, replaceFile bool) (Certificate, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Certificate{}, fmt.Errorf("begin update certificate: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	fileID := uuid.Nil
	if replaceFile {
		fileID, err = insertFile(ctx, tx, cert.ObjectKey, cert.OriginalName, cert.ContentType, cert.SizeBytes)
		if err != nil {
			return Certificate{}, err
		}
	}
	tag, err := tx.Exec(ctx, sqlUpdateCertificate, cert.ID, cert.Title, cert.SortOrder, cert.IsActive, replaceFile, fileID)
	if err != nil {
		return Certificate{}, fmt.Errorf("update certificate: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return Certificate{}, ErrCertificateNotFound
	}
	if replaceFile && cert.FileID != uuid.Nil {
		if err = deleteFiles(ctx, tx, []uuid.UUID{cert.FileID}); err != nil {
			return Certificate{}, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return Certificate{}, fmt.Errorf("commit update certificate: %w", err)
	}
	return s.GetCertificate(ctx, cert.ID)
}

func (s *Storage) DeleteCertificate(ctx context.Context, certID uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin delete certificate: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var fileID uuid.NullUUID
	if err = tx.QueryRow(ctx, `SELECT file_id FROM portal_certificates WHERE id = $1`, certID).Scan(&fileID); errors.Is(err, pgx.ErrNoRows) {
		return ErrCertificateNotFound
	} else if err != nil {
		return fmt.Errorf("select certificate file: %w", err)
	}
	if _, err = tx.Exec(ctx, sqlDeleteCertificate, certID); err != nil {
		return fmt.Errorf("delete certificate: %w", err)
	}
	if fileID.Valid {
		if err = deleteFiles(ctx, tx, []uuid.UUID{fileID.UUID}); err != nil {
			return err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit delete certificate: %w", err)
	}
	return nil
}
