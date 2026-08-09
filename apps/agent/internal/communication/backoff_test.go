package communication

import (
	"testing"
	"time"
)

func TestReconnectBackoffGrowthCapJitterAndReset(t *testing.T) {
	var limits []time.Duration
	backoff := newReconnectBackoff(func(limit time.Duration) time.Duration {
		limits = append(limits, limit)
		return limit / 2
	})

	wantLimits := []time.Duration{
		time.Second,
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
		16 * time.Second,
		30 * time.Second,
		30 * time.Second,
	}

	for _, wantLimit := range wantLimits {
		if got := backoff.next(); got != wantLimit/2 {
			t.Errorf("next delay = %s, want jittered delay %s", got, wantLimit/2)
		}
	}

	for i, want := range wantLimits {
		if limits[i] != want {
			t.Errorf("jitter limit %d = %s, want %s", i, limits[i], want)
		}
	}

	backoff.reset()

	if got := backoff.next(); got != reconnectBackoffInitial/2 {
		t.Errorf("delay after reset = %s, want %s", got, reconnectBackoffInitial/2)
	}
	if got := limits[len(limits)-1]; got != reconnectBackoffInitial {
		t.Errorf("jitter limit after reset = %s, want %s", got, reconnectBackoffInitial)
	}
}
