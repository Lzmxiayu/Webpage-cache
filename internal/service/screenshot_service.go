package service

import (
	"context"
	"time"
	"webpage-cache/internal/model"
	"webpage-cache/internal/queue"
	"webpage-cache/internal/repository"

	"github.com/google/uuid"
)

type ScreenshotService struct {
	queue queue.Queue
	repo  repository.TaskRepository
}

func NewScreenshotService(q queue.Queue, repo repository.TaskRepository) *ScreenshotService {
	return &ScreenshotService{
		queue: q,
		repo:  repo,
	}
}

func (s *ScreenshotService) CreateTask(ctx context.Context, url string) (model.Task, error) {
	task := model.Task{
		ID:        uuid.NewString(),
		URL:       url,
		Status:    model.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(task); err != nil {
		return model.Task{}, err
	}

	if err := s.queue.Push(ctx, task); err != nil {
		task.Status = model.StatusFailed
		task.ErrorMsg = "enqueue failed: " + err.Error()
		task.UpdatedAt = time.Now()
		_ = s.repo.Update(task)
		return model.Task{}, err
	}

	return task, nil
}

func (s *ScreenshotService) GetTask(id string) (model.Task, bool) {
	return s.repo.GetByID(id)
}
