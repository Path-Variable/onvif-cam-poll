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
	fmt.Printf("Connected to device %s\n", opts.Address)
	for true {
		gtreq := ptz.GetPresets{
			ProfileToken: "001",
		}
        println("Sending get presets command")
		r, err := sdk_ptz.Call_GetPresets(nil, cam, gtreq)
		if err != nil {
			fmt.Printf("Error while sending go to preset command %s\nExiting!\n", err)
			return
		}
		fmt.Printf("%v", r)
		fmt.Printf("Sleeping for %d minutes\n", opts.CooldownTimer)
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
