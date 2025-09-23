package store

import (
	"embed"
	"errors"
	"fmt"

	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/config"
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

func EnsureMigrations(cfg config.DbConfig) error {
	dbpath := cfg.Path

	fs, err := iofs.New(migrationFS, migrationPath)
	if err != nil {
		return fmt.Errorf(
			"unable to initialize migartionfs: %w: %w",
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
	m.Up()

	return nil
}
