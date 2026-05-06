package report

import (
	"fmt"
	"strings"
)

// ParseFormat converts a raw string to a Format, returning an error if unknown.
func ParseFormat(raw string) (Format, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "text", "":
		return FormatText, nil
	case "json":
		return FormatJSON, nil
	default:
		return "", fmt.Errorf("unknown report format %q: supported formats are text, json", raw)
	}
}

// Summary returns a one-line human-readable summary of the report.
func (r *Report) Summary() string {
	if len(r.Drifts) == 0 {
		return fmt.Sprintf("[%s] OK — no configuration drift detected", r.Environment)
	}
	keys := make([]string, 0, len(r.Drifts))
	for _, d := range r.Drifts {
		keys = append(keys, d.Key)
	}
	return fmt.Sprintf("[%s] DRIFT — %d key(s) differ: %s",
		r.Environment, len(r.Drifts), strings.Join(keys, ", "))
}
