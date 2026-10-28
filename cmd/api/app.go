package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/sixgillkrahs/backend-business-chat/docs" // swagger docs
	"github.com/sixgillkrahs/backend-business-chat/internal/application"
	"github.com/sixgillkrahs/backend-business-chat/internal/config"
	"github.com/sixgillkrahs/backend-business-chat/internal/infrastructure/database"
	"github.com/sixgillkrahs/backend-business-chat/internal/infrastructure/repository"
	"github.com/sixgillkrahs/backend-business-chat/internal/interface/http/handlers"
	"github.com/sixgillkrahs/backend-business-chat/internal/interface/http/middleware"
	"github.com/sixgillkrahs/backend-business-chat/internal/interface/http/routes"
	"github.com/sixgillkrahs/backend-business-chat/pkg/utils"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(cfg *config.Config, db *database.PostgresDB) *gin.Engine {
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	r := gin.New()
	r.Use(utils.GinLogger(), utils.GinRecovery(), middleware.Cors(cfg), middleware.PrometheusMiddleware(), middleware.ValidateParams(), middleware.LanguageMiddleware())

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
		authService := application.NewAuthService(cfg, actionRepo, authRepo, resourcesRepo, policyRepo)

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

	// Swagger UI (without production)
	if cfg.Server.Env != "production" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		log.Println("Swagger UI enabled at /swagger/index.html")
	}

	return r
}
