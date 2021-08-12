package storage

import (
	"context"
	"testing"
	"time"

	"dpd.de/idempotency-offloader/pkg/client"
	"dpd.de/idempotency-offloader/pkg/entity"
	"github.com/stretchr/testify/assert"
)

var repo = NewRepository(client.NewRedis().Client)

func TestRedisRepository(t *testing.T) {
	response := entity.ResponseBody{ID: 007, Message: "bar"}
	err := repo.Store(context.TODO(), "testLookup", &response)
	if err != nil {
		t.Error("Failed to set value")
	}

	lookUpResult, err := repo.LookUp(context.TODO(), "testLookup")

	assert.Nil(t, err)
	assert.Equal(t, lookUpResult.ID, 007)
	assert.Equal(t, lookUpResult.Message, "bar")
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

// func TestStoreTimeout(t *testing.T) {
// 	ctx := context.Background()
// 	ctx, _ = context.WithTimeout(ctx, 1*time.Microsecond)
// 	err := repo.Store(ctx, "doesNotExist", []byte("hello world"))

// 	assert.Contains(t, err.Error(), "context deadline exceeded")
// }
