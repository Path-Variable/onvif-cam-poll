package model

type BasicParameters struct {
	Username string `short:"u" long:"user" description:"The username for authenticating to the ONVIF device" required:"true"`
	Password string `short:"p" long:"password" description:"The password for authenticating to the ONVIF device" required:"true"`
	Address  string `short:"a" long:"address" description:"The address of the ONVIF device and its port separated by semicolon" required:"true"`
	Profile  string `short:"r" long:"profile" description:"The onvif profile to be used" required:"false" default:"000"`
}

type SlackParameters struct {
	SlackChannelID  string `short:"c" long:"channel-id" description:"The ID of the slack channel where snapshots will be posted if provided" required:"true"`
	SlackBotToken   string `short:"b" long:"bot-token" description:"The token for the slack bot that will upload a snapshot if provided" required:"true"`
	MessageTemplate string `short:"m" long:"message-template" description:"The message template in JSON format to use for notifications instead of the default one" required:"false" default:"Motion detected at %s"`
}

type PresetParameters struct {
	PositionPreset string `short:"l" long:"point" description:"The ptz preset the camera should move to" required:"false" default:"001"`
}

type CooldownParameters struct {
	CooldownTimer int `short:"t" long:"cooldown" description:"The integer value of the number of seconds after an event has occurred before polling resumes" required:"false" default:"60"`
}

type CameraNameParameters struct {
	CameraName string `short:"n" long:"name" description:"The name or location of the ONVIF device that will appear in all notifications" required:"true"`
}

type InterfaceParameters struct {
	Interface string `short:"i" long:"interface" description:"Prints a list of all discovered ONVIF devices on the specified interface" required:"true"`
}
