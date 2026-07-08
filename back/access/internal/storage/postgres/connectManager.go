package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	_ "github.com/lib/pq"

	"github.com/mbatimel/AMC/access/internal/config"
)

func OpenPostgres(cfg config.Config) (*sql.DB, error) {
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

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}
