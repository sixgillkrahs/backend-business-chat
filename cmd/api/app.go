package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sixgillkrahs/backend-business-chat/internal/config"
	"github.com/sixgillkrahs/backend-business-chat/pkg/utils"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	r := gin.New()
	r.Use(utils.GinLogger(), utils.GinRecovery())

	apiV1 := r.Group("/v1/api")
	{
		apiV1.GET("/ping", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{
				"message": "pong",
				"env":     cfg.Server.Env,
			})
		})
	}
	return r
}
