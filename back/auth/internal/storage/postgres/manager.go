package postgres

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgconn/stmtcache"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mbatimel/AMC/auth/internal/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const (
	MASTER  = true
	REPLICA = false
)

type ConnectManager interface {
	GetConnection(fromMaster bool) *pgxpool.Pool
}

type manager struct {
	log     zerolog.Logger
	master  *pgxpool.Pool
	replica *pgxpool.Pool
}

func (m *manager) GetConnection(fromMaster bool) (shard *pgxpool.Pool) {
	if fromMaster {
		return m.master
	}
	return m.replica
}

func NewManager(log zerolog.Logger, cfg config.Postgres) (ConnectManager, error) {
	ctx := context.Background()

	masterAddr, masterPort, err := parseDbAddressAndPort(cfg.MasterAddress)
	if err != nil {
		return nil, err
	}

	master, err := dbConnect(ctx, cfg, masterAddr, masterPort, cfg.DBName, cfg.UserName, cfg.Password)
	if err != nil {
		return nil, err
	}

	replicaAddr, replicaPort, err := parseDbAddressAndPort(cfg.ReplicaAddress)
	if err != nil {
		return nil, err
	}

	replica, err := dbConnect(ctx, cfg, replicaAddr, replicaPort, cfg.DBName, cfg.UserNameRO, cfg.PasswordRO)
	if err != nil {
		return nil, err
	}

	return &manager{
		log:     log,
		master:  master,
		replica: replica,
	}, nil
}

func parseDbAddressAndPort(conn string) (string, int, error) {
	splits := strings.Split(conn, ":")
	address := splits[0]
	port, err := strconv.Atoi(splits[1])
	if err != nil {
		return "", 0, fmt.Errorf("failed parse db address and port, connection string=%s", conn)
	}

	return address, port, nil
}

func dbConnect(ctx context.Context, postgresCfg config.Postgres, dbAddr string, dbPort int, db, user, password string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(fmt.Sprintf(
		"host=%s port=%d dbname=%s sslmode=disable user=%s password=%s pool_max_conns=%d",
		dbAddr, dbPort, db, user, password, postgresCfg.MaxConn,
	))
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed parse postgres dsn: %s:%v", dbAddr, dbAddr))
	}

	cfg.ConnConfig.LogLevel = pgx.LogLevelError

	mci, err := time.ParseDuration(postgresCfg.MaxIdleLifetime)
	if err != nil {
		return nil, errors.Wrap(err, "failed parse max idle conn lifetime to duration")
	}
	mc, err := time.ParseDuration(postgresCfg.MaxLifetime)
	if err != nil {
		return nil, errors.Wrap(err, "failed parse max conn lifetime to duration")
	}

	cfg.MaxConnIdleTime = mci
	cfg.MaxConnLifetime = mc

	cfg.ConnConfig.BuildStatementCache = func(conn *pgconn.PgConn) stmtcache.Cache {
		return stmtcache.New(conn, stmtcache.ModeDescribe, postgresCfg.PrepareCacheCap)
	}

	pool, err := pgxpool.ConnectConfig(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed connect to pg: %s:%v", dbAddr, dbPort))
	}

	return pool, nil
}
