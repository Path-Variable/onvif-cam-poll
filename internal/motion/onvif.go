package motion

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/use-go/onvif"
	"github.com/use-go/onvif/event"
	"github.com/use-go/onvif/xsd"
)

// PullPoint is the ONVIF pull-point implementation of Subscriber.
//
// The subscription is addressed through the device's event service, which
// is what use-go/onvif supports and what the cameras this tool was written
// for accept. Cameras that require the SubscriptionReference address will
// fail every Pull; the Poller then re-subscribes with backoff.
type PullPoint struct {
	Dev *onvif.Device
	// TTL is requested as the subscription's InitialTerminationTime.
	TTL time.Duration
	// PullTimeout is how long the camera may hold a Pull before answering.
	PullTimeout time.Duration
	// MessageLimit caps the events returned per Pull.
	MessageLimit int
}

// Subscribe creates a fresh pull-point subscription for changed events only.
func (p *PullPoint) Subscribe(ctx context.Context) error {
	req := event.CreatePullPointSubscription{
		SubscriptionPolicy:     event.SubscriptionPolicy{ChangedOnly: true},
		InitialTerminationTime: event.AbsoluteOrRelativeTimeType{Duration: xsd.Duration(isoDuration(p.TTL))},
	}
	resp, err := p.call(ctx, req)
	if err != nil {
		return err
	}
	return drain(resp)
}

// Pull fetches pending events and reports whether any of them is motion.
func (p *PullPoint) Pull(ctx context.Context) (bool, error) {
	req := event.PullMessages{
		Timeout:      xsd.Duration(isoDuration(p.PullTimeout)),
		MessageLimit: xsd.Int(p.MessageLimit),
	}
	resp, err := p.call(ctx, req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return ContainsMotion(resp.Body)
}

func (p *PullPoint) call(ctx context.Context, method any) (*http.Response, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	resp, err := p.Dev.CallMethod(method)
	if err != nil {
		return nil, fmt.Errorf("%T: %w", method, err)
	}
	if resp.StatusCode != http.StatusOK {
		_ = drain(resp)
		return nil, fmt.Errorf("%T: camera answered HTTP %d", method, resp.StatusCode)
	}
	return resp, nil
}

// drain reads and closes a response body so the connection can be reused.
func drain(resp *http.Response) error {
	_, err := io.Copy(io.Discard, resp.Body)
	if cerr := resp.Body.Close(); err == nil {
		err = cerr
	}
	return err
}

// isoDuration renders whole seconds as an ISO 8601 duration ("PT300S").
func isoDuration(d time.Duration) string {
	return fmt.Sprintf("PT%dS", int(d.Seconds()))
}
