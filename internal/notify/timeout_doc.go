// Package notify provides composable Sender implementations for delivering
// drift alerts through various channels and middleware layers.
//
// # Timeout
//
// Timeout wraps any Sender and enforces a maximum wall-clock duration on
// each Send call. If the inner sender does not complete within the
// configured deadline, Send returns an error immediately and the
// background goroutine is abandoned (the context passed to the inner
// sender is cancelled).
//
// This is useful when wrapping network-bound senders (webhook, Slack,
// email) where a stalled connection would otherwise block the scheduler
// indefinitely.
//
// Example:
//
//	base, _ := notify.NewWebhookSender(cfg)
//	guarded, _ := notify.NewTimeout(base, 5*time.Second)
//	// guarded.Send will return an error if base takes longer than 5 s.
package notify
