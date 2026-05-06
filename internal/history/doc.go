// Package history provides persistent storage and querying of drift detection
// records for driftwatch.
//
// Records are written as JSON files under a configurable directory, one file
// per monitored environment.  The package exposes three primary concerns:
//
//   - Store: low-level save/load of drift.Drift slices.
//   - Query / Summarize: filtered reads and aggregate statistics.
//   - Prune: time-based and count-based cleanup of old records.
//
// Typical usage:
//
//	store, err := history.NewStore("/var/lib/driftwatch/history")
//	if err != nil { ... }
//
//	// persist new drifts
//	_ = store.Save("production", drifts)
//
//	// query recent records
//	results, _ := store.Query(history.QueryOptions{
//		Environment: "production",
//		Since:       time.Now().Add(-24 * time.Hour),
//		Limit:       50,
//	})
//
//	// remove records older than 30 days
//	_, _ = store.Prune("production", history.PruneOptions{
//		OlderThan: 30 * 24 * time.Hour,
//	})
package history
