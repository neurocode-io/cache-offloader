package storage

import (
	"context"
	"sync"
	"time"

	"neurocode.io/cache-offloader/pkg/model"
)

type HashLRU struct {
	maxSize            float64
	size               float64
	oldCache, newCache map[string]model.Response
	lock               sync.RWMutex
	commandTimeout     time.Duration
}

func NewHashLRU(maxSizeMB float64) *HashLRU {
	numOfCaches := 2.0
	if maxSizeMB <= 0 {
		maxSizeMB = 50.0
	}

	return &HashLRU{
		maxSize:        maxSizeMB / numOfCaches,
		size:           0,
		oldCache:       make(map[string]model.Response),
		newCache:       make(map[string]model.Response),
		commandTimeout: commandTimeout,
	}
}

func (lru *HashLRU) update(key string, value model.Response) {
	lru.newCache[key] = value
	// number of bytes in a byte slice use the len function
	bodyInBytes := len(value.Body)
	bodyInMB := float64(bodyInBytes) / (1024 * 1024)
	lru.size += bodyInMB

	if lru.size >= lru.maxSize {
		lru.size = 0

		lru.oldCache = make(map[string]model.Response)
		for key, value := range lru.newCache {
			lru.oldCache[key] = value
		}

		lru.newCache = make(map[string]model.Response)
	}
}

func (lru *HashLRU) Store(ctx context.Context, key string, value *model.Response) error {
	lru.lock.Lock()
	if _, found := lru.newCache[key]; found {
		lru.newCache[key] = *value
	} else {
		lru.update(key, *value)
	}
	lru.lock.Unlock()

	return nil
}

func (lru *HashLRU) LookUp(ctx context.Context, key string) (*model.Response, error) {
	ctx, cancel := context.WithTimeout(ctx, lru.commandTimeout)
	defer cancel()

	proc := make(chan *model.Response, 1)
	go func() {
		lru.lock.RLock()

		if value, found := lru.newCache[key]; found {
			lru.lock.RUnlock()
			proc <- &value

			return
		}

		if value, found := lru.oldCache[key]; found {
			delete(lru.oldCache, key)
			lru.update(key, value)
			lru.lock.RUnlock()
			proc <- &value

			return
		}

		lru.lock.RUnlock()
		proc <- nil
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case value := <-proc:
		if value == nil {
			return nil, nil
		}

		return value, nil
	}
}
