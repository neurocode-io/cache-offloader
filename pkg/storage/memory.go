package storage

import (
	"errors"
	"math"
	"sync"
)

type HashLRU struct {
	maxSize            int
	size               int
	oldCache, newCache map[string]Response
	lock               sync.RWMutex
}

func NewHashLRU(maxSizeMB int) (*HashLRU, error) {

	if maxSizeMB <= 0 {
		return nil, errors.New("size must be a postive int")
	}

	lru := &HashLRU{
		maxSize:  maxSizeMB,
		size:     0,
		oldCache: make(map[string]Response),
		newCache: make(map[string]Response),
	}

	return lru, nil
}

func (lru *HashLRU) update(key string, value Response) {

	lru.newCache[key] = value
	lru.size = lru.size + len(value.Body)

	if lru.size >= lru.maxSize {
		lru.size = 0

		lru.oldCache = make(map[string]Response)
		for key, value := range lru.newCache {
			lru.oldCache[key] = value
		}

		lru.newCache = make(map[string]Response)
	}

}

func (lru *HashLRU) Set(key string, value Response) {

	lru.lock.Lock()

	if _, found := lru.newCache[key]; found {
		lru.newCache[key] = value
	} else {
		lru.update(key, value)
	}

	lru.lock.Unlock()

}

func (lru *HashLRU) Get(key string) (*Response, bool) {

	lru.lock.Lock()

	if value, found := lru.newCache[key]; found {
		lru.lock.Unlock()
		return &value, found
	}

	if value, found := lru.oldCache[key]; found {
		delete(lru.oldCache, key)
		lru.update(key, value)
		lru.lock.Unlock()
		return &value, found
	}

	lru.lock.Unlock()
	return nil, false

}

// Difference between get and peek is that we dont update after peek
func (lru *HashLRU) Peek(key string) (*Response, bool) {
	lru.lock.RLock()
	if value, found := lru.newCache[key]; found {
		lru.lock.RUnlock()
		return &value, found
	}

	if value, found := lru.oldCache[key]; found {
		lru.lock.RUnlock()
		return &value, found
	}

	lru.lock.RUnlock()
	return nil, false

}

func (lru *HashLRU) Has(key string) bool {

	lru.lock.RLock()

	_, cacheNew := lru.newCache[key]
	_, cacheOld := lru.oldCache[key]

	lru.lock.RUnlock()

	return cacheNew || cacheOld

}

func (lru *HashLRU) Remove(key string) bool {

	lru.lock.Lock()

	if _, found := lru.newCache[key]; found {
		delete(lru.newCache, key)
		lru.size--
		lru.lock.Unlock()
		return true
	}

	if _, found := lru.oldCache[key]; found {
		delete(lru.oldCache, key)
		lru.lock.Unlock()
		return true
	}

	lru.lock.Unlock()

	return false

}

func (lru *HashLRU) Len() int {

	lru.lock.RLock()

	if lru.size == 0 {
		lru.lock.RUnlock()
		return len(lru.oldCache)
	}

	oldCacheSize := 0

	for key, _ := range lru.oldCache {
		if _, found := lru.newCache[key]; !found {
			oldCacheSize++
		}
	}

	lru.lock.RUnlock()
	return int(math.Min(float64(lru.size+oldCacheSize), float64(lru.maxSize)))

}

func (lru *HashLRU) Clear() {

	lru.lock.Lock()

	lru.oldCache = make(map[string]Response)
	lru.newCache = make(map[string]Response)
	lru.size = 0

	lru.lock.Unlock()

}
