package worker

import (
	"sync"

	"github.com/rs/zerolog/log"
)

type UpdateQueue struct {
	mtx   sync.Mutex
	queue map[string]struct{}
	size  int
}

func NewUpdateQueue(size int) *UpdateQueue {
	if size <= 0 {
		size = 1000
	}
	return &UpdateQueue{
		queue: make(map[string]struct{}, size),
		size:  size,
	}
}

func (u *UpdateQueue) Start(key string, work func()) {
	u.mtx.Lock()
	// capacity check under lock to avoid races
	if len(u.queue) >= u.size {
		u.mtx.Unlock()
		log.Warn().Msg("UpdateQueue is full, dropping request")
		return
	}
	// if already queued, skip
	if _, ok := u.queue[key]; ok {
		u.mtx.Unlock()
		return
	}
	u.queue[key] = struct{}{}
	u.mtx.Unlock()

	defer func() {
		// ensure cleanup even on panic
		if r := recover(); r != nil {
			log.Error().Interface("panic", r).Msg("worker panic in revalidate")
		}
		u.mtx.Lock()
		delete(u.queue, key)
		u.mtx.Unlock()
	}()

	// execute work synchronously; caller can decide to run in a goroutine
	work()
}
