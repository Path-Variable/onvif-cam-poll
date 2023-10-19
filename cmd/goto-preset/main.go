package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/use-go/onvif"
	"github.com/use-go/onvif/ptz"
	sdk_ptz "github.com/use-go/onvif/sdk/ptz"
	"os"
	"time"
)

/**
Sends the camera to the target preset
*/

func main() {
	var opts gotoPresetOptions
	_, err := flags.ParseArgs(&opts, os.Args)

	if err != nil {
		fmt.Printf("%s Exiting!\n", err)
		return
	}

	cam, _ := onvif.NewDevice(onvif.DeviceParams{Xaddr: opts.Address, Username: opts.Username, Password: opts.Password})
	for true {
		gtreq := ptz.GotoPreset{
			PresetToken:  "001",
			ProfileToken: "000",
		}

		_, err := sdk_ptz.Call_GotoPreset(nil, cam, gtreq)
		if err != nil {
			fmt.Printf("Error while sending go to preset command %s\nExiting!\n", err)
			return
		}
		time.Sleep(time.Duration(opts.CooldownTimer) * time.Minute)
	}

}

type gotoPresetOptions struct {
	Username       string `short:"u" long:"user" description:"The username for authenticating to the ONVIF device" required:"true"`
	Password       string `short:"p" long:"password" description:"The password for authenticating to the ONVIF device" required:"true"`
	Address        string `short:"a" long:"address" description:"The address of the ONVIF device and its port separated by semicolon" required:"true"`
	PresetPosition string `short:"l" long:"point" description:"The ptz preset the camera should move to" required:"true" default:"001"`
	CooldownTimer  int    `short:"t" long:"cooldown" description:"The integer value of the number of seconds after an event has occurred before polling resumes" required:"false" default:"10"`
}
