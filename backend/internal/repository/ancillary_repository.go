package repository

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"simrs-asoy/backend/internal/domain"
)

type AncillaryRepository struct {
	db *sqlx.DB
}

func NewAncillaryRepository(db *sqlx.DB) *AncillaryRepository {
	return &AncillaryRepository{db: db}
}

func (r *AncillaryRepository) GenerateNomorOrder() (string, error) {
	dateStr := time.Now().Format("20060102")
	var count int
	query := `SELECT COUNT(*) FROM ancillary_orders WHERE nomor_order LIKE $1`
	err := r.db.Get(&count, query, fmt.Sprintf("ORD-%s-%%", dateStr))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("ORD-%s-%03d", dateStr, count+1), nil
}

func (r *AncillaryRepository) FindOrders(status, jenisOrder, search string, page, limit int) ([]domain.AncillaryOrder, int, error) {
	offset := (page - 1) * limit
	whereClauses := []string{"ao.deleted_at IS NULL"}
	args := []interface{}{}
	argIndex := 1

	if status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("ao.status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	if jenisOrder != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("ao.jenis_order = $%d", argIndex))
		args = append(args, jenisOrder)
		argIndex++
	}

	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(ao.nomor_order ILIKE $%d OR p.no_rm ILIKE $%d OR p.nama_pasien ILIKE $%d OR ao.nama_tindakan ILIKE $%d)", argIndex, argIndex, argIndex, argIndex))
		args = append(args, "%"+search+"%")
		argIndex++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM ancillary_orders ao
		JOIN patients p ON ao.patient_id = p.id
		WHERE %s
	`, whereSQL)

	var total int
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT ao.id, ao.nomor_order, ao.visit_id, ao.patient_id, ao.doctor_id,
		       ao.jenis_order, ao.nama_tindakan, ao.tarif, COALESCE(ao.catatan_klinis, '') AS catatan_klinis,
		       COALESCE(ao.hasil_pemeriksaan, '') AS hasil_pemeriksaan,
		       COALESCE(ao.lampiran_url, '') AS lampiran_url,
		       ao.status,
		       ao.verified_by, ao.verified_at, ao.created_at,
		       p.nama_pasien AS patient_name, p.no_rm,
		       COALESCE(u.full_name, '') AS doctor_name
		FROM ancillary_orders ao
		JOIN patients p ON ao.patient_id = p.id
		JOIN users u ON ao.doctor_id = u.id
		WHERE %s
		ORDER BY ao.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIndex, argIndex+1)

	args = append(args, limit, offset)
	var list []domain.AncillaryOrder
	err = r.db.Select(&list, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *AncillaryRepository) GetOrderByID(id string) (*domain.AncillaryOrder, error) {
	query := `
		SELECT ao.id, ao.nomor_order, ao.visit_id, ao.patient_id, ao.doctor_id,
		       ao.jenis_order, ao.nama_tindakan, ao.tarif, COALESCE(ao.catatan_klinis, '') AS catatan_klinis,
		       COALESCE(ao.hasil_pemeriksaan, '') AS hasil_pemeriksaan,
		       COALESCE(ao.lampiran_url, '') AS lampiran_url,
		       ao.status,
		       ao.verified_by, ao.verified_at, ao.created_at,
		       p.nama_pasien AS patient_name, p.no_rm,
		       COALESCE(u.full_name, '') AS doctor_name
		FROM ancillary_orders ao
		JOIN patients p ON ao.patient_id = p.id
		JOIN users u ON ao.doctor_id = u.id
		WHERE ao.id = $1 AND ao.deleted_at IS NULL
	`
	var o domain.AncillaryOrder
	err := r.db.Get(&o, query, id)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *AncillaryRepository) GetOrdersByVisitID(visitID string) ([]domain.AncillaryOrder, error) {
	query := `
		SELECT ao.id, ao.nomor_order, ao.visit_id, ao.patient_id, ao.doctor_id,
		       ao.jenis_order, ao.nama_tindakan, ao.tarif, COALESCE(ao.catatan_klinis, '') AS catatan_klinis,
		       COALESCE(ao.hasil_pemeriksaan, '') AS hasil_pemeriksaan,
		       COALESCE(ao.lampiran_url, '') AS lampiran_url,
		       ao.status,
		       ao.verified_by, ao.verified_at, ao.created_at,
		       p.nama_pasien AS patient_name, p.no_rm,
		       COALESCE(u.full_name, '') AS doctor_name
		FROM ancillary_orders ao
		JOIN patients p ON ao.patient_id = p.id
		JOIN users u ON ao.doctor_id = u.id
		WHERE ao.visit_id = $1 AND ao.deleted_at IS NULL
		ORDER BY ao.created_at DESC
	`
	var list []domain.AncillaryOrder
	err := r.db.Select(&list, query, visitID)
	return list, err
}

func (r *AncillaryRepository) CreateOrder(ao *domain.AncillaryOrder) error {
	if ao.NomorOrder == "" {
		noOrder, err := r.GenerateNomorOrder()
		if err != nil {
			return err
		}
		ao.NomorOrder = noOrder
	}

	query := `
		INSERT INTO ancillary_orders (nomor_order, visit_id, patient_id, doctor_id, jenis_order,
		                              nama_tindakan, tarif, catatan_klinis, status, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'MENUNGGU', $9, NOW(), NOW())
		RETURNING id, created_at
	`
	return r.db.QueryRow(query, ao.NomorOrder, ao.VisitID, ao.PatientID, ao.DoctorID, ao.JenisOrder,
		ao.NamaTindakan, ao.Tarif, ao.CatatanKlinis, ao.DoctorID).Scan(&ao.ID, &ao.CreatedAt)
}

func (r *AncillaryRepository) UpdateResult(id, hasil, lampiranURL, verifiedBy string) error {
	query := `
		UPDATE ancillary_orders
		SET hasil_pemeriksaan = $1, lampiran_url = $2, status = 'SELESAI',
		    verified_by = $3, verified_at = NOW(), updated_at = NOW()
		WHERE id = $4
	`
	_, err := r.db.Exec(query, hasil, lampiranURL, verifiedBy, id)
	return err
}
