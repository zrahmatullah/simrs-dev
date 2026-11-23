package usecase

import (
	"errors"
	"fmt"
	"time"

	"simrs-asoy/backend/internal/domain"
	"simrs-asoy/backend/internal/repository"
	"simrs-asoy/backend/internal/security"
)

type ClinicalUsecase struct {
	clinicalRepo *repository.ClinicalRepository
	visitRepo    *repository.VisitRepository
	patientRepo  *repository.PatientRepository
	auditRepo    *repository.AuditRepository
}

func NewClinicalUsecase(
	clinicalRepo *repository.ClinicalRepository,
	visitRepo *repository.VisitRepository,
	patientRepo *repository.PatientRepository,
	auditRepo *repository.AuditRepository,
) *ClinicalUsecase {
	return &ClinicalUsecase{
		clinicalRepo: clinicalRepo,
		visitRepo:    visitRepo,
		patientRepo:  patientRepo,
		auditRepo:    auditRepo,
	}
}

// 1. Vital Signs TTV (Nurse)
func (u *ClinicalUsecase) CreateVitalSign(req domain.CreateVitalSignRequest, currentUserID, currentUsername string) (*domain.VitalSign, error) {
	// OWASP: Validate vital signs physiological bounds
	if err := security.ValidateVitalSigns(req.TDSistole, req.TDDiastole, req.HR, req.RR, req.Temp, req.SpO2); err != nil {
		return nil, err
	}

	visit, err := u.visitRepo.FindByID(req.VisitID)
	if err != nil {
		return nil, errors.New("kunjungan tidak ditemukan")
	}

	triase := security.SanitizePlainText(req.TriaseLevel)
	if triase == "" {
		triase = "HIJAU"
	}

	kesadaran := security.SanitizePlainText(req.Kesadaran)
	if kesadaran == "" {
		kesadaran = "Compos Mentis"
	}

	vs := &domain.VitalSign{
		VisitID:            req.VisitID,
		PatientID:          visit.PatientID,
		NurseID:            &currentUserID,
		TDSistole:          req.TDSistole,
		TDDiastole:         req.TDDiastole,
		HR:                 req.HR,
		RR:                 req.RR,
		Temp:               req.Temp,
		SpO2:               req.SpO2,
		BB:                 req.BB,
		TB:                 req.TB,
		LK:                 req.LK,
		SkalaNyeri:         req.SkalaNyeri,
		Kesadaran:          kesadaran,
		TriaseLevel:        triase,
		CatatanKeperawatan: security.SanitizeText(req.CatatanKeperawatan),
	}

	err = u.clinicalRepo.CreateVitalSign(vs)
	if err != nil {
		return nil, err
	}

	// Move visit to DIPERIKSA_DOKTER
	_ = u.visitRepo.UpdateStatus(req.VisitID, "DIPERIKSA_DOKTER")

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "INPUT_TTV",
		EntityType: "vital_sign",
		EntityID:   vs.ID,
		Details:    `{"visit_id": "` + req.VisitID + `", "triase": "` + triase + `"}`,
	})

	return vs, nil
}

func (u *ClinicalUsecase) GetVitalSignsByVisit(visitID string) ([]domain.VitalSign, error) {
	return u.clinicalRepo.GetVitalSignsByVisitID(visitID)
}

func (u *ClinicalUsecase) GetVitalSignsByPatient(patientID string) ([]domain.VitalSign, error) {
	return u.clinicalRepo.GetVitalSignsByPatientID(patientID)
}

