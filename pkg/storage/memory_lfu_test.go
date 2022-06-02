package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"neurocode.io/cache-offloader/pkg/model"
)

func TestLFU_size0(t *testing.T) {
	cache := NewLFUCache(0)
	assert.NotNil(t, cache)
	assert.Equal(t, 50, cache.Capacity())
}

func TestLFU_functionality(t *testing.T) {
	cache := NewLFUCache(3)
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

func TestLFU_functionality2(t *testing.T) {
	cache := NewLFUCache(2)

	cache.Set("1", model.Response{
		Status: 1,
	})

	cache.Set("2", model.Response{
		Status: 2,
	})

	out := cache.Get("1")
	assert.Equal(t, out.Status, 1)

	cache.Set("3", model.Response{
		Status: 3,
	})

	out = cache.Get("2")
	assert.Nil(t, out)

	out = cache.Get("3")
	assert.Equal(t, out.Status, 3)

	cache.Set("4", model.Response{
		Status: 4,
	})

	out = cache.Get("1")
	assert.Nil(t, out)

	out = cache.Get("3")
	assert.Equal(t, out.Status, 3)
	out = cache.Get("4")
	assert.Equal(t, out.Status, 4)
}
