package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jmoloko/taskmange/docs"
	"github.com/jmoloko/taskmange/internal/cache"
	"github.com/jmoloko/taskmange/internal/config"
	"github.com/jmoloko/taskmange/internal/handler"
	"github.com/jmoloko/taskmange/internal/logger"
	"github.com/jmoloko/taskmange/internal/repository/postgres"
	"github.com/jmoloko/taskmange/internal/server"
	"github.com/jmoloko/taskmange/internal/service"
	"github.com/jmoloko/taskmange/internal/worker"
	"github.com/redis/go-redis/v9"
)

// @title Task Management API
// @version 1.0
// @description RESTful API for task management with user authentication
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.example.com/support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the access token.

func main() {

	// инициализируем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// инициализируем логгер
	appLogger := logger.NewSLogLogger(cfg.Logger)
	defer appLogger.Close()

	// инициализируем базу данных
	db, err := postgres.NewPostgresDB(cfg.Database)
	if err != nil {
		appLogger.Error("Failed to initialize db", err)
		return
	}
	appLogger.Info("Database connected successfully")

	// инициализируем Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		DB:   cfg.Redis.DB,
	})
	defer redisClient.Close()

	// Проверяем подключение к Redis
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		appLogger.Error("Failed to connect to Redis", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	appLogger.Info("Redis connected successfully")

	// инициализируем кэш Redis
	redisCache := cache.NewRedisCache(redisClient)

	// инициализируем репозитории
	userRepo := postgres.NewUserRepository(db)
	taskRepo := postgres.NewTaskRepository(db)

	// инициализируем сервисы
	authService := service.NewAuthService(userRepo, appLogger, cfg.Auth.SigningKey)
	taskService := service.NewTaskService(taskRepo, redisCache, appLogger)

	// инициализируем background worker
	backgroundWorker := worker.NewBackgroundWorker(taskService, redisCache, appLogger)
	backgroundWorker.Start()
	defer backgroundWorker.Stop()

	// инициализируем handlers
	authHandler := handler.NewAuthHandler(authService, appLogger)
	taskHandler := handler.NewTaskHandler(taskService, appLogger)
	handlers := handler.NewHandler(authHandler, taskHandler)

	// инициализируем метрики
	srv := server.NewServer(cfg, handlers, appLogger)

	// инициализируем контекст сервера
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Прослушивание сигналов системных вызовов для прерывания/завершения процесса
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Сигнал выключения с периодом отсрочки 30 секунд
		shutdownCtx, cancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer cancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				appLogger.Fatal("graceful shutdown timed out.. forcing exit")
			}
		}()

		// trigger graceful shutdown
		err := srv.Shutdown(shutdownCtx)
		if err != nil {
			appLogger.Fatal(err.Error())
		}
		serverStopCtx()
	}()

	// запуск сервера
	appLogger.Info(fmt.Sprintf("Starting server on port %d", cfg.Server.Port))
	err = srv.Run()
	if err != nil && err != http.ErrServerClosed {
		appLogger.Fatal(fmt.Sprintf("Error starting server: %s", err))
	}

	// ждем server context для остановки
	<-serverCtx.Done()
	appLogger.Info("Server stopped")
}
