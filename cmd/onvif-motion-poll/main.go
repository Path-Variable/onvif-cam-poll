// Command onvif-motion-poll watches an ONVIF camera for motion events and
// posts a Slack message plus a snapshot whenever one occurs. It runs until
// SIGINT/SIGTERM and survives camera and network hiccups by re-subscribing.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/path-variable/onvif-cam-poll/internal/backoff"
	"github.com/path-variable/onvif-cam-poll/internal/camera"
	"github.com/path-variable/onvif-cam-poll/internal/cli"
	"github.com/path-variable/onvif-cam-poll/internal/motion"
	"github.com/path-variable/onvif-cam-poll/internal/notify"
)

const (
	pollInterval    = time.Second
	subscriptionTTL = 5 * time.Minute
	renewMargin     = 30 * time.Second
	pullTimeout     = time.Second
	messageLimit    = 10
	retryMin        = 2 * time.Second
	retryMax        = 2 * time.Minute
)

type options struct {
	cli.Device
	cli.CameraName
	cli.Cooldown
	cli.Slack
}

func main() {
	var opts options
	if err := cli.Parse(&opts); err != nil {
		os.Exit(2)
	}
	log := cli.Logger().With("camera", opts.Name)
	ctx, stop := cli.Context()
	defer stop()

	dev, err := camera.Connect(opts.Device)
	if err != nil {
		log.Error("cannot connect", "err", err)
		os.Exit(1)
	}
	snapshotURL, err := camera.SnapshotURI(ctx, dev, opts.Profile)
	if err != nil {
		log.Warn("snapshots disabled", "err", err)
	}
	slack := notify.NewSlack(opts.BotToken, opts.ChannelID, log)

	poller := &motion.Poller{
		Sub:             &motion.PullPoint{Dev: dev, TTL: subscriptionTTL, PullTimeout: pullTimeout, MessageLimit: messageLimit},
		OnMotion:        func(ctx context.Context) { notifyMotion(ctx, slack, opts, snapshotURL, log) },
		PollInterval:    pollInterval,
		Cooldown:        opts.Cooldown.Duration(),
		SubscriptionTTL: subscriptionTTL - renewMargin,
		Backoff:         backoff.New(retryMin, retryMax),
		Log:             log,
	}
	log.Info("polling for motion", "address", opts.Address, "cooldown", opts.Cooldown.Duration())
	err = poller.Run(ctx)
	log.Info("stopped", "reason", err)
}

func notifyMotion(ctx context.Context, slack *notify.Slack, opts options, snapshotURL string, log interface{ Warn(string, ...any) }) {
	if err := slack.Post(ctx, fmt.Sprintf(opts.MessageTemplate, opts.Name)); err != nil {
		log.Warn("notification failed", "err", err)
	}
	if snapshotURL == "" {
		return
	}
	if err := slack.UploadSnapshot(ctx, snapshotURL); err != nil {
		log.Warn("snapshot failed", "err", err)
	}
}
