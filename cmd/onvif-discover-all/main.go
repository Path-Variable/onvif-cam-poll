package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/path-variable/onvif-cam-poll/pkg/model"
	"github.com/path-variable/onvif-cam-poll/pkg/utils"
	"github.com/use-go/onvif"
	"os"
	"strings"
	"time"
)

const commandName = "onvif-discover-all"

func main() {
	var opts discoverAllOptions
	_, err := flags.ParseArgs(&opts, os.Args)

	if err != nil {
		fmt.Printf(utils.ArgParseError, err)
		return
	}

	for {
		fmt.Printf(utils.CommandSend, commandName)
		res, err := onvif.GetAvailableDevicesAtSpecificEthernetInterface(opts.Interface)
		if err != nil {
			fmt.Printf(utils.CommandError, commandName, err)
			return
		}

		fmt.Printf("Discovered %d devices on interface %s\n", len(res), opts.Interface)

		for i := 0; i < len(res); i++ {
			dev := res[i]
			fmt.Printf("Device at %s\n", getAddressFromServices(dev))
		}

		fmt.Printf(utils.SleepTemplate, opts.CooldownTimer)
		time.Sleep(time.Duration(opts.CooldownTimer) * time.Second)
	}

}

func getAddressFromServices(device onvif.Device) string {
	val, found := device.GetServices()["device"]
	if found {
		return getAddressFromUrl(val)
	}
	return ""
}

func getAddressFromUrl(url string) string {
	return strings.Split(url, ":")[1][2:]
}

type discoverAllOptions struct {
	model.CooldownParameters
	model.InterfaceParameters
}
