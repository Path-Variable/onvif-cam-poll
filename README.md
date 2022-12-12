# onvif-cam-poll

[![Go](https://github.com/isaric/onvif-cam-poll/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/isaric/onvif-cam-poll/actions/workflows/go.yml)

Two simple golang scripts that provide ONVIF motion event polling and time/date setting. The scripts were created to allow 
the use of IP cameras without using a cloud service as a middle man between the user and the camera(s).

## motion-poll

The ONVIF standard exposed a set of services on a camera. One of them is the events service. Cameras with motion detection
capabilities publish motion events over this service. There exist two models with which to get events from an ONVIF camera.
One is polling - the client calls the device repeatedly and asks if any new events were created since the last call. The other
is a subscription model where the client leaves a callback url and the camera uses this url to post notifications when a new 
event is created.

Polling is the most common feature implemented, and it is the one that is utilized here. The cameras I have used in the making 
of this script don't support subscription.

In order to use the polling script you must provide the mandatory arguments or the script will fail immediately. These include the
base url of the camera, auth details and a slack webhook to which the notification will be posted.

It is also possible to provide the cooldown time and give the Slack channel and bot token. The script will then
grab a snapshot from the camera and upload it to slack. The bot must have file upload privileges. There are more
details in this guide [here](https://api.slack.com/methods/files.upload).

The script was written in Go and therefore must be compiled into a native executable before use like so:

    go build motion-poll.go
    
After that we can use the compiled native executable. Example:

    ./motion-poll -a my-camera-url:1234 -u admin -p nimda -n garden -s https://hooks.slack.com/my-fake-webook -t 30 -b xoxb-slack-bot-token -c slack-channel-id
    
With this command the script will keep running continuously and polling the camera. After it finds a motion event,
it will post to the slack url specified and identify the camera as "garden".

## set-time

The second script deals with another problem when opting to use these cameras without a cloud provider - time synchronization.
If cut off from access to the cloud, the cameras' system clock quickly falls out of sync with the world clock and is unable to 
contact an NTP server in order to correct itself. This script will continue updating the system clock with the local time of 
the server in which the script is running in regular intervals. Even with the low-cost nature of the clocks in these cameras, it is 
enough to maintain a sufficiently low drift for most surveillance purposes.

It requires four arguments for usage - the url and port (one string, seperated by semicolon), the username, the password and the
interval time expressed in minutes as an integer. After interval expires, the time is set again.
It must also be compiled beforehand.

    go build set-time.go
    

Example:

    ./set-time -a my-camera-url:1234 -u admin -p nimda -i 1
    
As with the previous script, this one keeps running and will post the current local time of the server to the camera, thereby synchronizing
its system clock with the servers.
