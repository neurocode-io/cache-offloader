package metrics

import (
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	cacheHit  = "cacheHit"
	cacheMiss = "cacheMiss"
)

type PrometheusCollector struct {
	httpMetrics *prometheus.CounterVec
}

func isValidHTTPMethod(maybeMethod string) bool {
	methods := []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "CONNECT", "OPTIONS", "TRACE"}

	for _, m := range methods {
		if m == strings.ToUpper(maybeMethod) {
			return true
		}
	}

	return false
}

func NewPrometheusCollector() *PrometheusCollector {
	httpMetricsCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Number of http requests by statusCode, http method and result",
		}, []string{"statusCode", "method", "result"})

	prometheus.Register(httpMetricsCounter)

	return &PrometheusCollector{httpMetrics: httpMetricsCounter}

}

func (m *PrometheusCollector) CacheHit(method string, statusCode int) {
	status := strconv.Itoa(statusCode)
	if !isValidHTTPMethod(method) {
		method = "NA"
	}

	m.httpMetrics.WithLabelValues(status, method, cacheHit).Inc()
}

func (m *PrometheusCollector) CacheMiss(method string, statusCode int) {
	status := strconv.Itoa(statusCode)
	if !isValidHTTPMethod(method) {
		method = "NA"
	}

	m.httpMetrics.WithLabelValues(status, method, cacheMiss).Inc()
}
