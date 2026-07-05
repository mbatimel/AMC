package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/mbatimel/AMC/migrations/pkg/migrations"
)

type PGConfig struct {
	Address  string // host:port
	DB       string // db
	User     string // user
	Password string // password
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})


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
