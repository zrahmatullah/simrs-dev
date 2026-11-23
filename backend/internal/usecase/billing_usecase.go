package usecase

import (
	"errors"
	"fmt"

	"simrs-asoy/backend/internal/domain"
	"simrs-asoy/backend/internal/repository"
	"simrs-asoy/backend/internal/security"
)

type BillingUsecase struct {
	billingRepo   *repository.BillingRepository
	visitRepo     *repository.VisitRepository
	pharmacyRepo  *repository.PharmacyRepository
	ancillaryRepo *repository.AncillaryRepository
	clinicalRepo  *repository.ClinicalRepository
	auditRepo     *repository.AuditRepository
}

func NewBillingUsecase(
	billingRepo *repository.BillingRepository,
	visitRepo *repository.VisitRepository,
	pharmacyRepo *repository.PharmacyRepository,
	ancillaryRepo *repository.AncillaryRepository,
	clinicalRepo *repository.ClinicalRepository,
	auditRepo *repository.AuditRepository,
) *BillingUsecase {
	return &BillingUsecase{
		billingRepo:   billingRepo,
		visitRepo:     visitRepo,
		pharmacyRepo:  pharmacyRepo,
		ancillaryRepo: ancillaryRepo,
		clinicalRepo:  clinicalRepo,
		auditRepo:     auditRepo,
	}
}

func (u *BillingUsecase) GetInvoices(status, search string, page, limit int) ([]domain.Invoice, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return u.billingRepo.FindInvoices(status, search, page, limit)
}

func (u *BillingUsecase) GetInvoiceByID(id string) (*domain.Invoice, error) {
	return u.billingRepo.GetInvoiceByID(id)
}

func (u *BillingUsecase) GetInvoiceByVisit(visitID string) (*domain.Invoice, error) {
	return u.billingRepo.GetInvoiceByVisitID(visitID)
}

// Auto-aggregate billing items from registration, doctor consultation, procedures, drugs, and lab/rad!
func (u *BillingUsecase) GenerateBillingFromVisit(visitID string, currentUserID, currentUsername string) (*domain.Invoice, error) {
	// Check if invoice already exists
	existing, err := u.billingRepo.GetInvoiceByVisitID(visitID)
	if err == nil && existing != nil {
		return existing, nil
	}

	visit, err := u.visitRepo.FindByID(visitID)
	if err != nil {
		return nil, errors.New("kunjungan tidak ditemukan")
	}

	var items []domain.InvoiceItemDTO

	// 1. Biaya Pendaftaran & Administrasi (Standar Rp 100.000 per aturan regulasi poli)
	items = append(items, domain.InvoiceItemDTO{
		ItemType:    "PENDAFTARAN",
		ReferenceID: visit.ID,
		ItemName:    "Administrasi Pendaftaran & Rekam Medis Poli",
		Qty:         1,
		UnitPrice:   100000.00,
	})

	// 2. Konsultasi Dokter
	items = append(items, domain.InvoiceItemDTO{
		ItemType:    "KONSULTASI",
		ReferenceID: visit.ID,
		ItemName:    "Konsultasi & Pemeriksaan Dokter (" + visit.PoliName + ")",
		Qty:         1,
		UnitPrice:   150000.00,
	})

	// 3. Prosedur dari Rekam Medis ICD-9 jika ada
	mr, _ := u.clinicalRepo.GetMedicalRecordByVisitID(visitID)
	if mr != nil && mr.ICD9Code != nil && *mr.ICD9Code != "" {
		items = append(items, domain.InvoiceItemDTO{
			ItemType:    "TINDAKAN",
			ReferenceID: mr.ID,
			ItemName:    "Prosedur ICD-9: " + mr.ProsedurDeskripsi,
			Qty:         1,
			UnitPrice:   150000.00,
		})
	}

	// 3b. Tindakan Klinis Langsung di Ruangan Poli
	procedures, _ := u.clinicalRepo.GetVisitProcedures(visitID)
	for _, proc := range procedures {
		items = append(items, domain.InvoiceItemDTO{
			ItemType:    "TINDAKAN",
			ReferenceID: proc.ID,
			ItemName:    "Tindakan Poli: " + proc.ProcedureName,
			Qty:         proc.Qty,
			UnitPrice:   proc.Tariff,
		})
	}

	// 4. Obat dari e-Resep jika ada
	presc, _ := u.pharmacyRepo.GetPrescriptionByVisitID(visitID)
	if presc != nil {
		for _, pi := range presc.Items {
			items = append(items, domain.InvoiceItemDTO{
				ItemType:    "OBAT",
				ReferenceID: pi.ID,
				ItemName:    "Obat: " + pi.ItemName + " (" + pi.Signa + ")",
				Qty:         pi.Jumlah,
				UnitPrice:   pi.HargaSatuan,
			})
		}
	}

	// 5. Penunjang (Lab / Rad) jika ada
	orders, _ := u.ancillaryRepo.GetOrdersByVisitID(visitID)
	for _, ord := range orders {
		items = append(items, domain.InvoiceItemDTO{
			ItemType:    ord.JenisOrder,
			ReferenceID: ord.ID,
			ItemName:    "Pemeriksaan Penunjang: " + ord.NamaTindakan,
			Qty:         1,
			UnitPrice:   ord.Tarif,
		})
	}

	var createdBy *string
	if currentUserID != "" {
		createdBy = &currentUserID
	}

	inv := &domain.Invoice{
		VisitID:       visit.ID,
		PatientID:     visit.PatientID,
		JenisPenjamin: visit.GuarantorName,
		NomorPenjamin: visit.NomorAsuransi,
		Diskon:        0,
		Pajak:         0,
		Catatan:       "Tagihan pelayanan rawat jalan " + visit.NoRawat,
		CreatedBy:     createdBy,
	}

	err = u.billingRepo.CreateInvoice(inv, items)
	if err != nil {
		return nil, err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "GENERATE_INVOICE",
		EntityType: "invoice",
		EntityID:   inv.ID,
		Details:    `{"nomor_invoice": "` + inv.NomorInvoice + `", "total": ` + fmt.Sprintf("%.2f", inv.TotalTagihan) + `}`,
	})

	return u.billingRepo.GetInvoiceByID(inv.ID)
}

