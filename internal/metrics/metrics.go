package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var Registry = prometheus.NewRegistry()

var (
	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "taskmanager",
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HttpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "taskmanager",
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "endpoint"},
	)

	TasksCreatedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "taskmanager",
			Name:      "tasks_created_total",
			Help:      "Total number of created tasks",
		},
	)

	TasksCompletedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "taskmanager",
			Name:      "tasks_completed_total",
			Help:      "Total number of completed tasks",
		},
	)

	TasksByStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "taskmanager",
			Name:      "tasks_by_status",
			Help:      "Number of tasks by status",
		},
		[]string{"status"},
	)
)

func init() {
	Registry.MustRegister(HttpRequestsTotal)
	Registry.MustRegister(HttpRequestDuration)
	Registry.MustRegister(TasksCreatedTotal)
	Registry.MustRegister(TasksCompletedTotal)
	Registry.MustRegister(TasksByStatus)

	Registry.MustRegister(prometheus.NewBuildInfoCollector())
	Registry.MustRegister(prometheus.NewGoCollector())
	Registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
}
