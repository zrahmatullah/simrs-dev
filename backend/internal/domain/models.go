package domain

import (
	"time"
)

// Role Models
type Role struct {
	ID          int       `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// User Models
type User struct {
	ID           string     `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"`
	FullName     string     `json:"full_name" db:"full_name"`
	NIP          string     `json:"nip" db:"nip"`
	Phone        string     `json:"phone" db:"phone"`
	RoleID       int        `json:"role_id" db:"role_id"`
	RoleCode     string     `json:"role_code,omitempty" db:"role_code"`
	RoleName     string     `json:"role_name,omitempty" db:"role_name"`
	PoliID       *int       `json:"poli_id" db:"poli_id"`
	PoliName     string     `json:"poli_name,omitempty" db:"poli_name"`
	StatusAktif  bool       `json:"status_aktif" db:"status_aktif"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Master Poli
type PoliUnit struct {
	ID          int       `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	StatusAktif bool      `json:"status_aktif" db:"status_aktif"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Master Guarantor
type Guarantor struct {
	ID          int       `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	Name        string    `json:"name" db:"name"`
	Category    string    `json:"category" db:"category"`
	StatusAktif bool      `json:"status_aktif" db:"status_aktif"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Master ICD-10
type ICD10 struct {
	ID          int       `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	NameID      string    `json:"name_id" db:"name_id"`
	NameEN      string    `json:"name_en" db:"name_en"`
	StatusAktif bool      `json:"status_aktif" db:"status_aktif"`
}

