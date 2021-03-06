package storage

import (
	"context"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/neurocode-io/cache-offloader/pkg/model"
	"github.com/stretchr/testify/assert"
)

func generateRandomBytes(b *testing.B) []byte {
	arr := make([]byte, 100)
	_, err := rand.Read(arr)
	if err != nil {
		b.Error(err)
	}

	return arr
}

func BenchmarkLRU(b *testing.B) {
	maxMem := 10000.0
	ctx := context.Background()
	lru := NewLRUCache(maxMem, 200)
	for i := 0; i < b.N; i++ {
		err := lru.Store(ctx, fmt.Sprintf("key%d", i), &model.Response{
			Status: 200,
			Body:   generateRandomBytes(b),
		})

		assert.Nil(b, err)
	}
	for i := 0; i < b.N; i++ {
		_, err := lru.LookUp(ctx, fmt.Sprintf("key%d", i))

		assert.Nil(b, err)
	}
}

func BenchmarkLFU(b *testing.B) {
	maxMem := 10000.0
	ctx := context.Background()
	lfu := NewLFUCache(maxMem, 200)
	for i := 0; i < b.N; i++ {
		err := lfu.Store(ctx, fmt.Sprintf("key%d", i), &model.Response{
			Status: 200,
			Body:   generateRandomBytes(b),
		})

		assert.Nil(b, err)
	}
	for i := 0; i < b.N; i++ {
		_, err := lfu.LookUp(ctx, fmt.Sprintf("key%d", i))

		assert.Nil(b, err)
	}
}
