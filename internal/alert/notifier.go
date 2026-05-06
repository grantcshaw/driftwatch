package alert

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/yourorg/driftwatch/internal/drift"
)

// Severity represents the urgency level of a drift alert.
type Severity string

const (
	SeverityInfo     Severity = "INFO"
	SeverityWarning  Severity = "WARNING"
	SeverityCritical Severity = "CRITICAL"
)

// Alert holds information about a detected drift event.
type Alert struct {
	Timestamp   time.Time
	Environment string
	Severity    Severity
	Drifts      []drift.DriftResult
}

// Notifier sends alerts to a configured output.
type Notifier struct {
	writer    io.Writer
	threshold int // minimum number of drifted keys to escalate to CRITICAL
}

// NewNotifier creates a Notifier that writes to the given writer.
// If writer is nil, os.Stdout is used.
func NewNotifier(writer io.Writer, criticalThreshold int) *Notifier {
	if writer == nil {
		writer = os.Stdout
	}
	if criticalThreshold <= 0 {
		criticalThreshold = 5
	}
	return &Notifier{writer: writer, threshold: criticalThreshold}
}

// Notify formats and writes an alert for the given drift results.
// Returns the Alert that was emitted, or nil if drifts is empty.
func (n *Notifier) Notify(environment string, drifts []drift.DriftResult) (*Alert, error) {
	if len(drifts) == 0 {
		return nil, nil
	}

	severity := SeverityWarning
	if len(drifts) >= n.threshold {
		severity = SeverityCritical
	}

	alert := &Alert{
		Timestamp:   time.Now().UTC(),
		Environment: environment,
		Severity:    severity,
		Drifts:      drifts,
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s] %s drift detected in environment %q — %d key(s) affected\n",
		alert.Timestamp.Format(time.RFC3339),
		string(severity),
		environment,
		len(drifts),
	))
	for _, d := range drifts {
		sb.WriteString(fmt.Sprintf("  key=%q baseline=%q target=%q\n", d.Key, d.BaselineValue, d.TargetValue))
	}

	_, err := fmt.Fprint(n.writer, sb.String())
	if err != nil {
		return nil, fmt.Errorf("alert: failed to write notification: %w", err)
	}
	return alert, nil
}
