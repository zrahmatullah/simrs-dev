package repository

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"simrs-asoy/backend/internal/domain"
)

type PatientRepository struct {
	db *sqlx.DB
}

func NewPatientRepository(db *sqlx.DB) *PatientRepository {
	return &PatientRepository{db: db}
}

func (r *PatientRepository) GenerateNoRM() (string, error) {
	year := time.Now().Format("2006")
	var count int
	query := `SELECT COUNT(*) FROM patients WHERE no_rm LIKE $1`
	err := r.db.Get(&count, query, fmt.Sprintf("RM-%s-%%", year))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("RM-%s-%04d", year, count+1), nil
}

func (r *PatientRepository) FindAll(search string, page, limit int) ([]domain.Patient, int, error) {
	offset := (page - 1) * limit
	whereClauses := []string{"deleted_at IS NULL"}
	args := []interface{}{}
	argIndex := 1

	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(no_rm ILIKE $%d OR nama_pasien ILIKE $%d OR nik ILIKE $%d OR no_hp ILIKE $%d)", argIndex, argIndex, argIndex, argIndex))
		args = append(args, "%"+search+"%")
		argIndex++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM patients WHERE %s", whereSQL)
	var total int
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, no_rm, nik, nama_pasien, jenis_kelamin,
		       COALESCE(tempat_lahir, '') AS tempat_lahir,
		       TO_CHAR(tanggal_lahir, 'YYYY-MM-DD') AS tanggal_lahir,
		       COALESCE(nama_orang_tua, '') AS nama_orang_tua,
		       COALESCE(no_hp, '') AS no_hp,
		       COALESCE(alamat, '') AS alamat,
		       riwayat_alergi,
		       COALESCE(golongan_darah, '') AS golongan_darah,
		       status_aktif, created_at, updated_at
		FROM patients
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIndex, argIndex+1)

	args = append(args, limit, offset)
	var patients []domain.Patient
	err = r.db.Select(&patients, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return patients, total, nil
}

func (r *PatientRepository) FindByID(id string) (*domain.Patient, error) {
	query := `
		SELECT id, no_rm, nik, nama_pasien, jenis_kelamin,
		       COALESCE(tempat_lahir, '') AS tempat_lahir,
		       TO_CHAR(tanggal_lahir, 'YYYY-MM-DD') AS tanggal_lahir,
		       COALESCE(nama_orang_tua, '') AS nama_orang_tua,
		       COALESCE(no_hp, '') AS no_hp,
		       COALESCE(alamat, '') AS alamat,
		       riwayat_alergi,
		       COALESCE(golongan_darah, '') AS golongan_darah,
		       status_aktif, created_at, updated_at
		FROM patients
		WHERE id = $1 AND deleted_at IS NULL
	`
	var p domain.Patient
	err := r.db.Get(&p, query, id)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PatientRepository) FindByNoRM(noRM string) (*domain.Patient, error) {
	query := `
		SELECT id, no_rm, nik, nama_pasien, jenis_kelamin,
		       COALESCE(tempat_lahir, '') AS tempat_lahir,
		       TO_CHAR(tanggal_lahir, 'YYYY-MM-DD') AS tanggal_lahir,
		       COALESCE(nama_orang_tua, '') AS nama_orang_tua,
		       COALESCE(no_hp, '') AS no_hp,
		       COALESCE(alamat, '') AS alamat,
		       riwayat_alergi,
		       COALESCE(golongan_darah, '') AS golongan_darah,
		       status_aktif, created_at, updated_at
		FROM patients
		WHERE no_rm = $1 AND deleted_at IS NULL
	`
	var p domain.Patient
	err := r.db.Get(&p, query, noRM)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PatientRepository) Create(p *domain.Patient) error {
	if p.NoRM == "" {
		noRM, err := r.GenerateNoRM()
		if err != nil {
			return err
		}
		p.NoRM = noRM
	}

	query := `
		INSERT INTO patients (no_rm, nik, nama_pasien, jenis_kelamin, tempat_lahir, tanggal_lahir,
		                      nama_orang_tua, no_hp, alamat, riwayat_alergi, golongan_darah, status_aktif,
		                      created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(query, p.NoRM, p.NIK, p.NamaPasien, p.JenisKelamin, p.TempatLahir, p.TanggalLahir,
		p.NamaOrangTua, p.NoHP, p.Alamat, p.RiwayatAlergi, p.GolonganDarah, p.StatusAktif, p.CreatedBy).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *PatientRepository) Update(p *domain.Patient) error {
	query := `
		UPDATE patients
		SET nik = $1, nama_pasien = $2, jenis_kelamin = $3, tempat_lahir = $4, tanggal_lahir = $5,
		    nama_orang_tua = $6, no_hp = $7, alamat = $8, riwayat_alergi = $9, golongan_darah = $10,
		    status_aktif = $11, updated_by = $12, updated_at = NOW()
		WHERE id = $13 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(query, p.NIK, p.NamaPasien, p.JenisKelamin, p.TempatLahir, p.TanggalLahir,
		p.NamaOrangTua, p.NoHP, p.Alamat, p.RiwayatAlergi, p.GolonganDarah, p.StatusAktif, p.UpdatedBy, p.ID)
	return err
}

func (r *PatientRepository) Delete(id, deletedBy string) error {
	query := `UPDATE patients SET deleted_at = NOW(), updated_by = $1 WHERE id = $2`
	_, err := r.db.Exec(query, deletedBy, id)
	return err
}
