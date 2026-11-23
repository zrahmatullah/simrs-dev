-- ========================================================
-- MIGRATION FASE 1: ROADMAP IMPLEMENTATION
-- SIMRS-ASOY E-CLINIC & HOSPITAL SYSTEM
-- ========================================================

-- 1. JADWAL PRAKTIK & MANAJEMEN KUOTA DOKTER
CREATE TABLE IF NOT EXISTS doctor_schedules (
    id SERIAL PRIMARY KEY,
    doctor_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    poli_id INT NOT NULL REFERENCES poli_units(id) ON DELETE CASCADE,
    day_of_week INT NOT NULL, -- 0=Minggu, 1=Senin, 2=Selasa, 3=Rabu, 4=Kamis, 5=Jumat, 6=Sabtu
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    max_quota INT NOT NULL DEFAULT 30,
    status_aktif BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. CUTI DOKTER
CREATE TABLE IF NOT EXISTS doctor_leaves (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    doctor_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    reason TEXT NOT NULL,
    substitute_doctor_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. INFORMED CONSENT DIGITAL (CANVAS SIGNATURE PAD)
CREATE TABLE IF NOT EXISTS informed_consents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    visit_id UUID NOT NULL REFERENCES visits(id) ON DELETE CASCADE,
    patient_id UUID NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    doctor_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    consent_type VARCHAR(50) NOT NULL, -- PERSETUJUAN_TINDAKAN, PENOLAKAN_TINDAKAN, ANESTESI, BEDAH_MINOR
    action_name VARCHAR(255) NOT NULL,
    risks_explained TEXT NOT NULL,
    signer_name VARCHAR(150) NOT NULL,
    signer_relation VARCHAR(50) NOT NULL, -- DIRI_SENDIRI, ORANG_TUA, SUAMI_ISTRI, ANAK, KELUARGA
    patient_signature_data TEXT NOT NULL, -- Base64 data URL
    doctor_signature_data TEXT NOT NULL,  -- Base64 data URL
    qr_code_token VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 4. RETUR OBAT FARMASI & PENYESUAIAN KASIR
CREATE TABLE IF NOT EXISTS medicine_returns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nomor_retur VARCHAR(50) NOT NULL UNIQUE,
    prescription_id UUID NULL REFERENCES prescriptions(id) ON DELETE SET NULL,
    visit_id UUID NOT NULL REFERENCES visits(id) ON DELETE CASCADE,
    patient_id UUID NOT NULL REFERENCES patients(id) ON DELETE RESTRICT,
    invoice_id UUID NULL REFERENCES invoices(id) ON DELETE SET NULL,
    alasan TEXT NOT NULL,
    total_refund NUMERIC(15, 2) NOT NULL DEFAULT 0,
    status VARCHAR(30) NOT NULL DEFAULT 'COMPLETED', -- COMPLETED, CANCELLED
    created_by UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS medicine_return_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    return_id UUID NOT NULL REFERENCES medicine_returns(id) ON DELETE CASCADE,
    item_id INT NOT NULL REFERENCES inventory_items(id) ON DELETE RESTRICT,
    item_name VARCHAR(200) NOT NULL,
    qty INT NOT NULL DEFAULT 1,
    unit_price NUMERIC(15, 2) NOT NULL DEFAULT 0,
    subtotal NUMERIC(15, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 5. STOCK OPNAME (SO) PERSERIAAN
CREATE TABLE IF NOT EXISTS stock_opnames (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    opname_number VARCHAR(50) NOT NULL UNIQUE,
    opname_date DATE NOT NULL DEFAULT CURRENT_DATE,
    depo_type VARCHAR(50) NOT NULL DEFAULT 'GUDANG_UTAMA', -- GUDANG_UTAMA, DEPO_FARMASI
    status VARCHAR(30) NOT NULL DEFAULT 'COMPLETED', -- DRAFT, COMPLETED
    total_items INT NOT NULL DEFAULT 0,
    total_deviasi_nilai NUMERIC(15, 2) NOT NULL DEFAULT 0,
    notes TEXT,
    created_by UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS stock_opname_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    opname_id UUID NOT NULL REFERENCES stock_opnames(id) ON DELETE CASCADE,
    item_id INT NOT NULL REFERENCES inventory_items(id) ON DELETE RESTRICT,
    item_name VARCHAR(200) NOT NULL,
    system_qty INT NOT NULL,
    physical_qty INT NOT NULL,
    difference_qty INT NOT NULL,
    hpp_price NUMERIC(15, 2) NOT NULL DEFAULT 0,
    total_difference_value NUMERIC(15, 2) NOT NULL DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 6. BERITA ACARA PEMUSNAHAN OBAT KEDALUWARSA / RUSAK
CREATE TABLE IF NOT EXISTS medicine_disposals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    disposal_number VARCHAR(50) NOT NULL UNIQUE,
    item_id INT NOT NULL REFERENCES inventory_items(id) ON DELETE RESTRICT,
    item_name VARCHAR(200) NOT NULL,
    batch_number VARCHAR(100) NOT NULL,
    expiry_date DATE NOT NULL,
    qty INT NOT NULL,
    unit_price NUMERIC(15, 2) NOT NULL DEFAULT 0,
    total_loss NUMERIC(15, 2) NOT NULL DEFAULT 0,
    reason TEXT NOT NULL, -- KEDALUWARSA, RUSAK_FISIK, TERKONTAMINASI
    witness_name VARCHAR(150) NOT NULL,
    created_by UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- INDEXES
CREATE INDEX IF NOT EXISTS idx_doctor_schedules_doc ON doctor_schedules(doctor_id);
CREATE INDEX IF NOT EXISTS idx_doctor_schedules_poli ON doctor_schedules(poli_id);
CREATE INDEX IF NOT EXISTS idx_doctor_leaves_doc ON doctor_leaves(doctor_id);
CREATE INDEX IF NOT EXISTS idx_informed_consents_visit ON informed_consents(visit_id);
CREATE INDEX IF NOT EXISTS idx_medicine_returns_visit ON medicine_returns(visit_id);
CREATE INDEX IF NOT EXISTS idx_stock_opnames_date ON stock_opnames(opname_date);

-- SEED DATA DEFAULT JADWAL DOKTER
INSERT INTO doctor_schedules (doctor_id, poli_id, day_of_week, start_time, end_time, max_quota, status_aktif)
SELECT u.id, u.poli_id, s.day_of_week, s.start_time, s.end_time, s.max_quota, TRUE
FROM users u
CROSS JOIN (
    VALUES 
        (1, '08:00'::TIME, '12:00'::TIME, 30), -- Senin
        (2, '08:00'::TIME, '12:00'::TIME, 30), -- Selasa
        (3, '08:00'::TIME, '12:00'::TIME, 30), -- Rabu
        (4, '08:00'::TIME, '12:00'::TIME, 30), -- Kamis
        (5, '08:00'::TIME, '11:30'::TIME, 25), -- Jumat
        (6, '08:00'::TIME, '13:00'::TIME, 20)  -- Sabtu
) AS s(day_of_week, start_time, end_time, max_quota)
WHERE u.role_id = 5 AND u.poli_id IS NOT NULL
ON CONFLICT DO NOTHING;
