package storage

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"neurocode.io/cache-offloader/config"
	"neurocode.io/cache-offloader/pkg/model"
)

func BenchmarkStoreLRU(b *testing.B) {
	maxMem := 10000.0
	ctx := context.Background()
	lru := NewHashLRU(maxMem, config.CacheConfig{CommandTimeout: 10})
	for i := 0; i < b.N; i++ {
		err := lru.Store(ctx, fmt.Sprintf("key%d", i), &model.Response{
			Status: 200,
			Body:   []byte(fmt.Sprintf("body%d", i)),
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
			Body:   []byte(fmt.Sprintf("body%d", i)),
		})
	}
}