// 2. Doctor EMR (SOAP)
func (u *ClinicalUsecase) CreateMedicalRecord(req domain.CreateMedicalRecordRequest, currentUserID, currentUsername string) (*domain.MedicalRecord, error) {
	if req.VisitID == "" || req.Subjektif == "" || req.Objektif == "" || req.Asesmen == "" || req.Planning == "" {
		return nil, errors.New("kunjungan dan kolom SOAP (Subjektif, Objektif, Asesmen, Planning) wajib diisi lengkap")
	}

	visit, err := u.visitRepo.FindByID(req.VisitID)
	if err != nil {
		return nil, errors.New("kunjungan tidak ditemukan")
	}

	status := security.SanitizePlainText(req.Status)
	if status == "" {
		status = "FINAL"
	}

	// OWASP: Sanitize clinical narratives to eradicate Stored XSS vectors
	mr := &domain.MedicalRecord{
		VisitID:            req.VisitID,
		PatientID:          visit.PatientID,
		DoctorID:           currentUserID,
		Subjektif:          security.SanitizeText(req.Subjektif),
		Objektif:           security.SanitizeText(req.Objektif),
		Asesmen:            security.SanitizeText(req.Asesmen),
		Planning:           security.SanitizeText(req.Planning),
		ICD10Code:          req.ICD10Code,
		ICD9Code:           req.ICD9Code,
		DiagnosisDeskripsi: security.SanitizeText(req.DiagnosisDeskripsi),
		ProsedurDeskripsi:  security.SanitizeText(req.ProsedurDeskripsi),
		Status:             status,
		IsLocked:           true,
	}

	err = u.clinicalRepo.CreateMedicalRecord(mr)
	if err != nil {
		return nil, err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "EMR_SOAP",
		EntityType: "medical_record",
		EntityID:   mr.ID,
		Details:    `{"visit_id": "` + req.VisitID + `", "icd10": "` + safeString(req.ICD10Code) + `"}`,
	})

	return mr, nil
}

func (u *ClinicalUsecase) GetMedicalRecordByVisit(visitID string) (*domain.MedicalRecord, error) {
	return u.clinicalRepo.GetMedicalRecordByVisitID(visitID)
}

func (u *ClinicalUsecase) GetMedicalRecordsByPatient(patientID string) ([]domain.MedicalRecord, error) {
	return u.clinicalRepo.GetMedicalRecordsByPatientID(patientID)
}

// 3. Medical Certificate (Surat Sakit / Rujukan)
func (u *ClinicalUsecase) CreateMedicalCertificate(req domain.CreateMedicalCertificateRequest, currentUserID, currentUsername string) (*domain.MedicalCertificate, error) {
	visit, err := u.visitRepo.FindByID(req.VisitID)
	if err != nil {
		return nil, errors.New("kunjungan tidak ditemukan")
	}

	mc := &domain.MedicalCertificate{
		VisitID:    req.VisitID,
		PatientID:  visit.PatientID,
		DoctorID:   currentUserID,
		JenisSurat: req.JenisSurat,
		Diagnosa:   req.Diagnosa,
		Catatan:    req.Catatan,
		TglMulai:   req.TglMulai,
		TglSelesai: req.TglSelesai,
	}

	err = u.clinicalRepo.CreateMedicalCertificate(mc)
	if err != nil {
		return nil, err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "CREATE_CERTIFICATE",
		EntityType: "medical_certificate",
		EntityID:   mc.ID,
		Details:    `{"nomor_surat": "` + mc.NomorSurat + `", "jenis": "` + mc.JenisSurat + `"}`,
	})

	return mc, nil
}

func (u *ClinicalUsecase) GetMedicalCertificateByVisit(visitID string) (*domain.MedicalCertificate, error) {
	return u.clinicalRepo.GetMedicalCertificateByVisitID(visitID)
}

// 4. Clinical Procedures in Poli Room
func (u *ClinicalUsecase) GetMasterProcedures() ([]domain.MasterProcedure, error) {
	return u.clinicalRepo.GetMasterProcedures()
}

func (u *ClinicalUsecase) CreateVisitProcedure(req domain.AddVisitProcedureRequest, currentUserID string) (*domain.VisitProcedure, error) {
	if req.VisitID == "" || req.ProcedureName == "" {
		return nil, errors.New("kunjungan dan nama tindakan wajib diisi")
	}
	qty := req.Qty
	if qty <= 0 {
		qty = 1
	}
	subtotal := req.Tariff * float64(qty)
	vp := &domain.VisitProcedure{
		VisitID:       req.VisitID,
		ProcedureName: req.ProcedureName,
		Tariff:        req.Tariff,
		Qty:           qty,
		Subtotal:      subtotal,
		DoctorID:      &currentUserID,
		Catatan:       req.Catatan,
	}
	err := u.clinicalRepo.CreateVisitProcedure(vp)
	if err != nil {
		return nil, err
	}
	return vp, nil
}

