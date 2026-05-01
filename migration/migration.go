package migration

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

//go:embed *.sql
var migrations embed.FS

func RollMigration(pool *pgxpool.Pool) error {
	dir, err := iofs.New(migrations, ".")
	if err != nil {
		return fmt.Errorf("unable to initiate fs driver: %w", err)
	}
	defer func() { _ = dir.Close() }()

	driver, err := postgres.WithInstance(
		sql.OpenDB(stdlib.GetPoolConnector(pool)),
		&postgres.Config{},
	)
	if err != nil {
		return fmt.Errorf("unable to initiate migration driver: %w", err)
	}
	defer func() { _ = driver.Close() }()

	migration, err := migrate.NewWithInstance(".", dir, "postgres", driver)
	if err != nil {
		return fmt.Errorf("unable to initiate migration: %w", err)
	}
	defer func() { _, _ = migration.Close() }()

	err = migration.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("unable to roll migration: %w", err)
	}

	return nil
}
