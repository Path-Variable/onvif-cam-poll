package main

import (
	"bytes"
	"fmt"
	"github.com/use-go/onvif"
	"github.com/use-go/onvif/event"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)


/**
 Script for polling an ONVIF camera and getting motion events - specifically designed for Hisseu cameras
 CLI args:
     0 - url and port, separated by semicolon
     1 - username
     2 - password
     3 - camera name (or location)
     4 - slack hook url
     5 - (optional) cooldown time after motion event detected
     6 - (optional) json message template for sprintf ex. ./motion-poll "${<file.json}"
 */
func main() {
	// get and validate number of cli args
	args := os.Args[1:]
	if len(args) < 5 {
		fmt.Println("Not enough arguments given. There must be at least 5! Exiting!")
		return
	}

	// make initial pull point subscription
	cam, _ := onvif.NewDevice(args[0])
	cam.Authenticate(args[1], args[2])
	res2 := &event.CreatePullPointSubscription{ SubscriptionPolicy: event.SubscriptionPolicy{ChangedOnly: true},
		InitialTerminationTime: event.AbsoluteOrRelativeTimeType {
		Duration: "PT300S",
	}}
	_, err := cam.CallMethod(res2)
	if err != nil {
		fmt.Printf("Aborting due to err when subscribing %s", err)
		return
	}

	// get slack message from template - use default if cli arg is not given
	camName := args[3]
	var msgT string
	if len(args) > 5 {
		msgT = args[6]
	} else {
		msgT = ` 
    	{
			"text" : "Motion detected at %s"
        }
    `
	}

	//get cooldown time from args, default 10 seconds
	cooldown := 10
	if len(args) > 5 {
		convInt, err := strconv.Atoi(args[5])
		if err == nil {
			cooldown = convInt
		} else {
			fmt.Printf("Could not parse cooldown time %s with error %s Defaulting to: %d\n", args[5], err, cooldown)
		}
	}


	// continue polling for motion events. if motion is detected, send slack notification
	for true {
		r2, _ := cam.CallMethod(event.PullMessages{})
		bodyBytes, _ := ioutil.ReadAll(r2.Body)
		bodyS := string(bodyBytes)
		if strings.Contains(bodyS, "<tt:SimpleItem Name=\"IsMotion\" Value=\"true\" />") {
			fmt.Printf("%s", bodyS)
			msg := fmt.Sprintf(msgT, camName)
			_, err = http.Post(args[4], "aplication/json", bytes.NewReader([]byte(msg)))
			if err != nil {
				fmt.Printf("there was an error while posting the slack notification %s", err)
			}
			time.Sleep(time.Duration(cooldown) * time.Second)
		}
		time.Sleep(1 * time.Second)
	}

}