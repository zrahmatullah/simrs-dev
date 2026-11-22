// API Types matching Backend Envelope & Models

export interface Meta {
  page: number
  limit: number
  total_data: number
  total_page: number
}

export interface ApiResponse<T = any> {
  success: boolean
  message: string
  data: T
  meta?: Meta
  errors?: any
}

export interface User {
  id: string
  username: string
  email: string
  full_name: string
  nama_lengkap?: string
  nip: string
  phone: string
  role_id: number
  role_code: string
  role_name: string
  poli_id?: number | null
  poli_name?: string
  status_aktif: boolean
  created_at: string
  updated_at: string
}

export interface Patient {
  id: string
  no_rm: string
  nik: string
  nama_pasien: string
  jenis_kelamin: string
  tempat_lahir?: string
  tanggal_lahir: string
  nama_orang_tua?: string
  no_hp?: string
  alamat?: string
  riwayat_alergi?: string | null
  golongan_darah?: string
  status_aktif: boolean
  created_at: string
  updated_at: string
}

export interface Visit {
  id: string
  no_rawat: string
  no_antrian: string
  nomor_antrean?: string
  patient_id: string
  poli_id: number
  doctor_id?: string | null
  guarantor_id: number
  nomor_asuransi?: string | null
  visit_date: string
  keluhan_utama: string
  status: 'MENUNGGU_TRIASE' | 'MENUNGGU_DOKTER' | 'DIPERIKSA_DOKTER' | 'FARMASI' | 'KASIR' | 'SELESAI' | 'BATAL' | string
  waktu_masuk: string
  waktu_selesai?: string | null
  created_at: string
  updated_at: string

  patient_name?: string
  no_rm?: string
  patient_gender?: string
  jenis_kelamin?: string
  patient_dob?: string
  tanggal_lahir?: string
  patient_allergy?: string | null
  alergi?: string | null
  poli_name?: string
  doctor_name?: string
  guarantor_name?: string
}

export interface Appointment {
  id: string
  patient_id: string
  poli_id: number
  doctor_id?: string | null
  appointment_date: string
  start_time: string
  end_time: string
  status: 'TERJADWAL' | 'HADIR' | 'RESCHEDULE' | 'BATAL'
  channel: string
  catatan: string
  checked_in_at?: string | null
  created_at: string

  patient_name?: string
  no_rm?: string
  no_hp?: string
  poli_name?: string
  doctor_name?: string
}

export interface VitalSign {
  id: string
  visit_id: string
  patient_id: string
  nurse_id?: string | null
  nurse_name?: string
  td_sistole?: number | null
  td_diastole?: number | null
  hr?: number | null
  rr?: number | null
  temp?: number | null
  spo2?: number | null
  bb?: number | null
  tb?: number | null
  lk?: number | null
  skala_nyeri: number
  kesadaran: string
  triase_level: 'MERAH' | 'KUNING' | 'HIJAU' | 'HITAM'
  catatan_keperawatan: string
  created_at: string
}

export interface MedicalRecord {
  id: string
  visit_id: string
  patient_id: string
  doctor_id: string
  doctor_name?: string
  subjektif: string
  objektif: string
  asesmen: string
  planning: string
  icd10_code?: string | null
  icd9_code?: string | null
  diagnosis_deskripsi: string
  prosedur_deskripsi: string
  status: 'DRAFT' | 'FINAL'
  is_locked: boolean
  created_at: string
  updated_at: string
}

export interface PrescriptionItem {
  id: string
  prescription_id: string
  item_id: number
  item_name?: string
  signa: string
  aturan_pakai: string
  jumlah: number
  harga_satuan: number
  subtotal: number
  is_racikan: boolean
  racikan_nama?: string
  created_at: string
}

export interface Prescription {
  id: string
  nomor_resep: string
  visit_id: string
  patient_id: string
  doctor_id: string
  doctor_name?: string
  patient_name?: string
  no_rm?: string
  status: 'MENUNGGU_TELAAH' | 'TELAAH_OK' | 'PENYIAPAN' | 'SELESAI' | 'BATAL'
  catatan_dokter: string
  catatan_telaah?: string
  is_verified: boolean
  verified_by?: string | null
  verified_at?: string | null
  dispensed_by?: string | null
  dispensed_at?: string | null
  created_at: string
  items?: PrescriptionItem[]
}

export interface AncillaryOrder {
  id: string
  nomor_order: string
  visit_id: string
  patient_id: string
  patient_name?: string
  no_rm?: string
  doctor_id: string
  doctor_name?: string
  jenis_order: 'LAB' | 'RAD' | 'FISIOTERAPI' | 'EKG'
  nama_tindakan: string
  tarif: number
  catatan_klinis: string
  hasil_pemeriksaan?: string
  lampiran_url?: string
  status: 'MENUNGGU' | 'DALAM_PROSES' | 'SELESAI' | 'BATAL'
  verified_by?: string | null
  verified_at?: string | null
  created_at: string
}

export interface InvoiceItem {
  id: string
  invoice_id: string
  item_type: string
  reference_id: string
  item_name: string
  qty: number
  unit_price: number
  subtotal: number
  created_at: string
}

export interface Invoice {
  id: string
  nomor_invoice: string
  visit_id: string
  no_rawat?: string
  patient_id: string
  patient_name?: string
  no_rm?: string
  subtotal: number
  diskon: number
  pajak: number
  total_tagihan: number
  total_terbayar: number
  sisa_tagihan: number
  status: 'BELUM_BAYAR' | 'LUNAS' | 'BATAL'
  jenis_penjamin: string
  nomor_penjamin?: string | null
  catatan: string
  promo_code?: string | null
  promo_name?: string | null
  created_at: string
  items?: InvoiceItem[]
  payments?: BillingPayment[]
}

