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

type AncillaryHandler struct {
	ancillaryUsecase *usecase.AncillaryUsecase
}

func NewAncillaryHandler(ancillaryUsecase *usecase.AncillaryUsecase) *AncillaryHandler {
	return &AncillaryHandler{ancillaryUsecase: ancillaryUsecase}
}

func (h *AncillaryHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	jenisOrder := r.URL.Query().Get("jenis_order")
	search := r.URL.Query().Get("search")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	list, total, err := h.ancillaryUsecase.GetOrders(status, jenisOrder, search, page, limit)
	if err != nil {
		RespondInternalError(w, err)
		return
	}

	RespondMeta(w, "Daftar order pemeriksaan penunjang berhasil diambil", list, page, limit, total)
}

func (h *AncillaryHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	o, err := h.ancillaryUsecase.GetOrderByID(id)
	if err != nil {
		RespondNotFound(w, "Order penunjang tidak ditemukan")
		return
	}
	RespondSuccess(w, "Data order penunjang berhasil diambil", o)
}

func (h *AncillaryHandler) GetOrdersByVisit(w http.ResponseWriter, r *http.Request) {
	visitID := chi.URLParam(r, "visit_id")
	list, err := h.ancillaryUsecase.GetOrdersByVisit(visitID)
	if err != nil {
		RespondInternalError(w, err)
		return
	}
	RespondSuccess(w, "Daftar order penunjang kunjungan berhasil diambil", list)
}

func (h *AncillaryHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateAncillaryOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	order, err := h.ancillaryUsecase.CreateOrder(req, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondCreated(w, "Order pemeriksaan penunjang berhasil dibuat", order)
}

func (h *AncillaryHandler) UpdateResult(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req domain.UpdateAncillaryResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	err := h.ancillaryUsecase.UpdateResult(id, req, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondSuccess(w, "Hasil pemeriksaan penunjang berhasil disimpan dan divalidasi", nil)
}