func (u *BillingUsecase) CreateInvoice(req domain.CreateInvoiceRequest, currentUserID, currentUsername string) (*domain.Invoice, error) {
	// OWASP Arithmetic & Integer Overflow Validation
	if err := security.ValidatePrice(req.Diskon, 100_000_000); err != nil {
		return nil, errors.New("nilai diskon tidak valid atau bernilai negatif")
	}
	if err := security.ValidatePrice(req.Pajak, 100_000_000); err != nil {
		return nil, errors.New("nilai pajak tidak valid atau bernilai negatif")
	}
	for _, item := range req.Items {
		if err := security.ValidateQuantity(item.Qty, 1000); err != nil {
			return nil, err
		}
		if err := security.ValidatePrice(item.UnitPrice, 1_000_000_000); err != nil {
			return nil, err
		}
	}

	visit, err := u.visitRepo.FindByID(req.VisitID)
	if err != nil {
		return nil, errors.New("kunjungan tidak ditemukan")
	}

	var createdBy *string
	if currentUserID != "" {
		createdBy = &currentUserID
	}

	inv := &domain.Invoice{
		VisitID:       req.VisitID,
		PatientID:     visit.PatientID,
		Diskon:        req.Diskon,
		Pajak:         req.Pajak,
		Catatan:       security.SanitizeText(req.Catatan),
		JenisPenjamin: visit.GuarantorName,
		NomorPenjamin: visit.NomorAsuransi,
		CreatedBy:     createdBy,
	}

	err = u.billingRepo.CreateInvoice(inv, req.Items)
	if err != nil {
		return nil, err
	}

	return u.billingRepo.GetInvoiceByID(inv.ID)
}

