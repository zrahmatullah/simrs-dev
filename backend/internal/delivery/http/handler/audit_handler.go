package handler

import (
	"net/http"
	"strconv"

	"simrs-asoy/backend/internal/usecase"
)

type AuditHandler struct {
	auditUsecase *usecase.AuditUsecase
}

func NewAuditHandler(auditUsecase *usecase.AuditUsecase) *AuditHandler {
	return &AuditHandler{auditUsecase: auditUsecase}
}

func (h *AuditHandler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	action := r.URL.Query().Get("action")
	entityType := r.URL.Query().Get("entity_type")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	logs, total, err := h.auditUsecase.GetAuditLogs(search, action, entityType, page, limit)
	if err != nil {
		RespondInternalError(w, err)
		return
	}

	RespondMeta(w, "Jejak audit (audit trail) sistem berhasil diambil", logs, page, limit, total)
}
