// Command onvif-set-time keeps a camera's clock in sync with this host,
// for cameras that cannot reach an NTP server. It re-sends the time after
// every cooldown until interrupted.
package main

import (
	"os"
	"time"

	sdkdevice "github.com/use-go/onvif/sdk/device"

	"github.com/path-variable/onvif-cam-poll/internal/camera"
	"github.com/path-variable/onvif-cam-poll/internal/cli"
	"github.com/path-variable/onvif-cam-poll/internal/systime"
)

type options struct {
	cli.Device
	cli.Cooldown
}

func main() {
	var opts options
	if err := cli.Parse(&opts); err != nil {
		os.Exit(2)
	}
	log := cli.Logger().With("address", opts.Address)
	ctx, stop := cli.Context()
	defer stop()

	dev, err := camera.Connect(opts.Device)
	if err != nil {
		log.Error("cannot connect", "err", err)
		os.Exit(1)
	}
	for {
		now := time.Now()
		if _, err := sdkdevice.Call_SetSystemDateAndTime(ctx, dev, systime.Request(now)); err != nil {
			log.Warn("set time failed", "err", err)
		} else {
			log.Info("time set", "utc", now.UTC().Format(time.RFC3339))
		}
		if !cli.Sleep(ctx, opts.Cooldown.Duration()) {
			return
		}
	}
}
