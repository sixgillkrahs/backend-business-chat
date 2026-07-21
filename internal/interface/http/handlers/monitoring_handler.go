package handlers

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MonitoringHandler struct{}

func NewMonitoringHandler() *MonitoringHandler {
	return &MonitoringHandler{}
}

func (mh *MonitoringHandler) GetMetrics(c *gin.Context) {
	h := promhttp.Handler()
	h.ServeHTTP(c.Writer, c.Request)
}

func (mh *MonitoringHandler) ReceiveAlerts(c *gin.Context) {
	var alertPayload map[string]interface{}
	if err := c.ShouldBindJSON(&alertPayload); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[ALERTMANAGER] Received alerts webhook payload: %+v", alertPayload)
	c.JSON(200, gin.H{"status": "success"})
}
