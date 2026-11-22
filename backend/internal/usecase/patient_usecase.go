package usecase

import (
	"errors"

	"simrs-asoy/backend/internal/domain"
	"simrs-asoy/backend/internal/repository"
)

type PatientUsecase struct {
	patientRepo *repository.PatientRepository
	auditRepo   *repository.AuditRepository
}

func NewPatientUsecase(patientRepo *repository.PatientRepository, auditRepo *repository.AuditRepository) *PatientUsecase {
	return &PatientUsecase{patientRepo: patientRepo, auditRepo: auditRepo}
}

func (u *PatientUsecase) GetPatients(search string, page, limit int) ([]domain.Patient, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return u.patientRepo.FindAll(search, page, limit)
}

func (u *PatientUsecase) GetPatientByID(id string) (*domain.Patient, error) {
	return u.patientRepo.FindByID(id)
}

func (u *PatientUsecase) GetPatientByNoRM(noRM string) (*domain.Patient, error) {
	return u.patientRepo.FindByNoRM(noRM)
}

func (u *PatientUsecase) CreatePatient(req domain.CreatePatientRequest, currentUserID, currentUsername string) (*domain.Patient, error) {
	if req.NamaPasien == "" || req.NIK == "" || req.TanggalLahir == "" || req.JenisKelamin == "" {
		return nil, errors.New("nama pasien, NIK, tanggal lahir, dan jenis kelamin wajib diisi")
	}

	patient := &domain.Patient{
		NIK:           req.NIK,
		NamaPasien:    req.NamaPasien,
		JenisKelamin:  req.JenisKelamin,
		TempatLahir:   req.TempatLahir,
		TanggalLahir:  req.TanggalLahir,
		NamaOrangTua:  req.NamaOrangTua,
		NoHP:          req.NoHP,
		Alamat:        req.Alamat,
		RiwayatAlergi: req.RiwayatAlergi,
		GolonganDarah: req.GolonganDarah,
		StatusAktif:   true,
		CreatedBy:     &currentUserID,
	}

	err := u.patientRepo.Create(patient)
	if err != nil {
		return nil, err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "CREATE",
		EntityType: "patient",
		EntityID:   patient.ID,
		Details:    `{"no_rm": "` + patient.NoRM + `", "nama_pasien": "` + patient.NamaPasien + `"}`,
	})

	return patient, nil
}

func (u *PatientUsecase) UpdatePatient(id string, req domain.UpdatePatientRequest, currentUserID, currentUsername string) error {
	patient, err := u.patientRepo.FindByID(id)
	if err != nil {
		return err
	}

	patient.NIK = req.NIK
	patient.NamaPasien = req.NamaPasien
	patient.JenisKelamin = req.JenisKelamin
	patient.TempatLahir = req.TempatLahir
	patient.TanggalLahir = req.TanggalLahir
	patient.NamaOrangTua = req.NamaOrangTua
	patient.NoHP = req.NoHP
	patient.Alamat = req.Alamat
	patient.RiwayatAlergi = req.RiwayatAlergi
	patient.GolonganDarah = req.GolonganDarah
	patient.StatusAktif = req.StatusAktif
	patient.UpdatedBy = &currentUserID

	err = u.patientRepo.Update(patient)
	if err != nil {
		return err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "UPDATE",
		EntityType: "patient",
		EntityID:   id,
		Details:    `{"nama_pasien": "` + patient.NamaPasien + `"}`,
	})

	return nil
}

func (u *PatientUsecase) DeletePatient(id, currentUserID, currentUsername string) error {
	err := u.patientRepo.Delete(id, currentUserID)
	if err != nil {
		return err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "DELETE",
		EntityType: "patient",
		EntityID:   id,
		Details:    `{"deleted": true}`,
	})

	return nil
}
