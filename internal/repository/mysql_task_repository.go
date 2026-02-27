package repository

import (
	"database/sql"
	"webpage-cache/internal/model"
)

type MySQLTaskRepository struct {
	db *sql.DB
}

func NewMySQLTaskRepository(db *sql.DB) *MySQLTaskRepository {
	return &MySQLTaskRepository{db: db}
}

func (r *MySQLTaskRepository) Create(task model.Task) error {
	_, err := r.db.Exec(
		`INSERT INTO tasks (id, url, status, result_url, error_msg, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		task.ID,
		task.URL,
		string(task.Status),
		task.ResultURL,
		task.ErrorMsg,
		task.CreatedAt,
		task.UpdatedAt,
	)
	return err
}

func (r *MySQLTaskRepository) Update(task model.Task) error {
	_, err := r.db.Exec(
		`UPDATE tasks
		 SET url = ?, status = ?, result_url = ?, error_msg = ?, updated_at = ?
		 WHERE id = ?`,
		task.URL,
		string(task.Status),
		task.ResultURL,
		task.ErrorMsg,
		task.UpdatedAt,
		task.ID,
	)
	return err
}

func (r *MySQLTaskRepository) GetByID(id string) (model.Task, bool) {
	var task model.Task
	var status string

	err := r.db.QueryRow(
		`SELECT id, url, status, result_url, error_msg, created_at, updated_at
		 FROM tasks WHERE id = ?`,
		id,
	).Scan(
		&task.ID,
		&task.URL,
		&status,
		&task.ResultURL,
		&task.ErrorMsg,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return model.Task{}, false
	}
	if err != nil {
		return model.Task{}, false
	}

	task.Status = model.TaskStatus(status)
	return task, true
}

