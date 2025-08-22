package observability

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP/gRPC request metrics
	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "requests_total",
			Help: "Total number of requests",
		},
		[]string{"service", "method", "status"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method"},
	)

	// Business metrics
	OrdersCreated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orders_created_total",
			Help: "Total number of orders created",
		},
		[]string{"status"},
	)

	PaymentsProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payments_processed_total",
			Help: "Total number of payments processed",
		},
		[]string{"status"},
	)

	InventoryReservations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "inventory_reservations_total",
			Help: "Total number of inventory reservations",
		},
		[]string{"product_id", "status"},
	)

	CurrentStock = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "current_stock",
			Help: "Current stock levels for products",
		},
		[]string{"product_id", "product_name"},
	)

	// System metrics
	ActiveConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Number of active gRPC connections",
		},
	)
)

// InitMetrics initializes and registers all metrics
func InitMetrics() {
	prometheus.MustRegister(
		RequestsTotal,
		RequestDuration,
		OrdersCreated,
		PaymentsProcessed,
		InventoryReservations,
		CurrentStock,
		ActiveConnections,
	)
}

// MetricsHandler returns an HTTP handler for Prometheus metrics
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

