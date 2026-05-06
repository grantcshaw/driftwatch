package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/your-org/driftwatch/internal/drift"
)

// PruneOptions controls which records are removed.
type PruneOptions struct {
	// OlderThan removes records detected before this duration ago.
	OlderThan time.Duration
	// KeepLast retains at most this many records per environment (0 = unlimited).
	KeepLast int
}

// Prune removes old drift records for the given environment according to opts.
// It rewrites the on-disk file atomically.
func (s *Store) Prune(env string, opts PruneOptions) (int, error) {
	records, err := s.LoadAll(env)
	if err != nil {
		return 0, err
	}

	cutoff := time.Now().Add(-opts.OlderThan)
	var kept []drift.Drift
	for _, r := range records {
		if opts.OlderThan > 0 && r.DetectedAt.Before(cutoff) {
			continue
		}
		kept = append(kept, r)
	}

	// Sort newest first so KeepLast retains the most recent.
	sort.Slice(kept, func(i, j int) bool {
		return kept[i].DetectedAt.After(kept[j].DetectedAt)
	})
	if opts.KeepLast > 0 && len(kept) > opts.KeepLast {
		kept = kept[:opts.KeepLast]
	}

	removed := len(records) - len(kept)

	path := filepath.Join(s.dir, env+".json")
	if len(kept) == 0 {
		_ = os.Remove(path)
		return removed, nil
	}

	data, err := json.MarshalIndent(kept, "", "  ")
	if err != nil {
		return 0, err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return 0, err
	}
	return removed, nil
}
