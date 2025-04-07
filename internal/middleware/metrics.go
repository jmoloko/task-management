package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoloko/taskmange/internal/metrics"
)

// MetricsMiddleware собираем метрики HTTP для Prometheus
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		metrics.HttpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			status,
		).Inc()

		metrics.HttpRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(duration)
	}
}
