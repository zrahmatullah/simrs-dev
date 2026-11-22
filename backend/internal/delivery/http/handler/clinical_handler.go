package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"simrs-asoy/backend/internal/domain"
	"simrs-asoy/backend/internal/middleware"
	"simrs-asoy/backend/internal/usecase"
)

type ClinicalHandler struct {
	clinicalUsecase *usecase.ClinicalUsecase
}

func NewClinicalHandler(clinicalUsecase *usecase.ClinicalUsecase) *ClinicalHandler {
	return &ClinicalHandler{clinicalUsecase: clinicalUsecase}
}

// 1. Vital Signs TTV (Nurse)
func (h *ClinicalHandler) CreateVitalSign(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateVitalSignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	vs, err := h.clinicalUsecase.CreateVitalSign(req, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondCreated(w, "Tanda-Tanda Vital (TTV) dan Triase berhasil disimpan", vs)
}

func (h *ClinicalHandler) GetVitalSignsByVisit(w http.ResponseWriter, r *http.Request) {
	visitID := chi.URLParam(r, "visit_id")
	list, err := h.clinicalUsecase.GetVitalSignsByVisit(visitID)
	if err != nil {
		RespondInternalError(w, err)
		return
	}
	RespondSuccess(w, "Data TTV kunjungan berhasil diambil", list)
}

func (h *ClinicalHandler) GetVitalSignsByPatient(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "patient_id")
	list, err := h.clinicalUsecase.GetVitalSignsByPatient(patientID)
	if err != nil {
		RespondInternalError(w, err)
		return
	}
	RespondSuccess(w, "Riwayat TTV pasien berhasil diambil", list)
}

// 2. Doctor EMR (SOAP)
func (h *ClinicalHandler) CreateMedicalRecord(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateMedicalRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	mr, err := h.clinicalUsecase.CreateMedicalRecord(req, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondCreated(w, "Rekam Medis (EMR SOAP) berhasil disimpan", mr)
}

func (h *ClinicalHandler) GetMedicalRecordByVisit(w http.ResponseWriter, r *http.Request) {
	visitID := chi.URLParam(r, "visit_id")
	mr, err := h.clinicalUsecase.GetMedicalRecordByVisit(visitID)
	if err != nil {
		RespondNotFound(w, "Rekam medis kunjungan ini belum diisi")
		return
	}
	RespondSuccess(w, "Rekam medis kunjungan berhasil diambil", mr)
}

func (h *ClinicalHandler) GetMedicalRecordsByPatient(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "patient_id")
	list, err := h.clinicalUsecase.GetMedicalRecordsByPatient(patientID)
	if err != nil {
		RespondInternalError(w, err)
		return
	}
	RespondSuccess(w, "Riwayat EMR pasien berhasil diambil", list)
}

// 3. Medical Certificate
func (h *ClinicalHandler) CreateMedicalCertificate(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateMedicalCertificateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	mc, err := h.clinicalUsecase.CreateMedicalCertificate(req, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondCreated(w, "Surat keterangan medis berhasil diterbitkan", mc)
}

func (h *ClinicalHandler) GetMedicalCertificateByVisit(w http.ResponseWriter, r *http.Request) {
	visitID := chi.URLParam(r, "visit_id")
	mc, err := h.clinicalUsecase.GetMedicalCertificateByVisit(visitID)
	if err != nil {
		RespondNotFound(w, "Surat keterangan medis tidak ditemukan")
		return
	}
	RespondSuccess(w, "Surat keterangan medis berhasil diambil", mc)
}

// 4. Clinical Procedures in Room
func (h *ClinicalHandler) GetMasterProcedures(w http.ResponseWriter, r *http.Request) {
	procedures, err := h.clinicalUsecase.GetMasterProcedures()
	if err != nil {
		RespondInternalError(w, err)
		return
	}
	RespondSuccess(w, "Daftar master tindakan poli berhasil diambil", procedures)
}

func (h *ClinicalHandler) CreateVisitProcedure(w http.ResponseWriter, r *http.Request) {
	var req domain.AddVisitProcedureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	vp, err := h.clinicalUsecase.CreateVisitProcedure(req, claims.UserID)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}
	RespondCreated(w, "Tindakan berhasil ditambahkan ke rekam pelayanan pasien", vp)
}

func (h *ClinicalHandler) GetVisitProcedures(w http.ResponseWriter, r *http.Request) {
	visitID := chi.URLParam(r, "visit_id")
	list, err := h.clinicalUsecase.GetVisitProcedures(visitID)
	if err != nil {
		RespondInternalError(w, err)
		return
	}
	RespondSuccess(w, "Daftar tindakan kunjungan berhasil diambil", list)
}

// 5. Detailed Nursing Assessment
func (h *ClinicalHandler) CreateNursingAssessment(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateNursingAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	na, err := h.clinicalUsecase.CreateNursingAssessment(req, claims.UserID)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}
	RespondCreated(w, "Asesmen keperawatan detail berhasil disimpan", na)
}

func (h *ClinicalHandler) GetNursingAssessment(w http.ResponseWriter, r *http.Request) {
	visitID := chi.URLParam(r, "visit_id")
	na, err := h.clinicalUsecase.GetNursingAssessment(visitID)
	if err != nil {
		RespondNotFound(w, "Asesmen keperawatan belum diisi untuk kunjungan ini")
		return
	}
	RespondSuccess(w, "Asesmen keperawatan berhasil diambil", na)
}
