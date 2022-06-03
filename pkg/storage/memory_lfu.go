package storage

import (
	"container/list"
	"sync"

	"neurocode.io/cache-offloader/pkg/model"
)

type LFUCache struct {
	mtx      sync.Mutex
	min      int
	capacity int
	lists    map[int]*FrequencyList
	nodes    map[string]*list.Element
}

type FrequencyList struct {
	lruCache *list.List
	counter  int
}

type LfuNode struct {
	parent *FrequencyList
	value  *model.Response
	key    string
}

func NewLFUCache(maxLFUSize int) *LFUCache {
	if maxLFUSize <= 0 {
		maxLFUSize = 50
	}

	return &LFUCache{
		min:      1,
		capacity: maxLFUSize,
		lists:    make(map[int]*FrequencyList),
		nodes:    make(map[string]*list.Element),
	}
}

func (lfu *LFUCache) Set(key string, value model.Response) {
	if lfu.capacity == 0 {
		return
	}

	if val, found := lfu.nodes[key]; found {
		val.Value.(*LfuNode).value = &value
		lfu.update(val)

		return
	}

	if len(lfu.nodes) == lfu.capacity {
		freqList := lfu.lists[lfu.min].lruCache
		ejectedLfuNode := freqList.Back()
		freqList.Remove(ejectedLfuNode)

		if freqList.Len() == 0 {
			delete(lfu.lists, lfu.min)
		}

		delete(lfu.nodes, ejectedLfuNode.Value.(*LfuNode).key)
	}

	node := &LfuNode{
		key:   key,
		value: &value,
	}

	addedLfuNode := lfu.moveLfuNode(node, 1)
	lfu.nodes[key] = addedLfuNode
	lfu.min = 1
}

func (lfu *LFUCache) Get(key string) *model.Response {
	lfu.mtx.Lock()
	defer lfu.mtx.Unlock()

	if val, found := lfu.nodes[key]; found {
		lfu.update(val)

		return val.Value.(*LfuNode).value
	}

	return nil
}

// Difference between get and peek is that we dont update after peek
func (lfu *LFUCache) Peek(key string) *model.Response {
	lfu.mtx.Lock()
	defer lfu.mtx.Unlock()

	if val, found := lfu.nodes[key]; found {
		return val.Value.(*LfuNode).value
	}

	return nil
}

func (lfu *LFUCache) Has(key string) bool {
	lfu.mtx.Lock()
	defer lfu.mtx.Unlock()

	_, found := lfu.nodes[key]

	return found
}

func (lfu *LFUCache) Len() int {
	lfu.mtx.Lock()
	defer lfu.mtx.Unlock()

	return len(lfu.nodes)
}

func (lfu *LFUCache) Clear() {
	lfu.mtx.Lock()
	defer lfu.mtx.Unlock()

	lfu.lists = make(map[int]*FrequencyList)
	lfu.nodes = make(map[string]*list.Element)
}

func (lfu *LFUCache) update(node *list.Element) {
	parent := node.Value.(*LfuNode).parent
	count := parent.counter
	freqList := parent.lruCache

	freqList.Remove(node)
	if freqList.Len() == 0 {
		if lfu.min == count {
			lfu.min = count + 1
		}

		delete(lfu.lists, count)
	}

	lfu.moveLfuNode(node.Value.(*LfuNode), count+1)
}

func (lfu *LFUCache) moveLfuNode(node *LfuNode, count int) *list.Element {
	if _, found := lfu.lists[count]; !found {
		lfu.lists[count] = &FrequencyList{
			lruCache: list.New(),
			counter:  count,
		}
	}

	returnedLfuNode := lfu.lists[count].lruCache.PushFront(node)
	returnedLfuNode.Value.(*LfuNode).parent = lfu.lists[count]
	lfu.nodes[returnedLfuNode.Value.(*LfuNode).key] = returnedLfuNode

	return returnedLfuNode
}

func (lfu *LFUCache) Capacity() int {
	return lfu.capacity
}
