package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoloko/taskmange/internal/domain/models"
	"github.com/jmoloko/taskmange/internal/logger"
	"github.com/jmoloko/taskmange/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockTaskService реализует интерфейс TaskService для тестов
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

// MockLogger реализует интерфейс Logger для тестов
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Info(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Warn(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Error(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Fatal(msg string, args ...interface{}) {
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

func setupTest() (*gin.Engine, *MockTaskService, *MockLogger) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	mockService := new(MockTaskService)
	mockLogger := new(MockLogger)
	handler := NewTaskHandler(mockService, mockLogger)

	// Add middleware to set user_id in context
	engine.Use(func(c *gin.Context) {
		if userID := c.GetHeader("X-User-ID"); userID != "" {
			c.Set("user_id", userID)
		}
		c.Next()
	})

	// настройка роутов
	engine.POST("/tasks", handler.CreateTask)
	engine.GET("/tasks/:id", handler.GetTask)
	engine.GET("/tasks", handler.GetTasks)
	engine.PUT("/tasks/:id", handler.UpdateTask)
	engine.DELETE("/tasks/:id", handler.DeleteTask)
	engine.POST("/tasks/import", handler.ImportTasks)
	engine.GET("/tasks/export", handler.ExportTasks)
	engine.GET("/tasks/analytics", handler.GetAnalytics)

	return engine, mockService, mockLogger
}

func TestCreateTask(t *testing.T) {
	mockService := new(MockTaskService)
	mockLogger := new(MockLogger)
	handler := NewTaskHandler(mockService, mockLogger)

	dueDate := time.Now().Add(24 * time.Hour)
	dueDateStr := dueDate.Format(time.RFC3339Nano)

	tests := []struct {
		name        string
		requestBody interface{}
		setupMocks  func()
		checkStatus int
		checkBody   gin.H
	}{
		{
			name: "Success",
			requestBody: models.Task{
				Title:       "Test Task",
				Description: "Test Description",
				Priority:    models.PriorityHigh,
				Status:      "pending",
				DueDate:     dueDate,
			},
			setupMocks: func() {
				mockService.On("CreateTask", mock.Anything, "test_user", mock.MatchedBy(func(task models.Task) bool {
					return task.Title == "Test Task" &&
						task.Description == "Test Description" &&
						task.Priority == models.PriorityHigh &&
						task.Status == "pending" &&
						task.DueDate.Equal(dueDate)
				})).Return(models.Task{
					ID:          "test_id",
					Title:       "Test Task",
					Description: "Test Description",
					Priority:    models.PriorityHigh,
					Status:      "pending",
					UserID:      "test_user",
					DueDate:     dueDate,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}, nil)
			},
			checkStatus: http.StatusCreated,
			checkBody: gin.H{
				"id":          "test_id",
				"title":       "Test Task",
				"description": "Test Description",
				"priority":    string(models.PriorityHigh),
				"status":      "pending",
				"user_id":     "test_user",
				"due_date":    dueDateStr,
			},
		},
		{
			name:        "Invalid_Request_Body",
			requestBody: "invalid json",
			setupMocks: func() {
				mockLogger.On("Error", mock.Anything, mock.Anything).Return()
			},
			checkStatus: http.StatusBadRequest,
			checkBody: gin.H{
				"error": "Invalid request body",
			},
		},
		{
			name: "Invalid_Task_Data",
			requestBody: models.Task{
				Title:    "Test Task",
				Priority: models.PriorityHigh,
				Status:   "invalid_status",
			},
			setupMocks: func() {
				mockService.On("CreateTask", mock.Anything, "test_user", mock.MatchedBy(func(task models.Task) bool {
					return task.Title == "Test Task" &&
						task.Priority == models.PriorityHigh &&
						task.Status == "invalid_status"
				})).Return(models.Task{}, service.ErrInvalidTaskData)
				mockLogger.On("Error", mock.Anything, mock.Anything).Return()
			},
			checkStatus: http.StatusBadRequest,
			checkBody: gin.H{
				"error": "Invalid task data",
			},
		},
		{
			name: "Internal_Server_Error",
			requestBody: models.Task{
				Title:    "Test Task",
				Priority: models.PriorityHigh,
				Status:   "pending",
			},
			setupMocks: func() {
				mockService.On("CreateTask", mock.Anything, "test_user", mock.MatchedBy(func(task models.Task) bool {
					return task.Title == "Test Task" &&
						task.Priority == models.PriorityHigh &&
						task.Status == "pending"
				})).Return(models.Task{}, errors.New("internal error"))
				mockLogger.On("Error", mock.Anything, mock.Anything).Return()
			},
			checkStatus: http.StatusInternalServerError,
			checkBody: gin.H{
				"error": "Failed to create task",
			},
		},
		{
			name: "Unauthorized",
			requestBody: models.Task{
				Title:    "Test Task",
				Priority: models.PriorityHigh,
				Status:   "pending",
			},
			setupMocks: func() {
				// No mocks needed as the handler should return early
			},
			checkStatus: http.StatusUnauthorized,
			checkBody: gin.H{
				"error": "Unauthorized",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			router := gin.New()
			if tt.name != "Unauthorized" {
				router.Use(func(c *gin.Context) {
					c.Set("user_id", "test_user")
					c.Next()
				})
			}
			router.POST("/tasks", handler.CreateTask)

			tt.setupMocks()

			// Create request
			var body bytes.Buffer
			if err := json.NewEncoder(&body).Encode(tt.requestBody); err != nil {
				t.Fatalf("Failed to encode request body: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/tasks", &body)
			req.Header.Set("Content-Type", "application/json")

			// Perform request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Check status
			assert.Equal(t, tt.checkStatus, w.Code)

			// Parse response
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response fields
			for key, expectedValue := range tt.checkBody {
				if key == "due_date" || key == "created_at" || key == "updated_at" {
					// Parse both times and compare them
					expectedTime, err := time.Parse(time.RFC3339Nano, expectedValue.(string))
					assert.NoError(t, err)
					actualTime, err := time.Parse(time.RFC3339Nano, response[key].(string))
					assert.NoError(t, err)
					assert.Equal(t, expectedTime.Unix(), actualTime.Unix(), "Field %s does not match", key)
				} else {
					assert.Equal(t, expectedValue, response[key], "Field %s does not match", key)
				}
			}

			// Дополнительные проверки для полей с временными метками в случае успеха
			if tt.name == "Success" {
				// Проверка наличия и правильного формата временных меток
				_, err = time.Parse(time.RFC3339Nano, response["created_at"].(string))
				assert.NoError(t, err, "created_at should be a valid RFC3339Nano timestamp")
				_, err = time.Parse(time.RFC3339Nano, response["updated_at"].(string))
				assert.NoError(t, err, "updated_at should be a valid RFC3339Nano timestamp")
			}

			mockService.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestGetTask(t *testing.T) {
	mockService := new(MockTaskService)
	mockLogger := new(MockLogger)
	handler := NewTaskHandler(mockService, mockLogger)

	tests := []struct {
		name        string
		taskID      string
		setupMocks  func()
		checkStatus int
		checkBody   interface{}
		isPlainText bool
	}{
		{
			name:   "Success",
			taskID: "test_id",
			setupMocks: func() {
				mockService.On("GetUserTask", mock.Anything, "test_user", "test_id").Return(models.Task{
					ID:          "test_id",
					Title:       "Test Task",
					Description: "Test Description",
					Priority:    models.PriorityHigh,
					Status:      "pending",
					UserID:      "test_user",
					DueDate:     time.Now(),
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}, nil)
			},
			checkStatus: http.StatusOK,
			checkBody: gin.H{
				"id":          "test_id",
				"title":       "Test Task",
				"description": "Test Description",
				"priority":    string(models.PriorityHigh),
				"status":      "pending",
				"user_id":     "test_user",
			},
		},
		{
			name:   "Task_Not_Found",
			taskID: "nonexistent_id",
			setupMocks: func() {
				mockService.On("GetUserTask", mock.Anything, "test_user", "nonexistent_id").Return(models.Task{}, service.ErrTaskNotFound)
			},
			checkStatus: http.StatusNotFound,
			checkBody: gin.H{
				"error": "Task not found",
			},
		},
		{
			name:        "Invalid_Task_ID",
			taskID:      "invalid/id",
			setupMocks:  func() {},
			checkStatus: http.StatusNotFound,
			checkBody:   "404 page not found",
			isPlainText: true,
		},
		{
			name:        "Unauthorized",
			taskID:      "test_id",
			setupMocks:  func() {},
			checkStatus: http.StatusUnauthorized,
			checkBody: gin.H{
				"error": "Unauthorized",
			},
		},
		{
			name:   "Internal_Server_Error",
			taskID: "test_task",
			setupMocks: func() {
				mockService.On("GetUserTask", mock.Anything, "test_user", "test_task").Return(models.Task{}, errors.New("database error"))
				mockLogger.On("Error", "Failed to get task: %v", mock.Anything).Return()
			},
			checkStatus: http.StatusInternalServerError,
			checkBody:   gin.H{"error": "Failed to get task"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			router := gin.New()
			if tt.name != "Unauthorized" {
				router.Use(func(c *gin.Context) {
					if userID := c.GetHeader("X-User-ID"); userID != "" {
						c.Set("user_id", userID)
					}
					c.Next()
				})
			}
			router.GET("/tasks/:id", handler.GetTask)

			tt.setupMocks()

			// Создание запроса
			req := httptest.NewRequest(http.MethodGet, "/tasks/"+tt.taskID, nil)
			if tt.name != "Unauthorized" {
				req.Header.Set("X-User-ID", "test_user")
			}

			// Выполняем запрос
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// проверяем статус
			assert.Equal(t, tt.checkStatus, w.Code)

			if tt.isPlainText {
				// Проверяем текстовый ответ
				assert.Equal(t, tt.checkBody, strings.TrimSpace(w.Body.String()))
			} else {
				// Разбор JSON-ответа
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				// Проверка полей ответа
				expectedBody := tt.checkBody.(gin.H)
				for key, expectedValue := range expectedBody {
					if key == "due_date" || key == "created_at" || key == "updated_at" {
						// Пропускаем проверку временных полей, так как они могут не присутствовать в ответе
						expectedTime, err := time.Parse(time.RFC3339Nano, expectedValue.(string))
						assert.NoError(t, err)
						actualTime, err := time.Parse(time.RFC3339Nano, response[key].(string))
						assert.NoError(t, err)
						assert.Equal(t, expectedTime.Unix(), actualTime.Unix(), "Field %s does not match", key)
					} else {
						assert.Equal(t, expectedValue, response[key], "Field %s does not match", key)
					}
				}
			}

			mockService.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestGetTasks(t *testing.T) {
	mockService := new(MockTaskService)
	mockLogger := new(MockLogger)
	handler := NewTaskHandler(mockService, mockLogger)

	dueDate := time.Now().Add(24 * time.Hour)
	tasks := []models.Task{
		{
			ID:          "task1",
			Title:       "Task 1",
			Description: "Description 1",
			Priority:    models.PriorityHigh,
			Status:      "pending",
			UserID:      "test_user",
			DueDate:     dueDate,
		},
		{
			ID:          "task2",
			Title:       "Task 2",
			Description: "Description 2",
			Priority:    models.PriorityMedium,
			Status:      "completed",
			UserID:      "test_user",
			DueDate:     dueDate,
		},
	}

	tests := []struct {
		name         string
		queryParams  map[string]string
		setupMocks   func()
		checkStatus  int
		checkBody    interface{}
		isAuthorized bool
	}{
		{
			name:         "Get_All_Tasks",
			queryParams:  map[string]string{},
			isAuthorized: true,
			setupMocks: func() {
				mockService.On("GetUserTasks", mock.Anything, "test_user", models.TaskFilters{
					UserID: "test_user",
				}).Return(tasks, nil)
			},
			checkStatus: http.StatusOK,
			checkBody:   tasks,
		},
		{
			name: "Get_Tasks_With_Filters",
			queryParams: map[string]string{
				"status":   "pending",
				"priority": "high",
			},
			isAuthorized: true,
			setupMocks: func() {
				mockService.On("GetUserTasks", mock.Anything, "test_user", models.TaskFilters{
					Status:   models.Status("pending"),
					Priority: models.Priority("high"),
					UserID:   "test_user",
				}).Return([]models.Task{tasks[0]}, nil)
			},
			checkStatus: http.StatusOK,
			checkBody:   []models.Task{tasks[0]},
		},
		{
			name: "Get_Tasks_With_Invalid_Due_Date",
			queryParams: map[string]string{
				"due_date": "invalid-date",
			},
			isAuthorized: true,
			setupMocks: func() {
				mockLogger.On("Error", "Invalid due_date format: %v", mock.Anything).Return()
			},
			checkStatus: http.StatusBadRequest,
			checkBody: gin.H{
				"error": "Invalid due_date format",
			},
		},
		{
			name:         "Get_Tasks_Unauthorized",
			queryParams:  map[string]string{},
			isAuthorized: false,
			setupMocks:   func() {},
			checkStatus:  http.StatusUnauthorized,
			checkBody: gin.H{
				"error": "Unauthorized",
			},
		},
		{
			name: "Get_Tasks_Internal_Error",
			queryParams: map[string]string{
				"status": "pending",
			},
			isAuthorized: true,
			setupMocks: func() {
				mockService.On("GetUserTasks", mock.Anything, "test_user", models.TaskFilters{
					Status: models.Status("pending"),
					UserID: "test_user",
				}).Return([]models.Task{}, errors.New("database error"))
				mockLogger.On("Error", "Failed to get tasks: %v", mock.Anything).Return()
			},
			checkStatus: http.StatusInternalServerError,
			checkBody: gin.H{
				"error": "Failed to get tasks",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			if tt.isAuthorized {
				router.Use(func(c *gin.Context) {
					c.Set("user_id", "test_user")
					c.Next()
				})
			}
			router.GET("/tasks", handler.GetTasks)

			tt.setupMocks()

			// Создаем запрос с параметрами
			req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			if tt.isAuthorized {
				req.Header.Set("X-User-ID", "test_user")
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.checkStatus, w.Code)

			var response interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tasks, ok := tt.checkBody.([]models.Task); ok {
				responseTasks := response.([]interface{})
				require.Equal(t, len(tasks), len(responseTasks))
				for i, task := range tasks {
					responseTask := responseTasks[i].(map[string]interface{})
					assert.Equal(t, task.ID, responseTask["id"])
					assert.Equal(t, task.Title, responseTask["title"])
					assert.Equal(t, task.Description, responseTask["description"])
					assert.Equal(t, string(task.Priority), responseTask["priority"])
					assert.Equal(t, string(task.Status), responseTask["status"])
					assert.Equal(t, task.UserID, responseTask["user_id"])
				}
			} else {
				responseMap := response.(map[string]interface{})
				errorResponse := gin.H{}
				for k, v := range responseMap {
					errorResponse[k] = v
				}
				assert.Equal(t, tt.checkBody, errorResponse)
			}

			mockService.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestUpdateTask(t *testing.T) {
	tests := []struct {
		name       string
		taskID     string
		userID     string
		body       interface{}
		setupMock  func(s *MockTaskService, l *MockLogger)
		checkBody  interface{}
		wantStatus int
	}{
		{
			name:   "Success",
			taskID: "test_task",
			userID: "test_user",
			body: gin.H{
				"title":       "Updated Task",
				"description": "Updated Description",
				"status":      "completed",
				"priority":    "high",
			},
			setupMock: func(s *MockTaskService, l *MockLogger) {
				updatedTask := &models.Task{
					ID:          "test_task",
					UserID:      "test_user",
					Title:       "Updated Task",
					Description: "Updated Description",
					Status:      "completed",
					Priority:    models.PriorityHigh,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
					DueDate:     time.Now().Add(24 * time.Hour),
				}
				s.On("UpdateUserTask", mock.Anything, "test_user", mock.MatchedBy(func(task models.Task) bool {
					return task.ID == "test_task" &&
						task.Title == "Updated Task" &&
						task.Description == "Updated Description" &&
						task.Status == "completed" &&
						task.Priority == models.PriorityHigh
				})).Return(*updatedTask, nil)
			},
			checkBody: gin.H{
				"id":          "test_task",
				"user_id":     "test_user",
				"title":       "Updated Task",
				"description": "Updated Description",
				"status":      "completed",
				"priority":    "high",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "Task_Not_Found",
			taskID: "nonexistent_task",
			userID: "test_user",
			body: gin.H{
				"title":    "Updated Task",
				"priority": "high",
			},
			setupMock: func(s *MockTaskService, l *MockLogger) {
				s.On("UpdateUserTask", mock.Anything, "test_user", mock.MatchedBy(func(task models.Task) bool {
					return task.ID == "nonexistent_task"
				})).Return(models.Task{}, service.ErrTaskNotFound)
			},
			checkBody: gin.H{
				"error": "Task not found",
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "Access_Denied",
			taskID: "other_user_task",
			userID: "test_user",
			body: gin.H{
				"title":    "Updated Task",
				"priority": "high",
			},
			setupMock: func(s *MockTaskService, l *MockLogger) {
				s.On("UpdateUserTask", mock.Anything, "test_user", mock.MatchedBy(func(task models.Task) bool {
					return task.ID == "other_user_task"
				})).Return(models.Task{}, service.ErrAccessDenied)
			},
			checkBody: gin.H{
				"error": "Access denied",
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:   "Invalid_Request_Body",
			taskID: "test_task",
			userID: "test_user",
			body: map[string]interface{}{
				"title": 123, // Invalid type for title
			},
			setupMock: func(s *MockTaskService, l *MockLogger) {
				l.On("Error", "Failed to parse task: %v", mock.Anything).Return()
			},
			checkBody: gin.H{
				"error": "Invalid request body",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "Unauthorized",
			taskID: "test_task",
			userID: "",
			body: gin.H{
				"title": "Updated Task",
			},
			setupMock: func(s *MockTaskService, l *MockLogger) {},
			checkBody: gin.H{
				"error": "Unauthorized",
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:   "Internal_Server_Error",
			taskID: "test_task",
			userID: "test_user",
			body: gin.H{
				"title":    "Updated Task",
				"priority": "high",
			},
			setupMock: func(s *MockTaskService, l *MockLogger) {
				s.On("UpdateUserTask", mock.Anything, "test_user", mock.MatchedBy(func(task models.Task) bool {
					return task.ID == "test_task"
				})).Return(models.Task{}, errors.New("database error"))
				l.On("Error", "Failed to update task: %v", mock.Anything).Return()
			},
			checkBody: gin.H{
				"error": "Failed to update task",
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockTaskService)
			mockLogger := new(MockLogger)
			handler := NewTaskHandler(mockService, mockLogger)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(func(c *gin.Context) {
				if tt.userID != "" {
					c.Set("user_id", tt.userID)
				}
				c.Next()
			})
			router.PUT("/tasks/:id", handler.UpdateTask)

			tt.setupMock(mockService, mockLogger)

			var body bytes.Buffer
			if err := json.NewEncoder(&body).Encode(tt.body); err != nil {
				t.Fatalf("Failed to encode request body: %v", err)
			}

			req := httptest.NewRequest(http.MethodPut, "/tasks/"+tt.taskID, &body)
			req.Header.Set("Content-Type", "application/json")
			if tt.userID != "" {
				req.Header.Set("X-User-ID", tt.userID)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var got gin.H
			err := json.Unmarshal(w.Body.Bytes(), &got)
			require.NoError(t, err)

			if tt.checkBody != nil {
				if tt.name == "Success" {
					expected := tt.checkBody.(gin.H)
					for k, v := range expected {
						require.Equal(t, v, got[k], "field %s does not match", k)
					}
					for _, field := range []string{"created_at", "updated_at", "due_date"} {
						timeStr, ok := got[field].(string)
						require.True(t, ok, "field %s should be a string", field)
						_, err := time.Parse(time.RFC3339Nano, timeStr)
						require.NoError(t, err, "field %s should be a valid RFC3339Nano timestamp", field)
					}
				} else {
					require.Equal(t, tt.checkBody, got)
				}
			}

			mockService.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestDeleteTask(t *testing.T) {
	tests := []struct {
		name       string
		taskID     string
		userID     string
		setupMock  func(s *MockTaskService, l *MockLogger)
		checkBody  gin.H
		wantStatus int
	}{
		{
			name:   "Success",
			taskID: "test_task",
			userID: "test_user",
			setupMock: func(s *MockTaskService, l *MockLogger) {
				s.On("DeleteUserTask", mock.Anything, "test_user", "test_task").Return(nil)
			},
			checkBody: gin.H{
				"message": "Task deleted successfully",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "Task_Not_Found",
			taskID: "nonexistent_task",
			userID: "test_user",
			setupMock: func(s *MockTaskService, l *MockLogger) {
				s.On("DeleteUserTask", mock.Anything, "test_user", "nonexistent_task").Return(service.ErrTaskNotFound)
			},
			checkBody: gin.H{
				"error": "Task not found",
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "Access_Denied",
			taskID: "other_user_task",
			userID: "test_user",
			setupMock: func(s *MockTaskService, l *MockLogger) {
				s.On("DeleteUserTask", mock.Anything, "test_user", "other_user_task").Return(service.ErrAccessDenied)
			},
			checkBody: gin.H{
				"error": "Access denied",
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:      "Unauthorized",
			taskID:    "test_task",
			userID:    "",
			setupMock: func(s *MockTaskService, l *MockLogger) {},
			checkBody: gin.H{
				"error": "Unauthorized",
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:   "Internal_Server_Error",
			taskID: "test_task",
			userID: "test_user",
			setupMock: func(s *MockTaskService, l *MockLogger) {
				s.On("DeleteUserTask", mock.Anything, "test_user", "test_task").Return(errors.New("database error"))
				l.On("Error", "Failed to delete task: %v", mock.Anything).Return()
			},
			checkBody: gin.H{
				"error": "Failed to delete task",
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockTaskService)
			mockLogger := new(MockLogger)
			handler := NewTaskHandler(mockService, mockLogger)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(func(c *gin.Context) {
				if tt.userID != "" {
					c.Set("user_id", tt.userID)
				}
				c.Next()
			})
			router.DELETE("/tasks/:id", handler.DeleteTask)

			tt.setupMock(mockService, mockLogger)

			req := httptest.NewRequest(http.MethodDelete, "/tasks/"+tt.taskID, nil)
			if tt.userID != "" {
				req.Header.Set("X-User-ID", tt.userID)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var got gin.H
			err := json.Unmarshal(w.Body.Bytes(), &got)
			require.NoError(t, err)
			require.Equal(t, tt.checkBody, got)

			mockService.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestGetAnalytics(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		period     string
		setupMock  func(s *MockTaskService, l *MockLogger)
		checkBody  gin.H
		wantStatus int
	}{
		{
			name:   "Success_Week_Period",
			userID: "test_user",
			period: "week",
			setupMock: func(s *MockTaskService, l *MockLogger) {
				s.On("GetUserAnalytics", mock.Anything, "test_user", "week").Return(models.Analytics{
					StatusCount: map[models.Status]int{
						models.StatusPending: 5,
						models.StatusDone:    3,
					},
					PriorityCount: map[models.Priority]int{
						models.PriorityHigh:   2,
						models.PriorityMedium: 4,
						models.PriorityLow:    2,
					},
					AvgCompletionTime:    72.5,
					OnTimeCompletionRate: 0.85,
					OverdueTasks:         1,
					Period:               "week",
					GeneratedAt:          time.Now(),
				}, nil)
			},
			checkBody: gin.H{
				"status_count": map[string]interface{}{
					"pending": float64(5),
					"done":    float64(3),
				},
				"priority_count": map[string]interface{}{
					"high":   float64(2),
					"medium": float64(4),
					"low":    float64(2),
				},
				"avg_completion_time":     float64(72.5),
				"on_time_completion_rate": float64(0.85),
				"overdue_tasks":           float64(1),
				"period":                  "week",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "Invalid_Period",
			userID: "user1",
			period: "invalid",
			setupMock: func(mockService *MockTaskService, mockLogger *MockLogger) {
			},
			checkBody:  gin.H{"error": "Invalid period"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "Unauthorized",
			userID:    "",
			period:    "week",
			setupMock: func(s *MockTaskService, l *MockLogger) {},
			checkBody: gin.H{
				"error": "Unauthorized",
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:   "Internal_Server_Error",
			userID: "test_user",
			period: "week",
			setupMock: func(s *MockTaskService, l *MockLogger) {
				s.On("GetUserAnalytics", mock.Anything, "test_user", "week").Return(models.Analytics{}, errors.New("database error"))
				l.On("Error", "Failed to get analytics: %v", mock.Anything).Return()
			},
			checkBody: gin.H{
				"error": "Failed to get analytics",
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockTaskService)
			mockLogger := new(MockLogger)
			handler := NewTaskHandler(mockService, mockLogger)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(func(c *gin.Context) {
				if tt.userID != "" {
					c.Set("user_id", tt.userID)
				}
				c.Next()
			})
			router.GET("/tasks/analytics", handler.GetAnalytics)

			tt.setupMock(mockService, mockLogger)

			req := httptest.NewRequest(http.MethodGet, "/tasks/analytics?period="+tt.period, nil)
			if tt.userID != "" {
				req.Header.Set("X-User-ID", tt.userID)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var got gin.H
			err := json.Unmarshal(w.Body.Bytes(), &got)
			require.NoError(t, err)

			if tt.name == "Success_Week_Period" {
				assert.Equal(t, tt.checkBody["status_count"], got["status_count"], "status_count mismatch")
				assert.Equal(t, tt.checkBody["priority_count"], got["priority_count"], "priority_count mismatch")
				assert.Equal(t, tt.checkBody["avg_completion_time"], got["avg_completion_time"], "avg_completion_time mismatch")
				assert.Equal(t, tt.checkBody["on_time_completion_rate"], got["on_time_completion_rate"], "on_time_completion_rate mismatch")
				assert.Equal(t, tt.checkBody["overdue_tasks"], got["overdue_tasks"], "overdue_tasks mismatch")
				assert.Equal(t, tt.checkBody["period"], got["period"], "period mismatch")

				generatedAt, ok := got["generated_at"].(string)
				require.True(t, ok, "generated_at should be a string")
				_, err := time.Parse(time.RFC3339Nano, generatedAt)
				require.NoError(t, err, "generated_at should be a valid RFC3339Nano timestamp")
			} else {
				require.Equal(t, tt.checkBody, got)
			}

			mockService.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}
