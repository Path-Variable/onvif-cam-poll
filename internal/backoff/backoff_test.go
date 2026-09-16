package backoff

import (
	"testing"
	"time"
)

func TestDoublesUpToMaxAndResets(t *testing.T) {
	b := New(time.Second, 5*time.Second)
	want := []time.Duration{time.Second, 2 * time.Second, 4 * time.Second, 5 * time.Second, 5 * time.Second}
	for i, w := range want {
		if got := b.Next(); got != w {
			t.Fatalf("call %d: got %v, want %v", i, got, w)
		}
	}
	b.Reset()
	if got := b.Next(); got != time.Second {
		t.Fatalf("after reset: got %v, want 1s", got)
	}
}

func TestMaxBelowMinIsClamped(t *testing.T) {
	b := New(3*time.Second, time.Second)
	if got := b.Next(); got != 3*time.Second {
		t.Fatalf("got %v, want 3s", got)
	}
}
