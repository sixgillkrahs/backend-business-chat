package database

import (
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/sixgillkrahs/backend-business-chat/internal/infrastructure/database/migrations"
)

func RunMigrations(dbConnStr string) error {
	d, err := iofs.New(migrations.MigrationFS, ".")
	if err != nil {
		return fmt.Errorf("failed to create iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dbConnStr)
	if err != nil {
		return fmt.Errorf("failed to init migration: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("PostgreSQL schema is up to date. No migration needed.")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("PostgreSQL database migrated successfully.")
	return nil
}
