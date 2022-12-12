package main

import (
	"fmt"
	"github.com/isaric/go-posix-time/pkg/p_time"
	"github.com/jessevdk/go-flags"
	"github.com/use-go/onvif"
	"github.com/use-go/onvif/device"
	"github.com/use-go/onvif/xsd"
	onvif2 "github.com/use-go/onvif/xsd/onvif"
	"os"
	"time"
)

/*
*
Script for setting the date and time on an ONVIF camera- specifically designed for Hisseu cameras
*/
func main() {
	// get and validate number of cli args
	var opts timeOptions
	_, err := flags.ParseArgs(&opts, os.Args)

	if err != nil {
		println("Invalid command line arguments! Exiting!")
		return
	}

	// create device and authenticate
	cam, _ := onvif.NewDevice(opts.Host)
	cam.Authenticate(opts.Username, opts.Password)

	// repeat call after interval passes
	for true {
		ct := time.Now()
		req := getOnvifDateTime(ct)
		_, err := cam.CallMethod(req)
		if err != nil {
			fmt.Printf("Could not set time. %s Exiting!\n", err)
			return
		}
		time.Sleep(time.Duration(opts.Interval) * time.Minute)
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
	Username string `short:"u" long:"user" description:"The username for authenticating to the ONVIF device" required:"true"`
	Password string `short:"p" long:"password" description:"The password for authenticating to the ONVIF device" required:"true"`
	Host     string `short:"h" long:"host" description:"The address of the ONVIF device and its port separated by semicolon" required:"true"`
	Interval int    `short:"i" long:"interval" description:"The integer representing the number of minutes to pass between polls" required:"false" default:"10"`
}
