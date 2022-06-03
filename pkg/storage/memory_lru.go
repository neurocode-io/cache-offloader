package storage

import (
	"container/list"
	"sync"

	"neurocode.io/cache-offloader/pkg/model"
)

type LRUCache struct {
	mtx        sync.Mutex
	responses  *list.List
	cache      map[string]*list.Element
	capacityMB float64
	sizeMB     float64
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
		capacityMB: maxSizeMB,
		sizeMB:     0.0,
		responses:  list.New(),
		cache:      make(map[string]*list.Element),
	}
}

func (lru *LRUCache) Store(key string, value model.Response) {
	lru.mtx.Lock()
	defer lru.mtx.Unlock()

	bodySizeMB := lru.getSize(value)

	val, found := lru.cache[key]

	if found {
		bodySizeMB = bodySizeMB - lru.getSize(*val.Value.(*Node).value)
	}

	lru.sizeMB += bodySizeMB

	if (lru.sizeMB) <= lru.capacityMB {
		if found {
			val.Value.(*Node).value = &value
			lru.responses.MoveToFront(val)
		} else {
			element := lru.responses.PushFront(&Node{value: &value, key: key})
			lru.cache[key] = element
		}

		return
	}

	ejectedNode := &list.Element{}

	for lru.sizeMB > lru.capacityMB {
		lru.responses.Remove(ejectedNode)
		ejectedNode = lru.responses.Back()
		delete(lru.cache, ejectedNode.Value.(*Node).key)

		lru.sizeMB -= lru.getSize(*ejectedNode.Value.(*Node).value)
	}

	if found {
		ejectedNode = val
	}

	ejectedNode.Value.(*Node).value = &value
	ejectedNode.Value.(*Node).key = key

	lru.cache[key] = ejectedNode
	lru.responses.MoveToFront(ejectedNode)
}

func (lru *LRUCache) LookUp(key string) *model.Response {
	lru.mtx.Lock()
	defer lru.mtx.Unlock()

	if value, found := lru.cache[key]; found {
		lru.responses.MoveToFront(value)
		return value.Value.(*Node).value
	}

	return nil
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
