package communication

import (
	"math/rand/v2"
	"time"
)

const (
	reconnectBackoffInitial = time.Second
	reconnectBackoffMax     = 30 * time.Second
)

type reconnectBackoff struct {
	current time.Duration
	jitter  func(time.Duration) time.Duration
}

func newReconnectBackoff(jitter func(time.Duration) time.Duration) *reconnectBackoff {
	return &reconnectBackoff{
		current: reconnectBackoffInitial,
		jitter:  jitter,
	}
}

func (b *reconnectBackoff) next() time.Duration {
	delay := b.jitter(b.current)

	b.current *= 2
	if b.current > reconnectBackoffMax {
		b.current = reconnectBackoffMax
	}

	return delay
}

func (b *reconnectBackoff) reset() {
	b.current = reconnectBackoffInitial
}

func fullJitter(limit time.Duration) time.Duration {
	return time.Duration(rand.Int64N(int64(limit)))
}
