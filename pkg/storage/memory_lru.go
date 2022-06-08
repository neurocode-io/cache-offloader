package storage

import (
	"container/list"
	"context"
	"sync"
	"time"

	"neurocode.io/cache-offloader/pkg/model"
)

type LRUCache struct {
	mtx            sync.Mutex
	responses      *list.List
	cache          map[string]*list.Element
	capacityMB     float64
	sizeMB         float64
	commandTimeout time.Duration
}

type Node struct {
	key   string
	value *model.Response
}

func NewLRUCache(maxSizeMB float64) *LRUCache {
	if maxSizeMB <= 0 {
		maxSizeMB = 50.0
	}

	return &LRUCache{
		capacityMB:     maxSizeMB,
		sizeMB:         0.0,
		responses:      list.New(),
		cache:          make(map[string]*list.Element),
		commandTimeout: commandTimeout,
	}
}

func (lru *LRUCache) Store(key string, value model.Response) {
	lru.mtx.Lock()
	defer lru.mtx.Unlock()

	bodySizeMB := lru.getSize(value)

	if bodySizeMB > lru.capacityMB {
		return
	}

	if val, found := lru.cache[key]; found {
		bodySizeMB -= lru.getSize(*val.Value.(*Node).value)
		val.Value.(*Node).value = &value
		lru.responses.MoveToFront(val)
	} else {
		element := lru.responses.PushFront(&Node{value: &value, key: key})
		lru.cache[key] = element
	}

	lru.sizeMB += bodySizeMB
	var ejectedNode *list.Element

	for lru.sizeMB > lru.capacityMB {
		ejectedNode = lru.responses.Back()
		delete(lru.cache, ejectedNode.Value.(*Node).key)
		lru.responses.Remove(ejectedNode)

		lru.sizeMB -= lru.getSize(*ejectedNode.Value.(*Node).value)
	}
}

func (lru *LRUCache) LookUp(ctx context.Context, key string) (*model.Response, error) {
	ctx, cancel := context.WithTimeout(ctx, lru.commandTimeout)
	defer cancel()

	proc := make(chan *model.Response, 1)

	go func() {
		lru.mtx.Lock()
		defer lru.mtx.Unlock()

		if value, found := lru.cache[key]; found {
			lru.responses.MoveToFront(value)
			proc <- value.Value.(*Node).value

			return
		}

		proc <- nil
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case value := <-proc:
		if value != nil {
			return value, nil
		}

		return nil, nil
	}
}

func (lru *LRUCache) Size() float64 {
	lru.mtx.Lock()
	defer lru.mtx.Unlock()

	return lru.sizeMB
}

func (lru *LRUCache) Capacity() float64 {
	lru.mtx.Lock()
	defer lru.mtx.Unlock()

	return lru.capacityMB
}

func (lru *LRUCache) getSize(value model.Response) float64 {
	sizeBytes := len(value.Body)
	sizeMB := float64(sizeBytes) / (1024 * 1024)

	return sizeMB
}
