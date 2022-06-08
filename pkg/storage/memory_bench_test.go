package storage

import (
	"context"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"neurocode.io/cache-offloader/pkg/model"
)

func generateRandomBytes(b *testing.B) []byte {
	arr := make([]byte, 100)
	_, err := rand.Read(arr)
	if err != nil {
		b.Error(err)
	}

	return arr
}

func BenchmarkStoreLRU(b *testing.B) {
	maxMem := 10000.0
	ctx := context.Background()
	lru := NewHashLRU(maxMem)
	for i := 0; i < b.N; i++ {
		err := lru.Store(ctx, fmt.Sprintf("key%d", i), &model.Response{
			Status: 200,
			Body:   generateRandomBytes(b),
		})

		assert.Nil(b, err)
	}
}

func BenchmarkStoreLRU2(b *testing.B) {
	maxMem := 10000.0
	lru := NewLRUCache(maxMem)
	for i := 0; i < b.N; i++ {
		lru.Store(fmt.Sprintf("key%d", i), model.Response{
			Status: 200,
			Body:   generateRandomBytes(b),
		})
	}
}
