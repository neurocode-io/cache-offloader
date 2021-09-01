package metrics

import (
	"strconv"
	"sync"

	"dpd.de/idempotency-offloader/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	success      = "Success"
	lookUpError  = "LookUpError"
	storageError = "StorageError"
)

const (
	cacheHit      = "cacheHit"
	downstreamHit = "downstreamHit"
)

type MetricCollector struct {
	repositoryMetrics *prometheus.CounterVec
	httpMetrics       *prometheus.CounterVec
}

var (
	metricCollectorInstance *MetricCollector
	metricsOnce             sync.Once
)

func NewMetricCollector() *MetricCollector {
	metricsOnce.Do(func() {
		repositoryMetricsCounter := prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "repository_operations_total",
				Help: "Number of operations by result",
			}, []string{"result"})

		httpMetricsCounter := prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Number of http requests by statusCode, http method and result",
			}, []string{"statusCode", "method", "result"})

		prometheus.Register(repositoryMetricsCounter)
		prometheus.Register(httpMetricsCounter)

		metricCollectorInstance = &MetricCollector{repositoryMetrics: repositoryMetricsCounter, httpMetrics: httpMetricsCounter}
	})

	return metricCollectorInstance
}

func (m *MetricCollector) StorageError() {
	m.repositoryMetrics.WithLabelValues(storageError).Inc()
}

func (m *MetricCollector) LookUpError() {
	m.repositoryMetrics.WithLabelValues(lookUpError).Inc()
}

func (m *MetricCollector) Success() {
	m.repositoryMetrics.WithLabelValues(success).Inc()
}

func (m *MetricCollector) CacheHit(statusCode int, method string) {
	status := strconv.Itoa(statusCode)
	if !utils.IsValidHTTPMethod(method) {
		method = "NA"
	}

	m.httpMetrics.WithLabelValues(status, method, cacheHit).Inc()
}

func (m *MetricCollector) DownstreamHit(statusCode int, method string) {
	status := strconv.Itoa(statusCode)
	if !utils.IsValidHTTPMethod(method) {
		method = "NA"
	}

	m.httpMetrics.WithLabelValues(status, method, downstreamHit).Inc()
}
