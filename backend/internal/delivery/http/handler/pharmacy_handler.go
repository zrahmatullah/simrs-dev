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

type PharmacyHandler struct {
	pharmacyUsecase *usecase.PharmacyUsecase
}

func NewPharmacyHandler(pharmacyUsecase *usecase.PharmacyUsecase) *PharmacyHandler {
	return &PharmacyHandler{pharmacyUsecase: pharmacyUsecase}
}

func (h *PharmacyHandler) GetPrescriptions(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	list, total, err := h.pharmacyUsecase.GetPrescriptions(status, search, page, limit)
	if err != nil {
		RespondInternalError(w, err)
		return
	}

	RespondMeta(w, "Daftar antrean e-Resep berhasil diambil", list, page, limit, total)
}

func (h *PharmacyHandler) GetPrescriptionByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	p, err := h.pharmacyUsecase.GetPrescriptionByID(id)
	if err != nil {
		RespondNotFound(w, "e-Resep tidak ditemukan")
		return
	}
	RespondSuccess(w, "Data e-Resep berhasil diambil", p)
}

func (h *PharmacyHandler) GetPrescriptionByVisit(w http.ResponseWriter, r *http.Request) {
	visitID := chi.URLParam(r, "visit_id")
	p, err := h.pharmacyUsecase.GetPrescriptionByVisit(visitID)
	if err != nil {
		RespondNotFound(w, "e-Resep untuk kunjungan ini belum ada")
		return
	}
	RespondSuccess(w, "Data e-Resep kunjungan berhasil diambil", p)
}

func (h *PharmacyHandler) CreatePrescription(w http.ResponseWriter, r *http.Request) {
	var req domain.CreatePrescriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	p, err := h.pharmacyUsecase.CreatePrescription(req, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondCreated(w, "e-Resep berhasil dibuat dan dikirim ke instalasi farmasi", p)
}

func (h *PharmacyHandler) ReviewPrescription(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req domain.ReviewPrescriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	err := h.pharmacyUsecase.ReviewPrescription(id, req, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondSuccess(w, "Telaah resep farmasi berhasil disimpan", nil)
}

func (h *PharmacyHandler) DispensePrescription(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.GetCurrentUser(r)

	err := h.pharmacyUsecase.DispensePrescription(id, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondSuccess(w, "Obat berhasil diserahkan dan stok depo farmasi otomatis terpotong", nil)
}

// -------------------------------------------------------------
// Retur Obat Pasien Handlers (Fase 1)
// -------------------------------------------------------------

func (h *PharmacyHandler) CreateMedicineReturn(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateMedicineReturnRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	ret, err := h.pharmacyUsecase.CreateMedicineReturn(req, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}
	RespondCreated(w, "Retur obat pasien berhasil diproses dan stok depo otomatis dikembalikan", ret)
}

func (h *PharmacyHandler) GetMedicineReturns(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	list, total, err := h.pharmacyUsecase.GetMedicineReturns(search, page, limit)
	if err != nil {
		RespondInternalError(w, err)
		return
	}
	RespondMeta(w, "Daftar retur obat berhasil diambil", list, page, limit, total)
}
