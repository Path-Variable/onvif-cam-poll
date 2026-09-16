package cli

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jessevdk/go-flags"
)

// Parse fills opts from the command line and environment. Usage errors are
// already printed by go-flags; the caller only needs to exit.
func Parse(opts any) error {
	_, err := flags.NewParser(opts, flags.Default).Parse()
	var flagsErr *flags.Error
	if errors.As(err, &flagsErr) && flagsErr.Type == flags.ErrHelp {
		os.Exit(0)
	}
	return err
}

// Context returns a context cancelled by SIGINT or SIGTERM.
func Context() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}

// Logger returns the process-wide structured logger writing to stderr.
func Logger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, nil))
}

// Sleep waits for d or until ctx is done. It reports whether the full
// duration elapsed, so loops can `if !cli.Sleep(ctx, d) { return }`.
func Sleep(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
