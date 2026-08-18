// back/admin/internal/storage/postgres/legal_docs.go
package postgres

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

//go:embed sql/listLegalDocs.sql
var sqlListLegalDocs string

//go:embed sql/getLegalDoc.sql
var sqlGetLegalDoc string

//go:embed sql/insertLegalDoc.sql
var sqlInsertLegalDoc string

//go:embed sql/insertLegalDocVersion.sql
var sqlInsertLegalDocVersion string

//go:embed sql/updateLegalDocCurrentVersion.sql
var sqlUpdateLegalDocCurrentVersion string

//go:embed sql/listLegalDocVersions.sql
var sqlListLegalDocVersions string

//go:embed sql/deleteLegalDoc.sql
var sqlDeleteLegalDoc string

var (
	ErrLegalDocNotFound      = errors.New("legal document not found")
	ErrLegalDocAlreadyExists = errors.New("legal document already exists")
)

type LegalDoc struct {
	ID             string
	Name           string
	Body           string
	CurrentVersion string
	UpdatedAt      time.Time
	FileID         uuid.UUID
	ObjectKey      string
	OriginalName   string
	ContentType    string
	SizeBytes      int64
}

type LegalDocVersion struct {
	ID           uuid.UUID
	DocID        string
	Version      string
	Summary      string
	Author       string
	CreatedAt    time.Time
	FileID       uuid.UUID
	ObjectKey    string
	OriginalName string
	ContentType  string
	SizeBytes    int64
}

func scanLegalDoc(row rowScanner) (LegalDoc, error) {
	var doc LegalDoc
	var fileID uuid.NullUUID
	err := row.Scan(
		&doc.ID, &doc.Name, &doc.Body, &doc.CurrentVersion, &doc.UpdatedAt,
		&fileID, &doc.ObjectKey, &doc.OriginalName, &doc.ContentType, &doc.SizeBytes,
	)
	if fileID.Valid {
		doc.FileID = fileID.UUID
	}
	return doc, err
}

func scanLegalDocVersion(row rowScanner) (LegalDocVersion, error) {
	var version LegalDocVersion
	var fileID uuid.NullUUID
	err := row.Scan(
		&version.ID, &version.DocID, &version.Version, &version.Summary, &version.Author, &version.CreatedAt,
		&fileID, &version.ObjectKey, &version.OriginalName, &version.ContentType, &version.SizeBytes,
	)
	if fileID.Valid {
		version.FileID = fileID.UUID
	}
	return version, err
}

func (s *Storage) ListLegalDocs(ctx context.Context) ([]LegalDoc, error) {
	rows, err := s.pool.Query(ctx, sqlListLegalDocs)
	if err != nil {
		return nil, fmt.Errorf("list legal docs: %w", err)
	}
	defer rows.Close()
	items := make([]LegalDoc, 0)
	for rows.Next() {
		doc, scanErr := scanLegalDoc(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan legal doc: %w", scanErr)
		}
		items = append(items, doc)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate legal docs: %w", err)
	}
	return items, nil
}

func (s *Storage) GetLegalDoc(ctx context.Context, docID string) (LegalDoc, error) {
	doc, err := scanLegalDoc(s.pool.QueryRow(ctx, sqlGetLegalDoc, docID))
	if errors.Is(err, pgx.ErrNoRows) {
		return LegalDoc{}, ErrLegalDocNotFound
	}
	if err != nil {
		return LegalDoc{}, fmt.Errorf("get legal doc: %w", err)
	}
	return doc, nil
}

