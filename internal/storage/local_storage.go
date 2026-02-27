package storage

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

type LocalStorage struct {
	baseDir string
	baseURL string
}

func NewLocalStorage(baseDir, baseURL string) *LocalStorage {
	return &LocalStorage{
		baseDir: baseDir,
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

func (s *LocalStorage) Save(_ context.Context, taskID string, data []byte) (string, error) {
	if err := os.MkdirAll(s.baseDir, 0o755); err != nil {
		return "", err
	}

	filename := taskID + ".png"
	fullPath := filepath.Join(s.baseDir, filename)
	if err := os.WriteFile(fullPath, data, 0o644); err != nil {
		return "", err
	}

	return s.baseURL + "/" + filename, nil
}

