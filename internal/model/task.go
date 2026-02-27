package model

import "time"

type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusProcessing TaskStatus = "processing"
	StatusDone       TaskStatus = "done"
	StatusFailed     TaskStatus = "failed"
)

type Task struct {
	ID         string
	URL        string
	Status     TaskStatus
	ResultURL  string
	ErrorMsg   string
	RetryCount int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
