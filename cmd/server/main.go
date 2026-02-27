package main

import (
	"log"
	"time"
	"webpage-cache/internal/api"
	"webpage-cache/internal/api/handler"
	"webpage-cache/internal/browser"
	"webpage-cache/internal/queue"
	"webpage-cache/internal/repository"
	"webpage-cache/internal/service"
	"webpage-cache/internal/storage"
	"webpage-cache/internal/worker"
)

func main() {
	repo := repository.NewMemoryTaskRepository()
	q := queue.NewMemoryQueue(100)

	screenshotter := browser.NewChromedpScreenshotter(30 * time.Second)
	defer screenshotter.Close()

	localStorage := storage.NewLocalStorage("./data/screenshots", "/static/screenshots")

	pool := worker.NewPool(5, q, repo, screenshotter, localStorage)
	pool.Start()

	svc := service.NewScreenshotService(q, repo)
	h := handler.NewScreenshotHandler(svc)
	r := api.NewRouter(h)

	log.Println("Server started at :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