func (u *BillingUsecase) ProcessPayment(invoiceID string, req domain.ProcessPaymentRequest, currentUserID, currentUsername string) (*domain.Invoice, error) {
	if req.MetodeBayar == "" {
		return nil, errors.New("metode bayar wajib diisi")
	}

	// OWASP: Validate payment amount strictly positive, finite, and within upper bound
	if err := security.ValidatePrice(req.Nominal, 1_000_000_000); err != nil || req.Nominal <= 0 {
		return nil, errors.New("nominal pembayaran harus lebih besar dari 0 dan valid")
	}

	// Check if active shift exists
	shift, err := u.billingRepo.GetActiveShift(currentUserID)
	if err != nil || shift == nil {
		return nil, errors.New("silakan buka shift kasir terlebih dahulu sebelum menerima pembayaran")
	}

	// Rule: Pembayaran TIDAK BISA diproses jika ada orderan resep atau penunjang yang belum selesai!
	inv, err := u.billingRepo.GetInvoiceByID(invoiceID)
	if err != nil || inv == nil {
		return nil, errors.New("invoice tidak ditemukan")
	}

	if inv.VisitID != "" {
		// 1. Cek status resep obat
		presc, _ := u.pharmacyRepo.GetPrescriptionByVisitID(inv.VisitID)
		if presc != nil && (presc.Status == "MENUNGGU_TELAAH" || presc.Status == "PENYIAPAN") {
			return nil, errors.New("pembayaran ditolak: resep obat masih dalam antrean telaah / penyiapan di Farmasi")
		}

		// 2. Cek status order penunjang (Lab, Radiologi, Fisio)
		orders, _ := u.ancillaryRepo.GetOrdersByVisitID(inv.VisitID)
		for _, ord := range orders {
			if ord.Status == "MENUNGGU" || ord.Status == "DALAM_PROSES" {
				return nil, fmt.Errorf("pembayaran ditolak: pemeriksaan penunjang %s (%s) belum selesai divalidasi", ord.JenisOrder, ord.NamaTindakan)
			}
		}
	}

	uangDiterima := req.UangDiterima
	if uangDiterima <= 0 {
		uangDiterima = req.Nominal
	}

	p := &domain.BillingPayment{
		InvoiceID:           invoiceID,
		KasirID:             currentUserID,
		Nominal:             req.Nominal,
		MetodeBayar:         req.MetodeBayar,
		ReferensiPembayaran: security.SanitizePlainText(req.ReferensiPembayaran),
		Catatan:             security.SanitizeText(req.Catatan),
		UangDiterima:        uangDiterima,
		Kembalian:           req.Kembalian,
	}

	err = u.billingRepo.AddPayment(p, req.PromoCode, req.PromoName, req.PromoDiscount)
	if err != nil {
		return nil, err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "PROCESS_PAYMENT",
		EntityType: "billing_payment",
		EntityID:   p.ID,
		Details:    `{"nomor_transaksi": "` + p.NomorTransaksi + `", "metode": "` + req.MetodeBayar + `"}`,
	})

	updatedInv, _ := u.billingRepo.GetInvoiceByID(invoiceID)
	if updatedInv != nil && updatedInv.SisaTagihan <= 0 && inv.VisitID != "" {
		_ = u.visitRepo.UpdateStatus(inv.VisitID, "SELESAI")
	}

	return updatedInv, nil
}

// Cashier Shifts
func (u *BillingUsecase) GetActiveShift(kasirID string) (*domain.CashierShift, error) {
	return u.billingRepo.GetActiveShift(kasirID)
}

func (u *BillingUsecase) OpenShift(req domain.OpenShiftRequest, currentUserID, currentUsername string) (*domain.CashierShift, error) {
	if err := security.ValidatePrice(req.SaldoAwal, 100_000_000); err != nil {
		return nil, errors.New("saldo awal tidak valid atau negatif")
	}

	existing, _ := u.billingRepo.GetActiveShift(currentUserID)
	if existing != nil {
		return existing, nil
	}

	shiftName := security.SanitizePlainText(req.ShiftName)
	if shiftName == "" {
		shiftName = "Shift Pagi"
	}

	s := &domain.CashierShift{
		KasirID:   currentUserID,
		ShiftName: shiftName,
		SaldoAwal: req.SaldoAwal,
	}

	err := u.billingRepo.OpenShift(s)
	if err != nil {
		return nil, err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "OPEN_SHIFT",
		EntityType: "cashier_shift",
		EntityID:   s.ID,
		Details:    `{"shift_name": "` + shiftName + `"}`,
	})

	return u.billingRepo.GetActiveShift(currentUserID)
}

