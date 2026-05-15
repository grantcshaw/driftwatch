// Package notify provides composable Sender implementations for delivering
// drift alerts through various channels and pipelines.
//
// # Schema
//
// Schema wraps a Sender and validates drift payloads against a declared set
// of required and forbidden keys before forwarding.
//
// Required keys must be present (as drift keys) in the payload; if any
// required key is absent, Send returns a validation error and the inner
// Sender is never called.
//
// Forbidden keys must not appear in the payload; if any forbidden key is
// found, Send returns a validation error immediately.
//
// This is useful for enforcing contracts between environments — for example,
// ensuring that critical configuration keys are always reported when they
// drift, or preventing sensitive keys (e.g. secrets) from being forwarded
// to external notification channels.
//
// Example:
//
//	s, err := notify.NewSchema(
//		webhookSender,
//		[]string{"db_host", "db_port"},  // required
//		[]string{"api_secret"},           // forbidden
//	)
package notify
