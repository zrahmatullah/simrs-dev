package domain

// Auth DTOs
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // Durasi dalam detik (misal 900 detik = 15 menit)
	Token        string `json:"token"`      // Alias backward compatibility untuk AccessToken
	User         *User  `json:"user"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	Token        string `json:"token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token,omitempty"`
}

// User DTOs
type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	NIP      string `json:"nip"`
	Phone    string `json:"phone"`
	RoleID   int    `json:"role_id"`
	PoliID   *int   `json:"poli_id"`
}

type UpdateUserRequest struct {
	Email       string `json:"email"`
	FullName    string `json:"full_name"`
	NIP         string `json:"nip"`
	Phone       string `json:"phone"`
	RoleID      int    `json:"role_id"`
	PoliID      *int   `json:"poli_id"`
	StatusAktif bool   `json:"status_aktif"`
	Password    string `json:"password,omitempty"`
}

// Patient DTOs
type CreatePatientRequest struct {
	NIK           string  `json:"nik"`
	NamaPasien    string  `json:"nama_pasien"`
	JenisKelamin  string  `json:"jenis_kelamin"`
	TempatLahir   string  `json:"tempat_lahir"`
	TanggalLahir  string  `json:"tanggal_lahir"` // YYYY-MM-DD
	NamaOrangTua  string  `json:"nama_orang_tua"`
	NoHP          string  `json:"no_hp"`
	Alamat        string  `json:"alamat"`
	RiwayatAlergi *string `json:"riwayat_alergi"`
	GolonganDarah string  `json:"golongan_darah"`
}

type UpdatePatientRequest struct {
	NIK           string  `json:"nik"`
	NamaPasien    string  `json:"nama_pasien"`
	JenisKelamin  string  `json:"jenis_kelamin"`
	TempatLahir   string  `json:"tempat_lahir"`
	TanggalLahir  string  `json:"tanggal_lahir"`
	NamaOrangTua  string  `json:"nama_orang_tua"`
	NoHP          string  `json:"no_hp"`
	Alamat        string  `json:"alamat"`
	RiwayatAlergi *string `json:"riwayat_alergi"`
	GolonganDarah string  `json:"golongan_darah"`
	StatusAktif   bool    `json:"status_aktif"`
}

// Visit DTOs
type CreateVisitRequest struct {
	PatientID     string  `json:"patient_id"`
	PoliID        int     `json:"poli_id"`
	DoctorID      *string `json:"doctor_id"`
	GuarantorID   int     `json:"guarantor_id"`
	NomorAsuransi *string `json:"nomor_asuransi"`
	KeluhanUtama  string  `json:"keluhan_utama"`
}

type UpdateVisitStatusRequest struct {
	Status string `json:"status"` // MENUNGGU_TRIASE, DIPERIKSA_DOKTER, FARMASI, KASIR, SELESAI, BATAL
}

// Appointment DTOs
type CreateAppointmentRequest struct {
	PatientID       string  `json:"patient_id"`
	PoliID          int     `json:"poli_id"`
	DoctorID        *string `json:"doctor_id"`
	AppointmentDate string  `json:"appointment_date"`
	StartTime       string  `json:"start_time"`
	EndTime         string  `json:"end_time"`
	Channel         string  `json:"channel"`
	Catatan         string  `json:"catatan"`
}

type RescheduleAppointmentRequest struct {
	AppointmentDate string `json:"appointment_date"`
	StartTime       string `json:"start_time"`
	EndTime         string `json:"end_time"`
	Catatan         string `json:"catatan"`
}

// Vital Sign DTOs
type CreateVitalSignRequest struct {
	VisitID            string   `json:"visit_id"`
	TDSistole          *int     `json:"td_sistole"`
	TDDiastole         *int     `json:"td_diastole"`
	HR                 *float64 `json:"hr"`
	RR                 *float64 `json:"rr"`
	Temp               *float64 `json:"temp"`
	SpO2               *float64 `json:"spo2"`
	BB                 *float64 `json:"bb"`
	TB                 *float64 `json:"tb"`
	LK                 *float64 `json:"lk"`
	SkalaNyeri         int      `json:"skala_nyeri"`
	Kesadaran          string   `json:"kesadaran"`
	TriaseLevel        string   `json:"triase_level"` // MERAH, KUNING, HIJAU, HITAM
	CatatanKeperawatan string   `json:"catatan_keperawatan"`
}

// Medical Record DTOs
type CreateMedicalRecordRequest struct {
	VisitID            string  `json:"visit_id"`
	Subjektif          string  `json:"subjektif"`
	Objektif           string  `json:"objektif"`
	Asesmen            string  `json:"asesmen"`
	Planning           string  `json:"planning"`
	ICD10Code          *string `json:"icd10_code"`
	ICD9Code           *string `json:"icd9_code"`
	DiagnosisDeskripsi string  `json:"diagnosis_deskripsi"`
	ProsedurDeskripsi  string  `json:"prosedur_deskripsi"`
	Status             string  `json:"status"` // DRAFT, FINAL
}

// Prescription DTOs
type PrescriptionItemDTO struct {
	ItemID      int     `json:"item_id"`
	Signa       string  `json:"signa"`
	AturanPakai string  `json:"aturan_pakai"`
	Jumlah      int     `json:"jumlah"`
	IsRacikan   bool    `json:"is_racikan"`
	RacikanNama string  `json:"racikan_nama"`
}

type CreatePrescriptionRequest struct {
	VisitID       string                `json:"visit_id"`
	CatatanDokter string                `json:"catatan_dokter"`
	Items         []PrescriptionItemDTO `json:"items"`
}

type ReviewPrescriptionRequest struct {
	CatatanTelaah string `json:"catatan_telaah"`
	Status        string `json:"status"` // TELAAH_OK, BATAL
}

// Ancillary Order DTOs
type CreateAncillaryOrderRequest struct {
	VisitID       string  `json:"visit_id"`
	JenisOrder    string  `json:"jenis_order"` // LAB, RAD, FISIOTERAPI, EKG
	NamaTindakan  string  `json:"nama_tindakan"`
	Tarif         float64 `json:"tarif"`
	CatatanKlinis string  `json:"catatan_klinis"`
}

type UpdateAncillaryResultRequest struct {
	HasilPemeriksaan string `json:"hasil_pemeriksaan"`
	LampiranURL      string `json:"lampiran_url"`
	Status           string `json:"status"` // SELESAI
}

// Billing DTOs
type InvoiceItemDTO struct {
	ItemType    string  `json:"item_type"` // PENDAFTARAN, KONSULTASI, TINDAKAN, OBAT, LAB, RAD
	ReferenceID string  `json:"reference_id"`
	ItemName    string  `json:"item_name"`
	Qty         int     `json:"qty"`
	UnitPrice   float64 `json:"unit_price"`
}

type CreateInvoiceRequest struct {
	VisitID       string           `json:"visit_id"`
	Diskon        float64          `json:"diskon"`
	Pajak         float64          `json:"pajak"`
	Catatan       string           `json:"catatan"`
	Items         []InvoiceItemDTO `json:"items"`
}

type ProcessPaymentRequest struct {
	Nominal             float64 `json:"nominal"`
	MetodeBayar         string  `json:"metode_bayar"` // TUNAI, QRIS, DEBIT, KREDIT, TRANSFER
	ReferensiPembayaran string  `json:"referensi_pembayaran"`
	Catatan             string  `json:"catatan"`
	PromoCode           string  `json:"promo_code,omitempty"`
	PromoName           string  `json:"promo_name,omitempty"`
	PromoDiscount       float64 `json:"promo_discount,omitempty"`
	UangDiterima        float64 `json:"uang_diterima,omitempty"`
	Kembalian           float64 `json:"kembalian,omitempty"`
}

type OpenShiftRequest struct {
	ShiftName string  `json:"shift_name"`
	SaldoAwal float64 `json:"saldo_awal"`
}

type CloseShiftRequest struct {
	SaldoAkhirFisik float64 `json:"saldo_akhir_fisik"`
	CatatanShift    string  `json:"catatan_shift"`
}

// Logistics DTOs
type POItemDTO struct {
	ItemID    int     `json:"item_id"`
	Qty       int     `json:"qty"`
	UnitPrice float64 `json:"unit_price"`
}

type CreatePORequest struct {
	SupplierID   *int        `json:"supplier_id"`
	SupplierName string      `json:"supplier_name"`
	OrderDate    string      `json:"order_date"`
	ExpectedDate *string     `json:"expected_date"`
	Catatan      string      `json:"catatan"`
	Items        []POItemDTO `json:"items"`
}

type CreateSupplierRequest struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	PicName  string `json:"pic_name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Address  string `json:"address"`
	City     string `json:"city"`
}

