package usecase

import (
	"simrs-asoy/backend/internal/domain"
	"simrs-asoy/backend/internal/repository"
)

type MasterUsecase struct {
	masterRepo *repository.MasterRepository
}

func NewMasterUsecase(masterRepo *repository.MasterRepository) *MasterUsecase {
	return &MasterUsecase{masterRepo: masterRepo}
}

func (u *MasterUsecase) GetRoles() ([]domain.Role, error) {
	return u.masterRepo.GetRoles()
}

func (u *MasterUsecase) GetPoliUnits() ([]domain.PoliUnit, error) {
	return u.masterRepo.GetPoliUnits()
}

func (u *MasterUsecase) GetGuarantors() ([]domain.Guarantor, error) {
	return u.masterRepo.GetGuarantors()
}

func (u *MasterUsecase) GetICD10(search string) ([]domain.ICD10, error) {
	return u.masterRepo.GetICD10(search)
}

func (u *MasterUsecase) GetICD9(search string) ([]domain.ICD9, error) {
	return u.masterRepo.GetICD9(search)
}

func (u *MasterUsecase) GetCategories() ([]domain.InventoryCategory, error) {
	return u.masterRepo.GetCategories()
}

func (u *MasterUsecase) GetPromos() ([]domain.Promo, error) {
	return u.masterRepo.GetPromos()
}

func (u *MasterUsecase) CreatePromo(req domain.CreatePromoRequest) (*domain.Promo, error) {
	p := &domain.Promo{
		Code:           req.Code,
		Name:           req.Name,
		DiscountType:   req.DiscountType,
		DiscountValue:  req.DiscountValue,
		MinTransaction: req.MinTransaction,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		StatusAktif:    req.StatusAktif,
	}
	err := u.masterRepo.CreatePromo(p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (u *MasterUsecase) GetSuppliers(search string) ([]domain.Supplier, error) {
	return u.masterRepo.GetSuppliers(search)
}

func (u *MasterUsecase) CreateSupplier(req domain.CreateSupplierRequest) (*domain.Supplier, error) {
	s := &domain.Supplier{
		Code:        req.Code,
		Name:        req.Name,
		PicName:     req.PicName,
		Phone:       req.Phone,
		Email:       req.Email,
		Address:     req.Address,
		City:        req.City,
		StatusAktif: true,
	}
	err := u.masterRepo.CreateSupplier(s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (u *MasterUsecase) CreatePoli(req domain.CreatePoliRequest) (*domain.PoliUnit, error) {
	p := &domain.PoliUnit{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		StatusAktif: true,
	}
	err := u.masterRepo.CreatePoli(p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (u *MasterUsecase) GetFaskes(search string) ([]domain.FaskesRujukan, error) {
	return u.masterRepo.GetFaskes(search)
}

func (u *MasterUsecase) CreateFaskes(req domain.CreateFaskesRequest) (*domain.FaskesRujukan, error) {
	if req.Tipe == "" {
		req.Tipe = "RSUD"
	}
	f := &domain.FaskesRujukan{
		Code:        req.Code,
		Name:        req.Name,
		Tipe:        req.Tipe,
		Alamat:      req.Alamat,
		Kota:        req.Kota,
		NoTelp:      req.NoTelp,
		StatusAktif: true,
	}
	err := u.masterRepo.CreateFaskes(f)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (u *MasterUsecase) GetSpesialis(search string) ([]domain.MasterSpesialis, error) {
	return u.masterRepo.GetSpesialis(search)
}

func (u *MasterUsecase) CreateSpesialis(req domain.CreateSpesialisRequest) (*domain.MasterSpesialis, error) {
	s := &domain.MasterSpesialis{
		Code:         req.Code,
		NamaDokter:   req.NamaDokter,
		Spesialisasi: req.Spesialisasi,
		RumahSakit:   req.RumahSakit,
		StatusAktif:  true,
	}
	err := u.masterRepo.CreateSpesialis(s)
	if err != nil {
		return nil, err
	}
	return s, nil
}
