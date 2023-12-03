package main

import (
	"context"
	"fmt"
	"os"

	"github.com/path-variable/onvif-cam-poll/pkg/model"
	"github.com/path-variable/onvif-cam-poll/pkg/utils"

	"github.com/jessevdk/go-flags"
	"github.com/use-go/onvif"
	"github.com/use-go/onvif/ptz"
	sdk_ptz "github.com/use-go/onvif/sdk/ptz"
	token "github.com/use-go/onvif/xsd/onvif"
)

const commandName = "onvif-set-preset"

/**
Records the current camera position on the passed preset token
*/

func main() {
	var opts setPresetOptions
	_, err := flags.ParseArgs(&opts, os.Args)

	if err != nil {
		fmt.Printf(utils.ArgParseError, err)
		return
	}

	cam, err := onvif.NewDevice(onvif.DeviceParams{Xaddr: opts.Address, Username: opts.Username, Password: opts.Password})
	if err != nil {
		fmt.Printf(utils.ConnectionError, err)
		return
	}
	gtreq := ptz.SetPreset{
		PresetToken:  token.ReferenceToken(opts.PositionPreset),
		ProfileToken: token.ReferenceToken(opts.Profile),
	}
	fmt.Printf(utils.CommandSend, commandName)
	_, err = sdk_ptz.Call_SetPreset(context.TODO(), cam, gtreq)
	if err != nil {
		fmt.Printf(utils.CommandError, commandName, err)
		return
	}

}

type setPresetOptions struct {
	model.BasicParameters
	model.PresetParameters
}
