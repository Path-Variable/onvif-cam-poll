// Package backoff implements a minimal exponential backoff.
package backoff

import "time"

// Backoff doubles the delay on every Next call up to Max and starts again
// from Min after Reset. The zero value is not usable; use New.
type Backoff struct {
	min, max, next time.Duration
}

// New returns a backoff starting at min and capped at max.
func New(min, max time.Duration) *Backoff {
	if max < min {
		max = min
	}
	return &Backoff{min: min, max: max, next: min}
}

// Next returns the current delay and doubles it for the following call.
func (b *Backoff) Next() time.Duration {
	d := b.next
	if b.next < b.max {
		b.next = min(b.next*2, b.max)
	}
	return d
}

// Reset returns the delay to its minimum after a success.
func (b *Backoff) Reset() { b.next = b.min }
