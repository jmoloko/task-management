package service

import (
	"context"
	"testing"
	"time"

	"github.com/jmoloko/taskmange/internal/domain/models"
	"github.com/jmoloko/taskmange/internal/domain/repository"
	"github.com/jmoloko/taskmange/internal/logger"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	mockRepo   *MockTaskRepository
	mockLogger *MockLogger
	mockCache  *MockCache
)

// MockTaskRepository реализует интерфейс repository.TaskRepository для тестов
type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) Create(ctx context.Context, task *models.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepository) GetByID(ctx context.Context, id string) (*models.Task, error) {
	args := m.Called(ctx, id)
	if task, ok := args.Get(0).(*models.Task); ok {
		return task, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTaskRepository) GetAll(ctx context.Context, filters models.TaskFilters) ([]models.Task, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]models.Task), args.Error(1)
}

func (m *MockTaskRepository) Update(ctx context.Context, task *models.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockLogger реализует интерфейс logger.Logger для тестов
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

// MockCache implements repository.AnalyticsCache
type MockCache struct {
	mock.Mock
}

func (m *MockCache) GetUserAnalytics(ctx context.Context, userID, period string) (*repository.CachedAnalytics, error) {
	args := m.Called(ctx, userID, period)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
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

func TestCreate(t *testing.T) {
	tests := []struct {
		name    string
		task    models.Task
		want    models.Task
		setup   func()
		wantErr bool
	}{
		{
			name: "Valid task with all fields",
			task: models.Task{
				Title:       "Test Task",
				Description: "Test Description",
				Status:      models.StatusPending,
				Priority:    models.PriorityHigh,
				DueDate:     time.Now().Add(24 * time.Hour),
			},
			want: models.Task{
				Title:       "Test Task",
				Description: "Test Description",
				Status:      models.StatusPending,
				Priority:    models.PriorityHigh,
				DueDate:     time.Now().Add(24 * time.Hour),
			},
			setup: func() {
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Task")).Return(nil).Once()
				mockLogger.On("Info", "Creating new task", mock.Anything).Return()
				mockLogger.On("Info", "Task created successfully", mock.Anything).Return()
			},
			wantErr: false,
		},
		{
			name: "Valid task with minimal fields",
			task: models.Task{
				Title: "Test Task",
			},
			want: models.Task{
				Title:    "Test Task",
				Status:   models.StatusPending,           // Default value
				Priority: models.PriorityMedium,          // Default value
				DueDate:  time.Now().Add(24 * time.Hour), // Default value
			},
			setup: func() {
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Task")).Return(nil).Once()
				mockLogger.On("Info", "Creating new task", mock.Anything).Return()
				mockLogger.On("Info", "Setting default status: pending", mock.Anything).Return()
				mockLogger.On("Info", "Setting default priority: medium", mock.Anything).Return()
				mockLogger.On("Info", "Setting default due date", mock.Anything).Return()
				mockLogger.On("Info", "Task created successfully", mock.Anything).Return()
			},
			wantErr: false,
		},
		{
			name: "Invalid task - empty title",
			task: models.Task{
				Description: "Test Description",
			},
			want: models.Task{},
			setup: func() {
				mockLogger.On("Info", "Creating new task", mock.Anything).Return()
				mockLogger.On("Error", "Invalid task data: title is required", mock.Anything).Return()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo = new(MockTaskRepository)
			mockLogger = new(MockLogger)
			mockCache = new(MockCache)
			tt.setup()

			service := NewTaskService(mockRepo, mockCache, mockLogger)
			got, err := service.CreateTask(context.Background(), "user1", tt.task)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.want, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.Title, got.Title)
				assert.Equal(t, tt.want.Description, got.Description)
				assert.Equal(t, tt.want.Status, got.Status)
				assert.Equal(t, tt.want.Priority, got.Priority)
				assert.Equal(t, tt.want.DueDate.Unix(), got.DueDate.Unix())
			}

			mockRepo.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestGetByID(t *testing.T) {
	mockRepo = new(MockTaskRepository)
	mockLogger = new(MockLogger)
	mockCache = new(MockCache)
	service := NewTaskService(mockRepo, mockCache, mockLogger)

	taskID := "test-id"
	userID := "user1"
	task := &models.Task{
		ID:          taskID,
		UserID:      userID,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      models.StatusPending,
		Priority:    models.PriorityHigh,
		DueDate:     time.Now().Add(24 * time.Hour),
	}

	tests := []struct {
		name        string
		taskID      string
		userID      string
		setup       func()
		want        models.Task
		wantErr     bool
		wantErrType error
	}{
		{
			name:   "Existing task - authorized user",
			taskID: taskID,
			userID: userID,
			setup: func() {
				mockRepo.On("GetByID", mock.Anything, taskID).Return(task, nil).Once()
			},
			want:    *task,
			wantErr: false,
		},
		{
			name:   "Existing task - unauthorized user",
			taskID: taskID,
			userID: "other-user",
			setup: func() {
				mockRepo.On("GetByID", mock.Anything, taskID).Return(&models.Task{
					ID:     taskID,
					UserID: userID,
				}, nil).Once()
			},
			want:        models.Task{},
			wantErr:     true,
			wantErrType: ErrAccessDenied,
		},
		{
			name:   "Non-existing task",
			taskID: taskID,
			userID: userID,
			setup: func() {
				mockRepo.On("GetByID", mock.Anything, taskID).Return(nil, ErrTaskNotFound).Once()
			},
			want:        models.Task{},
			wantErr:     true,
			wantErrType: ErrTaskNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			got, err := service.GetUserTask(context.Background(), tt.userID, tt.taskID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErrType, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockRepo.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestGetAll(t *testing.T) {
	mockRepo = new(MockTaskRepository)
	mockLogger = new(MockLogger)
	mockCache = new(MockCache)
	service := NewTaskService(mockRepo, mockCache, mockLogger)

	userID := "user1"
	tasks := []models.Task{
		{ID: "1", Title: "Task 1", UserID: userID},
		{ID: "2", Title: "Task 2", UserID: userID},
	}

	tests := []struct {
		name    string
		userID  string
		filters models.TaskFilters
		want    []models.Task
		wantErr bool
	}{
		{
			name:   "Get all tasks - no filters",
			userID: userID,
			filters: models.TaskFilters{
				UserID: userID,
			},
			want:    tasks,
			wantErr: false,
		},
		{
			name:   "Get tasks with status filter",
			userID: userID,
			filters: models.TaskFilters{
				UserID: userID,
				Status: models.StatusPending,
			},
			want:    tasks[:1],
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.On("GetAll", mock.Anything, tt.filters).Return(tt.want, nil).Once()

			got, err := service.GetUserTasks(context.Background(), userID, tt.filters)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockRepo.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestUpdate(t *testing.T) {
	mockRepo = new(MockTaskRepository)
	mockLogger = new(MockLogger)
	mockCache = new(MockCache)
	service := NewTaskService(mockRepo, mockCache, mockLogger)

	taskID := "test-id"
	userID := "user1"
	existingTask := &models.Task{
		ID:     taskID,
		Title:  "Old Title",
		UserID: userID,
	}

	tests := []struct {
		name    string
		taskID  string
		userID  string
		update  models.Task
		setup   func()
		wantErr bool
	}{
		{
			name:   "Valid update - authorized user",
			taskID: taskID,
			userID: userID,
			update: models.Task{
				ID:          taskID,
				Title:       "New Title",
				Description: "New Description",
			},
			setup: func() {
				mockRepo.On("GetByID", mock.Anything, taskID).Return(existingTask, nil).Once()
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Task")).Return(nil).Once()
				mockLogger.On("Info", "Updating task", mock.Anything).Return()
				mockLogger.On("Info", "Task updated successfully", mock.Anything).Return()
			},
			wantErr: false,
		},
		{
			name:   "Update - unauthorized user",
			taskID: taskID,
			userID: "other-user",
			update: models.Task{
				ID:    taskID,
				Title: "New Title",
			},
			setup: func() {
				mockRepo.On("GetByID", mock.Anything, taskID).Return(existingTask, nil).Once()
				mockLogger.On("Info", "Updating task", mock.Anything).Return()
				mockLogger.On("Error", "Access denied to task", mock.Anything).Return()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			got, err := service.UpdateUserTask(context.Background(), tt.userID, tt.update)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, models.Task{}, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.update.Title, got.Title)
				assert.Equal(t, tt.update.Description, got.Description)
			}

			mockRepo.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestDelete(t *testing.T) {
	mockRepo = new(MockTaskRepository)
	mockLogger = new(MockLogger)
	mockCache = new(MockCache)
	service := NewTaskService(mockRepo, mockCache, mockLogger)

	taskID := "test-id"
	userID := "user1"
	existingTask := &models.Task{
		ID:     taskID,
		Title:  "Test Task",
		UserID: userID,
	}

	tests := []struct {
		name    string
		taskID  string
		userID  string
		setup   func()
		wantErr bool
	}{
		{
			name:   "Delete - authorized user",
			taskID: taskID,
			userID: userID,
			setup: func() {
				mockRepo.On("GetByID", mock.Anything, taskID).Return(existingTask, nil).Once()
				mockRepo.On("Delete", mock.Anything, taskID).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:   "Delete - unauthorized user",
			taskID: taskID,
			userID: "other-user",
			setup: func() {
				mockRepo.On("GetByID", mock.Anything, taskID).Return(existingTask, nil).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			err := service.DeleteUserTask(context.Background(), tt.userID, tt.taskID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, ErrAccessDenied, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestGetAnalytics(t *testing.T) {
	mockRepo = new(MockTaskRepository)
	mockLogger = new(MockLogger)
	mockCache = new(MockCache)
	service := NewTaskService(mockRepo, mockCache, mockLogger)

	userID := "user1"
	now := time.Now()
	tasks := []models.Task{
		{
			ID:          "1",
			Title:       "Task 1",
			UserID:      userID,
			Status:      models.StatusDone,
			Priority:    models.PriorityHigh,
			CreatedAt:   now.Add(-48 * time.Hour),
			CompletedAt: &now,
			DueDate:     now.Add(24 * time.Hour),
		},
		{
			ID:        "2",
			Title:     "Task 2",
			UserID:    userID,
			Status:    models.StatusPending,
			Priority:  models.PriorityMedium,
			CreatedAt: now.Add(-24 * time.Hour),
			DueDate:   now.Add(-1 * time.Hour),
		},
	}

	tests := []struct {
		name    string
		userID  string
		period  string
		setup   func()
		want    models.Analytics
		wantErr bool
	}{
		{
			name:   "Get analytics - week period",
			userID: userID,
			period: "week",
			setup: func() {
				filters := models.TaskFilters{
					UserID: userID,
				}
				mockRepo.On("GetAll", mock.Anything, filters).Return(tasks, nil).Once()
				mockCache.On("GetUserAnalytics", mock.Anything, userID, "week").Return(nil, redis.Nil).Once()
				mockCache.On("SetUserAnalytics", mock.Anything, mock.MatchedBy(func(analytics repository.CachedAnalytics) bool {
					return analytics.UserID == userID && analytics.Period == "week"
				})).Return(nil).Once()
				mockLogger.On("Error", mock.Anything, mock.Anything).Return()
			},
			want: models.Analytics{
				StatusCount: map[models.Status]int{
					models.StatusDone:    1,
					models.StatusPending: 1,
				},
				PriorityCount: map[models.Priority]int{
					models.PriorityHigh:   1,
					models.PriorityMedium: 1,
				},
				AvgCompletionTime:    float64(48),
				OnTimeCompletionRate: 1.0,
				OverdueTasks:         1,
				Period:               "week",
				GeneratedAt:          now,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			got, err := service.GetUserAnalytics(context.Background(), userID, tt.period)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.StatusCount, got.StatusCount)
				assert.Equal(t, tt.want.PriorityCount, got.PriorityCount)
				assert.Equal(t, tt.want.OverdueTasks, got.OverdueTasks)
				assert.Equal(t, tt.want.Period, got.Period)
				assert.NotZero(t, got.AvgCompletionTime)
				assert.NotZero(t, got.OnTimeCompletionRate)
			}
		})
	}
}
