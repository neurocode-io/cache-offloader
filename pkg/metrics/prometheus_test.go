package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestPrometheusMetrics(t *testing.T) {
	t.Run("should return a prometheus registry", func(t *testing.T) {
		collector := NewPrometheusCollector()
		assert.NotNil(t, collector)
	})

	t.Run("should use NA for invalid HTTP method", func(t *testing.T) {
		collector := NewPrometheusCollector()
		collector.CacheHit("FOO", 200)

		metric, err := collector.httpMetrics.GetMetricWith(prometheus.Labels{"statusCode": "200", "method": "FOO", "result": "cacheHit"})

		assert.Nil(t, err)
		assert.NotNil(t, metric)
	})
}
