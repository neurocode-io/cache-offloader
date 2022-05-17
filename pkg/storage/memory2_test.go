package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLRU_size0(t *testing.T) {
	cache, err := NewLRUMap(0)
	assert.Nil(t, cache)
	assert.NotNil(t, err)
}

func TestLRU_functionality(t *testing.T) {
	cache, err := NewLRUMap(3)
	assert.Nil(t, err)
	assert.NotNil(t, cache)

	assert.Equal(t, 0, cache.Len())

	cache.Set("1", Response{
		Status: 200,
	})
	assert.Equal(t, 1, cache.Len())

	cache.Set("2", Response{
		Status: 200,
	})
	assert.Equal(t, 2, cache.Len())

	cache.Set("3", Response{
		Status: 200,
	})
	assert.Equal(t, 3, cache.Len())

	val := cache.Peek("1")
	assert.Equal(t, val.Status, 200)

	cache.Set("1", Response{
		Status: 300,
	})
	assert.Equal(t, 3, cache.Len())

	val = cache.Peek("1")
	assert.Equal(t, val.Status, 300)

	cache.Set("4", Response{
		Status: 200,
	})
	assert.Equal(t, 3, cache.Len())

	val = cache.Peek("2")
	assert.Nil(t, val)

	cache.Clear()
	assert.Equal(t, 0, cache.Len())
}
