// Package history persists drift detection results over time and provides
// querying, pruning, and summarisation utilities.
//
// # Store
//
// NewStore creates a file-backed store rooted at a given directory. Each
// environment gets its own JSON record file. Use Save to append a new drift
// event and LoadAll to retrieve the full history for an environment.
//
// # Query helpers
//
// Since filters records to those detected after a given time.
// Limit returns the most-recent N records.
// Summarize returns per-key drift counts across a slice of records.
//
// # Summary
//
// BuildEnvSummary aggregates stored records for an environment into an
// EnvSummary, including total event count, time range, and the most
// frequently drifting keys. WriteSummary formats the summary for human
// consumption.
//
// # Pruner
//
// Prune removes records that are older than a retention threshold or trims
// history to a maximum number of records per environment, keeping the newest.
package history
