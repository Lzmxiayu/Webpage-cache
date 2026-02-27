package handler

import (
	"net/http"
	"webpage-cache/internal/service"

	"github.com/gin-gonic/gin"
)

type ScreenshotHandler struct {
	service *service.ScreenshotService
}

func NewScreenshotHandler(s *service.ScreenshotService) *ScreenshotHandler {
	return &ScreenshotHandler{service: s}
}

type CreateRequest struct {
	URL string `json:"url" binding:"required"`
}

func (h *ScreenshotHandler) CreateTask(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, _ := h.service.CreateTask(req.URL)

	c.JSON(http.StatusOK, gin.H{
		"task_id": task.ID,
		"status":  task.Status,
	})
}

func (h *ScreenshotHandler) GetTask(c *gin.Context) {

	id := c.Param("id")

	task, ok := h.service.GetTask(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}
