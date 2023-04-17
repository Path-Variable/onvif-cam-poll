package motion-poll

import (
	"github.com/jessevdk/go-flags"
	"github.com/slack-go/slack"
	"github.com/use-go/onvif"
	"github.com/use-go/onvif/event"
	"github.com/use-go/onvif/media"
	sdk "github.com/use-go/onvif/sdk/media"

	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

/**
Script for polling an ONVIF camera and getting motion events - specifically designed for Hisseu cameras
*/

const ssErrorTemplate = "Error while getting snapshot %s\n"

func main() {
	var opts options
	_, err := flags.ParseArgs(&opts, os.Args)

	if err != nil {
		fmt.Printf("%s Exiting!\n", err)
		return
	}

	// make initial pull point subscription
	cam, _ := onvif.NewDevice(onvif.DeviceParams{Xaddr: opts.Address, Username: opts.Username, Password: opts.Password})
	res := &event.CreatePullPointSubscription{SubscriptionPolicy: event.SubscriptionPolicy{ChangedOnly: true},
		InitialTerminationTime: event.AbsoluteOrRelativeTimeType{
			Duration: "PT300S",
		}}
	_, err = cam.CallMethod(res)
	if err != nil {
		fmt.Printf("Aborting due to err when subscribing %s", err)
		return
	}

	// retrieve the snapshot url
	ssur, err := sdk.Call_GetSnapshotUri(nil, cam, media.GetSnapshotUri{ProfileToken: "000"})
	ssUrl := string(ssur.MediaUri.Uri)
	fmt.Printf("Snapshot url is %s\n", ssUrl)

	slackClient := slack.New(opts.SlackBotToken)

	// continue polling for motion events. if motion is detected, send Slack notification
	for true {
		r2, _ := cam.CallMethod(event.PullMessages{})
		bodyBytes, _ := io.ReadAll(r2.Body)
		bodyS := string(bodyBytes)
		if strings.Contains(bodyS, "<tt:SimpleItem Name=\"IsMotion\" Value=\"true\" />") {
			msg := fmt.Sprintf(opts.MessageTemplate, opts.CameraName)
			err = slack.PostWebhook(opts.SlackHook, &slack.WebhookMessage{Text: msg})
			if err != nil {
				fmt.Printf("there was an error while posting the slack notification %s", err)
			}
			if ssUrl != "" && opts.SlackBotToken != "token" && opts.SlackChannelID != "" {
				getAndUploadSnapshot(ssUrl, opts.SlackChannelID, *slackClient)
			}
			time.Sleep(time.Duration(opts.CooldownTimer) * time.Second)
		}
		time.Sleep(1 * time.Second)
	}

}

func getAndUploadSnapshot(url, channelID string, slackClient slack.Client) {
	r, e := http.Get(url)
	if e != nil {
		fmt.Printf(ssErrorTemplate, e)
		return
	}
	defer r.Body.Close()
	if e != nil {
		fmt.Printf(ssErrorTemplate, e)
		return
	}

	_, err := slackClient.UploadFile(slack.FileUploadParameters{
		Reader:   r.Body,
		Filetype: "image/png",
		Filename: fmt.Sprintf("%s.png", time.Now().Format("20060102150405")),
		Channels: []string{channelID},
	})

	if err != nil {
		fmt.Printf("error while posting snapshot %s", err)
	}
}

type options struct {
	Username        string `short:"u" long:"user" description:"The username for authenticating to the ONVIF device" required:"true"`
	Password        string `short:"p" long:"password" description:"The password for authenticating to the ONVIF device" required:"true"`
	Address         string `short:"a" long:"address" description:"The address of the ONVIF device and its port separated by semicolon" required:"true"`
	CameraName      string `short:"n" long:"name" description:"The name or location of the ONVIF device that will appear in all notifications" required:"true"`
	SlackHook       string `short:"s" long:"slack-hook" description:"The address of the slack hook to which notifications will be posted" required:"true"`
	CooldownTimer   int    `short:"t" long:"cooldown" description:"The integer value of the number of seconds after an event has occurred before polling resumes" required:"false" default:"10"`
	SlackChannelID  string `short:"c" long:"channel-id" description:"The ID of the slack channel where snapshots will be posted if provided" required:"false"`
	SlackBotToken   string `short:"b" long:"bot-token" description:"The token for the slack bot that will upload a snapshot if provided" required:"false" default:"token"`
	MessageTemplate string `short:"m" long:"message-template" description:"The message template in JSON format to use for notifications instead of the default one" required:"false" default:"Motion detected at %s"`
}
