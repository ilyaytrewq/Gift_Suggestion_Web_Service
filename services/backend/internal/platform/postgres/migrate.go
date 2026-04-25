package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	projectmigrations "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/migrations"
)

func RunMigrations(ctx context.Context, db *sql.DB) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}

	driver, err := migratepostgres.WithConnection(ctx, conn, &migratepostgres.Config{
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		if closeErr := conn.Close(); closeErr != nil {
			return errors.Join(err, closeErr)
		}

		return err
	}

	sourceDriver, err := iofs.New(projectmigrations.Files, ".")
	if err != nil {
		if closeErr := driver.Close(); closeErr != nil {
			return errors.Join(err, closeErr)
		}

		return err
	}

	migrator, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		if closeErr := driver.Close(); closeErr != nil {
			return errors.Join(err, closeErr)
		}

		return err
	}

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		sourceErr, databaseErr := migrator.Close()

		return errors.Join(err, sourceErr, databaseErr)
	}

	sourceErr, databaseErr := migrator.Close()
	if sourceErr != nil {
		return sourceErr
	}

	return databaseErr
}
