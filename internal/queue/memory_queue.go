package queue

import (
	"context"
	"errors"
	"webpage-cache/internal/model"
)

type MemoryQueue struct {
	ch chan model.Task
}

func NewMemoryQueue(size int) *MemoryQueue {
	return &MemoryQueue{
		ch: make(chan model.Task, size),
	}
}

func (q *MemoryQueue) Push(ctx context.Context, task model.Task) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case q.ch <- task:
		return nil
	}
}

func (q *MemoryQueue) Pop(ctx context.Context) (model.Task, error) {
	select {
	case <-ctx.Done():
		return model.Task{}, ctx.Err()
	case task, ok := <-q.ch:
		if !ok {
			return model.Task{}, errors.New("queue closed")
		}
		return task, nil
	}
}
