package notify

import (
	"net/smtp"
	"strings"
	"testing"

	"github.com/yourorg/driftwatch/internal/drift"
)

func makeEmailDrifts(n int) []drift.Drift {
	out := make([]drift.Drift, n)
	for i := range out {
		out[i] = drift.Drift{
			Key:           fmt.Sprintf("KEY_%d", i),
			BaselineValue: "old",
			TargetValue:   "new",
		}
	}
	return out
}

func validEmailCfg() EmailConfig {
	return EmailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
		From:     "alerts@example.com",
		To:       []string{"ops@example.com"},
	}
}

func TestEmailSender_NewSender_EmptyHost(t *testing.T) {
	cfg := validEmailCfg()
	cfg.Host = ""
	_, err := NewEmailSender(cfg)
	if err == nil {
		t.Fatal("expected error for empty host")
	}
}

func TestEmailSender_NewSender_NoRecipients(t *testing.T) {
	cfg := validEmailCfg()
	cfg.To = nil
	_, err := NewEmailSender(cfg)
	if err == nil {
		t.Fatal("expected error for no recipients")
	}
}

func TestEmailSender_Send_NoDrifts(t *testing.T) {
	s, _ := NewEmailSender(validEmailCfg())
	called := false
	s.dial = func(_ string, _ smtp.Auth, _ string, _ []string, _ []byte) error {
		called = true
		return nil
	}
	if err := s.Send("staging", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("dial should not be called when no drifts")
	}
}

func TestEmailSender_Send_Success(t *testing.T) {
	s, _ := NewEmailSender(validEmailCfg())
	var capturedMsg []byte
	s.dial = func(_ string, _ smtp.Auth, _ string, _ []string, msg []byte) error {
		capturedMsg = msg
		return nil
	}
	drifts := []drift.Drift{{Key: "DB_HOST", BaselineValue: "prod-db", TargetValue: "staging-db"}}
	if err := s.Send("staging", drifts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	body := string(capturedMsg)
	if !strings.Contains(body, "DB_HOST") {
		t.Error("expected body to contain drift key")
	}
	if !strings.Contains(body, "staging") {
		t.Error("expected body to reference environment")
	}
}

func TestEmailSender_Send_DialError(t *testing.T) {
	s, _ := NewEmailSender(validEmailCfg())
	s.dial = func(_ string, _ smtp.Auth, _ string, _ []string, _ []byte) error {
		return fmt.Errorf("connection refused")
	}
	err := s.Send("prod", []drift.Drift{{Key: "X", BaselineValue: "a", TargetValue: "b"}})
	if err == nil {
		t.Fatal("expected error from dial failure")
	}
}
