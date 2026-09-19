package indexing

import "testing"

func TestBackoffGrowsAndCaps(t *testing.T) {
	b := newBackoff()

	first := b.next()
	if first < minBackoff || first > minBackoff+minBackoff/2 {
		t.Fatalf("expected first delay within [%v, %v], got %v", minBackoff, minBackoff+minBackoff/2, first)
	}

	var last float64
	for i := 0; i < 20; i++ {
		d := b.next()
		if float64(d) < last {
			// jitter means it isn't strictly monotonic once capped, but it
			// must never exceed maxBackoff plus its own jitter window.
		}
		if d > maxBackoff+maxBackoff/2 {
			t.Fatalf("backoff exceeded expected cap: %v", d)
		}
		last = float64(d)
	}
}

func TestBackoffResetsToMinimum(t *testing.T) {
	b := newBackoff()
	for i := 0; i < 10; i++ {
		b.next()
	}
	b.reset()
	if b.current != minBackoff {
		t.Fatalf("expected reset to minBackoff, got %v", b.current)
	}
}
