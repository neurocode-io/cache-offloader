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
	redisStore := storage.NewRepository(r.Client, 1*time.Hour, 1*time.Second)

	handler := http.HandlerFunc(IdempotencyHandler(redisStore, downstreamURL))
	res := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/headers", nil)

	client.NewRedis().Client.Del(req.Context(), "TestIdempotency")

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

	client.NewRedis().Client.Del(req.Context(), "TestIdempotency")

}

func Test5xxResponses(t *testing.T) {
	downstreamURL, err := url.Parse(config.New().ServerConfig.DownstreamHost)
	assert.Nil(t, err)

	r := client.NewRedis()
	redisStore := storage.NewRepository(r.Client, 1*time.Hour, 1*time.Second)

	handler := http.HandlerFunc(IdempotencyHandler(redisStore, downstreamURL))
	res := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/status/500", nil)
	req.Header.Set("request-id", "Test5xxResponses")

	client.NewRedis().Client.Del(req.Context(), "Test5xxResponses")

	handler.ServeHTTP(res, req)

	assert.Equal(t, res.Code, 500)
	assert.Equal(t, res.Header()["Access-Control-Allow-Credentials"], []string{"true"})
	assert.NotNil(t, res.Header()["Date"])
	assert.Equal(t, res.Header()["Content-Length"], []string{"0"})

	lookUpResult, err := redisStore.LookUp(req.Context(), "Test5xxResponses")
	assert.Nil(t, lookUpResult)

	newRes := httptest.NewRecorder()
	time.Sleep(1 * time.Second)
	handler.ServeHTTP(newRes, req)
	// different time means it was not cached
	assert.NotEqual(t, newRes.Header()["Date"], res.Header()["Date"])
}
