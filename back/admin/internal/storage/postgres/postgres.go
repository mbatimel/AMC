package postgres

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

//go:embed sql/insertAuditLogEntry.sql
var sqlInsertAuditLogEntry string

//go:embed sql/listAuditLogEntries.sql
var sqlListAuditLogEntries string

//go:embed sql/countAuditLogEntries.sql
var sqlCountAuditLogEntries string

//go:embed sql/listBanners.sql
var sqlListBanners string

//go:embed sql/getBanner.sql
var sqlGetBanner string

//go:embed sql/insertBannerFile.sql
var sqlInsertBannerFile string

//go:embed sql/insertBanner.sql
var sqlInsertBanner string

//go:embed sql/updateBanner.sql
var sqlUpdateBanner string

//go:embed sql/deleteBanner.sql
var sqlDeleteBanner string

//go:embed sql/getBannerDelay.sql
var sqlGetBannerDelay string

//go:embed sql/updateBannerDelay.sql
var sqlUpdateBannerDelay string

var ErrBannerNotFound = errors.New("banner not found")

type AuditLogEntry struct {
	ID          uuid.UUID
	ActorUserID uuid.UUID
	ActorLabel  string
	Action      string
	CreatedAt   time.Time
}

type Banner struct {
	ID           uuid.UUID
	Title        string
	Subtitle     string
	ImageURL     string
	Link         string
	SortOrder    int
	IsActive     bool
	DateFrom     *time.Time
	DateTo       *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	FileID       uuid.UUID
	ObjectKey    string
	OriginalName string
	ContentType  string
	SizeBytes    int64
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

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanBanner(row rowScanner) (Banner, error) {
	var banner Banner
	var dateFrom, dateTo sql.NullTime
	var fileID uuid.NullUUID
	err := row.Scan(
		&banner.ID, &banner.Title, &banner.Subtitle, &banner.ImageURL,
		&banner.Link, &banner.SortOrder, &banner.IsActive, &dateFrom, &dateTo,
		&banner.CreatedAt, &banner.UpdatedAt, &fileID, &banner.ObjectKey,
		&banner.OriginalName, &banner.ContentType, &banner.SizeBytes,
	)
	if dateFrom.Valid {
		value := dateFrom.Time
		banner.DateFrom = &value
	}
	if dateTo.Valid {
		value := dateTo.Time
		banner.DateTo = &value
	}
	if fileID.Valid {
		banner.FileID = fileID.UUID
	}
	return banner, err
}

func (s *Storage) ListBanners(ctx context.Context, visibleOnly bool) ([]Banner, int, error) {
	rows, err := s.pool.Query(ctx, sqlListBanners, visibleOnly)
	if err != nil {
		return nil, 0, fmt.Errorf("list banners: %w", err)
	}
	defer rows.Close()
	items := make([]Banner, 0)
	for rows.Next() {
		banner, scanErr := scanBanner(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("scan banner: %w", scanErr)
		}
		items = append(items, banner)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate banners: %w", err)
	}
	var delay int
	if err = s.pool.QueryRow(ctx, sqlGetBannerDelay).Scan(&delay); err != nil {
		return nil, 0, fmt.Errorf("get banner delay: %w", err)
	}
	return items, delay, nil
}

func (s *Storage) GetBanner(ctx context.Context, bannerID uuid.UUID) (Banner, error) {
	banner, err := scanBanner(s.pool.QueryRow(ctx, sqlGetBanner, bannerID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Banner{}, ErrBannerNotFound
	}
	if err != nil {
		return Banner{}, fmt.Errorf("get banner: %w", err)
	}
	return banner, nil
}

func insertBannerFile(ctx context.Context, tx pgx.Tx, banner Banner) (uuid.UUID, error) {
	var fileID uuid.UUID
	if err := tx.QueryRow(ctx, sqlInsertBannerFile, banner.ObjectKey, banner.OriginalName, banner.ContentType, banner.SizeBytes).Scan(&fileID); err != nil {
		return uuid.Nil, fmt.Errorf("insert banner file: %w", err)
	}
	return fileID, nil
}

func (s *Storage) CreateBanner(ctx context.Context, banner Banner) (Banner, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Banner{}, fmt.Errorf("begin create banner: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	fileID, err := insertBannerFile(ctx, tx, banner)
	if err != nil {
		return Banner{}, err
	}
	if _, err = tx.Exec(ctx, sqlInsertBanner, banner.ID, banner.Title, banner.Subtitle, banner.ImageURL, banner.Link, banner.SortOrder, banner.IsActive, banner.DateFrom, banner.DateTo, fileID); err != nil {
		return Banner{}, fmt.Errorf("insert banner: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return Banner{}, fmt.Errorf("commit create banner: %w", err)
	}
	return s.GetBanner(ctx, banner.ID)
}

func (s *Storage) UpdateBanner(ctx context.Context, banner Banner, replaceImage bool) (Banner, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Banner{}, fmt.Errorf("begin update banner: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	fileID := uuid.Nil
	if replaceImage {
		fileID, err = insertBannerFile(ctx, tx, banner)
		if err != nil {
			return Banner{}, err
		}
	}
	tag, err := tx.Exec(ctx, sqlUpdateBanner, banner.ID, banner.Title, banner.Subtitle, banner.ImageURL, banner.Link, banner.SortOrder, banner.IsActive, banner.DateFrom, banner.DateTo, replaceImage, fileID)
	if err != nil {
		return Banner{}, fmt.Errorf("update banner: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return Banner{}, ErrBannerNotFound
	}
	if replaceImage && banner.FileID != uuid.Nil {
		if _, err = tx.Exec(ctx, `DELETE FROM files WHERE id = $1`, banner.FileID); err != nil {
			return Banner{}, fmt.Errorf("delete replaced banner file metadata: %w", err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return Banner{}, fmt.Errorf("commit update banner: %w", err)
	}
	return s.GetBanner(ctx, banner.ID)
}

func (s *Storage) DeleteBanner(ctx context.Context, bannerID uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin delete banner: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var fileID uuid.NullUUID
	if err = tx.QueryRow(ctx, `SELECT file_id FROM portal_banners WHERE id = $1`, bannerID).Scan(&fileID); errors.Is(err, pgx.ErrNoRows) {
		return ErrBannerNotFound
	} else if err != nil {
		return fmt.Errorf("select banner file: %w", err)
	}
	if _, err = tx.Exec(ctx, sqlDeleteBanner, bannerID); err != nil {
		return fmt.Errorf("delete banner: %w", err)
	}
	if fileID.Valid {
		if _, err = tx.Exec(ctx, `DELETE FROM files WHERE id = $1`, fileID.UUID); err != nil {
			return fmt.Errorf("delete banner file metadata: %w", err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit delete banner: %w", err)
	}
	return nil
}

func (s *Storage) UpdateBannerDelay(ctx context.Context, delaySec int) error {
	if _, err := s.pool.Exec(ctx, sqlUpdateBannerDelay, delaySec); err != nil {
		return fmt.Errorf("update banner delay: %w", err)
	}
	return nil
}
