package storage

import (
	"context"
	"sync"
	"time"

	"neurocode.io/cache-offloader/config"
	"neurocode.io/cache-offloader/pkg/model"
)

type hashLRU struct {
	cfg                config.CacheConfig
	maxSize            float64
	size               float64
	oldCache, newCache map[string]model.Response
	lock               sync.RWMutex
}

func NewHashLRU(maxSizeMB float64, cfg config.CacheConfig) hashLRU {
	if maxSizeMB <= 0 {
		maxSizeMB = 50.0
	}

	return hashLRU{
		maxSize:  maxSizeMB,
		size:     0,
		oldCache: make(map[string]model.Response),
		newCache: make(map[string]model.Response),
		cfg:      cfg,
	}
}

func (lru hashLRU) update(key string, value model.Response) {

	lru.newCache[key] = value
	// number of bytes in a byte slice use the len function
	bodyInBytes := len(value.Body)
	bodyInMB := float64(bodyInBytes) / (1024 * 1024)
	lru.size = lru.size + bodyInMB

	if lru.size >= lru.maxSize {
		lru.size = 0

		lru.oldCache = make(map[string]model.Response)
		for key, value := range lru.newCache {
			lru.oldCache[key] = value
		}

		lru.newCache = make(map[string]model.Response)
	}

}

func (lru hashLRU) Store(ctx context.Context, key string, value *model.Response) error {
	ctx, cancel := context.WithTimeout(ctx, lru.cfg.CommandTimeoutMilliseconds*time.Millisecond)
	defer cancel()

	proc := make(chan struct{}, 1)
	go func() {
		lru.lock.Lock()
		if _, found := lru.newCache[key]; found {
			lru.newCache[key] = *value
		} else {
			lru.update(key, *value)
		}
		lru.lock.Unlock()
		proc <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-proc:
		return nil
	}

}

func (lru hashLRU) LookUp(ctx context.Context, key string) (*model.Response, error) {
	ctx, cancel := context.WithTimeout(ctx, lru.cfg.CommandTimeoutMilliseconds*time.Millisecond)
	defer cancel()

	proc := make(chan *model.Response, 1)
	go func() {
		lru.lock.Lock()

		if value, found := lru.newCache[key]; found {
			lru.lock.Unlock()
			proc <- &value
			return
		}

		if value, found := lru.oldCache[key]; found {
			delete(lru.oldCache, key)
			lru.update(key, value)
			lru.lock.Unlock()
			proc <- &value
			return
		}

		lru.lock.Unlock()
		proc <- nil
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case value := <-proc:
		if value != nil {
			return value, nil
		} else {
			return nil, nil
		}
	}
}
