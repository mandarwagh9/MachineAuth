package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"machineauth/internal/db"
	"machineauth/internal/models"
)

// Webhook event type constants
const (
	EventAgentCreated            = "agent.created"
	EventAgentDeleted            = "agent.deleted"
	EventAgentUpdated            = "agent.updated"
	EventAgentCredentialsRotated = "agent.credentials_rotated"
	EventTokenIssued             = "token.issued"
	EventTokenValidationSuccess  = "token.validation_success"
	EventTokenValidationFailed   = "token.validation_failed"
	EventWebhookCreated          = "webhook.created"
	EventWebhookUpdated          = "webhook.updated"
	EventWebhookDeleted          = "webhook.deleted"
	EventWebhookTest             = "webhook.test"
)

// Delivery status constants
const (
	DeliveryStatusPending   = "pending"
	DeliveryStatusDelivered = "delivered"
	DeliveryStatusFailed    = "failed"
	DeliveryStatusRetrying  = "retrying"
	DeliveryStatusDead      = "dead"
)

// Default configuration
const (
	DefaultMaxRetries       = 10
	DefaultRetryBackoffBase = 2
	DefaultDeliveryTimeout  = 10 * time.Second
	MaxConsecutiveFails     = 10
)

// AllWebhookEvents returns all known event types
func AllWebhookEvents() []string {
	return []string{
		EventAgentCreated,
		EventAgentDeleted,
		EventAgentUpdated,
		EventAgentCredentialsRotated,
		EventTokenIssued,
		EventTokenValidationSuccess,
		EventTokenValidationFailed,
		EventWebhookCreated,
		EventWebhookUpdated,
		EventWebhookDeleted,
		EventWebhookTest,
	}
}

type WebhookService struct {
	db         db.Database
	httpClient *http.Client
}

func NewWebhookService(database db.Database) *WebhookService {
	return &WebhookService{
		db: database,
		httpClient: &http.Client{
			Timeout: DefaultDeliveryTimeout,
		},
	}
}

// CreateWebhook creates a new webhook configuration
func (s *WebhookService) CreateWebhook(req models.CreateWebhookRequest) (*models.CreateWebhookResponse, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("webhook name is required")
	}
	if req.URL == "" {
		return nil, fmt.Errorf("webhook URL is required")
	}
	if err := validateWebhookURL(req.URL); err != nil {
		return nil, err
	}
	if len(req.Events) == 0 {
		return nil, fmt.Errorf("at least one event is required")
	}
	if err := validateEvents(req.Events); err != nil {
		return nil, err
	}

	id := uuid.New()
	secret := generateSecureSecret(32)

	maxRetries := DefaultMaxRetries
	if req.MaxRetries != nil && *req.MaxRetries >= 0 {
		maxRetries = *req.MaxRetries
	}

	retryBackoffBase := DefaultRetryBackoffBase
	if req.RetryBackoffBase != nil && *req.RetryBackoffBase >= 1 {
		retryBackoffBase = *req.RetryBackoffBase
	}

	now := time.Now()
	webhook := db.WebhookConfig{
		ID:               id.String(),
		Name:             req.Name,
		URL:              req.URL,
		Secret:           secret,
		Events:           req.Events,
		IsActive:         true,
		MaxRetries:       maxRetries,
		RetryBackoffBase: retryBackoffBase,
		CreatedAt:        now,
		UpdatedAt:        now,
		ConsecutiveFails: 0,
	}

	if err := s.db.CreateWebhook(webhook); err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}

	return &models.CreateWebhookResponse{
		Webhook: toModelWebhook(&webhook),
		Secret:  secret,
	}, nil
}

// GetWebhook retrieves a webhook by ID
func (s *WebhookService) GetWebhook(id uuid.UUID) (*models.WebhookConfig, error) {
	webhook, err := s.db.GetWebhook(id.String())
	if err != nil {
		return nil, fmt.Errorf("webhook not found: %w", err)
	}
	result := toModelWebhook(webhook)
	return &result, nil
}

// ListWebhooks returns all webhooks
func (s *WebhookService) ListWebhooks() ([]models.WebhookConfig, error) {
	webhooks, err := s.db.ListWebhooks()
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}

	result := make([]models.WebhookConfig, len(webhooks))
	for i, w := range webhooks {
		result[i] = toModelWebhook(&w)
	}
	return result, nil
}

