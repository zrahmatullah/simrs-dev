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

type PatientHandler struct {
	patientUsecase *usecase.PatientUsecase
}

func NewPatientHandler(patientUsecase *usecase.PatientUsecase) *PatientHandler {
	return &PatientHandler{patientUsecase: patientUsecase}
}

func (h *PatientHandler) GetPatients(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	patients, total, err := h.patientUsecase.GetPatients(search, page, limit)
	if err != nil {
		RespondInternalError(w, err)
		return
	}

	RespondMeta(w, "Daftar pasien berhasil diambil", patients, page, limit, total)
}

func (h *PatientHandler) GetPatientByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	patient, err := h.patientUsecase.GetPatientByID(id)
	if err != nil {
		RespondNotFound(w, "Data pasien tidak ditemukan")
		return
	}
	RespondSuccess(w, "Data pasien berhasil diambil", patient)
}

func (h *PatientHandler) GetPatientByNoRM(w http.ResponseWriter, r *http.Request) {
	noRM := chi.URLParam(r, "no_rm")
	patient, err := h.patientUsecase.GetPatientByNoRM(noRM)
	if err != nil {
		RespondNotFound(w, "Pasien dengan No. RM tersebut tidak ditemukan")
		return
	}
	RespondSuccess(w, "Data pasien berhasil diambil", patient)
}

func (h *PatientHandler) CreatePatient(w http.ResponseWriter, r *http.Request) {
	var req domain.CreatePatientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	patient, err := h.patientUsecase.CreatePatient(req, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondCreated(w, "Pasien baru berhasil didaftarkan", patient)
}

func (h *PatientHandler) UpdatePatient(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req domain.UpdatePatientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	err := h.patientUsecase.UpdatePatient(id, req, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondSuccess(w, "Data pasien berhasil diperbarui", nil)
}

func (h *PatientHandler) DeletePatient(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.GetCurrentUser(r)

	err := h.patientUsecase.DeletePatient(id, claims.UserID, claims.Username)
	if err != nil {
		RespondInternalError(w, err)
		return
	}

	RespondSuccess(w, "Data pasien berhasil dihapus", nil)
}
