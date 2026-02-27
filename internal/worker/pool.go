package worker

import (
	"log"
	"time"
	"webpage-cache/internal/model"
	"webpage-cache/internal/repository"
)

type Pool struct {
	workerCount int
	jobChan     chan model.Task
	repo        repository.TaskRepository
}

func NewPool(workerCount int, jobChan chan model.Task, repo repository.TaskRepository) *Pool {
	return &Pool{
		workerCount: workerCount,
		jobChan:     jobChan,
		repo:        repo,
	}
}

func (p *Pool) Start() {
	for i := 0; i < p.workerCount; i++ {
		go p.worker(i)
	}
}

func (p *Pool) worker(id int) {
	for task := range p.jobChan {

		task.Status = model.StatusProcessing
		p.repo.Update(task)

		log.Printf("[Worker %d] processing %s\n", id, task.ID)

		time.Sleep(3 * time.Second)

		task.Status = model.StatusDone
		task.ResultURL = "https://fake-cdn.com/" + task.ID + ".png"
		p.repo.Update(task)

		log.Printf("[Worker %d] finished %s\n", id, task.ID)
	}
}
