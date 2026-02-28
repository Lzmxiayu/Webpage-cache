package worker

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"
	"webpage-cache/internal/browser"
	"webpage-cache/internal/model"
	"webpage-cache/internal/observability"
	"webpage-cache/internal/queue"
	"webpage-cache/internal/repository"
	"webpage-cache/internal/storage"
)

type Pool struct {
	workerCount     int
	maxRetryCount   int
	taskExecTimeout time.Duration
	queue           queue.Queue
	repo            repository.TaskRepository
	screenshotter   browser.Screenshotter
	storage         storage.Storage
	logger          *slog.Logger
	wg              sync.WaitGroup
}

func NewPool(
	workerCount int,
	maxRetryCount int,
	taskExecTimeout time.Duration,
	q queue.Queue,
	repo repository.TaskRepository,
	screenshotter browser.Screenshotter,
	storage storage.Storage,
	logger *slog.Logger,
) *Pool {
	return &Pool{
		workerCount:     workerCount,
		maxRetryCount:   maxRetryCount,
		taskExecTimeout: taskExecTimeout,
		queue:           q,
		repo:            repo,
		screenshotter:   screenshotter,
		storage:         storage,
		logger:          logger,
	}
}

func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(ctx, i)
	}
}

func (p *Pool) Wait() {
	p.wg.Wait()
}

func (p *Pool) worker(ctx context.Context, id int) {
	defer p.wg.Done()

	for {
		task, err := p.queue.Pop(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
				p.logger.Info("worker_stopped", "worker_id", id)
				return
			}
			p.logger.Error("queue_pop_failed", "worker_id", id, "error", err)
			continue
		}

		startedAt := time.Now()
		task.Status = model.StatusProcessing
		task.UpdatedAt = time.Now()
		if err := p.repo.Update(task); err != nil {
			p.logger.Error("task_update_processing_failed", "worker_id", id, "task_id", task.ID, "error", err)
			continue
		}

		p.logger.Info("task_processing_started", "worker_id", id, "task_id", task.ID, "retry_count", task.RetryCount)

		taskCtx := ctx
		cancel := func() {}
		if p.taskExecTimeout > 0 {
			taskCtx, cancel = context.WithTimeout(ctx, p.taskExecTimeout)
		}
		if task.LastProxy != "" {
			taskCtx = browser.WithAvoidProxy(taskCtx, task.LastProxy)
		}

		img, usedProxy, err := p.screenshotter.Capture(taskCtx, task.URL)
		if usedProxy != "" {
			task.LastProxy = usedProxy
		}
		if err != nil {
			cancel()
			observability.ObserveTaskProcessed("failed", time.Since(startedAt))
			p.handleFailure(ctx, id, &task, "capture screenshot failed: "+err.Error())
			continue
		}

		resultURL, err := p.storage.Save(taskCtx, task.ID, img)
		if err != nil {
			cancel()
			observability.ObserveTaskProcessed("failed", time.Since(startedAt))
			p.handleFailure(ctx, id, &task, "save screenshot failed: "+err.Error())
			continue
		}
		cancel()

		task.Status = model.StatusDone
		task.ResultURL = resultURL
		task.ErrorMsg = ""
		task.UpdatedAt = time.Now()
		if err := p.repo.Update(task); err != nil {
			p.logger.Error("task_update_done_failed", "worker_id", id, "task_id", task.ID, "error", err)
			continue
		}

		observability.ObserveTaskProcessed("done", time.Since(startedAt))
		p.logger.Info("task_processing_done", "worker_id", id, "task_id", task.ID, "duration_ms", time.Since(startedAt).Milliseconds())
	}
}

func (p *Pool) handleFailure(ctx context.Context, workerID int, task *model.Task, errMsg string) {
	if task.RetryCount < p.maxRetryCount {
		task.RetryCount++
		task.Status = model.StatusPending
		task.ErrorMsg = errMsg
		task.UpdatedAt = time.Now()

		if err := p.repo.Update(*task); err != nil {
			p.logger.Error("task_update_retry_failed", "worker_id", workerID, "task_id", task.ID, "error", err)
			return
		}
		if err := p.queue.Push(ctx, *task); err != nil {
			p.logger.Error("task_requeue_failed", "worker_id", workerID, "task_id", task.ID, "error", err)
			p.markFailed(workerID, task, "requeue failed: "+err.Error())
			return
		}

		observability.IncTaskRetry()
		p.logger.Warn(
			"task_retry_scheduled",
			"worker_id", workerID,
			"task_id", task.ID,
			"retry_count", task.RetryCount,
			"max_retry_count", p.maxRetryCount,
			"reason", errMsg,
		)
		return
	}

	p.markFailed(workerID, task, errMsg)
}

func (p *Pool) markFailed(workerID int, task *model.Task, errMsg string) {
	task.Status = model.StatusFailed
	task.ErrorMsg = errMsg
	task.UpdatedAt = time.Now()
	if err := p.repo.Update(*task); err != nil {
		p.logger.Error("task_update_failed_status_failed", "worker_id", workerID, "task_id", task.ID, "error", err)
		return
	}

	p.logger.Error(
		"task_failed_permanently",
		"worker_id", workerID,
		"task_id", task.ID,
		"retry_count", task.RetryCount,
		"reason", errMsg,
	)
}
