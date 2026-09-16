package motion

import (
	"strings"
	"testing"
)

const envelope = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope" xmlns:tt="http://www.onvif.org/ver10/schema" xmlns:tev="http://www.onvif.org/ver10/events/wsdl">
<SOAP-ENV:Body><tev:PullMessagesResponse><wsnt:NotificationMessage xmlns:wsnt="http://docs.oasis-open.org/wsn/b-2">
<wsnt:Message><tt:Message UtcTime="2024-01-01T00:00:00Z" PropertyOperation="Changed">
<tt:Source><tt:SimpleItem Name="VideoSourceConfigurationToken" Value="000"/></tt:Source>
<tt:Data>%s</tt:Data>
</tt:Message></wsnt:Message></wsnt:NotificationMessage></tev:PullMessagesResponse></SOAP-ENV:Body></SOAP-ENV:Envelope>`

func TestContainsMotion(t *testing.T) {
	cases := map[string]struct {
		data string
		want bool
	}{
		"motion true":            {`<tt:SimpleItem Name="IsMotion" Value="true" />`, true},
		"motion false":           {`<tt:SimpleItem Name="IsMotion" Value="false" />`, false},
		"attribute order":        {`<tt:SimpleItem Value="true" Name="IsMotion"/>`, true},
		"no namespace prefix":    {`<SimpleItem Name="IsMotion" Value="true"></SimpleItem>`, true},
		"other item true":        {`<tt:SimpleItem Name="IsTamper" Value="true"/>`, false},
		"empty data":             {``, false},
		"motion among many":      {`<tt:SimpleItem Name="State" Value="idle"/><tt:SimpleItem Name="IsMotion" Value="true"/>`, true},
		"value casing not truth": {`<tt:SimpleItem Name="IsMotion" Value="True"/>`, false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := ContainsMotion(strings.NewReader(strings.Replace(envelope, "%s", tc.data, 1)))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestContainsMotionMalformed(t *testing.T) {
	if _, err := ContainsMotion(strings.NewReader(`<a><b></a>`)); err == nil {
		t.Fatal("expected a parse error")
	}
}

func TestContainsMotionEmptyBody(t *testing.T) {
	got, err := ContainsMotion(strings.NewReader(""))
	if err != nil || got {
		t.Fatalf("got %v, %v; want false, nil", got, err)
	}
}
