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
