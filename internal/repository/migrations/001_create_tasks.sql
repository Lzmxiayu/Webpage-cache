CREATE TABLE IF NOT EXISTS tasks (
    id VARCHAR(64) PRIMARY KEY,
    url TEXT NOT NULL,
    status VARCHAR(32) NOT NULL,
    result_url TEXT NOT NULL DEFAULT '',
    error_msg TEXT NOT NULL DEFAULT '',
    retry_count INT NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    INDEX idx_tasks_status_created_at (status, created_at)
);

