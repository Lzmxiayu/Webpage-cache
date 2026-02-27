package repository

import "webpage-cache/internal/model"

type TaskRepository interface {
	Create(task model.Task) error
	Update(task model.Task) error
	GetByID(id string) (model.Task, bool)
}