// CreateLegalDoc inserts a new document together with its file and first
// version record in a single transaction.
func (s *Storage) CreateLegalDoc(ctx context.Context, doc LegalDoc, version LegalDocVersion) (LegalDoc, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return LegalDoc{}, fmt.Errorf("begin create legal doc: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	fileID, err := insertFile(ctx, tx, doc.ObjectKey, doc.OriginalName, doc.ContentType, doc.SizeBytes)
	if err != nil {
		return LegalDoc{}, err
	}
	if _, err = tx.Exec(ctx, sqlInsertLegalDoc, doc.ID, doc.Name, doc.Body, doc.CurrentVersion, fileID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return LegalDoc{}, ErrLegalDocAlreadyExists
		}
		return LegalDoc{}, fmt.Errorf("insert legal doc: %w", err)
	}
	if _, err = tx.Exec(ctx, sqlInsertLegalDocVersion, doc.ID, version.Version, version.Summary, version.Author, fileID); err != nil {
		return LegalDoc{}, fmt.Errorf("insert legal doc version: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return LegalDoc{}, fmt.Errorf("commit create legal doc: %w", err)
	}
	return s.GetLegalDoc(ctx, doc.ID)
}

// ReplaceLegalDocFile uploads a new version of an existing document: it
// keeps every previous file/version so the history can never be lost, and
// only repoints the document's "current" pointer.
func (s *Storage) ReplaceLegalDocFile(ctx context.Context, docID string, version LegalDocVersion) (LegalDoc, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return LegalDoc{}, fmt.Errorf("begin replace legal doc file: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	fileID, err := insertFile(ctx, tx, version.ObjectKey, version.OriginalName, version.ContentType, version.SizeBytes)
	if err != nil {
		return LegalDoc{}, err
	}
	if _, err = tx.Exec(ctx, sqlInsertLegalDocVersion, docID, version.Version, version.Summary, version.Author, fileID); err != nil {
		return LegalDoc{}, fmt.Errorf("insert legal doc version: %w", err)
	}
	tag, err := tx.Exec(ctx, sqlUpdateLegalDocCurrentVersion, docID, version.Version, fileID)
	if err != nil {
		return LegalDoc{}, fmt.Errorf("update legal doc current version: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return LegalDoc{}, ErrLegalDocNotFound
	}
	if err = tx.Commit(ctx); err != nil {
		return LegalDoc{}, fmt.Errorf("commit replace legal doc file: %w", err)
	}
	return s.GetLegalDoc(ctx, docID)
}

func (s *Storage) ListLegalDocVersions(ctx context.Context, docID string) ([]LegalDocVersion, error) {
	rows, err := s.pool.Query(ctx, sqlListLegalDocVersions, docID)
	if err != nil {
		return nil, fmt.Errorf("list legal doc versions: %w", err)
	}
	defer rows.Close()
	items := make([]LegalDocVersion, 0)
	for rows.Next() {
		version, scanErr := scanLegalDocVersion(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan legal doc version: %w", scanErr)
		}
		items = append(items, version)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate legal doc versions: %w", err)
	}
	return items, nil
}

// DeleteLegalDoc removes the document, its version history (cascaded by the
// database) and every file row those versions reference.
func (s *Storage) DeleteLegalDoc(ctx context.Context, docID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin delete legal doc: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rows, err := tx.Query(ctx, `SELECT file_id FROM portal_legal_doc_versions WHERE doc_id = $1 AND file_id IS NOT NULL`, docID)
	if err != nil {
		return fmt.Errorf("select legal doc file ids: %w", err)
	}
	fileIDs := make([]uuid.UUID, 0)
	for rows.Next() {
		var fileID uuid.UUID
		if scanErr := rows.Scan(&fileID); scanErr != nil {
			rows.Close()
			return fmt.Errorf("scan legal doc file id: %w", scanErr)
		}
		fileIDs = append(fileIDs, fileID)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return fmt.Errorf("iterate legal doc file ids: %w", err)
	}

	tag, err := tx.Exec(ctx, sqlDeleteLegalDoc, docID)
	if err != nil {
		return fmt.Errorf("delete legal doc: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrLegalDocNotFound
	}
	if err = deleteFiles(ctx, tx, fileIDs); err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit delete legal doc: %w", err)
	}
	return nil
}
