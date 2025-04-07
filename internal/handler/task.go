package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoloko/taskmange/internal/domain/models"
	domainService "github.com/jmoloko/taskmange/internal/domain/service"
	"github.com/jmoloko/taskmange/internal/logger"
	"github.com/jmoloko/taskmange/internal/service"
)

// TaskHandler обрабатывает HTTP-запросы для задач
type TaskHandler struct {
	service domainService.TaskService
	logger  logger.Logger
}

// NewTaskHandler создаёт новый обработчик для задач
func NewTaskHandler(service domainService.TaskService, logger logger.Logger) *TaskHandler {
	return &TaskHandler{
		service: service,
		logger:  logger,
	}
}

// GetTasks получение списка задач
// @Summary Get all tasks
// @Description Get all tasks with optional filtering
// @Tags tasks
// @Accept json
// @Produce json
// @Param status query string false "Filter by status"
// @Param priority query string false "Filter by priority"
// @Param due_date query string false "Filter by due date (RFC3339 format)"
// @Param search query string false "Search in title and description"
// @Security BearerAuth
// @Success 200 {array} models.Task
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /tasks [get]
func (h *TaskHandler) GetTasks(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	filters := models.TaskFilters{
		Status:   models.Status(c.Query("status")),
		Priority: models.Priority(c.Query("priority")),
		UserID:   userID.(string),
		Search:   c.Query("search"),
	}

	if dueDateStr := c.Query("due_date"); dueDateStr != "" {
		dueDate, err := time.Parse(time.RFC3339, dueDateStr)
		if err != nil {
			h.logger.Error("Invalid due_date format: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid due_date format"})
			return
		}
		filters.DueDate = &dueDate
	}

	tasks, err := h.service.GetUserTasks(c.Request.Context(), userID.(string), filters)
	if err != nil {
		h.logger.Error("Failed to get tasks: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// GetTask получение задачи по ID
// @Summary Get a task by ID
// @Description Get a task by its ID
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Security BearerAuth
// @Success 200 {object} models.Task
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /tasks/{id} [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task ID is required"})
		return
	}

	task, err := h.service.GetUserTask(c.Request.Context(), userID.(string), taskID)
	if err != nil {
		if err == service.ErrTaskNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		if err == service.ErrAccessDenied {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		h.logger.Error("Failed to get task: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get task"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// CreateTask создание новой задачи
// @Summary Create a new task
// @Description Create a new task
// @Tags tasks
// @Accept json
// @Produce json
// @Param task body models.Task true "Task object to create"
// @Security BearerAuth
// @Success 201 {object} models.Task
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		h.logger.Error("Failed to parse task: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// инициализация полей задачи
	if task.ID == "" {
		task.ID = uuid.New().String()
	}

	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now

	createdTask, err := h.service.CreateTask(c.Request.Context(), userID.(string), task)
	if err != nil {
		if err == service.ErrInvalidTaskData {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task data"})
			return
		}
		h.logger.Error("Failed to create task: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, createdTask)
}

// UpdateTask обновление задачи
// @Summary Update a task
// @Description Update an existing task
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param task body models.Task true "Task object with updates"
// @Security BearerAuth
// @Success 200 {object} models.Task
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /tasks/{id} [put]
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task ID is required"})
		return
	}

	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		h.logger.Error("Failed to parse task: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	task.ID = taskID
	task.UpdatedAt = time.Now()

	updatedTask, err := h.service.UpdateUserTask(c.Request.Context(), userID.(string), task)
	if err != nil {
		if err == service.ErrTaskNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		if err == service.ErrAccessDenied {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		h.logger.Error("Failed to update task: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	c.JSON(http.StatusOK, updatedTask)
}

// DeleteTask удаление задачи
// @Summary Delete a task
// @Description Delete a task by ID
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Security BearerAuth
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task ID is required"})
		return
	}

	if err := h.service.DeleteUserTask(c.Request.Context(), userID.(string), taskID); err != nil {
		if err == service.ErrTaskNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		if err == service.ErrAccessDenied {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		h.logger.Error("Failed to delete task: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

// ImportTasks импортируем задачи из файла
// @Summary Import tasks
// @Description Import tasks from a JSON file
// @Tags tasks
// @Accept json
// @Produce json
// @Param tasks body []models.Task true "Array of tasks to import"
// @Security BearerAuth
// @Success 201 {object} map[string]string "Tasks imported successfully"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /tasks/import [post]
func (h *TaskHandler) ImportTasks(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var tasks []models.Task
	if err := c.ShouldBindJSON(&tasks); err != nil {
		h.logger.Error("Failed to parse tasks: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.service.ImportTasks(c.Request.Context(), userID.(string), tasks); err != nil {
		h.logger.Error("Failed to import tasks: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to import tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tasks imported successfully"})
}

// ExportTasks экспортируем задачи в файл
// @Summary Export tasks
// @Description Export all user's tasks as JSON
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Task
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /tasks/export [get]
func (h *TaskHandler) ExportTasks(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	tasks, err := h.service.ExportUserTasks(c.Request.Context(), userID.(string))
	if err != nil {
		h.logger.Error("Failed to export tasks: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export tasks"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// GetAnalytics получаем аналитику
// @Summary Get task analytics
// @Description Get analytics for user's tasks
// @Tags analytics
// @Accept json
// @Produce json
// @Param period query string true "Analytics period (day/week/month)"
// @Security BearerAuth
// @Success 200 {object} models.Analytics
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /analytics [get]
func (h *TaskHandler) GetAnalytics(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	period := c.Query("period")
	if period == "" {
		period = "week"
	}

	if !isValidPeriod(period) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period"})
		return
	}

	analytics, err := h.service.GetUserAnalytics(c.Request.Context(), userID.(string), period)
	if err != nil {
		h.logger.Error("Failed to get analytics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get analytics"})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// isValidPeriod проверяем валидность периода
func isValidPeriod(period string) bool {
	validPeriods := map[string]bool{
		"day":   true,
		"week":  true,
		"month": true,
	}
	return validPeriods[period]
}
