package main

import (
	"context"
	"fmt"
	"log"

	"github.com/sixgillkrahs/backend-business-chat/internal/config"
	"github.com/sixgillkrahs/backend-business-chat/internal/infrastructure/database"
	"github.com/sixgillkrahs/backend-business-chat/internal/infrastructure/redis"
	"github.com/sixgillkrahs/backend-business-chat/pkg/utils"
)

func main() {
	utils.InitLogger()
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	if err := database.RunMigrations(cfg.Postgres.URI); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}
	ctx := context.Background()
	dbConn, err := database.NewPostgresConnection(ctx, cfg.Postgres.URI)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer dbConn.Close()
	log.Println("PostgreSQL connection pool initialized and warmed up.")
	rdb, err := redis.NewRedisConnection(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer rdb.Close()
	log.Println("Redis client initialized and warmed up.")
	app := SetupRouter(cfg, dbConn)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	if err := app.Run(addr); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
