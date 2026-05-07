package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/example/driftwatch/internal/drift"
)

// WebhookPayload is the JSON body sent to a webhook endpoint.
type WebhookPayload struct {
	Timestamp   time.Time    `json:"timestamp"`
	Environment string       `json:"environment"`
	DriftCount  int          `json:"drift_count"`
	Drifts      []DriftEntry `json:"drifts"`
}

// DriftEntry is a single drift item serialised for the webhook.
type DriftEntry struct {
	Key      string `json:"key"`
	Baseline string `json:"baseline"`
	Current  string `json:"current"`
}

// WebhookSender sends drift notifications to an HTTP webhook.
type WebhookSender struct {
	URL     string
	Timeout time.Duration
	client  *http.Client
}

// NewWebhookSender creates a WebhookSender. timeout <= 0 defaults to 10s.
func NewWebhookSender(url string, timeout time.Duration) *WebhookSender {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &WebhookSender{
		URL:     url,
		Timeout: timeout,
		client:  &http.Client{Timeout: timeout},
	}
}

// Send posts the drift results for the named environment to the webhook URL.
func (w *WebhookSender) Send(env string, drifts []drift.Drift) error {
	entries := make([]DriftEntry, 0, len(drifts))
	for _, d := range drifts {
		entries = append(entries, DriftEntry{
			Key:      d.Key,
			Baseline: d.Baseline,
			Current:  d.Current,
		})
	}

	payload := WebhookPayload{
		Timestamp:   time.Now().UTC(),
		Environment: env,
		DriftCount:  len(drifts),
		Drifts:      entries,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: marshal payload: %w", err)
	}

	resp, err := w.client.Post(w.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: post to %s: %w", w.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Read up to 256 bytes of the response body to aid debugging.
		preview, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return fmt.Errorf("webhook: unexpected status %d from %s: %s", resp.StatusCode, w.URL, bytes.TrimSpace(preview))
	}
	return nil
}
