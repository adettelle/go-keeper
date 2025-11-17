// Package migrator provides functionality for applying database migrations
// using embedded SQL files and the `golang-migrate` library.
package migrator

import (
	"database/sql"
	"embed"
	"errors"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// migrationsDir specifies the directory containing migration files within the embedded filesystem.
const migrationsDir = "migration"

// MigrationFS is an embedded filesystem containing SQL migration files.
//
// The //go:embed directive embeds all SQL files in the `migration` directory into the binary.
//
//go:embed migration/*.sql
var MigrationFS embed.FS

// MustApplyMigrations applies all pending migrations to the PostgreSQL database specified by dbParams.
// The migrations are sourced from the embedded `MigrationFS`.
//
// Parameters:
//   - dbParams: A connection string containing database configuration details.
func MustApplyMigrations(dbParams string) {
	// Create a new source driver from the embedded filesystem
	srcDriver, err := iofs.New(MigrationFS, migrationsDir)
	if err != nil {
		log.Fatal(err)
	}

	// Open the database connection
	db, err := sql.Open("pgx", dbParams)
	if err != nil {
		log.Fatal(err)
	}

	// Create a PostgreSQL driver instance
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("unable to create db instance: %v", err)
	}

	// Create a new migrator instance with the embedded migration files
	migrator, err := migrate.NewWithInstance("migration_embeded_sql_files", srcDriver, "psql_db", driver)
	if err != nil {
		log.Fatalf("unable to create migration: %v", err)
	}

	// Ensure the migrator is closed when done
	defer func() {
		migrator.Close()
	}()

	// Apply all migrations; ignore the error if there are no changes
	if err = migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("unable to apply migrations %v", err)
	}

	log.Println("Migrations applied")
}
