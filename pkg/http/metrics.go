package http

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	StatusCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_to_downstream_urls",
			Help: "Number of downstream requests by status.",
		},
		[]string{"status"},
	)

	StorageCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cached_http_requests",
			Help: "Number of cached http requests by status.",
		},
		[]string{"status"},
	)

	ResponseSourceCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_responses",
			Help: "Number of http responses with status<500 by source.",
		},
		[]string{"source"},
	)
)

func MetricsHandler() http.Handler {
	prometheus.MustRegister(StatusCounter)
	prometheus.MustRegister(StorageCounter)
	prometheus.MustRegister(ResponseSourceCounter)

	return promhttp.Handler()
}
