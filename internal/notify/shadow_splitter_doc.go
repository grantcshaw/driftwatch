// Package notify provides a suite of composable Sender implementations for
// dispatching drift alerts through various channels and pipelines.
//
// Shadow and Splitter senders:
//
// ShadowSender runs two senders in parallel — a primary and a shadow — and
// returns only the primary result. It records whether the two outcomes
// matched, enabling dark-launch testing of new notification backends without
// affecting production alerting. Use Results() to inspect mismatches and
// Reset() to clear the log between test cycles.
//
//	s, _ := notify.NewShadowSender(productionSender, candidateSender)
//	_ = s.Send("prod", drifts)
//	for _, r := range s.Results() {
//		if r.Mismatch { log.Println("shadow mismatch detected") }
//	}
//
// Splitter partitions a drift slice by a caller-supplied key function and
// routes each partition to a registered sender. A default sender handles
// any bucket without an explicit registration. Unrouted drifts with no
// default are silently dropped, preserving forward-compatibility when new
// severity levels or categories are introduced.
//
//	sp, _ := notify.NewSplitter(func(d drift.Drift) string { return d.Severity })
//	_ = sp.Register("critical", pagingSender)
//	_ = sp.Register("warning", emailSender)
//	_ = sp.SetDefault(logSender)
//	_ = sp.Send("staging", drifts)
package notify