type CreateStockMutationRequest struct {
	ItemID          int     `json:"item_id"`
	TransactionType string  `json:"transaction_type"` // IN_PO, OUT_DISPENSE, MUTASI_DEPO, ADJUSTMENT
	Qty             int     `json:"qty"`
	BatchNumber     string  `json:"batch_number"`
	ExpiryDate      *string `json:"expiry_date"`
	SourceDepo      string  `json:"source_depo"`
	DestDepo        string  `json:"dest_depo"`
	Notes           string  `json:"notes"`
}

// Medical Certificate DTO
type CreateMedicalCertificateRequest struct {
	VisitID    string  `json:"visit_id"`
	JenisSurat string  `json:"jenis_surat"` // SURAT_SAKIT, SURAT_SEHAT, SURAT_RUJUKAN, SURAT_NARKOBA, SURAT_LAIK_TERBANG, SURAT_IZIN
	Diagnosa   string  `json:"diagnosa"`
	Catatan    string  `json:"catatan"`
	TglMulai   *string `json:"tgl_mulai"`
	TglSelesai *string `json:"tgl_selesai"`
}

// Promo DTO
type CreatePromoRequest struct {
	Code           string  `json:"code"`
	Name           string  `json:"name"`
	DiscountType   string  `json:"discount_type"` // PERCENTAGE / FIXED
	DiscountValue  float64 `json:"discount_value"`
	MinTransaction float64 `json:"min_transaction"`
	StartDate      string  `json:"start_date"`
	EndDate        string  `json:"end_date"`
	StatusAktif    bool    `json:"status_aktif"`
}

