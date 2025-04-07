package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoloko/taskmange/internal/domain/models"
	"github.com/stretchr/testify/require"
)

// createTestUser создает тестового пользователя и возвращает токен
func createTestUser(t *testing.T, env *TestEnv) (models.User, string) {
	// Регистрируем пользователя с уникальным email
	email := fmt.Sprintf("test_%s@example.com", uuid.New().String())
	registerReq := RegisterRequest{
		Email:    email,
		Password: "password123",
	}

	resp, err := makeRequest(env, "POST", "/api/auth/register", registerReq, "")
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	// Логинимся для получения токена
	loginReq := LoginRequest{
		Email:    registerReq.Email,
		Password: registerReq.Password,
	}

	resp, err = makeRequest(env, "POST", "/api/auth/login", loginReq, "")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var loginResp LoginResponse
	err = json.NewDecoder(resp.Body).Decode(&loginResp)
	require.NoError(t, err)

	user := models.User{
		Email: registerReq.Email,
	}

	return user, loginResp.Token
}

// createTestTask создает тестовую задачу
func createTestTask(t *testing.T, env *TestEnv, token string) models.Task {
	req := CreateTaskRequest{
		Title:       "Test Task",
		Description: "Test Description",
		Status:      string(models.StatusPending),
		Priority:    string(models.PriorityMedium),
		DueDate:     time.Now().Add(24 * time.Hour),
	}

	resp, err := makeRequest(env, "POST", "/api/tasks", req, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var task models.Task
	err = json.NewDecoder(resp.Body).Decode(&task)
	require.NoError(t, err)

	return task
}

// createTestTasks создает указанное количество тестовых задач
func createTestTasks(t *testing.T, env *TestEnv, token string, count int) []models.Task {
	tasks := make([]models.Task, count)
	for i := 0; i < count; i++ {
		tasks[i] = createTestTask(t, env, token)
	}
	return tasks
}

// verifyTaskInDB проверяет наличие задачи в БД
func verifyTaskInDB(t *testing.T, env *TestEnv, taskID string) models.Task {
	var task models.Task
	err := env.DB.QueryRow(
		"SELECT id, title, description, status, priority, user_id, due_date, created_at, updated_at FROM tasks WHERE id = $1",
		taskID,
	).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.UserID,
		&task.DueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	require.NoError(t, err)
	return task
}

// ptr возвращает указатель на значение
func ptr[T any](v T) *T {
	return &v
}

func clearTestData(env *TestEnv) error {
	_, err := env.DB.Exec("TRUNCATE users, tasks CASCADE")
	return err
}
