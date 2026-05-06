package notify

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/example/driftwatch/internal/drift"
)

func makeSlackDrifts(n int) []drift.Drift {
	out := make([]drift.Drift, n)
	for i := 0; i < n; i++ {
		out[i] = drift.Drift{
			Key:           fmt.Sprintf("KEY_%d", i),
			BaselineValue: "old",
			CurrentValue:  "new",
		}
	}
	return out
}

func TestSlackSender_NewSender_EmptyURL(t *testing.T) {
	_, err := NewSlackSender("")
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestSlackSender_Send_NoDrifts(t *testing.T) {
	s, _ := NewSlackSender("http://example.com/hook")
	if err := s.Send("staging", nil); err != nil {
		t.Fatalf("expected no error for empty drifts, got %v", err)
	}
}

func TestSlackSender_Send_Success(t *testing.T) {
	var captured []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s, _ := NewSlackSender(ts.URL)
	drifts := makeSlackDrifts(2)
	if err := s.Send("production", drifts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var payload slackPayload
	if err := json.Unmarshal(captured, &payload); err != nil {
		t.Fatalf("invalid JSON sent to Slack: %v", err)
	}
	if !strings.Contains(payload.Text, "production") {
		t.Errorf("expected env name in text, got: %s", payload.Text)
	}
	if len(payload.Attachments) == 0 {
		t.Fatal("expected at least one attachment")
	}
}

func TestSlackSender_Send_WarningColor(t *testing.T) {
	var captured []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s, _ := NewSlackSender(ts.URL)
	if err := s.Send("staging", makeSlackDrifts(3)); err != nil {
		t.Fatal(err)
	}
	var payload slackPayload
	json.Unmarshal(captured, &payload)
	if payload.Attachments[0].Color != "warning" {
		t.Errorf("expected warning color, got %s", payload.Attachments[0].Color)
	}
}

func TestSlackSender_Send_DangerColor(t *testing.T) {
	var captured []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s, _ := NewSlackSender(ts.URL)
	if err := s.Send("prod", makeSlackDrifts(6)); err != nil {
		t.Fatal(err)
	}
	var payload slackPayload
	json.Unmarshal(captured, &payload)
	if payload.Attachments[0].Color != "danger" {
		t.Errorf("expected danger color, got %s", payload.Attachments[0].Color)
	}
}

func TestSlackSender_Send_NonSuccessStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	s, _ := NewSlackSender(ts.URL)
	err := s.Send("env", makeSlackDrifts(1))
	if err == nil {
		t.Fatal("expected error on non-2xx status")
	}
}
