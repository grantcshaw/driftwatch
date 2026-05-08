// Package notify provides pluggable notification senders for drift alerts.
//
// Senders implement the Sender interface and can be composed using the
// provided wrappers:
//
//   - MultiSender   — fan-out to multiple senders
//   - Filter        — skip sends below a severity threshold or within cooldown
//   - RateLimiter   — cap sends per environment per time window
//   - Throttle      — suppress repeated sends within a fixed interval
//   - Retry         — retry failed sends with configurable attempts
//   - DeadLetter    — persist failed payloads to disk for later inspection
//   - Digest        — batch drifts over a time window before forwarding
//   - CircuitBreaker — open the circuit after consecutive failures
//   - SenderMetrics  — instrument any sender with send counts and latency
//   - LoggerSender  — write drift events to an io.Writer
//
// Concrete transport implementations:
//
//   - WebhookSender — HTTP POST JSON payload
//   - SlackSender   — Slack incoming webhook with colour-coded attachments
//   - EmailSender   — SMTP plain-text email
//
// Compose wrappers around a transport to build a resilient notification
// pipeline, e.g.:
//
//	sender := notify.NewRetry(
//		notify.NewCircuitBreaker(
//			notify.NewWebhookSender(url),
//			5, time.Minute,
//		),
//		3,
//	)
package notify