// UpdateWebhook updates a webhook configuration
func (s *WebhookService) UpdateWebhook(id uuid.UUID, req models.UpdateWebhookRequest) (*models.WebhookConfig, error) {
	if req.URL != nil {
		if err := validateWebhookURL(*req.URL); err != nil {
			return nil, err
		}
	}
	if req.Events != nil {
		if err := validateEvents(req.Events); err != nil {
			return nil, err
		}
	}

	var result *models.WebhookConfig
	err := s.db.UpdateWebhook(id.String(), func(wh *db.WebhookConfig) error {
		if req.Name != nil {
			wh.Name = *req.Name
		}
		if req.URL != nil {
			wh.URL = *req.URL
		}
		if req.Events != nil {
			wh.Events = req.Events
		}
		if req.IsActive != nil {
			wh.IsActive = *req.IsActive
			if *req.IsActive {
				wh.ConsecutiveFails = 0
			}
		}
		if req.MaxRetries != nil {
			wh.MaxRetries = *req.MaxRetries
		}
		if req.RetryBackoffBase != nil {
			wh.RetryBackoffBase = *req.RetryBackoffBase
		}
		r := toModelWebhook(wh)
		result = &r
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update webhook: %w", err)
	}
	return result, nil
}

// DeleteWebhook removes a webhook configuration
func (s *WebhookService) DeleteWebhook(id uuid.UUID) error {
	return s.db.DeleteWebhook(id.String())
}

// TriggerEvent fires webhooks for a given event
func (s *WebhookService) TriggerEvent(event string, payload interface{}) {
	webhooks, err := s.db.ListActiveWebhooksForEvent(event)
	if err != nil {
		log.Printf("failed to list webhooks for event %s: %v", event, err)
		return
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("failed to marshal webhook payload for event %s: %v", event, err)
		return
	}

	for _, wh := range webhooks {
		deliveryID := uuid.New()
		now := time.Now()
		delivery := db.WebhookDelivery{
			ID:              deliveryID.String(),
			WebhookConfigID: wh.ID,
			Event:           event,
			Payload:         string(payloadBytes),
			Status:          DeliveryStatusPending,
			Attempts:        0,
			CreatedAt:       now,
		}

		if err := s.db.AddWebhookDelivery(delivery); err != nil {
			log.Printf("failed to create webhook delivery for webhook %s: %v", wh.ID, err)
			continue
		}

		// Dispatch async delivery
		go s.processDelivery(deliveryID.String(), wh, delivery)
	}
}

// processDelivery attempts to deliver a webhook payload
func (s *WebhookService) processDelivery(deliveryID string, wh db.WebhookConfig, delivery db.WebhookDelivery) {
	for attempt := 0; attempt <= wh.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := computeBackoff(attempt, wh.RetryBackoffBase)
			time.Sleep(backoff)
		}

		now := time.Now()
		err := s.sendWebhook(wh, delivery)

		if err == nil {
			// Success
			s.db.UpdateWebhookDelivery(deliveryID, func(d *db.WebhookDelivery) error {
				d.Status = DeliveryStatusDelivered
				d.Attempts = attempt + 1
				d.LastAttemptAt = &now
				d.LastError = ""
				return nil
			})

			// Reset consecutive fails
			s.db.UpdateWebhook(wh.ID, func(w *db.WebhookConfig) error {
				w.ConsecutiveFails = 0
				return nil
			})

			log.Printf("webhook delivered: id=%s event=%s webhook=%s attempt=%d", deliveryID, delivery.Event, wh.ID, attempt+1)
			return
		}

		// Update delivery with error
		errMsg := err.Error()
		s.db.UpdateWebhookDelivery(deliveryID, func(d *db.WebhookDelivery) error {
			d.Attempts = attempt + 1
			d.LastAttemptAt = &now
			d.LastError = errMsg
			if attempt < wh.MaxRetries {
				d.Status = DeliveryStatusRetrying
				nextRetry := now.Add(computeBackoff(attempt+1, wh.RetryBackoffBase))
				d.NextRetryAt = &nextRetry
			}
			return nil
		})

		log.Printf("webhook delivery attempt failed: id=%s event=%s webhook=%s attempt=%d err=%v", deliveryID, delivery.Event, wh.ID, attempt+1, err)
	}

	// All retries exhausted
	now := time.Now()
	s.db.UpdateWebhookDelivery(deliveryID, func(d *db.WebhookDelivery) error {
		d.Status = DeliveryStatusDead
		d.LastAttemptAt = &now
		return nil
	})

	// Increment consecutive fails and potentially auto-disable
	s.db.UpdateWebhook(wh.ID, func(w *db.WebhookConfig) error {
		w.ConsecutiveFails++
		if w.ConsecutiveFails >= MaxConsecutiveFails {
			w.IsActive = false
			log.Printf("webhook auto-disabled after %d consecutive failures: id=%s", w.ConsecutiveFails, wh.ID)
		}
		return nil
	})

	log.Printf("webhook delivery dead-lettered: id=%s event=%s webhook=%s", deliveryID, delivery.Event, wh.ID)
}

