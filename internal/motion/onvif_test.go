package motion

import (
	"testing"
	"time"
)

func TestISODuration(t *testing.T) {
	cases := map[time.Duration]string{0: "PT0S", time.Second: "PT1S", 5 * time.Minute: "PT300S", 1500 * time.Millisecond: "PT1S"}
	for in, want := range cases {
		if got := isoDuration(in); got != want {
			t.Errorf("isoDuration(%v) = %q, want %q", in, got, want)
		}
	}
}
