package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Технические метрики
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	httpResponseTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_response_time_seconds",
		Help:    "Duration of HTTP requests",
		Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5},
	}, []string{"method", "path"})

	// Бизнесовые метрики
	PickupPointsCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pickup_points_created_total",
		Help: "Total number of created pickup points",
	})

	OrderAcceptancesCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "order_acceptances_created_total",
		Help: "Total number of created order acceptances",
	})

	ProductsAdded = promauto.NewCounter(prometheus.CounterOpts{
		Name: "products_added_total",
		Help: "Total number of added products",
	})
)
