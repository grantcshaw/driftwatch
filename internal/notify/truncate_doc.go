// Package notify provides composable Sender middleware for delivering
// drift alerts through various channels and pipelines.
//
// # Truncate
//
// Truncate limits the number of drift entries forwarded to an inner Sender.
// This is useful when downstream systems (e.g. Slack, email) impose message
// size constraints or when noisy environments would otherwise flood alerts.
//
// When the number of incoming drifts exceeds the configured limit, the first
// N entries are forwarded and a synthetic "__truncated__" entry is appended
// describing how many additional drifts were suppressed.
//
// Example:
//
//	base := notify.NewSlackSender(webhookURL)
//	sender, err := notify.NewTruncate(base, 10)
//	if err != nil {
//		log.Fatal(err)
//	}
//	// sender will forward at most 10 drifts + a summary entry.
package notify
