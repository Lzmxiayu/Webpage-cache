package worker

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
	"webpage-cache/internal/browser"
	"webpage-cache/internal/model"
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
) *Pool {
	return &Pool{
		workerCount:     workerCount,
		maxRetryCount:   maxRetryCount,
		taskExecTimeout: taskExecTimeout,
		queue:           q,
		repo:            repo,
		screenshotter:   screenshotter,
		storage:         storage,
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
				log.Printf("[Worker %d] stopped\n", id)
				return
			}
			log.Printf("[Worker %d] queue error: %v\n", id, err)
			continue
		}

		task.Status = model.StatusProcessing
		task.UpdatedAt = time.Now()
		if err := p.repo.Update(task); err != nil {
			log.Printf("[Worker %d] update processing status failed for %s: %v\n", id, task.ID, err)
			continue
		}

		log.Printf("[Worker %d] processing %s (retry=%d)\n", id, task.ID, task.RetryCount)

		taskCtx := ctx
		cancel := func() {}
		if p.taskExecTimeout > 0 {
			taskCtx, cancel = context.WithTimeout(ctx, p.taskExecTimeout)
		}

		img, err := p.screenshotter.Capture(taskCtx, task.URL)
		if err != nil {
			cancel()
			p.handleFailure(ctx, id, &task, "capture screenshot failed: "+err.Error())
			continue
		}

		resultURL, err := p.storage.Save(taskCtx, task.ID, img)
		if err != nil {
			cancel()
			p.handleFailure(ctx, id, &task, "save screenshot failed: "+err.Error())
			continue
		}
		cancel()

		task.Status = model.StatusDone
		task.ResultURL = resultURL
		task.ErrorMsg = ""
		task.UpdatedAt = time.Now()
		if err := p.repo.Update(task); err != nil {
			log.Printf("[Worker %d] update done status failed for %s: %v\n", id, task.ID, err)
			continue
		}

		log.Printf("[Worker %d] finished %s\n", id, task.ID)
	}
}

func (p *Pool) handleFailure(ctx context.Context, workerID int, task *model.Task, errMsg string) {
	if task.RetryCount < p.maxRetryCount {
		task.RetryCount++
		task.Status = model.StatusPending
		task.ErrorMsg = errMsg
		task.UpdatedAt = time.Now()

		if err := p.repo.Update(*task); err != nil {
			log.Printf("[Worker %d] update retry status failed for %s: %v\n", workerID, task.ID, err)
			return
		}
		if err := p.queue.Push(ctx, *task); err != nil {
			log.Printf("[Worker %d] requeue failed for %s: %v\n", workerID, task.ID, err)
			p.markFailed(workerID, task, "requeue failed: "+err.Error())
			return
		}

		log.Printf("[Worker %d] task %s retry scheduled (%d/%d): %s\n", workerID, task.ID, task.RetryCount, p.maxRetryCount, errMsg)
		return
	}

	p.markFailed(workerID, task, errMsg)
}

func (p *Pool) markFailed(workerID int, task *model.Task, errMsg string) {
	task.Status = model.StatusFailed
	task.ErrorMsg = errMsg
	task.UpdatedAt = time.Now()
	if err := p.repo.Update(*task); err != nil {
		log.Printf("[Worker %d] update failed status failed for %s: %v\n", workerID, task.ID, err)
		return
	}

	log.Printf("[Worker %d] task %s failed permanently after %d retries: %s\n", workerID, task.ID, task.RetryCount, errMsg)
}
