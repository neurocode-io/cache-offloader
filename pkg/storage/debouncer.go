package storage

import (
	"container/list"
	"net/http"
	"sync"
)

type Debouncer struct {
	mtx        sync.RWMutex
	requests   *list.List
	cache      map[string]*list.Element
	capacityMB float64
	sizeMB     float64
}

type RequestNode struct {
	value *http.Request
}

func NewDebouncer(maxSizeMB float64) *Debouncer {
	if maxSizeMB <= 0 {
		maxSizeMB = 50.0
	}

	return &Debouncer{
		capacityMB: maxSizeMB,
		sizeMB:     0.0,
		requests:   list.New(),
		cache:      make(map[string]*list.Element),
	}
}

func (debouncer *Debouncer) Store(key string, value *http.Request) {
	debouncer.mtx.Lock()
	defer debouncer.mtx.Unlock()

	bodySizeMB := debouncer.getSize(*value)

	if bodySizeMB < 0 || (bodySizeMB+debouncer.sizeMB > debouncer.capacityMB) {
		return
	}

	if _, found := debouncer.cache[key]; !found {
		element := debouncer.requests.PushFront(&RequestNode{value: value})
		debouncer.cache[key] = element
	}
}

func (debouncer *Debouncer) GetNext() *http.Request {
	debouncer.mtx.RLock()
	defer debouncer.mtx.RUnlock()

	return debouncer.requests.Back().Value.(*RequestNode).value
}

func (debouncer *Debouncer) Erase(key string) bool {
	debouncer.mtx.RLock()
	defer debouncer.mtx.RUnlock()

	if val, found := debouncer.cache[key]; found {
		delete(debouncer.cache, key)
		debouncer.requests.Remove(val)

		return true
	}

	return false
}

func (debouncer *Debouncer) getSize(value http.Request) float64 {
	sizeBytes := value.ContentLength

	if sizeBytes < 0 {
		return -1.0
	}

	sizeMB := float64(sizeBytes) / (1024 * 1024)

	return sizeMB
}
