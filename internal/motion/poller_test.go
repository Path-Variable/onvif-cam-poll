package motion

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/path-variable/onvif-cam-poll/internal/backoff"
)

// fakeSub replays scripted results: each Pull returns the next entry.
type fakeSub struct {
	mu         sync.Mutex
	subscribes int
	pulls      []pullResult
	next       int
	cancel     context.CancelFunc
	failSub    int // number of Subscribe calls that fail first
}

type pullResult struct {
	motion bool
	err    error
}

func (f *fakeSub) Subscribe(context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.subscribes++
	if f.subscribes <= f.failSub {
		return errors.New("camera busy")
	}
	return nil
}

func (f *fakeSub) Pull(context.Context) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.next >= len(f.pulls) {
		f.cancel() // script exhausted: stop the poller
		return false, nil
	}
	r := f.pulls[f.next]
	f.next++
	return r.motion, r.err
}

func newPoller(sub *fakeSub, onMotion func(context.Context)) *Poller {
	return &Poller{
		Sub:             sub,
		OnMotion:        onMotion,
		PollInterval:    time.Millisecond,
		Cooldown:        2 * time.Millisecond,
		SubscriptionTTL: time.Minute,
		Backoff:         backoff.New(time.Millisecond, 2*time.Millisecond),
		Log:             slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

func run(t *testing.T, sub *fakeSub, p *Poller) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	sub.cancel = cancel
	if err := p.Run(ctx); !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run returned %v", err)
	}
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatal("poller did not finish the script in time")
	}
}

func TestMotionTriggersHandlerOncePerEvent(t *testing.T) {
	sub := &fakeSub{pulls: []pullResult{{false, nil}, {true, nil}, {false, nil}, {true, nil}}}
	var events int
	run(t, sub, newPoller(sub, func(context.Context) { events++ }))
	if events != 2 {
		t.Fatalf("handler called %d times, want 2", events)
	}
	if sub.subscribes != 1 {
		t.Fatalf("subscribed %d times, want 1", sub.subscribes)
	}
}

func TestPullErrorResubscribesInsteadOfExiting(t *testing.T) {
	sub := &fakeSub{pulls: []pullResult{{false, errors.New("timeout")}, {false, nil}, {true, nil}}}
	var events int
	run(t, sub, newPoller(sub, func(context.Context) { events++ }))
	if sub.subscribes != 2 {
		t.Fatalf("subscribed %d times, want 2 (initial + after the failed pull)", sub.subscribes)
	}
	if events != 1 {
		t.Fatalf("handler called %d times, want 1", events)
	}
}

func TestSubscribeFailuresAreRetried(t *testing.T) {
	sub := &fakeSub{failSub: 3, pulls: []pullResult{{true, nil}}}
	var events int
	run(t, sub, newPoller(sub, func(context.Context) { events++ }))
	if sub.subscribes != 4 {
		t.Fatalf("subscribed %d times, want 4", sub.subscribes)
	}
	if events != 1 {
		t.Fatalf("handler called %d times, want 1", events)
	}
}

func TestSubscriptionIsRenewedAfterTTL(t *testing.T) {
	sub := &fakeSub{pulls: []pullResult{{false, nil}, {false, nil}, {false, nil}, {false, nil}}}
	p := newPoller(sub, func(context.Context) {})
	p.SubscriptionTTL = 0 // expires immediately: every loop iteration re-subscribes
	run(t, sub, p)
	if sub.subscribes < 2 {
		t.Fatalf("subscribed %d times, want at least 2", sub.subscribes)
	}
}

func TestCancelStopsRunPromptly(t *testing.T) {
	sub := &fakeSub{failSub: 1 << 30} // subscribe never succeeds
	p := newPoller(sub, func(context.Context) {})
	p.Backoff = backoff.New(time.Hour, time.Hour) // would block for an hour if cancellation were ignored
	ctx, cancel := context.WithCancel(context.Background())
	sub.cancel = cancel
	done := make(chan error, 1)
	go func() { done <- p.Run(ctx) }()
	time.Sleep(5 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("got %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run did not return after cancel")
	}
}
