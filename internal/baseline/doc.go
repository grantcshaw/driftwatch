// Package baseline provides storage and management of reference snapshots
// used as the authoritative state for drift detection.
//
// # Overview
//
// A baseline is a persisted [environment.Snapshot] that represents the known-good
// or expected configuration for a given environment. The drift detector compares
// a freshly collected snapshot against the stored baseline to identify changes.
//
// # Usage
//
//	store, err := baseline.NewStore("/var/lib/driftwatch/baselines")
//	reg := environment.NewRegistry()
//	mgr := baseline.NewManager(store, reg)
//
//	// Capture a fresh baseline from the live environment:
//	snap, err := mgr.Capture("production")
//
//	// Or promote an already-collected snapshot:
//	err = mgr.Promote(snap)
//
//	// Retrieve the current baseline:
//	baseline, err := mgr.Current("production")
//
// Baselines are stored as JSON files under the configured directory,
// one file per environment (e.g. production.baseline.json).
package baseline
