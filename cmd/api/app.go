package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sixgillkrahs/backend-business-chat/internal/application"
	"github.com/sixgillkrahs/backend-business-chat/internal/config"
	"github.com/sixgillkrahs/backend-business-chat/internal/infrastructure/database"
	"github.com/sixgillkrahs/backend-business-chat/internal/infrastructure/repository"
	"github.com/sixgillkrahs/backend-business-chat/internal/interface/http/handlers"
	"github.com/sixgillkrahs/backend-business-chat/internal/interface/http/middleware"
	"github.com/sixgillkrahs/backend-business-chat/internal/interface/http/routes"
	"github.com/sixgillkrahs/backend-business-chat/pkg/utils"
)

func SetupRouter(cfg *config.Config, db *database.PostgresDB) *gin.Engine {
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	r := gin.New()
	r.Use(utils.GinLogger(), utils.GinRecovery(), middleware.Cors(cfg), middleware.PrometheusMiddleware())

	apiV1 := r.Group("/v1/api")
	{
		apiV1.GET("/ping", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{
				"message": "pong",
				"env":     cfg.Server.Env,
			})
		})
		actionRepo := repository.NewActionRepository(db)
		authRepo := repository.NewAuthRepository(db)
		resourcesRepo := repository.NewResourceRepository(db)
		policyRepo := repository.NewPolicyRepository(db)
		authService := application.NewAuthService(actionRepo, authRepo, resourcesRepo, policyRepo)

		// Auto-initialize default resources on app startup
		if err := authService.InitDefaultResources(context.Background()); err != nil {
			log.Printf("Warning: Failed to auto-initialize default resources: %v", err)
		} else {
			log.Println("Auto-initialized default resources successfully.")
		}

		authHandler := handlers.NewAuthHandler(authService)
		routes.AuthRoutes(apiV1, authHandler)
	}
	mr := r.Group("api")
	{
		monitoringHandler := handlers.NewMonitoringHandler()
		routes.MonitoringRouter(mr, monitoringHandler)
	}
	return r
}
