package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/example/driftwatch/internal/drift"
)

// SlackSender sends drift alerts to a Slack webhook URL.
type SlackSender struct {
	webhookURL string
	client     *http.Client
}

type slackPayload struct {
	Text        string            `json:"text"`
	Attachments []slackAttachment `json:"attachments,omitempty"`
}

type slackAttachment struct {
	Color  string `json:"color"`
	Title  string `json:"title"`
	Text   string `json:"text"`
	Footer string `json:"footer"`
}

// NewSlackSender creates a SlackSender that posts to the given Slack incoming webhook URL.
func NewSlackSender(webhookURL string) (*SlackSender, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("slack webhook URL must not be empty")
	}
	return &SlackSender{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Send posts a Slack message summarising the detected drifts.
// It is a no-op when drifts is empty.
func (s *SlackSender) Send(env string, drifts []drift.Drift) error {
	if len(drifts) == 0 {
		return nil
	}

	color := "warning"
	if len(drifts) >= 5 {
		color = "danger"
	}

	body := ""
	for _, d := range drifts {
		body += fmt.Sprintf("• *%s*: `%s` → `%s`\n", d.Key, d.BaselineValue, d.CurrentValue)
	}

	payload := slackPayload{
		Text: fmt.Sprintf(":rotating_light: Drift detected in *%s* (%d keys)", env, len(drifts)),
		Attachments: []slackAttachment{
			{
				Color:  color,
				Title:  "Changed Keys",
				Text:   body,
				Footer: "driftwatch",
			},
		},
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("slack: marshal payload: %w", err)
	}

	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("slack: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack: unexpected status %d", resp.StatusCode)
	}
	return nil
}
