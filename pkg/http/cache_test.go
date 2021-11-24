package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"neurocode.io/cache-offloader/config"
	"neurocode.io/cache-offloader/pkg/client"
	"neurocode.io/cache-offloader/pkg/storage"
)

type repositoryMockImpl struct{}

func setup(t *testing.T) func() {
	redisStore := setupRedisStore()
	handler := setupHandler(redisStore)
	srv := httptest.NewServer(handler)
	os.Setenv(t.Name(), srv.URL)

	return func() {
		os.Unsetenv(t.Name())
		client.NewRedis().FlushAll(context.TODO())
		srv.Close()
	}
}

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
	redisStore := storage.NewRepository(r.Client, 1*time.Second)

	return redisStore
}

func setupHandler(store storage.Repository) http.HandlerFunc {
	downstreamURL, _ := url.Parse(config.New().ServerConfig.DownstreamHost)
	return http.HandlerFunc(CacheHandler(store, downstreamURL))
}

func TestCacheHandler(t *testing.T) {
	redisStore := setupRedisStore()
	handler := setupHandler(redisStore)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/headers?q=1", nil)

	handler.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)

	newRes := httptest.NewRecorder()
	handler.ServeHTTP(newRes, req)

	assert.Equal(t, http.StatusOK, newRes.Code)
	assert.Equal(t, res.Body, newRes.Body)
	assert.Equal(t, newRes.Header(), res.Header())

}

func Test5xxResponses(t *testing.T) {
	redisStore := setupRedisStore()
	handler := setupHandler(redisStore)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/status/500", nil)

	handler.ServeHTTP(res, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)

	newRes := httptest.NewRecorder()
	time.Sleep(1 * time.Second)
	handler.ServeHTTP(newRes, req)
	// different time means it was not cached
	assert.NotEqual(t, newRes.Header()["Date"], res.Header()["Date"])
}

func TestPassThrough(t *testing.T) {
	redisStore := setupRedisStore()
	handler := setupHandler(redisStore)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/status/200", nil)

	handler.ServeHTTP(res, req)
	lookUpResult, err := redisStore.LookUp(req.Context(), "f7 b4 73 2c d5 f2 3b 90 3c 75 50 7f c0 54 ff 6c ef 4e b3 bc")
	assert.Nil(t, lookUpResult)
	assert.Nil(t, err)
}

func TestRepoTimeoutResponses(t *testing.T) {
	handle := setupHandler(&repositoryMockImpl{})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/headers", nil)

	handle.ServeHTTP(res, req)
	assert.Equal(t, http.StatusBadGateway, res.Code)
}

func TestPassthroguhIsNeverCached(t *testing.T) {
	defer setup(t)()
	c := &http.Client{}

	url := fmt.Sprintf("%s/headers", os.Getenv(t.Name()))
	req, _ := http.NewRequest("GET", url, nil)

	resp, err := c.Do(req)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	newUrl := fmt.Sprintf("%s/status/200", os.Getenv(t.Name()))
	newReq, _ := http.NewRequest("GET", newUrl, nil)

	newResp, err := c.Do(newReq)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, newResp.StatusCode)

}
