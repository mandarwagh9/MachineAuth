package services

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	TokensIssued = promauto.NewCounter(prometheus.CounterOpts{
		Name: "machineauth_tokens_issued_total",
		Help: "Total number of tokens issued",
	})

	TokensRefreshed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "machineauth_tokens_refreshed_total",
		Help: "Total number of tokens refreshed",
	})

	TokensRevoked = promauto.NewCounter(prometheus.CounterOpts{
		Name: "machineauth_tokens_revoked_total",
		Help: "Total number of tokens revoked",
	})

	TokenValidationAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "machineauth_token_validation_attempts_total",
			Help: "Total number of token validation attempts",
		},
		[]string{"result"},
	)

	ActiveAgents = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "machineauth_active_agents",
		Help: "Number of active agents",
	})

	TokenIssuanceDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "machineauth_token_issuance_duration_seconds",
		Help:    "Time taken to issue tokens",
		Buckets: prometheus.DefBuckets,
	})

	WebhookDeliveries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "machineauth_webhook_deliveries_total",
			Help: "Total number of webhook deliveries",
		},
		[]string{"event", "status"},
	)

	WebhookDeliveryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "machineauth_webhook_delivery_duration_seconds",
			Help:    "Time taken to deliver webhooks",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"event"},
	)

	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "machineauth_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "machineauth_http_request_duration_seconds",
			Help:    "Time taken to process HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)
