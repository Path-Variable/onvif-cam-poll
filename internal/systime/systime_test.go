package systime

import (
	"strings"
	"testing"
	"time"
)

func TestRequestSendsUTCAndPosixZone(t *testing.T) {
	madrid, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		t.Skip("tzdata not available")
	}
	// 15 July 2024 14:30 CEST is 12:30 UTC.
	now := time.Date(2024, 7, 15, 14, 30, 5, 0, madrid)
	req := Request(now)

	if int(req.UTCDateTime.Time.Hour) != 12 || int(req.UTCDateTime.Time.Minute) != 30 || int(req.UTCDateTime.Time.Second) != 5 {
		t.Errorf("UTC time = %d:%d:%d, want 12:30:05", req.UTCDateTime.Time.Hour, req.UTCDateTime.Time.Minute, req.UTCDateTime.Time.Second)
	}
	if int(req.UTCDateTime.Date.Year) != 2024 || int(req.UTCDateTime.Date.Month) != 7 || int(req.UTCDateTime.Date.Day) != 15 {
		t.Errorf("UTC date = %d-%d-%d, want 2024-7-15", req.UTCDateTime.Date.Year, req.UTCDateTime.Date.Month, req.UTCDateTime.Date.Day)
	}
	if !bool(req.DaylightSavings) {
		t.Error("DaylightSavings should be true in July")
	}
	// go-posix-time < 1.2 writes the default 02:00 explicitly ("M3.5.0/2"); both are valid POSIX.
	if got := string(req.TimeZone.TZ); !strings.HasPrefix(got, "CET-1CEST,M3.5.0") || !strings.HasSuffix(got, ",M10.5.0/3") {
		t.Errorf("TZ = %q, want CET-1CEST,M3.5.0[/2],M10.5.0/3", got)
	}
}

func TestRequestCrossesMidnightCorrectly(t *testing.T) {
	tokyo, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Skip("tzdata not available")
	}
	// 01:00 JST on the 2nd is 16:00 UTC on the 1st.
	req := Request(time.Date(2024, 3, 2, 1, 0, 0, 0, tokyo))
	if int(req.UTCDateTime.Date.Day) != 1 || int(req.UTCDateTime.Time.Hour) != 16 {
		t.Errorf("got day %d hour %d, want day 1 hour 16", req.UTCDateTime.Date.Day, req.UTCDateTime.Time.Hour)
	}
	if bool(req.DaylightSavings) {
		t.Error("Tokyo has no DST")
	}
}
