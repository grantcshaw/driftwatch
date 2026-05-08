// Package notify provides composable Sender implementations for delivering
// drift alerts to external systems.
//
// Core senders:
//   - WebhookSender  — HTTP POST to a configurable endpoint
//   - SlackSender    — Slack incoming webhook with colour-coded severity
//   - EmailSender    — SMTP email delivery
//   - LoggerSender   — writes structured drift lines to an io.Writer
//
// Middleware / decorators (wrap any Sender):
//   - MultiSender    — fan-out to multiple senders in parallel
//   - Filter         — gate on minimum severity or cooldown
//   - RateLimiter    — cap sends per environment within a sliding window
//   - Throttle       — suppress repeated sends within a fixed interval
//   - Retry          — retry transient failures with configurable attempts
//   - DeadLetter     — persist failed payloads to disk for later replay
//   - Digest         — batch drifts over a time window before forwarding
//   - CircuitBreaker — open/half-open/closed protection against downstream failures
//
// Typical composition:
//
//	base := notify.NewWebhookSender(cfg)
//	withRetry, _ := notify.NewRetry(base, 3)
//	withCB, _ := notify.NewCircuitBreaker(withRetry, 5, time.Minute)
//	withRL, _ := notify.NewRateLimiter(withCB, time.Hour, 10)
//	sender := notify.NewFilter(withRL, drift.SeverityWarning, 0)
package notify
