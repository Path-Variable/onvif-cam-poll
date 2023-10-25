package main

import (
	"context"
	"fmt"

	"github.com/isaric/go-posix-time/pkg/p_time"
	"github.com/path-variable/onvif-cam-poll/pkg/model"
	"github.com/path-variable/onvif-cam-poll/pkg/utils"

	"os"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/use-go/onvif"
	"github.com/use-go/onvif/device"
	sdk "github.com/use-go/onvif/sdk/device"
	"github.com/use-go/onvif/xsd"
	onvif2 "github.com/use-go/onvif/xsd/onvif"
)

const commandName = "set-time"
/*
*
Script for setting the date and time on an ONVIF camera- specifically designed for Hisseu cameras
*/
func main() {
	// get and validate number of cli args
	var opts timeOptions
	_, err := flags.ParseArgs(&opts, os.Args)

	if err != nil {
		fmt.Printf(utils.ArgParseError, err)
		return
	}

	// create device and authenticate
	cam, err := onvif.NewDevice(onvif.DeviceParams{Xaddr: opts.Address, Username: opts.Username, Password: opts.Password})
	if err != nil {
		fmt.Printf(utils.ConnectionError, err)
	}

	// repeat call after interval passes
	for {
		ct := time.Now()
		req := getOnvifDateTime(ct)
		fmt.Printf(utils.CommandSend, commandName)
		_, err := sdk.Call_SetSystemDateAndTime(context.TODO(), cam, req)
		if err != nil {
			fmt.Printf(utils.CommandError,commandName, err)
			return
		}
		fmt.Printf(utils.SleepTemplate, opts.CooldownTimer)
		time.Sleep(time.Duration(opts.CooldownTimer) * time.Second)
	}

}

func getOnvifDateTime(ct time.Time) device.SetSystemDateAndTime {
	diff := time.Duration(p_time.GetPosixOffset(ct)) * time.Hour
	ct = ct.Add(diff)

	return device.SetSystemDateAndTime{
		DaylightSavings: xsd.Boolean(ct.IsDST()),
		TimeZone:        onvif2.TimeZone{TZ: xsd.Token(p_time.FormatTimeZone(ct))},
		DateTimeType:    "Manual",
		UTCDateTime: onvif2.DateTime(struct {
			Time onvif2.Time
			Date onvif2.Date
		}{Time: onvif2.Time(struct {
			Hour   xsd.Int
			Minute xsd.Int
			Second xsd.Int
		}{Hour: xsd.Int(ct.Hour()), Minute: xsd.Int(ct.Minute()), Second: xsd.Int(ct.Second())}), Date: onvif2.Date(struct {
			Year  xsd.Int
			Month xsd.Int
			Day   xsd.Int
		}{Year: xsd.Int(ct.Year()), Month: xsd.Int(ct.Month()), Day: xsd.Int(ct.Day())})})}
}

type timeOptions struct {
	model.BasicParameters
	model.CooldownParameters
}
