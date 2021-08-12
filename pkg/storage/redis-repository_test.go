package storage

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"dpd.de/idempotency-offloader/pkg/client"
	"dpd.de/idempotency-offloader/pkg/entity"
	"github.com/stretchr/testify/assert"
)

var repo = NewRepository(client.NewRedis().Client)

func TestRedisRepository(t *testing.T) {
	response := entity.ResponseBody{ID: 007, Message: "bar"}
	res, _ := json.Marshal(response)
	err := repo.Store(context.TODO(), "testLookup", res)
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
