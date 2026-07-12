package main

import (
	"fmt"
	"log"

	"github.com/sixgillkrahs/backend-business-chat/internal/config"
	"github.com/sixgillkrahs/backend-business-chat/pkg/utils"
)

func main() {
	utils.InitLogger()
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	app := SetupRouter(cfg)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	if err := app.Run(addr); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
