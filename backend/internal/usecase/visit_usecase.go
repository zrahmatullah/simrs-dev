package usecase

import (
	"errors"

	"simrs-asoy/backend/internal/domain"
	"simrs-asoy/backend/internal/repository"
)

type VisitUsecase struct {
	visitRepo   *repository.VisitRepository
	patientRepo *repository.PatientRepository
	auditRepo   *repository.AuditRepository
}

func NewVisitUsecase(visitRepo *repository.VisitRepository, patientRepo *repository.PatientRepository, auditRepo *repository.AuditRepository) *VisitUsecase {
	return &VisitUsecase{visitRepo: visitRepo, patientRepo: patientRepo, auditRepo: auditRepo}
}

func (u *VisitUsecase) GetVisits(date, status string, poliID int, search string, page, limit int) ([]domain.Visit, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return u.visitRepo.FindAll(date, status, poliID, search, page, limit)
}

func (u *VisitUsecase) GetVisitByID(id string) (*domain.Visit, error) {
	return u.visitRepo.FindByID(id)
}

func (u *VisitUsecase) CreateVisit(req domain.CreateVisitRequest, currentUserID, currentUsername string) (*domain.Visit, error) {
	if req.PatientID == "" || req.PoliID == 0 || req.GuarantorID == 0 {
		return nil, errors.New("pasien, poli, dan penjamin wajib diisi")
	}

	// Verify patient exists
	_, err := u.patientRepo.FindByID(req.PatientID)
	if err != nil {
		return nil, errors.New("data pasien tidak ditemukan")
	}

	visit := &domain.Visit{
		PatientID:     req.PatientID,
		PoliID:        req.PoliID,
		DoctorID:      req.DoctorID,
		GuarantorID:   req.GuarantorID,
		NomorAsuransi: req.NomorAsuransi,
		KeluhanUtama:  req.KeluhanUtama,
		CreatedBy:     &currentUserID,
	}

	err = u.visitRepo.Create(visit)
	if err != nil {
		return nil, err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "CREATE",
		EntityType: "visit",
		EntityID:   visit.ID,
		Details:    `{"no_rawat": "` + visit.NoRawat + `", "no_antrian": "` + visit.NoAntrian + `"}`,
	})

	return u.visitRepo.FindByID(visit.ID)
}

func (u *VisitUsecase) UpdateVisitStatus(id, status, currentUserID, currentUsername string) error {
	validStatuses := map[string]bool{
		"MENUNGGU_TRIASE":  true,
		"DIPERIKSA_DOKTER": true,
		"FARMASI":          true,
		"KASIR":            true,
		"SELESAI":          true,
		"BATAL":            true,
	}

	if !validStatuses[status] {
		return errors.New("status kunjungan tidak valid")
	}

	err := u.visitRepo.UpdateStatus(id, status)
	if err != nil {
		return err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "UPDATE_STATUS",
		EntityType: "visit",
		EntityID:   id,
		Details:    `{"new_status": "` + status + `"}`,
	})

	return nil
}

func (u *VisitUsecase) GetQueueMetrics() (map[string]int, error) {
	return u.visitRepo.GetQueueMetrics()
}
