# Onvif-Cam-Poll

[![Go](https://github.com/Path-Variable/onvif-cam-poll/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/Path-Variable/onvif-cam-poll/actions/workflows/go.yml)

Five small Go programs for running ONVIF IP cameras without a vendor cloud in
the middle: motion alerts to Slack with a snapshot, clock synchronisation for
cameras that cannot reach NTP, PTZ preset handling and device discovery.

They are built to run unattended. `onvif-motion-poll` keeps its event
subscription alive, re-subscribes with exponential backoff when the camera or
the network misbehaves, and stops cleanly on SIGINT/SIGTERM.

## Install

    go install ./...

The binaries land in `$(go env GOPATH)/bin`. `make install` additionally
creates symlinks and systemd unit templates (see `install_scripts/`).

## Credentials

Every flag that carries a secret can also come from the environment, which
keeps it out of `ps` and shell history:

| Flag | Environment variable |
|---|---|
| `-u, --user` | `ONVIF_USER` |
| `-p, --password` | `ONVIF_PASSWORD` |
| `-a, --address` | `ONVIF_ADDRESS` |
| `-b, --bot-token` | `SLACK_BOT_TOKEN` |
| `-c, --channel-id` | `SLACK_CHANNEL_ID` |

Run any command with `--help` for the full list.

## Commands

### onvif-motion-poll

The ONVIF standard exposed a set of services on a camera. One of them is the events service. Cameras with motion detection
capabilities publish motion events over this service. There exist two models with which to get events from an ONVIF camera.
One is polling - the client calls the device repeatedly and asks if any new events were created since the last call. The other
is a subscription model where the client leaves a callback url and the camera uses this url to post notifications when a new 
event is created.

Polling is the most common feature implemented, and it is the one that is utilized here. The cameras I have used in the making 
of this script don't support subscription.

In order to use the polling script you must provide the mandatory arguments or the script will fail immediately. These 
include the base url of the camera, auth details and a slack configuration for the bot we will use to post notifications.

On every motion event the daemon posts the message and then fetches a snapshot from the camera and uploads it to the
channel (the bot needs `files:write`). A snapshot that is not an image, for example a login page, is rejected rather
than uploaded. Snapshot failures are logged and never stop the alert.

    
After that we can use the compiled native executable. Example:

    onvif-motion-poll -a my-camera-url:1234 -u admin -p nimda -n garden -t 30 -b xoxb-slack-bot-token -c slack-channel-id
    
With this command the script will keep running continuously and polling the camera. After it finds a motion event,
it will post to the slack url specified and identify the camera as "garden".

### onvif-set-time

The second script deals with another problem when opting to use these cameras without a cloud provider - time synchronization.
If cut off from access to the cloud, the cameras' system clock quickly falls out of sync with the world clock and is unable to 
contact an NTP server in order to correct itself. This script will continue updating the system clock with the local time of 
the server in which the script is running in regular intervals. Even with the low-cost nature of the clocks in these cameras, it is 
enough to maintain a sufficiently low drift for most surveillance purposes.

It requires four arguments for usage - the url and port (one string, seperated by semicolon), the username, the password and the
interval time expressed in minutes as an integer. After interval expires, the time is set again.
    
Example:

    onvif-set-time -a my-camera-url:1234 -u admin -p nimda -t 1
    
As with the previous script, this one keeps running and will post the current local time of the server to the camera, 
thereby synchronizing its system clock with the servers.

### onvif-goto-preset

The third script will move the camera to a predefined ptz preset. The purpose of this script is to counter tampering and 
accidental camera misalignment due to power outages. Besides the authentication details, this script accepts the ptz preset 
name and the onvif profile name.

Example:

    onvif-goto-preset -a my-camera-url:1234 -u admin -p nimda -t 1 -l 001 -r 000

### onvif-set-preset

The fourth script will record a ptz preset for the current position of the camera. It requires the authentication details, 
the preset name and, optionally, the onvif profile name.

Example:

    onvif-set-preset -a my-camera-url:1234 -u admin -p nimda -l 003 -r 000

### onvif-discover-all

This script will run an onvif discovery broadcast on the interface specifed in the entry parameter. The interface must be
specified by name. You cannot use a CIDR address to define it. If you are running this over Wireguard, it won't work as 
Wireguard does not support broadcasting.

Example:

    onvif-discover-all -i wlo1

The command will output the number of discovered devices along with their IP addresses.

## Service Templates
Please check the services folder for examples of what a systemd service template should look like if you choose to run 
motion-poll or set-time as a service. By using environment files and templating, the user can handle multiple cameras 
with only one service file.
For example, I will define my environment file `hallway.env` and place it in the configuration folder.
In order to run `motion-poll` as a service I will execute the following command:

    systemctl start motion@hallway

This will use the `motion@.service` file and take all the relevant variables from the `hallway.env` file.
If I want the service to restart when the server restarts I need to enable it.

    systemctl enable motion@hallway

The commands are the same as when running any other systemd service.

## System-wide Install
A Makefile is provided in the project root that will make a system-wide install of the go binaries once compiled. It 
will then use the provided service templates to create systemd services for every compiled command and place it in a 
config folder that the user can specify.

The user can then place `.env` files into the config folder and run `systemctl` to start any of the command services with
the name of the environment file. Be sure to enable any such services if you want them to run on restart!

## Behaviour under failure

* `onvif-motion-poll` creates a pull-point subscription with a five-minute
  lifetime and renews it 30 seconds before expiry. A failed pull or subscribe
  is retried with exponential backoff (2 s doubling to 2 min); the process
  never exits on its own.
* The looping commands (`goto-preset`, `set-time`, `discover-all`) log a
  failed call and try again after the cooldown instead of exiting.
* `onvif-set-time` sends the clock in UTC together with a POSIX TZ string
  (`CET-1CEST,M3.5.0,M10.5.0/3`) so the camera renders local time and
  follows daylight-saving changes itself.
* All output is structured logging on stderr (`log/slog`).

## Development

    go vet ./... && go test -race ./...
    go run honnef.co/go/tools/cmd/staticcheck@latest ./...

Layout: one `main` per command under `cmd/`, shared code under `internal/`:
`cli` (flags, signals), `camera` (device calls), `motion` (event parsing and
the poll loop), `notify` (Slack and snapshot fetching), `systime` (clock
request), `backoff`. The poll loop is tested against a fake subscriber; the
ONVIF calls themselves need a real camera.