// Clinical Procedure DTO
type AddVisitProcedureRequest struct {
	VisitID       string  `json:"visit_id"`
	ProcedureName string  `json:"procedure_name"`
	Tariff        float64 `json:"tariff"`
	Qty           int     `json:"qty"`
	Catatan       string  `json:"catatan"`
}

// Nursing Assessment DTO
type CreateNursingAssessmentRequest struct {
	VisitID                 string `json:"visit_id"`
	PatientID               string `json:"patient_id"`
	KeluhanKeperawatan      string `json:"keluhan_keperawatan"`
	RiwayatPenyakitSekarang string `json:"riwayat_penyakit_sekarang"`
	RiwayatAlergi           string `json:"riwayat_alergi"`
	SkalaNyeriSkor          int    `json:"skala_nyeri_skor"`
	SkalaNyeriLokasi        string `json:"skala_nyeri_lokasi"`
	SkalaNyeriDurasi        string `json:"skala_nyeri_durasi"`
	RisikoJatuhTingkat      string `json:"risiko_jatuh_tingkat"`
	SkriningGiziTurunBB     bool   `json:"skrining_gizi_turun_bb"`
	SkriningGiziNafsuMakan  bool   `json:"skrining_gizi_nafsu_makan"`
	MasalahKeperawatan      string `json:"masalah_keperawatan"`
	RencanaKeperawatan      string `json:"rencana_keperawatan"`
	EvaluasiKeperawatan     string `json:"evaluasi_keperawatan"`
}

// Logistics PR & GRN DTOs
type PRItemDTO struct {
	ItemID    int     `json:"item_id"`
	Qty       int     `json:"qty"`
	UnitPrice float64 `json:"unit_price"`
}

type CreatePRRequest struct {
	Department string      `json:"department"`
	Notes      string      `json:"notes"`
	Items      []PRItemDTO `json:"items"`
}

type CreateGRNRequest struct {
	POID                string `json:"po_id"`
	ReceiverName        string `json:"receiver_name"`
	DeliveryOrderNumber string `json:"delivery_order_number"`
	Notes               string `json:"notes"`
}

// Master Faskes, Spesialis, & Poli DTOs
type CreateFaskesRequest struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Tipe   string `json:"tipe"`
	Alamat string `json:"alamat"`
	Kota   string `json:"kota"`
	NoTelp string `json:"no_telp"`
}

type CreateSpesialisRequest struct {
	Code         string `json:"code"`
	NamaDokter   string `json:"nama_dokter"`
	Spesialisasi string `json:"spesialisasi"`
	RumahSakit   string `json:"rumah_sakit"`
}

type CreatePoliRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// -------------------------------------------------------------
// FASE 1 DTOS
// -------------------------------------------------------------

