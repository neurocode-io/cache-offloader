package http

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func prepareTest(counter *prometheus.CounterVec, testCounterStart int) *httptest.ResponseRecorder {
	handler := MetricsHandler()
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metrics", nil)

	counter.WithLabelValues("Test" + strconv.Itoa(testCounterStart)).Add(4)
	counter.WithLabelValues("Test" + strconv.Itoa(testCounterStart+1)).Add(2)
	counter.WithLabelValues("Test" + strconv.Itoa(testCounterStart+2)).Add(3)

	handler.ServeHTTP(res, req)

	return res
}

func TestStatusCounter(t *testing.T) {
	res := prepareTest(StatusCounter, 1)

	assert.True(t, strings.Contains(string(res.Body.Bytes()), "# HELP http_request_to_downstream_urls Number of downstream requests by status"))
	assert.True(t, strings.Contains(string(res.Body.Bytes()), "# TYPE http_request_to_downstream_urls counter"))
	assert.True(t, strings.Contains(string(res.Body.Bytes()), "http_request_to_downstream_urls{status=\"Test1\"} 4"))
	assert.True(t, strings.Contains(string(res.Body.Bytes()), "http_request_to_downstream_urls{status=\"Test2\"} 2"))
	assert.True(t, strings.Contains(string(res.Body.Bytes()), "http_request_to_downstream_urls{status=\"Test3\"} 3"))
}

func TestStorageCounter(t *testing.T) {
	res := prepareTest(StorageCounter, 4)

	assert.True(t, strings.Contains(string(res.Body.Bytes()), "# HELP cached_http_requests Number of cached http requests by status."))
	assert.True(t, strings.Contains(string(res.Body.Bytes()), "# TYPE cached_http_requests counter"))
	assert.True(t, strings.Contains(string(res.Body.Bytes()), "cached_http_requests{status=\"Test4\"} 4"))
	assert.True(t, strings.Contains(string(res.Body.Bytes()), "cached_http_requests{status=\"Test5\"} 2"))
	assert.True(t, strings.Contains(string(res.Body.Bytes()), "cached_http_requests{status=\"Test6\"} 3"))
}

func TestResponseSourceCounter(t *testing.T) {
	res := prepareTest(ResponseSourceCounter, 7)

	assert.True(t, strings.Contains(string(res.Body.Bytes()), "# HELP http_responses Number of http responses with status<500 by source."))
	assert.True(t, strings.Contains(string(res.Body.Bytes()), "# TYPE http_responses counter"))
	assert.True(t, strings.Contains(string(res.Body.Bytes()), "http_responses{source=\"Test7\"} 4"))
	assert.True(t, strings.Contains(string(res.Body.Bytes()), "http_responses{source=\"Test8\"} 2"))
	assert.True(t, strings.Contains(string(res.Body.Bytes()), "http_responses{source=\"Test9\"} 3"))
}
