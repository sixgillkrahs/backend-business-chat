package main

import (
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
	r.Use(utils.GinLogger(), utils.GinRecovery(), middleware.Cors(cfg))

	apiV1 := r.Group("/v1/api")
	{
		apiV1.GET("/ping", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{
				"message": "pong",
				"env":     cfg.Server.Env,
			})
		})
		authRepo := repository.NewAuthRepository(db)
		authService := application.NewAuthService(authRepo)
		authHandler := handlers.NewAuthHandler(authService)
		routes.AuthRoutes(apiV1, authHandler)
	}
	return r
}
