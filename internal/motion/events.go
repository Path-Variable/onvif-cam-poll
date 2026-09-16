// Package motion turns an ONVIF pull-point subscription into a stream of
// "motion detected" callbacks that survives camera hiccups.
package motion

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
)

// ContainsMotion reports whether an ONVIF PullMessages response carries a
// SimpleItem named IsMotion with the value true. It walks the XML tokens
// instead of matching text, so namespace prefixes, attribute order and
// whitespace do not matter.
func ContainsMotion(r io.Reader) (bool, error) {
	dec := xml.NewDecoder(r)
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			return false, nil
		}
		if err != nil {
			return false, fmt.Errorf("parse event message: %w", err)
		}
		start, ok := tok.(xml.StartElement)
		if !ok || start.Name.Local != "SimpleItem" {
			continue
		}
		if isMotionTrue(start.Attr) {
			return true, nil
		}
	}
}

func isMotionTrue(attrs []xml.Attr) bool {
	var name, value string
	for _, a := range attrs {
		switch a.Name.Local {
		case "Name":
			name = a.Value
		case "Value":
			value = a.Value
		}
	}
	return name == "IsMotion" && value == "true"
}
