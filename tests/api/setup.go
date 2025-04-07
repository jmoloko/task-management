package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoloko/taskmange/internal/cache"
	"github.com/jmoloko/taskmange/internal/config"
	"github.com/jmoloko/taskmange/internal/handler"
	"github.com/jmoloko/taskmange/internal/logger"
	"github.com/jmoloko/taskmange/internal/middleware"
	"github.com/jmoloko/taskmange/internal/repository/postgres"
	"github.com/jmoloko/taskmange/internal/service"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestEnv содержит все зависимости для тестов
type TestEnv struct {
	API       *gin.Engine
	DB        *sql.DB
	Redis     *redis.Client
	PostgresC testcontainers.Container
	RedisC    testcontainers.Container
	Config    *config.Config
	Server    *httptest.Server
}

// setupPostgres создает и настраивает контейнер с PostgreSQL
func setupPostgres(t *testing.T) testcontainers.Container {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections"),
			wait.ForListeningPort("5432/tcp"),
		),
	}

	postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	return postgres
}

// setupRedis создает и настраивает контейнер с Redis
func setupRedis(t *testing.T) testcontainers.Container {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "redis:7",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForLog("Ready to accept connections"),
			wait.ForListeningPort("6379/tcp"),
		),
	}

	redis, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	return redis
}

// connectDB устанавливает соединение с БД
func connectDB(t *testing.T, container testcontainers.Container) *sql.DB {
	dsn := "postgres://postgres:postgres@localhost:5433/taskmanager?sslmode=disable"
	log.Printf("Connecting to PostgreSQL with DSN: %s", dsn)

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)

	// Проверяем соединение
	log.Printf("Pinging PostgreSQL...")
	require.NoError(t, db.Ping())
	log.Printf("PostgreSQL connection established")

	// Применяем миграции
	log.Printf("Applying migrations...")
	require.NoError(t, applyMigrations(t, db))
	log.Printf("Migrations applied successfully")

	return db
}

// applyMigrations применяет миграции к базе данных
func applyMigrations(t *testing.T, db *sql.DB) error {
	// Читаем файл миграции
	migrationPath := filepath.Join("migrations", "000001_init.up.sql")
	log.Printf("Reading migration file: %s", migrationPath)

	migration, err := os.ReadFile(migrationPath)
	if err != nil {
		log.Printf("Failed to read migration file: %v", err)
		return fmt.Errorf("failed to read migration file: %w", err)
	}
	log.Printf("Migration file read successfully")

	// Применяем миграцию
	log.Printf("Executing migration...")
	_, err = db.Exec(string(migration))
	if err != nil {
		log.Printf("Failed to apply migration: %v", err)
		return fmt.Errorf("failed to apply migration: %w", err)
	}
	log.Printf("Migration executed successfully")

	return nil
}

// connectRedis устанавливает соединение с Redis
func connectRedis(t *testing.T, container testcontainers.Container) *redis.Client {
	addr := "localhost:6380"
	log.Printf("Connecting to Redis at: %s", addr)

	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// Проверяем соединение
	log.Printf("Pinging Redis...")
	require.NoError(t, client.Ping(context.Background()).Err())
	log.Printf("Redis connection established")

	return client
}

// setupAPI создает и настраивает экземпляр API
func setupAPI(db *sql.DB, redisClient *redis.Client) *gin.Engine {
	gin.SetMode(gin.TestMode)

	// Инициализируем зависимости
	taskRepo := postgres.NewTaskRepository(db)
	userRepo := postgres.NewUserRepository(db)
	redisCache := cache.NewRedisCache(redisClient)
	log := &logger.MockLogger{} // Используем мок логгер для тестов

	// Создаем сервисы
	taskService := service.NewTaskService(taskRepo, redisCache, log)
	authService := service.NewAuthService(userRepo, log, "your-secret-key")

	// Создаем обработчики
	taskHandler := handler.NewTaskHandler(taskService, log)
	authHandler := handler.NewAuthHandler(authService, log)

	// Создаем и настраиваем роутер
	router := gin.New()

	// Регистрируем маршруты
	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		tasks := api.Group("/tasks")
		tasks.Use(middleware.AuthMiddleware(authService))
		{
			tasks.POST("", taskHandler.CreateTask)
			tasks.GET("", taskHandler.GetTasks)
			tasks.GET("/:id", taskHandler.GetTask)
			tasks.PUT("/:id", taskHandler.UpdateTask)
			tasks.DELETE("/:id", taskHandler.DeleteTask)
			tasks.GET("/analytics", taskHandler.GetAnalytics)
		}
	}

	return router
}

// SetupTestEnv создает тестовое окружение
func SetupTestEnv(t *testing.T) (*TestEnv, func()) {
	log.Printf("Setting up test environment...")

	// Поднимаем контейнеры
	log.Printf("Starting PostgreSQL container...")
	postgresC := setupPostgres(t)
	log.Printf("PostgreSQL container started")

	log.Printf("Starting Redis container...")
	redisC := setupRedis(t)
	log.Printf("Redis container started")

	// Подключаемся к сервисам
	log.Printf("Connecting to services...")
	db := connectDB(t, postgresC)
	redis := connectRedis(t, redisC)
	log.Printf("Services connected")

	// Создаем API
	log.Printf("Setting up API...")
	api := setupAPI(db, redis)
	log.Printf("API setup completed")

	// Создаем тестовый сервер
	log.Printf("Creating test server...")
	server := httptest.NewServer(api)
	log.Printf("Test server created at: %s", server.URL)

	env := &TestEnv{
		API:       api,
		DB:        db,
		Redis:     redis,
		PostgresC: postgresC,
		RedisC:    redisC,
		Server:    server,
	}

	// Функция очистки
	cleanup := func() {
		log.Printf("Cleaning up test environment...")
		server.Close()
		db.Close()
		redis.Close()
		postgresC.Terminate(context.Background())
		redisC.Terminate(context.Background())
		log.Printf("Cleanup completed")
	}

	log.Printf("Test environment setup completed")
	return env, cleanup
}

// makeRequest выполняет HTTP запрос к API
func makeRequest(env *TestEnv, method, path string, body interface{}, token string) (*http.Response, error) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, env.Server.URL+path, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	return client.Do(req)
}
