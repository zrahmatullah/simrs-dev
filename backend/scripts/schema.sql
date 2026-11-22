-- ========================================================
-- SIMRS-ASOY (E-Clinic) PostgreSQL Schema & Initial Seeder
-- Aligned with AGENTS.md Architecture & Standards
-- ========================================================

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Drop existing tables if needed (clean slate)
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS medical_certificates CASCADE;
DROP TABLE IF EXISTS stock_mutations CASCADE;
DROP TABLE IF EXISTS purchase_order_items CASCADE;
DROP TABLE IF EXISTS purchase_orders CASCADE;
DROP TABLE IF EXISTS cashier_shifts CASCADE;
DROP TABLE IF EXISTS billing_payments CASCADE;
DROP TABLE IF EXISTS invoice_items CASCADE;
DROP TABLE IF EXISTS invoices CASCADE;
DROP TABLE IF EXISTS ancillary_orders CASCADE;
DROP TABLE IF EXISTS prescription_items CASCADE;
DROP TABLE IF EXISTS prescriptions CASCADE;
DROP TABLE IF EXISTS medical_records CASCADE;
DROP TABLE IF EXISTS vital_signs CASCADE;
DROP TABLE IF EXISTS appointments CASCADE;
DROP TABLE IF EXISTS visits CASCADE;
DROP TABLE IF EXISTS patients CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS inventory_items CASCADE;
DROP TABLE IF EXISTS inventory_categories CASCADE;
DROP TABLE IF EXISTS icd9_codes CASCADE;
DROP TABLE IF EXISTS icd10_codes CASCADE;
DROP TABLE IF EXISTS guarantors CASCADE;
DROP TABLE IF EXISTS poli_units CASCADE;
DROP TABLE IF EXISTS roles CASCADE;

