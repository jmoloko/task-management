package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoloko/taskmange/internal/config"
	"github.com/jmoloko/taskmange/internal/handler"
	"github.com/jmoloko/taskmange/internal/logger"
	"github.com/jmoloko/taskmange/internal/metrics"
	"github.com/jmoloko/taskmange/internal/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	httpServer    *http.Server
	metricsServer *http.Server
}

// NewServer новый экземпляр сервера
func NewServer(cfg *config.Config, handlers *handler.Handler, logger logger.Logger) *Server {
	router := gin.New()

	router.Use(middleware.LoggerMiddleware(logger))
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RecoveryMiddleware(logger))

	// отдельный маршрутизатор для метрик
	metricsRouter := gin.New()
	metricsRouter.GET("/metrics", gin.WrapH(promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{})))

	router.Use(middleware.MetricsMiddleware())

	// документация Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.URL("http://localhost:8080/docs/swagger.json"),
		ginSwagger.DefaultModelsExpandDepth(-1)))

	// статические файлы Swagger
	router.Static("/docs", "./docs")

	// настройка маршрутов
	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Auth.Register)
			auth.POST("/login", handlers.Auth.Login)
		}

		tasks := api.Group("/tasks")
		tasks.Use(middleware.AuthMiddleware(handlers.Auth.GetService()))
		{
			tasks.POST("", handlers.Task.CreateTask)
			tasks.GET("", handlers.Task.GetTasks)
			tasks.GET("/:id", handlers.Task.GetTask)
			tasks.PUT("/:id", handlers.Task.UpdateTask)
			tasks.DELETE("/:id", handlers.Task.DeleteTask)
			tasks.POST("/import", handlers.Task.ImportTasks)
			tasks.GET("/export", handlers.Task.ExportTasks)
			tasks.GET("/analytics", handlers.Task.GetAnalytics)
		}
	}

	return &Server{
		httpServer: &http.Server{
			Addr:           fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
			Handler:        router,
			ReadTimeout:    cfg.Server.ReadTimeout,
			WriteTimeout:   cfg.Server.WriteTimeout,
			MaxHeaderBytes: 1 << 20,
		},
		metricsServer: &http.Server{
			Addr:    ":9090",
			Handler: metricsRouter,
		},
	}
}

// запускаем HTTP-сервер
func (s *Server) Run() error {
	// Start metrics server
	go func() {
		if err := s.metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Error starting metrics server: %v", err)
		}
	}()

	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully останавливаем сервер
func (s *Server) Shutdown(ctx context.Context) error {
	// останавливаем метрики сервера
	if err := s.metricsServer.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down metrics server: %v", err)
	}

	return s.httpServer.Shutdown(ctx)
}
