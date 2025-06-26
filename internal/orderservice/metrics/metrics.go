package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto" // For convenience
)

var (
	OrdersCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "orders_created_total",
		Help: "Total number of new orders created.",
	})

	OrderCreationDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "order_creation_duration_seconds",
		Help:    "Duration of order creation calls in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"status"})

	OrdersRetrievedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "orders_retrieved_total",
		Help: "Total number of orders retrieved.",
	})

	OrderRetrievalDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "order_retrieval_duration_seconds",
		Help:    "Duration of order retrieval calls in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"status"})
)