func (u *BillingUsecase) CloseShift(shiftID string, req domain.CloseShiftRequest, currentUserID, currentUsername string) error {
	err := u.billingRepo.CloseShift(shiftID, req.SaldoAkhirFisik, req.CatatanShift)
	if err != nil {
		return err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "CLOSE_SHIFT",
		EntityType: "cashier_shift",
		EntityID:   shiftID,
		Details:    `{"catatan": "` + req.CatatanShift + `"}`,
	})

	return nil
}

func (u *BillingUsecase) GetShiftHistory(kasirID string, limit int) ([]domain.CashierShift, error) {
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return u.billingRepo.GetShiftHistory(kasirID, limit)
}

// -------------------------------------------------------------
// Split Payment & Reports (Fase 1)
// -------------------------------------------------------------

func (u *BillingUsecase) ProcessSplitPayment(invoiceID string, req domain.ProcessSplitPaymentRequest, currentUserID, currentUsername string) (*domain.Invoice, error) {
	if len(req.Payments) == 0 {
		return nil, errors.New("daftar rincian metode pembayaran wajib diisi")
	}

	shift, err := u.billingRepo.GetActiveShift(currentUserID)
	if err != nil || shift == nil {
		return nil, errors.New("silakan buka shift kasir terlebih dahulu sebelum menerima pembayaran")
	}

	inv, err := u.billingRepo.GetInvoiceByID(invoiceID)
	if err != nil || inv == nil {
		return nil, errors.New("invoice tidak ditemukan")
	}

	if inv.VisitID != "" {
		presc, _ := u.pharmacyRepo.GetPrescriptionByVisitID(inv.VisitID)
		if presc != nil && (presc.Status == "MENUNGGU_TELAAH" || presc.Status == "PENYIAPAN") {
			return nil, errors.New("pembayaran ditolak: resep obat masih dalam antrean telaah / penyiapan di Farmasi")
		}

		orders, _ := u.ancillaryRepo.GetOrdersByVisitID(inv.VisitID)
		for _, ord := range orders {
			if ord.Status == "MENUNGGU" || ord.Status == "DALAM_PROSES" {
				return nil, fmt.Errorf("pembayaran ditolak: pemeriksaan penunjang %s (%s) belum selesai divalidasi", ord.JenisOrder, ord.NamaTindakan)
			}
		}
	}

	var totalNominal float64
	for _, p := range req.Payments {
		if p.Nominal <= 0 {
			return nil, errors.New("nominal pembayaran pada masing-masing metode harus lebih besar dari 0")
		}
		totalNominal += p.Nominal
	}

	updatedInv, err := u.billingRepo.AddSplitPayments(invoiceID, req.Payments, req.PromoCode, req.PromoName, req.PromoDiscount, req.UangDiterima, req.Kembalian, currentUserID, currentUsername)
	if err != nil {
		return nil, err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "PROCESS_SPLIT_PAYMENT",
		EntityType: "invoice",
		EntityID:   invoiceID,
		Details:    fmt.Sprintf(`{"total_paid": %.2f, "methods_count": %d}`, totalNominal, len(req.Payments)),
	})

	return updatedInv, nil
}

func (u *BillingUsecase) GetPnLReport(startDate, endDate string) (*domain.PnLReportDTO, error) {
	return u.billingRepo.GetPnLReport(startDate, endDate)
}

func (u *BillingUsecase) GetTopDiseases(startDate, endDate string, limit int) ([]domain.TopDiseaseDTO, error) {
	return u.billingRepo.GetTopDiseases(startDate, endDate, limit)
}

func (u *BillingUsecase) VerifyPublicDocument(docType, code string) (*domain.PublicDocumentVerification, error) {
	return u.billingRepo.VerifyPublicDocument(docType, code)
}
