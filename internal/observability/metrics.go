package observability

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	taskCreatedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "screenshot_task_created_total",
			Help: "Total number of created screenshot tasks.",
		},
	)

	taskProcessedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "screenshot_task_processed_total",
			Help: "Total number of processed screenshot tasks grouped by result.",
		},
		[]string{"result"},
	)

	taskProcessingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "screenshot_task_processing_duration_seconds",
			Help:    "Task processing duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"result"},
	)

	taskRetryTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "screenshot_task_retry_total",
			Help: "Total number of task retries.",
		},
	)
)

func init() {
	prometheus.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
		taskCreatedTotal,
		taskProcessedTotal,
		taskProcessingDuration,
		taskRetryTotal,
	)
}

func RecordHTTPRequest(method, path string, status int, d time.Duration) {
	httpRequestsTotal.WithLabelValues(method, path, strconv.Itoa(status)).Inc()
	httpRequestDuration.WithLabelValues(method, path).Observe(d.Seconds())
}

func IncTaskCreated() {
	taskCreatedTotal.Inc()
}

func IncTaskRetry() {
	taskRetryTotal.Inc()
}

func ObserveTaskProcessed(result string, d time.Duration) {
	taskProcessedTotal.WithLabelValues(result).Inc()
	taskProcessingDuration.WithLabelValues(result).Observe(d.Seconds())
}
