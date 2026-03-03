package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"machineauth/internal/db"
	"machineauth/internal/models"
)

func setupTestDB(t *testing.T) db.Database {
	t.Helper()
	tmpFile := fmt.Sprintf("test_db_%d.json", time.Now().UnixNano())
	t.Cleanup(func() { os.Remove(tmpFile) })
	database, err := db.Connect("json:" + tmpFile)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	return database
}

func TestWebhookService_CreateWebhook(t *testing.T) {
	database := setupTestDB(t)
	svc := NewWebhookService(database)

	tests := []struct {
		name    string
		req     models.CreateWebhookRequest
		wantErr bool
	}{
		{
			name: "valid webhook",
			req: models.CreateWebhookRequest{
				Name:   "Test Webhook",
				URL:    "https://example.com/hook",
				Events: []string{"agent.created"},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			req: models.CreateWebhookRequest{
				URL:    "https://example.com/hook",
				Events: []string{"agent.created"},
			},
			wantErr: true,
		},
		{
			name: "missing URL",
			req: models.CreateWebhookRequest{
				Name:   "Test",
				Events: []string{"agent.created"},
			},
			wantErr: true,
		},
		{
			name: "invalid URL scheme",
			req: models.CreateWebhookRequest{
				Name:   "Test",
				URL:    "ftp://example.com/hook",
				Events: []string{"agent.created"},
			},
			wantErr: true,
		},
		{
			name: "no events",
			req: models.CreateWebhookRequest{
				Name:   "Test",
				URL:    "https://example.com/hook",
				Events: []string{},
			},
			wantErr: true,
		},
		{
			name: "unknown event",
			req: models.CreateWebhookRequest{
				Name:   "Test",
				URL:    "https://example.com/hook",
				Events: []string{"unknown.event"},
			},
			wantErr: true,
		},
		{
			name: "wildcard event",
			req: models.CreateWebhookRequest{
				Name:   "All Events",
				URL:    "https://example.com/hook",
				Events: []string{"*"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.CreateWebhook(tt.req)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.Secret == "" {
				t.Error("expected secret to be generated")
			}
			if resp.Webhook.ID.String() == "" {
				t.Error("expected ID to be set")
			}
			if !resp.Webhook.IsActive {
				t.Error("expected webhook to be active by default")
			}
		})
	}
}

func TestWebhookService_CRUD(t *testing.T) {
	database := setupTestDB(t)
	svc := NewWebhookService(database)

	// Create
	resp, err := svc.CreateWebhook(models.CreateWebhookRequest{
		Name:   "CRUD Test",
		URL:    "https://example.com/hook",
		Events: []string{"agent.created", "agent.deleted"},
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	id := resp.Webhook.ID

	// Get
	webhook, err := svc.GetWebhook(id)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if webhook.Name != "CRUD Test" {
		t.Errorf("expected name 'CRUD Test', got '%s'", webhook.Name)
	}

	// List
	webhooks, err := svc.ListWebhooks()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(webhooks) != 1 {
		t.Errorf("expected 1 webhook, got %d", len(webhooks))
	}

	// Update
	newName := "Updated Name"
	isActive := false
	updated, err := svc.UpdateWebhook(id, models.UpdateWebhookRequest{
		Name:     &newName,
		IsActive: &isActive,
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Name != "Updated Name" {
		t.Errorf("expected updated name, got '%s'", updated.Name)
	}
	if updated.IsActive {
		t.Error("expected webhook to be inactive")
	}

	// Delete
	if err := svc.DeleteWebhook(id); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	webhooks, _ = svc.ListWebhooks()
	if len(webhooks) != 0 {
		t.Errorf("expected 0 webhooks after delete, got %d", len(webhooks))
	}
}

func TestHMACSignature(t *testing.T) {
	payload := `{"event":"test","data":"hello"}`
	secret := "my-secret-key"

	signature := GenerateHMACSignature(payload, secret)

	if signature == "" {
		t.Fatal("expected non-empty signature")
	}

	// Verify correct signature
	if !VerifyHMACSignature(payload, secret, signature) {
		t.Error("expected signature to be valid")
	}

	// Verify wrong signature
	if VerifyHMACSignature(payload, secret, "wrong-signature") {
		t.Error("expected wrong signature to fail verification")
	}

	// Verify wrong secret
	if VerifyHMACSignature(payload, "wrong-secret", signature) {
		t.Error("expected wrong secret to fail verification")
	}

	// Verify tampered payload
	if VerifyHMACSignature(payload+"tampered", secret, signature) {
		t.Error("expected tampered payload to fail verification")
	}
}

func TestWebhookService_TriggerEvent(t *testing.T) {
	database := setupTestDB(t)
	svc := NewWebhookService(database)

	var received int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&received, 1)

		// Verify headers
		if r.Header.Get("X-Webhook-Event") == "" {
			t.Error("missing X-Webhook-Event header")
		}
		if r.Header.Get("X-Webhook-Signature-256") == "" {
			t.Error("missing X-Webhook-Signature-256 header")
		}
		if r.Header.Get("X-Webhook-Delivery") == "" {
			t.Error("missing X-Webhook-Delivery header")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Read and verify body
		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Errorf("failed to unmarshal payload: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a webhook listening for agent.created
	_, err := svc.CreateWebhook(models.CreateWebhookRequest{
		Name:   "Test Trigger",
		URL:    server.URL,
		Events: []string{"agent.created"},
	})
	if err != nil {
		t.Fatalf("create webhook failed: %v", err)
	}

	// Trigger event
	svc.TriggerEvent("agent.created", map[string]string{
		"agent_id": "test-123",
	})

	// Wait for async delivery
	time.Sleep(500 * time.Millisecond)

	if atomic.LoadInt32(&received) != 1 {
		t.Errorf("expected 1 delivery, got %d", atomic.LoadInt32(&received))
	}
}

func TestWebhookService_RetryOnFailure(t *testing.T) {
	database := setupTestDB(t)
	svc := NewWebhookService(database)

	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		if count <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	maxRetries := 3
	retryBackoff := 1
	_, err := svc.CreateWebhook(models.CreateWebhookRequest{
		Name:             "Retry Test",
		URL:              server.URL,
		Events:           []string{"agent.created"},
		MaxRetries:       &maxRetries,
		RetryBackoffBase: &retryBackoff,
	})
	if err != nil {
		t.Fatalf("create webhook failed: %v", err)
	}

	svc.TriggerEvent("agent.created", map[string]string{"test": "retry"})

	// Wait for retries (1s + 1s backoff + delivery)
	time.Sleep(5 * time.Second)

	totalAttempts := atomic.LoadInt32(&attempts)
	if totalAttempts < 3 {
		t.Errorf("expected at least 3 attempts, got %d", totalAttempts)
	}
}

func TestWebhookService_NoTriggerForInactiveWebhook(t *testing.T) {
	database := setupTestDB(t)
	svc := NewWebhookService(database)

	var received int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&received, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	resp, err := svc.CreateWebhook(models.CreateWebhookRequest{
		Name:   "Inactive Test",
		URL:    server.URL,
		Events: []string{"agent.created"},
	})
	if err != nil {
		t.Fatalf("create webhook failed: %v", err)
	}

	// Disable webhook
	isActive := false
	_, err = svc.UpdateWebhook(resp.Webhook.ID, models.UpdateWebhookRequest{
		IsActive: &isActive,
	})
	if err != nil {
		t.Fatalf("update webhook failed: %v", err)
	}

	svc.TriggerEvent("agent.created", map[string]string{"test": "inactive"})

	time.Sleep(500 * time.Millisecond)

	if atomic.LoadInt32(&received) != 0 {
		t.Error("expected no deliveries for inactive webhook")
	}
}

func TestWebhookService_EventFiltering(t *testing.T) {
	database := setupTestDB(t)
	svc := NewWebhookService(database)

	var received int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&received, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_, err := svc.CreateWebhook(models.CreateWebhookRequest{
		Name:   "Filter Test",
		URL:    server.URL,
		Events: []string{"agent.created"},
	})
	if err != nil {
		t.Fatalf("create webhook failed: %v", err)
	}

	// Trigger non-matching event
	svc.TriggerEvent("agent.deleted", map[string]string{"test": "filter"})

	time.Sleep(500 * time.Millisecond)

	if atomic.LoadInt32(&received) != 0 {
		t.Error("expected no deliveries for non-matching event")
	}
}

func TestWebhookService_TestWebhook(t *testing.T) {
	database := setupTestDB(t)
	svc := NewWebhookService(database)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	resp, err := svc.CreateWebhook(models.CreateWebhookRequest{
		Name:   "Test Webhook",
		URL:    server.URL,
		Events: []string{"webhook.test"},
	})
	if err != nil {
		t.Fatalf("create webhook failed: %v", err)
	}

	result, err := svc.TestWebhook(resp.Webhook.ID, models.TestWebhookRequest{})
	if err != nil {
		t.Fatalf("test webhook failed: %v", err)
	}
	if !result.Success {
		t.Errorf("expected test to succeed, got error: %s", result.Error)
	}
	if result.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", result.StatusCode)
	}
}

func TestComputeBackoff(t *testing.T) {
	tests := []struct {
		attempt  int
		base     int
		expected time.Duration
	}{
		{0, 2, 1 * time.Second},
		{1, 2, 2 * time.Second},
		{2, 2, 4 * time.Second},
		{3, 2, 8 * time.Second},
		{4, 2, 16 * time.Second},
		{1, 3, 3 * time.Second},
		{2, 3, 9 * time.Second},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("attempt=%d,base=%d", tt.attempt, tt.base), func(t *testing.T) {
			got := computeBackoff(tt.attempt, tt.base)
			if got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestComputeBackoff_Cap(t *testing.T) {
	// Very high attempt should be capped at 3600s (1 hour)
	got := computeBackoff(20, 2)
	if got > 3600*time.Second {
		t.Errorf("expected backoff capped at 3600s, got %v", got)
	}
}

func TestValidateWebhookURL(t *testing.T) {
	tests := []struct {
		url     string
		wantErr bool
	}{
		{"https://example.com/hook", false},
		{"http://localhost:8080/hook", false},
		{"ftp://example.com", true},
		{"not-a-url", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			err := validateWebhookURL(tt.url)
			if tt.wantErr && err == nil {
				t.Error("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateEvents(t *testing.T) {
	tests := []struct {
		name    string
		events  []string
		wantErr bool
	}{
		{"valid single", []string{"agent.created"}, false},
		{"valid multiple", []string{"agent.created", "token.issued"}, false},
		{"wildcard", []string{"*"}, false},
		{"unknown", []string{"foo.bar"}, true},
		{"mixed valid and invalid", []string{"agent.created", "foo.bar"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEvents(tt.events)
			if tt.wantErr && err == nil {
				t.Error("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestWebhookService_DeliveryHistory(t *testing.T) {
	database := setupTestDB(t)
	svc := NewWebhookService(database)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	resp, err := svc.CreateWebhook(models.CreateWebhookRequest{
		Name:   "History Test",
		URL:    server.URL,
		Events: []string{"agent.created"},
	})
	if err != nil {
		t.Fatalf("create webhook failed: %v", err)
	}

	svc.TriggerEvent("agent.created", map[string]string{"test": "history"})
	time.Sleep(500 * time.Millisecond)

	deliveries, err := svc.GetDeliveries(resp.Webhook.ID)
	if err != nil {
		t.Fatalf("get deliveries failed: %v", err)
	}
	if len(deliveries) != 1 {
		t.Errorf("expected 1 delivery, got %d", len(deliveries))
	}
	if len(deliveries) > 0 && deliveries[0].Status != DeliveryStatusDelivered {
		t.Errorf("expected status 'delivered', got '%s'", deliveries[0].Status)
	}
}
