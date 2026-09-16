// Package systime builds the ONVIF SetSystemDateAndTime request from a Go time.
package systime

import (
	"time"

	"github.com/isaric/go-posix-time/pkg/p_time"
	"github.com/use-go/onvif/device"
	"github.com/use-go/onvif/xsd"
	xonvif "github.com/use-go/onvif/xsd/onvif"
)

// Request describes now to the camera: the clock is sent in UTC, as the
// ONVIF field name says, and the zone is sent as a POSIX TZ string so the
// camera can render local time and follow daylight-saving changes itself.
func Request(now time.Time) device.SetSystemDateAndTime {
	utc := now.UTC()
	return device.SetSystemDateAndTime{
		DateTimeType:    "Manual",
		DaylightSavings: xsd.Boolean(now.IsDST()),
		TimeZone:        xonvif.TimeZone{TZ: xsd.Token(p_time.FormatTimeZone(now))},
		UTCDateTime: xonvif.DateTime{
			Time: xonvif.Time{Hour: xsd.Int(utc.Hour()), Minute: xsd.Int(utc.Minute()), Second: xsd.Int(utc.Second())},
			Date: xonvif.Date{Year: xsd.Int(utc.Year()), Month: xsd.Int(utc.Month()), Day: xsd.Int(utc.Day())},
		},
	}
}
