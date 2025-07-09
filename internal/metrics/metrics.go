package metrics

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// HTTP request metrics
	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status_code"},
	)

	HttpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Authentication specific metrics
	AuthSignupTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_signup_total",
			Help: "Total number of signup attempts",
		},
		[]string{"status"},
	)

	AuthLoginTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_login_total",
			Help: "Total number of login attempts",
		},
		[]string{"status"},
	)

	AuthPasswordResetTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_password_reset_total",
			Help: "Total number of password reset attempts",
		},
		[]string{"status"},
	)

	// Database metrics
	DatabaseOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "database_operations_total",
			Help: "Total number of database operations",
		},
		[]string{"operation", "status"},
	)

	DatabaseOperationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_operation_duration_seconds",
			Help:    "Duration of database operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// Email metrics
	EmailSentTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "email_sent_total",
			Help: "Total number of emails sent",
		},
		[]string{"type", "status"},
	)

	// JWT metrics
	JWTTokenGeneratedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "jwt_token_generated_total",
			Help: "Total number of JWT tokens generated",
		},
		[]string{"token_type"},
	)

	JWTTokenValidatedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "jwt_token_validated_total",
			Help: "Total number of JWT tokens validated",
		},
		[]string{"status"},
	)

	// Active users metric
	ActiveUsers = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users",
			Help: "Number of currently active users",
		},
	)
)

var registered = false

// RegisterMetrics registers all metrics with Prometheus
func RegisterMetrics() {
	if registered {
		return
	}
	
	prometheus.MustRegister(
		HttpRequestsTotal,
		HttpRequestDuration,
		AuthSignupTotal,
		AuthLoginTotal,
		AuthPasswordResetTotal,
		DatabaseOperationsTotal,
		DatabaseOperationDuration,
		EmailSentTotal,
		JWTTokenGeneratedTotal,
		JWTTokenValidatedTotal,
		ActiveUsers,
	)
	
	registered = true
}

// PrometheusMiddleware creates a middleware that collects HTTP metrics
func PrometheusMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Response().StatusCode())
		method := c.Method()
		path := c.Path()

		// Increment request counter
		HttpRequestsTotal.WithLabelValues(method, path, status).Inc()

		// Record request duration
		HttpRequestDuration.WithLabelValues(method, path).Observe(duration)

		return err
	}
}

// RecordAuthSignup records signup metrics
func RecordAuthSignup(success bool) {
	status := "failure"
	if success {
		status = "success"
	}
	AuthSignupTotal.WithLabelValues(status).Inc()
}

// RecordAuthLogin records login metrics
func RecordAuthLogin(success bool) {
	status := "failure"
	if success {
		status = "success"
	}
	AuthLoginTotal.WithLabelValues(status).Inc()
}

// RecordAuthPasswordReset records password reset metrics
func RecordAuthPasswordReset(success bool) {
	status := "failure"
	if success {
		status = "success"
	}
	AuthPasswordResetTotal.WithLabelValues(status).Inc()
}

// RecordDatabaseOperation records database operation metrics
func RecordDatabaseOperation(operation string, success bool, duration time.Duration) {
	status := "failure"
	if success {
		status = "success"
	}
	DatabaseOperationsTotal.WithLabelValues(operation, status).Inc()
	DatabaseOperationDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// RecordEmailSent records email sending metrics
func RecordEmailSent(emailType string, success bool) {
	status := "failure"
	if success {
		status = "success"
	}
	EmailSentTotal.WithLabelValues(emailType, status).Inc()
}

// RecordJWTTokenGenerated records JWT token generation metrics
func RecordJWTTokenGenerated(tokenType string) {
	JWTTokenGeneratedTotal.WithLabelValues(tokenType).Inc()
}

// RecordJWTTokenValidated records JWT token validation metrics
func RecordJWTTokenValidated(success bool) {
	status := "failure"
	if success {
		status = "success"
	}
	JWTTokenValidatedTotal.WithLabelValues(status).Inc()
}

// UpdateActiveUsers updates the active users gauge
func UpdateActiveUsers(count int) {
	ActiveUsers.Set(float64(count))
} 