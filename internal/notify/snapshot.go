// Package notify posts motion notifications and camera snapshots to Slack.
package notify

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// MaxSnapshotBytes bounds what is read from the camera; snapshots are a few
// hundred kilobytes, so anything larger is a misbehaving device.
const MaxSnapshotBytes = 20 << 20

// FetchSnapshot downloads a snapshot and refuses anything that is not an
// image, so a login page or an error document is never uploaded as a photo.
func FetchSnapshot(ctx context.Context, client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build snapshot request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch snapshot: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch snapshot: camera answered HTTP %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "image/") {
		return nil, fmt.Errorf("fetch snapshot: expected an image, got content-type %q", ct)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxSnapshotBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read snapshot: %w", err)
	}
	if len(body) > MaxSnapshotBytes {
		return nil, errors.New("read snapshot: larger than the 20 MiB limit")
	}
	if len(body) == 0 {
		return nil, errors.New("read snapshot: empty body")
	}
	return body, nil
}
