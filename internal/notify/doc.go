// Package notify provides integrations for delivering drift alerts to
// external systems.
//
// Currently supported sinks:
//
//	- WebhookSender: HTTP POST of a JSON payload to a configurable URL.
//
// # Webhook payload
//
// The JSON body contains the environment name, a timestamp (UTC), the total
// number of drifted keys, and a list of individual drift entries each
// carrying the key name together with its baseline and current values.
//
// # Usage
//
//	sender := notify.NewWebhookSender("https://hooks.example.com/drift", 10*time.Second)
//	if err := sender.Send("production", drifts); err != nil {
//		log.Printf("webhook send failed: %v", err)
//	}
//
// A timeout of zero or negative defaults to 10 seconds.
package notify
