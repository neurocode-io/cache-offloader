package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"neurocode.io/cache-offloader/pkg/model"
)

func TestLRU_size0(t *testing.T) {
	cache := NewLRUCache(0)
	assert.NotNil(t, cache)
	assert.Equal(t, 50.0, cache.Capacity())
}

func TestLRU_functionality(t *testing.T) {
	cache := NewLRUCache(0.00001)
	assert.NotNil(t, cache)

	cache.Store("1", model.Response{
		Status: 100,
		Body:   []byte{1, 2, 3},
	})
	assert.Equal(t, 100, cache.LookUp("1").Status)

	cache.Store("2", model.Response{
		Status: 200,
		Body:   []byte{1, 2, 3},
	})
	assert.Equal(t, 200, cache.LookUp("2").Status)

	cache.Store("3", model.Response{
		Status: 300,
		Body:   []byte{1, 2, 3},
	})
	assert.Equal(t, 300, cache.LookUp("3").Status)

	cache.Store("1", model.Response{
		Status: 200,
		Body:   []byte{1, 2, 3},
	})

	cache.Store("4", model.Response{
		Status: 400,
		Body:   []byte{1, 2, 3},
	})

	assert.Equal(t, 400, cache.LookUp("4").Status)
	assert.Nil(t, cache.LookUp("2"))

	cache.Store("5", model.Response{
		Status: 500,
		Body:   []byte{1, 2, 3, 4, 5},
	})

	assert.Nil(t, cache.LookUp("1"))
	assert.Nil(t, cache.LookUp("3"))
	assert.Equal(t, 500, cache.LookUp("5").Status)
}

func TestLRU_functionality2(t *testing.T) {
	cache := NewLRUCache(0.00001)
	assert.NotNil(t, cache)

	cache.Store("1", model.Response{
		Status: 100,
		Body:   []byte{1, 2, 3},
	})
	assert.Equal(t, 100, cache.LookUp("1").Status)

	cache.Store("2", model.Response{
		Status: 200,
		Body:   []byte{1, 2, 3},
	})
	assert.Equal(t, 200, cache.LookUp("2").Status)

	cache.Store("3", model.Response{
		Status: 300,
		Body:   []byte{1, 2, 3},
	})
	assert.Equal(t, 300, cache.LookUp("3").Status)

	cache.Store("1", model.Response{
		Status: 200,
		Body:   []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
	})

	assert.Equal(t, 100, cache.LookUp("1").Status)
	assert.Equal(t, 200, cache.LookUp("2").Status)
	assert.Equal(t, 300, cache.LookUp("3").Status)

	cache.Store("1", model.Response{
		Status: 200,
		Body:   []byte{1, 2, 3, 4, 5, 6, 7, 8, 9},
	})

	assert.Nil(t, cache.LookUp("3"))
	assert.Nil(t, cache.LookUp("2"))
	assert.Equal(t, 200, cache.LookUp("1").Status)
}
