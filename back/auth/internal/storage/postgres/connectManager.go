package postgres

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/mbatimel/AMC/auth/internal/config"
)

func NewPool(cfg config.Config) (*pgxpool.Pool, error) {
	port, err := strconv.Atoi(cfg.PGPort)
	if err != nil {
		return nil, fmt.Errorf("parse PG_PORT: %w", err)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d dbname=%s sslmode=disable user=%s password=%s",
		cfg.PGHost,
		port,
		cfg.PGDB,
		cfg.PGUser,
		cfg.PGPassword,
	)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse postgres dsn: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.ConnectConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}
