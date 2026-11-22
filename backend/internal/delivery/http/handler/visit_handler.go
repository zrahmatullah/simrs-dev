package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"simrs-asoy/backend/internal/domain"
	"simrs-asoy/backend/internal/middleware"
	"simrs-asoy/backend/internal/usecase"
)

type VisitHandler struct {
	visitUsecase *usecase.VisitUsecase
}

func NewVisitHandler(visitUsecase *usecase.VisitUsecase) *VisitHandler {
	return &VisitHandler{visitUsecase: visitUsecase}
}

func (h *VisitHandler) GetVisits(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	status := r.URL.Query().Get("status")
	poliID, _ := strconv.Atoi(r.URL.Query().Get("poli_id"))
	search := r.URL.Query().Get("search")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	visits, total, err := h.visitUsecase.GetVisits(date, status, poliID, search, page, limit)
	if err != nil {
		RespondInternalError(w, err)
		return
	}

	RespondMeta(w, "Daftar kunjungan / antrean berhasil diambil", visits, page, limit, total)
}

func (h *VisitHandler) GetVisitByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	visit, err := h.visitUsecase.GetVisitByID(id)
	if err != nil {
		RespondNotFound(w, "Data kunjungan tidak ditemukan")
		return
	}
	RespondSuccess(w, "Data kunjungan berhasil diambil", visit)
}

func (h *VisitHandler) CreateVisit(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateVisitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	visit, err := h.visitUsecase.CreateVisit(req, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondCreated(w, "Pendaftaran kunjungan rawat jalan berhasil dibuat", visit)
}

func (h *VisitHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req domain.UpdateVisitStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	err := h.visitUsecase.UpdateVisitStatus(id, req.Status, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondSuccess(w, "Status kunjungan berhasil diperbarui", nil)
}

func (h *VisitHandler) GetQueueMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.visitUsecase.GetQueueMetrics()
	if err != nil {
		RespondInternalError(w, err)
		return
	}
	RespondSuccess(w, "Metrik antrean real-time hari ini berhasil diambil", metrics)
}
