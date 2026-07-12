package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sixgillkrahs/backend-business-chat/internal/config"
)

func Cors(config *config.Config) gin.HandlerFunc {
	return cors.New(
		cors.Config{
			AllowOrigins:     []string{config.Server.Cors.AllowOrigins},
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Authorization", "Content-Type", "Accept"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		},
	)
}
