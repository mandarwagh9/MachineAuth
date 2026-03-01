package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"agentauth/internal/models"
	"agentauth/internal/services"
)

type WebhookHandler struct {
	webhookService *services.WebhookService
	auditService   *services.AuditService
}

func NewWebhookHandler(webhookService *services.WebhookService, auditService *services.AuditService) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
		auditService:   auditService,
	}
}

// ListAndCreate handles GET /api/webhooks and POST /api/webhooks
func (h *WebhookHandler) ListAndCreate(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listWebhooks(w, r)
	case http.MethodPost:
		h.createWebhook(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleWebhook handles /api/webhooks/:id routes
func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if !strings.HasPrefix(path, "/api/webhooks/") {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	remaining := strings.TrimPrefix(path, "/api/webhooks/")
	parts := strings.SplitN(remaining, "/", 2)

	idStr := parts[0]
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid webhook ID", http.StatusBadRequest)
		return
	}

	// Check for sub-routes: /api/webhooks/:id/test, /api/webhooks/:id/deliveries
	if len(parts) > 1 {
		subRoute := parts[1]
		switch subRoute {
		case "test":
			h.testWebhook(w, r, id)
			return
		case "deliveries":
			h.listDeliveries(w, r, id)
			return
		}
		// Check for /api/webhooks/:id/deliveries/:deliveryId
		if strings.HasPrefix(subRoute, "deliveries/") {
			deliveryIDStr := strings.TrimPrefix(subRoute, "deliveries/")
			deliveryID, err := uuid.Parse(deliveryIDStr)
			if err != nil {
				http.Error(w, "invalid delivery ID", http.StatusBadRequest)
				return
			}
			h.getDelivery(w, r, deliveryID)
			return
		}

		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getWebhook(w, r, id)
	case http.MethodPut:
		h.updateWebhook(w, r, id)
	case http.MethodDelete:
		h.deleteWebhook(w, r, id)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleEvents returns the list of available webhook events
func (h *WebhookHandler) HandleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	events := services.AllWebhookEvents()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": events,
	})
}

func (h *WebhookHandler) listWebhooks(w http.ResponseWriter, r *http.Request) {
	webhooks, err := h.webhookService.ListWebhooks()
	if err != nil {
		log.Printf("failed to list webhooks: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if webhooks == nil {
		webhooks = []models.WebhookConfig{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.WebhooksListResponse{Webhooks: webhooks})
}

func (h *WebhookHandler) createWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req models.CreateWebhookRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	resp, err := h.webhookService.CreateWebhook(req)
	if err != nil {
		log.Printf("failed to create webhook: %v", err)
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.auditService.LogWebhook(services.EventWebhookCreated, resp.Webhook.ID, r.RemoteAddr, r.UserAgent())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *WebhookHandler) getWebhook(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	webhook, err := h.webhookService.GetWebhook(id)
	if err != nil {
		log.Printf("failed to get webhook %s: %v", id, err)
		http.Error(w, "webhook not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.WebhookResponse{Webhook: *webhook})
}

func (h *WebhookHandler) updateWebhook(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req models.UpdateWebhookRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	webhook, err := h.webhookService.UpdateWebhook(id, req)
	if err != nil {
		log.Printf("failed to update webhook %s: %v", id, err)
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.auditService.LogWebhook(services.EventWebhookUpdated, id, r.RemoteAddr, r.UserAgent())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.WebhookResponse{Webhook: *webhook})
}

func (h *WebhookHandler) deleteWebhook(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	if err := h.webhookService.DeleteWebhook(id); err != nil {
		log.Printf("failed to delete webhook %s: %v", id, err)
		http.Error(w, "failed to delete webhook", http.StatusInternalServerError)
		return
	}

	h.auditService.LogWebhook(services.EventWebhookDeleted, id, r.RemoteAddr, r.UserAgent())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func (h *WebhookHandler) testWebhook(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.TestWebhookRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(body) > 0 {
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
	}

	result, err := h.webhookService.TestWebhook(id, req)
	if err != nil {
		log.Printf("failed to test webhook %s: %v", id, err)
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *WebhookHandler) listDeliveries(w http.ResponseWriter, r *http.Request, webhookID uuid.UUID) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	deliveries, err := h.webhookService.GetDeliveries(webhookID)
	if err != nil {
		log.Printf("failed to list deliveries for webhook %s: %v", webhookID, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if deliveries == nil {
		deliveries = []models.WebhookDelivery{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.WebhookDeliveriesListResponse{Deliveries: deliveries})
}

func (h *WebhookHandler) getDelivery(w http.ResponseWriter, r *http.Request, deliveryID uuid.UUID) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	delivery, err := h.webhookService.GetDelivery(deliveryID)
	if err != nil {
		log.Printf("failed to get delivery %s: %v", deliveryID, err)
		http.Error(w, "delivery not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.WebhookDeliveryResponse{Delivery: *delivery})
}

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(models.ErrorResponse{
		Error:            "invalid_request",
		ErrorDescription: message,
	})
}
