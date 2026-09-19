package indexing

import (
	"context"
	"math/rand/v2"
	"time"
)

const (
	minBackoff = 1 * time.Second
	maxBackoff = 60 * time.Second
)

// backoff is a simple exponential backoff with jitter, used to avoid
// hammering a struggling RPC provider on repeated transient failures.
type backoff struct {
	current time.Duration
}

func newBackoff() *backoff {
	return &backoff{current: minBackoff}
}

func (b *backoff) next() time.Duration {
	delay := b.current
	b.current *= 2
	if b.current > maxBackoff {
		b.current = maxBackoff
	}
	jitter := time.Duration(rand.Int64N(int64(delay) / 2))
	return delay + jitter
}

func (b *backoff) reset() {
	b.current = minBackoff
}

// sleepOrDone waits for either the duration to elapse or ctx to be
// cancelled, whichever comes first. Returns false if ctx was cancelled.
func sleepOrDone(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	}
}
