package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/use-go/onvif"
	"github.com/use-go/onvif/ptz"
	sdk_ptz "github.com/use-go/onvif/sdk/ptz"
	"os"
)

/**
Records the current camera position on the passed preset token
*/

func main() {
	var opts setPresetOptions
	_, err := flags.ParseArgs(&opts, os.Args)

	if err != nil {
		fmt.Printf("%s Exiting!\n", err)
		return
	}

	cam, _ := onvif.NewDevice(onvif.DeviceParams{Xaddr: opts.Address, Username: opts.Username, Password: opts.Password})
	gtreq := ptz.SetPreset{
		PresetToken:  "001",
		ProfileToken: "000",
	}

	_, err = sdk_ptz.Call_SetPreset(nil, cam, gtreq)
	if err != nil {
		fmt.Printf("Error while sending set preset command %s\nExiting!\n", err)
		return
	}

}

type setPresetOptions struct {
	Username       string `short:"u" long:"user" description:"The username for authenticating to the ONVIF device" required:"true"`
	Password       string `short:"p" long:"password" description:"The password for authenticating to the ONVIF device" required:"true"`
	Address        string `short:"a" long:"address" description:"The address of the ONVIF device and its port separated by semicolon" required:"true"`
	PresetPosition string `short:"l" long:"point" description:"The ptz preset the camera should move to" required:"false" default:"001"`
}
