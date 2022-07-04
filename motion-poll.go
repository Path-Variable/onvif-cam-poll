package main

import (
	"bytes"
	"fmt"
	"github.com/use-go/onvif"
	"github.com/use-go/onvif/event"
	"gopkg.in/xmlpath.v2"
	"io"
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
	 5 - snapshot save path
     6 - (optional) cooldown time after motion event detected
     7 - (optional) json message template for sprintf ex. ./motion-poll "${<file.json}"
*/

const ssErrorTemplate = "Error while getting snapshot %s\n"

const soap = `
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope"
xmlns:trt="http://www.onvif.org/ver10/media/wsdl"
xmlns:tt="http://www.onvif.org/ver10/schema">
  <soap:Body>
    <trt:GetSnapshotUri >     
      <trt:ProfileToken>%s</trt:ProfileToken>
    </trt:GetSnapshotUri>
  </soap:Body>
</soap:Envelope>
`

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
	res := &event.CreatePullPointSubscription{SubscriptionPolicy: event.SubscriptionPolicy{ChangedOnly: true},
		InitialTerminationTime: event.AbsoluteOrRelativeTimeType{
			Duration: "PT300S",
		}}
	_, err := cam.CallMethod(res)
	if err != nil {
		fmt.Printf("Aborting due to err when subscribing %s", err)
		return
	}

	// get slack message from template - use default if cli arg is not given
	camName := args[3]
	var msgT string
	if len(args) > 7 {
		msgT = args[7]
	} else {
		msgT = ` 
    	{
			"text" : "Motion detected at %s"
        }
    `
	}

	//get cooldown time from args, default 10 seconds
	cooldown := 10
	if len(args) > 6 {
		convInt, err := strconv.Atoi(args[6])
		if err == nil {
			cooldown = convInt
		} else {
			fmt.Printf("Could not parse cooldown time %s with error %s Defaulting to: %d\n", args[5], err, cooldown)
		}
	}

	r, err := http.Post(fmt.Sprintf("http://%s", args[0]), "application/soap+xml", strings.NewReader(fmt.Sprintf(soap, "000")))
	ssUrl := ""
	path := xmlpath.MustCompile("//*/Uri")
	if err == nil {
		data, _ := ioutil.ReadAll(r.Body)
		root, _ := xmlpath.Parse(strings.NewReader(string(data)))
		if err == nil {
			ssUrl, _ = path.String(root)
		}
	}
	fmt.Printf("Snapshot url is %s\n", ssUrl)

	// continue polling for motion events. if motion is detected, send slack notification
	for true {
		r2, _ := cam.CallMethod(event.PullMessages{})
		bodyBytes, _ := ioutil.ReadAll(r2.Body)
		bodyS := string(bodyBytes)
		if strings.Contains(bodyS, "<tt:SimpleItem Name=\"IsMotion\" Value=\"true\" />") {
			msg := fmt.Sprintf(msgT, camName)
			_, err = http.Post(args[4], "aplication/json", bytes.NewReader([]byte(msg)))
			if err != nil {
				fmt.Printf("there was an error while posting the slack notification %s", err)
			}
			if ssUrl != "" {
				getSnapshot(ssUrl, args[5])
			}

			time.Sleep(time.Duration(cooldown) * time.Second)
		}
		time.Sleep(1 * time.Second)
	}

}

func getSnapshot(url, path string) {
	r, e := http.Get(url)
	if e != nil {
		fmt.Printf(ssErrorTemplate, e)
		return
	}
	defer r.Body.Close()

	file, e := os.Create(fmt.Sprintf("%s/%s.jpeg", path, time.Now().Format("20060102150405")))
	if e != nil {
		fmt.Printf(ssErrorTemplate, e)
		return
	}
	defer file.Close()

	_, e = io.Copy(file, r.Body)
	if e != nil {
		fmt.Printf(ssErrorTemplate, e)
	}
}
