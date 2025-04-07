package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/jmoloko/taskmange/internal/domain/models"
	"github.com/lib/pq"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// создаём новую задачу
func (r *TaskRepository) Create(ctx context.Context, task *models.Task) error {
	query := `
		INSERT INTO tasks (id, title, description, status, priority, user_id, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	slog.Info("Creating task in database",
		"task_id", task.ID,
		"user_id", task.UserID,
		"title", task.Title,
		"status", task.Status,
		"priority", task.Priority,
		"due_date", task.DueDate)

	result, err := r.db.ExecContext(ctx, query,
		task.ID, task.Title, task.Description, task.Status, task.Priority,
		task.UserID, task.DueDate, task.CreatedAt, task.UpdatedAt)
	if err != nil {
		slog.Error("Failed to create task in database",
			"error", err,
			"task_id", task.ID,
			"user_id", task.UserID,
			"error_details", err.Error())

		if pqErr, ok := err.(*pq.Error); ok {
			slog.Error("PostgreSQL error details",
				"code", pqErr.Code,
				"constraint", pqErr.Constraint,
				"detail", pqErr.Detail,
				"message", pqErr.Message)
		}
		return fmt.Errorf("failed to create task: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		slog.Error("Failed to get affected rows",
			"error", err,
			"task_id", task.ID,
			"user_id", task.UserID)
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		slog.Error("No rows affected",
			"task_id", task.ID,
			"user_id", task.UserID)
		return fmt.Errorf("no rows affected")
	}

	return nil
}

// обновляем существующую задачу
func (r *TaskRepository) Update(ctx context.Context, task *models.Task) error {
	query := `
		UPDATE tasks
		SET title = $1, description = $2, status = $3, priority = $4, due_date = $5, updated_at = $6
		WHERE id = $7 AND user_id = $8
	`
	result, err := r.db.ExecContext(ctx, query,
		task.Title, task.Description, task.Status, task.Priority,
		task.DueDate, task.UpdatedAt, task.ID, task.UserID)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("task not found or not owned by user")
	}

	return nil
}

// удаляет задачу по ID
func (r *TaskRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM tasks WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("task not found")
	}

	return nil
}

// получаем задачу по ID
func (r *TaskRepository) GetByID(ctx context.Context, id string) (*models.Task, error) {
	query := `
		SELECT id, title, description, status, priority, user_id, due_date, created_at, updated_at, completed_at
		FROM tasks
		WHERE id = $1
	`
	var task models.Task
	var completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID, &task.Title, &task.Description, &task.Status, &task.Priority,
		&task.UserID, &task.DueDate, &task.CreatedAt, &task.UpdatedAt, &completedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("task not found")
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}

	return &task, nil
}

// список задач с применением фильтров
func (r *TaskRepository) GetAll(ctx context.Context, filters models.TaskFilters) ([]models.Task, error) {
	query := `
		SELECT id, title, description, status, priority, user_id, due_date, created_at, updated_at, completed_at
		FROM tasks
		WHERE user_id = $1
	`
	args := []interface{}{filters.UserID}
	argCount := 2

	// Добавляем фильтры, если они указаны
	if filters.Status != "" {
		query += ` AND status = $` + strconv.Itoa(argCount)
		args = append(args, filters.Status)
		argCount++
	}

	if filters.Priority != "" {
		query += ` AND priority = $` + strconv.Itoa(argCount)
		args = append(args, filters.Priority)
		argCount++
	}

	if filters.DueDate != nil {
		query += ` AND due_date::date = $` + strconv.Itoa(argCount) + `::date`
		args = append(args, filters.DueDate)
		argCount++
	}

	if filters.Search != "" {
		query += ` AND (title ILIKE $` + strconv.Itoa(argCount) + ` OR description ILIKE $` + strconv.Itoa(argCount) + `)`
		args = append(args, "%"+filters.Search+"%")
		argCount++
	}

	query += ` ORDER BY due_date ASC, priority DESC, created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		var completedAt sql.NullTime

		err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.Status, &task.Priority,
			&task.UserID, &task.DueDate, &task.CreatedAt, &task.UpdatedAt, &completedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}

		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tasks: %w", err)
	}

	return tasks, nil
}
