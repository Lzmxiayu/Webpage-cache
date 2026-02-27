package queue

import (
	"context"
	"webpage-cache/internal/model"
)

type Queue interface {
	Push(ctx context.Context, task model.Task) error
	Pop(ctx context.Context) (model.Task, error)
}
