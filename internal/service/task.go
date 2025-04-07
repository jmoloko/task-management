package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoloko/taskmange/internal/domain/models"
	"github.com/jmoloko/taskmange/internal/domain/repository"
	domainService "github.com/jmoloko/taskmange/internal/domain/service"
	"github.com/jmoloko/taskmange/internal/logger"
	"github.com/jmoloko/taskmange/internal/metrics"
)

var (
	// ErrTaskNotFound возвращается, когда задача не найдена
	ErrTaskNotFound = errors.New("task not found")
	// ErrInvalidTaskData возвращается при некорректных данных задачи
	ErrInvalidTaskData = errors.New("invalid task data")
	// ErrAccessDenied возвращается при попытке доступа к чужой задаче
	ErrAccessDenied = errors.New("access denied")
)

// TaskServiceImpl реализует интерфейс domainService.TaskService
type TaskServiceImpl struct {
	repo   repository.TaskRepository
	cache  repository.AnalyticsCache
	logger logger.Logger
}

// NewTaskService создает новый экземпляр TaskServiceImpl
func NewTaskService(repo repository.TaskRepository, cache repository.AnalyticsCache, logger logger.Logger) domainService.TaskService {
	return &TaskServiceImpl{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

// Create создает новую задачу
func (s *TaskServiceImpl) Create(ctx context.Context, task models.Task) (models.Task, error) {
	s.logger.Info("Creating new task", map[string]interface{}{
		"title":    task.Title,
		"status":   task.Status,
		"priority": task.Priority,
		"due_date": task.DueDate,
	})

	if task.Title == "" {
		s.logger.Error("Invalid task data: title is required")
		return models.Task{}, ErrInvalidTaskData
	}

	if task.Status == "" {
		s.logger.Info("Setting default status: pending")
		task.Status = models.StatusPending
	}

	if task.Priority == "" {
		s.logger.Info("Setting default priority: medium")
		task.Priority = models.PriorityMedium
	}

	if task.DueDate.IsZero() {
		tomorrow := time.Now().AddDate(0, 0, 1)
		s.logger.Info("Setting default due date", map[string]interface{}{
			"due_date": tomorrow,
		})
		task.DueDate = tomorrow
	}

	if err := s.repo.Create(ctx, &task); err != nil {
		s.logger.Error("Failed to create task in repository", map[string]interface{}{
			"error": err.Error(),
		})
		return models.Task{}, err
	}

	metrics.TasksCreatedTotal.Inc()
	metrics.TasksByStatus.WithLabelValues(string(task.Status)).Inc()

	s.logger.Info("Task created successfully", map[string]interface{}{
		"task_id": task.ID,
	})

	return task, nil
}

// GetByID возвращает задачу по ID
func (s *TaskServiceImpl) GetByID(ctx context.Context, id, userID string) (models.Task, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return models.Task{}, ErrTaskNotFound
	}

	if task.UserID != userID {
		return models.Task{}, ErrAccessDenied
	}

	return *task, nil
}

// GetAll возвращает все задачи с применением фильтров
func (s *TaskServiceImpl) GetAll(ctx context.Context, userID string, filters models.TaskFilters) ([]models.Task, error) {
	return s.repo.GetAll(ctx, filters)
}

// Update обновляет существующую задачу
func (s *TaskServiceImpl) Update(ctx context.Context, id, userID string, task models.Task) (models.Task, error) {
	s.logger.Info("Updating task", map[string]interface{}{
		"task_id": id,
		"user_id": userID,
	})

	existingTask, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Task not found", map[string]interface{}{
			"task_id": id,
			"error":   err.Error(),
		})
		return models.Task{}, ErrTaskNotFound
	}

	if existingTask.UserID != userID {
		s.logger.Error("Access denied to task", map[string]interface{}{
			"task_id": id,
			"user_id": userID,
		})
		return models.Task{}, ErrAccessDenied
	}

	if task.Title != "" {
		existingTask.Title = task.Title
	}

	if task.Description != existingTask.Description {
		existingTask.Description = task.Description
	}

	if task.Status != "" {
		existingTask.Status = task.Status

		if task.Status == models.StatusDone && (existingTask.CompletedAt == nil || *existingTask.CompletedAt == time.Time{}) {
			now := time.Now()
			existingTask.CompletedAt = &now
			s.logger.Info("Task marked as completed", map[string]interface{}{
				"task_id":      id,
				"completed_at": now,
			})
		}
	}

	if task.Priority != "" {
		existingTask.Priority = task.Priority
	}

	if !task.DueDate.IsZero() {
		existingTask.DueDate = task.DueDate
	}

	existingTask.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, existingTask); err != nil {
		s.logger.Error("Failed to update task", map[string]interface{}{
			"task_id": id,
			"error":   err.Error(),
		})
		return models.Task{}, err
	}

	if task.Status == models.StatusDone {
		metrics.TasksCompletedTotal.Inc()
	}

	metrics.TasksByStatus.WithLabelValues(string(task.Status)).Inc()

	s.logger.Info("Task updated successfully", map[string]interface{}{
		"task_id": id,
	})

	return *existingTask, nil
}

