package notify_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/drift"
	"github.com/example/driftwatch/internal/notify"
)

func makeDrifts(keys ...string) []drift.Drift {
	ds := make([]drift.Drift, 0, len(keys))
	for _, k := range keys {
		ds = append(ds, drift.Drift{Key: k, Baseline: "old", Current: "new"})
	}
	return ds
}

func TestWebhookSender_Send_Success(t *testing.T) {
	var received notify.WebhookPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	sender := notify.NewWebhookSender(ts.URL, 5*time.Second)
	drifts := makeDrifts("DB_HOST", "API_KEY")

	if err := sender.Send("staging", drifts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received.Environment != "staging" {
		t.Errorf("env = %q, want %q", received.Environment, "staging")
	}
	if received.DriftCount != 2 {
		t.Errorf("drift_count = %d, want 2", received.DriftCount)
	}
	if len(received.Drifts) != 2 {
		t.Fatalf("drifts len = %d, want 2", len(received.Drifts))
	}
	if received.Drifts[0].Key != "DB_HOST" {
		t.Errorf("first key = %q, want DB_HOST", received.Drifts[0].Key)
	}
}

func TestWebhookSender_Send_NoDrifts(t *testing.T) {
	var received notify.WebhookPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received) //nolint:errcheck
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	sender := notify.NewWebhookSender(ts.URL, 0)
	if err := sender.Send("prod", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.DriftCount != 0 {
		t.Errorf("expected 0 drifts, got %d", received.DriftCount)
	}
}

func TestWebhookSender_Send_NonSuccessStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	sender := notify.NewWebhookSender(ts.URL, 5*time.Second)
	err := sender.Send("prod", makeDrifts("KEY"))
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

func TestWebhookSender_Send_BadURL(t *testing.T) {
	sender := notify.NewWebhookSender("http://127.0.0.1:1", 500*time.Millisecond)
	err := sender.Send("dev", makeDrifts("X"))
	if err == nil {
		t.Fatal("expected connection error, got nil")
	}
}
