package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/sixgillkrahs/backend-business-chat/internal/interface/http/handlers"
)

func AuthRoutes(r *gin.RouterGroup, authHandler handlers.AuthHandler) {
	api := r.Group("/auth")
	{
		api.GET("/actions", authHandler.ListActions)
		api.GET("/resources", authHandler.ListResources)
		api.GET("/policies", authHandler.ListPolicies)
	}
}
