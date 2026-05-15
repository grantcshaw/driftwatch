// Package notify provides a collection of Sender decorators for delivering
// drift alerts through various channels and with various delivery guarantees.
//
// Window
//
// Window is a rolling-window accumulator that suppresses low-volume drift
// noise. It buffers incoming drift events and only forwards them to the
// wrapped Sender once at least MinCount distinct keys have drifted within
// the configured time window. After forwarding, the window resets.
//
// This is useful when infrastructure churn is expected during deployments:
// a single key flapping should not wake on-call engineers, but five or more
// keys drifting simultaneously likely indicates a real problem.
//
// Example usage:
//
//	inner := notify.NewWebhookSender(cfg)
//	w, err := notify.NewWindow(inner, 5*time.Minute, 3)
//	if err != nil {
//		log.Fatal(err)
//	}
//	// w.Send will only call inner.Send when ≥3 unique keys drift
//	// within any 5-minute window.
package notify
