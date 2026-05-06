package history

import (
	"fmt"
	"io"
	"sort"
	"time"
)

// EnvSummary holds aggregated drift statistics for a single environment.
type EnvSummary struct {
	Environment  string
	TotalDrifts  int
	FirstSeen    time.Time
	LastSeen     time.Time
	TopKeys      []string // keys that drifted most frequently
}

// BuildEnvSummary aggregates drift records for a given environment from the store.
func BuildEnvSummary(store *Store, env string, since time.Time) (EnvSummary, error) {
	records, err := store.LoadAll(env)
	if err != nil {
		return EnvSummary{}, fmt.Errorf("load history for %q: %w", env, err)
	}

	filtered := Since(records, since)
	if len(filtered) == 0 {
		return EnvSummary{Environment: env}, nil
	}

	keyCounts := make(map[string]int)
	var first, last time.Time

	for _, r := range filtered {
		if first.IsZero() || r.DetectedAt.Before(first) {
			first = r.DetectedAt
		}
		if r.DetectedAt.After(last) {
			last = r.DetectedAt
		}
		for _, d := range r.Drifts {
			keyCounts[d.Key]++
		}
	}

	topKeys := rankKeys(keyCounts, 5)

	return EnvSummary{
		Environment: env,
		TotalDrifts: len(filtered),
		FirstSeen:   first,
		LastSeen:    last,
		TopKeys:     topKeys,
	}, nil
}

// WriteSummary writes a human-readable summary to w.
func WriteSummary(w io.Writer, s EnvSummary) {
	if s.TotalDrifts == 0 {
		fmt.Fprintf(w, "[%s] No drift records found.\n", s.Environment)
		return
	}
	fmt.Fprintf(w, "[%s] Drift summary\n", s.Environment)
	fmt.Fprintf(w, "  Total drift events : %d\n", s.TotalDrifts)
	fmt.Fprintf(w, "  First seen         : %s\n", s.FirstSeen.Format(time.RFC3339))
	fmt.Fprintf(w, "  Last seen          : %s\n", s.LastSeen.Format(time.RFC3339))
	if len(s.TopKeys) > 0 {
		fmt.Fprintf(w, "  Most-drifted keys  : %v\n", s.TopKeys)
	}
}

// rankKeys returns up to n keys sorted by descending count.
func rankKeys(counts map[string]int, n int) []string {
	type kv struct {
		key   string
		count int
	}
	var pairs []kv
	for k, v := range counts {
		pairs = append(pairs, kv{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].count != pairs[j].count {
			return pairs[i].count > pairs[j].count
		}
		return pairs[i].key < pairs[j].key
	})
	if len(pairs) > n {
		pairs = pairs[:n]
	}
	keys := make([]string, len(pairs))
	for i, p := range pairs {
		keys[i] = p.key
	}
	return keys
}
