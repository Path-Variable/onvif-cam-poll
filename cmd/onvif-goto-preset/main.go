// Command onvif-goto-preset keeps sending a camera back to a PTZ preset,
// once per cooldown, until interrupted.
package main

import (
	"os"

	"github.com/use-go/onvif/ptz"
	sdkptz "github.com/use-go/onvif/sdk/ptz"
	xonvif "github.com/use-go/onvif/xsd/onvif"

	"github.com/path-variable/onvif-cam-poll/internal/camera"
	"github.com/path-variable/onvif-cam-poll/internal/cli"
)

type options struct {
	cli.Device
	cli.Cooldown
	cli.Preset
}

func main() {
	var opts options
	if err := cli.Parse(&opts); err != nil {
		os.Exit(2)
	}
	log := cli.Logger().With("address", opts.Address, "preset", opts.Token)
	ctx, stop := cli.Context()
	defer stop()

	dev, err := camera.Connect(opts.Device)
	if err != nil {
		log.Error("cannot connect", "err", err)
		os.Exit(1)
	}
	req := ptz.GotoPreset{PresetToken: xonvif.ReferenceToken(opts.Token), ProfileToken: xonvif.ReferenceToken(opts.Profile)}
	for {
		if _, err := sdkptz.Call_GotoPreset(ctx, dev, req); err != nil {
			log.Warn("goto preset failed", "err", err)
		} else {
			log.Info("moved to preset")
		}
		if !cli.Sleep(ctx, opts.Cooldown.Duration()) {
			return
		}
	}
}
