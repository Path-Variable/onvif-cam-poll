package camera

import "testing"

func TestHostFromServiceURL(t *testing.T) {
	cases := []struct {
		in, want string
		wantErr  bool
	}{
		{"http://192.168.1.10:8080/onvif/device_service", "192.168.1.10:8080", false},
		{"http://cam.local/onvif/device_service", "cam.local", false},
		{"https://[fe80::1]:443/x", "[fe80::1]:443", false},
		{"", "", true},
		{"not a url", "", true},
		{"://bad", "", true},
	}
	for _, tc := range cases {
		got, err := HostFromServiceURL(tc.in)
		if (err != nil) != tc.wantErr {
			t.Errorf("%q: err = %v, wantErr %v", tc.in, err, tc.wantErr)
		}
		if got != tc.want {
			t.Errorf("%q: got %q, want %q", tc.in, got, tc.want)
		}
	}
}
