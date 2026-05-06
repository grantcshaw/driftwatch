// Package notify provides pluggable notification backends for DriftWatch.
//
// Supported senders:
//
//   - WebhookSender: posts JSON payloads to an HTTP endpoint.
//   - SlackSender:   sends formatted messages to a Slack incoming webhook.
//   - EmailSender:   delivers drift alerts via SMTP.
//   - MultiSender:   fans out to multiple Sender implementations, collecting
//     all errors without short-circuiting.
//
// All senders implement the Sender interface:
//
//	type Sender interface {
//	    Send(env string, drifts []drift.Drift) error
//	}
//
// When no drifts are present, senders should return nil without performing
// any network I/O. Use MultiSender to compose several backends together:
//
//	m := notify.NewMultiSender(webhookSender, slackSender, emailSender)
//	if err := m.Send("production", drifts); err != nil {
//	    log.Printf("one or more notifications failed: %v", err)
//	}
package notify
