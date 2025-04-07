package worker

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jmoloko/taskmange/internal/cache"
	"github.com/jmoloko/taskmange/internal/domain/models"
	"github.com/jmoloko/taskmange/internal/domain/repository"
	"github.com/jmoloko/taskmange/internal/logger"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTaskService реализует интерфейс TaskService для тестирования
type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) CreateTask(ctx context.Context, userID string, task models.Task) (models.Task, error) {
	args := m.Called(ctx, userID, task)
	return args.Get(0).(models.Task), args.Error(1)
}

func (m *MockTaskService) GetUserTask(ctx context.Context, userID, taskID string) (models.Task, error) {
	args := m.Called(ctx, userID, taskID)
	return args.Get(0).(models.Task), args.Error(1)
}

func (m *MockTaskService) GetUserTasks(ctx context.Context, userID string, filters models.TaskFilters) ([]models.Task, error) {
	args := m.Called(ctx, userID, filters)
	return args.Get(0).([]models.Task), args.Error(1)
}

func (m *MockTaskService) UpdateUserTask(ctx context.Context, userID string, task models.Task) (models.Task, error) {
	args := m.Called(ctx, userID, task)
	return args.Get(0).(models.Task), args.Error(1)
}

func (m *MockTaskService) DeleteUserTask(ctx context.Context, userID, taskID string) error {
	args := m.Called(ctx, userID, taskID)
	return args.Error(0)
}

func (m *MockTaskService) ImportTasks(ctx context.Context, userID string, tasks []models.Task) error {
	args := m.Called(ctx, userID, tasks)
	return args.Error(0)
}

func (m *MockTaskService) ExportUserTasks(ctx context.Context, userID string) ([]models.Task, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Task), args.Error(1)
}

func (m *MockTaskService) GetUserAnalytics(ctx context.Context, userID string, period string) (models.Analytics, error) {
	args := m.Called(ctx, userID, period)
	return args.Get(0).(models.Analytics), args.Error(1)
}

func (m *MockTaskService) Delete(ctx context.Context, taskID, userID string) error {
	args := m.Called(ctx, taskID, userID)
	return args.Error(0)
}

func (m *MockTaskService) GetActiveUsers(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockTaskService) GetAll(ctx context.Context, userID string, filters models.TaskFilters) ([]models.Task, error) {
	args := m.Called(ctx, userID, filters)
	return args.Get(0).([]models.Task), args.Error(1)
}

func (m *MockTaskService) GetAnalytics(ctx context.Context, userID string, period string) (models.Analytics, error) {
	args := m.Called(ctx, userID, period)
	return args.Get(0).(models.Analytics), args.Error(1)
}

// MockCache реализует интерфейс AnalyticsCache для тестирования
type MockCache struct {
	mock.Mock
}

func (m *MockCache) GetAnalytics(ctx context.Context) (repository.CachedAnalytics, error) {
	args := m.Called(ctx)
	return args.Get(0).(repository.CachedAnalytics), args.Error(1)
}

func (m *MockCache) SetAnalytics(ctx context.Context, analytics repository.CachedAnalytics) error {
	args := m.Called(ctx, analytics)
	return args.Error(0)
}

func (m *MockCache) GetUserAnalytics(ctx context.Context, userID, period string) (*repository.CachedAnalytics, error) {
	args := m.Called(ctx, userID, period)
	return args.Get(0).(*repository.CachedAnalytics), args.Error(1)
}

func (m *MockCache) SetUserAnalytics(ctx context.Context, analytics repository.CachedAnalytics) error {
	args := m.Called(ctx, analytics)
	return args.Error(0)
}

func (m *MockCache) InvalidateUserAnalytics(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// MockLogger реализует интерфейс Logger для тестирования
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Info(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Error(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Fatal(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Warn(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) WithFields(fields map[string]interface{}) logger.Logger {
	args := m.Called(fields)
	return args.Get(0).(logger.Logger)
}

func (m *MockLogger) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestBackgroundWorker_Start(t *testing.T) {
	mockTaskService := new(MockTaskService)
	mockCache := new(MockCache)
	mockLogger := new(MockLogger)

	worker := NewBackgroundWorker(mockTaskService, mockCache, mockLogger)
	assert.NotNil(t, worker)

	worker.Start()

	time.Sleep(100 * time.Millisecond)
	worker.Stop()
}

func TestBackgroundWorker_CleanupExpiredTasks(t *testing.T) {
	mockTaskService := new(MockTaskService)
	mockCache := new(MockCache)
	mockLogger := new(MockLogger)

	worker := NewBackgroundWorker(mockTaskService, mockCache, mockLogger)
	assert.NotNil(t, worker)

	expiredTasks := []models.Task{
		{ID: "1", UserID: "user1", Title: "Expired Task 1"},
		{ID: "2", UserID: "user2", Title: "Expired Task 2"},
	}

	mockTaskService.On("GetAll", mock.Anything, "", mock.Anything).Return(expiredTasks, nil)
	mockTaskService.On("Delete", mock.Anything, "1", "user1").Return(nil)
	mockTaskService.On("Delete", mock.Anything, "2", "user2").Return(nil)
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

	err := worker.cleanupExpiredTasks()
	assert.NoError(t, err)

	mockTaskService.AssertExpectations(t)
}

func TestBackgroundWorker_GenerateAnalytics(t *testing.T) {
	mockTaskService := new(MockTaskService)
	mockCache := new(MockCache)
	mockLogger := new(MockLogger)

	worker := NewBackgroundWorker(mockTaskService, mockCache, mockLogger)
	assert.NotNil(t, worker)

	users := []string{"user1", "user2"}
	analytics := models.Analytics{
		StatusCount: map[models.Status]int{
			models.StatusDone:    5,
			models.StatusPending: 3,
		},
		AvgCompletionTime: 24.5,
	}

	mockTaskService.On("GetActiveUsers", mock.Anything).Return(users, nil)
	for _, userID := range users {
		for _, period := range []string{"day", "week", "month"} {
			mockTaskService.On("GetAnalytics", mock.Anything, userID, period).Return(analytics, nil)
			mockCache.On("SetUserAnalytics", mock.Anything, mock.MatchedBy(func(a repository.CachedAnalytics) bool {
				return a.UserID == userID && a.Period == period
			})).Return(nil)
		}
	}
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

	err := worker.generateAnalytics()
	assert.NoError(t, err)

	mockTaskService.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestBackgroundWorker_StartStop(t *testing.T) {
	// Подготовка моков
	mockTaskService := new(MockTaskService)
	redisClient := redis.NewClient(&redis.Options{})
	mockCache := cache.NewRedisCache(redisClient)
	mockLogger := new(MockLogger)

	// Создаем worker
	worker := NewBackgroundWorker(mockTaskService, mockCache, mockLogger)

	// Запускаем worker
	worker.Start()

	// Даем немного времени на запуск горутин
	time.Sleep(100 * time.Millisecond)

	// Останавливаем worker
	worker.Stop()

	// Проверяем, что все горутины завершились
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Проверяем, что повторный вызов Stop не вызывает панику
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Error("Stop() вызвал панику при повторном вызове")
				}
			}()
			worker.Stop()
		}()
	}()

	// Ждем с таймаутом
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Все ок
	case <-time.After(time.Second):
		t.Fatal("Worker.Stop() заблокировался")
	}
}
