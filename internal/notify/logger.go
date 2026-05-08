package notify

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/yourusername/driftwatch/internal/drift"
)

// LoggerSender is a Sender that writes drift notifications to a writer (e.g. stdout or a log file).
type LoggerSender struct {
	out    io.Writer
	prefix string
}

// NewLoggerSender creates a LoggerSender that writes to out.
// If out is nil, os.Stdout is used. prefix is prepended to each log line.
func NewLoggerSender(out io.Writer, prefix string) (*LoggerSender, error) {
	if out == nil {
		out = os.Stdout
	}
	return &LoggerSender{out: out, prefix: prefix}, nil
}

// Send writes a formatted log line for each drift to the configured writer.
func (l *LoggerSender) Send(env string, drifts []drift.Drift) error {
	if len(drifts) == 0 {
		return nil
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)
	for _, d := range drifts {
		line := fmt.Sprintf("%s [%s] env=%s key=%s baseline=%q current=%q\n",
			timestamp, l.prefix, env, d.Key, d.BaselineValue, d.CurrentValue)
		if _, err := fmt.Fprint(l.out, line); err != nil {
			return fmt.Errorf("logger sender: write failed: %w", err)
		}
	}
	return nil
}
