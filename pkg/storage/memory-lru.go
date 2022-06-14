package storage

import (
	"container/list"
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"neurocode.io/cache-offloader/pkg/model"
)

type LRUCache struct {
	mtx           sync.RWMutex
	responses     *list.List
	cache         map[string]*list.Element
	capacityMB    float64
	sizeMB        float64
	lookupTimeout time.Duration
	staleDuration int64
}

type LRUNode struct {
	key       string
	value     *model.Response
	timeStamp int64
}

func NewLRUCache(maxSizeMB float64, staleInSeconds int64) *LRUCache {
	if maxSizeMB <= 0 {
		maxSizeMB = 50.0
	}

	return &LRUCache{
		capacityMB:    maxSizeMB,
		sizeMB:        0.0,
		lookupTimeout: lookupTimeout,
		responses:     list.New(),
		cache:         make(map[string]*list.Element),
		staleDuration: staleInSeconds,
	}
}

func (lru *LRUCache) Store(ctx context.Context, key string, value *model.Response) error {
	lru.mtx.Lock()
	defer lru.mtx.Unlock()

	bodySizeMB := lru.getSize(*value)

	if bodySizeMB > lru.capacityMB {
		log.Ctx(ctx).Warn().Msg("The size of the body is bigger than the configured LRU cache maxSize. The body will not be stored.")

		return nil
	}

	if val, found := lru.cache[key]; found {
		bodySizeMB -= lru.getSize(*val.Value.(*LRUNode).value)
		node := val.Value.(*LRUNode)
		node.value = value
		node.timeStamp = time.Now().Unix()

		lru.responses.MoveToFront(val)
	} else {
		element := lru.responses.PushFront(&LRUNode{value: value, key: key, timeStamp: time.Now().Unix()})
		lru.cache[key] = element
	}

	lru.sizeMB += bodySizeMB
	var ejectedNode *list.Element

	for lru.sizeMB > lru.capacityMB {
		ejectedNode = lru.responses.Back()
		delete(lru.cache, ejectedNode.Value.(*LRUNode).key)
		lru.responses.Remove(ejectedNode)

		lru.sizeMB -= lru.getSize(*ejectedNode.Value.(*LRUNode).value)
	}

	return nil
}

func (lru *LRUCache) LookUp(ctx context.Context, key string) (*model.Response, error) {
	ctx, cancel := context.WithTimeout(ctx, lru.lookupTimeout)
	defer cancel()

	proc := make(chan *model.Response, 1)

	go func() {
		lru.mtx.RLock()
		defer lru.mtx.RUnlock()

		if value, found := lru.cache[key]; found {
			lru.responses.MoveToFront(value)
			node := value.Value.(*LRUNode)
			response := node.value
			if (time.Now().Unix() - node.timeStamp) >= lru.staleDuration {
				response.StaleValue = 0
			} else {
				response.StaleValue = 1
			}

			proc <- response

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

func (lru *LRUCache) getSize(value model.Response) float64 {
	sizeBytes := len(value.Body)
	sizeMB := float64(sizeBytes) / (1024 * 1024)

	return sizeMB
}
