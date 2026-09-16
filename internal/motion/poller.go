package motion

import (
	"context"
	"log/slog"
	"time"

	"github.com/path-variable/onvif-cam-poll/internal/backoff"
	"github.com/path-variable/onvif-cam-poll/internal/cli"
)

// Subscriber is the camera side of the poll loop. PullPoint implements it
// for ONVIF; tests supply a fake.
type Subscriber interface {
	// Subscribe creates (or re-creates) the event subscription.
	Subscribe(ctx context.Context) error
	// Pull fetches pending events and reports whether any of them is motion.
	Pull(ctx context.Context) (bool, error)
}

// Poller drives a Subscriber until the context is cancelled. It never gives
// up on errors: it re-subscribes with exponential backoff, because the whole
// point of the daemon is to keep running unattended.
type Poller struct {
	Sub      Subscriber
	OnMotion func(ctx context.Context)

	PollInterval    time.Duration // pause between successful pulls
	Cooldown        time.Duration // pause after a motion event
	SubscriptionTTL time.Duration // how long a subscription is trusted before renewing
	Backoff         *backoff.Backoff
	Log             *slog.Logger
}

// Run blocks until ctx is done and returns ctx.Err().
func (p *Poller) Run(ctx context.Context) error {
	for ctx.Err() == nil {
		if err := p.Sub.Subscribe(ctx); err != nil {
			p.retry(ctx, "subscribe failed", err)
			continue
		}
		p.Backoff.Reset()
		p.Log.Info("subscribed to motion events", "ttl", p.SubscriptionTTL)
		if err := p.pullUntilExpiry(ctx); err != nil {
			p.retry(ctx, "pull failed; re-subscribing", err)
		}
	}
	return ctx.Err()
}

// pullUntilExpiry polls until the subscription is about to expire (nil) or
// a pull fails (error). Either way the caller re-subscribes. A fresh
// subscription is always pulled at least once.
func (p *Poller) pullUntilExpiry(ctx context.Context) error {
	renewAt := time.Now().Add(p.SubscriptionTTL)
	for first := true; ctx.Err() == nil && (first || time.Now().Before(renewAt)); first = false {
		motion, err := p.Sub.Pull(ctx)
		if err != nil {
			return err
		}
		p.Backoff.Reset()
		if motion {
			p.Log.Info("motion detected")
			p.OnMotion(ctx)
			if !cli.Sleep(ctx, p.Cooldown) {
				return nil
			}
		}
		if !cli.Sleep(ctx, p.PollInterval) {
			return nil
		}
	}
	return nil
}

func (p *Poller) retry(ctx context.Context, msg string, err error) {
	if ctx.Err() != nil {
		return
	}
	delay := p.Backoff.Next()
	p.Log.Warn(msg, "err", err, "retry_in", delay)
	cli.Sleep(ctx, delay)
}
