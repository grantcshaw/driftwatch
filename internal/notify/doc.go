// Package notify provides a composable set of Sender implementations for
// delivering drift alerts through various channels and with various
// reliability and flow-control policies.
//
// # Channels
//
//   - WebhookSender  – HTTP POST to an arbitrary endpoint
//   - SlackSender    – Slack incoming webhook with colour-coded attachments
//   - EmailSender    – SMTP email delivery
//   - LoggerSender   – writes human-readable lines to an io.Writer
//
// # Reliability & flow control wrappers
//
//   - Retry        – retries on transient errors (fixed attempts)
//   - Backoff      – retries with configurable exponential back-off
//   - CircuitBreaker – opens after repeated failures, resets after a timeout
//   - RateLimiter  – caps sends per environment per rolling window
//   - Throttle     – enforces a minimum interval between sends per environment
//   - Dedup        – suppresses identical drift fingerprints within a window
//   - Filter       – drops drifts below a minimum severity or within a cooldown
//   - Digest       – batches drifts and delivers them on a fixed schedule
//   - Sampler      – probabilistically samples a fraction of drift events
//   - DeadLetter   – persists failed notifications to disk for later review
//   - Audit        – appends every attempted send to an on-disk audit log
//   - SenderMetrics – records send latency and outcome counters
//   - MultiSender  – fans out to multiple Sender implementations in parallel
//
// All wrappers accept a Sender interface so they can be freely composed.
package notify
