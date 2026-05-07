package notify

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/org/driftwatch/internal/drift"
)

// DeadLetterEntry records a failed notification attempt.
type DeadLetterEntry struct {
	Timestamp time.Time    `json:"timestamp"`
	Env       string       `json:"env"`
	Drifts    []drift.Drift `json:"drifts"`
	Error     string       `json:"error"`
}

// DeadLetter wraps a Sender and writes failed deliveries to a directory on
// disk so they can be inspected or replayed later.
type DeadLetter struct {
	inner Sender
	dir   string
}

// NewDeadLetter creates a DeadLetter sender. dir is created if it does not
// exist.
func NewDeadLetter(inner Sender, dir string) (*DeadLetter, error) {
	if inner == nil {
		return nil, fmt.Errorf("deadletter: inner sender must not be nil")
	}
	if dir == "" {
		return nil, fmt.Errorf("deadletter: dir must not be empty")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("deadletter: create dir: %w", err)
	}
	return &DeadLetter{inner: inner, dir: dir}, nil
}

// Send delivers drifts via the inner sender. On failure the attempt is
// persisted as a JSON file under the configured directory.
func (d *DeadLetter) Send(env string, drifts []drift.Drift) error {
	if err := d.inner.Send(env, drifts); err != nil {
		entry := DeadLetterEntry{
			Timestamp: time.Now().UTC(),
			Env:       env,
			Drifts:    drifts,
			Error:     err.Error(),
		}
		if writeErr := d.persist(entry); writeErr != nil {
			return fmt.Errorf("%w (dead-letter write failed: %v)", err, writeErr)
		}
		return err
	}
	return nil
}

func (d *DeadLetter) persist(e DeadLetterEntry) error {
	name := fmt.Sprintf("%d_%s.json", e.Timestamp.UnixNano(), e.Env)
	path := filepath.Join(d.dir, name)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(e)
}
