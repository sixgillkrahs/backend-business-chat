package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/sixgillkrahs/backend-business-chat/internal/interface/http/handlers"
)

func MonitoringRouter(r *gin.RouterGroup, monitoringHandler *handlers.MonitoringHandler) {
	api := r.Group("/monitoring")
	{
		api.GET("/metrics", monitoringHandler.GetMetrics)
	}
}
