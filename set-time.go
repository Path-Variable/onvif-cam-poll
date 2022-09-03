package main

import (
	"fmt"
	"github.com/isaric/go-posix-time/pkg/p_time"
	"github.com/use-go/onvif"
	"github.com/use-go/onvif/device"
	"github.com/use-go/onvif/xsd"
	onvif2 "github.com/use-go/onvif/xsd/onvif"
	"os"
	"strconv"
	"time"
)

/*
*
Script for setting the date and time on an ONVIF camera- specifically designed for Hisseu cameras
CLI args:

	0 - url and port, separated by semicolon
	1 - username
	2 - password
*/
func main() {
	// get and validate number of cli args
	args := os.Args[1:]
	if len(args) < 3 {
		fmt.Println("Not enough arguments given. There must be at least 3! Exiting!")
		return
	}

	// create device and authenticate
	cam, _ := onvif.NewDevice(args[0])
	cam.Authenticate(args[1], args[2])

	var interval = 30
	if len(args) > 4 {
		convInt, err := strconv.Atoi(args[4])
		if err == nil {
			interval = convInt
		}
	}

	// repeat call after interval passes
	for true {
		ct := time.Now()
		req := getOnvifDateTime(ct)
		_, err := cam.CallMethod(req)
		if err != nil {
			fmt.Printf("Could not set time. %s Exiting!\n", err)
			return
		}
		time.Sleep(time.Duration(interval) * time.Minute)
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
