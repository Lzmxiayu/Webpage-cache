package repository

import (
	"sync"
	"webpage-cache/internal/model"
)

type MemoryTaskRepository struct {
	mu    sync.RWMutex
	tasks map[string]model.Task
}

func NewMemoryTaskRepository() *MemoryTaskRepository {
	return &MemoryTaskRepository{
		tasks: make(map[string]model.Task),
	}
}

func (r *MemoryTaskRepository) Create(task model.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[task.ID] = task
	return nil
}

func (r *MemoryTaskRepository) Update(task model.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[task.ID] = task
	return nil
}

func (r *MemoryTaskRepository) GetByID(id string) (model.Task, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	task, ok := r.tasks[id]
	return task, ok
}
