package repository

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"simrs-asoy/backend/internal/domain"
)

type VisitRepository struct {
	db *sqlx.DB
}

func NewVisitRepository(db *sqlx.DB) *VisitRepository {
	return &VisitRepository{db: db}
}

func (r *VisitRepository) GenerateNoRawat() (string, error) {
	dateStr := time.Now().Format("20060102")
	var count int
	query := `SELECT COUNT(*) FROM visits WHERE no_rawat LIKE $1`
	err := r.db.Get(&count, query, fmt.Sprintf("RAWAT-%s-%%", dateStr))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("RAWAT-%s-%03d", dateStr, count+1), nil
}

func (r *VisitRepository) GenerateNoAntrian(poliID int) (string, error) {
	prefix := "A"
	if poliID == 2 {
		prefix = "A" // Poli Anak
	} else if poliID == 1 {
		prefix = "U" // Poli Umum
	} else if poliID == 3 {
		prefix = "G" // Poli Gigi
	} else if poliID == 4 {
		prefix = "F" // Fisioterapi
	} else {
		prefix = "Q"
	}

	var count int
	query := `SELECT COUNT(*) FROM visits WHERE visit_date = CURRENT_DATE AND poli_id = $1`
	err := r.db.Get(&count, query, poliID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%03d", prefix, count+1), nil
}

func (r *VisitRepository) FindAll(date, status string, poliID int, search string, page, limit int) ([]domain.Visit, int, error) {
	offset := (page - 1) * limit
	whereClauses := []string{"v.deleted_at IS NULL"}
	args := []interface{}{}
	argIndex := 1

	if date != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("v.visit_date = $%d", argIndex))
		args = append(args, date)
		argIndex++
	}

	if status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("v.status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	if poliID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("v.poli_id = $%d", argIndex))
		args = append(args, poliID)
		argIndex++
	}

	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(v.no_rawat ILIKE $%d OR v.no_antrian ILIKE $%d OR p.no_rm ILIKE $%d OR p.nama_pasien ILIKE $%d)", argIndex, argIndex, argIndex, argIndex))
		args = append(args, "%"+search+"%")
		argIndex++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM visits v
		JOIN patients p ON v.patient_id = p.id
		WHERE %s
	`, whereSQL)

	var total int
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT v.id, v.no_rawat, v.no_antrian, v.patient_id, v.poli_id, v.doctor_id, v.guarantor_id,
		       v.nomor_asuransi, TO_CHAR(v.visit_date, 'YYYY-MM-DD') AS visit_date, COALESCE(v.keluhan_utama, '') AS keluhan_utama,
		       v.status, v.waktu_masuk, v.waktu_selesai, v.created_at, v.updated_at,
		       p.nama_pasien AS patient_name, p.no_rm, p.jenis_kelamin AS patient_gender,
		       TO_CHAR(p.tanggal_lahir, 'YYYY-MM-DD') AS patient_dob, p.riwayat_alergi AS patient_allergy,
		       pu.name AS poli_name, COALESCE(u.full_name, '') AS doctor_name, g.name AS guarantor_name
		FROM visits v
		JOIN patients p ON v.patient_id = p.id
		JOIN poli_units pu ON v.poli_id = pu.id
		LEFT JOIN users u ON v.doctor_id = u.id
		JOIN guarantors g ON v.guarantor_id = g.id
		WHERE %s
		ORDER BY v.waktu_masuk ASC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIndex, argIndex+1)

	args = append(args, limit, offset)
	var visits []domain.Visit
	err = r.db.Select(&visits, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return visits, total, nil
}

func (r *VisitRepository) FindByID(id string) (*domain.Visit, error) {
	query := `
		SELECT v.id, v.no_rawat, v.no_antrian, v.patient_id, v.poli_id, v.doctor_id, v.guarantor_id,
		       v.nomor_asuransi, TO_CHAR(v.visit_date, 'YYYY-MM-DD') AS visit_date, COALESCE(v.keluhan_utama, '') AS keluhan_utama,
		       v.status, v.waktu_masuk, v.waktu_selesai, v.created_at, v.updated_at,
		       p.nama_pasien AS patient_name, p.no_rm, p.jenis_kelamin AS patient_gender,
		       TO_CHAR(p.tanggal_lahir, 'YYYY-MM-DD') AS patient_dob, p.riwayat_alergi AS patient_allergy,
		       pu.name AS poli_name, COALESCE(u.full_name, '') AS doctor_name, g.name AS guarantor_name
		FROM visits v
		JOIN patients p ON v.patient_id = p.id
		JOIN poli_units pu ON v.poli_id = pu.id
		LEFT JOIN users u ON v.doctor_id = u.id
		JOIN guarantors g ON v.guarantor_id = g.id
		WHERE v.id = $1 AND v.deleted_at IS NULL
	`
	var v domain.Visit
	err := r.db.Get(&v, query, id)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *VisitRepository) Create(v *domain.Visit) error {
	if v.NoRawat == "" {
		noRawat, err := r.GenerateNoRawat()
		if err != nil {
			return err
		}
		v.NoRawat = noRawat
	}

	if v.NoAntrian == "" {
		noAntrian, err := r.GenerateNoAntrian(v.PoliID)
		if err != nil {
			return err
		}
		v.NoAntrian = noAntrian
	}

	query := `
		INSERT INTO visits (no_rawat, no_antrian, patient_id, poli_id, doctor_id, guarantor_id,
		                    nomor_asuransi, visit_date, keluhan_utama, status, waktu_masuk,
		                    created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_DATE, $8, 'MENUNGGU_TRIASE', NOW(), $9, NOW(), NOW())
		RETURNING id, created_at, updated_at, waktu_masuk
	`
	return r.db.QueryRow(query, v.NoRawat, v.NoAntrian, v.PatientID, v.PoliID, v.DoctorID, v.GuarantorID,
		v.NomorAsuransi, v.KeluhanUtama, v.CreatedBy).
		Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt, &v.WaktuMasuk)
}

func (r *VisitRepository) UpdateStatus(id, status string) error {
	var query string
	if status == "SELESAI" {
		query = `UPDATE visits SET status = $1, waktu_selesai = NOW(), updated_at = NOW() WHERE id = $2`
	} else {
		query = `UPDATE visits SET status = $1, updated_at = NOW() WHERE id = $2`
	}
	_, err := r.db.Exec(query, status, id)
	return err
}

func (r *VisitRepository) GetQueueMetrics() (map[string]int, error) {
	query := `
		SELECT status, COUNT(*) AS count 
		FROM visits 
		WHERE visit_date = CURRENT_DATE AND deleted_at IS NULL
		GROUP BY status
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := map[string]int{
		"MENUNGGU_TRIASE":  0,
		"DIPERIKSA_DOKTER": 0,
		"FARMASI":          0,
		"KASIR":            0,
		"SELESAI":          0,
		"TOTAL":            0,
	}

	for rows.Next() {
		var status string
		var cnt int
		if err := rows.Scan(&status, &cnt); err == nil {
			counts[status] = cnt
			counts["TOTAL"] += cnt
		}
	}

	return counts, nil
}
