package repository

import (
	"context"
	"time"

	"github.com/jmoloko/taskmange/internal/domain/models"
)

// TaskCreator создание задач
type TaskCreator interface {
	Create(ctx context.Context, task *models.Task) error
}

// TaskReader чтение задач
type TaskReader interface {
	GetByID(ctx context.Context, id string) (*models.Task, error)
	GetAll(ctx context.Context, filters models.TaskFilters) ([]models.Task, error)
}

// TaskUpdater обновление задач
type TaskUpdater interface {
	Update(ctx context.Context, task *models.Task) error
}

// TaskDeleter удаление задач
type TaskDeleter interface {
	Delete(ctx context.Context, id string) error
}

// TaskRepository объединяет все операции с задачами (для обратной совместимости)
type TaskRepository interface {
	TaskCreator
	TaskReader
	TaskUpdater
	TaskDeleter
}

// UserCreator создание пользователя
type UserCreator interface {
	Create(ctx context.Context, user *models.User) error
}

// UserReader получение пользователя
type UserReader interface {
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
}

// UserRepository объединяет все операции с пользователями (для обратной совместимости)
type UserRepository interface {
	UserCreator
	UserReader
}

// AnalyticsReader чтение аналитики из кэша
type AnalyticsReader interface {
	GetUserAnalytics(ctx context.Context, userID, period string) (*CachedAnalytics, error)
}

// AnalyticsWriter запись аналитики в кэш
type AnalyticsWriter interface {
	SetUserAnalytics(ctx context.Context, analytics CachedAnalytics) error
}

// AnalyticsInvalidator инвалидация кэша аналитики
type AnalyticsInvalidator interface {
	InvalidateUserAnalytics(ctx context.Context, userID string) error
}

// AnalyticsCache объединяет операции с кэшем аналитики
type AnalyticsCache interface {
	AnalyticsReader
	AnalyticsWriter
	AnalyticsInvalidator
}

// CachedAnalytics представляет данные аналитики в кэше
type CachedAnalytics struct {
	UserID    string           `json:"user_id"`
	Period    string           `json:"period"`
	Analytics models.Analytics `json:"analytics"`
	CachedAt  time.Time        `json:"cached_at"`
}

// Repositories содержит все репозитории (для обратной совместимости)
type Repositories struct {
	Tasks     TaskRepository
	Users     UserRepository
	Analytics AnalyticsCache
}
