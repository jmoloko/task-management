package service

import (
	"context"

	"github.com/jmoloko/taskmange/internal/domain/models"
)

// TaskCreator создание задачи
type TaskCreator interface {
	CreateTask(ctx context.Context, userID string, task models.Task) (models.Task, error)
}

// TaskReader чтение задачи
type TaskReader interface {
	GetUserTask(ctx context.Context, userID, taskID string) (models.Task, error)
	GetUserTasks(ctx context.Context, userID string, filters models.TaskFilters) ([]models.Task, error)
	GetAll(ctx context.Context, userID string, filters models.TaskFilters) ([]models.Task, error)
	GetActiveUsers(ctx context.Context) ([]string, error)
}

// TaskUpdater обновление задачи
type TaskUpdater interface {
	UpdateUserTask(ctx context.Context, userID string, task models.Task) (models.Task, error)
}

// TaskDeleter удаление задачи
type TaskDeleter interface {
	DeleteUserTask(ctx context.Context, userID, taskID string) error
	Delete(ctx context.Context, taskID, userID string) error
}

// TaskImporter импорт задачи
type TaskImporter interface {
	ImportTasks(ctx context.Context, userID string, tasks []models.Task) error
}

// TaskExporter экспорт задачи
type TaskExporter interface {
	ExportUserTasks(ctx context.Context, userID string) ([]models.Task, error)
}

// TaskAnalytics аналитика задач
type TaskAnalytics interface {
	GetUserAnalytics(ctx context.Context, userID string, period string) (models.Analytics, error)
	GetAnalytics(ctx context.Context, userID string, period string) (models.Analytics, error)
}

// TaskManager объединяет основные операции с задачами
type TaskManager interface {
	TaskCreator
	TaskReader
	TaskUpdater
	TaskDeleter
}

// TaskDataProcessor объединяет операции обработки данных задач
type TaskDataProcessor interface {
	TaskImporter
	TaskExporter
	TaskAnalytics
}

// TaskService объединяет все операции с задачами (для обратной совместимости)
type TaskService interface {
	TaskManager
	TaskDataProcessor
}

// TaskServiceProvider фабрика для создания TaskService
type TaskServiceProvider interface {
	NewTaskService() TaskService
}
