package repository

import (
	"fmt"
	"time"

	"simrs-asoy/backend/internal/domain"

	"github.com/jmoiron/sqlx"
)

type ClinicalRepository struct {
	db *sqlx.DB
}

func NewClinicalRepository(db *sqlx.DB) *ClinicalRepository {
	return &ClinicalRepository{db: db}
}

// 1. Vital Signs
func (r *ClinicalRepository) CreateVitalSign(vs *domain.VitalSign) error {
	query := `
		INSERT INTO vital_signs (visit_id, patient_id, nurse_id, td_sistole, td_diastole, hr, rr, temp, spo2,
		                         bb, tb, lk, skala_nyeri, kesadaran, triase_level, catatan_keperawatan,
		                         created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, NOW(), NOW())
		RETURNING id, created_at
	`
	return r.db.QueryRow(query, vs.VisitID, vs.PatientID, vs.NurseID, vs.TDSistole, vs.TDDiastole, vs.HR, vs.RR, vs.Temp,
		vs.SpO2, vs.BB, vs.TB, vs.LK, vs.SkalaNyeri, vs.Kesadaran, vs.TriaseLevel, vs.CatatanKeperawatan, vs.NurseID).
		Scan(&vs.ID, &vs.CreatedAt)
}

func (r *ClinicalRepository) GetVitalSignsByVisitID(visitID string) ([]domain.VitalSign, error) {
	query := `
		SELECT vs.id, vs.visit_id, vs.patient_id, vs.nurse_id, COALESCE(u.full_name, '') AS nurse_name,
		       vs.td_sistole, vs.td_diastole, vs.hr, vs.rr, vs.temp, vs.spo2, vs.bb, vs.tb, vs.lk,
		       vs.skala_nyeri, vs.kesadaran, vs.triase_level, vs.catatan_keperawatan, vs.created_at
		FROM vital_signs vs
		LEFT JOIN users u ON vs.nurse_id = u.ids
		WHERE vs.visit_id = $1 AND vs.deleted_at IS NULL
		ORDER BY vs.created_at DESC
	`
	var list []domain.VitalSign
	err := r.db.Select(&list, query, visitID)
	return list, err
}

func (r *ClinicalRepository) GetVitalSignsByPatientID(patientID string) ([]domain.VitalSign, error) {
	query := `
		SELECT vs.id, vs.visit_id, vs.patient_id, vs.nurse_id, COALESCE(u.full_name, '') AS nurse_name,
		       vs.td_sistole, vs.td_diastole, vs.hr, vs.rr, vs.temp, vs.spo2, vs.bb, vs.tb, vs.lk,
		       vs.skala_nyeri, vs.kesadaran, vs.triase_level, vs.catatan_keperawatan, vs.created_at
		FROM vital_signs vs
		LEFT JOIN users u ON vs.nurse_id = u.id
		WHERE vs.patient_id = $1 AND vs.deleted_at IS NULL
		ORDER BY vs.created_at DESC
		LIMIT 10
	`
	var list []domain.VitalSign
	err := r.db.Select(&list, query, patientID)
	return list, err
}

