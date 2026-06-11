package pubsub

import (
	"context"
	"log/slog"
	"sync"
)

const bufferSize = 64

type Broker[T any] struct {
	subs        map[chan Event[T]]struct{}
	mu          sync.RWMutex
	done        chan struct{}
	subCount    int
	maxEvents   int
	overflowLog int // count of overflowed events for this broker
}

func NewBroker[T any]() *Broker[T] {
	return NewBrokerWithOptions[T](bufferSize, 1000)
}

func NewBrokerWithOptions[T any](channelBufferSize, maxEvents int) *Broker[T] {
	return &Broker[T]{
		subs:      make(map[chan Event[T]]struct{}),
		done:      make(chan struct{}),
		maxEvents: maxEvents,
	}
}

func (b *Broker[T]) Shutdown() {
	select {
	case <-b.done: // Already closed
		return
	default:
		close(b.done)
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	for ch := range b.subs {
		delete(b.subs, ch)
		close(ch)
	}

	b.subCount = 0
}

func (b *Broker[T]) Subscribe(ctx context.Context) <-chan Event[T] {
	b.mu.Lock()
	defer b.mu.Unlock()

	select {
	case <-b.done:
		ch := make(chan Event[T])
		close(ch)
		return ch
	default:
	}

	sub := make(chan Event[T], bufferSize)
	b.subs[sub] = struct{}{}
	b.subCount++

	go func() {
		<-ctx.Done()

		b.mu.Lock()
		defer b.mu.Unlock()

		select {
		case <-b.done:
			return
		default:
		}

		delete(b.subs, sub)
		close(sub)
		b.subCount--
	}()

	return sub
}

func (b *Broker[T]) GetSubscriberCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.subCount
}

func (b *Broker[T]) GetOverflowCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.overflowLog
}

func (b *Broker[T]) Publish(t EventType, payload T) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	select {
	case <-b.done:
		return
	default:
	}

	event := Event[T]{Type: t, Payload: payload}

	for sub := range b.subs {
		select {
		case sub <- event:
		default:
			// Channel is full, subscriber is slow - log overflow once per
			// 100 dropped events to avoid log spam while still surfacing
			// the problem.
			b.overflowLog++
			if b.overflowLog%100 == 0 {
				slog.Warn("Subscriber channel full, dropping events",
					"dropped_total", b.overflowLog,
					"subscriber_count", b.subCount,
				)
			}
		}
	}
}
