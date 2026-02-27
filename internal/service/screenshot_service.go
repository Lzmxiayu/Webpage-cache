package service

import (
	"time"
	"webpage-cache/internal/model"
	"webpage-cache/internal/repository"

	"github.com/google/uuid"
)

type ScreenshotService struct {
	jobChan chan model.Task
	repo    repository.TaskRepository
}

func NewScreenshotService(jobChan chan model.Task, repo repository.TaskRepository) *ScreenshotService {
	return &ScreenshotService{
		jobChan: jobChan,
		repo:    repo,
	}
}

func (s *ScreenshotService) CreateTask(url string) (model.Task, error) {

	task := model.Task{
		ID:        uuid.NewString(),
		URL:       url,
		Status:    model.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 存储任务
	s.repo.Create(task)

	// 推入 worker
	s.jobChan <- task

	return task, nil
}

func (s *ScreenshotService) GetTask(id string) (model.Task, bool) {
	return s.repo.GetByID(id)
}