-- 1. MASTER TABLES (BIGSERIAL / INT)
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE poli_units (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    status_aktif BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE guarantors (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    category VARCHAR(50) NOT NULL DEFAULT 'UMUM', -- UMUM, BPJS, ASURANSI
    status_aktif BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE icd10_codes (
    id SERIAL PRIMARY KEY,
    code VARCHAR(20) NOT NULL UNIQUE,
    name_id VARCHAR(255) NOT NULL,
    name_en VARCHAR(255),
    status_aktif BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE icd9_codes (
    id SERIAL PRIMARY KEY,
    code VARCHAR(20) NOT NULL UNIQUE,
    name_id VARCHAR(255) NOT NULL,
    tariff NUMERIC(15, 2) NOT NULL DEFAULT 0,
    status_aktif BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE inventory_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE inventory_items (
    id SERIAL PRIMARY KEY,
    category_id INT REFERENCES inventory_categories(id) ON DELETE SET NULL,
    item_code VARCHAR(50) NOT NULL UNIQUE,
    item_name VARCHAR(200) NOT NULL,
    generic_name VARCHAR(200),
    unit VARCHAR(50) NOT NULL, -- Tablet, Botol, Tube, Box, Ampul, Pcs
    buy_price NUMERIC(15, 2) NOT NULL DEFAULT 0,
    sell_price NUMERIC(15, 2) NOT NULL DEFAULT 0,
    minimum_stock INT NOT NULL DEFAULT 10,
    total_stock INT NOT NULL DEFAULT 0,
    depo_stock INT NOT NULL DEFAULT 0, -- Stock in Pharmacy Depo
    is_medicine BOOLEAN NOT NULL DEFAULT TRUE,
    rack_location VARCHAR(50),
    status_aktif BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. USER & RBAC (UUID)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(150) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(150) NOT NULL,
    nip VARCHAR(50),
    phone VARCHAR(30),
    role_id INT NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
    poli_id INT REFERENCES poli_units(id) ON DELETE SET NULL,
    status_aktif BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,
    created_by UUID NULL,
    updated_by UUID NULL
);

-- 3. PATIENTS (UUID)
CREATE TABLE patients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    no_rm VARCHAR(50) NOT NULL UNIQUE,
    nik VARCHAR(20) NOT NULL,
    nama_pasien VARCHAR(150) NOT NULL,
    jenis_kelamin VARCHAR(10) NOT NULL, -- L / P
    tempat_lahir VARCHAR(100),
    tanggal_lahir DATE NOT NULL,
    nama_orang_tua VARCHAR(150),
    no_hp VARCHAR(30),
    alamat TEXT,
    riwayat_alergi TEXT,
    golongan_darah VARCHAR(5),
    status_aktif BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL
);

-- 4. VISITS / REGISTRATIONS (UUID)
CREATE TABLE visits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    no_rawat VARCHAR(50) NOT NULL UNIQUE,
    no_antrian VARCHAR(20) NOT NULL,
    patient_id UUID NOT NULL REFERENCES patients(id) ON DELETE RESTRICT,
    poli_id INT NOT NULL REFERENCES poli_units(id) ON DELETE RESTRICT,
    doctor_id UUID REFERENCES users(id) ON DELETE SET NULL,
    guarantor_id INT NOT NULL REFERENCES guarantors(id) ON DELETE RESTRICT,
    nomor_asuransi VARCHAR(50),
    visit_date DATE NOT NULL DEFAULT CURRENT_DATE,
    keluhan_utama TEXT,
    status VARCHAR(30) NOT NULL DEFAULT 'MENUNGGU_TRIASE', 
    -- Status: MENUNGGU_TRIASE, DIPERIKSA_DOKTER, FARMASI, KASIR, SELESAI, BATAL
    waktu_masuk TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    waktu_selesai TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL
);

-- 5. APPOINTMENTS / CALL CENTER (UUID)
CREATE TABLE appointments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id UUID NOT NULL REFERENCES patients(id) ON DELETE RESTRICT,
    poli_id INT NOT NULL REFERENCES poli_units(id) ON DELETE RESTRICT,
    doctor_id UUID REFERENCES users(id) ON DELETE SET NULL,
    appointment_date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'TERJADWAL', -- TERJADWAL, HADIR, RESCHEDULE, BATAL
    channel VARCHAR(30) NOT NULL DEFAULT 'TELEPON', -- TELEPON, WHATSAPP, WEBSITE, LOKET
    catatan TEXT,
    checked_in_at TIMESTAMPTZ NULL,
    checked_in_by UUID REFERENCES users(id) ON DELETE SET NULL,
    visit_id UUID REFERENCES visits(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL
);

-- 6. VITAL SIGNS / TTV NURSE (UUID)
CREATE TABLE vital_signs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    visit_id UUID NOT NULL REFERENCES visits(id) ON DELETE CASCADE,
    patient_id UUID NOT NULL REFERENCES patients(id) ON DELETE RESTRICT,
    nurse_id UUID REFERENCES users(id) ON DELETE SET NULL,
    td_sistole INT,
    td_diastole INT,
    hr NUMERIC(5, 1), -- Heart Rate / Nadi
    rr NUMERIC(5, 1), -- Respiration Rate / Laju Napas
    temp NUMERIC(4, 1), -- Suhu Celcius
    spo2 NUMERIC(5, 1), -- SpO2 %
    bb NUMERIC(5, 2), -- Berat Badan kg
    tb NUMERIC(5, 2), -- Tinggi Badan cm
    lk NUMERIC(5, 2), -- Lingkar Kepala cm (pediatrik)
    skala_nyeri INT DEFAULT 0, -- 0-10
    kesadaran VARCHAR(50) DEFAULT 'Compos Mentis',
    triase_level VARCHAR(20) DEFAULT 'HIJAU', -- MERAH, KUNING, HIJAU, HITAM
    catatan_keperawatan TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL
);

-- 7. MEDICAL RECORDS / DOCTOR EMR (UUID)
CREATE TABLE medical_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    visit_id UUID NOT NULL REFERENCES visits(id) ON DELETE CASCADE,
    patient_id UUID NOT NULL REFERENCES patients(id) ON DELETE RESTRICT,
    doctor_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    subjektif TEXT NOT NULL,
    objektif TEXT NOT NULL,
    asesmen TEXT NOT NULL,
    planning TEXT NOT NULL,
    icd10_code VARCHAR(20) REFERENCES icd10_codes(code) ON DELETE SET NULL,
    icd9_code VARCHAR(20) REFERENCES icd9_codes(code) ON DELETE SET NULL,
    diagnosis_deskripsi TEXT,
    prosedur_deskripsi TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'FINAL', -- DRAFT, FINAL
    is_locked BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL
);

-- 8. PRESCRIPTIONS / FARMASI (UUID)
CREATE TABLE prescriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nomor_resep VARCHAR(50) NOT NULL UNIQUE,
    visit_id UUID NOT NULL REFERENCES visits(id) ON DELETE CASCADE,
    patient_id UUID NOT NULL REFERENCES patients(id) ON DELETE RESTRICT,
    doctor_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status VARCHAR(30) NOT NULL DEFAULT 'MENUNGGU_TELAAH', 
    -- MENUNGGU_TELAAH, TELAAH_OK, PENYIAPAN, SELESAI, BATAL
    catatan_dokter TEXT,
    catatan_telaah TEXT,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    verified_by UUID REFERENCES users(id) ON DELETE SET NULL,
    verified_at TIMESTAMPTZ NULL,
    dispensed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    dispensed_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE prescription_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    prescription_id UUID NOT NULL REFERENCES prescriptions(id) ON DELETE CASCADE,
    item_id INT NOT NULL REFERENCES inventory_items(id) ON DELETE RESTRICT,
    signa VARCHAR(100) NOT NULL, -- e.g. "3x1 sendok teh sesudah makan"
    aturan_pakai VARCHAR(150),
    jumlah INT NOT NULL,
    harga_satuan NUMERIC(15, 2) NOT NULL DEFAULT 0,
    subtotal NUMERIC(15, 2) NOT NULL DEFAULT 0,
    is_racikan BOOLEAN NOT NULL DEFAULT FALSE,
    racikan_nama VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 9. ANCILLARY ORDERS / PENUNJANG (LAB/RAD/FISIO) (UUID)
