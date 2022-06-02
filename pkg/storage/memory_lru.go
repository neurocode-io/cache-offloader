package storage

import (
	"container/list"
	"sync"

	"neurocode.io/cache-offloader/pkg/model"
)

type LRUCache struct {
	mtx       sync.Mutex
	responses *list.List
	cache     map[string]*list.Element
	capacity  int
}

type Node struct {
	key   string
	value *model.Response
}

func NewLRUCache(maxLRUSize int) *LRUCache {

	if maxLRUSize <= 0 {
		maxLRUSize = 50
	}

	lru := LRUCache{
		capacity:  maxLRUSize,
		responses: list.New(),
		cache:     make(map[string]*list.Element),
	}

	return &lru
}

func (lru *LRUCache) Set(key string, value model.Response) {

	lru.mtx.Lock()
	defer lru.mtx.Unlock()

	if val, found := lru.cache[key]; found {
		val.Value.(*Node).value = &value
		lru.responses.MoveToFront(val)
		return
	}

	if len(lru.cache) >= lru.capacity {
		ejectedNode := lru.responses.Back()
		delete(lru.cache, ejectedNode.Value.(*Node).key)

		ejectedNode.Value.(*Node).value = &value
		ejectedNode.Value.(*Node).key = key

		lru.cache[key] = ejectedNode
		lru.responses.MoveToFront(ejectedNode)

		return
	}

	element := lru.responses.PushFront(&Node{value: &value, key: key})
	lru.cache[key] = element
}

func (lru *LRUCache) Get(key string) *model.Response {
	lru.mtx.Lock()
	defer lru.mtx.Unlock()

	if value, found := lru.cache[key]; found {
		lru.responses.MoveToFront(value)
		return value.Value.(*Node).value
	}

	return nil
}

// Difference between get and peek is that we dont update after peek
func (lru *LRUCache) Peek(key string) *model.Response {
	lru.mtx.Lock()
	defer lru.mtx.Unlock()

	if value, found := lru.cache[key]; found {
		return value.Value.(*Node).value
	}

	return nil
}

func (lru *LRUCache) Has(key string) bool {

	lru.mtx.Lock()
	defer lru.mtx.Unlock()

	_, found := lru.cache[key]

	return found
}

func (lru *LRUCache) Len() int {

	lru.mtx.Lock()
	defer lru.mtx.Unlock()

	return len(lru.cache)
}

func (lru *LRUCache) Capacity() int {
	return lru.capacity
}

func (lru *LRUCache) Clear() {

	lru.mtx.Lock()
	defer lru.mtx.Unlock()

	lru.responses.Init()
	lru.cache = make(map[string]*list.Element)
}
