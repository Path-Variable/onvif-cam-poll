// Package cli holds the flag groups shared by the commands and the small
// amount of process plumbing (signals, sleeping) they have in common.
package cli

import "time"

// Device identifies and authenticates against one ONVIF device.
// Secrets can be passed through the environment so they stay out of `ps`.
type Device struct {
	Username string `short:"u" long:"user" env:"ONVIF_USER" required:"true" description:"Username for the ONVIF device"`
	Password string `short:"p" long:"password" env:"ONVIF_PASSWORD" required:"true" description:"Password for the ONVIF device (prefer ONVIF_PASSWORD; flags are visible in ps)"`
	Address  string `short:"a" long:"address" env:"ONVIF_ADDRESS" required:"true" description:"host:port of the ONVIF device"`
	Profile  string `short:"r" long:"profile" default:"000" description:"ONVIF media profile token"`
}

// Slack configures where notifications and snapshots go.
type Slack struct {
	ChannelID       string `short:"c" long:"channel-id" env:"SLACK_CHANNEL_ID" required:"true" description:"Slack channel ID for notifications and snapshots"`
	BotToken        string `short:"b" long:"bot-token" env:"SLACK_BOT_TOKEN" required:"true" description:"Slack bot token (prefer SLACK_BOT_TOKEN)"`
	MessageTemplate string `short:"m" long:"message-template" default:"Motion detected at %s" description:"fmt template for the notification; %s is the camera name"`
}

// Preset names a PTZ preset token.
type Preset struct {
	Token string `short:"l" long:"point" default:"001" description:"PTZ preset token"`
}

// Cooldown is the pause after an event, or between repeated commands.
type Cooldown struct {
	Seconds int `short:"t" long:"cooldown" default:"60" description:"Seconds to wait after an event before polling resumes"`
}

// Duration converts the flag to a time.Duration.
func (c Cooldown) Duration() time.Duration { return time.Duration(c.Seconds) * time.Second }

// CameraName labels the device in notifications.
type CameraName struct {
	Name string `short:"n" long:"name" required:"true" description:"Name or location of the device, used in notifications"`
}

// Interface selects the network interface for WS-Discovery.
type Interface struct {
	Name string `short:"i" long:"interface" required:"true" description:"Network interface to probe for ONVIF devices"`
}
