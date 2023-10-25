package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/path-variable/onvif-cam-poll/pkg/model"
	"github.com/path-variable/onvif-cam-poll/pkg/utils"

	"github.com/jessevdk/go-flags"
	"github.com/use-go/onvif"
	"github.com/use-go/onvif/ptz"
	sdk_ptz "github.com/use-go/onvif/sdk/ptz"
	token "github.com/use-go/onvif/xsd/onvif"
)

const commandName = "goto-preset"

/**
Sends the camera to the target preset
*/

func main() {
	var opts gotoPresetOptions
	_, err := flags.ParseArgs(&opts, os.Args)

	if err != nil {
		fmt.Printf(utils.ArgParseError, err)
		return
	}

	cam, _ := onvif.NewDevice(onvif.DeviceParams{Xaddr: opts.Address, Username: opts.Username, Password: opts.Password})
	fmt.Printf(utils.ConnectionOK, opts.Address)
	for {
		gtreq := ptz.GotoPreset{
			PresetToken:  token.ReferenceToken(opts.PositionPreset),
			ProfileToken: token.ReferenceToken(opts.Profile),
		}
        fmt.Printf(utils.CommandSend, commandName)
		_, err := sdk_ptz.Call_GotoPreset(context.TODO(), cam, gtreq)
		if err != nil {
			fmt.Printf(utils.CommandError,commandName, err)
			return
		}
		fmt.Printf(utils.SleepTemplate, opts.CooldownTimer)
		time.Sleep(time.Duration(opts.CooldownTimer) * time.Second)
	}

}

type gotoPresetOptions struct {
	model.BasicParameters
	model.CooldownParameters
	model.PresetParameters
}
