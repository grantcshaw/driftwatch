package schedule

import (
	"context"
	"log"
	"time"

	"github.com/example/driftwatch/internal/alert"
	"github.com/example/driftwatch/internal/drift"
	"github.com/example/driftwatch/internal/environment"
	"github.com/example/driftwatch/internal/report"
)

// Runner periodically collects snapshots, detects drift, and notifies.
type Runner struct {
	registry  *environment.Registry
	detector  *drift.Detector
	notifier  *alert.Notifier
	reporter  *report.Report
	interval  time.Duration
	baseline  string
	targets   []string
}

// NewRunner creates a Runner with the provided dependencies.
func NewRunner(
	reg *environment.Registry,
	det *drift.Detector,
	not *alert.Notifier,
	rep *report.Report,
	interval time.Duration,
	baseline string,
	targets []string,
) *Runner {
	return &Runner{
		registry: reg,
		detector: det,
		notifier: not,
		reporter: rep,
		interval: interval,
		baseline: baseline,
		targets:  targets,
	}
}

// Run starts the periodic drift-check loop until ctx is cancelled.
func (r *Runner) Run(ctx context.Context) error {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	log.Printf("runner: starting drift checks every %s", r.interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("runner: context cancelled, stopping")
			return ctx.Err()
		case <-ticker.C:
			if err := r.runOnce(ctx); err != nil {
				log.Printf("runner: cycle error: %v", err)
			}
		}
	}
}

// RunOnce executes a single drift-check cycle (exported for testing).
func (r *Runner) RunOnce(ctx context.Context) error {
	return r.runOnce(ctx)
}

func (r *Runner) runOnce(ctx context.Context) error {
	snapshots, err := r.registry.CollectAll(ctx)
	if err != nil {
		return fmt.Errorf("collect snapshots: %w", err)
	}

	baseSnap, ok := snapshots[r.baseline]
	if !ok {
		return fmt.Errorf("baseline environment %q not found in snapshots", r.baseline)
	}

	var allDrifts []drift.Drift
	for _, target := range r.targets {
		tSnap, ok := snapshots[target]
		if !ok {
			log.Printf("runner: target %q missing, skipping", target)
			continue
		}
		drifts := r.detector.Detect(baseSnap, tSnap)
		allDrifts = append(allDrifts, drifts...)
	}

	if err := r.notifier.Notify(allDrifts); err != nil {
		log.Printf("runner: notify error: %v", err)
	}

	r.reporter.Write(allDrifts)
	return nil
}