// Delete удаляет задачу
func (s *TaskServiceImpl) Delete(ctx context.Context, taskID, userID string) error {
	// Проверяем существование задачи и права доступа
	task, err := s.GetByID(ctx, taskID, userID)
	if err != nil {
		return err
	}

	if task.UserID != userID {
		return ErrAccessDenied
	}

	return s.repo.Delete(ctx, taskID)
}

// Import импортирует список задач
func (s *TaskServiceImpl) Import(ctx context.Context, userID string, tasks []models.Task) error {
	for i := range tasks {
		tasks[i].UserID = userID
		tasks[i].ID = uuid.New().String()
		tasks[i].CreatedAt = time.Now()
		tasks[i].UpdatedAt = time.Now()

		if tasks[i].Status == "" {
			tasks[i].Status = models.StatusPending
		}

		if tasks[i].Priority == "" {
			tasks[i].Priority = models.PriorityMedium
		}

		if tasks[i].DueDate.IsZero() {
			tasks[i].DueDate = time.Now().AddDate(0, 0, 1)
		}

		if err := s.repo.Create(ctx, &tasks[i]); err != nil {
			return err
		}
	}

	return nil
}

// Export экспортирует задачи пользователя
func (s *TaskServiceImpl) Export(ctx context.Context, userID string) ([]models.Task, error) {
	return s.repo.GetAll(ctx, models.TaskFilters{UserID: userID})
}

// GetAnalytics возвращает аналитику по задачам (алиас для GetUserAnalytics)
func (s *TaskServiceImpl) GetAnalytics(ctx context.Context, userID string, period string) (models.Analytics, error) {
	return s.GetUserAnalytics(ctx, userID, period)
}

// CreateTask создает новую задачу
func (s *TaskServiceImpl) CreateTask(ctx context.Context, userID string, task models.Task) (models.Task, error) {
	task.UserID = userID
	return s.Create(ctx, task)
}

// GetUserTask возвращает задачу по ID
func (s *TaskServiceImpl) GetUserTask(ctx context.Context, userID, taskID string) (models.Task, error) {
	return s.GetByID(ctx, taskID, userID)
}

// GetUserTasks возвращает задачи по фильтрам
func (s *TaskServiceImpl) GetUserTasks(ctx context.Context, userID string, filters models.TaskFilters) ([]models.Task, error) {
	return s.GetAll(ctx, userID, filters)
}

// UpdateUserTask обновляет существующую задачу
func (s *TaskServiceImpl) UpdateUserTask(ctx context.Context, userID string, task models.Task) (models.Task, error) {
	return s.Update(ctx, task.ID, userID, task)
}

