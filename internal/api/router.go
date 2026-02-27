package api

import (
	"log/slog"
	"webpage-cache/internal/api/handler"
	"webpage-cache/internal/observability"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(logger *slog.Logger, h *handler.ScreenshotHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(observability.RequestMetricsAndLoggingMiddleware(logger))

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.Static("/static/screenshots", "./data/screenshots")
	r.POST("/screenshot", h.CreateTask)
	r.GET("/screenshot/:id", h.GetTask)
	return r
}