export interface BillingPayment {
  id: string
  nomor_transaksi: string
  invoice_id: string
  kasir_id: string
  kasir_name?: string
  nominal: number
  metode_bayar: 'TUNAI' | 'QRIS' | 'DEBIT' | 'KREDIT' | 'TRANSFER'
  referensi_pembayaran?: string
  catatan?: string
  uang_diterima?: number
  kembalian?: number
  created_at: string
}

export interface CashierShift {
  id: string
  kasir_id: string
  kasir_name?: string
  shift_name: string
  waktu_buka: string
  waktu_tutup?: string | null
  saldo_awal: number
  total_tunai_sistem: number
  total_non_tunai_sistem: number
  saldo_akhir_fisik?: number | null
  selisih?: number | null
  status: 'OPEN' | 'CLOSED'
  catatan_shift?: string
  created_at: string
}

export interface InventoryItem {
  id: number
  category_id: number
  category_name?: string
  item_code: string
  item_name: string
  generic_name: string
  unit: string
  buy_price: number
  sell_price: number
  minimum_stock: number
  total_stock: number
  depo_stock: number
  is_medicine: boolean
  rack_location: string
  status_aktif: boolean
  created_at: string
}

export interface Supplier {
  id: number
  code: string
  name: string
  pic_name?: string
  phone?: string
  email?: string
  address?: string
  city?: string
  status_aktif: boolean
  created_at?: string
}

export interface PurchaseOrderItem {
  id: string
  po_id: string
  item_id: number
  item_name?: string
  unit?: string
  qty: number
  unit_price: number
  subtotal: number
}

export interface PurchaseOrder {
  id: string
  po_number: string
  supplier_id?: number
  supplier_name: string
  order_date: string
  expected_date?: string | null
  status: 'DRAFT' | 'SUBMITTED' | 'RECEIVED' | 'CANCELLED'
  total_amount: number
  catatan: string
  created_at: string
  items?: PurchaseOrderItem[]
}

export interface StockMutation {
  id: string
  item_id: number
  item_name?: string
  transaction_type: 'IN_PO' | 'OUT_DISPENSE' | 'MUTASI_DEPO' | 'ADJUSTMENT'
  qty: number
  batch_number?: string
  expiry_date?: string | null
  source_depo: string
  dest_depo: string
  notes: string
  created_at: string
}

export interface MasterPoli {
  id: number
  code: string
  name: string
  description: string
  status_aktif?: boolean
}

export type PoliUnit = MasterPoli

export interface MasterGuarantor {
  id: number
  code: string
  name: string
  category: string
}

export interface MasterICD10 {
  id: number
  code: string
  name_id: string
  name_en?: string
}

export interface MasterICD9 {
  id: number
  code: string
  name_id: string
  tariff: number
}

export interface MasterCategory {
  id: number
  name: string
  description?: string
}

export interface AuditLog {
  id: number
  user_id?: string | null
  username: string
  action: string
  entity_type: string
  entity_id: string
  details: string
  ip_address: string
  user_agent: string
  created_at: string
}

export interface MasterProcedure {
  id: number
  code: string
  name: string
  category: string
  tariff: number
  status_aktif: boolean
  created_at: string
}

export interface VisitProcedure {
  id: string
  visit_id: string
  procedure_name: string
  tariff: number
  qty: number
  subtotal: number
  doctor_id?: string
  doctor_name?: string
  catatan?: string
  created_at: string
}

export interface NursingAssessment {
  id: string
  visit_id: string
  patient_id: string
  nurse_id?: string
  nurse_name?: string
  keluhan_keperawatan: string
  riwayat_penyakit_sekarang: string
  riwayat_alergi: string
  skala_nyeri_skor: number
  skala_nyeri_lokasi?: string
  skala_nyeri_durasi?: string
  risiko_jatuh_tingkat: 'RENDAH' | 'SEDANG' | 'TINGGI'
  skrining_gizi_turun_bb: boolean
  skrining_gizi_nafsu_makan: boolean
  masalah_keperawatan: string
  rencana_keperawatan: string
  evaluasi_keperawatan: string
  created_at: string
  updated_at: string
}

export interface Promo {
  id: number
  code: string
  name: string
  discount_type: 'PERCENTAGE' | 'FIXED'
  discount_value: number
  min_transaction: number
  start_date: string
  end_date: string
  status_aktif: boolean
  created_at: string
  updated_at: string
}

export interface PurchaseRequisitionItem {
  id: string
  pr_id: string
  item_id: number
  item_name?: string
  qty: number
  unit_price: number
  subtotal: number
}

export interface PurchaseRequisition {
  id: string
  pr_number: string
  requested_by?: string
  requester_name?: string
  request_date: string
  department: string
  status: 'DRAFT' | 'SUBMITTED' | 'APPROVED' | 'REJECTED' | 'CONVERTED_TO_PO'
  total_estimated: number
  notes?: string
  created_at: string
  updated_at: string
  items?: PurchaseRequisitionItem[]
}

export interface GoodsReceiptNote {
  id: string
  grn_number: string
  po_id: string
  po_number?: string
  supplier_name?: string
  received_date: string
  receiver_name: string
  delivery_order_number?: string
  status: string
  notes?: string
  created_at: string
}

export interface FaskesRujukan {
  id: number
  code: string
  name: string
  tipe: string
  alamat?: string
  kota?: string
  no_telp?: string
  status_aktif: boolean
  created_at?: string
  updated_at?: string
}

export interface MasterSpesialis {
  id: number
  code: string
  nama_dokter: string
  spesialisasi: string
  rumah_sakit?: string
  status_aktif: boolean
  created_at?: string
  updated_at?: string
}
