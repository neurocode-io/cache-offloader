package storage

import (
	"sync"
)

type Debouncer struct {
	mtx   sync.RWMutex
	queue map[string]bool
}

func NewDebouncer(size int) *Debouncer {
	if size <= 0 {
		size = 1000
	}

	return &Debouncer{
		queue: make(map[string]bool, size),
	}
}

func (debouncer *Debouncer) Start(key string, work func()) {
	debouncer.mtx.Lock()

	if _, ok := debouncer.queue[key]; ok {
		return
	}
	debouncer.queue[key] = true
	debouncer.mtx.Unlock()

	work()

	debouncer.mtx.Lock()
	delete(debouncer.queue, key)
	debouncer.mtx.Unlock()
}
