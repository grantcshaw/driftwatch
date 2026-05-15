// Package notify provides a composable set of Sender implementations for
// delivering drift alerts through various channels and pipelines.
//
// LabelSender
//
// LabelSender is a middleware Sender that attaches static key/value labels
// to every Drift before forwarding to an inner Sender. This is useful for
// annotating alerts with environment metadata such as region, team, or
// deployment tier without modifying upstream collection logic.
//
// Example usage:
//
//	 base := notify.NewWebhookSender(cfg)
//	 labeled, err := notify.NewLabelSender(base, map[string]string{
//	     "region": "us-east-1",
//	     "team":   "platform",
//	 })
//	 if err != nil {
//	     log.Fatal(err)
//	 }
//	 // labeled now injects region and team into every drift's Metadata.
//
Existing drift Metadata keys take precedence; LabelSender will not
overwrite values already set on a Drift.
package notify
