package main

import (
	"database/sql"
	"log"
	"webpage-cache/internal/api"
	"webpage-cache/internal/api/handler"
	"webpage-cache/internal/browser"
	"webpage-cache/internal/config"
	"webpage-cache/internal/queue"
	"webpage-cache/internal/repository"
	"webpage-cache/internal/service"
	"webpage-cache/internal/storage"
	"webpage-cache/internal/worker"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("mysql", cfg.MySQLDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	repo := repository.NewMySQLTaskRepository(db)
	q := queue.NewMemoryQueue(cfg.QueueSize)

	screenshotter := browser.NewChromedpScreenshotter(cfg.ScreenshotTimeout)
	defer screenshotter.Close()

	localStorage := storage.NewLocalStorage(cfg.ScreenshotDir, cfg.ScreenshotBaseURL)

	pool := worker.NewPool(cfg.WorkerCount, q, repo, screenshotter, localStorage)
	pool.Start()

	svc := service.NewScreenshotService(q, repo)
	h := handler.NewScreenshotHandler(svc)
	r := api.NewRouter(h)

	log.Printf("Server started at %s\n", cfg.HTTPAddr)
	if err := r.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}

