package backend

import (
	"sync"
	"time"
)

type EventBus struct {
	mu     sync.RWMutex
	subs   map[int]chan Event
	nextID int
}

func NewEventBus() *EventBus {
	return &EventBus{
		subs: make(map[int]chan Event),
	}
}

func (b *EventBus) Subscribe(handler func(Event)) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.nextID++
	id := b.nextID
	ch := make(chan Event, 100)
	b.subs[id] = ch

	go func() {
		for ev := range ch {
			handler(ev)
		}
	}()

	return id
}

func (b *EventBus) Unsubscribe(id int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if ch, ok := b.subs[id]; ok {
		close(ch)
		delete(b.subs, id)
	}
}

func (b *EventBus) Emit(ev Event) {
	if ev.Time.IsZero() {
		ev.Time = time.Now()
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subs {
		select {
		case ch <- ev:
		default:
			// Buffer cheio, ignora para este assinante
		}
	}
}