// DeleteUserTask удаляет задачу
func (s *TaskServiceImpl) DeleteUserTask(ctx context.Context, userID, taskID string) error {
	return s.Delete(ctx, taskID, userID)
}

// ImportTasks импортирует список задач
func (s *TaskServiceImpl) ImportTasks(ctx context.Context, userID string, tasks []models.Task) error {
	return s.Import(ctx, userID, tasks)
}

// ExportUserTasks экспортирует задачи пользователя
func (s *TaskServiceImpl) ExportUserTasks(ctx context.Context, userID string) ([]models.Task, error) {
	return s.Export(ctx, userID)
}

// GetUserAnalytics возвращает аналитику по задачам
func (s *TaskServiceImpl) GetUserAnalytics(ctx context.Context, userID string, period string) (models.Analytics, error) {
	// Пытаемся получить данные из кэша
	cachedData, err := s.cache.GetUserAnalytics(ctx, userID, period)
	if err != nil {
		s.logger.Error("Failed to get analytics from cache", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
			"period":  period,
		})
	} else if cachedData != nil {
		s.logger.Info("Analytics retrieved from cache", map[string]interface{}{
			"user_id": userID,
			"period":  period,
		})
		return cachedData.Analytics, nil
	}

	// Если данных в кэше нет или произошла ошибка, вычисляем аналитику
	filters := models.TaskFilters{
		UserID: userID,
	}

	tasks, err := s.repo.GetAll(ctx, filters)
	if err != nil {
		return models.Analytics{}, err
	}

	analytics := models.Analytics{
		StatusCount:   make(map[models.Status]int),
		PriorityCount: make(map[models.Priority]int),
		Period:        period,
		GeneratedAt:   time.Now(),
	}

	var completedTasks, overdueTasks, onTimeTasks int
	var totalCompletionTime float64

	for _, task := range tasks {
		// Подсчет по статусам
		analytics.StatusCount[task.Status]++

		// Подсчет по приоритетам
		analytics.PriorityCount[task.Priority]++

		// Анализ выполненных задач
		if task.Status == models.StatusDone && task.CompletedAt != nil {
			completedTasks++
			completionTime := task.CompletedAt.Sub(task.CreatedAt).Hours()
			totalCompletionTime += completionTime

			if task.CompletedAt.Before(task.DueDate) {
				onTimeTasks++
			}
		}

		// Подсчет просроченных задач
		if task.Status != models.StatusDone && time.Now().After(task.DueDate) {
			overdueTasks++
		}
	}

	// Вычисление среднего времени выполнения
	if completedTasks > 0 {
		analytics.AvgCompletionTime = totalCompletionTime / float64(completedTasks)
	}

	// Вычисление процента выполнения в срок
	if completedTasks > 0 {
		analytics.OnTimeCompletionRate = float64(onTimeTasks) / float64(completedTasks) * 100
	}

	analytics.OverdueTasks = overdueTasks

	// Сохраняем результаты в кэш
	if err := s.cache.SetUserAnalytics(ctx, repository.CachedAnalytics{
		UserID:    userID,
		Period:    period,
		Analytics: analytics,
		CachedAt:  time.Now(),
	}); err != nil {
		s.logger.Error("Failed to cache analytics", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
			"period":  period,
		})
	}

	return analytics, nil
}

// GetActiveUsers возвращает список ID пользователей с активными задачами
func (s *TaskServiceImpl) GetActiveUsers(ctx context.Context) ([]string, error) {
	// Получаем все задачи
	tasks, err := s.repo.GetAll(ctx, models.TaskFilters{})
	if err != nil {
		return nil, err
	}

	// Создаем map для уникальных ID пользователей
	userIDs := make(map[string]struct{})
	for _, task := range tasks {
		userIDs[task.UserID] = struct{}{}
	}

	// Преобразуем map в slice
	result := make([]string, 0, len(userIDs))
	for userID := range userIDs {
		result = append(result, userID)
	}

	return result, nil
}
