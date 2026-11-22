-- ========================================================
-- SIMRS-ASOY Database Migration Additions
-- Promos, Clinical Procedures, Nursing Assessments, PR & GRN
-- ========================================================

-- 1. Master Promo
CREATE TABLE IF NOT EXISTS promos (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(150) NOT NULL,
    discount_type VARCHAR(20) NOT NULL DEFAULT 'PERCENTAGE', -- PERCENTAGE / FIXED
    discount_value NUMERIC(15, 2) NOT NULL DEFAULT 0,
    min_transaction NUMERIC(15, 2) NOT NULL DEFAULT 0,
    start_date DATE NOT NULL DEFAULT CURRENT_DATE,
    end_date DATE NOT NULL DEFAULT CURRENT_DATE + INTERVAL '1 year',
    status_aktif BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed Sample Promos
INSERT INTO promos (code, name, discount_type, discount_value, min_transaction, start_date, end_date, status_aktif)
VALUES
('PROMOSEHAT10', 'Diskon Spesial Hari Kesehatan 10%', 'PERCENTAGE', 10, 150000, CURRENT_DATE, CURRENT_DATE + INTERVAL '6 month', true),
('POTONGAN25K', 'Voucher Potongan Biaya Rawat Jalan Rp 25.000', 'FIXED', 25000, 200000, CURRENT_DATE, CURRENT_DATE + INTERVAL '6 month', true),
('MEMBERVIP', 'Diskon Khusus Pasien VIP / Rekanan 15%', 'PERCENTAGE', 15, 100000, CURRENT_DATE, CURRENT_DATE + INTERVAL '1 year', true)
ON CONFLICT (code) DO NOTHING;

-- 2. Master & Visit Clinical Procedures (Tindakan di Ruangan Poli)
CREATE TABLE IF NOT EXISTS master_procedures (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(200) NOT NULL,
    category VARCHAR(100) NOT NULL DEFAULT 'TINDAKAN_POLI',
    tariff NUMERIC(15, 2) NOT NULL DEFAULT 0,
    status_aktif BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO master_procedures (code, name, category, tariff)
VALUES
('TND-001', 'Pemeriksaan Fisik Khusus & Konsultasi Spesialis', 'KONSULTASI', 150000),
('TND-002', 'Nebulisasi Anak / Dewasa per Sesi', 'TERAPI', 75000),
('TND-003', 'Injeksi Intramuskular / Intravena', 'INJEKSI', 35000),
('TND-004', 'Rawat Luka Ringan & Ganti Verban (Wound Dressing)', 'TINDAKAN_BEDAH', 60000),
('TND-005', 'Jahit Luka Sederhana (1-3 Simpul)', 'TINDAKAN_BEDAH', 120000),
('TND-006', 'Pembersihan Serumen Telinga (Ear Toilet)', 'TINDAKAN_THT', 85000),
('TND-007', 'Pengambilan Benda Asing Hidung / Telinga', 'TINDAKAN_THT', 100000),
('TND-008', 'Pemeriksaan EKG 12 Lead & Interpretasi', 'DIAGNOSTIK', 90000),
('TND-009', 'Suction Jalan Napas', 'TERAPI', 50000),
('TND-010', 'Pemasangan Kateter Urin', 'TINDAKAN_MEDIS', 95000)
ON CONFLICT (code) DO NOTHING;

CREATE TABLE IF NOT EXISTS visit_procedures (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    visit_id UUID NOT NULL REFERENCES visits(id) ON DELETE CASCADE,
    procedure_name VARCHAR(200) NOT NULL,
    tariff NUMERIC(15, 2) NOT NULL DEFAULT 0,
    qty INT NOT NULL DEFAULT 1,
    subtotal NUMERIC(15, 2) NOT NULL DEFAULT 0,
    doctor_id UUID REFERENCES users(id) ON DELETE SET NULL,
    catatan TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. Nursing Detailed Assessments (Pengkajian Keperawatan Lengkap)
CREATE TABLE IF NOT EXISTS nursing_assessments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    visit_id UUID NOT NULL UNIQUE REFERENCES visits(id) ON DELETE CASCADE,
    patient_id UUID NOT NULL REFERENCES patients(id) ON DELETE RESTRICT,
    nurse_id UUID REFERENCES users(id) ON DELETE SET NULL,
    keluhan_keperawatan TEXT,
    riwayat_penyakit_sekarang TEXT,
    riwayat_alergi TEXT,
    skala_nyeri_skor INT DEFAULT 0,
    skala_nyeri_lokasi VARCHAR(100),
    skala_nyeri_durasi VARCHAR(100),
    risiko_jatuh_tingkat VARCHAR(50) DEFAULT 'RENDAH', -- RENDAH, SEDANG, TINGGI
    skrining_gizi_turun_bb BOOLEAN DEFAULT FALSE,
    skrining_gizi_nafsu_makan BOOLEAN DEFAULT FALSE,
    masalah_keperawatan TEXT,
    rencana_keperawatan TEXT,
    evaluasi_keperawatan TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 4. Purchase Requisitions (PR) in Logistics
CREATE TABLE IF NOT EXISTS purchase_requisitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pr_number VARCHAR(50) NOT NULL UNIQUE,
    requested_by UUID REFERENCES users(id) ON DELETE SET NULL,
    request_date DATE NOT NULL DEFAULT CURRENT_DATE,
    department VARCHAR(100) NOT NULL DEFAULT 'FARMASI',
    status VARCHAR(30) NOT NULL DEFAULT 'SUBMITTED', -- DRAFT, SUBMITTED, APPROVED, REJECTED, CONVERTED_TO_PO
    total_estimated NUMERIC(15, 2) NOT NULL DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS purchase_requisition_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pr_id UUID NOT NULL REFERENCES purchase_requisitions(id) ON DELETE CASCADE,
    item_id INT NOT NULL REFERENCES inventory_items(id) ON DELETE RESTRICT,
    qty INT NOT NULL,
    unit_price NUMERIC(15, 2) NOT NULL DEFAULT 0,
    subtotal NUMERIC(15, 2) NOT NULL DEFAULT 0
);

-- 5. Goods Receipt Notes (GRN) in Logistics
CREATE TABLE IF NOT EXISTS goods_receipt_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    grn_number VARCHAR(50) NOT NULL UNIQUE,
    po_id UUID NOT NULL REFERENCES purchase_orders(id) ON DELETE RESTRICT,
    received_date DATE NOT NULL DEFAULT CURRENT_DATE,
    receiver_name VARCHAR(150) NOT NULL,
    delivery_order_number VARCHAR(100),
    status VARCHAR(30) NOT NULL DEFAULT 'RECEIVED',
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 6. Master Supplier / PBF (Pedagang Besar Farmasi)
CREATE TABLE IF NOT EXISTS suppliers (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(150) NOT NULL,
    pic_name VARCHAR(100) DEFAULT '',
    phone VARCHAR(50) DEFAULT '',
    email VARCHAR(100) DEFAULT '',
    address TEXT DEFAULT '',
    city VARCHAR(100) DEFAULT '',
    status_aktif BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed Standard PBF Suppliers
INSERT INTO suppliers (code, name, pic_name, phone, email, address, city, status_aktif)
VALUES
('PBF-KF-01', 'PT. Kimia Farma Trading & Distribution', 'Budi Santoso, S.Si, Apt', '021-3847709', 'kftd@kimiafarma.co.id', 'Jl. Veteran No. 9', 'Jakarta Pusat', true),
('PBF-ENS-02', 'PT. Enseval Putera Megatrading (Kalbe)', 'Hendra Gunawan', '021-4609042', 'cs@enseval.com', 'Kawasan Industri Pulogadung', 'Jakarta Timur', true),
('PBF-AAM-03', 'PT. Anugrah Argon Medica (AAM)', 'Dewi Lestari, S.Farm', '021-89901234', 'info@argon-group.com', 'Titan Center, Bintaro Jaya', 'Tangerang Selatan', true),
('PBF-MBS-04', 'PT. Mensa Bina Sukses (MBS)', 'Agus Priyono', '021-5818989', 'contact@mbs.co.id', 'Jl. Tomang Raya No. 21', 'Jakarta Barat', true),
('PBF-PTV-05', 'PT. Penta Valent', 'Rina Wulandari', '021-65302222', 'sales@pentavalent.co.id', 'Sunter Agung Podomoro', 'Jakarta Utara', true),
('PBF-DNR-06', 'PT. Dos Ni Roha (DNR Logistics)', 'Eko Prasetyo', '021-83788888', 'support@dnr.id', 'Jl. Tebet Raya No. 42', 'Jakarta Selatan', true),
('PBF-PPG-07', 'PT. Parit Padang Global (PPG)', 'Maya Kusuma', '021-4608888', 'care@paritpadang.co.id', 'Pulo Gadung Trade Center', 'Jakarta Timur', true),
('PBF-BSP-08', 'PT. Bina San Prima (BSP)', 'Ahmad Fauzi', '022-7561234', 'info@binasanprima.com', 'Jl. Soekarno Hatta No. 450', 'Bandung', true)
ON CONFLICT (code) DO NOTHING;

-- Add supplier_id to purchase_orders if not present
ALTER TABLE purchase_orders ADD COLUMN IF NOT EXISTS supplier_id INT REFERENCES suppliers(id) ON DELETE SET NULL;

-- 7. Kuitansi & Kasir: Promo & Kembalian details
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS promo_code VARCHAR(50);
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS promo_name VARCHAR(150);
ALTER TABLE billing_payments ADD COLUMN IF NOT EXISTS uang_diterima NUMERIC(15, 2) DEFAULT 0;
ALTER TABLE billing_payments ADD COLUMN IF NOT EXISTS kembalian NUMERIC(15, 2) DEFAULT 0;
