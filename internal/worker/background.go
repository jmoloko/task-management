package worker

import (
	"context"
	"sync"
	"time"

	"github.com/jmoloko/taskmange/internal/domain/models"
	"github.com/jmoloko/taskmange/internal/domain/repository"
	domainService "github.com/jmoloko/taskmange/internal/domain/service"
	"github.com/jmoloko/taskmange/internal/logger"
)

// BackgroundWorker фоновые задачи
type BackgroundWorker struct {
	taskService domainService.TaskService
	cache       repository.AnalyticsCache
	logger      logger.Logger
	stopChan    chan struct{}
	wg          sync.WaitGroup
	stopOnce    sync.Once
}

func NewBackgroundWorker(taskService domainService.TaskService, cache repository.AnalyticsCache, logger logger.Logger) *BackgroundWorker {
	return &BackgroundWorker{
		taskService: taskService,
		cache:       cache,
		logger:      logger,
		stopChan:    make(chan struct{}),
	}
}

// запуск фоновых задач
func (w *BackgroundWorker) Start() {
	w.wg.Add(2)

	// очистка просроченных задач
	go func() {
		defer w.wg.Done()
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := w.cleanupExpiredTasks(); err != nil {
					w.logger.Error("Failed to cleanup expired tasks", map[string]interface{}{
						"error": err.Error(),
					})
				}
			case <-w.stopChan:
				return
			}
		}
	}()

	// генерация аналитики
	go func() {
		defer w.wg.Done()
		ticker := time.NewTicker(6 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := w.generateAnalytics(); err != nil {
					w.logger.Error("Failed to generate analytics", map[string]interface{}{
						"error": err.Error(),
					})
				}
			case <-w.stopChan:
				return
			}
		}
	}()
}

// корректная остановка фоновых задач
func (w *BackgroundWorker) Stop() {
	w.stopOnce.Do(func() {
		close(w.stopChan)
		w.wg.Wait()
	})
}

// удаление просроченных задач
func (w *BackgroundWorker) cleanupExpiredTasks() error {
	ctx := context.Background()
	expiredDate := time.Now().AddDate(0, 0, -7) // Tasks expired for 7 days
	filters := models.TaskFilters{
		DueDate: &expiredDate,
	}

	tasks, err := w.taskService.GetAll(ctx, "", filters)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if err := w.taskService.Delete(ctx, task.ID, task.UserID); err != nil {
			w.logger.Error("Failed to delete expired task", map[string]interface{}{
				"task_id": task.ID,
				"error":   err.Error(),
			})
		}
	}

	return nil
}

// генеририруем и кэширует аналитику для всех пользователей
func (w *BackgroundWorker) generateAnalytics() error {
	ctx := context.Background()

	// Получаем список всех пользователей с активными задачами
	users, err := w.taskService.GetActiveUsers(ctx)
	if err != nil {
		return err
	}

	// Для каждого пользователя обновляем кэш аналитики
	for _, userID := range users {
		// Генерируем аналитику за разные периоды
		periods := []string{"day", "week", "month"}
		for _, period := range periods {
			analytics, err := w.taskService.GetAnalytics(ctx, userID, period)
			if err != nil {
				w.logger.Error("Failed to generate analytics", map[string]interface{}{
					"user_id": userID,
					"period":  period,
					"error":   err.Error(),
				})
				continue
			}

			// Сохраняем в кэш
			if err := w.cache.SetUserAnalytics(ctx, repository.CachedAnalytics{
				UserID:    userID,
				Period:    period,
				Analytics: analytics,
				CachedAt:  time.Now(),
			}); err != nil {
				w.logger.Error("Failed to cache analytics", map[string]interface{}{
					"user_id": userID,
					"period":  period,
					"error":   err.Error(),
				})
			}
		}
	}

	return nil
}
