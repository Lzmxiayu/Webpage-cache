package handler

import (
	"net/http"
	"net/url"
	"strings"
	"time"
	"webpage-cache/internal/api/response"
	"webpage-cache/internal/model"
	"webpage-cache/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

type CreateTaskData struct {
	TaskID string           `json:"task_id"`
	Status model.TaskStatus `json:"status"`
}

type TaskData struct {
	ID         string           `json:"id"`
	URL        string           `json:"url"`
	Status     model.TaskStatus `json:"status"`
	ResultURL  string           `json:"result_url"`
	ErrorMsg   string           `json:"error_msg"`
	RetryCount int              `json:"retry_count"`
	CreatedAt  string           `json:"created_at"`
	UpdatedAt  string           `json:"updated_at"`
}

func (h *ScreenshotHandler) CreateTask(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(response.CodeInvalidRequest, "invalid request body"))
		return
	}

	if !isValidHTTPURL(req.URL) {
		c.JSON(http.StatusBadRequest, response.Error(response.CodeInvalidURL, "url must be a valid http/https URL"))
		return
	}

	task, err := h.service.CreateTask(req.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(response.CodeInternalError, "failed to create task"))
		return
	}

	c.JSON(http.StatusAccepted, response.Success(response.CodeAccepted, "task accepted", CreateTaskData{
		TaskID: task.ID,
		Status: task.Status,
	}))
}

func (h *ScreenshotHandler) GetTask(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(response.CodeInvalidTaskID, "task id must be a valid UUID"))
		return
	}

	task, ok := h.service.GetTask(id)
	if !ok {
		c.JSON(http.StatusNotFound, response.Error(response.CodeTaskNotFound, "task not found"))
		return
	}

	c.JSON(http.StatusOK, response.Success(response.CodeOK, "ok", toTaskData(task)))
}

func toTaskData(task model.Task) TaskData {
	return TaskData{
		ID:         task.ID,
		URL:        task.URL,
		Status:     task.Status,
		ResultURL:  task.ResultURL,
		ErrorMsg:   task.ErrorMsg,
		RetryCount: task.RetryCount,
		CreatedAt:  task.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  task.UpdatedAt.Format(time.RFC3339),
	}
}

func isValidHTTPURL(raw string) bool {
	u, err := url.ParseRequestURI(strings.TrimSpace(raw))
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return u.Host != ""
}
