package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)


func main() {
	var storagePath, migrationsPath, migrationsTable string

	flag.StringVar(&storagePath, "storage-path", "storage.db", "path to storage file")
	flag.StringVar(&migrationsPath, "migrations-path", "migrations", "path to migrations files")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "table name for migrations")
	flag.Parse()


	if storagePath == "" {
		log.Fatal("storage path is required")
	}

	if migrationsPath == "" {
		log.Fatal("migrations path is required")
	}

	if migrationsTable == "" {
		log.Fatal("migrations table is required")
	}

	m, err := migrate.New(
		"file://" + migrationsPath,
		fmt.Sprintf("sqlite://%s?x-migrations-table=%s", storagePath, migrationsTable),
	)
	if err != nil {
		log.Fatalf("failed to create migrator: %v", err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("no migrations to apply")
			os.Exit(0)
		}
		log.Fatalf("failed to apply migrations: %v", err)
	}

	slog.Info("migrations applied successfully")
}