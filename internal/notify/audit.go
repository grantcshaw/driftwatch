package notify

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/yourusername/driftwatch/internal/drift"
)

// AuditEntry records a single notification attempt.
type AuditEntry struct {
	Timestamp time.Time    `json:"timestamp"`
	Env       string       `json:"env"`
	Severity  string       `json:"severity"`
	DriftKeys []string     `json:"drift_keys"`
	Success   bool         `json:"success"`
	Error     string       `json:"error,omitempty"`
}

// AuditSender wraps a Sender and writes an audit log entry for every Send call.
type AuditSender struct {
	inner   Sender
	dir     string
	mu      sync.Mutex
}

// NewAuditSender creates an AuditSender that writes audit logs to dir.
func NewAuditSender(inner Sender, dir string) (*AuditSender, error) {
	if inner == nil {
		return nil, fmt.Errorf("audit: inner sender must not be nil")
	}
	if dir == "" {
		return nil, fmt.Errorf("audit: dir must not be empty")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("audit: creating dir: %w", err)
	}
	return &AuditSender{inner: inner, dir: dir}, nil
}

// Send forwards drifts to the inner sender and records the outcome.
func (a *AuditSender) Send(env string, drifts []drift.Drift) error {
	keys := make([]string, 0, len(drifts))
	severity := "warning"
	for _, d := range drifts {
		keys = append(keys, d.Key)
		if d.Severity == "critical" {
			severity = "critical"
		}
	}

	entry := AuditEntry{
		Timestamp: time.Now().UTC(),
		Env:       env,
		Severity:  severity,
		DriftKeys: keys,
		Success:   true,
	}

	if len(drifts) == 0 {
		return nil
	}

	sendErr := a.inner.Send(env, drifts)
	if sendErr != nil {
		entry.Success = false
		entry.Error = sendErr.Error()
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.appendEntry(entry); err != nil {
		// best-effort: don't mask the original error
		_ = err
	}
	return sendErr
}

func (a *AuditSender) appendEntry(entry AuditEntry) error {
	filename := filepath.Join(a.dir, fmt.Sprintf("audit-%s.jsonl", entry.Env))
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	return enc.Encode(entry)
}
