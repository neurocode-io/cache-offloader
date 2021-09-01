package http

import (
	"context"
	"errors"
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

type repositoryMockImpl struct{}

func (r *repositoryMockImpl) LookUp(ctx context.Context, key string) (*storage.Response, error) {
	return nil, errors.New("timeout")
}

func (r *repositoryMockImpl) Store(ctx context.Context, key string, repo *storage.Response) error {
	return nil
}

func (r *repositoryMockImpl) CheckConnection(ctx context.Context) error {
	return nil
}

func setupRedisStore() storage.Repository {
	r := client.NewRedis()
	redisStore := storage.NewRepository(r.Client, &storage.ExpirationTime{Value: 1 * time.Hour}, &storage.CommandTimeout{Value: 1 * time.Second})

	return redisStore
}

func setupHandler(store storage.Repository) http.HandlerFunc {
	downstreamURL, _ := url.Parse(config.New().ServerConfig.DownstreamHost)
	return http.HandlerFunc(IdempotencyHandler(store, downstreamURL))
}

func TestIdempotency(t *testing.T) {
	redisStore := setupRedisStore()
	handler := setupHandler(redisStore)

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
	redisStore := setupRedisStore()
	handler := setupHandler(redisStore)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/status/500", nil)
	client.NewRedis().Client.Del(req.Context(), "Test5xxResponses")
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
	redisStore := setupRedisStore()
	handler := setupHandler(redisStore)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/status/200", nil)
	req.Header.Set("request-id", "ShouldNotStore")

	handler.ServeHTTP(res, req)
	lookUpResult, err := redisStore.LookUp(req.Context(), "ShouldNotStore")
	assert.Nil(t, lookUpResult)
	assert.Nil(t, err)
}

func TestRepoTimeoutResponses(t *testing.T) {
	handle := setupHandler(&repositoryMockImpl{})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/headers", nil)
	req.Header.Set("request-id", "LookupTimeout")

	handle.ServeHTTP(res, req)
	assert.Equal(t, res.Code, http.StatusBadGateway)
}
