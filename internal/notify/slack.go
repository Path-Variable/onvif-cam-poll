package notify

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/slack-go/slack"
)

// Slack posts messages and snapshots to one channel.
type Slack struct {
	client  *slack.Client
	channel string
	http    *http.Client
	log     *slog.Logger
}

// NewSlack builds a notifier for the channel using the bot token.
func NewSlack(token, channel string, log *slog.Logger) *Slack {
	return &Slack{
		client:  slack.New(token),
		channel: channel,
		http:    &http.Client{Timeout: 15 * time.Second},
		log:     log,
	}
}

// Post sends a text message to the channel.
func (s *Slack) Post(ctx context.Context, text string) error {
	if _, _, err := s.client.PostMessageContext(ctx, s.channel, slack.MsgOptionText(text, false)); err != nil {
		return fmt.Errorf("post slack message: %w", err)
	}
	return nil
}

// UploadSnapshot fetches the snapshot URL and uploads it as a timestamped
// JPEG. Failures are returned, not fatal: a missing picture must not stop
// the alert that was already posted.
func (s *Slack) UploadSnapshot(ctx context.Context, url string) error {
	img, err := FetchSnapshot(ctx, s.http, url)
	if err != nil {
		return err
	}
	_, err = s.client.UploadFileV2Context(ctx, slack.UploadFileV2Parameters{
		Reader:   bytes.NewReader(img),
		FileSize: len(img),
		Filename: time.Now().UTC().Format("20060102T150405Z") + ".jpg",
		Channel:  s.channel,
	})
	if err != nil {
		return fmt.Errorf("upload snapshot to slack: %w", err)
	}
	return nil
}