// 2. Medical Records EMR (SOAP)
func (r *ClinicalRepository) CreateMedicalRecord(mr *domain.MedicalRecord) error {
	query := `
		INSERT INTO medical_records (visit_id, patient_id, doctor_id, subjektif, objektif, asesmen, planning,
		                             icd10_code, icd9_code, diagnosis_deskripsi, prosedur_deskripsi, status,
		                             is_locked, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, true, $13, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(query, mr.VisitID, mr.PatientID, mr.DoctorID, mr.Subjektif, mr.Objektif, mr.Asesmen, mr.Planning,
		mr.ICD10Code, mr.ICD9Code, mr.DiagnosisDeskripsi, mr.ProsedurDeskripsi, mr.Status, mr.DoctorID).
		Scan(&mr.ID, &mr.CreatedAt, &mr.UpdatedAt)
}

func (r *ClinicalRepository) GetMedicalRecordByVisitID(visitID string) (*domain.MedicalRecord, error) {
	query := `
		SELECT mr.id, mr.visit_id, mr.patient_id, mr.doctor_id, COALESCE(u.full_name, '') AS doctor_name,
		       mr.subjektif, mr.objektif, mr.asesmen, mr.planning, mr.icd10_code, mr.icd9_code,
		       mr.diagnosis_deskripsi, mr.prosedur_deskripsi, mr.status, mr.is_locked, mr.created_at, mr.updated_at
		FROM medical_records mr
		JOIN users u ON mr.doctor_id = u.id
		WHERE mr.visit_id = $1 AND mr.deleted_at IS NULL
		LIMIT 1
	`
	var mr domain.MedicalRecord
	err := r.db.Get(&mr, query, visitID)
	if err != nil {
		return nil, err
	}
	return &mr, nil
}

func (r *ClinicalRepository) GetMedicalRecordsByPatientID(patientID string) ([]domain.MedicalRecord, error) {
	query := `
		SELECT mr.id, mr.visit_id, mr.patient_id, mr.doctor_id, COALESCE(u.full_name, '') AS doctor_name,
		       mr.subjektif, mr.objektif, mr.asesmen, mr.planning, mr.icd10_code, mr.icd9_code,
		       mr.diagnosis_deskripsi, mr.prosedur_deskripsi, mr.status, mr.is_locked, mr.created_at, mr.updated_at
		FROM medical_records mr
		JOIN users u ON mr.doctor_id = u.id
		WHERE mr.patient_id = $1 AND mr.deleted_at IS NULL
		ORDER BY mr.created_at DESC
	`
	var list []domain.MedicalRecord
	err := r.db.Select(&list, query, patientID)
	return list, err
}

// 3. Medical Certificate (Surat Sakit / Rujukan)
func (r *ClinicalRepository) CreateMedicalCertificate(mc *domain.MedicalCertificate) error {
	dateStr := time.Now().Format("20060102")
	var count int
	_ = r.db.Get(&count, "SELECT COUNT(*) FROM medical_certificates WHERE nomor_surat LIKE $1", fmt.Sprintf("SK-%s-%%", dateStr))
	mc.NomorSurat = fmt.Sprintf("SK-%s-%03d", dateStr, count+1)

	query := `
		INSERT INTO medical_certificates (nomor_surat, visit_id, patient_id, doctor_id, jenis_surat, diagnosa, catatan,
		                                  tgl_mulai, tgl_selesai, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, created_at
	`
	return r.db.QueryRow(query, mc.NomorSurat, mc.VisitID, mc.PatientID, mc.DoctorID, mc.JenisSurat, mc.Diagnosa, mc.Catatan,
		mc.TglMulai, mc.TglSelesai).Scan(&mc.ID, &mc.CreatedAt)
}

func (r *ClinicalRepository) GetMedicalCertificateByVisitID(visitID string) (*domain.MedicalCertificate, error) {
	query := `
		SELECT id, nomor_surat, visit_id, patient_id, doctor_id, jenis_surat, diagnosa, catatan,
		       TO_CHAR(tgl_mulai, 'YYYY-MM-DD') AS tgl_mulai, TO_CHAR(tgl_selesai, 'YYYY-MM-DD') AS tgl_selesai,
		       created_at
		FROM medical_certificates
		WHERE visit_id = $1 AND deleted_at IS NULL
		LIMIT 1
	`
	var mc domain.MedicalCertificate
	err := r.db.Get(&mc, query, visitID)
	if err != nil {
		return nil, err
	}
	return &mc, nil
}

// 4. Clinical Procedures (Tindakan di Ruangan Poli)
func (r *ClinicalRepository) GetMasterProcedures() ([]domain.MasterProcedure, error) {
	var list []domain.MasterProcedure
	err := r.db.Select(&list, "SELECT id, code, name, category, tariff, status_aktif, created_at FROM master_procedures WHERE status_aktif = true ORDER BY name ASC")
	return list, err
}

func (r *ClinicalRepository) CreateVisitProcedure(vp *domain.VisitProcedure) error {
	vp.Subtotal = vp.Tariff * float64(vp.Qty)
	query := `
		INSERT INTO visit_procedures (visit_id, procedure_name, tariff, qty, subtotal, doctor_id, catatan, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		RETURNING id, created_at
	`
	return r.db.QueryRow(query, vp.VisitID, vp.ProcedureName, vp.Tariff, vp.Qty, vp.Subtotal, vp.DoctorID, vp.Catatan).
		Scan(&vp.ID, &vp.CreatedAt)
}

func (r *ClinicalRepository) GetVisitProcedures(visitID string) ([]domain.VisitProcedure, error) {
	var list []domain.VisitProcedure
	query := `
		SELECT id, visit_id, procedure_name, tariff, qty, subtotal, doctor_id, catatan, created_at
		FROM visit_procedures
		WHERE visit_id = $1
		ORDER BY created_at ASC
	`
	err := r.db.Select(&list, query, visitID)
	return list, err
}

// 5. Nursing Assessment (Pengkajian Keperawatan Detail)
func (r *ClinicalRepository) CreateNursingAssessment(na *domain.NursingAssessment) error {
	query := `
		INSERT INTO nursing_assessments (
			visit_id, patient_id, nurse_id, keluhan_keperawatan, riwayat_penyakit_sekarang, riwayat_alergi,
			skala_nyeri_skor, skala_nyeri_lokasi, skala_nyeri_durasi, risiko_jatuh_tingkat,
			skrining_gizi_turun_bb, skrining_gizi_nafsu_makan, masalah_keperawatan, rencana_keperawatan,
			evaluasi_keperawatan, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW(), NOW()
		)
		ON CONFLICT (visit_id) DO UPDATE SET
			keluhan_keperawatan = EXCLUDED.keluhan_keperawatan,
			riwayat_penyakit_sekarang = EXCLUDED.riwayat_penyakit_sekarang,
			riwayat_alergi = EXCLUDED.riwayat_alergi,
			skala_nyeri_skor = EXCLUDED.skala_nyeri_skor,
			skala_nyeri_lokasi = EXCLUDED.skala_nyeri_lokasi,
			skala_nyeri_durasi = EXCLUDED.skala_nyeri_durasi,
			risiko_jatuh_tingkat = EXCLUDED.risiko_jatuh_tingkat,
			skrining_gizi_turun_bb = EXCLUDED.skrining_gizi_turun_bb,
			skrining_gizi_nafsu_makan = EXCLUDED.skrining_gizi_nafsu_makan,
			masalah_keperawatan = EXCLUDED.masalah_keperawatan,
			rencana_keperawatan = EXCLUDED.rencana_keperawatan,
			evaluasi_keperawatan = EXCLUDED.evaluasi_keperawatan,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(query,
		na.VisitID, na.PatientID, na.NurseID, na.KeluhanKeperawatan, na.RiwayatPenyakitSekarang, na.RiwayatAlergi,
		na.SkalaNyeriSkor, na.SkalaNyeriLokasi, na.SkalaNyeriDurasi, na.RisikoJatuhTingkat,
		na.SkriningGiziTurunBB, na.SkriningGiziNafsuMakan, na.MasalahKeperawatan, na.RencanaKeperawatan,
		na.EvaluasiKeperawatan,
	).Scan(&na.ID, &na.CreatedAt, &na.UpdatedAt)
}

func (r *ClinicalRepository) GetNursingAssessmentByVisitID(visitID string) (*domain.NursingAssessment, error) {
	var na domain.NursingAssessment
	query := `
		SELECT id, visit_id, patient_id, nurse_id,
		       COALESCE(keluhan_keperawatan, '') AS keluhan_keperawatan,
		       COALESCE(riwayat_penyakit_sekarang, '') AS riwayat_penyakit_sekarang,
		       COALESCE(riwayat_alergi, '') AS riwayat_alergi,
		       skala_nyeri_skor,
		       COALESCE(skala_nyeri_lokasi, '') AS skala_nyeri_lokasi,
		       COALESCE(skala_nyeri_durasi, '') AS skala_nyeri_durasi,
		       COALESCE(risiko_jatuh_tingkat, '') AS risiko_jatuh_tingkat,
		       skrining_gizi_turun_bb, skrining_gizi_nafsu_makan,
		       COALESCE(masalah_keperawatan, '') AS masalah_keperawatan,
		       COALESCE(rencana_keperawatan, '') AS rencana_keperawatan,
		       COALESCE(evaluasi_keperawatan, '') AS evaluasi_keperawatan,
		       created_at, updated_at
		FROM nursing_assessments
		WHERE visit_id = $1
		LIMIT 1
	`
	err := r.db.Get(&na, query, visitID)
	if err != nil {
		return nil, err
	}
	return &na, nil
}

// -------------------------------------------------------------
// Informed Consent Digital (Fase 1)
// -------------------------------------------------------------

func (r *ClinicalRepository) CreateInformedConsent(c *domain.InformedConsent) error {
	query := `
		INSERT INTO informed_consents (visit_id, patient_id, doctor_id, consent_type, action_name,
		                               risks_explained, signer_name, signer_relation, patient_signature_data,
		                               doctor_signature_data, qr_code_token, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
		RETURNING id, created_at
	`
	return r.db.QueryRow(query, c.VisitID, c.PatientID, c.DoctorID, c.ConsentType, c.ActionName,
		c.RisksExplained, c.SignerName, c.SignerRelation, c.PatientSignatureData,
		c.DoctorSignatureData, c.QRCodeToken).Scan(&c.ID, &c.CreatedAt)
}

func (r *ClinicalRepository) GetInformedConsentsByVisitID(visitID string) ([]domain.InformedConsent, error) {
	query := `
		SELECT ic.id, ic.visit_id, v.no_rawat, ic.patient_id, p.nama_pasien AS patient_name, p.no_rm,
		       ic.doctor_id, u.full_name AS doctor_name, ic.consent_type, ic.action_name,
		       ic.risks_explained, ic.signer_name, ic.signer_relation, ic.patient_signature_data,
		       ic.doctor_signature_data, ic.qr_code_token, ic.created_at
		FROM informed_consents ic
		JOIN visits v ON ic.visit_id = v.id
		JOIN patients p ON ic.patient_id = p.id
		JOIN users u ON ic.doctor_id = u.id
		WHERE ic.visit_id = $1
		ORDER BY ic.created_at DESC
	`
	var list []domain.InformedConsent
	err := r.db.Select(&list, query, visitID)
	return list, err
}

func (r *ClinicalRepository) GetInformedConsentByID(id string) (*domain.InformedConsent, error) {
	query := `
		SELECT ic.id, ic.visit_id, v.no_rawat, ic.patient_id, p.nama_pasien AS patient_name, p.no_rm,
		       ic.doctor_id, u.full_name AS doctor_name, ic.consent_type, ic.action_name,
		       ic.risks_explained, ic.signer_name, ic.signer_relation, ic.patient_signature_data,
		       ic.doctor_signature_data, ic.qr_code_token, ic.created_at
		FROM informed_consents ic
		JOIN visits v ON ic.visit_id = v.id
		JOIN patients p ON ic.patient_id = p.id
		JOIN users u ON ic.doctor_id = u.id
		WHERE ic.id = $1
	`
	var ic domain.InformedConsent
	err := r.db.Get(&ic, query, id)
	return &ic, err
}
