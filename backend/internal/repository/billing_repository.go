package repository

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"simrs-asoy/backend/internal/domain"
)

type BillingRepository struct {
	db *sqlx.DB
}

func NewBillingRepository(db *sqlx.DB) *BillingRepository {
	return &BillingRepository{db: db}
}

func (r *BillingRepository) GenerateNomorInvoice() (string, error) {
	dateStr := time.Now().Format("20060102")
	var count int
	query := `SELECT COUNT(*) FROM invoices WHERE nomor_invoice LIKE $1`
	err := r.db.Get(&count, query, fmt.Sprintf("INV-%s-%%", dateStr))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("INV-%s-%03d", dateStr, count+1), nil
}

func (r *BillingRepository) GenerateNomorTransaksi() (string, error) {
	dateStr := time.Now().Format("20060102")
	var count int
	query := `SELECT COUNT(*) FROM billing_payments WHERE nomor_transaksi LIKE $1`
	err := r.db.Get(&count, query, fmt.Sprintf("TRX-%s-%%", dateStr))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("TRX-%s-%03d", dateStr, count+1), nil
}

func (r *BillingRepository) FindInvoices(status, search string, page, limit int) ([]domain.Invoice, int, error) {
	offset := (page - 1) * limit
	whereClauses := []string{"inv.deleted_at IS NULL"}
	args := []interface{}{}
	argIndex := 1

	if status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("inv.status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(inv.nomor_invoice ILIKE $%d OR p.no_rm ILIKE $%d OR p.nama_pasien ILIKE $%d OR v.no_rawat ILIKE $%d)", argIndex, argIndex, argIndex, argIndex))
		args = append(args, "%"+search+"%")
		argIndex++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM invoices inv
		JOIN patients p ON inv.patient_id = p.id
		JOIN visits v ON inv.visit_id = v.id
		WHERE %s
	`, whereSQL)

	var total int
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT inv.id, inv.nomor_invoice, inv.visit_id, inv.patient_id,
		       inv.subtotal, inv.diskon, inv.pajak, inv.total_tagihan,
		       inv.total_terbayar, inv.sisa_tagihan, inv.status,
		       inv.jenis_penjamin, inv.nomor_penjamin, COALESCE(inv.catatan, '') AS catatan,
		       COALESCE(inv.promo_code, '') AS promo_code, COALESCE(inv.promo_name, '') AS promo_name,
		       inv.created_at,
		       p.nama_pasien AS patient_name, p.no_rm, COALESCE(v.no_rawat, '') AS no_rawat
		FROM invoices inv
		JOIN patients p ON inv.patient_id = p.id
		JOIN visits v ON inv.visit_id = v.id
		WHERE %s
		ORDER BY inv.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIndex, argIndex+1)

	args = append(args, limit, offset)
	var list []domain.Invoice
	err = r.db.Select(&list, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *BillingRepository) GetInvoiceByID(id string) (*domain.Invoice, error) {
	query := `
		SELECT inv.id, inv.nomor_invoice, inv.visit_id, inv.patient_id,
		       inv.subtotal, inv.diskon, inv.pajak, inv.total_tagihan,
		       inv.total_terbayar, inv.sisa_tagihan, inv.status,
		       inv.jenis_penjamin, inv.nomor_penjamin, COALESCE(inv.catatan, '') AS catatan,
		       COALESCE(inv.promo_code, '') AS promo_code, COALESCE(inv.promo_name, '') AS promo_name,
		       inv.created_at,
		       p.nama_pasien AS patient_name, p.no_rm, COALESCE(v.no_rawat, '') AS no_rawat
		FROM invoices inv
		JOIN patients p ON inv.patient_id = p.id
		JOIN visits v ON inv.visit_id = v.id
		WHERE inv.id = $1 AND inv.deleted_at IS NULL
	`
	var inv domain.Invoice
	err := r.db.Get(&inv, query, id)
	if err != nil {
		return nil, err
	}

	itemsQuery := `
		SELECT id, invoice_id, item_type, reference_id, item_name, qty, unit_price, subtotal, created_at
		FROM invoice_items
		WHERE invoice_id = $1
		ORDER BY created_at ASC
	`
	_ = r.db.Select(&inv.Items, itemsQuery, inv.ID)

	paymentsQuery := `
		SELECT bp.id, bp.nomor_transaksi, bp.invoice_id, bp.kasir_id,
		       COALESCE(u.full_name, '') as kasir_name,
		       bp.nominal, bp.metode_bayar,
		       COALESCE(bp.referensi_pembayaran, '') as referensi_pembayaran,
		       COALESCE(bp.catatan, '') as catatan,
		       COALESCE(bp.uang_diterima, bp.nominal) as uang_diterima,
		       COALESCE(bp.kembalian, 0) as kembalian,
		       bp.created_at
		FROM billing_payments bp
		LEFT JOIN users u ON bp.kasir_id = u.id
		WHERE bp.invoice_id = $1
		ORDER BY bp.created_at ASC
	`
	_ = r.db.Select(&inv.Payments, paymentsQuery, inv.ID)

	return &inv, nil
}

func (r *BillingRepository) GetInvoiceByVisitID(visitID string) (*domain.Invoice, error) {
	query := `
		SELECT inv.id, inv.nomor_invoice, inv.visit_id, inv.patient_id,
		       inv.subtotal, inv.diskon, inv.pajak, inv.total_tagihan,
		       inv.total_terbayar, inv.sisa_tagihan, inv.status,
		       inv.jenis_penjamin, inv.nomor_penjamin, COALESCE(inv.catatan, '') AS catatan,
		       COALESCE(inv.promo_code, '') AS promo_code, COALESCE(inv.promo_name, '') AS promo_name,
		       inv.created_at,
		       p.nama_pasien AS patient_name, p.no_rm, COALESCE(v.no_rawat, '') AS no_rawat
		FROM invoices inv
		JOIN patients p ON inv.patient_id = p.id
		JOIN visits v ON inv.visit_id = v.id
		WHERE inv.visit_id = $1 AND inv.deleted_at IS NULL
		ORDER BY inv.created_at DESC
		LIMIT 1
	`
	var inv domain.Invoice
	err := r.db.Get(&inv, query, visitID)
	if err != nil {
		return nil, err
	}

	itemsQuery := `
		SELECT id, invoice_id, item_type, reference_id, item_name, qty, unit_price, subtotal, created_at
		FROM invoice_items
		WHERE invoice_id = $1
		ORDER BY created_at ASC
	`
	_ = r.db.Select(&inv.Items, itemsQuery, inv.ID)

	return &inv, nil
}

func (r *BillingRepository) CreateInvoice(inv *domain.Invoice, items []domain.InvoiceItemDTO) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if inv.NomorInvoice == "" {
		noInv, err := r.GenerateNomorInvoice()
		if err != nil {
			return err
		}
		inv.NomorInvoice = noInv
	}

	var subtotal float64
	for _, it := range items {
		subtotal += it.UnitPrice * float64(it.Qty)
	}

	total := subtotal - inv.Diskon + inv.Pajak
	if total < 0 {
		total = 0
	}
	inv.Subtotal = subtotal
	inv.TotalTagihan = total
	inv.SisaTagihan = total
	inv.Status = "BELUM_BAYAR"

	var createdBy *string
	if inv.CreatedBy != nil && *inv.CreatedBy != "" {
		createdBy = inv.CreatedBy
	}

	query := `
		INSERT INTO invoices (nomor_invoice, visit_id, patient_id, subtotal, diskon, pajak,
		                      total_tagihan, total_terbayar, sisa_tagihan, status, jenis_penjamin,
		                      nomor_penjamin, catatan, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 0, $8, $9, $10, $11, $12, $13, NOW(), NOW())
		RETURNING id, created_at
	`
	err = tx.QueryRow(query, inv.NomorInvoice, inv.VisitID, inv.PatientID, inv.Subtotal, inv.Diskon, inv.Pajak,
		inv.TotalTagihan, inv.SisaTagihan, inv.Status, inv.JenisPenjamin, inv.NomorPenjamin, inv.Catatan, createdBy).
		Scan(&inv.ID, &inv.CreatedAt)
	if err != nil {
		return err
	}

	for _, it := range items {
		itemSubtotal := it.UnitPrice * float64(it.Qty)
		itemQuery := `
			INSERT INTO invoice_items (invoice_id, item_type, reference_id, item_name, qty, unit_price, subtotal, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		`
		_, err = tx.Exec(itemQuery, inv.ID, it.ItemType, it.ReferenceID, it.ItemName, it.Qty, it.UnitPrice, itemSubtotal)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *BillingRepository) AddPayment(p *domain.BillingPayment, promoCode, promoName string, promoDiscount float64) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if p.NomorTransaksi == "" {
		noTrx, err := r.GenerateNomorTransaksi()
		if err != nil {
			return err
		}
		p.NomorTransaksi = noTrx
	}

	// Insert payment with tendered cash and change
	query := `
		INSERT INTO billing_payments (nomor_transaksi, invoice_id, kasir_id, nominal, metode_bayar,
		                              referensi_pembayaran, catatan, uang_diterima, kembalian, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		RETURNING id, created_at
	`
	err = tx.QueryRow(query, p.NomorTransaksi, p.InvoiceID, p.KasirID, p.Nominal, p.MetodeBayar,
		p.ReferensiPembayaran, p.Catatan, p.UangDiterima, p.Kembalian).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		return err
	}

	// Update invoice totals & promo if any
	var inv domain.Invoice
	err = tx.Get(&inv, "SELECT id, visit_id, subtotal, diskon, pajak, total_tagihan, total_terbayar, sisa_tagihan FROM invoices WHERE id = $1", p.InvoiceID)
	if err != nil {
		return err
	}

	newDiskon := inv.Diskon
	if promoDiscount > 0 {
		newDiskon += promoDiscount
	}
	newTotalTagihan := inv.Subtotal - newDiskon + inv.Pajak
	if newTotalTagihan < 0 {
		newTotalTagihan = 0
	}

	// Calculate net amount applied to pay the bill
	paidForInvoice := p.Nominal
	if p.MetodeBayar == "TUNAI" && p.Kembalian > 0 {
		paidForInvoice = p.UangDiterima - p.Kembalian
	}
	newTerbayar := inv.TotalTerbayar + paidForInvoice
	newSisa := newTotalTagihan - newTerbayar
	newStatus := "BELUM_BAYAR"
	if newSisa <= 0 {
		newSisa = 0
		newStatus = "LUNAS"
	}

	_, err = tx.Exec(`
		UPDATE invoices
		SET diskon = $1, total_tagihan = $2, total_terbayar = $3, sisa_tagihan = $4, status = $5,
		    promo_code = CASE WHEN $6 != '' THEN $6 ELSE promo_code END,
		    promo_name = CASE WHEN $7 != '' THEN $7 ELSE promo_name END,
		    updated_at = NOW()
		WHERE id = $8
	`, newDiskon, newTotalTagihan, newTerbayar, newSisa, newStatus, promoCode, promoName, p.InvoiceID)
	if err != nil {
		return err
	}

	// If lunas, update visit to SELESAI
	if newStatus == "LUNAS" {
		_, _ = tx.Exec("UPDATE visits SET status = 'SELESAI', waktu_selesai = NOW(), updated_at = NOW() WHERE id = $1", inv.VisitID)
	}

	// Update active shift of cashier (net cash in drawer = uang_diterima - kembalian)
	if p.MetodeBayar == "TUNAI" {
		_, _ = tx.Exec("UPDATE cashier_shifts SET total_tunai_sistem = total_tunai_sistem + $1, updated_at = NOW() WHERE kasir_id = $2 AND status = 'OPEN'", paidForInvoice, p.KasirID)
	} else {
		_, _ = tx.Exec("UPDATE cashier_shifts SET total_non_tunai_sistem = total_non_tunai_sistem + $1, updated_at = NOW() WHERE kasir_id = $2 AND status = 'OPEN'", p.Nominal, p.KasirID)
	}

	return tx.Commit()
}

// Cashier Shifts
func (r *BillingRepository) GetActiveShift(kasirID string) (*domain.CashierShift, error) {
	query := `
		SELECT cs.id, cs.kasir_id, COALESCE(u.full_name, '') AS kasir_name,
		       cs.shift_name, cs.waktu_buka, cs.waktu_tutup, cs.saldo_awal,
		       cs.total_tunai_sistem, cs.total_non_tunai_sistem, cs.saldo_akhir_fisik,
		       cs.selisih, cs.status, COALESCE(cs.catatan_shift, '') AS catatan_shift, cs.created_at
		FROM cashier_shifts cs
		JOIN users u ON cs.kasir_id = u.id
		WHERE cs.kasir_id = $1 AND cs.status = 'OPEN'
		ORDER BY cs.created_at DESC
		LIMIT 1
	`
	var s domain.CashierShift
	err := r.db.Get(&s, query, kasirID)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *BillingRepository) OpenShift(s *domain.CashierShift) error {
	query := `
		INSERT INTO cashier_shifts (kasir_id, shift_name, waktu_buka, saldo_awal, status, created_at, updated_at)
		VALUES ($1, $2, NOW(), $3, 'OPEN', NOW(), NOW())
		RETURNING id, waktu_buka, created_at
	`
	return r.db.QueryRow(query, s.KasirID, s.ShiftName, s.SaldoAwal).Scan(&s.ID, &s.WaktuBuka, &s.CreatedAt)
}

func (r *BillingRepository) CloseShift(shiftID string, saldoAkhirFisik float64, catatan string) error {
	var current domain.CashierShift
	err := r.db.Get(&current, "SELECT saldo_awal, total_tunai_sistem FROM cashier_shifts WHERE id = $1", shiftID)
	if err != nil {
		return err
	}

	expectedCash := current.SaldoAwal + current.TotalTunaiSistem
	selisih := saldoAkhirFisik - expectedCash

	query := `
		UPDATE cashier_shifts
		SET waktu_tutup = NOW(), saldo_akhir_fisik = $1, selisih = $2,
		    status = 'CLOSED', catatan_shift = $3, updated_at = NOW()
		WHERE id = $4
	`
	_, err = r.db.Exec(query, saldoAkhirFisik, selisih, catatan, shiftID)
	return err
}

func (r *BillingRepository) GetShiftHistory(kasirID string, limit int) ([]domain.CashierShift, error) {
	query := `
		SELECT cs.id, cs.kasir_id, COALESCE(u.full_name, '') AS kasir_name,
		       cs.shift_name, cs.waktu_buka, cs.waktu_tutup, cs.saldo_awal,
		       cs.total_tunai_sistem, cs.total_non_tunai_sistem, cs.saldo_akhir_fisik,
		       cs.selisih, cs.status, COALESCE(cs.catatan_shift, '') AS catatan_shift, cs.created_at
		FROM cashier_shifts cs
		JOIN users u ON cs.kasir_id = u.id
		WHERE ($1 = '' OR cs.kasir_id::text = $1)
		ORDER BY cs.created_at DESC
		LIMIT $2
	`
	var list []domain.CashierShift
	err := r.db.Select(&list, query, kasirID, limit)
	return list, err
}

// -------------------------------------------------------------
// Split Payment Kasir & Reports (Fase 1)
// -------------------------------------------------------------

func (r *BillingRepository) AddSplitPayments(invoiceID string, payments []domain.SplitPaymentItem, promoCode, promoName string, promoDiscount, uangDiterima, kembalian float64, currentUserID, currentUsername string) (*domain.Invoice, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. Get current invoice with row lock
	var inv domain.Invoice
	err = tx.Get(&inv, "SELECT id, visit_id, subtotal, diskon, pajak, total_tagihan, total_terbayar, sisa_tagihan FROM invoices WHERE id = $1 FOR UPDATE", invoiceID)
	if err != nil {
		return nil, err
	}

	newDiskon := inv.Diskon
	if promoDiscount > 0 {
		newDiskon += promoDiscount
	}
	newTotalTagihan := inv.Subtotal - newDiskon + inv.Pajak
	if newTotalTagihan < 0 {
		newTotalTagihan = 0
	}

	var totalPaidNet float64
	dateStr := time.Now().Format("20060102")

	for i, pay := range payments {
		var trxCount int
		_ = tx.Get(&trxCount, "SELECT COUNT(*) FROM billing_payments WHERE nomor_transaksi LIKE $1", fmt.Sprintf("TRX-%s-%%", dateStr))
		nomorTrx := fmt.Sprintf("TRX-%s-%03d", dateStr, trxCount+1+i)

		payUangDiterima := pay.Nominal
		payKembalian := 0.0
		if pay.MetodeBayar == "TUNAI" && uangDiterima > 0 {
			payUangDiterima = uangDiterima
			payKembalian = kembalian
		}

		query := `
			INSERT INTO billing_payments (nomor_transaksi, invoice_id, kasir_id, nominal, metode_bayar,
			                              referensi_pembayaran, catatan, uang_diterima, kembalian, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		`
		_, err = tx.Exec(query, nomorTrx, invoiceID, currentUserID, pay.Nominal, pay.MetodeBayar,
			pay.ReferensiPembayaran, pay.Catatan, payUangDiterima, payKembalian)
		if err != nil {
			return nil, err
		}

		// Update shift
		paidNet := pay.Nominal
		if pay.MetodeBayar == "TUNAI" {
			_, _ = tx.Exec("UPDATE cashier_shifts SET total_tunai_sistem = total_tunai_sistem + $1, updated_at = NOW() WHERE kasir_id = $2 AND status = 'OPEN'", paidNet, currentUserID)
		} else {
			_, _ = tx.Exec("UPDATE cashier_shifts SET total_non_tunai_sistem = total_non_tunai_sistem + $1, updated_at = NOW() WHERE kasir_id = $2 AND status = 'OPEN'", paidNet, currentUserID)
		}

		totalPaidNet += paidNet
	}

	newTerbayar := inv.TotalTerbayar + totalPaidNet
	newSisa := newTotalTagihan - newTerbayar
	newStatus := "BELUM_BAYAR"
	if newSisa <= 0 {
		newSisa = 0
		newStatus = "LUNAS"
	}

	_, err = tx.Exec(`
		UPDATE invoices
		SET diskon = $1, total_tagihan = $2, total_terbayar = $3, sisa_tagihan = $4, status = $5,
		    promo_code = CASE WHEN $6 != '' THEN $6 ELSE promo_code END,
		    promo_name = CASE WHEN $7 != '' THEN $7 ELSE promo_name END,
		    updated_at = NOW()
		WHERE id = $8
	`, newDiskon, newTotalTagihan, newTerbayar, newSisa, newStatus, promoCode, promoName, invoiceID)
	if err != nil {
		return nil, err
	}

	if newStatus == "LUNAS" && inv.VisitID != "" {
		_, _ = tx.Exec("UPDATE visits SET status = 'SELESAI', waktu_selesai = NOW(), updated_at = NOW() WHERE id = $1", inv.VisitID)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetInvoiceByID(invoiceID)
}

func (r *BillingRepository) GetPnLReport(startDate, endDate string) (*domain.PnLReportDTO, error) {
	if startDate == "" {
		startDate = time.Now().Format("2006-01") + "-01"
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	dto := &domain.PnLReportDTO{}

	// Total Paid Invoices & Revenue Breakdown by Item Type
	type ItemSum struct {
		ItemType string  `db:"item_type"`
		Total    float64 `db:"total"`
	}
	var sums []ItemSum
	query := `
		SELECT ii.item_type, COALESCE(SUM(ii.subtotal), 0) AS total
		FROM invoice_items ii
		JOIN invoices inv ON ii.invoice_id = inv.id
		WHERE inv.status = 'LUNAS'
		  AND inv.created_at::DATE BETWEEN $1::DATE AND $2::DATE
		GROUP BY ii.item_type
	`
	_ = r.db.Select(&sums, query, startDate, endDate)

	for _, s := range sums {
		switch s.ItemType {
		case "PENDAFTARAN":
			dto.RegistrationIncome += s.Total
		case "KONSULTASI":
			dto.ConsultationIncome += s.Total
		case "TINDAKAN":
			dto.ProcedureIncome += s.Total
		case "OBAT":
			dto.PharmacyIncome += s.Total
		default:
			dto.AncillaryIncome += s.Total
		}
		dto.TotalIncome += s.Total
	}

	// Calculate Estimated Pharmacy HPP (Cost of Goods Sold ~ 70% of drug retail price)
	dto.PharmacyHppCost = dto.PharmacyIncome * 0.70
	dto.GrossProfit = dto.TotalIncome - dto.PharmacyHppCost

	_ = r.db.Get(&dto.TotalInvoicesPaid, `
		SELECT COUNT(*) FROM invoices 
		WHERE status = 'LUNAS' AND created_at::DATE BETWEEN $1::DATE AND $2::DATE
	`, startDate, endDate)

	return dto, nil
}

func (r *BillingRepository) GetTopDiseases(startDate, endDate string, limit int) ([]domain.TopDiseaseDTO, error) {
	if limit <= 0 {
		limit = 10
	}
	if startDate == "" {
		startDate = "2026-01-01"
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	query := `
		SELECT mr.primary_diagnosis_code AS icd10_code,
		       COALESCE(c.description, mr.primary_diagnosis_desc, '') AS description,
		       COUNT(*) AS total_cases
		FROM medical_records mr
		LEFT JOIN icd10_codes c ON mr.primary_diagnosis_code = c.code
		WHERE mr.primary_diagnosis_code IS NOT NULL AND mr.primary_diagnosis_code != ''
		  AND mr.created_at::DATE BETWEEN $1::DATE AND $2::DATE
		GROUP BY mr.primary_diagnosis_code, c.description, mr.primary_diagnosis_desc
		ORDER BY total_cases DESC
		LIMIT $3
	`
	var list []domain.TopDiseaseDTO
	err := r.db.Select(&list, query, startDate, endDate, limit)
	if err != nil {
		return nil, err
	}

	var totalAll int
	for _, it := range list {
		totalAll += it.TotalCases
	}
	if totalAll > 0 {
		for i := range list {
			list[i].Percentage = (float64(list[i].TotalCases) / float64(totalAll)) * 100.0
		}
	}

	return list, nil
}

func (r *BillingRepository) VerifyPublicDocument(docType, code string) (*domain.PublicDocumentVerification, error) {
	ver := &domain.PublicDocumentVerification{
		DocumentType:   docType,
		DocumentNumber: code,
		ClinicName:     "RUMAH SAKIT & KLINIK SIMRS-ASOY",
		Status:         "SAH & TERVERIFIKASI SISTEM",
	}

	switch docType {
	case "INVOICE", "KUITANSI":
		var inv domain.Invoice
		err := r.db.Get(&inv, `
			SELECT inv.nomor_invoice, inv.status, inv.total_tagihan, inv.created_at,
			       p.nama_pasien AS patient_name, p.no_rm
			FROM invoices inv
			JOIN patients p ON inv.patient_id = p.id
			WHERE inv.nomor_invoice = $1 OR inv.id::text = $1
		`, code)
		if err != nil {
			return nil, fmt.Errorf("dokumen kuitansi %s tidak ditemukan atau tidak terdaftar", code)
		}
		ver.DocumentNumber = inv.NomorInvoice
		ver.PatientName = inv.PatientName
		ver.NoRM = inv.NoRM
		ver.IssueDate = inv.CreatedAt.Format("02 Jan 2006 15:04 WIB")
		ver.Details = fmt.Sprintf("Kuitansi Pelayanan Medis dengan Status: %s. Total Tagihan: Rp %.0f", inv.Status, inv.TotalTagihan)

	case "CONSENT", "INFORMED_CONSENT":
		var ic domain.InformedConsent
		err := r.db.Get(&ic, `
			SELECT ic.action_name, ic.consent_type, ic.signer_name, ic.created_at,
			       p.nama_pasien AS patient_name, p.no_rm, u.full_name AS doctor_name
			FROM informed_consents ic
			JOIN patients p ON ic.patient_id = p.id
			JOIN users u ON ic.doctor_id = u.id
			WHERE ic.qr_code_token = $1 OR ic.id::text = $1
		`, code)
		if err != nil {
			return nil, fmt.Errorf("dokumen informed consent dengan token %s tidak valid", code)
		}
		ver.PatientName = ic.PatientName
		ver.NoRM = ic.NoRM
		ver.DoctorName = ic.DoctorName
		ver.IssueDate = ic.CreatedAt.Format("02 Jan 2006 15:04 WIB")
		ver.Details = fmt.Sprintf("Formulir %s untuk Tindakan: %s. Ditandatangani oleh: %s", ic.ConsentType, ic.ActionName, ic.SignerName)

	default: // SURAT_MEDIS
		var mc domain.MedicalCertificate
		err := r.db.Get(&mc, `
			SELECT mc.nomor_surat, mc.jenis_surat, mc.diagnosa, mc.created_at,
			       p.nama_pasien AS patient_name, p.no_rm, u.full_name AS doctor_name
			FROM medical_certificates mc
			JOIN patients p ON mc.patient_id = p.id
			JOIN users u ON mc.doctor_id = u.id
			WHERE mc.nomor_surat = $1 OR mc.id::text = $1
		`, code)
		if err != nil {
			return nil, fmt.Errorf("surat medis %s tidak valid atau tidak terdaftar", code)
		}
		ver.DocumentNumber = mc.NomorSurat
		ver.PatientName = mc.PatientName
		ver.NoRM = mc.NoRM
		ver.DoctorName = mc.DoctorName
		ver.IssueDate = mc.CreatedAt.Format("02 Jan 2006 15:04 WIB")
		ver.Details = fmt.Sprintf("%s dengan Diagnosa Medis: %s", mc.JenisSurat, mc.Diagnosa)
	}

	return ver, nil
}
