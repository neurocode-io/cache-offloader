package storage

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"dpd.de/idempotency-offloader/pkg/client"
	"github.com/stretchr/testify/assert"
)

var repo = NewRepository(client.NewRedis().Client)

type TestResponseBody struct {
	ID      int
	Message string
}

func TestRedisRepository(t *testing.T) {
	responseBody := TestResponseBody{ID: 007, Message: "bar"}
	res, _ := json.Marshal(responseBody)
	err := repo.Store(context.TODO(), "testLookup", res)
	if err != nil {
		t.Error("Failed to set value")
	}

	lookUpResult, err := repo.LookUp(context.TODO(), "testLookup")

	outBody := TestResponseBody{}
	json.Unmarshal(lookUpResult, &outBody)
	assert.Nil(t, err)
	assert.Equal(t, outBody.ID, 007)
	assert.Equal(t, outBody.Message, "bar")
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
