package handlers

import (
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
