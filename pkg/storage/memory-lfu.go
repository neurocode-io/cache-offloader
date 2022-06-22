package storage

import (
	"container/list"
	"context"
	"math"
	"sync"
	"time"

	"github.com/neurocode-io/cache-offloader/pkg/model"
	"github.com/rs/zerolog/log"
)

type LFUCache struct {
	mtx           sync.RWMutex
	min           uint
	capacityMB    float64
	sizeMB        float64
	lookupTimeout time.Duration
	lists         map[uint]*FrequencyList
	cache         map[string]*list.Element
	staleDuration int
}

type FrequencyList struct {
	lruCache *list.List
	counter  uint
}

type LfuNode struct {
	parent    *FrequencyList
	value     *model.Response
	timeStamp int64
	key       string
}

func NewLFUCache(maxSizeMB float64, staleInSeconds int) *LFUCache {
	if maxSizeMB <= 0 {
		maxSizeMB = 50.0
	}

	return &LFUCache{
		min:           1,
		capacityMB:    maxSizeMB,
		lookupTimeout: lookupTimeout,
		sizeMB:        0,
		lists:         make(map[uint]*FrequencyList),
		cache:         make(map[string]*list.Element),
		staleDuration: staleInSeconds,
	}
}

func (lfu *LFUCache) Store(ctx context.Context, key string, value *model.Response) error {
	lfu.mtx.Lock()
	defer lfu.mtx.Unlock()

	bodySizeMB := getSize(*value)

	if bodySizeMB > lfu.capacityMB {
		log.Ctx(ctx).Warn().Msg("The size of the body is bigger than the configured LRU cache maxSize. The body will not be stored.")

		return nil
	}

	val, found := lfu.cache[key]

	if found {
		bodySizeMB -= getSize(*val.Value.(*LfuNode).value)
		node, ok := val.Value.(*LfuNode)
		if !ok {
			log.Warn().Msg("The node is not a LfuNode")
		}
		node.value = value
		node.timeStamp = time.Now().Unix()
		lfu.update(val)
	}

	lfu.sizeMB += bodySizeMB

	for lfu.sizeMB > lfu.capacityMB {
		lfu.ejectNode()
	}

	if !found {
		node := &LfuNode{
			key:       key,
			value:     value,
			timeStamp: time.Now().Unix(),
		}

		addedLfuNode := lfu.moveNode(node, 1)
		lfu.cache[key] = addedLfuNode
		lfu.min = 1
	}

	return nil
}

func (lfu *LFUCache) LookUp(ctx context.Context, key string) (*model.Response, error) {
	ctx, cancel := context.WithTimeout(ctx, lfu.lookupTimeout)
	defer cancel()

	proc := make(chan *model.Response, 1)

	go func() {
		lfu.mtx.Lock()
		defer lfu.mtx.Unlock()

		if val, found := lfu.cache[key]; found {
			lfu.update(val)
			node, ok := val.Value.(*LfuNode)
			if !ok {
				log.Warn().Msg("The node is not a LfuNode")
			}
			response := node.value
			response.StaleValue = getStaleStatus(node.timeStamp, lfu.staleDuration)

			proc <- response

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
	if node == nil {
		return
	}
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

func (lfu *LFUCache) moveNode(node *LfuNode, count uint) *list.Element {
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

func (lfu *LFUCache) ejectNode() {
	freqList := lfu.lists[lfu.min].lruCache
	ejectedNode := freqList.Back()
	freqList.Remove(ejectedNode)

	if freqList.Len() == 0 {
		delete(lfu.lists, lfu.min)
		lfu.min = lfu.findNextMin()
	}

	delete(lfu.cache, ejectedNode.Value.(*LfuNode).key)

	lfu.sizeMB -= getSize(*ejectedNode.Value.(*LfuNode).value)
}

func (lfu *LFUCache) findNextMin() uint {
	if _, found := lfu.lists[lfu.min+1]; found {
		return lfu.min + 1
	}

	var newMin uint = math.MaxUint

	for k := range lfu.lists {
		if k < newMin {
			newMin = k
		}
	}

	return newMin
}
