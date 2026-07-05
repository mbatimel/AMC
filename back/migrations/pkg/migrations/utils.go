package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"strconv"
	"strings"

	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog/log"
	"github.com/mbatimel/AMC/migrations/pkg/goose_logger"
)

type DatabaseConnectionParams struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
}

func setup(fs embed.FS) (err error) {
	goose.SetTableName("public.goose_db_version")
	goose.SetBaseFS(fs)
	if err = goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("can not set dialect for migrations, %w", err)
	}
	goose.SetVerbose(true)
	goose.SetLogger(goose_logger.NewGooseLoggerAdapter(&log.Logger))
	return
}

func openSqlConnection(connParam *DatabaseConnectionParams) (*sql.DB, error) {
	url := fmt.Sprintf("host=%s port=%d dbname=%s sslmode=disable user=%s password=%s",
		connParam.Host, connParam.Port, connParam.Database, connParam.User, connParam.Password)
	return sql.Open("postgres", url)
}

func ParseDbAddressAndPort(conn string) (string, int) {
	splits := strings.Split(conn, ":")
	address := splits[0]
	port, err := strconv.Atoi(splits[1])
	if err != nil {
		log.Fatal().Msgf("failed parse db address and port, connection string=%s, %v", conn, err)
	}
	return address, port
}
