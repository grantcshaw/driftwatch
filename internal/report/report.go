package report

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/driftwatch/internal/drift"
)

// Format defines the output format for a report.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// Report holds the results of a drift detection run.
type Report struct {
	GeneratedAt time.Time
	Environment string
	Drifts      []drift.Drift
}

// New creates a new Report for the given environment and drifts.
func New(env string, drifts []drift.Drift) *Report {
	return &Report{
		GeneratedAt: time.Now().UTC(),
		Environment: env,
		Drifts:      drifts,
	}
}

// Write renders the report in the specified format to w.
func (r *Report) Write(w io.Writer, format Format) error {
	switch format {
	case FormatJSON:
		return r.writeJSON(w)
	case FormatText:
		return r.writeText(w)
	default:
		return fmt.Errorf("unsupported report format: %q", format)
	}
}

func (r *Report) writeText(w io.Writer) error {
	_, err := fmt.Fprintf(w, "DriftWatch Report\n")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "Generated : %s\n", r.GeneratedAt.Format(time.RFC3339))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "Environment: %s\n", r.Environment)
	if err != nil {
		return err
	}
	if len(r.Drifts) == 0 {
		_, err = fmt.Fprintln(w, "Status     : OK — no drift detected")
		return err
	}
	_, err = fmt.Fprintf(w, "Drifts     : %d\n%s\n", len(r.Drifts), strings.Repeat("-", 40))
	if err != nil {
		return err
	}
	for _, d := range r.Drifts {
		_, err = fmt.Fprintf(w, "  key=%-30s baseline=%-20s target=%s\n",
			d.Key, d.BaselineValue, d.TargetValue)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Report) writeJSON(w io.Writer) error {
	// Manual JSON to avoid importing encoding/json for a lightweight impl.
	var sb strings.Builder
	sb.WriteString("{\n")
	sb.WriteString(fmt.Sprintf("  \"generated_at\": %q,\n", r.GeneratedAt.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("  \"environment\": %q,\n", r.Environment))
	sb.WriteString(fmt.Sprintf("  \"drift_count\": %d,\n", len(r.Drifts)))
	sb.WriteString("  \"drifts\": [\n")
	for i, d := range r.Drifts {
		comma := ","
		if i == len(r.Drifts)-1 {
			comma = ""
		}
		sb.WriteString(fmt.Sprintf("    {\"key\": %q, \"baseline\": %q, \"target\": %q}%s\n",
			d.Key, d.BaselineValue, d.TargetValue, comma))
	}
	sb.WriteString("  ]\n}\n")
	_, err := io.WriteString(w, sb.String())
	return err
}
