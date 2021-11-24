package storage

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"neurocode.io/cache-offloader/pkg/client"
)

var repo = NewRepository(client.NewRedis().Client, 1*time.Second)

type TestResponseBody struct {
	ID      int
	Message string
}

var TestHeader map[string][]string = map[string][]string{}

func TestRedisRepository(t *testing.T) {
	TestHeader["test"] = []string{"test"}
	responseBody := TestResponseBody{ID: 007, Message: "bar"}
	body, _ := json.Marshal(responseBody)
	err := repo.Store(context.TODO(), "testLookup", &Response{Body: body, Header: TestHeader})
	if err != nil {
		t.Error("Failed to set value")
	}

	lookUpResult, err := repo.LookUp(context.TODO(), "testLookup")
	assert.Nil(t, err)

	outBody := TestResponseBody{}
	json.Unmarshal(lookUpResult.Body, &outBody)

	assert.Equal(t, outBody.ID, 007)
	assert.Equal(t, outBody.Message, "bar")
	assert.Equal(t, lookUpResult.Header["test"][0], "test")

	client.NewRedis().Client.Del(context.TODO(), "testLookup")
	assert.Nil(t, err)

	lookUpResult, err = repo.LookUp(context.TODO(), "testLookup")
	assert.Nil(t, err)
	assert.Nil(t, lookUpResult)
}

func TestLookupResultAndErrorNil(t *testing.T) {
	result, err := repo.LookUp(context.TODO(), "doesNotExist")

	assert.Nil(t, result)
	assert.Equal(t, err, nil)
}

func TestLookupTimeout(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 1*time.Microsecond)
	defer cancel()

	result, err := repo.LookUp(ctx, "doesNotExist")

	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}
