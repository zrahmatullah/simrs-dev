package usecase

import (
	"errors"
	"fmt"

	"simrs-asoy/backend/internal/domain"
	"simrs-asoy/backend/internal/repository"
	"simrs-asoy/backend/internal/security"
)

type PharmacyUsecase struct {
	pharmacyRepo *repository.PharmacyRepository
	visitRepo    *repository.VisitRepository
	auditRepo    *repository.AuditRepository
}

func NewPharmacyUsecase(
	pharmacyRepo *repository.PharmacyRepository,
	visitRepo *repository.VisitRepository,
	auditRepo *repository.AuditRepository,
) *PharmacyUsecase {
	return &PharmacyUsecase{
		pharmacyRepo: pharmacyRepo,
		visitRepo:    visitRepo,
		auditRepo:    auditRepo,
	}
}

func (u *PharmacyUsecase) GetPrescriptions(status, search string, page, limit int) ([]domain.Prescription, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return u.pharmacyRepo.FindPrescriptions(status, search, page, limit)
}

func (u *PharmacyUsecase) GetPrescriptionByID(id string) (*domain.Prescription, error) {
	return u.pharmacyRepo.GetPrescriptionByID(id)
}

func (u *PharmacyUsecase) GetPrescriptionByVisit(visitID string) (*domain.Prescription, error) {
	return u.pharmacyRepo.GetPrescriptionByVisitID(visitID)
}

func (u *PharmacyUsecase) CreatePrescription(req domain.CreatePrescriptionRequest, currentUserID, currentUsername string) (*domain.Prescription, error) {
	if req.VisitID == "" || len(req.Items) == 0 {
		return nil, errors.New("kunjungan dan minimal satu item obat wajib diisi")
	}

	// OWASP: Integer Flow and bounds validation for prescription quantities
	for i, item := range req.Items {
		if err := security.ValidateQuantity(item.Jumlah, 1000); err != nil {
			return nil, fmt.Errorf("item obat ke-%d: %w", i+1, err)
		}
		req.Items[i].Signa = security.SanitizeText(item.Signa)
		req.Items[i].AturanPakai = security.SanitizeText(item.AturanPakai)
		req.Items[i].RacikanNama = security.SanitizePlainText(item.RacikanNama)
	}

	visit, err := u.visitRepo.FindByID(req.VisitID)
	if err != nil {
		return nil, errors.New("kunjungan tidak ditemukan")
	}

	p := &domain.Prescription{
		VisitID:       req.VisitID,
		PatientID:     visit.PatientID,
		DoctorID:      currentUserID,
		CatatanDokter: security.SanitizeText(req.CatatanDokter),
	}

	err = u.pharmacyRepo.CreatePrescription(p, req.Items)
	if err != nil {
		return nil, err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "CREATE_PRESCRIPTION",
		EntityType: "prescription",
		EntityID:   p.ID,
		Details:    fmt.Sprintf(`{"nomor_resep": "%s", "items_count": %d}`, p.NomorResep, len(req.Items)),
	})

	return u.pharmacyRepo.GetPrescriptionByID(p.ID)
}

func (u *PharmacyUsecase) ReviewPrescription(id string, req domain.ReviewPrescriptionRequest, currentUserID, currentUsername string) error {
	status := req.Status
	if status == "" {
		status = "TELAAH_OK"
	}

	err := u.pharmacyRepo.ReviewPrescription(id, req.CatatanTelaah, currentUserID, status)
	if err != nil {
		return err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "REVIEW_PRESCRIPTION",
		EntityType: "prescription",
		EntityID:   id,
		Details:    `{"status": "` + status + `", "catatan": "` + req.CatatanTelaah + `"}`,
	})

	return nil
}

func (u *PharmacyUsecase) DispensePrescription(id string, currentUserID, currentUsername string) error {
	err := u.pharmacyRepo.DispensePrescription(id, currentUserID)
	if err != nil {
		return err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "DISPENSE_PRESCRIPTION",
		EntityType: "prescription",
		EntityID:   id,
		Details:    `{"message": "Obat diserahkan dan stok depo farmasi terpotong otomatis"}`,
	})

	return nil
}
