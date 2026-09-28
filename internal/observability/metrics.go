package observability

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request latency in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	OrdersCreated = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "orders_created_total",
		Help: "Total orders created",
	}, []string{"channel", "status"})

	OrdersPaid = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "orders_paid_total",
		Help: "Total orders paid",
	}, []string{"channel"})

	RedemptionsCreated = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "redemptions_created_total",
		Help: "Total crypto redemptions created",
	}, []string{"asset"})

	RedemptionsConfirmed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "redemptions_confirmed_total",
		Help: "Total crypto redemptions confirmed on-chain",
	}, []string{"asset"})

	SettlementsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "settlements_processed_total",
		Help: "Total settlement batches processed",
	}, []string{"status"})

	RefundsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "refunds_processed_total",
		Help: "Total refunds processed",
	}, []string{"channel", "status"})
)

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &statusWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(wrapped, r)
		duration := time.Since(start).Seconds()
		HTTPRequestsTotal.WithLabelValues(r.Method, r.URL.Path, fmt.Sprintf("%d", wrapped.status)).Inc()
		HTTPRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (s *statusWriter) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
