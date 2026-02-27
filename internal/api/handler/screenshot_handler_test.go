package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"webpage-cache/internal/api/response"
	"webpage-cache/internal/model"
	"webpage-cache/internal/queue"
	"webpage-cache/internal/repository"
	"webpage-cache/internal/service"

	"github.com/gin-gonic/gin"
)

func TestCreateTaskInvalidBody(t *testing.T) {
	h := newTestHandler()
	r := gin.New()
	r.POST("/screenshot", h.CreateTask)

	req := httptest.NewRequest(http.MethodPost, "/screenshot", bytes.NewBufferString(`{`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	resp := decodeAPIResponse(t, w.Body.Bytes())
	if resp.Code != response.CodeInvalidRequest {
		t.Fatalf("expected code %s, got %s", response.CodeInvalidRequest, resp.Code)
	}
}

func TestCreateTaskInvalidURL(t *testing.T) {
	h := newTestHandler()
	r := gin.New()
	r.POST("/screenshot", h.CreateTask)

	req := httptest.NewRequest(http.MethodPost, "/screenshot", bytes.NewBufferString(`{"url":"ftp://example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	resp := decodeAPIResponse(t, w.Body.Bytes())
	if resp.Code != response.CodeInvalidURL {
		t.Fatalf("expected code %s, got %s", response.CodeInvalidURL, resp.Code)
	}
}

func TestCreateTaskAccepted(t *testing.T) {
	h := newTestHandler()
	r := gin.New()
	r.POST("/screenshot", h.CreateTask)

	req := httptest.NewRequest(http.MethodPost, "/screenshot", bytes.NewBufferString(`{"url":"https://example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d", w.Code)
	}

	resp := decodeAPIResponse(t, w.Body.Bytes())
	if resp.Code != response.CodeAccepted {
		t.Fatalf("expected code %s, got %s", response.CodeAccepted, resp.Code)
	}

	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("response data is not object: %#v", resp.Data)
	}
	if data["task_id"] == "" {
		t.Fatalf("task_id is empty")
	}
	if data["status"] != string(model.StatusPending) {
		t.Fatalf("expected status pending, got %v", data["status"])
	}
}

func TestGetTaskInvalidID(t *testing.T) {
	h := newTestHandler()
	r := gin.New()
	r.GET("/screenshot/:id", h.GetTask)

	req := httptest.NewRequest(http.MethodGet, "/screenshot/not-a-uuid", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	resp := decodeAPIResponse(t, w.Body.Bytes())
	if resp.Code != response.CodeInvalidTaskID {
		t.Fatalf("expected code %s, got %s", response.CodeInvalidTaskID, resp.Code)
	}
}

func TestGetTaskNotFound(t *testing.T) {
	h := newTestHandler()
	r := gin.New()
	r.GET("/screenshot/:id", h.GetTask)

	req := httptest.NewRequest(http.MethodGet, "/screenshot/7f656523-24cb-4a6e-b239-4f45b70a8cc8", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}

	resp := decodeAPIResponse(t, w.Body.Bytes())
	if resp.Code != response.CodeTaskNotFound {
		t.Fatalf("expected code %s, got %s", response.CodeTaskNotFound, resp.Code)
	}
}

func TestGetTaskOK(t *testing.T) {
	repo := repository.NewMemoryTaskRepository()
	q := queue.NewMemoryQueue(10)
	svc := service.NewScreenshotService(q, repo)
	h := NewScreenshotHandler(svc)
	r := gin.New()
	r.GET("/screenshot/:id", h.GetTask)

	task, err := svc.CreateTask(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("create task failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/screenshot/"+task.ID, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	resp := decodeAPIResponse(t, w.Body.Bytes())
	if resp.Code != response.CodeOK {
		t.Fatalf("expected code %s, got %s", response.CodeOK, resp.Code)
	}
	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("response data is not object: %#v", resp.Data)
	}
	if data["id"] != task.ID {
		t.Fatalf("expected task id %s, got %v", task.ID, data["id"])
	}
}

func newTestHandler() *ScreenshotHandler {
	repo := repository.NewMemoryTaskRepository()
	q := queue.NewMemoryQueue(10)
	svc := service.NewScreenshotService(q, repo)
	return NewScreenshotHandler(svc)
}

func decodeAPIResponse(t *testing.T, body []byte) response.APIResponse {
	t.Helper()
	var resp response.APIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	return resp
}
