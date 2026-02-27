package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

type Config struct {
	MySQLDSN          string
	RedisAddr         string
	RedisPassword     string
	RedisDB           int
	RedisQueueKey     string
	WorkerCount       int
	QueueSize         int
	ScreenshotDir     string
	ScreenshotBaseURL string
	ScreenshotTimeout time.Duration
	HTTPAddr          string
}

func Load() (Config, error) {
	cfg := Config{
		MySQLDSN:          os.Getenv("MYSQL_DSN"),
		RedisAddr:         getStringEnv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword:     os.Getenv("REDIS_PASSWORD"),
		RedisDB:           getIntEnv("REDIS_DB", 0),
		RedisQueueKey:     getStringEnv("REDIS_QUEUE_KEY", "screenshot:tasks"),
		WorkerCount:       getIntEnv("WORKER_COUNT", 5),
		QueueSize:         getIntEnv("QUEUE_SIZE", 100),
		ScreenshotDir:     getStringEnv("SCREENSHOT_DIR", "./data/screenshots"),
		ScreenshotBaseURL: getStringEnv("SCREENSHOT_BASE_URL", "/static/screenshots"),
		ScreenshotTimeout: time.Duration(getIntEnv("SHOT_TIMEOUT_SEC", 30)) * time.Second,
		HTTPAddr:          getStringEnv("HTTP_ADDR", ":8080"),
	}

	if cfg.MySQLDSN == "" {
		return Config{}, fmt.Errorf("MYSQL_DSN is required")
	}
	if cfg.RedisAddr == "" {
		return Config{}, fmt.Errorf("REDIS_ADDR is required")
	}
	if cfg.RedisQueueKey == "" {
		return Config{}, fmt.Errorf("REDIS_QUEUE_KEY is required")
	}

	mysqlCfg, err := mysqlDriver.ParseDSN(cfg.MySQLDSN)
	if err != nil {
		return Config{}, fmt.Errorf("invalid MYSQL_DSN: %w", err)
	}
	mysqlCfg.ParseTime = true
	cfg.MySQLDSN = mysqlCfg.FormatDSN()

	return cfg, nil
}

func getIntEnv(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}

	return v
}

func getStringEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
