-- ========================================================
-- SIMRS-ASOY: Migration for Referral Hospitals (Faskes) & Specialists
-- ========================================================

CREATE TABLE IF NOT EXISTS master_faskes_rujukan (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(200) NOT NULL,
    tipe VARCHAR(50) NOT NULL DEFAULT 'RSUD', -- RSUP, RSUD, RS Swasta, Klinik Utama, Puskesmas
    alamat TEXT,
    kota VARCHAR(100),
    no_telp VARCHAR(50),
    status_aktif BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS master_spesialis (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    nama_dokter VARCHAR(200) NOT NULL,
    spesialisasi VARCHAR(100) NOT NULL,
    rumah_sakit VARCHAR(200),
    status_aktif BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed Initial Faskes / RS Rujukan
INSERT INTO master_faskes_rujukan (code, name, tipe, alamat, kota, no_telp, status_aktif)
VALUES
    ('RS-001', 'RSUP Dr. Sardjito', 'RSUP', 'Jl. Kesehatan No. 1, Sendangadi, Mlati', 'Yogyakarta', '0274-587333', true),
    ('RS-002', 'RSUP Dr. Kariadi', 'RSUP', 'Jl. Dr. Sutomo No. 16, Randusari', 'Semarang', '024-8413476', true),
    ('RS-003', 'RSUP H. Adam Malik', 'RSUP', 'Jl. Bunga Lau No. 17, Medan Tuntungan', 'Medan', '061-8360143', true),
    ('RS-004', 'RSUP Fatmawati', 'RSUP', 'Jl. RS Fatmawati Raya No. 4, Cilandak', 'Jakarta Selatan', '021-7660552', true),
    ('RS-005', 'RSUD Pasar Minggu', 'RSUD', 'Jl. TB Simatupang No. 1, Ragunan', 'Jakarta Selatan', '021-29555555', true),
    ('RS-006', 'RS Jantung & Pembuluh Darah Harapan Kita', 'RS Khusus', 'Jl. Letjen S. Parman Kav. 87', 'Jakarta Barat', '021-5684085', true),
    ('RS-007', 'RS Kanker Dharmais', 'RS Khusus', 'Jl. Letjen S. Parman Kav. 84-86', 'Jakarta Barat', '021-5681570', true)
ON CONFLICT (code) DO NOTHING;

-- Seed Initial Dokter Spesialis & Subspesialis
INSERT INTO master_spesialis (code, nama_dokter, spesialisasi, rumah_sakit, status_aktif)
VALUES
    ('DR-SPA-01', 'dr. Bambang Soepardi, Sp.A(K)', 'Spesialis Anak (Konsultan Tumbuh Kembang)', 'RSUP Dr. Sardjito', true),
    ('DR-SPA-02', 'dr. Ratna Kartika, Sp.A(K)', 'Spesialis Anak (Konsultan Kardiologi Anak)', 'RSUP Dr. Kariadi', true),
    ('DR-SPJP-01', 'dr. Irfan Maulana, Sp.JP(K), FIHA', 'Spesialis Jantung & Pembuluh Darah', 'RS Jantung Harapan Kita', true),
    ('DR-SPPD-01', 'dr. Rahmat Hidayat, Sp.PD-KGEH', 'Spesialis Penyakit Dalam (Gastroenterohepatologi)', 'RSUP Fatmawati', true),
    ('DR-SPB-01', 'dr. Dimas Prasetyo, Sp.B, FICS', 'Spesialis Bedah Umum & Onkologi', 'RS Kanker Dharmais', true),
    ('DR-SPS-01', 'dr. Farida Rahmawati, Sp.N', 'Spesialis Neurologi / Saraf', 'RSUD Pasar Minggu', true),
    ('DR-SPO-01', 'dr. Suryo Wibowo, Sp.OT', 'Spesialis Orthopedi & Traumatologi', 'RSUP Fatmawati', true)
ON CONFLICT (code) DO NOTHING;
