package repository

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"simrs-asoy/backend/internal/domain"
)

type PharmacyRepository struct {
	db *sqlx.DB
}

func NewPharmacyRepository(db *sqlx.DB) *PharmacyRepository {
	return &PharmacyRepository{db: db}
}

func (r *PharmacyRepository) GenerateNomorResep() (string, error) {
	dateStr := time.Now().Format("20060102")
	var count int
	query := `SELECT COUNT(*) FROM prescriptions WHERE nomor_resep LIKE $1`
	err := r.db.Get(&count, query, fmt.Sprintf("RXP-%s-%%", dateStr))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("RXP-%s-%03d", dateStr, count+1), nil
}

func (r *PharmacyRepository) FindPrescriptions(status, search string, page, limit int) ([]domain.Prescription, int, error) {
	offset := (page - 1) * limit
	whereClauses := []string{"pr.deleted_at IS NULL"}
	args := []interface{}{}
	argIndex := 1

	if status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("pr.status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(pr.nomor_resep ILIKE $%d OR p.no_rm ILIKE $%d OR p.nama_pasien ILIKE $%d)", argIndex, argIndex, argIndex))
		args = append(args, "%"+search+"%")
		argIndex++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM prescriptions pr
		JOIN patients p ON pr.patient_id = p.id
		WHERE %s
	`, whereSQL)

	var total int
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT pr.id, pr.nomor_resep, pr.visit_id, pr.patient_id, pr.doctor_id,
		       pr.status, COALESCE(pr.catatan_dokter, '') AS catatan_dokter, COALESCE(pr.catatan_telaah, '') AS catatan_telaah, pr.is_verified,
		       pr.verified_by, pr.verified_at, pr.dispensed_by, pr.dispensed_at, pr.created_at,
		       p.nama_pasien AS patient_name, p.no_rm,
		       COALESCE(u.full_name, '') AS doctor_name
		FROM prescriptions pr
		JOIN patients p ON pr.patient_id = p.id
		JOIN users u ON pr.doctor_id = u.id
		WHERE %s
		ORDER BY pr.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIndex, argIndex+1)

	args = append(args, limit, offset)
	var list []domain.Prescription
	err = r.db.Select(&list, query, args...)
	if err != nil {
		return nil, 0, err
	}

	// Fetch items for each prescription
	for i := range list {
		itemsQuery := `
			SELECT pi.id, pi.prescription_id, pi.item_id, ii.item_name,
			       COALESCE(pi.signa, '') AS signa, COALESCE(pi.aturan_pakai, '') AS aturan_pakai, pi.jumlah, pi.harga_satuan, pi.subtotal,
			       pi.is_racikan, COALESCE(pi.racikan_nama, '') AS racikan_nama, pi.created_at
			FROM prescription_items pi
			JOIN inventory_items ii ON pi.item_id = ii.id
			WHERE pi.prescription_id = $1
		`
		var items []domain.PrescriptionItem
		_ = r.db.Select(&items, itemsQuery, list[i].ID)
		list[i].Items = items
	}

	return list, total, nil
}

func (r *PharmacyRepository) GetPrescriptionByID(id string) (*domain.Prescription, error) {
	query := `
		SELECT pr.id, pr.nomor_resep, pr.visit_id, pr.patient_id, pr.doctor_id,
		       pr.status, COALESCE(pr.catatan_dokter, '') AS catatan_dokter, COALESCE(pr.catatan_telaah, '') AS catatan_telaah, pr.is_verified,
		       pr.verified_by, pr.verified_at, pr.dispensed_by, pr.dispensed_at, pr.created_at,
		       p.nama_pasien AS patient_name, p.no_rm,
		       COALESCE(u.full_name, '') AS doctor_name
		FROM prescriptions pr
		JOIN patients p ON pr.patient_id = p.id
		JOIN users u ON pr.doctor_id = u.id
		WHERE pr.id = $1 AND pr.deleted_at IS NULL
	`
	var p domain.Prescription
	err := r.db.Get(&p, query, id)
	if err != nil {
		return nil, err
	}

	itemsQuery := `
		SELECT pi.id, pi.prescription_id, pi.item_id, ii.item_name,
		       COALESCE(pi.signa, '') AS signa, COALESCE(pi.aturan_pakai, '') AS aturan_pakai, pi.jumlah, pi.harga_satuan, pi.subtotal,
		       pi.is_racikan, COALESCE(pi.racikan_nama, '') AS racikan_nama, pi.created_at
		FROM prescription_items pi
		JOIN inventory_items ii ON pi.item_id = ii.id
		WHERE pi.prescription_id = $1
	`
	_ = r.db.Select(&p.Items, itemsQuery, p.ID)

	return &p, nil
}

func (r *PharmacyRepository) GetPrescriptionByVisitID(visitID string) (*domain.Prescription, error) {
	query := `
		SELECT pr.id, pr.nomor_resep, pr.visit_id, pr.patient_id, pr.doctor_id,
		       pr.status, COALESCE(pr.catatan_dokter, '') AS catatan_dokter, COALESCE(pr.catatan_telaah, '') AS catatan_telaah, pr.is_verified,
		       pr.verified_by, pr.verified_at, pr.dispensed_by, pr.dispensed_at, pr.created_at,
		       p.nama_pasien AS patient_name, p.no_rm,
		       COALESCE(u.full_name, '') AS doctor_name
		FROM prescriptions pr
		JOIN patients p ON pr.patient_id = p.id
		JOIN users u ON pr.doctor_id = u.id
		WHERE pr.visit_id = $1 AND pr.deleted_at IS NULL
		ORDER BY pr.created_at DESC
		LIMIT 1
	`
	var p domain.Prescription
	err := r.db.Get(&p, query, visitID)
	if err != nil {
		return nil, err
	}

	itemsQuery := `
		SELECT pi.id, pi.prescription_id, pi.item_id, ii.item_name,
		       COALESCE(pi.signa, '') AS signa, COALESCE(pi.aturan_pakai, '') AS aturan_pakai, pi.jumlah, pi.harga_satuan, pi.subtotal,
		       pi.is_racikan, COALESCE(pi.racikan_nama, '') AS racikan_nama, pi.created_at
		FROM prescription_items pi
		JOIN inventory_items ii ON pi.item_id = ii.id
		WHERE pi.prescription_id = $1
	`
	_ = r.db.Select(&p.Items, itemsQuery, p.ID)

	return &p, nil
}

func (r *PharmacyRepository) CreatePrescription(p *domain.Prescription, items []domain.PrescriptionItemDTO) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if p.NomorResep == "" {
		noResep, err := r.GenerateNomorResep()
		if err != nil {
			return err
		}
		p.NomorResep = noResep
	}

	query := `
		INSERT INTO prescriptions (nomor_resep, visit_id, patient_id, doctor_id, status, catatan_dokter, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'MENUNGGU_TELAAH', $5, $6, NOW(), NOW())
		RETURNING id, created_at
	`
	err = tx.QueryRow(query, p.NomorResep, p.VisitID, p.PatientID, p.DoctorID, p.CatatanDokter, p.DoctorID).
		Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		return err
	}

	for _, it := range items {
		// fetch price from inventory
		var sellPrice float64
		_ = tx.Get(&sellPrice, "SELECT sell_price FROM inventory_items WHERE id = $1", it.ItemID)
		subtotal := sellPrice * float64(it.Jumlah)

		itemQuery := `
			INSERT INTO prescription_items (prescription_id, item_id, signa, aturan_pakai, jumlah, harga_satuan, subtotal, is_racikan, racikan_nama, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		`
		_, err = tx.Exec(itemQuery, p.ID, it.ItemID, it.Signa, it.AturanPakai, it.Jumlah, sellPrice, subtotal, it.IsRacikan, it.RacikanNama)
		if err != nil {
			return err
		}
	}

	// Update visit status to FARMASI
	_, _ = tx.Exec("UPDATE visits SET status = 'FARMASI', updated_at = NOW() WHERE id = $1", p.VisitID)

	return tx.Commit()
}

func (r *PharmacyRepository) ReviewPrescription(id, catatanTelaah, verifiedBy string, status string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Update prescription status and verification
	query := `
		UPDATE prescriptions 
		SET status = $1, catatan_telaah = $2, is_verified = TRUE, verified_by = $3, verified_at = NOW(), updated_at = NOW()
		WHERE id = $4
	`
	_, err = tx.Exec(query, status, catatanTelaah, verifiedBy, id)
	if err != nil {
		return err
	}

	// 2. Automatically deduct stock upon verification
	var items []domain.PrescriptionItem
	err = tx.Select(&items, "SELECT item_id, jumlah FROM prescription_items WHERE prescription_id = $1", id)
	if err != nil {
		return err
	}

	var nomorResep string
	_ = tx.Get(&nomorResep, "SELECT nomor_resep FROM prescriptions WHERE id = $1", id)

	for _, it := range items {
		_, err = tx.Exec(`
			UPDATE inventory_items 
			SET depo_stock = GREATEST(0, depo_stock - $1), 
			    total_stock = GREATEST(0, total_stock - $1), 
			    updated_at = NOW() 
			WHERE id = $2
		`, it.Jumlah, it.ItemID)
		if err != nil {
			return err
		}

		// Log into stock mutations
		_, _ = tx.Exec(`
			INSERT INTO stock_mutations (item_id, transaction_type, qty, source_depo, dest_depo, notes, created_by, created_at)
			VALUES ($1, 'OUT_PRESCRIPTION', $2, 'DEPO_FARMASI', 'PASIEN', $3, $4, NOW())
		`, it.ItemID, it.Jumlah, "Verifikasi Resep Dokter: "+nomorResep, verifiedBy)
	}

	return tx.Commit()
}

func (r *PharmacyRepository) DispensePrescription(id, dispensedBy string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get items
	var items []domain.PrescriptionItem
	err = tx.Select(&items, "SELECT item_id, jumlah FROM prescription_items WHERE prescription_id = $1", id)
	if err != nil {
		return err
	}

	// Deduct stock
	for _, it := range items {
		_, err = tx.Exec("UPDATE inventory_items SET depo_stock = GREATEST(0, depo_stock - $1), total_stock = GREATEST(0, total_stock - $1), updated_at = NOW() WHERE id = $2", it.Jumlah, it.ItemID)
		if err != nil {
			return err
		}

		// Log mutation
		_, _ = tx.Exec(`
			INSERT INTO stock_mutations (item_id, transaction_type, qty, source_depo, dest_depo, notes, created_by, created_at)
			VALUES ($1, 'OUT_DISPENSE', $2, 'DEPO_FARMASI', 'PASIEN', 'Pengeluaran obat e-Resep', $3, NOW())
		`, it.ItemID, it.Jumlah, dispensedBy)
	}

	// Update prescription
	query := `
		UPDATE prescriptions 
		SET status = 'SELESAI', dispensed_by = $1, dispensed_at = NOW(), updated_at = NOW()
		WHERE id = $2
		RETURNING visit_id
	`
	var visitID string
	err = tx.QueryRow(query, dispensedBy, id).Scan(&visitID)
	if err != nil {
		return err
	}

	// Progress visit to KASIR
	_, _ = tx.Exec("UPDATE visits SET status = 'KASIR', updated_at = NOW() WHERE id = $1", visitID)

	return tx.Commit()
}

// -------------------------------------------------------------
// Retur Obat Pasien (Fase 1)
// -------------------------------------------------------------

func (r *PharmacyRepository) CreateMedicineReturn(ret *domain.MedicineReturn, items []domain.MedicineReturnItemDTO, currentUserID string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	dateStr := time.Now().Format("20060102")
	var count int
	_ = tx.Get(&count, "SELECT COUNT(*) FROM medicine_returns WHERE nomor_retur LIKE $1", fmt.Sprintf("RET-%s-%%", dateStr))
	ret.NomorRetur = fmt.Sprintf("RET-%s-%03d", dateStr, count+1)

	// Auto-link active invoice for the visit if not explicitly supplied
	if ret.InvoiceID == nil {
		var invID string
		_ = tx.Get(&invID, "SELECT id FROM invoices WHERE visit_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 1", ret.VisitID)
		if invID != "" {
			ret.InvoiceID = &invID
		}
	}

	var totalRefund float64
	for _, it := range items {
		totalRefund += float64(it.Qty) * it.UnitPrice
	}
	ret.TotalRefund = totalRefund

	query := `
		INSERT INTO medicine_returns (nomor_retur, prescription_id, visit_id, patient_id, invoice_id,
		                              alasan, total_refund, status, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'COMPLETED', $8, NOW())
		RETURNING id, created_at
	`
	err = tx.QueryRow(query, ret.NomorRetur, ret.PrescriptionID, ret.VisitID, ret.PatientID, ret.InvoiceID,
		ret.Alasan, ret.TotalRefund, currentUserID).Scan(&ret.ID, &ret.CreatedAt)
	if err != nil {
		return err
	}

	for _, it := range items {
		subtotal := float64(it.Qty) * it.UnitPrice

		// ACID Row-Level Locking on inventory_items
		var currentStock int
		_ = tx.Get(&currentStock, "SELECT depo_stock FROM inventory_items WHERE id = $1 FOR UPDATE", it.ItemID)

		// Restock back to Depo Farmasi & Total Stock
		_, err = tx.Exec(`
			UPDATE inventory_items 
			SET depo_stock = depo_stock + $1, total_stock = total_stock + $1, updated_at = NOW() 
			WHERE id = $2
		`, it.Qty, it.ItemID)
		if err != nil {
			return err
		}

		// Insert stock mutation IN_RETUR
		_, _ = tx.Exec(`
			INSERT INTO stock_mutations (item_id, transaction_type, qty, source_depo, dest_depo, notes, created_by, created_at)
			VALUES ($1, 'IN_RETUR', $2, 'PASIEN', 'DEPO_FARMASI', $3, $4, NOW())
		`, it.ItemID, it.Qty, "Retur obat pasien: "+ret.Alasan, currentUserID)

		// Insert return item
		_, err = tx.Exec(`
			INSERT INTO medicine_return_items (return_id, item_id, item_name, qty, unit_price, subtotal, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW())
		`, ret.ID, it.ItemID, it.ItemName, it.Qty, it.UnitPrice, subtotal)
		if err != nil {
			return err
		}
	}

	// If invoice exists, adjust invoice total
	if ret.InvoiceID != nil && *ret.InvoiceID != "" {
		_, _ = tx.Exec(`
			UPDATE invoices
			SET subtotal = GREATEST(0, subtotal - $1),
			    total_tagihan = GREATEST(0, total_tagihan - $1),
			    sisa_tagihan = GREATEST(0, sisa_tagihan - $1),
			    updated_at = NOW()
			WHERE id = $2
		`, ret.TotalRefund, *ret.InvoiceID)
	}

	return tx.Commit()
}

func (r *PharmacyRepository) GetMedicineReturns(search string, page, limit int) ([]domain.MedicineReturn, int, error) {
	offset := (page - 1) * limit
	where := []string{"1=1"}
	args := []interface{}{}
	idx := 1

	if search != "" {
		where = append(where, fmt.Sprintf("(mr.nomor_retur ILIKE $%d OR p.nama_pasien ILIKE $%d OR p.no_rm ILIKE $%d)", idx, idx, idx))
		args = append(args, "%"+search+"%")
		idx++
	}

	whereSQL := strings.Join(where, " AND ")
	var total int
	err := r.db.Get(&total, fmt.Sprintf("SELECT COUNT(*) FROM medicine_returns mr JOIN patients p ON mr.patient_id = p.id WHERE %s", whereSQL), args...)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT mr.id, mr.nomor_retur, mr.prescription_id, mr.visit_id, v.no_rawat,
		       mr.patient_id, p.nama_pasien AS patient_name, mr.invoice_id,
		       mr.alasan, mr.total_refund, mr.status, mr.created_by, mr.created_at
		FROM medicine_returns mr
		JOIN patients p ON mr.patient_id = p.id
		JOIN visits v ON mr.visit_id = v.id
		WHERE %s
		ORDER BY mr.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, idx, idx+1)
	args = append(args, limit, offset)

	var list []domain.MedicineReturn
	err = r.db.Select(&list, query, args...)
	if err != nil {
		return nil, 0, err
	}

	// Fetch items for each return
	for i := range list {
		var itms []domain.MedicineReturnItem
		_ = r.db.Select(&itms, "SELECT id, return_id, item_id, item_name, qty, unit_price, subtotal, created_at FROM medicine_return_items WHERE return_id = $1", list[i].ID)
		list[i].Items = itms
	}

	return list, total, nil
}
