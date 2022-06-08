package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"neurocode.io/cache-offloader/pkg/model"
)

func TestLFU_size0(t *testing.T) {
	cache := NewLFUCache(0)
	assert.NotNil(t, cache)
	assert.Equal(t, 50.0, cache.Capacity())
}

func TestLFU_functionality(t *testing.T) {
	cache := NewLFUCache(0.00001)
	assert.NotNil(t, cache)

	ctx := context.Background()

	cache.Store("1", model.Response{
		Status: 100,
		Body:   []byte{1, 2, 3},
	})

	resp, err := cache.LookUp(ctx, "1")
	assert.Nil(t, err)
	assert.Equal(t, 100, resp.Status)

	cache.Store("2", model.Response{
		Status: 200,
		Body:   []byte{1, 2, 3},
	})

	resp, err = cache.LookUp(ctx, "2")
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.Status)

	cache.Store("3", model.Response{
		Status: 300,
		Body:   []byte{1, 2, 3},
	})

	resp, err = cache.LookUp(ctx, "3")
	assert.Nil(t, err)
	assert.Equal(t, 300, resp.Status)

	cache.Store("1", model.Response{
		Status: 200,
		Body:   []byte{1, 2, 3},
	})

	cache.Store("4", model.Response{
		Status: 400,
		Body:   []byte{1, 2, 3},
	})

	resp, err = cache.LookUp(ctx, "4")
	assert.Nil(t, err)
	assert.Equal(t, 400, resp.Status)

	resp, err = cache.LookUp(ctx, "2")
	assert.Nil(t, err)
	assert.Nil(t, resp)

	cache.Store("5", model.Response{
		Status: 500,
		Body:   []byte{1, 2, 3, 4, 5},
	})

	resp, err = cache.LookUp(ctx, "3")
	assert.Nil(t, err)
	assert.Nil(t, resp)

	resp, err = cache.LookUp(ctx, "4")
	assert.Nil(t, err)
	assert.Nil(t, resp)

	resp, err = cache.LookUp(ctx, "5")
	assert.Nil(t, err)
	assert.Equal(t, 500, resp.Status)
}

func TestLFU_functionality2(t *testing.T) {
	cache := NewLFUCache(0.00001)
	assert.NotNil(t, cache)

	ctx := context.Background()

	cache.Store("1", model.Response{
		Status: 100,
		Body:   []byte{1, 2, 3},
	})

	resp, err := cache.LookUp(ctx, "1")
	assert.Nil(t, err)
	assert.Equal(t, 100, resp.Status)

	cache.Store("2", model.Response{
		Status: 200,
		Body:   []byte{1, 2, 3},
	})

	resp, err = cache.LookUp(ctx, "2")
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.Status)

	cache.Store("3", model.Response{
		Status: 300,
		Body:   []byte{1, 2, 3},
	})

	resp, err = cache.LookUp(ctx, "3")
	assert.Nil(t, err)
	assert.Equal(t, 300, resp.Status)

	// body size > capacity, nothing should happen
	cache.Store("1", model.Response{
		Status: 200,
		Body:   []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
	})

	resp, err = cache.LookUp(ctx, "1")
	assert.Nil(t, err)
	assert.Equal(t, 100, resp.Status)

	resp, err = cache.LookUp(ctx, "2")
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.Status)

	resp, err = cache.LookUp(ctx, "3")
	assert.Nil(t, err)
	assert.Equal(t, 300, resp.Status)

	cache.Store("1", model.Response{
		Status: 200,
		Body:   []byte{1, 2, 3, 4, 5, 6, 7, 8, 9},
	})

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

func TestHashLFUCommandExeeeded(t *testing.T) {
	oneMegaByte := 1000000.0 / 1024 / 1024
	lru := NewLFUCache(oneMegaByte)
	ctx := context.Background()

	lru.commandTimeout = 0
	resp, err := lru.LookUp(ctx, "1")

	assert.Nil(t, resp)
	assert.EqualError(t, err, "context deadline exceeded")
}
