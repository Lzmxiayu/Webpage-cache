package storage

import "context"

type Storage interface {
	Save(ctx context.Context, taskID string, data []byte) (string, error)
}

