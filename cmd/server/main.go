package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"webpage-cache/internal/api"
	"webpage-cache/internal/api/handler"
	"webpage-cache/internal/browser"
	"webpage-cache/internal/config"
	"webpage-cache/internal/observability"
	"webpage-cache/internal/queue"
	"webpage-cache/internal/repository"
	"webpage-cache/internal/service"
	"webpage-cache/internal/storage"
	"webpage-cache/internal/worker"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	logger := observability.NewLogger(cfg.LogLevel)

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := sql.Open("mysql", cfg.MySQLDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.PingContext(rootCtx); err != nil {
		log.Fatal(err)
	}
	if err := repository.RunMigrations(db); err != nil {
		log.Fatal(err)
	}

	repo := repository.NewMySQLTaskRepository(db)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(rootCtx).Err(); err != nil {
		log.Fatal(fmt.Errorf("redis ping failed: %w", err))
	}

	q := queue.NewRedisQueue(redisClient, cfg.RedisQueueKey)

	screenshotter, err := browser.NewPooledChromedpScreenshotter(cfg.BrowserPoolSize, cfg.MaxTabsPerBrowser, cfg.ScreenshotTimeout)
	if err != nil {
		log.Fatal(err)
	}
	defer screenshotter.Close()

	localStorage := storage.NewLocalStorage(cfg.ScreenshotDir, cfg.ScreenshotBaseURL)

	pool := worker.NewPool(cfg.WorkerCount, cfg.MaxRetryCount, cfg.TaskExecTimeout, q, repo, screenshotter, localStorage, logger)
	pool.Start(rootCtx)

	svc := service.NewScreenshotService(q, repo)
	h := handler.NewScreenshotHandler(svc)
	r := api.NewRouter(logger, h)

	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: r,
	}

	go func() {
		logger.Info("server_started", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-rootCtx.Done()
	logger.Info("shutdown_signal_received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("http_shutdown_failed", "error", err)
	}

	pool.Wait()
	logger.Info("server_shutdown_completed")
}
