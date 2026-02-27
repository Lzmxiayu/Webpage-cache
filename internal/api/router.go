package api

import (
	"webpage-cache/internal/api/handler"

	"github.com/gin-gonic/gin"
)

func NewRouter(h *handler.ScreenshotHandler) *gin.Engine {
	r := gin.Default()

	r.POST("/screenshot", h.CreateTask)
	r.GET("/screenshot/:id", h.GetTask)
	return r
}
