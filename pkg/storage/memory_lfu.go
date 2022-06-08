package storage

import (
	"container/list"
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"neurocode.io/cache-offloader/pkg/model"
)

type LFUCache struct {
	mtx            sync.RWMutex
	min            int
	capacityMB     float64
	sizeMB         float64
	commandTimeout time.Duration
	lists          map[int]*FrequencyList
	cache          map[string]*list.Element
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

func NewLFUCache(maxSizeMB float64) *LFUCache {
	if maxSizeMB <= 0 {
		maxSizeMB = 50.0
	}

	return &LFUCache{
		min:            1,
		capacityMB:     maxSizeMB,
		sizeMB:         0,
		commandTimeout: commandTimeout,
		lists:          make(map[int]*FrequencyList),
		cache:          make(map[string]*list.Element),
	}
}

func (lfu *LFUCache) Store(ctx context.Context, key string, value *model.Response) error {
	lfu.mtx.Lock()
	defer lfu.mtx.Unlock()

	bodySizeMB := lfu.getSize(*value)

	if bodySizeMB > lfu.capacityMB {
		log.Ctx(ctx).Warn().Msg("The size of the body is bigger than the configured LRU cache maxSize. The body will not be stored.")

		return nil
	}

	val, found := lfu.cache[key]

	if found {
		bodySizeMB -= lfu.getSize(*val.Value.(*LfuNode).value)
		val.Value.(*LfuNode).value = value
		lfu.update(val)
	}

	lfu.sizeMB += bodySizeMB

	for lfu.sizeMB > lfu.capacityMB {
		freqList := lfu.lists[lfu.min].lruCache
		ejectedNode := freqList.Back()
		freqList.Remove(ejectedNode)

		if freqList.Len() == 0 {
			delete(lfu.lists, lfu.min)
		}

		delete(lfu.cache, ejectedNode.Value.(*LfuNode).key)

		lfu.sizeMB -= lfu.getSize(*ejectedNode.Value.(*LfuNode).value)
	}

	if !found {
		node := &LfuNode{
			key:   key,
			value: value,
		}

		addedLfuNode := lfu.moveNode(node, 1)
		lfu.cache[key] = addedLfuNode
		lfu.min = 1
	}

	return nil
}

func (lfu *LFUCache) LookUp(ctx context.Context, key string) (*model.Response, error) {
	ctx, cancel := context.WithTimeout(ctx, lfu.commandTimeout)
	defer cancel()

	proc := make(chan *model.Response, 1)

	go func() {
		lfu.mtx.RLock()
		defer lfu.mtx.RUnlock()

		if val, found := lfu.cache[key]; found {
			lfu.update(val)
			proc <- val.Value.(*LfuNode).value

			return
		}

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

	lfu.moveNode(node.Value.(*LfuNode), count+1)
}

func (lfu *LFUCache) moveNode(node *LfuNode, count int) *list.Element {
	if _, found := lfu.lists[count]; !found {
		lfu.lists[count] = &FrequencyList{
			lruCache: list.New(),
			counter:  count,
		}
	}

	returnedLfuNode := lfu.lists[count].lruCache.PushFront(node)
	returnedLfuNode.Value.(*LfuNode).parent = lfu.lists[count]
	lfu.cache[returnedLfuNode.Value.(*LfuNode).key] = returnedLfuNode

	return returnedLfuNode
}

func (lfu *LFUCache) getSize(value model.Response) float64 {
	sizeBytes := len(value.Body)
	sizeMB := float64(sizeBytes) / (1024 * 1024)

	return sizeMB
}
