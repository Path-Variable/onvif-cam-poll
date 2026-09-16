// Command onvif-set-preset stores the camera's current position under a
// PTZ preset token.
package main

import (
	"context"
	"os"

	"github.com/use-go/onvif/ptz"
	sdkptz "github.com/use-go/onvif/sdk/ptz"
	xonvif "github.com/use-go/onvif/xsd/onvif"

	"github.com/path-variable/onvif-cam-poll/internal/camera"
	"github.com/path-variable/onvif-cam-poll/internal/cli"
)

type options struct {
	cli.Device
	cli.Preset
}

func main() {
	var opts options
	if err := cli.Parse(&opts); err != nil {
		os.Exit(2)
	}
	log := cli.Logger().With("address", opts.Address, "preset", opts.Token)

	dev, err := camera.Connect(opts.Device)
	if err != nil {
		log.Error("cannot connect", "err", err)
		os.Exit(1)
	}
	req := ptz.SetPreset{PresetToken: xonvif.ReferenceToken(opts.Token), ProfileToken: xonvif.ReferenceToken(opts.Profile)}
	if _, err := sdkptz.Call_SetPreset(context.Background(), dev, req); err != nil {
		log.Error("set preset failed", "err", err)
		os.Exit(1)
	}
	log.Info("preset stored")
}
