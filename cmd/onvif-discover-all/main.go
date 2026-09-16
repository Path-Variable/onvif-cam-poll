// Command onvif-discover-all lists the ONVIF devices answering WS-Discovery
// probes on a network interface, repeating after every cooldown.
package main

import (
	"os"

	"github.com/path-variable/onvif-cam-poll/internal/camera"
	"github.com/path-variable/onvif-cam-poll/internal/cli"
)

type options struct {
	cli.Cooldown
	cli.Interface
}

func main() {
	var opts options
	if err := cli.Parse(&opts); err != nil {
		os.Exit(2)
	}
	log := cli.Logger()
	ctx, stop := cli.Context()
	defer stop()

	for {
		hosts, err := camera.Discover(opts.Interface.Name)
		if err != nil {
			log.Warn("discovery failed", "err", err)
		} else {
			log.Info("discovery finished", "interface", opts.Interface.Name, "devices", len(hosts))
			for _, host := range hosts {
				log.Info("device", "address", host)
			}
		}
		if !cli.Sleep(ctx, opts.Cooldown.Duration()) {
			return
		}
	}
}