type CreateDoctorScheduleRequest struct {
	DoctorID  string `json:"doctor_id"`
	PoliID    int    `json:"poli_id"`
	DayOfWeek int    `json:"day_of_week"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	MaxQuota  int    `json:"max_quota"`
}

type CreateDoctorLeaveRequest struct {
	DoctorID           string  `json:"doctor_id"`
	StartDate          string  `json:"start_date"`
	EndDate            string  `json:"end_date"`
	Reason             string  `json:"reason"`
	SubstituteDoctorID *string `json:"substitute_doctor_id"`
}

type CreateInformedConsentRequest struct {
	VisitID              string `json:"visit_id"`
	ConsentType          string `json:"consent_type"`
	ActionName           string `json:"action_name"`
	RisksExplained       string `json:"risks_explained"`
	SignerName           string `json:"signer_name"`
	SignerRelation       string `json:"signer_relation"`
	PatientSignatureData string `json:"patient_signature_data"`
	DoctorSignatureData  string `json:"doctor_signature_data"`
}

type MedicineReturnItemDTO struct {
	ItemID    int     `json:"item_id"`
	ItemName  string  `json:"item_name"`
	Qty       int     `json:"qty"`
	UnitPrice float64 `json:"unit_price"`
}

type CreateMedicineReturnRequest struct {
	VisitID        string                  `json:"visit_id"`
	PrescriptionID *string                 `json:"prescription_id"`
	Alasan         string                  `json:"alasan"`
	Items          []MedicineReturnItemDTO `json:"items"`
}

type StockOpnameItemDTO struct {
	ItemID      int    `json:"item_id"`
	ItemName    string `json:"item_name"`
	PhysicalQty int    `json:"physical_qty"`
	Notes       string `json:"notes"`
}

type CreateStockOpnameRequest struct {
	DepoType string               `json:"depo_type"` // GUDANG_UTAMA, DEPO_FARMASI
	Notes    string               `json:"notes"`
	Items    []StockOpnameItemDTO `json:"items"`
}

type CreateMedicineDisposalRequest struct {
	ItemID      int    `json:"item_id"`
	BatchNumber string `json:"batch_number"`
	ExpiryDate  string `json:"expiry_date"`
	Qty         int    `json:"qty"`
	Reason      string `json:"reason"`
	WitnessName string `json:"witness_name"`
}

type SplitPaymentItem struct {
	MetodeBayar         string  `json:"metode_bayar"`
	Nominal             float64 `json:"nominal"`
	ReferensiPembayaran string  `json:"referensi_pembayaran"`
	Catatan             string  `json:"catatan"`
}

type ProcessSplitPaymentRequest struct {
	Payments      []SplitPaymentItem `json:"payments"`
	PromoCode     string             `json:"promo_code,omitempty"`
	PromoName     string             `json:"promo_name,omitempty"`
	PromoDiscount float64            `json:"promo_discount,omitempty"`
	UangDiterima  float64            `json:"uang_diterima,omitempty"`
	Kembalian     float64            `json:"kembalian,omitempty"`
}

type PnLReportDTO struct {
	TotalIncome        float64 `json:"total_income"`
	RegistrationIncome float64 `json:"registration_income"`
	ConsultationIncome float64 `json:"consultation_income"`
	ProcedureIncome    float64 `json:"procedure_income"`
	PharmacyIncome     float64 `json:"pharmacy_income"`
	AncillaryIncome    float64 `json:"ancillary_income"`
	PharmacyHppCost    float64 `json:"pharmacy_hpp_cost"`
	GrossProfit        float64 `json:"gross_profit"`
	TotalInvoicesPaid  int     `json:"total_invoices_paid"`
}

type TopDiseaseDTO struct {
	ICD10Code   string  `json:"icd10_code"`
	Description string  `json:"description"`
	TotalCases  int     `json:"total_cases"`
	Percentage  float64 `json:"percentage"`
}

type PublicDocumentVerification struct {
	DocumentType   string `json:"document_type"`
	DocumentNumber string `json:"document_number"`
	PatientName    string `json:"patient_name"`
	NoRM           string `json:"no_rm"`
	DoctorName     string `json:"doctor_name"`
	IssueDate      string `json:"issue_date"`
	ClinicName     string `json:"clinic_name"`
	Status         string `json:"status"`
	Details        string `json:"details"`
}
