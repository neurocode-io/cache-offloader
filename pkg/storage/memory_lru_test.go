package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"neurocode.io/cache-offloader/pkg/model"
)

func TestLRU_size0(t *testing.T) {
	cache := NewLRUCache(0)
	assert.NotNil(t, cache)
	assert.Equal(t, 50, cache.Capacity())
}

func TestLRU_functionality(t *testing.T) {
	cache := NewLRUCache(3)
	assert.NotNil(t, cache)

	assert.Equal(t, 0, cache.Len())

	cache.Set("1", model.Response{
		Status: 200,
	})
	assert.Equal(t, 1, cache.Len())

	cache.Set("2", model.Response{
		Status: 200,
	})
	assert.Equal(t, 2, cache.Len())

	cache.Set("3", model.Response{
		Status: 200,
	})
	assert.Equal(t, 3, cache.Len())

	val := cache.Peek("1")
	assert.Equal(t, val.Status, 200)

	cache.Set("1", model.Response{
		Status: 300,
	})
	assert.Equal(t, 3, cache.Len())

	val = cache.Peek("1")
	assert.Equal(t, val.Status, 300)

	cache.Set("4", model.Response{
		Status: 200,
	})
	assert.Equal(t, 3, cache.Len())

	val = cache.Peek("2")
	assert.Nil(t, val)

	cache.Clear()
	assert.Equal(t, 0, cache.Len())
}
