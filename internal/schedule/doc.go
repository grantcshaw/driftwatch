// Package schedule provides the Runner type, which orchestrates periodic
// infrastructure drift checks for driftwatch.
//
// A Runner ties together the environment Registry (snapshot collection),
// drift Detector, alert Notifier, and report Writer into a single control
// loop that fires on a configurable interval.
//
// Typical usage:
//
//	runner := schedule.NewRunner(
//		registry,
//		detector,
//		notifier,
//		reporter,
//		5*time.Minute,
//		"production",          // baseline environment
//		[]string{"staging"},   // environments to compare against baseline
//	)
//	if err := runner.Run(ctx); err != nil && err != context.Canceled {
//		log.Fatalf("runner exited: %v", err)
//	}
package schedule
