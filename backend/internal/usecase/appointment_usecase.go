package usecase

import (
	"errors"

	"simrs-asoy/backend/internal/domain"
	"simrs-asoy/backend/internal/repository"
)

type AppointmentUsecase struct {
	aptRepo     *repository.AppointmentRepository
	visitRepo   *repository.VisitRepository
	patientRepo *repository.PatientRepository
	auditRepo   *repository.AuditRepository
}

func NewAppointmentUsecase(
	aptRepo *repository.AppointmentRepository,
	visitRepo *repository.VisitRepository,
	patientRepo *repository.PatientRepository,
	auditRepo *repository.AuditRepository,
) *AppointmentUsecase {
	return &AppointmentUsecase{
		aptRepo:     aptRepo,
		visitRepo:   visitRepo,
		patientRepo: patientRepo,
		auditRepo:   auditRepo,
	}
}

func (u *AppointmentUsecase) GetAppointments(date, status, doctorID string, page, limit int) ([]domain.Appointment, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return u.aptRepo.FindAll(date, status, doctorID, page, limit)
}

func (u *AppointmentUsecase) GetAppointmentByID(id string) (*domain.Appointment, error) {
	return u.aptRepo.FindByID(id)
}

func (u *AppointmentUsecase) CreateAppointment(req domain.CreateAppointmentRequest, currentUserID, currentUsername string) (*domain.Appointment, error) {
	if req.PatientID == "" || req.PoliID == 0 || req.AppointmentDate == "" || req.StartTime == "" {
		return nil, errors.New("pasien, poli, tanggal janji temu, dan jam mulai wajib diisi")
	}

	apt := &domain.Appointment{
		PatientID:       req.PatientID,
		PoliID:          req.PoliID,
		DoctorID:        req.DoctorID,
		AppointmentDate: req.AppointmentDate,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		Channel:         req.Channel,
		Catatan:         req.Catatan,
	}

	err := u.aptRepo.Create(apt)
	if err != nil {
		return nil, err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "CREATE",
		EntityType: "appointment",
		EntityID:   apt.ID,
		Details:    `{"date": "` + apt.AppointmentDate + `", "time": "` + apt.StartTime + `"}`,
	})

	return u.aptRepo.FindByID(apt.ID)
}

func (u *AppointmentUsecase) Reschedule(id string, req domain.RescheduleAppointmentRequest, currentUserID, currentUsername string) error {
	err := u.aptRepo.Reschedule(id, req.AppointmentDate, req.StartTime, req.EndTime, req.Catatan)
	if err != nil {
		return err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "RESCHEDULE",
		EntityType: "appointment",
		EntityID:   id,
		Details:    `{"new_date": "` + req.AppointmentDate + `"}`,
	})

	return nil
}

func (u *AppointmentUsecase) Cancel(id, notes, currentUserID, currentUsername string) error {
	err := u.aptRepo.Cancel(id, notes)
	if err != nil {
		return err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "CANCEL",
		EntityType: "appointment",
		EntityID:   id,
		Details:    `{"notes": "` + notes + `"}`,
	})

	return nil
}

func (u *AppointmentUsecase) CheckIn(id, currentUserID, currentUsername string) (*domain.Visit, error) {
	apt, err := u.aptRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("janji temu tidak ditemukan")
	}

	if apt.Status == "HADIR" {
		return nil, errors.New("pasien sudah melakukan check-in")
	}

	// Create visit automatically
	visit := &domain.Visit{
		PatientID:    apt.PatientID,
		PoliID:       apt.PoliID,
		DoctorID:     apt.DoctorID,
		GuarantorID:  1, // Default to UMUM unless updated
		KeluhanUtama: "Janji Temu: " + apt.Catatan,
		CreatedBy:    &currentUserID,
	}

	err = u.visitRepo.Create(visit)
	if err != nil {
		return nil, err
	}

	// Link visit to appointment
	err = u.aptRepo.CheckIn(id, currentUserID, visit.ID)
	if err != nil {
		return nil, err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "CHECK_IN",
		EntityType: "appointment",
		EntityID:   id,
		Details:    `{"visit_id": "` + visit.ID + `", "no_rawat": "` + visit.NoRawat + `"}`,
	})

	return u.visitRepo.FindByID(visit.ID)
}
