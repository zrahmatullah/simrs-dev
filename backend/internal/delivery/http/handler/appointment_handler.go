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

type AppointmentHandler struct {
	aptUsecase *usecase.AppointmentUsecase
}

func NewAppointmentHandler(aptUsecase *usecase.AppointmentUsecase) *AppointmentHandler {
	return &AppointmentHandler{aptUsecase: aptUsecase}
}

func (h *AppointmentHandler) GetAppointments(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	status := r.URL.Query().Get("status")
	doctorID := r.URL.Query().Get("doctor_id")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	list, total, err := h.aptUsecase.GetAppointments(date, status, doctorID, page, limit)
	if err != nil {
		RespondInternalError(w, err)
		return
	}

	RespondMeta(w, "Daftar janji temu (appointment) berhasil diambil", list, page, limit, total)
}

func (h *AppointmentHandler) GetAppointmentByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	apt, err := h.aptUsecase.GetAppointmentByID(id)
	if err != nil {
		RespondNotFound(w, "Janji temu tidak ditemukan")
		return
	}
	RespondSuccess(w, "Data janji temu berhasil diambil", apt)
}

func (h *AppointmentHandler) CreateAppointment(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateAppointmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	apt, err := h.aptUsecase.CreateAppointment(req, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondCreated(w, "Janji temu berhasil dijadwalkan", apt)
}

func (h *AppointmentHandler) Reschedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req domain.RescheduleAppointmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	err := h.aptUsecase.Reschedule(id, req, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondSuccess(w, "Janji temu berhasil dijadwalkan ulang", nil)
}

func (h *AppointmentHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Notes string `json:"notes"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	claims := middleware.GetCurrentUser(r)
	err := h.aptUsecase.Cancel(id, req.Notes, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondSuccess(w, "Janji temu berhasil dibatalkan", nil)
}

func (h *AppointmentHandler) CheckIn(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.GetCurrentUser(r)

	visit, err := h.aptUsecase.CheckIn(id, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondSuccess(w, "Pasien berhasil check-in dan masuk ke antrean poli", visit)
}

// -------------------------------------------------------------
// Doctor Schedules & Leaves Handlers (Fase 1)
// -------------------------------------------------------------

func (h *AppointmentHandler) GetDoctorSchedules(w http.ResponseWriter, r *http.Request) {
	doctorID := r.URL.Query().Get("doctor_id")
	poliID, _ := strconv.Atoi(r.URL.Query().Get("poli_id"))
	dayOfWeek := -1
	if dayStr := r.URL.Query().Get("day_of_week"); dayStr != "" {
		dayOfWeek, _ = strconv.Atoi(dayStr)
	}

	list, err := h.aptUsecase.GetDoctorSchedules(doctorID, poliID, dayOfWeek)
	if err != nil {
		RespondInternalError(w, err)
		return
	}
	RespondSuccess(w, "Jadwal dokter berhasil diambil", list)
}

func (h *AppointmentHandler) CreateDoctorSchedule(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateDoctorScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	s, err := h.aptUsecase.CreateDoctorSchedule(req, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}
	RespondCreated(w, "Jadwal dokter berhasil ditambahkan", s)
}

func (h *AppointmentHandler) DeleteDoctorSchedule(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		RespondBadRequest(w, "ID jadwal tidak valid", nil)
		return
	}

	claims := middleware.GetCurrentUser(r)
	err = h.aptUsecase.DeleteDoctorSchedule(id, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}
	RespondSuccess(w, "Jadwal dokter berhasil dinonaktifkan", nil)
}

func (h *AppointmentHandler) GetDoctorLeaves(w http.ResponseWriter, r *http.Request) {
	doctorID := r.URL.Query().Get("doctor_id")
	date := r.URL.Query().Get("date")

	list, err := h.aptUsecase.GetDoctorLeaves(doctorID, date)
	if err != nil {
		RespondInternalError(w, err)
		return
	}
	RespondSuccess(w, "Data cuti dokter berhasil diambil", list)
}

func (h *AppointmentHandler) CreateDoctorLeave(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateDoctorLeaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	l, err := h.aptUsecase.CreateDoctorLeave(req, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}
	RespondCreated(w, "Pengajuan cuti dokter berhasil disimpan", l)
}
