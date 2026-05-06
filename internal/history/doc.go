// Package history provides persistent storage for drift detection records.
//
// Each time a drift check runs, the results can be saved to a Store which
// writes JSON-encoded Record files to a configured directory. Records are
// named using the environment name and a UTC timestamp, making them easy
// to inspect manually or load back for trend analysis.
//
// Example usage:
//
//	store, err := history.NewStore("/var/lib/driftwatch/history")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// After running drift detection:
//	if err := store.Save("staging", drifts); err != nil {
//		log.Printf("warn: could not save history: %v", err)
//	}
//
//	// Load all past records for an environment:
//	records, err := store.LoadAll("staging")
//	if err != nil {
//		log.Fatal(err)
//	}
package history
