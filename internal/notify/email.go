package notify

import (
	"bytes"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/yourorg/driftwatch/internal/drift"
)

// EmailConfig holds SMTP configuration for sending email alerts.
type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	To       []string
}

// EmailSender sends drift alerts via email.
type EmailSender struct {
	cfg  EmailConfig
	auth smtp.Auth
	dial func(addr string, a smtp.Auth, from string, to []string, msg []byte) error
}

// NewEmailSender creates a new EmailSender. Returns an error if configuration is invalid.
func NewEmailSender(cfg EmailConfig) (*EmailSender, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("email: host must not be empty")
	}
	if len(cfg.To) == 0 {
		return nil, fmt.Errorf("email: at least one recipient required")
	}
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	return &EmailSender{
		cfg:  cfg,
		auth: auth,
		dial: smtp.SendMail,
	}, nil
}

// Send sends an email notification if drifts are present.
func (e *EmailSender) Send(env string, drifts []drift.Drift) error {
	if len(drifts) == 0 {
		return nil
	}

	var body bytes.Buffer
	fmt.Fprintf(&body, "Subject: [DriftWatch] Drift detected in %s\r\n", env)
	fmt.Fprintf(&body, "From: %s\r\n", e.cfg.From)
	fmt.Fprintf(&body, "To: %s\r\n", strings.Join(e.cfg.To, ", "))
	fmt.Fprintf(&body, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&body, "Content-Type: text/plain; charset=utf-8\r\n\r\n")
	fmt.Fprintf(&body, "Drift detected in environment: %s\n\n", env)
	fmt.Fprintf(&body, "%d key(s) differ:\n", len(drifts))
	for _, d := range drifts {
		fmt.Fprintf(&body, "  - %s: baseline=%q target=%q\n", d.Key, d.BaselineValue, d.TargetValue)
	}

	addr := fmt.Sprintf("%s:%d", e.cfg.Host, e.cfg.Port)
	return e.dial(addr, e.auth, e.cfg.From, e.cfg.To, body.Bytes())
}