func (u *ClinicalUsecase) GetVisitProcedures(visitID string) ([]domain.VisitProcedure, error) {
	return u.clinicalRepo.GetVisitProcedures(visitID)
}

// 5. Detailed Nursing Assessment
func (u *ClinicalUsecase) CreateNursingAssessment(req domain.CreateNursingAssessmentRequest, currentUserID string) (*domain.NursingAssessment, error) {
	visit, err := u.visitRepo.FindByID(req.VisitID)
	if err != nil {
		return nil, errors.New("kunjungan tidak ditemukan")
	}

	na := &domain.NursingAssessment{
		VisitID:                 req.VisitID,
		PatientID:               visit.PatientID,
		NurseID:                 &currentUserID,
		KeluhanKeperawatan:      req.KeluhanKeperawatan,
		RiwayatPenyakitSekarang: req.RiwayatPenyakitSekarang,
		RiwayatAlergi:           req.RiwayatAlergi,
		SkalaNyeriSkor:          req.SkalaNyeriSkor,
		SkalaNyeriLokasi:        req.SkalaNyeriLokasi,
		SkalaNyeriDurasi:        req.SkalaNyeriDurasi,
		RisikoJatuhTingkat:      req.RisikoJatuhTingkat,
		SkriningGiziTurunBB:     req.SkriningGiziTurunBB,
		SkriningGiziNafsuMakan:  req.SkriningGiziNafsuMakan,
		MasalahKeperawatan:      req.MasalahKeperawatan,
		RencanaKeperawatan:      req.RencanaKeperawatan,
		EvaluasiKeperawatan:     req.EvaluasiKeperawatan,
	}

	err = u.clinicalRepo.CreateNursingAssessment(na)
	if err != nil {
		return nil, err
	}
	return na, nil
}

func (u *ClinicalUsecase) GetNursingAssessment(visitID string) (*domain.NursingAssessment, error) {
	return u.clinicalRepo.GetNursingAssessmentByVisitID(visitID)
}

// 6. Informed Consent Digital (Fase 1)
func (u *ClinicalUsecase) CreateInformedConsent(req domain.CreateInformedConsentRequest, currentUserID, currentUsername string) (*domain.InformedConsent, error) {
	if req.VisitID == "" || req.ActionName == "" || req.SignerName == "" || req.PatientSignatureData == "" {
		return nil, errors.New("tindakan medis, nama penandatangan, dan tanda tangan digital wajib diisi")
	}

	visit, err := u.visitRepo.FindByID(req.VisitID)
	if err != nil {
		return nil, errors.New("kunjungan tidak ditemukan")
	}

	tok, _ := security.GenerateRandomToken(8)
	qrToken := fmt.Sprintf("IC-%s-%s", time.Now().Format("20060102"), tok)

	consent := &domain.InformedConsent{
		VisitID:              req.VisitID,
		PatientID:            visit.PatientID,
		DoctorID:             currentUserID,
		ConsentType:          req.ConsentType,
		ActionName:           req.ActionName,
		RisksExplained:       req.RisksExplained,
		SignerName:           req.SignerName,
		SignerRelation:       req.SignerRelation,
		PatientSignatureData: req.PatientSignatureData,
		DoctorSignatureData:  req.DoctorSignatureData,
		QRCodeToken:          qrToken,
	}

	err = u.clinicalRepo.CreateInformedConsent(consent)
	if err != nil {
		return nil, err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "CREATE_INFORMED_CONSENT",
		EntityType: "informed_consent",
		EntityID:   consent.ID,
		Details:    fmt.Sprintf(`{"action_name": "%s", "signer": "%s"}`, consent.ActionName, consent.SignerName),
	})

	return consent, nil
}

func (u *ClinicalUsecase) GetInformedConsentsByVisit(visitID string) ([]domain.InformedConsent, error) {
	return u.clinicalRepo.GetInformedConsentsByVisitID(visitID)
}

func (u *ClinicalUsecase) GetInformedConsentByID(id string) (*domain.InformedConsent, error) {
	return u.clinicalRepo.GetInformedConsentByID(id)
}

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
