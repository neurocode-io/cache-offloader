package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"neurocode.io/cache-offloader/pkg/model"
)

func TestLRU_functionality(t *testing.T) {
	cache := NewLRUCache(0.00001)
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

	err = cache.Store(ctx, "2", &model.Response{
		Status: 200,
		Body:   []byte{1, 2, 3},
	})

	assert.Nil(t, err)

	resp, err = cache.LookUp(ctx, "2")
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.Status)

	err = cache.Store(ctx, "3", &model.Response{
		Status: 300,
		Body:   []byte{1, 2, 3},
	})

	assert.Nil(t, err)

	resp, err = cache.LookUp(ctx, "3")
	assert.Nil(t, err)
	assert.Equal(t, 300, resp.Status)

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
}

func TestLRU_functionality2(t *testing.T) {
	cache := NewLRUCache(0.00001)
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

	err = cache.Store(ctx, "2", &model.Response{
		Status: 200,
		Body:   []byte{1, 2, 3},
	})

	assert.Nil(t, err)

	resp, err = cache.LookUp(ctx, "2")
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.Status)

	err = cache.Store(ctx, "3", &model.Response{
		Status: 300,
		Body:   []byte{1, 2, 3},
	})

	assert.Nil(t, err)

	resp, err = cache.LookUp(ctx, "3")
	assert.Nil(t, err)
	assert.Equal(t, 300, resp.Status)

	err = cache.Store(ctx, "1", &model.Response{
		Status: 200,
		Body:   []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
	})

	assert.Nil(t, err)

	resp, err = cache.LookUp(ctx, "1")
	assert.Nil(t, err)
	assert.Equal(t, 100, resp.Status)

	resp, err = cache.LookUp(ctx, "2")
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.Status)

	resp, err = cache.LookUp(ctx, "3")
	assert.Nil(t, err)
	assert.Equal(t, 300, resp.Status)

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
}

func TestLRUCacheCommandExeeeded(t *testing.T) {
	oneMegaByte := 1000000.0 / 1024 / 1024
	lru := NewLRUCache(oneMegaByte)
	ctx := context.Background()

	lru.commandTimeout = 0
	resp, err := lru.LookUp(ctx, "1")

	assert.Nil(t, resp)
	assert.EqualError(t, err, "context deadline exceeded")
}
