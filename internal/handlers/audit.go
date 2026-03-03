package handlers

import (
	"net/http"
	"strconv"
	"time"

	"machineauth/internal/models"
	"machineauth/internal/services"
)

// AuditHandler exposes audit log querying.
type AuditHandler struct {
	auditService *services.AuditService
}

func NewAuditHandler(auditService *services.AuditService) *AuditHandler {
	return &AuditHandler{auditService: auditService}
}

// ListAuditLogs handles GET /api/audit-logs with query params:
//
//	?agent_id=&action=&ip_address=&from=&to=&page=&limit=
func (h *AuditHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	query := models.AuditLogQuery{
		AgentID:   q.Get("agent_id"),
		Action:    q.Get("action"),
		IPAddress: q.Get("ip_address"),
		Page:      page,
		Limit:     limit,
	}

	if fromStr := q.Get("from"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			query.From = &t
		}
	}
	if toStr := q.Get("to"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			query.To = &t
		}
	}

	logs, total, err := h.auditService.ListAuditLogs(query)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "server_error", "failed to query audit logs")
		return
	}

	totalPages := total / limit
	if total%limit != 0 {
		totalPages++
	}

	writeJSON(w, models.AuditLogsListResponse{
		Logs: logs,
		Pagination: &models.Pagination{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	})
}
