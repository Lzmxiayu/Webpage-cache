package worker

import (
	"context"
	"errors"
	"testing"
	"time"
	"webpage-cache/internal/model"
	"webpage-cache/internal/observability"
	"webpage-cache/internal/queue"
	"webpage-cache/internal/repository"
)

type fakeScreenshotter struct {
	failCount int
	calls     int
}

func (f *fakeScreenshotter) Capture(_ context.Context, _ string) ([]byte, string, error) {
	f.calls++
	if f.calls <= f.failCount {
		return nil, "http://proxy-a:8080", errors.New("screenshot failed")
	}
	return []byte("png"), "http://proxy-b:8080", nil
}

func (f *fakeScreenshotter) Close() {}

type fakeStorage struct{}

func (s *fakeStorage) Save(_ context.Context, taskID string, _ []byte) (string, error) {
	return "/static/screenshots/" + taskID + ".png", nil
}

func TestPoolTaskDone(t *testing.T) {
	repo := repository.NewMemoryTaskRepository()
	q := queue.NewMemoryQueue(10)
	logger := observability.NewLogger("error")

	shot := &fakeScreenshotter{}
	pool := NewPool(1, 2, time.Second, q, repo, shot, &fakeStorage{}, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)

	task := model.Task{
		ID:        "task-done",
		URL:       "https://example.com",
		Status:    model.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := repo.Create(task); err != nil {
		t.Fatalf("repo create failed: %v", err)
	}
	if err := q.Push(context.Background(), task); err != nil {
		t.Fatalf("queue push failed: %v", err)
	}

	got, ok := waitForStatus(repo, task.ID, model.StatusDone, 3*time.Second)
	if !ok {
		t.Fatalf("task did not reach done status in time")
	}
	if got.ResultURL == "" {
		t.Fatalf("result url is empty")
	}
	if got.RetryCount != 0 {
		t.Fatalf("expected retry_count 0, got %d", got.RetryCount)
	}

	cancel()
	pool.Wait()
}

func TestPoolRetryThenDone(t *testing.T) {
	repo := repository.NewMemoryTaskRepository()
	q := queue.NewMemoryQueue(10)
	logger := observability.NewLogger("error")

	shot := &fakeScreenshotter{failCount: 1}
	pool := NewPool(1, 2, time.Second, q, repo, shot, &fakeStorage{}, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)

	task := model.Task{
		ID:        "task-retry",
		URL:       "https://example.com",
		Status:    model.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = repo.Create(task)
	_ = q.Push(context.Background(), task)

	got, ok := waitForStatus(repo, task.ID, model.StatusDone, 3*time.Second)
	if !ok {
		t.Fatalf("task did not reach done status in time")
	}
	if got.RetryCount != 1 {
		t.Fatalf("expected retry_count 1, got %d", got.RetryCount)
	}
	if shot.calls < 2 {
		t.Fatalf("expected at least 2 capture calls, got %d", shot.calls)
	}

	cancel()
	pool.Wait()
}

func TestPoolPermanentFailure(t *testing.T) {
	repo := repository.NewMemoryTaskRepository()
	q := queue.NewMemoryQueue(10)
	logger := observability.NewLogger("error")

	shot := &fakeScreenshotter{failCount: 10}
	pool := NewPool(1, 1, time.Second, q, repo, shot, &fakeStorage{}, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)

	task := model.Task{
		ID:        "task-failed",
		URL:       "https://example.com",
		Status:    model.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = repo.Create(task)
	_ = q.Push(context.Background(), task)

	got, ok := waitForStatus(repo, task.ID, model.StatusFailed, 3*time.Second)
	if !ok {
		t.Fatalf("task did not reach failed status in time")
	}
	if got.RetryCount != 1 {
		t.Fatalf("expected retry_count 1, got %d", got.RetryCount)
	}
	if got.ErrorMsg == "" {
		t.Fatalf("expected error message to be set")
	}

	cancel()
	pool.Wait()
}

func waitForStatus(repo *repository.MemoryTaskRepository, taskID string, status model.TaskStatus, timeout time.Duration) (model.Task, bool) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		task, ok := repo.GetByID(taskID)
		if ok && task.Status == status {
			return task, true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return model.Task{}, false
}
