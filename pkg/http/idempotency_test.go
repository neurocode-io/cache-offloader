package http

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"dpd.de/idempotency-offloader/config"
	"dpd.de/idempotency-offloader/pkg/client"
	"dpd.de/idempotency-offloader/pkg/storage"
	"github.com/stretchr/testify/assert"
)

func TestIdempotency(t *testing.T) {
	downstreamURL, err := url.Parse(config.New().ServerConfig.DownstreamHost)
	assert.Nil(t, err)

	r := client.NewRedis()
	redisStore := storage.NewRepository(r.Client, 1*time.Hour)

	handler := http.HandlerFunc(IdempotencyHandler(redisStore, downstreamURL))
	res := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/headers", nil)
	req.Header.Set("request-id", "TestIdempotency")
	req.Header.Set("x-b3-traceid", "test1234")
	handler.ServeHTTP(res, req)

	assert.Equal(t, res.Code, http.StatusOK)
	assert.Equal(t, res.Header()["Content-Type"], []string{"application/json"})
	assert.Equal(t, res.Header()["Access-Control-Allow-Credentials"], []string{"true"})
	assert.Equal(t, res.Header()["Server"], []string{"gunicorn/19.9.0"})
	assert.NotNil(t, res.Header()["Date"])
	assert.NotNil(t, res.Header()["Content-Length"])
	assert.NotNil(t, res.Body)
	assert.Greater(t, res.Body.Len(), 0)

	newRes := httptest.NewRecorder()
	handler.ServeHTTP(newRes, req)

	assert.Equal(t, newRes.Code, http.StatusOK)
	assert.Equal(t, newRes.Body, res.Body)
	assert.Equal(t, newRes.Header(), res.Header())

	redisStore.Delete(req.Context(), "TestIdempotency")
}
