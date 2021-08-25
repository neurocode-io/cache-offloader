package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func prepareTest(counter *prometheus.CounterVec, counterBaseName string) *httptest.ResponseRecorder {
	handler := MetricsHandler()
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/management/prometheus", nil)

	counter.WithLabelValues(counterBaseName + "1").Add(1)
	counter.WithLabelValues(counterBaseName + "2").Add(2)
	counter.WithLabelValues(counterBaseName + "3").Add(3)

	handler.ServeHTTP(res, req)

	return res
}

func TestStatusCounter(t *testing.T) {
	res := prepareTest(StatusCounter, "StatusTest")

	assert.True(t, strings.Contains(string(res.Body.Bytes()), "# HELP http_request_to_downstream_urls Number of downstream requests by status"))
	assert.True(t, strings.Contains(string(res.Body.Bytes()), "# TYPE http_request_to_downstream_urls counter"))
	assert.True(t, strings.Contains(string(res.Body.Bytes()), "http_request_to_downstream_urls{status=\"StatusTest1\"} 1"))
	assert.True(t, strings.Contains(string(res.Body.Bytes()), "http_request_to_downstream_urls{status=\"StatusTest2\"} 2"))
	assert.True(t, strings.Contains(string(res.Body.Bytes()), "http_request_to_downstream_urls{status=\"StatusTest3\"} 3"))
}

func TestStorageCounter(t *testing.T) {
	res := prepareTest(StorageCounter, "StorageTest")

	assert.True(t, strings.Contains(string(res.Body.Bytes()), "# HELP cached_http_requests Number of cached http requests by status."))
	assert.True(t, strings.Contains(string(res.Body.Bytes()), "# TYPE cached_http_requests counter"))
	assert.True(t, strings.Contains(string(res.Body.Bytes()), "cached_http_requests{status=\"StorageTest1\"} 1"))
	assert.True(t, strings.Contains(string(res.Body.Bytes()), "cached_http_requests{status=\"StorageTest2\"} 2"))
	assert.True(t, strings.Contains(string(res.Body.Bytes()), "cached_http_requests{status=\"StorageTest3\"} 3"))
}

func TestResponseSourceCounter(t *testing.T) {
	res := prepareTest(ResponseSourceCounter, "ResSourceTest")

	assert.True(t, strings.Contains(string(res.Body.Bytes()), "# HELP http_responses Number of http responses with status<500 by source."))
	assert.True(t, strings.Contains(string(res.Body.Bytes()), "# TYPE http_responses counter"))
	assert.True(t, strings.Contains(string(res.Body.Bytes()), "http_responses{source=\"ResSourceTest1\"} 1"))
	assert.True(t, strings.Contains(string(res.Body.Bytes()), "http_responses{source=\"ResSourceTest2\"} 2"))
	assert.True(t, strings.Contains(string(res.Body.Bytes()), "http_responses{source=\"ResSourceTest3\"} 3"))
}
