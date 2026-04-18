package postgres

import (
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	projectmigrations "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/migrations"
)

func RunMigrations(db *sql.DB) error {
	driver, err := migratepostgres.WithInstance(db, &migratepostgres.Config{
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		return err
	}

	sourceDriver, err := iofs.New(projectmigrations.Files, ".")
	if err != nil {
		return err
	}

	migrator, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return err
	}

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	sourceErr, databaseErr := migrator.Close()
	if sourceErr != nil {
		return sourceErr
	}

	return databaseErr
}
