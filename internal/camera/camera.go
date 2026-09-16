// Package camera wraps the use-go/onvif device calls the commands need.
package camera

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/use-go/onvif"
	"github.com/use-go/onvif/media"
	sdkmedia "github.com/use-go/onvif/sdk/media"
	xonvif "github.com/use-go/onvif/xsd/onvif"

	"github.com/path-variable/onvif-cam-poll/internal/cli"
)

// Connect authenticates against the device and fetches its capabilities.
func Connect(d cli.Device) (*onvif.Device, error) {
	dev, err := onvif.NewDevice(onvif.DeviceParams{Xaddr: d.Address, Username: d.Username, Password: d.Password})
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", d.Address, err)
	}
	return dev, nil
}

// SnapshotURI asks the device where a JPEG snapshot for the profile can be fetched.
func SnapshotURI(ctx context.Context, dev *onvif.Device, profile string) (string, error) {
	res, err := sdkmedia.Call_GetSnapshotUri(ctx, dev, media.GetSnapshotUri{ProfileToken: xonvif.ReferenceToken(profile)})
	if err != nil {
		return "", fmt.Errorf("get snapshot uri for profile %q: %w", profile, err)
	}
	uri := string(res.MediaUri.Uri)
	if uri == "" {
		return "", errors.New("device returned an empty snapshot uri")
	}
	return uri, nil
}

// Discover probes the interface with WS-Discovery and returns the host:port
// of every device that answered with a parseable device-service address.
func Discover(iface string) ([]string, error) {
	devices, err := onvif.GetAvailableDevicesAtSpecificEthernetInterface(iface)
	if err != nil {
		return nil, fmt.Errorf("discover on %s: %w", iface, err)
	}
	hosts := make([]string, 0, len(devices))
	for _, dev := range devices {
		if host, err := HostFromServiceURL(dev.GetServices()["device"]); err == nil {
			hosts = append(hosts, host)
		}
	}
	return hosts, nil
}

// HostFromServiceURL extracts host:port from a device service URL such as
// http://192.168.1.10:8080/onvif/device_service.
func HostFromServiceURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse service url %q: %w", raw, err)
	}
	if u.Host == "" {
		return "", fmt.Errorf("service url %q has no host", raw)
	}
	return u.Host, nil
}
