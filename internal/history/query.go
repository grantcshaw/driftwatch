package history

import (
	"sort"
	"time"

	"github.com/your-org/driftwatch/internal/drift"
)

// QueryOptions controls filtering when reading history records.
type QueryOptions struct {
	Environment string
	Since       time.Time
	Limit       int
}

// Query returns drift records for the given environment filtered by options.
func (s *Store) Query(opts QueryOptions) ([]drift.Drift, error) {
	records, err := s.LoadAll(opts.Environment)
	if err != nil {
		return nil, err
	}

	var filtered []drift.Drift
	for _, r := range records {
		if !opts.Since.IsZero() && r.DetectedAt.Before(opts.Since) {
			continue
		}
		filtered = append(filtered, r)
	}

	// Sort descending by detection time (newest first).
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].DetectedAt.After(filtered[j].DetectedAt)
	})

	if opts.Limit > 0 && len(filtered) > opts.Limit {
		filtered = filtered[:opts.Limit]
	}

	return filtered, nil
}

// Summary holds aggregated statistics for a set of drift records.
type Summary struct {
	Environment string
	Total       int
	Critical    int
	Warning     int
	Oldest      time.Time
	Newest      time.Time
}

// Summarize computes statistics over all stored records for an environment.
func (s *Store) Summarize(env string) (Summary, error) {
	records, err := s.LoadAll(env)
	if err != nil {
		return Summary{}, err
	}

	sum := Summary{Environment: env, Total: len(records)}
	for _, r := range records {
		switch r.Severity {
		case "critical":
			sum.Critical++
		case "warning":
			sum.Warning++
		}
		if sum.Oldest.IsZero() || r.DetectedAt.Before(sum.Oldest) {
			sum.Oldest = r.DetectedAt
		}
		if r.DetectedAt.After(sum.Newest) {
			sum.Newest = r.DetectedAt
		}
	}
	return sum, nil
}