// sendWebhook performs the HTTP POST to the webhook URL
func (s *WebhookService) sendWebhook(wh db.WebhookConfig, delivery db.WebhookDelivery) error {
	body := delivery.Payload
	signature := GenerateHMACSignature(body, wh.Secret)

	req, err := http.NewRequest(http.MethodPost, wh.URL, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Event", delivery.Event)
	req.Header.Set("X-Webhook-Delivery", delivery.ID)
	req.Header.Set("X-Webhook-Signature-256", "sha256="+signature)
	req.Header.Set("User-Agent", "AgentAuth-Webhook/1.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read body for error reporting
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
}

// TestWebhook sends a test delivery to a webhook endpoint
func (s *WebhookService) TestWebhook(id uuid.UUID, req models.TestWebhookRequest) (*models.TestWebhookResponse, error) {
	webhook, err := s.db.GetWebhook(id.String())
	if err != nil {
		return nil, fmt.Errorf("webhook not found: %w", err)
	}

	event := req.Event
	if event == "" {
		event = EventWebhookTest
	}

	payload := req.Payload
	if payload == "" {
		testPayload := map[string]interface{}{
			"event":      event,
			"webhook_id": id.String(),
			"test":       true,
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		}
		payloadBytes, _ := json.Marshal(testPayload)
		payload = string(payloadBytes)
	}

	signature := GenerateHMACSignature(payload, webhook.Secret)

	httpReq, err := http.NewRequest(http.MethodPost, webhook.URL, strings.NewReader(payload))
	if err != nil {
		return &models.TestWebhookResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to create request: %v", err),
		}, nil
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Webhook-Event", event)
	httpReq.Header.Set("X-Webhook-Delivery", "test-"+uuid.New().String())
	httpReq.Header.Set("X-Webhook-Signature-256", "sha256="+signature)
	httpReq.Header.Set("User-Agent", "AgentAuth-Webhook/1.0")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return &models.TestWebhookResponse{
			Success: false,
			Error:   fmt.Sprintf("request failed: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	// Update last tested timestamp
	now := time.Now()
	s.db.UpdateWebhook(id.String(), func(wh *db.WebhookConfig) error {
		wh.LastTestedAt = &now
		return nil
	})

	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	result := &models.TestWebhookResponse{
		Success:    success,
		StatusCode: resp.StatusCode,
	}
	if !success {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		result.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return result, nil
}

// GetDeliveries returns delivery history for a webhook
func (s *WebhookService) GetDeliveries(webhookID uuid.UUID) ([]models.WebhookDelivery, error) {
	deliveries, err := s.db.ListWebhookDeliveries(webhookID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to list deliveries: %w", err)
	}

	result := make([]models.WebhookDelivery, len(deliveries))
	for i, d := range deliveries {
		result[i] = toModelDelivery(&d)
	}
	return result, nil
}

// GetDelivery returns a single delivery
func (s *WebhookService) GetDelivery(id uuid.UUID) (*models.WebhookDelivery, error) {
	delivery, err := s.db.GetWebhookDelivery(id.String())
	if err != nil {
		return nil, fmt.Errorf("delivery not found: %w", err)
	}
	result := toModelDelivery(delivery)
	return &result, nil
}

// HMAC Signature methods

// GenerateHMACSignature creates HMAC-SHA256 signature for payload verification
func GenerateHMACSignature(payload, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyHMACSignature verifies a payload against its HMAC-SHA256 signature
func VerifyHMACSignature(payload, secret, signature string) bool {
	expected := GenerateHMACSignature(payload, secret)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// Helper functions

func validateWebhookURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("URL must use http or https scheme")
	}
	if u.Host == "" {
		return fmt.Errorf("URL must have a host")
	}
	return nil
}

func validateEvents(events []string) error {
	known := make(map[string]bool)
	for _, e := range AllWebhookEvents() {
		known[e] = true
	}
	known["*"] = true

	for _, e := range events {
		if !known[e] {
			return fmt.Errorf("unknown event type: %s", e)
		}
	}
	return nil
}

func computeBackoff(attempt, base int) time.Duration {
	if base < 1 {
		base = 2
	}
	seconds := 1
	for i := 0; i < attempt; i++ {
		seconds *= base
		if seconds > 3600 {
			seconds = 3600
			break
		}
	}
	return time.Duration(seconds) * time.Second
}

func toModelWebhook(w *db.WebhookConfig) models.WebhookConfig {
	return models.WebhookConfig{
		ID:               uuid.MustParse(w.ID),
		OrganizationID:   w.OrganizationID,
		TeamID:           w.TeamID,
		Name:             w.Name,
		URL:              w.URL,
		Events:           w.Events,
		IsActive:         w.IsActive,
		MaxRetries:       w.MaxRetries,
		RetryBackoffBase: w.RetryBackoffBase,
		CreatedAt:        w.CreatedAt,
		UpdatedAt:        w.UpdatedAt,
		LastTestedAt:     w.LastTestedAt,
		ConsecutiveFails: w.ConsecutiveFails,
	}
}

func toModelDelivery(d *db.WebhookDelivery) models.WebhookDelivery {
	return models.WebhookDelivery{
		ID:              uuid.MustParse(d.ID),
		WebhookConfigID: uuid.MustParse(d.WebhookConfigID),
		Event:           d.Event,
		Payload:         d.Payload,
		Headers:         d.Headers,
		Status:          d.Status,
		Attempts:        d.Attempts,
		LastAttemptAt:   d.LastAttemptAt,
		LastError:       d.LastError,
		NextRetryAt:     d.NextRetryAt,
		CreatedAt:       d.CreatedAt,
	}
}
