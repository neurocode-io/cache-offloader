package storage

import (
	"context"
	"testing"
	"time"

	"github.com/neurocode-io/cache-offloader/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestLRUFunctionality(t *testing.T) {
	cache := NewLRUCache(0.00001, 50)
	assert.NotNil(t, cache)

	ctx := context.Background()

	err := cache.Store(ctx, "1", &model.Response{
		Status: 100,
		Body:   []byte{1, 2, 3},
	})

	assert.Nil(t, err)

	resp, err := cache.LookUp(ctx, "1")
	assert.Nil(t, err)
	assert.Equal(t, 100, resp.Status)
	assert.False(t, resp.IsStale())

	err = cache.Store(ctx, "2", &model.Response{
		Status: 200,
		Body:   []byte{1, 2, 3},
	})

	assert.Nil(t, err)

	resp, err = cache.LookUp(ctx, "2")
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.Status)
	assert.False(t, resp.IsStale())

	err = cache.Store(ctx, "3", &model.Response{
		Status: 300,
		Body:   []byte{1, 2, 3},
	})

	assert.Nil(t, err)

	resp, err = cache.LookUp(ctx, "3")
	assert.Nil(t, err)
	assert.Equal(t, 300, resp.Status)
	assert.False(t, resp.IsStale())

	err = cache.Store(ctx, "1", &model.Response{
		Status: 200,
		Body:   []byte{1, 2, 3},
	})

	assert.Nil(t, err)

	err = cache.Store(ctx, "4", &model.Response{
		Status: 400,
		Body:   []byte{1, 2, 3},
	})

	assert.Nil(t, err)

	resp, err = cache.LookUp(ctx, "4")
	assert.Nil(t, err)
	assert.Equal(t, 400, resp.Status)
	assert.False(t, resp.IsStale())

	resp, err = cache.LookUp(ctx, "2")
	assert.Nil(t, err)
	assert.Nil(t, resp)

	err = cache.Store(ctx, "5", &model.Response{
		Status: 500,
		Body:   []byte{1, 2, 3, 4, 5},
	})

	assert.Nil(t, err)

	resp, err = cache.LookUp(ctx, "1")
	assert.Nil(t, err)
	assert.Nil(t, resp)

	resp, err = cache.LookUp(ctx, "3")
	assert.Nil(t, err)
	assert.Nil(t, resp)

	resp, err = cache.LookUp(ctx, "5")
	assert.Nil(t, err)
	assert.Equal(t, 500, resp.Status)
	assert.False(t, resp.IsStale())
}

func TestLRUFunctionality2(t *testing.T) {
	cache := NewLRUCache(0.00001, 50)
	assert.NotNil(t, cache)

	ctx := context.Background()

	err := cache.Store(ctx, "1", &model.Response{
		Status: 100,
		Body:   []byte{1, 2, 3},
	})

	assert.Nil(t, err)

	resp, err := cache.LookUp(ctx, "1")
	assert.Nil(t, err)
	assert.Equal(t, 100, resp.Status)
	assert.False(t, resp.IsStale())

	err = cache.Store(ctx, "2", &model.Response{
		Status: 200,
		Body:   []byte{1, 2, 3},
	})

	assert.Nil(t, err)

	resp, err = cache.LookUp(ctx, "2")
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.Status)
	assert.False(t, resp.IsStale())

	err = cache.Store(ctx, "3", &model.Response{
		Status: 300,
		Body:   []byte{1, 2, 3},
	})

	assert.Nil(t, err)

	resp, err = cache.LookUp(ctx, "3")
	assert.Nil(t, err)
	assert.Equal(t, 300, resp.Status)
	assert.False(t, resp.IsStale())

	err = cache.Store(ctx, "1", &model.Response{
		Status: 200,
		Body:   []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
	})

	assert.Nil(t, err)

	resp, err = cache.LookUp(ctx, "1")
	assert.Nil(t, err)
	assert.Equal(t, 100, resp.Status)
	assert.False(t, resp.IsStale())

	resp, err = cache.LookUp(ctx, "2")
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.Status)
	assert.False(t, resp.IsStale())

	resp, err = cache.LookUp(ctx, "3")
	assert.Nil(t, err)
	assert.Equal(t, 300, resp.Status)
	assert.False(t, resp.IsStale())

	err = cache.Store(ctx, "1", &model.Response{
		Status: 200,
		Body:   []byte{1, 2, 3, 4, 5, 6, 7, 8, 9},
	})

	assert.Nil(t, err)

	resp, err = cache.LookUp(ctx, "3")
	assert.Nil(t, err)
	assert.Nil(t, resp)

	resp, err = cache.LookUp(ctx, "2")
	assert.Nil(t, err)
	assert.Nil(t, resp)

	resp, err = cache.LookUp(ctx, "1")
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.Status)
	assert.False(t, resp.IsStale())
}

func TestLRUCacheCommandExeeeded(t *testing.T) {
	oneMegaByte := 1000000.0 / 1024 / 1024
	lru := NewLRUCache(oneMegaByte, 50)
	ctx := context.Background()

	lru.lookupTimeout = 0
	resp, err := lru.LookUp(ctx, "1")

	assert.Nil(t, resp)
	assert.EqualError(t, err, "context deadline exceeded")
}

func TestLRUStaleStatus(t *testing.T) {
	oneMegaByte := 1000000.0 / 1024 / 1024
	lfu := NewLRUCache(oneMegaByte, 1)
	ctx := context.Background()

	lfu.Store(ctx, "1", &model.Response{
		Status: 100,
		Body:   []byte{1, 2, 3},
	})

	resp, _ := lfu.LookUp(ctx, "1")
	assert.False(t, resp.IsStale())

	time.Sleep(1 * time.Second)
	resp, _ = lfu.LookUp(ctx, "1")
	assert.True(t, resp.IsStale())
}

func TestLRUStaleStatus2(t *testing.T) {
	oneMegaByte := 1000000.0 / 1024 / 1024
	lfu := NewLRUCache(oneMegaByte, 1)
	ctx := context.Background()

	lfu.Store(ctx, "1", &model.Response{
		Status: 100,
		Body:   []byte{1, 2, 3},
	})

	time.Sleep(1 * time.Second)

	resp, _ := lfu.LookUp(ctx, "1")
	assert.True(t, resp.IsStale())

	lfu.Store(ctx, "1", &model.Response{
		Status: 100,
		Body:   []byte{1, 2, 3},
	})

	resp, _ = lfu.LookUp(ctx, "1")
	assert.False(t, resp.IsStale())
}
