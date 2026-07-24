package postgres

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

//go:embed sql/getCities.sql
var sqlGetCities string

type City struct {
	ID   uuid.UUID
	Name string
}

type Storage struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Storage {
	return &Storage{pool: pool}
}

func (s *Storage) GetCities(ctx context.Context) ([]City, error) {
	rows, err := s.pool.Query(ctx, sqlGetCities)
	if err != nil {
		return nil, fmt.Errorf("get cities: %w", err)
	}
	defer rows.Close()

	cities := make([]City, 0)
	for rows.Next() {
		var city City
		if err = rows.Scan(&city.ID, &city.Name); err != nil {
			return nil, fmt.Errorf("scan city: %w", err)
		}
		cities = append(cities, city)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cities: %w", err)
	}

	return cities, nil
}
