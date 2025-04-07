package models

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// Status статус задачи
type Status string

// Priority приоритет задачи
type Priority string

// Константы для статусов задач
const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

// Константы для приоритетов задач
const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

// Value реализует интерфейс driver.Valuer для типа Priority
func (p Priority) Value() (driver.Value, error) {
	return string(p), nil
}

// Scan реализует интерфейс sql.Scanner для типа Priority
func (p *Priority) Scan(value interface{}) error {
	if value == nil {
		*p = PriorityMedium
		return nil
	}

	switch v := value.(type) {
	case []byte:
		*p = Priority(string(v))
	case string:
		*p = Priority(v)
	default:
		return fmt.Errorf("invalid priority value: %v", value)
	}
	return nil
}

// Value реализует интерфейс driver.Valuer для типа Status
func (s Status) Value() (driver.Value, error) {
	return string(s), nil
}

// Scan реализует интерфейс sql.Scanner для типа Status
func (s *Status) Scan(value interface{}) error {
	if value == nil {
		*s = StatusPending
		return nil
	}

	switch v := value.(type) {
	case []byte:
		*s = Status(string(v))
	case string:
		*s = Status(v)
	default:
		return fmt.Errorf("invalid status value: %v", value)
	}
	return nil
}

// Task представляет модель задачи
type Task struct {
	ID          string     `json:"id" db:"id"`
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	Status      Status     `json:"status" db:"status"`
	Priority    Priority   `json:"priority" db:"priority"`
	UserID      string     `json:"user_id" db:"user_id"`
	DueDate     time.Time  `json:"due_date" db:"due_date"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`
}

// TaskFilters представляет фильтры для запросов к задачам
type TaskFilters struct {
	Status   Status
	Priority Priority
	DueDate  *time.Time
	UserID   string
	Search   string
}

// Analytics представляет аналитические данные по задачам
type Analytics struct {
	// Количество задач по статусам
	StatusCount map[Status]int `json:"status_count"`

	// Количество задач по приоритетам
	PriorityCount map[Priority]int `json:"priority_count"`

	// Среднее время выполнения задачи (от создания до завершения) в часах
	AvgCompletionTime float64 `json:"avg_completion_time"`

	// Процент задач, выполненных в срок
	OnTimeCompletionRate float64 `json:"on_time_completion_rate"`

	// Текущее количество просроченных задач
	OverdueTasks int `json:"overdue_tasks"`

	// Период, за который собрана аналитика
	Period string `json:"period"`

	// Дата и время формирования отчета
	GeneratedAt time.Time `json:"generated_at"`
}
