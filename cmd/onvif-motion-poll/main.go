package main

import (
	"context"

	"github.com/jessevdk/go-flags"
	"github.com/path-variable/onvif-cam-poll/pkg/model"
	"github.com/path-variable/onvif-cam-poll/pkg/utils"
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
const commandName = "onvif-motion-poll"

func main() {
	var opts options
	_, err := flags.ParseArgs(&opts, os.Args)

	if err != nil {
		fmt.Printf(utils.ArgParseError, err)
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
	ssur, _ := sdk.Call_GetSnapshotUri(context.TODO(), cam, media.GetSnapshotUri{ProfileToken: "000"})
	ssUrl := string(ssur.MediaUri.Uri)
	fmt.Printf("Snapshot url is %s\n", ssUrl)

	slackClient := slack.New(opts.SlackBotToken)

	// continue polling for motion events. if motion is detected, send Slack notification
	for {
		fmt.Printf(utils.CommandSend, commandName)
		r2, err := cam.CallMethod(event.PullMessages{})
		if err != nil {
			fmt.Printf(utils.CommandError, commandName, err)
			return
		}
		bodyBytes, _ := io.ReadAll(r2.Body)
		bodyS := string(bodyBytes)
		if strings.Contains(bodyS, "<tt:SimpleItem Name=\"IsMotion\" Value=\"true\" />") {
			msg := fmt.Sprintf(opts.MessageTemplate, opts.CameraName)
			_, _, err = slackClient.PostMessage(opts.SlackChannelID, slack.MsgOptionText(msg, false))
			if err != nil {
				fmt.Printf("there was an error while posting the slack notification %s", err)
			}
			if ssUrl != "" && opts.SlackBotToken != "token" && opts.SlackChannelID != "" {
				getAndUploadSnapshot(ssUrl, opts.SlackChannelID, *slackClient)
			}
			fmt.Printf(utils.SleepTemplate, opts.CooldownTimer)
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
	model.BasicParameters
	model.CameraNameParameters
	model.CooldownParameters
	model.SlackParameters
}
