// back/admin/internal/storage/postgres/signup_requests.go
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

//go:embed sql/insertSignupRequest.sql
var sqlInsertSignupRequest string

//go:embed sql/listSignupRequests.sql
var sqlListSignupRequests string

//go:embed sql/getSignupRequest.sql
var sqlGetSignupRequest string

//go:embed sql/decideSignupRequest.sql
var sqlDecideSignupRequest string

var (
	ErrSignupRequestNotFound       = errors.New("signup request not found")
	ErrSignupRequestAlreadyDecided = errors.New("signup request already decided")
)

type SignupRequest struct {
	ID           uuid.UUID
	Company      string
	INN          string
	Contact      string
	Email        string
	Phone        string
	RequestType  string
	Status       string
	RejectReason string
	CreatedAt    time.Time
	DecidedAt    *time.Time
}

func scanSignupRequest(row rowScanner) (SignupRequest, error) {
	var request SignupRequest
	var decidedAt *time.Time
	err := row.Scan(
		&request.ID, &request.Company, &request.INN, &request.Contact,
		&request.Email, &request.Phone, &request.RequestType, &request.Status,
		&request.RejectReason, &request.CreatedAt, &decidedAt,
	)
	request.DecidedAt = decidedAt
	return request, err
}

func (s *Storage) CreateSignupRequest(ctx context.Context, request SignupRequest) (SignupRequest, error) {
	created, err := scanSignupRequest(s.pool.QueryRow(
		ctx, sqlInsertSignupRequest,
		request.ID, request.Company, request.INN, request.Contact,
		request.Email, request.Phone, request.RequestType,
	))
	if err != nil {
		return SignupRequest{}, fmt.Errorf("insert signup request: %w", err)
	}
	return created, nil
}

func (s *Storage) ListSignupRequests(ctx context.Context, status string) ([]SignupRequest, error) {
	rows, err := s.pool.Query(ctx, sqlListSignupRequests, status)
	if err != nil {
		return nil, fmt.Errorf("list signup requests: %w", err)
	}
	defer rows.Close()

	items := make([]SignupRequest, 0)
	for rows.Next() {
		item, scanErr := scanSignupRequest(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan signup request: %w", scanErr)
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate signup requests: %w", err)
	}
	return items, nil
}

func (s *Storage) signupRequestExists(ctx context.Context, id uuid.UUID) (bool, error) {
	_, err := scanSignupRequest(s.pool.QueryRow(ctx, sqlGetSignupRequest, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("get signup request: %w", err)
	}
	return true, nil
}

func (s *Storage) DecideSignupRequest(ctx context.Context, id uuid.UUID, status string, rejectReason string) (SignupRequest, error) {
	updated, err := scanSignupRequest(s.pool.QueryRow(ctx, sqlDecideSignupRequest, id, status, rejectReason))
	if errors.Is(err, pgx.ErrNoRows) {
		exists, existsErr := s.signupRequestExists(ctx, id)
		if existsErr != nil {
			return SignupRequest{}, existsErr
		}
		if exists {
			return SignupRequest{}, ErrSignupRequestAlreadyDecided
		}
		return SignupRequest{}, ErrSignupRequestNotFound
	}
	if err != nil {
		return SignupRequest{}, fmt.Errorf("decide signup request: %w", err)
	}
	return updated, nil
}
