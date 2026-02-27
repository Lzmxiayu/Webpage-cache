package queue

import "webpage-cache/internal/model"

type Queue interface {
	Push(task model.Task) error
	Pop() (model.Task, error)
}
