package queue

import (
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

func (q *MemoryQueue) Push(task model.Task) error {
	q.ch <- task
	return nil
}

func (q *MemoryQueue) Pop() (model.Task, error) {
	task, ok := <-q.ch
	if !ok {
		return model.Task{}, errors.New("queue closed")
	}
	return task, nil
}
