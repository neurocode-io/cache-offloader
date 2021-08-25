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

	req, _ := http.NewRequest("GET", "/headers?q=1", nil)

	req.Header.Set("request-id", "TestIdempotency")
	handler.ServeHTTP(res, req)

	assert.Equal(t, res.Code, http.StatusOK)

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

	handler.ServeHTTP(res, req)

	assert.Equal(t, res.Code, http.StatusInternalServerError)

	lookUpResult, err := redisStore.LookUp(req.Context(), "Test5xxResponses")
	assert.Nil(t, lookUpResult)
	assert.Nil(t, err)

	newRes := httptest.NewRecorder()
	time.Sleep(1 * time.Second)
	handler.ServeHTTP(newRes, req)
	// different time means it was not cached
	assert.NotEqual(t, newRes.Header()["Date"], res.Header()["Date"])
}

func TestWrongRegexResponses(t *testing.T) {
	downstreamURL, err := url.Parse(config.New().ServerConfig.DownstreamHost)
	assert.Nil(t, err)

	r := client.NewRedis()
	redisStore := storage.NewRepository(r.Client, 1*time.Hour, 1*time.Second)

	handler := http.HandlerFunc(IdempotencyHandler(redisStore, downstreamURL))
	res := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/ShouldNotStore/", nil)
	req.Header.Set("request-id", "ShouldNotStore")

	handler.ServeHTTP(res, req)
	lookUpResult, err := redisStore.LookUp(req.Context(), "ShouldNotStore")
	assert.Nil(t, lookUpResult)
	assert.Nil(t, err)
}
