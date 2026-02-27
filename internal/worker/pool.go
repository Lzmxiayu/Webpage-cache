package worker

import (
	"context"
	"log"
	"time"
	"webpage-cache/internal/browser"
	"webpage-cache/internal/model"
	"webpage-cache/internal/queue"
	"webpage-cache/internal/repository"
	"webpage-cache/internal/storage"
)

type Pool struct {
	workerCount   int
	maxRetryCount int
	queue         queue.Queue
	repo          repository.TaskRepository
	screenshotter browser.Screenshotter
	storage       storage.Storage
}

func NewPool(
	workerCount int,
	maxRetryCount int,
	q queue.Queue,
	repo repository.TaskRepository,
	screenshotter browser.Screenshotter,
	storage storage.Storage,
) *Pool {
	return &Pool{
		workerCount:   workerCount,
		maxRetryCount: maxRetryCount,
		queue:         q,
		repo:          repo,
		screenshotter: screenshotter,
		storage:       storage,
	}
}

func (p *Pool) Start() {
	for i := 0; i < p.workerCount; i++ {
		go p.worker(i)
	}
}

func (p *Pool) worker(id int) {
	for {
		task, err := p.queue.Pop()
		if err != nil {
			log.Println("queue error:", err)
			continue
		}

		task.Status = model.StatusProcessing
		task.UpdatedAt = time.Now()
		if err := p.repo.Update(task); err != nil {
			log.Printf("[Worker %d] update processing status failed for %s: %v\n", id, task.ID, err)
			continue
		}

		log.Printf("[Worker %d] processing %s (retry=%d)\n", id, task.ID, task.RetryCount)

		img, err := p.screenshotter.Capture(context.Background(), task.URL)
		if err != nil {
			p.handleFailure(id, &task, "capture screenshot failed: "+err.Error())
			continue
		}

		resultURL, err := p.storage.Save(context.Background(), task.ID, img)
		if err != nil {
			p.handleFailure(id, &task, "save screenshot failed: "+err.Error())
			continue
		}

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

func (p *Pool) handleFailure(workerID int, task *model.Task, errMsg string) {
	if task.RetryCount < p.maxRetryCount {
		task.RetryCount++
		task.Status = model.StatusPending
		task.ErrorMsg = errMsg
		task.UpdatedAt = time.Now()

		if err := p.repo.Update(*task); err != nil {
			log.Printf("[Worker %d] update retry status failed for %s: %v\n", workerID, task.ID, err)
			return
		}
		if err := p.queue.Push(*task); err != nil {
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
