// Package notify provides composable notification senders for drift alerts.
//
// # Senders
//
// The core interface is Sender, which accepts an environment name and a slice
// of drift.Drift values. Implementations include:
//
//   - WebhookSender  — posts JSON payloads to an HTTP endpoint
//   - SlackSender    — formats and posts messages to a Slack webhook
//   - EmailSender    — sends SMTP email notifications
//   - LoggerSender   — writes drift events to an io.Writer (default: stdout)
//
// # Middleware / Decorators
//
// Senders can be wrapped with middleware to add cross-cutting behaviour:
//
//   - Filter        — skips sends below a minimum severity or within a cooldown window
//   - RateLimiter   — caps the number of sends per environment per time window
//   - Throttle      — suppresses repeated sends within a configurable interval
//   - Retry         — retries failed sends with a fixed delay
//   - DeadLetter    — persists failed payloads to disk for later inspection
//   - Digest        — batches drifts and sends a single digest after a time window
//   - CircuitBreaker — opens after N consecutive failures and resets after a timeout
//   - SenderMetrics  — records send counts and latencies via expvar counters
//   - AuditSender   — appends a JSONL audit log entry for every send attempt
//
// # Composition
//
// Use MultiSender to fan out to multiple senders simultaneously.
package notify
