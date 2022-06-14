package worker

import (
	"sync"

	"github.com/rs/zerolog/log"
)

type UpdateQueue struct {
	mtx   sync.RWMutex
	queue map[string]bool
	size  int
}

func NewUpdateQueue(size int) *UpdateQueue {
	if size <= 0 {
		size = 1000
	}

	return &UpdateQueue{
		queue: make(map[string]bool, size),
		size:  size,
	}
}

func (debouncer *UpdateQueue) Start(key string, work func()) {
	if len(debouncer.queue) >= debouncer.size {
		log.Warn().Msg("UpdateQueue is full, dropping request")

		return
	}

	debouncer.mtx.Lock()

	if _, ok := debouncer.queue[key]; ok {
		debouncer.mtx.Unlock()

		return
	}
	debouncer.queue[key] = true
	debouncer.mtx.Unlock()

	work()

	debouncer.mtx.Lock()
	delete(debouncer.queue, key)
	debouncer.mtx.Unlock()
}
