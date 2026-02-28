package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

type Config struct {
	MySQLDSN          string
	RedisAddr         string
	RedisPassword     string
	RedisDB           int
	RedisQueueKey     string
	MaxRetryCount     int
	WorkerCount       int
	BrowserPoolSize   int
	MaxTabsPerBrowser int
	QueueSize         int
	ProxyURLs         []string
	ScreenshotDir     string
	ScreenshotBaseURL string
	ScreenshotTimeout time.Duration
	TaskExecTimeout   time.Duration
	ShutdownTimeout   time.Duration
	HTTPAddr          string
	LogLevel          string
}

func Load() (Config, error) {
	cfg := Config{
		MySQLDSN:          os.Getenv("MYSQL_DSN"),
		RedisAddr:         getStringEnv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword:     os.Getenv("REDIS_PASSWORD"),
		RedisDB:           getIntEnv("REDIS_DB", 0),
		RedisQueueKey:     getStringEnv("REDIS_QUEUE_KEY", "screenshot:tasks"),
		MaxRetryCount:     getIntEnv("MAX_RETRY_COUNT", 2),
		WorkerCount:       getIntEnv("WORKER_COUNT", 5),
		BrowserPoolSize:   getIntEnv("BROWSER_POOL_SIZE", 2),
		MaxTabsPerBrowser: getIntEnv("MAX_TABS_PER_BROWSER", 1),
		QueueSize:         getIntEnv("QUEUE_SIZE", 100),
		ProxyURLs:         splitCSVEnv("PROXY_URLS"),
		ScreenshotDir:     getStringEnv("SCREENSHOT_DIR", "./data/screenshots"),
		ScreenshotBaseURL: getStringEnv("SCREENSHOT_BASE_URL", "/static/screenshots"),
		ScreenshotTimeout: time.Duration(getIntEnv("SHOT_TIMEOUT_SEC", 30)) * time.Second,
		TaskExecTimeout:   time.Duration(getIntEnv("TASK_TIMEOUT_SEC", 45)) * time.Second,
		ShutdownTimeout:   time.Duration(getIntEnv("SHUTDOWN_TIMEOUT_SEC", 10)) * time.Second,
		HTTPAddr:          getStringEnv("HTTP_ADDR", ":8080"),
		LogLevel:          getStringEnv("LOG_LEVEL", "info"),
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
	if cfg.MaxRetryCount < 0 {
		return Config{}, fmt.Errorf("MAX_RETRY_COUNT must be >= 0")
	}
	if cfg.BrowserPoolSize <= 0 {
		return Config{}, fmt.Errorf("BROWSER_POOL_SIZE must be > 0")
	}
	if cfg.MaxTabsPerBrowser <= 0 {
		return Config{}, fmt.Errorf("MAX_TABS_PER_BROWSER must be > 0")
	}
	if cfg.TaskExecTimeout <= 0 {
		return Config{}, fmt.Errorf("TASK_TIMEOUT_SEC must be > 0")
	}
	if cfg.ShutdownTimeout <= 0 {
		return Config{}, fmt.Errorf("SHUTDOWN_TIMEOUT_SEC must be > 0")
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

func splitCSVEnv(key string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}
