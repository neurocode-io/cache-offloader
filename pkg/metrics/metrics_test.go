package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStorageCounter(t *testing.T) {
	opts := CounterVecOpts{
		Name:       "cached_http_requests",
		Help:       "Number of cached http requests by status.",
		LabelNames: "status",
	}

	counter := AddCounterVec(&opts)
	handler := MetricsHandler()
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/management/prometheus", nil)

	counter.WithLabelValues("Test1").Add(1)
	counter.WithLabelValues("Test2").Add(2)
	counter.WithLabelValues("Test3").Add(3)

	handler.ServeHTTP(res, req)

	assert.True(t, strings.Contains(string(res.Body.Bytes()), "# HELP cached_http_requests Number of cached http requests by status."))
	assert.True(t, strings.Contains(string(res.Body.Bytes()), "# TYPE cached_http_requests counter"))
	assert.True(t, strings.Contains(string(res.Body.Bytes()), "cached_http_requests{status=\"Test1\"} 1"))
	assert.True(t, strings.Contains(string(res.Body.Bytes()), "cached_http_requests{status=\"Test2\"} 2"))
	assert.True(t, strings.Contains(string(res.Body.Bytes()), "cached_http_requests{status=\"Test3\"} 3"))
}
