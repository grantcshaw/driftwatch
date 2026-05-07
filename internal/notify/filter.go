package notify

import (
	"time"

	"github.com/yourorg/driftwatch/internal/drift"
)

// FilterConfig controls which drift events are forwarded to senders.
type FilterConfig struct {
	// MinSeverity filters out drifts below this level ("warning" or "critical").
	MinSeverity string

	// Cooldown suppresses repeated notifications for the same key within the
	// given duration. Zero means no cooldown.
	Cooldown time.Duration
}

// Filter wraps a Sender and applies suppression rules before forwarding.
type Filter struct {
	sender  Sender
	cfg     FilterConfig
	seen    map[string]time.Time
	nowFunc func() time.Time
}

// NewFilter creates a Filter wrapping the provided sender.
func NewFilter(s Sender, cfg FilterConfig) *Filter {
	return &Filter{
		sender:  s,
		cfg:     cfg,
		seen:    make(map[string]time.Time),
		nowFunc: time.Now,
	}
}

// Send applies filtering rules and forwards qualifying drifts to the inner sender.
func (f *Filter) Send(env string, drifts []drift.Drift) error {
	filtered := f.applyFilters(drifts)
	if len(filtered) == 0 {
		return nil
	}
	return f.sender.Send(env, filtered)
}

func (f *Filter) applyFilters(drifts []drift.Drift) []drift.Drift {
	now := f.nowFunc()
	var out []drift.Drift
	for _, d := range drifts {
		if !f.severityAllowed(d) {
			continue
		}
		if f.cfg.Cooldown > 0 {
			if last, ok := f.seen[d.Key]; ok && now.Sub(last) < f.cfg.Cooldown {
				continue
			}
			f.seen[d.Key] = now
		}
		out = append(out, d)
	}
	return out
}

func (f *Filter) severityAllowed(d drift.Drift) bool {
	if f.cfg.MinSeverity == "critical" {
		return d.Severity == "critical"
	}
	// "warning" or empty — allow everything
	return true
}
