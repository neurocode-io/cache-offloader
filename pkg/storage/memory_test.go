package storage

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"neurocode.io/cache-offloader/config"
	"neurocode.io/cache-offloader/pkg/model"
)

func TestHashLRU(t *testing.T) {
	oneHoundredBytes := 100.0 / 1024 / 1024
	lru := NewHashLRU(oneHoundredBytes, config.CacheConfig{CommandTimeoutMilliseconds: 10})
	ctx := context.Background()

	for i := 0; i < 100; i++ {
		err := lru.Store(ctx, fmt.Sprintf("key%d", i), &model.Response{
			Status: 200,
			Body:   []byte(fmt.Sprintf("body%d", i)),
		})
		assert.Nil(t, err)
	}

	resp, err := lru.LookUp(ctx, "key99")
	assert.Nil(t, err)
	assert.Equal(t, resp.Status, 200)

	resp, err = lru.LookUp(ctx, "key0")
	assert.Nil(t, err)
	assert.Nil(t, resp)
}

func TestHashLRUCommandExeeeded(t *testing.T) {
	oneMegaByte := 1000000.0 / 1024 / 1024
	lru := NewHashLRU(oneMegaByte, config.CacheConfig{CommandTimeoutMilliseconds: 0})
	ctx := context.Background()

	err := lru.Store(ctx, "1", &model.Response{
		Status: 200,
		Body:   []byte("body1"),
	})

	assert.EqualError(t, err, "context deadline exceeded")
}
