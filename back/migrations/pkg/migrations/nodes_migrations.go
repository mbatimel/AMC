package migrations

import (
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

//go:embed */*.sql
var nodesMigrationsFs embed.FS

func StartNodesMigration(nodes []DatabaseConnectionParams, nodesMigrationsDirectory string) (err error) {
	if len(nodes) == 0 {
		return fmt.Errorf("database count is zero, can not apply migrations")
	}
	err = setup(nodesMigrationsFs)
	if err != nil {
		return
	}
	connections, err := OpenConnections(nodes)
	if err != nil {
		return
	}
	defer closeConnections(connections)

	versions, err := getCurrentNodesVersions(connections)
	if err != nil {
		return
	}

	log.Info().Msg("Database migration launched")
	err = applyMigrations(connections, nodesMigrationsDirectory)
	if err != nil {
		log.Info().Msg("Starting rollback migrations")
		rollbackError := rollbackMigrations(connections, versions, nodesMigrationsDirectory)
		if rollbackError != nil {
			log.Error().Err(rollbackError).Msg("rollback error")
		}
		return
	}
	log.Info().Msg("Database migration completed successfully")
	return
}

func OpenConnections(databaseConnectionParams []DatabaseConnectionParams) (connections []*sql.DB, err error) {
	connections = make([]*sql.DB, len(databaseConnectionParams))
	for i, node := range databaseConnectionParams {
		conn, err := openSqlConnection(&node)
		if err != nil {
			break
		}
		connections[i] = conn
	}
	if err != nil {
		for _, conn := range connections {
			if conn != nil {
				conn.Close()
			}
		}
	}
	return
}

func closeConnections(connections []*sql.DB) (err error) {
	for _, conn := range connections {
		err = conn.Close()
		return
	}
	return
}

func getCurrentNodesVersions(connections []*sql.DB) ([]int64, error) {
	versions := make([]int64, len(connections))
	for i, conn := range connections {
		version, err := goose.GetDBVersion(conn)
		if err != nil {
			return nil, fmt.Errorf("can not get version for db %d, %w", i+1, err)
		}
		versions[i] = version
	}
	return versions, nil
}

func applyMigrations(connections []*sql.DB, nodesMigrationsDirectory string) (err error) {
	group := new(errgroup.Group)
	for i, conn := range connections {
		nodeConn := conn
		nodeIndex := i
		group.Go(func() error {
			log.Info().Msg(fmt.Sprintf("Staring migration to node %d", nodeIndex+1))
			_ = goose.Status(nodeConn, nodesMigrationsDirectory)
			if err = goose.Up(nodeConn, nodesMigrationsDirectory, goose.WithAllowMissing()); err != nil {
				log.Error().Err(err)
				return err
			}
			log.Info().Msg(fmt.Sprintf("Migration applied to node %d", nodeIndex+1))
			return nil
		})
	}
	return group.Wait()
}

func rollbackMigrations(connections []*sql.DB, versions []int64, nodesMigrationsDirectory string) (err error) {
	group := new(errgroup.Group)
	for i, conn := range connections {
		nodeConn := conn
		nodeIndex := i
		group.Go(func() error {
			if err = goose.DownTo(nodeConn, nodesMigrationsDirectory, versions[nodeIndex]); err != nil {
				log.Error().Err(err)
				return fmt.Errorf("can not rollback migrations, %w", err)
			}
			return nil
		})
	}
	return group.Wait()
}
