package storage

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"net/http"

	"dpd.de/indempotency-offloader/config"
	"github.com/go-redis/redis/v8"
)

var (
	response http.Response = http.Response{
		Status:     "HelloTest",
		StatusCode: 200,
	}

	ctx        context.Context = context.Background()
	redisCache Repository      = NewRedisCache(5)
	client     *redis.Client   = redis.NewClient(&redis.Options{
		Addr:     strings.Join([]string{config.GetEnv("REDIS_HOST", "127.0.0.1"), config.GetEnv("REDIS_PORT", "6379")}, ":"),
		Password: config.GetEnv("REDIS_PASSWORD", "development"),
		DB:       0,
	})
)

func TestStorage(t *testing.T) {
	redisCache := NewRedisCache(5)
	if redisCache == nil {
		t.Error("redis cache could not be created")
	}
}

func TestLookup(t *testing.T) {
	jsonVal, err := json.Marshal(response)

	if err != nil {
		t.Error("Failed to json marshal")
	}

	result, err := client.Set(ctx, "testLookup", jsonVal, 5*time.Second).Result()
	if err != nil {
		t.Error("Failed to set value")
	}

	if result != "OK" {
		t.Error("Failed to set value")
	}

	lookUpResult, err := redisCache.LookUp("testLookup")

	if lookUpResult.Status != "HelloTest" {
		t.Error("Test failed")
	}

}

func TestStore(t *testing.T) {
	err := redisCache.Store("testStore", &response)
	if err != nil {
		t.Error("Failed to set value")
	}

	getResult, err := client.Get(ctx, "testStore").Result()
	t.Log(getResult)

	if err != nil {
		t.Error("Failed to get the stored result")
	}

	response := http.Response{}
	err = json.Unmarshal([]byte(getResult), &response)

	if response.Status != "HelloTest" {
		t.Error("Test failed")
	}
}
