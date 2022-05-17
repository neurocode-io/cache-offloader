package storage

import (
	"container/list"
	"errors"
	"sync"
)

type LRUMap struct {
	mtx       sync.Mutex
	responses *list.List
	cache     map[string]*list.Element
	maxSize   int
}

type Node struct {
	key   string
	value *Response
}

func NewLRUMap(maxLRUSize int) (*LRUMap, error) {

	if maxLRUSize <= 0 {
		return nil, errors.New("size must be a postive int")
	}

	lru := &LRUMap{
		maxSize:   maxLRUSize,
		responses: list.New(),
		cache:     make(map[string]*list.Element),
	}

	return lru, nil
}

func (lru *LRUMap) Set(key string, value Response) {

	lru.mtx.Lock()
	defer lru.mtx.Unlock()

	if val, found := lru.cache[key]; found {
		val.Value.(*Node).value = &value
		lru.responses.MoveToFront(val)
		return
	}

	if len(lru.cache) >= lru.maxSize {
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

func (lru *LRUMap) Get(key string) (*Response, bool) {
	lru.mtx.Lock()
	defer lru.mtx.Unlock()

	if value, found := lru.cache[key]; found {
		lru.responses.MoveToFront(value)
		return value.Value.(*Node).value, true
	}

	return nil, false
}

// Difference between get and peek is that we dont update after peek
func (lru *LRUMap) Peek(key string) (*Response, bool) {
	lru.mtx.Lock()
	defer lru.mtx.Unlock()

	if value, found := lru.cache[key]; found {
		return value.Value.(*Node).value, true
	}

	return nil, false
}

func (lru *LRUMap) Has(key string) bool {

	lru.mtx.Lock()
	defer lru.mtx.Unlock()

	_, found := lru.cache[key]

	return found
}

func (lru *LRUMap) Len() int {

	lru.mtx.Lock()
	defer lru.mtx.Unlock()

	return len(lru.cache)
}

func (lru *LRUMap) Clear() {

	lru.mtx.Lock()
	defer lru.mtx.Unlock()

	lru.responses.Init()
	lru.cache = make(map[string]*list.Element)
}
