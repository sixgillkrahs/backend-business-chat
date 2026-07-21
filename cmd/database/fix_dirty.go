package main

import (
	"log"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/sixgillkrahs/backend-business-chat/internal/config"
	"github.com/sixgillkrahs/backend-business-chat/internal/infrastructure/database/migrations"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	version := 5
	if len(os.Args) > 1 {
		v, err := strconv.Atoi(os.Args[1])
		if err == nil {
			version = v
		} else {
			log.Printf("Invalid version argument '%s', defaulting to 5", os.Args[1])
		}
	}

	log.Printf("Forcing migration version to %d...", version)

	d, err := iofs.New(migrations.MigrationFS, ".")
	if err != nil {
		log.Fatalf("Failed to create iofs driver: %v", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, cfg.Postgres.URI)
	if err != nil {
		log.Fatalf("Failed to init migration: %v", err)
	}
	defer m.Close()

	if err := m.Force(version); err != nil {
		log.Fatalf("Failed to force migration version to %d: %v", version, err)
	}

	log.Printf("Successfully forced migration version to %d. You can now run the app or migrations again.", version)
}
