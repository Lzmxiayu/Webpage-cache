package main

import (
	"log"
	"webpage-cache/internal/api"
	"webpage-cache/internal/api/handler"
	"webpage-cache/internal/model"
	"webpage-cache/internal/repository"
	"webpage-cache/internal/service"
	"webpage-cache/internal/worker"
)

func main() {

	// 创建任务通道
	jobChan := make(chan model.Task, 100)

	repo := repository.NewMemoryTaskRepository()

	// 启动 worker pool
	pool := worker.NewPool(5, jobChan, repo)
	pool.Start()

	// 初始化 service
	svc := service.NewScreenshotService(jobChan, repo)

	// 初始化 handler
	h := handler.NewScreenshotHandler(svc)

	// 启动 HTTP 服务
	r := api.NewRouter(h)

	log.Println("Server started at :8080")
	r.Run(":8080")
}
