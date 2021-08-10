package storage

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"net/http"

	"dpd.de/indempotency-offloader/pkg/client"
	"github.com/stretchr/testify/assert"
)

var repo = NewRepository(client.NewRedis().Client)

func TestRedisRepository(t *testing.T) {
	jsonVal, err := json.Marshal(http.Response{Status: "HelloTest", StatusCode: 200})
	if err != nil {
		t.Error("Failed to json marshal")
	}

	err = repo.Store(context.TODO(), "testLookup", jsonVal)
	if err != nil {
		t.Error("Failed to set value")
	}

	lookUpResult, err := repo.LookUp(context.TODO(), "testLookup")

	if lookUpResult.Status != "HelloTest" {
		t.Error("Test failed")
	}

}

func TestLookupResultAndErrorNil(t *testing.T) {
	result, err := repo.LookUp(context.TODO(), "doesNotExist")

	assert.Nil(t, result)
	assert.Equal(t, err, nil)
}

func TestLookupTimeout(t *testing.T) {
	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 1*time.Microsecond)
	result, err := repo.LookUp(ctx, "doesNotExist")

	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestStoreTimeout(t *testing.T) {
	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 1*time.Microsecond)
	err := repo.Store(ctx, "doesNotExist", []byte("hello world"))

	assert.Contains(t, err.Error(), "context deadline exceeded")
}