// Master ICD-9
type ICD9 struct {
	ID          int       `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	NameID      string    `json:"name_id" db:"name_id"`
	Tariff      float64   `json:"tariff" db:"tariff"`
	StatusAktif bool      `json:"status_aktif" db:"status_aktif"`
}

// Master Supplier / PBF (Pedagang Besar Farmasi)
type Supplier struct {
	ID          int       `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	Name        string    `json:"name" db:"name"`
	PicName     string    `json:"pic_name" db:"pic_name"`
	Phone       string    `json:"phone" db:"phone"`
	Email       string    `json:"email" db:"email"`
	Address     string    `json:"address" db:"address"`
	City        string    `json:"city" db:"city"`
	StatusAktif bool      `json:"status_aktif" db:"status_aktif"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Master Rumah Sakit / Faskes Rujukan Eksternal
type FaskesRujukan struct {
	ID          int       `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	Name        string    `json:"name" db:"name"`
	Tipe        string    `json:"tipe" db:"tipe"` // RSUP, RSUD, RS Swasta, Klinik Utama, Puskesmas
	Alamat      string    `json:"alamat" db:"alamat"`
	Kota        string    `json:"kota" db:"kota"`
	NoTelp      string    `json:"no_telp" db:"no_telp"`
	StatusAktif bool      `json:"status_aktif" db:"status_aktif"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Master Dokter Spesialis & Subspesialis Rujukan
type MasterSpesialis struct {
	ID           int       `json:"id" db:"id"`
	Code         string    `json:"code" db:"code"`
	NamaDokter   string    `json:"nama_dokter" db:"nama_dokter"`
	Spesialisasi string    `json:"spesialisasi" db:"spesialisasi"`
	RumahSakit   string    `json:"rumah_sakit" db:"rumah_sakit"`
	StatusAktif  bool      `json:"status_aktif" db:"status_aktif"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Patient Model
type Patient struct {
	ID            string     `json:"id" db:"id"`
	NoRM          string     `json:"no_rm" db:"no_rm"`
	NIK           string     `json:"nik" db:"nik"`
	NamaPasien    string     `json:"nama_pasien" db:"nama_pasien"`
	JenisKelamin  string     `json:"jenis_kelamin" db:"jenis_kelamin"`
	TempatLahir   string     `json:"tempat_lahir" db:"tempat_lahir"`
	TanggalLahir  string     `json:"tanggal_lahir" db:"tanggal_lahir"`
	NamaOrangTua  string     `json:"nama_orang_tua" db:"nama_orang_tua"`
	NoHP          string     `json:"no_hp" db:"no_hp"`
	Alamat        string     `json:"alamat" db:"alamat"`
	RiwayatAlergi *string    `json:"riwayat_alergi" db:"riwayat_alergi"`
	GolonganDarah string     `json:"golongan_darah" db:"golongan_darah"`
	StatusAktif   bool       `json:"status_aktif" db:"status_aktif"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
	CreatedBy     *string    `json:"created_by" db:"created_by"`
	UpdatedBy     *string    `json:"updated_by" db:"updated_by"`
}

// Visit / Patient Registration Model
type Visit struct {
	ID             string     `json:"id" db:"id"`
	NoRawat        string     `json:"no_rawat" db:"no_rawat"`
	NoAntrian      string     `json:"no_antrian" db:"no_antrian"`
	PatientID      string     `json:"patient_id" db:"patient_id"`
	PoliID         int        `json:"poli_id" db:"poli_id"`
	DoctorID       *string    `json:"doctor_id" db:"doctor_id"`
	GuarantorID    int        `json:"guarantor_id" db:"guarantor_id"`
	NomorAsuransi  *string    `json:"nomor_asuransi" db:"nomor_asuransi"`
	VisitDate      string     `json:"visit_date" db:"visit_date"`
	KeluhanUtama   string     `json:"keluhan_utama" db:"keluhan_utama"`
	Status         string     `json:"status" db:"status"` // MENUNGGU_TRIASE, DIPERIKSA_DOKTER, FARMASI, KASIR, SELESAI, BATAL
	WaktuMasuk     time.Time  `json:"waktu_masuk" db:"waktu_masuk"`
	WaktuSelesai   *time.Time `json:"waktu_selesai" db:"waktu_selesai"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
	CreatedBy      *string    `json:"created_by" db:"created_by"`

	// Joined details
	PatientName    string     `json:"patient_name,omitempty" db:"patient_name"`
	NoRM           string     `json:"no_rm,omitempty" db:"no_rm"`
	PatientGender  string     `json:"patient_gender,omitempty" db:"patient_gender"`
	PatientDOB     string     `json:"patient_dob,omitempty" db:"patient_dob"`
	PatientAllergy *string    `json:"patient_allergy,omitempty" db:"patient_allergy"`
	PoliName       string     `json:"poli_name,omitempty" db:"poli_name"`
	DoctorName     string     `json:"doctor_name,omitempty" db:"doctor_name"`
	GuarantorName  string     `json:"guarantor_name,omitempty" db:"guarantor_name"`
}

// Appointment Model
type Appointment struct {
	ID              string     `json:"id" db:"id"`
	PatientID       string     `json:"patient_id" db:"patient_id"`
	PoliID          int        `json:"poli_id" db:"poli_id"`
	DoctorID        *string    `json:"doctor_id" db:"doctor_id"`
	AppointmentDate string     `json:"appointment_date" db:"appointment_date"`
	StartTime       string     `json:"start_time" db:"start_time"`
	EndTime         string     `json:"end_time" db:"end_time"`
	Status          string     `json:"status" db:"status"` // TERJADWAL, HADIR, RESCHEDULE, BATAL
	Channel         string     `json:"channel" db:"channel"`
	Catatan         string     `json:"catatan" db:"catatan"`
	CheckedInAt     *time.Time `json:"checked_in_at" db:"checked_in_at"`
	CheckedInBy     *string    `json:"checked_in_by" db:"checked_in_by"`
	VisitID         *string    `json:"visit_id" db:"visit_id"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`

	PatientName     string     `json:"patient_name,omitempty" db:"patient_name"`
	NoRM            string     `json:"no_rm,omitempty" db:"no_rm"`
	NoHP            string     `json:"no_hp,omitempty" db:"no_hp"`
	PoliName        string     `json:"poli_name,omitempty" db:"poli_name"`
	DoctorName      string     `json:"doctor_name,omitempty" db:"doctor_name"`
}

// Vital Sign / TTV Model
type VitalSign struct {
	ID                 string     `json:"id" db:"id"`
	VisitID            string     `json:"visit_id" db:"visit_id"`
	PatientID          string     `json:"patient_id" db:"patient_id"`
	NurseID            *string    `json:"nurse_id" db:"nurse_id"`
	NurseName          string     `json:"nurse_name,omitempty" db:"nurse_name"`
	TDSistole          *int       `json:"td_sistole" db:"td_sistole"`
	TDDiastole         *int       `json:"td_diastole" db:"td_diastole"`
	HR                 *float64   `json:"hr" db:"hr"`
	RR                 *float64   `json:"rr" db:"rr"`
	Temp               *float64   `json:"temp" db:"temp"`
	SpO2               *float64   `json:"spo2" db:"spo2"`
	BB                 *float64   `json:"bb" db:"bb"`
	TB                 *float64   `json:"tb" db:"tb"`
	LK                 *float64   `json:"lk" db:"lk"`
	SkalaNyeri         int        `json:"skala_nyeri" db:"skala_nyeri"`
	Kesadaran          string     `json:"kesadaran" db:"kesadaran"`
	TriaseLevel        string     `json:"triase_level" db:"triase_level"` // MERAH, KUNING, HIJAU, HITAM
	CatatanKeperawatan string     `json:"catatan_keperawatan" db:"catatan_keperawatan"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
}

// Medical Record / EMR Model
type MedicalRecord struct {
	ID                 string     `json:"id" db:"id"`
	VisitID            string     `json:"visit_id" db:"visit_id"`
	PatientID          string     `json:"patient_id" db:"patient_id"`
	DoctorID           string     `json:"doctor_id" db:"doctor_id"`
	DoctorName         string     `json:"doctor_name,omitempty" db:"doctor_name"`
	Subjektif          string     `json:"subjektif" db:"subjektif"`
	Objektif           string     `json:"objektif" db:"objektif"`
	Asesmen            string     `json:"asesmen" db:"asesmen"`
	Planning           string     `json:"planning" db:"planning"`
	ICD10Code          *string    `json:"icd10_code" db:"icd10_code"`
	ICD9Code           *string    `json:"icd9_code" db:"icd9_code"`
	DiagnosisDeskripsi string     `json:"diagnosis_deskripsi" db:"diagnosis_deskripsi"`
	ProsedurDeskripsi  string     `json:"prosedur_deskripsi" db:"prosedur_deskripsi"`
	Status             string     `json:"status" db:"status"` // DRAFT, FINAL
	IsLocked           bool       `json:"is_locked" db:"is_locked"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

// Prescription & Items
type Prescription struct {
	ID            string              `json:"id" db:"id"`
	NomorResep    string              `json:"nomor_resep" db:"nomor_resep"`
	VisitID       string              `json:"visit_id" db:"visit_id"`
	PatientID     string              `json:"patient_id" db:"patient_id"`
	DoctorID      string              `json:"doctor_id" db:"doctor_id"`
	DoctorName    string              `json:"doctor_name,omitempty" db:"doctor_name"`
	PatientName   string              `json:"patient_name,omitempty" db:"patient_name"`
	NoRM          string              `json:"no_rm,omitempty" db:"no_rm"`
	Status        string              `json:"status" db:"status"` // MENUNGGU_TELAAH, TELAAH_OK, PENYIAPAN, SELESAI, BATAL
	CatatanDokter string              `json:"catatan_dokter" db:"catatan_dokter"`
	CatatanTelaah string              `json:"catatan_telaah" db:"catatan_telaah"`
	IsVerified    bool                `json:"is_verified" db:"is_verified"`
	VerifiedBy    *string             `json:"verified_by" db:"verified_by"`
	VerifiedAt    *time.Time          `json:"verified_at" db:"verified_at"`
	DispensedBy   *string             `json:"dispensed_by" db:"dispensed_by"`
	DispensedAt   *time.Time          `json:"dispensed_at" db:"dispensed_at"`
	CreatedAt     time.Time           `json:"created_at" db:"created_at"`
	Items         []PrescriptionItem  `json:"items,omitempty"`
}

type PrescriptionItem struct {
	ID             string    `json:"id" db:"id"`
	PrescriptionID string    `json:"prescription_id" db:"prescription_id"`
	ItemID         int       `json:"item_id" db:"item_id"`
	ItemName       string    `json:"item_name,omitempty" db:"item_name"`
	Signa          string    `json:"signa" db:"signa"`
	AturanPakai    string    `json:"aturan_pakai" db:"aturan_pakai"`
	Jumlah         int       `json:"jumlah" db:"jumlah"`
	HargaSatuan    float64   `json:"harga_satuan" db:"harga_satuan"`
	Subtotal       float64   `json:"subtotal" db:"subtotal"`
	IsRacikan      bool      `json:"is_racikan" db:"is_racikan"`
	RacikanNama    string    `json:"racikan_nama" db:"racikan_nama"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// Ancillary Order (Lab / Rad / Fisio)
type AncillaryOrder struct {
	ID               string     `json:"id" db:"id"`
	NomorOrder       string     `json:"nomor_order" db:"nomor_order"`
	VisitID          string     `json:"visit_id" db:"visit_id"`
	PatientID        string     `json:"patient_id" db:"patient_id"`
	PatientName      string     `json:"patient_name,omitempty" db:"patient_name"`
	NoRM             string     `json:"no_rm,omitempty" db:"no_rm"`
	DoctorID         string     `json:"doctor_id" db:"doctor_id"`
	DoctorName       string     `json:"doctor_name,omitempty" db:"doctor_name"`
	JenisOrder       string     `json:"jenis_order" db:"jenis_order"` // LAB, RAD, FISIOTERAPI, EKG
	NamaTindakan     string     `json:"nama_tindakan" db:"nama_tindakan"`
	Tarif            float64    `json:"tarif" db:"tarif"`
	CatatanKlinis    string     `json:"catatan_klinis" db:"catatan_klinis"`
	HasilPemeriksaan string     `json:"hasil_pemeriksaan" db:"hasil_pemeriksaan"`
	LampiranURL      string     `json:"lampiran_url" db:"lampiran_url"`
	Status           string     `json:"status" db:"status"` // MENUNGGU, DALAM_PROSES, SELESAI, BATAL
	VerifiedBy       *string    `json:"verified_by" db:"verified_by"`
	VerifiedAt       *time.Time `json:"verified_at" db:"verified_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

// Invoice & Billing
type Invoice struct {
	ID            string        `json:"id" db:"id"`
	NomorInvoice  string        `json:"nomor_invoice" db:"nomor_invoice"`
	VisitID       string        `json:"visit_id" db:"visit_id"`
	NoRawat       string        `json:"no_rawat,omitempty" db:"no_rawat"`
	PatientID     string        `json:"patient_id" db:"patient_id"`
	PatientName   string        `json:"patient_name,omitempty" db:"patient_name"`
	NoRM          string        `json:"no_rm,omitempty" db:"no_rm"`
	Subtotal      float64       `json:"subtotal" db:"subtotal"`
	Diskon        float64       `json:"diskon" db:"diskon"`
	Pajak         float64       `json:"pajak" db:"pajak"`
	TotalTagihan  float64       `json:"total_tagihan" db:"total_tagihan"`
	TotalTerbayar float64       `json:"total_terbayar" db:"total_terbayar"`
	SisaTagihan   float64       `json:"sisa_tagihan" db:"sisa_tagihan"`
	Status        string        `json:"status" db:"status"` // BELUM_BAYAR, LUNAS, BATAL
	JenisPenjamin string        `json:"jenis_penjamin" db:"jenis_penjamin"`
	NomorPenjamin *string          `json:"nomor_penjamin" db:"nomor_penjamin"`
	Catatan       string           `json:"catatan" db:"catatan"`
	PromoCode     string           `json:"promo_code,omitempty" db:"promo_code"`
	PromoName     string           `json:"promo_name,omitempty" db:"promo_name"`
	CreatedBy     *string          `json:"created_by,omitempty" db:"created_by"`
	CreatedAt     time.Time        `json:"created_at" db:"created_at"`
	Items         []InvoiceItem    `json:"items,omitempty"`
	Payments      []BillingPayment `json:"payments,omitempty"`
}

type InvoiceItem struct {
	ID          string    `json:"id" db:"id"`
	InvoiceID   string    `json:"invoice_id" db:"invoice_id"`
	ItemType    string    `json:"item_type" db:"item_type"` // PENDAFTARAN, KONSULTASI, TINDAKAN, OBAT, LAB, RAD
	ReferenceID string    `json:"reference_id" db:"reference_id"`
	ItemName    string    `json:"item_name" db:"item_name"`
	Qty         int       `json:"qty" db:"qty"`
	UnitPrice   float64   `json:"unit_price" db:"unit_price"`
	Subtotal    float64   `json:"subtotal" db:"subtotal"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type BillingPayment struct {
	ID                  string    `json:"id" db:"id"`
	NomorTransaksi      string    `json:"nomor_transaksi" db:"nomor_transaksi"`
	InvoiceID           string    `json:"invoice_id" db:"invoice_id"`
	KasirID             string    `json:"kasir_id" db:"kasir_id"`
	KasirName           string    `json:"kasir_name,omitempty" db:"kasir_name"`
	Nominal             float64   `json:"nominal" db:"nominal"`
	MetodeBayar         string    `json:"metode_bayar" db:"metode_bayar"` // TUNAI, QRIS, DEBIT, KREDIT, TRANSFER
	ReferensiPembayaran string    `json:"referensi_pembayaran" db:"referensi_pembayaran"`
	Catatan             string    `json:"catatan" db:"catatan"`
	UangDiterima        float64   `json:"uang_diterima" db:"uang_diterima"`
	Kembalian           float64   `json:"kembalian" db:"kembalian"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
}

type CashierShift struct {
	ID                  string     `json:"id" db:"id"`
	KasirID             string     `json:"kasir_id" db:"kasir_id"`
	KasirName           string     `json:"kasir_name,omitempty" db:"kasir_name"`
	ShiftName           string     `json:"shift_name" db:"shift_name"`
	WaktuBuka           time.Time  `json:"waktu_buka" db:"waktu_buka"`
	WaktuTutup          *time.Time `json:"waktu_tutup" db:"waktu_tutup"`
	SaldoAwal           float64    `json:"saldo_awal" db:"saldo_awal"`
	TotalTunaiSistem    float64    `json:"total_tunai_sistem" db:"total_tunai_sistem"`
	TotalNonTunaiSistem float64    `json:"total_non_tunai_sistem" db:"total_non_tunai_sistem"`
	SaldoAkhirFisik     *float64   `json:"saldo_akhir_fisik" db:"saldo_akhir_fisik"`
	Selisih             *float64   `json:"selisih" db:"selisih"`
	Status              string     `json:"status" db:"status"` // OPEN, CLOSED
	CatatanShift        string     `json:"catatan_shift" db:"catatan_shift"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
}

// Inventory & Logistics
type InventoryCategory struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
}

type InventoryItem struct {
	ID           int       `json:"id" db:"id"`
	CategoryID   int       `json:"category_id" db:"category_id"`
	CategoryName string    `json:"category_name,omitempty" db:"category_name"`
	ItemCode     string    `json:"item_code" db:"item_code"`
	ItemName     string    `json:"item_name" db:"item_name"`
	GenericName  string    `json:"generic_name" db:"generic_name"`
	Unit         string    `json:"unit" db:"unit"`
	BuyPrice     float64   `json:"buy_price" db:"buy_price"`
	SellPrice    float64   `json:"sell_price" db:"sell_price"`
	MinimumStock int       `json:"minimum_stock" db:"minimum_stock"`
	TotalStock   int       `json:"total_stock" db:"total_stock"`
	DepoStock    int       `json:"depo_stock" db:"depo_stock"`
	IsMedicine   bool      `json:"is_medicine" db:"is_medicine"`
	RackLocation string    `json:"rack_location" db:"rack_location"`
	StatusAktif  bool      `json:"status_aktif" db:"status_aktif"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type PurchaseOrder struct {
	ID           string              `json:"id" db:"id"`
	PONumber     string              `json:"po_number" db:"po_number"`
	SupplierID   *int                `json:"supplier_id,omitempty" db:"supplier_id"`
	SupplierName string              `json:"supplier_name" db:"supplier_name"`
	OrderDate    string              `json:"order_date" db:"order_date"`
	ExpectedDate *string             `json:"expected_date" db:"expected_date"`
	Status       string              `json:"status" db:"status"` // DRAFT, SUBMITTED, RECEIVED, CANCELLED
	TotalAmount  float64             `json:"total_amount" db:"total_amount"`
	Catatan      string              `json:"catatan" db:"catatan"`
	CreatedBy    *string             `json:"created_by" db:"created_by"`
	CreatedAt    time.Time           `json:"created_at" db:"created_at"`
	Items        []PurchaseOrderItem `json:"items,omitempty"`
}

type PurchaseOrderItem struct {
	ID        string    `json:"id" db:"id"`
	POID      string    `json:"po_id" db:"po_id"`
	ItemID    int       `json:"item_id" db:"item_id"`
	ItemName  string    `json:"item_name,omitempty" db:"item_name"`
	Unit      string    `json:"unit,omitempty" db:"unit"`
	Qty       int       `json:"qty" db:"qty"`
	UnitPrice float64   `json:"unit_price" db:"unit_price"`
	Subtotal  float64   `json:"subtotal" db:"subtotal"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type StockMutation struct {
	ID              string    `json:"id" db:"id"`
	ItemID          int       `json:"item_id" db:"item_id"`
	ItemName        string    `json:"item_name,omitempty" db:"item_name"`
	TransactionType string    `json:"transaction_type" db:"transaction_type"` // IN_PO, OUT_DISPENSE, MUTASI_DEPO, ADJUSTMENT
	Qty             int       `json:"qty" db:"qty"`
	BatchNumber     string    `json:"batch_number" db:"batch_number"`
	ExpiryDate      *string   `json:"expiry_date" db:"expiry_date"`
	SourceDepo      string    `json:"source_depo" db:"source_depo"`
	DestDepo        string    `json:"dest_depo" db:"dest_depo"`
	Notes           string    `json:"notes" db:"notes"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type MedicalCertificate struct {
	ID          string    `json:"id" db:"id"`
	NomorSurat  string    `json:"nomor_surat" db:"nomor_surat"`
	VisitID     string    `json:"visit_id" db:"visit_id"`
	PatientID   string    `json:"patient_id" db:"patient_id"`
	DoctorID    string    `json:"doctor_id" db:"doctor_id"`
	JenisSurat  string    `json:"jenis_surat" db:"jenis_surat"` // SURAT_SAKIT, SURAT_SEHAT, SURAT_RUJUKAN
	Diagnosa    string    `json:"diagnosa" db:"diagnosa"`
	Catatan     string    `json:"catatan" db:"catatan"`
	TglMulai    *string   `json:"tgl_mulai" db:"tgl_mulai"`
	TglSelesai  *string   `json:"tgl_selesai" db:"tgl_selesai"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`

	// Joined details
	PatientName string    `json:"patient_name,omitempty" db:"patient_name"`
	NoRM        string    `json:"no_rm,omitempty" db:"no_rm"`
	DoctorName  string    `json:"doctor_name,omitempty" db:"doctor_name"`
}

type AuditLog struct {
	ID         int64     `json:"id" db:"id"`
	UserID     *string   `json:"user_id" db:"user_id"`
	Username   string    `json:"username" db:"username"`
	Action     string    `json:"action" db:"action"`
	EntityType string    `json:"entity_type" db:"entity_type"`
	EntityID   string    `json:"entity_id" db:"entity_id"`
	Details    string    `json:"details" db:"details"`
	IPAddress  string    `json:"ip_address" db:"ip_address"`
	UserAgent  string    `json:"user_agent" db:"user_agent"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// Master Promo
type Promo struct {
	ID             int       `json:"id" db:"id"`
	Code           string    `json:"code" db:"code"`
	Name           string    `json:"name" db:"name"`
	DiscountType   string    `json:"discount_type" db:"discount_type"` // PERCENTAGE / FIXED
	DiscountValue  float64   `json:"discount_value" db:"discount_value"`
	MinTransaction float64   `json:"min_transaction" db:"min_transaction"`
	StartDate      string    `json:"start_date" db:"start_date"`
	EndDate        string    `json:"end_date" db:"end_date"`
	StatusAktif    bool      `json:"status_aktif" db:"status_aktif"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Master & Visit Clinical Procedures (Tindakan di Ruangan Poli)
type MasterProcedure struct {
	ID          int       `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	Name        string    `json:"name" db:"name"`
	Category    string    `json:"category" db:"category"`
	Tariff      float64   `json:"tariff" db:"tariff"`
	StatusAktif bool      `json:"status_aktif" db:"status_aktif"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type VisitProcedure struct {
	ID            string    `json:"id" db:"id"`
	VisitID       string    `json:"visit_id" db:"visit_id"`
	ProcedureName string    `json:"procedure_name" db:"procedure_name"`
	Tariff        float64   `json:"tariff" db:"tariff"`
	Qty           int       `json:"qty" db:"qty"`
	Subtotal      float64   `json:"subtotal" db:"subtotal"`
	DoctorID      *string   `json:"doctor_id" db:"doctor_id"`
	Catatan       string    `json:"catatan" db:"catatan"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// Detailed Nursing Assessment
type NursingAssessment struct {
	ID                       string    `json:"id" db:"id"`
	VisitID                  string    `json:"visit_id" db:"visit_id"`
	PatientID                string    `json:"patient_id" db:"patient_id"`
	NurseID                  *string   `json:"nurse_id" db:"nurse_id"`
	KeluhanKeperawatan       string    `json:"keluhan_keperawatan" db:"keluhan_keperawatan"`
	RiwayatPenyakitSekarang  string    `json:"riwayat_penyakit_sekarang" db:"riwayat_penyakit_sekarang"`
	RiwayatAlergi            string    `json:"riwayat_alergi" db:"riwayat_alergi"`
	SkalaNyeriSkor           int       `json:"skala_nyeri_skor" db:"skala_nyeri_skor"`
	SkalaNyeriLokasi         string    `json:"skala_nyeri_lokasi" db:"skala_nyeri_lokasi"`
	SkalaNyeriDurasi         string    `json:"skala_nyeri_durasi" db:"skala_nyeri_durasi"`
	RisikoJatuhTingkat       string    `json:"risiko_jatuh_tingkat" db:"risiko_jatuh_tingkat"`
	SkriningGiziTurunBB      bool      `json:"skrining_gizi_turun_bb" db:"skrining_gizi_turun_bb"`
	SkriningGiziNafsuMakan   bool      `json:"skrining_gizi_nafsu_makan" db:"skrining_gizi_nafsu_makan"`
	MasalahKeperawatan       string    `json:"masalah_keperawatan" db:"masalah_keperawatan"`
	RencanaKeperawatan       string    `json:"rencana_keperawatan" db:"rencana_keperawatan"`
	EvaluasiKeperawatan      string    `json:"evaluasi_keperawatan" db:"evaluasi_keperawatan"`
	CreatedAt                time.Time `json:"created_at" db:"created_at"`
	UpdatedAt                time.Time `json:"updated_at" db:"updated_at"`
}

// Logistics PR (Purchase Requisition) & GRN (Goods Receipt Note)
type PurchaseRequisitionItem struct {
	ID        string    `json:"id" db:"id"`
	PRID      string    `json:"pr_id" db:"pr_id"`
	ItemID    int       `json:"item_id" db:"item_id"`
	ItemName  string    `json:"item_name,omitempty" db:"item_name"`
	Qty       int       `json:"qty" db:"qty"`
	UnitPrice float64   `json:"unit_price" db:"unit_price"`
	Subtotal  float64   `json:"subtotal" db:"subtotal"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type PurchaseRequisition struct {
	ID             string                    `json:"id" db:"id"`
	PRNumber       string                    `json:"pr_number" db:"pr_number"`
	RequestedBy    *string                   `json:"requested_by" db:"requested_by"`
	RequesterName  string                    `json:"requester_name,omitempty" db:"requester_name"`
	RequestDate    string                    `json:"request_date" db:"request_date"`
	Department     string                    `json:"department" db:"department"`
	Status         string                    `json:"status" db:"status"` // SUBMITTED, APPROVED, REJECTED, CONVERTED_TO_PO
	TotalEstimated float64                   `json:"total_estimated" db:"total_estimated"`
	Notes          string                    `json:"notes" db:"notes"`
	CreatedAt      time.Time                 `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time                 `json:"updated_at" db:"updated_at"`
	Items          []PurchaseRequisitionItem `json:"items,omitempty"`
}

type GoodsReceiptNote struct {
	ID                  string    `json:"id" db:"id"`
	GRNNumber           string    `json:"grn_number" db:"grn_number"`
	POID                string    `json:"po_id" db:"po_id"`
	PONumber            string    `json:"po_number,omitempty" db:"po_number"`
	SupplierName        string    `json:"supplier_name,omitempty" db:"supplier_name"`
	ReceivedDate        string    `json:"received_date" db:"received_date"`
	ReceiverName        string    `json:"receiver_name" db:"receiver_name"`
	DeliveryOrderNumber string    `json:"delivery_order_number" db:"delivery_order_number"`
	Status              string    `json:"status" db:"status"`
	Notes               string    `json:"notes" db:"notes"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
}

// -------------------------------------------------------------
// FASE 1 MODELS
// -------------------------------------------------------------

// Doctor Schedule & Leave
type DoctorSchedule struct {
	ID          int       `json:"id" db:"id"`
	DoctorID    string    `json:"doctor_id" db:"doctor_id"`
	DoctorName  string    `json:"doctor_name,omitempty" db:"doctor_name"`
	PoliID      int       `json:"poli_id" db:"poli_id"`
	PoliName    string    `json:"poli_name,omitempty" db:"poli_name"`
	DayOfWeek   int       `json:"day_of_week" db:"day_of_week"` // 0=Minggu, 1=Senin..
	StartTime   string    `json:"start_time" db:"start_time"`
	EndTime     string    `json:"end_time" db:"end_time"`
	MaxQuota    int       `json:"max_quota" db:"max_quota"`
	StatusAktif bool      `json:"status_aktif" db:"status_aktif"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type DoctorLeave struct {
	ID                   string    `json:"id" db:"id"`
	DoctorID             string    `json:"doctor_id" db:"doctor_id"`
	DoctorName           string    `json:"doctor_name,omitempty" db:"doctor_name"`
	StartDate            string    `json:"start_date" db:"start_date"`
	EndDate              string    `json:"end_date" db:"end_date"`
	Reason               string    `json:"reason" db:"reason"`
	SubstituteDoctorID   *string   `json:"substitute_doctor_id" db:"substitute_doctor_id"`
	SubstituteDoctorName *string   `json:"substitute_doctor_name,omitempty" db:"substitute_doctor_name"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
}

// Informed Consent Digital
type InformedConsent struct {
	ID                   string    `json:"id" db:"id"`
	VisitID              string    `json:"visit_id" db:"visit_id"`
	NoRawat              string    `json:"no_rawat,omitempty" db:"no_rawat"`
	PatientID            string    `json:"patient_id" db:"patient_id"`
	PatientName          string    `json:"patient_name,omitempty" db:"patient_name"`
	NoRM                 string    `json:"no_rm,omitempty" db:"no_rm"`
	DoctorID             string    `json:"doctor_id" db:"doctor_id"`
	DoctorName           string    `json:"doctor_name,omitempty" db:"doctor_name"`
	ConsentType          string    `json:"consent_type" db:"consent_type"`
	ActionName           string    `json:"action_name" db:"action_name"`
	RisksExplained       string    `json:"risks_explained" db:"risks_explained"`
	SignerName           string    `json:"signer_name" db:"signer_name"`
	SignerRelation       string    `json:"signer_relation" db:"signer_relation"`
	PatientSignatureData string    `json:"patient_signature_data" db:"patient_signature_data"`
	DoctorSignatureData  string    `json:"doctor_signature_data" db:"doctor_signature_data"`
	QRCodeToken          string    `json:"qr_code_token" db:"qr_code_token"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
}

// Medicine Returns
type MedicineReturnItem struct {
	ID        string    `json:"id" db:"id"`
	ReturnID  string    `json:"return_id" db:"return_id"`
	ItemID    int       `json:"item_id" db:"item_id"`
	ItemName  string    `json:"item_name" db:"item_name"`
	Qty       int       `json:"qty" db:"qty"`
	UnitPrice float64   `json:"unit_price" db:"unit_price"`
	Subtotal  float64   `json:"subtotal" db:"subtotal"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type MedicineReturn struct {
	ID             string               `json:"id" db:"id"`
	NomorRetur     string               `json:"nomor_retur" db:"nomor_retur"`
	PrescriptionID *string              `json:"prescription_id" db:"prescription_id"`
	VisitID        string               `json:"visit_id" db:"visit_id"`
	NoRawat        string               `json:"no_rawat,omitempty" db:"no_rawat"`
	PatientID      string               `json:"patient_id" db:"patient_id"`
	PatientName    string               `json:"patient_name,omitempty" db:"patient_name"`
	InvoiceID      *string              `json:"invoice_id" db:"invoice_id"`
	Alasan         string               `json:"alasan" db:"alasan"`
	TotalRefund    float64              `json:"total_refund" db:"total_refund"`
	Status         string               `json:"status" db:"status"`
	CreatedBy      *string              `json:"created_by" db:"created_by"`
	CreatedAt      time.Time            `json:"created_at" db:"created_at"`
	Items          []MedicineReturnItem `json:"items,omitempty"`
}

// Stock Opname
type StockOpnameItem struct {
	ID                   string    `json:"id" db:"id"`
	OpnameID             string    `json:"opname_id" db:"opname_id"`
	ItemID               int       `json:"item_id" db:"item_id"`
	ItemName             string    `json:"item_name" db:"item_name"`
	SystemQty            int       `json:"system_qty" db:"system_qty"`
	PhysicalQty          int       `json:"physical_qty" db:"physical_qty"`
	DifferenceQty        int       `json:"difference_qty" db:"difference_qty"`
	HppPrice             float64   `json:"hpp_price" db:"hpp_price"`
	TotalDifferenceValue float64   `json:"total_difference_value" db:"total_difference_value"`
	Notes                string    `json:"notes" db:"notes"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
}

type StockOpname struct {
	ID                string            `json:"id" db:"id"`
	OpnameNumber      string            `json:"opname_number" db:"opname_number"`
	OpnameDate        string            `json:"opname_date" db:"opname_date"`
	DepoType          string            `json:"depo_type" db:"depo_type"`
	Status            string            `json:"status" db:"status"`
	TotalItems        int               `json:"total_items" db:"total_items"`
	TotalDeviasiNilai float64           `json:"total_deviasi_nilai" db:"total_deviasi_nilai"`
	Notes             string            `json:"notes" db:"notes"`
	CreatedBy         *string           `json:"created_by" db:"created_by"`
	CreatedAt         time.Time         `json:"created_at" db:"created_at"`
	Items             []StockOpnameItem `json:"items,omitempty"`
}

// Medicine Disposal (Pemusnahan Obat)
type MedicineDisposal struct {
	ID             string    `json:"id" db:"id"`
	DisposalNumber string    `json:"disposal_number" db:"disposal_number"`
	ItemID         int       `json:"item_id" db:"item_id"`
	ItemName       string    `json:"item_name" db:"item_name"`
	BatchNumber    string    `json:"batch_number" db:"batch_number"`
	ExpiryDate     string    `json:"expiry_date" db:"expiry_date"`
	Qty            int       `json:"qty" db:"qty"`
	UnitPrice      float64   `json:"unit_price" db:"unit_price"`
	TotalLoss      float64   `json:"total_loss" db:"total_loss"`
	Reason         string    `json:"reason" db:"reason"`
	WitnessName    string    `json:"witness_name" db:"witness_name"`
	CreatedBy      *string   `json:"created_by" db:"created_by"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}
