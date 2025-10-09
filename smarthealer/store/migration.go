package store

import (
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

const (
	migrationPath = "migrations"
)

var (
	ErrMigrationFailed = errors.New("migration failed")
)

func EnsureMigrations(dbpath string) error {

	fs, err := iofs.New(migrationFS, migrationPath)
	if err != nil {
		return fmt.Errorf(
			"unable to initialize migrationfs: %w: %w",
			ErrMigrationFailed,
			err,
		)
	}
	dsn := fmt.Sprintf("sqlite3://%s", dbpath)

	m, err := migrate.NewWithSourceInstance("iofs", fs, dsn)
	if err != nil {
		return fmt.Errorf(
			"failed to create migrate instance: %w: %w",
			ErrMigrationFailed,
			err,
		)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf(
			"failed to run migrations: %w: %w",
			ErrMigrationFailed,
			err,
		)
	}

	return nil
}
