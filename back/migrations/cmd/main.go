package main

import (
	"bufio"
	"os"
	"strings"

	"github.com/mbatimel/AMC/migrations/pkg/migrations"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type PGConfig struct {
	Address  string // host:port
	DB       string // db
	User     string // user
	Password string // password
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	if err := loadEnvFile(".env"); err != nil && !os.IsNotExist(err) {
		panic(err)
	}

	if _, err := os.Stat("./pkg/migrations/data"); os.IsNotExist(err) {
		panic("Migration directory is not exist")
	}

	pgConfig := PGConfig{}

	pgConfig.Address = os.Getenv("PG_ADDRESS")   // host:port
	pgConfig.DB = os.Getenv("PG_DB")             // db
	pgConfig.User = os.Getenv("PG_USER")         // user
	pgConfig.Password = os.Getenv("PG_PASSWORD") // password
	if len(pgConfig.Address) == 0 || len(pgConfig.DB) == 0 || len(pgConfig.User) == 0 || len(pgConfig.Password) == 0 {
		panic("Address, DB, user and password must be specified")
	}

	nodeAddress, nodePort := migrations.ParseDbAddressAndPort(pgConfig.Address)

	migrationParams := migrations.DatabaseConnectionParams{
		Host: nodeAddress, Port: nodePort, Database: pgConfig.DB, User: pgConfig.User, Password: pgConfig.Password,
	}

	err := migrations.StartNodesMigration([]migrations.DatabaseConnectionParams{migrationParams}, "data")
	if err != nil {
		panic(err)
	}
}

func loadEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		if key == "" || os.Getenv(key) != "" {
			continue
		}

		value = strings.Trim(strings.TrimSpace(value), `"'`)
		value = os.Expand(value, func(name string) string {
			return os.Getenv(name)
		})
		if err = os.Setenv(key, value); err != nil {
			return err
		}
	}
	return scanner.Err()
}
