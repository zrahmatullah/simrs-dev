package usecase

import (
	"errors"

	"simrs-asoy/backend/internal/domain"
	"simrs-asoy/backend/internal/repository"
)

type AncillaryUsecase struct {
	ancillaryRepo *repository.AncillaryRepository
	visitRepo     *repository.VisitRepository
	auditRepo     *repository.AuditRepository
}

func NewAncillaryUsecase(
	ancillaryRepo *repository.AncillaryRepository,
	visitRepo *repository.VisitRepository,
	auditRepo *repository.AuditRepository,
) *AncillaryUsecase {
	return &AncillaryUsecase{
		ancillaryRepo: ancillaryRepo,
		visitRepo:     visitRepo,
		auditRepo:     auditRepo,
	}
}

func (u *AncillaryUsecase) GetOrders(status, jenisOrder, search string, page, limit int) ([]domain.AncillaryOrder, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return u.ancillaryRepo.FindOrders(status, jenisOrder, search, page, limit)
}

func (u *AncillaryUsecase) GetOrderByID(id string) (*domain.AncillaryOrder, error) {
	return u.ancillaryRepo.GetOrderByID(id)
}

func (u *AncillaryUsecase) GetOrdersByVisit(visitID string) ([]domain.AncillaryOrder, error) {
	return u.ancillaryRepo.GetOrdersByVisitID(visitID)
}

func (u *AncillaryUsecase) CreateOrder(req domain.CreateAncillaryOrderRequest, currentUserID, currentUsername string) (*domain.AncillaryOrder, error) {
	if req.VisitID == "" || req.NamaTindakan == "" || req.JenisOrder == "" {
		return nil, errors.New("kunjungan, nama tindakan, dan jenis order wajib diisi")
	}

	visit, err := u.visitRepo.FindByID(req.VisitID)
	if err != nil {
		return nil, errors.New("kunjungan tidak ditemukan")
	}

	order := &domain.AncillaryOrder{
		VisitID:       req.VisitID,
		PatientID:     visit.PatientID,
		DoctorID:      currentUserID,
		JenisOrder:    req.JenisOrder,
		NamaTindakan:  req.NamaTindakan,
		Tarif:         req.Tarif,
		CatatanKlinis: req.CatatanKlinis,
	}

	err = u.ancillaryRepo.CreateOrder(order)
	if err != nil {
		return nil, err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "CREATE_ANCILLARY_ORDER",
		EntityType: "ancillary_order",
		EntityID:   order.ID,
		Details:    `{"nomor_order": "` + order.NomorOrder + `", "jenis": "` + order.JenisOrder + `"}`,
	})

	return order, nil
}

func (u *AncillaryUsecase) UpdateResult(id string, req domain.UpdateAncillaryResultRequest, currentUserID, currentUsername string) error {
	if req.HasilPemeriksaan == "" {
		return errors.New("hasil pemeriksaan wajib diisi")
	}

	err := u.ancillaryRepo.UpdateResult(id, req.HasilPemeriksaan, req.LampiranURL, currentUserID)
	if err != nil {
		return err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "UPDATE_ANCILLARY_RESULT",
		EntityType: "ancillary_order",
		EntityID:   id,
		Details:    `{"message": "Hasil penunjang divalidasi"}`,
	})

	return nil
}