CREATE TABLE ancillary_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nomor_order VARCHAR(50) NOT NULL UNIQUE,
    visit_id UUID NOT NULL REFERENCES visits(id) ON DELETE CASCADE,
    patient_id UUID NOT NULL REFERENCES patients(id) ON DELETE RESTRICT,
    doctor_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    jenis_order VARCHAR(30) NOT NULL, -- LAB, RAD, FISIOTERAPI, EKG
    nama_tindakan VARCHAR(200) NOT NULL,
    tarif NUMERIC(15, 2) NOT NULL DEFAULT 0,
    catatan_klinis TEXT,
    hasil_pemeriksaan TEXT,
    lampiran_url TEXT,
    status VARCHAR(30) NOT NULL DEFAULT 'MENUNGGU', -- MENUNGGU, DALAM_PROSES, SELESAI, BATAL
    verified_by UUID REFERENCES users(id) ON DELETE SET NULL,
    verified_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL
);

-- 10. BILLING & INVOICES / KASIR (UUID)
CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nomor_invoice VARCHAR(50) NOT NULL UNIQUE,
    visit_id UUID NOT NULL REFERENCES visits(id) ON DELETE RESTRICT,
    patient_id UUID NOT NULL REFERENCES patients(id) ON DELETE RESTRICT,
    subtotal NUMERIC(15, 2) NOT NULL DEFAULT 0,
    diskon NUMERIC(15, 2) NOT NULL DEFAULT 0,
    pajak NUMERIC(15, 2) NOT NULL DEFAULT 0,
    total_tagihan NUMERIC(15, 2) NOT NULL DEFAULT 0,
    total_terbayar NUMERIC(15, 2) NOT NULL DEFAULT 0,
    sisa_tagihan NUMERIC(15, 2) NOT NULL DEFAULT 0,
    status VARCHAR(30) NOT NULL DEFAULT 'BELUM_BAYAR', -- BELUM_BAYAR, LUNAS, BATAL
    jenis_penjamin VARCHAR(50) NOT NULL DEFAULT 'UMUM',
    nomor_penjamin VARCHAR(50),
    catatan TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE invoice_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    item_type VARCHAR(50) NOT NULL, -- PENDAFTARAN, KONSULTASI, TINDAKAN, OBAT, LAB, RAD
    reference_id VARCHAR(50),
    item_name VARCHAR(200) NOT NULL,
    qty INT NOT NULL DEFAULT 1,
    unit_price NUMERIC(15, 2) NOT NULL DEFAULT 0,
    subtotal NUMERIC(15, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE billing_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nomor_transaksi VARCHAR(50) NOT NULL UNIQUE,
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    kasir_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    nominal NUMERIC(15, 2) NOT NULL,
    metode_bayar VARCHAR(30) NOT NULL, -- TUNAI, QRIS, DEBIT, KREDIT, TRANSFER
    referensi_pembayaran VARCHAR(100),
    catatan TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE cashier_shifts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kasir_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    shift_name VARCHAR(50) NOT NULL, -- Shift Pagi, Shift Siang, Shift Malam
    waktu_buka TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    waktu_tutup TIMESTAMPTZ NULL,
    saldo_awal NUMERIC(15, 2) NOT NULL DEFAULT 0,
    total_tunai_sistem NUMERIC(15, 2) NOT NULL DEFAULT 0,
    total_non_tunai_sistem NUMERIC(15, 2) NOT NULL DEFAULT 0,
    saldo_akhir_fisik NUMERIC(15, 2) NULL,
    selisih NUMERIC(15, 2) NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN', -- OPEN, CLOSED
    catatan_shift TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 11. LOGISTICS & INVENTORY TRANSACTIONS (UUID)
CREATE TABLE purchase_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    po_number VARCHAR(50) NOT NULL UNIQUE,
    supplier_name VARCHAR(150) NOT NULL,
    order_date DATE NOT NULL DEFAULT CURRENT_DATE,
    expected_date DATE,
    status VARCHAR(30) NOT NULL DEFAULT 'SUBMITTED', -- DRAFT, SUBMITTED, RECEIVED, CANCELLED
    total_amount NUMERIC(15, 2) NOT NULL DEFAULT 0,
    catatan TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE purchase_order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    po_id UUID NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    item_id INT NOT NULL REFERENCES inventory_items(id) ON DELETE RESTRICT,
    qty INT NOT NULL,
    unit_price NUMERIC(15, 2) NOT NULL,
    subtotal NUMERIC(15, 2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE stock_mutations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id INT NOT NULL REFERENCES inventory_items(id) ON DELETE RESTRICT,
    transaction_type VARCHAR(30) NOT NULL, -- IN_PO, OUT_DISPENSE, MUTASI_DEPO, ADJUSTMENT
    qty INT NOT NULL,
    batch_number VARCHAR(50),
    expiry_date DATE,
    source_depo VARCHAR(50) DEFAULT 'GUDANG_UTAMA',
    dest_depo VARCHAR(50) DEFAULT 'DEPO_FARMASI',
    notes TEXT,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 12. MEDICAL CERTIFICATES (UUID)
CREATE TABLE medical_certificates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nomor_surat VARCHAR(50) NOT NULL UNIQUE,
    visit_id UUID NOT NULL REFERENCES visits(id) ON DELETE CASCADE,
    patient_id UUID NOT NULL REFERENCES patients(id) ON DELETE RESTRICT,
    doctor_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    jenis_surat VARCHAR(30) NOT NULL, -- SURAT_SAKIT, SURAT_SEHAT, SURAT_RUJUKAN
    diagnosa TEXT NOT NULL,
    catatan TEXT,
    tgl_mulai DATE,
    tgl_selesai DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

-- 13. AUDIT LOGS (BIGSERIAL)
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    username VARCHAR(100),
    action VARCHAR(50) NOT NULL, -- CREATE, UPDATE, DELETE, LOGIN, LOGOUT
    entity_type VARCHAR(50) NOT NULL, -- patient, visit, emr, prescription, invoice, etc.
    entity_id VARCHAR(100),
    details JSONB,
    ip_address VARCHAR(50),
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- INDEXES
CREATE INDEX idx_patients_no_rm ON patients(no_rm);
CREATE INDEX idx_patients_nik ON patients(nik);
CREATE INDEX idx_visits_no_rawat ON visits(no_rawat);
CREATE INDEX idx_visits_visit_date ON visits(visit_date);
CREATE INDEX idx_visits_status ON visits(status);
CREATE INDEX idx_visits_patient_id ON visits(patient_id);
CREATE INDEX idx_appointments_date ON appointments(appointment_date);
CREATE INDEX idx_appointments_status ON appointments(status);
CREATE INDEX idx_vital_signs_visit ON vital_signs(visit_id);
CREATE INDEX idx_medical_records_visit ON medical_records(visit_id);
CREATE INDEX idx_prescriptions_visit ON prescriptions(visit_id);
CREATE INDEX idx_prescriptions_status ON prescriptions(status);
CREATE INDEX idx_ancillary_orders_visit ON ancillary_orders(visit_id);
CREATE INDEX idx_ancillary_orders_status ON ancillary_orders(status);
CREATE INDEX idx_invoices_visit ON invoices(visit_id);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_inventory_items_code ON inventory_items(item_code);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);

-- ========================================================
-- SEED DATA
-- ========================================================

-- Roles
INSERT INTO roles (id, code, name, description) VALUES
(1, 'SUPERADMIN', 'Superadmin', 'Akses penuh seluruh sistem, manajemen user, tarif & audit trail'),
(2, 'REGISTRATION', 'Pegawai Pendaftaran', 'Registrasi pasien baru/lama, No. RM, pendaftaran kunjungan & antrean'),
(3, 'CALL_CENTER', 'Petugas Call Center', 'Manajemen jadwal dokter, appointment booking & check-in'),
(4, 'NURSE', 'Perawat', 'Antrean poli perawat, pengkajian awal, Tanda-Tanda Vital (TTV) & triase'),
(5, 'DOCTOR', 'Dokter', 'Akses EMR, SOAP, diagnosa ICD-10/ICD-9, e-Resep & order penunjang'),
(6, 'PHARMACY', 'Pegawai Farmasi', 'Antrean e-Resep, telaah resep, penyiapan obat, etiket & stok depo'),
(7, 'SUPPORT', 'Pegawai Penunjang', 'Pemeriksaan laboratorium, radiologi, fisioterapi & validasi hasil'),
(8, 'LOGISTICS', 'Pegawai Logistik', 'Manajemen gudang utama, purchase orders, penerimaan & mutasi stok'),
(9, 'CASHIER', 'Kasir', 'Billing pendaftaran, tindakan, obat, pembayaran multi-metode & shift closing');

ALTER SEQUENCE roles_id_seq RESTART WITH 10;

-- Poli Units
INSERT INTO poli_units (id, code, name, description) VALUES
(1, 'POLI-UMUM', 'Poli Umum', 'Pelayanan rawat jalan umum dewasa dan anak'),
(2, 'POLI-ANAK', 'Poli Anak & Tumbuh Kembang', 'Pelayanan spesialis anak dan terapi tumbuh kembang'),
(3, 'POLI-GIGI', 'Poli Gigi & Mulut', 'Pemeriksaan kesehatan gigi dan mulut'),
(4, 'POLI-FISIO', 'Poli Fisioterapi / Rehabilitasi Medis', 'Layanan terapi okupasi, wicara, dan fisik'),
(5, 'IGD', 'Instalasi Gawat Darurat (IGD)', 'Pelayanan kedaruratan medik 24 jam'),
(6, 'LAB', 'Laboratorium Medis', 'Pemeriksaan hematologi, kimia darah, dan urin'),
(7, 'RAD', 'Radiologi', 'Pemeriksaan X-Ray, USG, dan pencitraan medik'),
(8, 'FAR', 'Instalasi Farmasi Depo', 'Dispensing obat dan perbekalan farmasi');

ALTER SEQUENCE poli_units_id_seq RESTART WITH 9;

-- Guarantors
INSERT INTO guarantors (id, code, name, category) VALUES
(1, 'UMUM', 'Umum / Pembayaran Mandiri', 'UMUM'),
(2, 'BPJS-PBI', 'BPJS Kesehatan - PBI', 'BPJS'),
(3, 'BPJS-NONPBI', 'BPJS Kesehatan - Non PBI', 'BPJS'),
(4, 'ASR-PRU', 'Asuransi Swasta - Prudential', 'ASURANSI'),
(5, 'ASR-INHEALTH', 'Asuransi Swasta - Mandiri Inhealth', 'ASURANSI');

ALTER SEQUENCE guarantors_id_seq RESTART WITH 6;

-- ICD-10 Codes
INSERT INTO icd10_codes (code, name_id, name_en) VALUES
('F80.9', 'Gangguan Perkembangan Bicara dan Bahasa, Tidak Spesifik', 'Developmental disorder of speech and language, unspecified'),
('F84.0', 'Autisme Masa Kanak-Kanak', 'Childhood autism'),
('F90.0', 'Gangguan Aktivitas dan Perhatian (ADHD)', 'Disturbance of activity and attention'),
('J06.9', 'Infeksi Saluran Pernapasan Akut Atas (ISPA)', 'Acute upper respiratory infection, unspecified'),
('K29.7', 'Gastritis, Tidak Spesifik', 'Gastritis, unspecified'),
('R50.9', 'Demam, Tidak Spesifik', 'Fever, unspecified'),
('A09.9', 'Gastroenteritis dan Kolitis Akut', 'Gastroenteritis and colitis, unspecified'),
('Z00.0', 'Pemeriksaan Medis Umum', 'General medical examination');

-- ICD-9 Codes (Procedures)
INSERT INTO icd9_codes (code, name_id, tariff) VALUES
('89.07', 'Konsultasi dan Pemeriksaan Dokter Spesialis', 150000.00),
('89.03', 'Konsultasi dan Pemeriksaan Dokter Umum', 75000.00),
('93.39', 'Terapi Wicara dan Komunikasi Interaktif', 175000.00),
('93.89', 'Terapi Okupasi dan Sensorik Integrasi', 200000.00),
('93.11', 'Latihan Fisik Fisioterapi Motorik Kasar', 150000.00),
('90.59', 'Pemeriksaan Darah Rutin / Lengkap', 85000.00),
('87.44', 'Rontgen Thorax AP / PA', 160000.00),
('96.59', 'Nebulisasi Terapi Saluran Pernapasan', 65000.00);

-- Inventory Categories
INSERT INTO inventory_categories (id, name, description) VALUES
(1, 'Obat Sirup Anak', 'Sediaan cair oral untuk anak'),
(2, 'Obat Tablet & Kapsul', 'Sediaan padat oral'),
(3, 'Obat Tetes & Salep', 'Sediaan tetes dan topikal'),
(4, 'Alat Kesehatan & BHP', 'Barang habis pakai klinis medis'),
(5, 'Suplemen & Vitamin', 'Suplemen daya tahan tubuh dan nutrisi');

ALTER SEQUENCE inventory_categories_id_seq RESTART WITH 6;

-- Inventory Items
INSERT INTO inventory_items (category_id, item_code, item_name, generic_name, unit, buy_price, sell_price, minimum_stock, total_stock, depo_stock, is_medicine, rack_location) VALUES
(1, 'OBAT-001', 'Paracetamol Sirup 120mg/5ml 60ml', 'Paracetamol', 'Botol', 8000.00, 15000.00, 20, 150, 45, true, 'RAK-A1'),
(1, 'OBAT-002', 'Amoxicillin Sirup Kering 125mg/5ml', 'Amoxicillin Trihydrate', 'Botol', 9500.00, 18000.00, 15, 80, 25, true, 'RAK-A2'),
(1, 'OBAT-003', 'Cetirizine Sirup 5mg/5ml 60ml', 'Cetirizine HCl', 'Botol', 18000.00, 32000.00, 10, 60, 20, true, 'RAK-A3'),
(5, 'OBAT-004', 'Curcuma Plus Sirup Emulsi 200ml', 'Curcuma + Multivitamin', 'Botol', 22000.00, 38000.00, 10, 50, 15, true, 'RAK-B1'),
(3, 'OBAT-005', 'Gentamicin Salep Kulit 0.1% 5g', 'Gentamicin Sulfate', 'Tube', 7000.00, 14000.00, 10, 40, 12, true, 'RAK-B2'),
(2, 'OBAT-006', 'Ibuprofen Tablet 400mg', 'Ibuprofen', 'Strip', 4500.00, 8500.00, 30, 200, 60, true, 'RAK-C1'),
(4, 'BHP-001', 'Spuit 3cc Terumo Disposible', 'Syringe 3ml', 'Pcs', 1500.00, 3500.00, 50, 500, 120, false, 'RAK-D1'),
(4, 'BHP-002', 'Alkohol Swab Sensi Box (Isi 100)', 'Isopropyl Alcohol 70%', 'Box', 12000.00, 20000.00, 10, 45, 10, false, 'RAK-D2'),
(4, 'BHP-003', 'Kassa Steril Onemed Box', 'Hydrophile Gauze', 'Box', 8500.00, 15000.00, 15, 60, 18, false, 'RAK-D3'),
(4, 'BHP-004', 'Masker Bedah 3-Ply Earloop Box', 'Medical Surgical Mask', 'Box', 18000.00, 30000.00, 20, 80, 25, false, 'RAK-D4');

-- Users for all 9 Roles (password is 'password123')
-- Bcrypt Hash: $2a$10$luZpclYpdE9s/xmUtDT14uRXxm60p8Vm4jgzZwkGfRmOrCkid58wi
INSERT INTO users (id, username, email, password_hash, full_name, nip, phone, role_id, poli_id) VALUES
('a0000000-0000-0000-0000-000000000001', 'superadmin', 'admin@simrs-asoy.local', '$2a$10$luZpclYpdE9s/xmUtDT14uRXxm60p8Vm4jgzZwkGfRmOrCkid58wi', 'Super Administrator, S.Kom', 'ADM-001', '081234567801', 1, NULL),
('a0000000-0000-0000-0000-000000000002', 'pendaftaran', 'regis@simrs-asoy.local', '$2a$10$luZpclYpdE9s/xmUtDT14uRXxm60p8Vm4jgzZwkGfRmOrCkid58wi', 'Siti Rahmawati, A.Md.RMIK', 'REG-002', '081234567802', 2, 1),
('a0000000-0000-0000-0000-000000000003', 'callcenter', 'callcenter@simrs-asoy.local', '$2a$10$luZpclYpdE9s/xmUtDT14uRXxm60p8Vm4jgzZwkGfRmOrCkid58wi', 'Budi Santoso, S.Sos', 'CC-003', '081234567803', 3, NULL),
('a0000000-0000-0000-0000-000000000004', 'perawat', 'nurse@simrs-asoy.local', '$2a$10$luZpclYpdE9s/xmUtDT14uRXxm60p8Vm4jgzZwkGfRmOrCkid58wi', 'Ns. Dian Anggraini, S.Kep', 'NUR-004', '081234567804', 4, 2),
('a0000000-0000-0000-0000-000000000005', 'dokter', 'doctor@simrs-asoy.local', '$2a$10$luZpclYpdE9s/xmUtDT14uRXxm60p8Vm4jgzZwkGfRmOrCkid58wi', 'dr. Aris Munandar, Sp.A', 'DOC-005', '081234567805', 5, 2),
('a0000000-0000-0000-0000-000000000006', 'farmasi', 'pharmacy@simrs-asoy.local', '$2a$10$luZpclYpdE9s/xmUtDT14uRXxm60p8Vm4jgzZwkGfRmOrCkid58wi', 'apt. Rina Marlina, S.Farm', 'PHA-006', '081234567806', 6, 8),
('a0000000-0000-0000-0000-000000000007', 'penunjang', 'support@simrs-asoy.local', '$2a$10$luZpclYpdE9s/xmUtDT14uRXxm60p8Vm4jgzZwkGfRmOrCkid58wi', 'Ahmad Fauzi, A.Md.AK', 'SUP-007', '081234567807', 7, 6),
('a0000000-0000-0000-0000-000000000008', 'logistik', 'logistics@simrs-asoy.local', '$2a$10$luZpclYpdE9s/xmUtDT14uRXxm60p8Vm4jgzZwkGfRmOrCkid58wi', 'Hendra Wijaya, S.E', 'LOG-008', '081234567808', 8, NULL),
('a0000000-0000-0000-0000-000000000009', 'kasir', 'cashier@simrs-asoy.local', '$2a$10$luZpclYpdE9s/xmUtDT14uRXxm60p8Vm4jgzZwkGfRmOrCkid58wi', 'Dewi Lestari, A.Md.Akt', 'CSH-009', '081234567809', 9, NULL);

-- Sample Patients
INSERT INTO patients (id, no_rm, nik, nama_pasien, jenis_kelamin, tempat_lahir, tanggal_lahir, nama_orang_tua, no_hp, alamat, riwayat_alergi, golongan_darah) VALUES
('b0000000-0000-0000-0000-000000000001', 'RM-2026-0001', '3201015504180001', 'Hilmi Al-Fatih', 'L', 'Bandung', '2020-04-15', 'Rahman & Anisa', '081321000111', 'Jl. Sukajadi No. 45, Bandung', 'Alergi Penisilin, Seafood', 'O'),
('b0000000-0000-0000-0000-000000000002', 'RM-2026-0002', '3201026008210002', 'Aisyah Putri Azzahra', 'P', 'Cimahi', '2022-08-20', 'Bambang & Ratih', '081321000222', 'Jl. Cihanjuang No. 12, Cimahi', NULL, 'A'),
('b0000000-0000-0000-0000-000000000003', 'RM-2026-0003', '3201031102190003', 'Kenzie Raditya', 'L', 'Bandung', '2019-02-11', 'Agus & Maya', '081321000333', 'Jl. Buah Batu No. 88, Bandung', 'Alergi Dingin, Amoxicillin', 'B');

-- Sample Active Visits (Different lifecycle stages)
INSERT INTO visits (id, no_rawat, no_antrian, patient_id, poli_id, doctor_id, guarantor_id, visit_date, keluhan_utama, status, waktu_masuk) VALUES
-- Visit 1: Sedang Menunggu Triase Perawat
('c0000000-0000-0000-0000-000000000001', 'RAWAT-20261122-001', 'A-001', 'b0000000-0000-0000-0000-000000000001', 2, 'a0000000-0000-0000-0000-000000000005', 1, CURRENT_DATE, 'Demam 3 hari naik turun disertai batuk pilek', 'MENUNGGU_TRIASE', NOW() - INTERVAL '45 MINUTE'),
-- Visit 2: Sedang Diperiksa Dokter
('c0000000-0000-0000-0000-000000000002', 'RAWAT-20261122-002', 'A-002', 'b0000000-0000-0000-0000-000000000002', 2, 'a0000000-0000-0000-0000-000000000005', 2, CURRENT_DATE, 'Konsultasi tumbuh kembang bicara terlambat', 'DIPERIKSA_DOKTER', NOW() - INTERVAL '30 MINUTE'),
-- Visit 3: Di Farmasi / Menunggu Obat
('c0000000-0000-0000-0000-000000000003', 'RAWAT-20261122-003', 'A-003', 'b0000000-0000-0000-0000-000000000003', 2, 'a0000000-0000-0000-0000-000000000005', 1, CURRENT_DATE, 'Pemeriksaan rutin dan terapi wicara', 'FARMASI', NOW() - INTERVAL '90 MINUTE');

-- Sample Vital Signs (for Visit 2 & 3)
INSERT INTO vital_signs (id, visit_id, patient_id, nurse_id, td_sistole, td_diastole, hr, rr, temp, spo2, bb, tb, lk, skala_nyeri, kesadaran, triase_level, catatan_keperawatan) VALUES
('d0000000-0000-0000-0000-000000000002', 'c0000000-0000-0000-0000-000000000002', 'b0000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000004', 100, 65, 92.0, 22.0, 36.6, 99.0, 14.5, 96.0, 48.0, 0, 'Compos Mentis', 'HIJAU', 'Anak aktif dan kooperatif, tidak ada tanda distres'),
('d0000000-0000-0000-0000-000000000003', 'c0000000-0000-0000-0000-000000000003', 'b0000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000004', 105, 70, 88.0, 20.0, 36.8, 98.0, 21.0, 115.0, 50.0, 1, 'Compos Mentis', 'HIJAU', 'Tanda vital dalam batas normal anak');

-- Sample Medical Record EMR (for Visit 3)
INSERT INTO medical_records (id, visit_id, patient_id, doctor_id, subjektif, objektif, asesmen, planning, icd10_code, icd9_code, diagnosis_deskripsi, prosedur_deskripsi, status) VALUES
('e0000000-0000-0000-0000-000000000003', 'c0000000-0000-0000-0000-000000000003', 'b0000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000005',
'Ibu mengeluhkan anak masih kesulitan merangkai lebih dari 2 kata pada usia 5 tahun.',
'Keadaan umum baik, kontak mata cukup, organ wicara normal, TTV stabil.',
'Suspek Gangguan Perkembangan Bahasa Ekspresif.',
'Jadwalkan Terapi Wicara 2x seminggu. Edukasi orang tua untuk stimulasi bicara di rumah.',
'F80.9', '93.39', 'Speech & language developmental disorder', 'Terapi wicara dan komunikasi', 'FINAL');

-- Sample Prescription (for Visit 3)
INSERT INTO prescriptions (id, nomor_resep, visit_id, patient_id, doctor_id, status, catatan_dokter, catatan_telaah, is_verified, verified_by, verified_at) VALUES
('f0000000-0000-0000-0000-000000000003', 'RXP-20261122-001', 'c0000000-0000-0000-0000-000000000003', 'b0000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000005',
'MENUNGGU_TELAAH', 'Berikan suplemen nafsu makan dan multivitamin', 'Resep memenuhi persyaratan administratif dan klinis', false, NULL, NULL);

INSERT INTO prescription_items (id, prescription_id, item_id, signa, aturan_pakai, jumlah, harga_satuan, subtotal, is_racikan) VALUES
(gen_random_uuid(), 'f0000000-0000-0000-0000-000000000003', 4, '1 x 1 sendok makan sehari', 'Sesudah makan pagi', 1, 38000.00, 38000.00, false);

-- Sample Ancillary Order (Lab order for Visit 2)
INSERT INTO ancillary_orders (id, nomor_order, visit_id, patient_id, doctor_id, jenis_order, nama_tindakan, tarif, catatan_klinis, status) VALUES
('70000000-0000-0000-0000-000000000002', 'ORD-20261122-001', 'c0000000-0000-0000-0000-000000000002', 'b0000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000005',
'LAB', 'Pemeriksaan Darah Rutin / Lengkap', 85000.00, 'Evaluasi status hematologi dan nutrisi anak', 'MENUNGGU');

-- Sample Appointments for Call Center
INSERT INTO appointments (id, patient_id, poli_id, doctor_id, appointment_date, start_time, end_time, status, channel, catatan) VALUES
('80000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001', 2, 'a0000000-0000-0000-0000-000000000005', CURRENT_DATE, '09:00:00', '09:30:00', 'TERJADWAL', 'WHATSAPP', 'Pasien booking melalui WA konfirmasi kehadiran jam 08:30'),
('80000000-0000-0000-0000-000000000002', 'b0000000-0000-0000-0000-000000000002', 4, NULL, CURRENT_DATE + INTERVAL '1 DAY', '10:00:00', '11:00:00', 'TERJADWAL', 'TELEPON', 'Sesi evaluasi sensori integrasi');

-- Sample Open Cashier Shift
INSERT INTO cashier_shifts (id, kasir_id, shift_name, waktu_buka, saldo_awal, status) VALUES
('90000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000009', 'Shift Pagi', NOW() - INTERVAL '3 HOUR', 500000.00, 'OPEN');

-- Sample Purchase Order in Logistics
INSERT INTO purchase_orders (id, po_number, supplier_name, order_date, expected_date, status, total_amount, catatan, created_by) VALUES
('60000000-0000-0000-0000-000000000001', 'PO-202611-001', 'PT. Kimia Farma Trading & Distribution', CURRENT_DATE - INTERVAL '2 DAY', CURRENT_DATE + INTERVAL '1 DAY', 'SUBMITTED', 1850000.00, 'Pengadaan stok sirup antibiotik dan BHP bulanan', 'a0000000-0000-0000-0000-000000000008');

INSERT INTO purchase_order_items (id, po_id, item_id, qty, unit_price, subtotal) VALUES
(gen_random_uuid(), '60000000-0000-0000-0000-000000000001', 1, 50, 8000.00, 400000.00),
(gen_random_uuid(), '60000000-0000-0000-0000-000000000001', 2, 50, 9500.00, 475000.00),
(gen_random_uuid(), '60000000-0000-0000-0000-000000000001', 7, 200, 1500.00, 300000.00),
(gen_random_uuid(), '60000000-0000-0000-0000-000000000001', 8, 30, 12000.00, 360000.00),
(gen_random_uuid(), '60000000-0000-0000-0000-000000000001', 9, 35, 8500.00, 297500.00);

-- Sample Stock Mutations (Initial Stock)
INSERT INTO stock_mutations (item_id, transaction_type, qty, batch_number, expiry_date, source_depo, dest_depo, notes) VALUES
(1, 'IN_PO', 150, 'BATCH-PCM-01', CURRENT_DATE + INTERVAL '500 DAY', 'SUPPLIER', 'GUDANG_UTAMA', 'Penerimaan stok awal'),
(2, 'IN_PO', 80, 'BATCH-AMX-01', CURRENT_DATE + INTERVAL '400 DAY', 'SUPPLIER', 'GUDANG_UTAMA', 'Penerimaan stok awal'),
(3, 'IN_PO', 60, 'BATCH-CTZ-01', CURRENT_DATE + INTERVAL '600 DAY', 'SUPPLIER', 'GUDANG_UTAMA', 'Penerimaan stok awal');

-- Initial System Audit Log
INSERT INTO audit_logs (username, action, entity_type, entity_id, details) VALUES
('superadmin', 'INITIALIZE', 'system', 'system', '{"message": "Database schema and initial seed data initialized successfully"}');
