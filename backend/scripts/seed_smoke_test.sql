-- ========================================================
-- SIMRS-ASOY: Smoke Test Seed Data (2 Patients Complete Flow)
-- ========================================================

BEGIN;

-- Master ICD-10 additions if missing
INSERT INTO icd10_codes (code, name_id, name_en, status_aktif)
VALUES 
    ('I10', 'Hipertensi Esensial (Primer)', 'Essential (primary) hypertension', true),
    ('J02.9', 'Faringitis Akut, Tidak Spesifik', 'Acute pharyngitis, unspecified', true)
ON CONFLICT (code) DO NOTHING;

-- 1. PASIEN 1: Ananda Rizky Pratama (Anak, Poli Anak, Siap Diperiksa Dokter)
INSERT INTO patients (id, no_rm, nik, nama_pasien, jenis_kelamin, tanggal_lahir, no_hp, alamat, golongan_darah, riwayat_alergi, created_at, updated_at)
VALUES (
    'b0000000-0000-0000-0000-000000000004',
    'RM-2026-0004',
    '3276012304190001',
    'Ananda Rizky Pratama',
    'L',
    '2019-04-23',
    '081298765432',
    'Jl. Melati Indah No. 45, RT 03/RW 04, Jakarta Selatan',
    'O+',
    'Amoxicillin (Urtikaria/Ruam Merah), Udang',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO UPDATE 
SET nama_pasien = EXCLUDED.nama_pasien, riwayat_alergi = EXCLUDED.riwayat_alergi;

-- Visit Pasien 1 (Status: MENUNGGU_DOKTER)
INSERT INTO visits (id, no_rawat, no_antrian, patient_id, doctor_id, poli_id, guarantor_id, visit_date, keluhan_utama, status, created_at, updated_at)
VALUES (
    'c0000000-0000-0000-0000-000000000005',
    'RAWAT-20261122-005',
    'A-005',
    'b0000000-0000-0000-0000-000000000004',
    'a0000000-0000-0000-0000-000000000005', -- dr. Aris Munandar, Sp.A
    2, -- Poli Anak
    1, -- UMUM
    CURRENT_DATE,
    'Demam tinggi 3 hari, batuk berdahak, pilek encer, nafsu makan menurun.',
    'MENUNGGU_DOKTER',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO UPDATE 
SET status = 'MENUNGGU_DOKTER', no_antrian = 'A-005';

-- TTV Perawat Pasien 1
INSERT INTO vital_signs (id, visit_id, patient_id, nurse_id, td_sistole, td_diastole, hr, rr, temp, spo2, bb, tb, skala_nyeri, kesadaran, triase_level, catatan_keperawatan, created_at, updated_at)
VALUES (
    'd0000000-0000-0000-0000-000000000005',
    'c0000000-0000-0000-0000-000000000005',
    'b0000000-0000-0000-0000-000000000004',
    'a0000000-0000-0000-0000-000000000004', -- Ns. Dian Anggraini
    100, 70, 98.0, 22.0, 38.4, 99.0, 21.5, 116.0, 2, 'Compos Mentis', 'KUNING',
    'Anak tampak rewel, akral hangat, riwayat alergi amoxicillin (+). Sudah diberikan kompres hangat axilla.',
    NOW(), NOW()
)
ON CONFLICT (id) DO UPDATE 
SET temp = 38.4, triase_level = 'KUNING';

-- Asesmen Keperawatan Pasien 1
INSERT INTO nursing_assessments (id, visit_id, patient_id, nurse_id, keluhan_keperawatan, riwayat_penyakit_sekarang, riwayat_alergi, skala_nyeri_skor, skala_nyeri_lokasi, skala_nyeri_durasi, risiko_jatuh_tingkat, skrining_gizi_turun_bb, skrining_gizi_nafsu_makan, masalah_keperawatan, rencana_keperawatan, created_at, updated_at)
VALUES (
    'e0000000-0000-0000-0000-000000000005',
    'c0000000-0000-0000-0000-000000000005',
    'b0000000-0000-0000-0000-000000000004',
    'a0000000-0000-0000-0000-000000000004',
    'Demam 3 hari naik turun, batuk pilek berdahak',
    'Demam dirasakan sejak 3 hari lalu terutama sore-malam hari, batuk pilek sekret kental kekuningan.',
    'Alergi Amoxicillin (gatal dan bentol kemerahan)',
    2, 'Tenggorokan saat menelan', 'Intermiten saat batuk',
    'RENDAH',
    false, true,
    '1. Hipertermia berhubungan dengan proses penyakit inflamasi saluran napas.
2. Bersihan jalan napas tidak efektif berhubungan dengan hipersekresi jalan napas.',
    '1. Pantau suhu tubuh dan tanda vital setiap 2-4 jam.
2. Anjurkan orang tua memberikan hidrasi air hangat sering.
3. Lakukan kompres hangat pada dahi dan lipat ketiak.
4. Kolaborasi dokter poli anak untuk terapi antipiretik dan mukolitik.',
    NOW(), NOW()
)
ON CONFLICT (visit_id) DO UPDATE 
SET keluhan_keperawatan = EXCLUDED.keluhan_keperawatan;

-- Invoice Awal Pasien 1 (Biaya Registrasi Rp 100.000)
DELETE FROM invoice_items WHERE invoice_id IN (SELECT id FROM invoices WHERE visit_id = 'c0000000-0000-0000-0000-000000000005');
DELETE FROM invoices WHERE visit_id = 'c0000000-0000-0000-0000-000000000005';

INSERT INTO invoices (id, nomor_invoice, visit_id, patient_id, subtotal, diskon, pajak, total_tagihan, total_terbayar, sisa_tagihan, status, jenis_penjamin, catatan, created_at, updated_at)
VALUES (
    'f0000000-0000-0000-0000-000000000005',
    'INV-20261122-005',
    'c0000000-0000-0000-0000-000000000005',
    'b0000000-0000-0000-0000-000000000004',
    100000.00, 0, 0, 100000.00, 0, 100000.00, 'BELUM_BAYAR', 'UMUM',
    'Biaya Administrasi Pendaftaran Otomatis Pasien Poli Anak',
    NOW(), NOW()
);

INSERT INTO invoice_items (id, invoice_id, item_type, reference_id, item_name, qty, unit_price, subtotal, created_at)
VALUES (
    'f1000000-0000-0000-0000-000000000005',
    'f0000000-0000-0000-0000-000000000005',
    'ADMINISTRASI',
    'c0000000-0000-0000-0000-000000000005',
    'Biaya Administrasi & Registrasi Rawat Jalan',
    1, 100000.00, 100000.00, NOW()
);


-- ========================================================
-- 2. PASIEN 2: Siti Nurhaliza (Dewasa, Selesai Dokter & Farmasi, Siap Bayar di Kasir)
-- ========================================================
INSERT INTO patients (id, no_rm, nik, nama_pasien, jenis_kelamin, tanggal_lahir, no_hp, alamat, golongan_darah, riwayat_alergi, created_at, updated_at)
VALUES (
    'b0000000-0000-0000-0000-000000000006',
    'RM-2026-0005',
    '3276015508880002',
    'Siti Nurhaliza',
    'P',
    '1988-08-15',
    '081387654321',
    'Jl. Cempaka Putih Barat No. 12, RT 02/RW 01, Jakarta Pusat',
    'B+',
    'Tidak Ada Alergi Obat (NKDA)',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO UPDATE 
SET nama_pasien = EXCLUDED.nama_pasien;

-- Visit Pasien 2 (Status: KASIR)
INSERT INTO visits (id, no_rawat, no_antrian, patient_id, doctor_id, poli_id, guarantor_id, visit_date, keluhan_utama, status, created_at, updated_at)
VALUES (
    'c0000000-0000-0000-0000-000000000006',
    'RAWAT-20261122-006',
    'U-006',
    'b0000000-0000-0000-0000-000000000006',
    'a0000000-0000-0000-0000-000000000005', -- dr. Aris Munandar
    1, -- Poli Umum
    1, -- UMUM
    CURRENT_DATE,
    'Pusing berputar, tengkuk terasa tegang sejak 2 hari, riwayat tensi tinggi.',
    'KASIR',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO UPDATE 
SET status = 'KASIR', no_antrian = 'U-006';

-- TTV Pasien 2
INSERT INTO vital_signs (id, visit_id, patient_id, nurse_id, td_sistole, td_diastole, hr, rr, temp, spo2, bb, tb, skala_nyeri, kesadaran, triase_level, catatan_keperawatan, created_at, updated_at)
VALUES (
    'd0000000-0000-0000-0000-000000000006',
    'c0000000-0000-0000-0000-000000000006',
    'b0000000-0000-0000-0000-000000000006',
    'a0000000-0000-0000-0000-000000000004',
    150, 95, 84.0, 18.0, 36.7, 99.0, 64.0, 158.0, 1, 'Compos Mentis', 'HIJAU',
    'Pasien mengeluh tengkuk kaku, tensi 150/95 mmHg, tidak ada keluhan sesak napas.',
    NOW(), NOW()
)
ON CONFLICT (id) DO UPDATE 
SET td_sistole = 150, td_diastole = 95;

-- Asesmen Keperawatan Pasien 2
INSERT INTO nursing_assessments (id, visit_id, patient_id, nurse_id, keluhan_keperawatan, riwayat_penyakit_sekarang, riwayat_alergi, skala_nyeri_skor, skala_nyeri_lokasi, skala_nyeri_durasi, risiko_jatuh_tingkat, skrining_gizi_turun_bb, skrining_gizi_nafsu_makan, masalah_keperawatan, rencana_keperawatan, created_at, updated_at)
VALUES (
    'e0000000-0000-0000-0000-000000000006',
    'c0000000-0000-0000-0000-000000000006',
    'b0000000-0000-0000-0000-000000000006',
    'a0000000-0000-0000-0000-000000000004',
    'Sakit kepala bagian belakang & leher kaku',
    'Keluhan dirasakan sejak 2 hari memberat saat beraktivitas padat, tidur kurang nyenyak.',
    'Tidak ada riwayat alergi obat maupun makanan',
    1, 'Tengkuk leher belakang', 'Hilang timbul saat lelah',
    'RENDAH',
    false, false,
    '1. Risiko perfusi serebral tidak efektif berhubungan dengan hipertensi.
2. Defisit pengetahuan manajemen hipertensi berhubungan dengan kurang terpapar informasi.',
    '1. Edukasi istirahat cukup dan pembatasan asupan natrium/garam dapur.
2. Pantau tekanan darah berkala.
3. Kolaborasi dokter untuk pemberian terapi antihipertensi oral.',
    NOW(), NOW()
)
ON CONFLICT (visit_id) DO UPDATE 
SET keluhan_keperawatan = EXCLUDED.keluhan_keperawatan;

-- Medical Record (CPPT Dokter) Pasien 2
DELETE FROM medical_records WHERE visit_id = 'c0000000-0000-0000-0000-000000000006';
INSERT INTO medical_records (id, visit_id, patient_id, doctor_id, subjektif, objektif, asesmen, planning, icd10_code, icd9_code, diagnosis_deskripsi, prosedur_deskripsi, status, is_locked, created_at, updated_at)
VALUES (
    'a1000000-0000-0000-0000-000000000006',
    'c0000000-0000-0000-0000-000000000006',
    'b0000000-0000-0000-0000-000000000006',
    'a0000000-0000-0000-0000-000000000005',
    'Pasien mengeluh pusing dan kaku tengkuk sejak 2 hari lalu. Riwayat tensi tinggi namun jarang kontrol minum obat.',
    'KU baik, CM. TD 150/95 mmHg, Nadi 84x/m regular, Suhu 36.7C, SpO2 99%. Mata: konjungtiva tidak anemis. Cor: BJ I-II murni regular, murmur (-). Pulmo: VBS normal kiri=kanan. Ekstremitas: akral hangat, edema (-).',
    'Hipertensi Primer / Stage 1 (ICD-10: I10) + Cephalea Tension-Type',
    '1. Paracetamol 500mg 3x1 tab bila sakit kepala.
2. Ibuprofen tablet 400mg 2x1 tab sesudah makan.
3. Edukasi modifikasi gaya hidup, kurangi makanan asin/garam berlebih, cukup tidur.
4. Kontrol ulang tensi 2 minggu lagi di poli.',
    'I10',
    '89.07',
    'Essential (primary) hypertension',
    'Consultation, physical exam and therapy evaluation',
    'FINAL',
    true,
    NOW(),
    NOW()
);

-- Tindakan Pasien 2
DELETE FROM visit_procedures WHERE visit_id = 'c0000000-0000-0000-0000-000000000006';
INSERT INTO visit_procedures (id, visit_id, procedure_name, tariff, qty, subtotal, doctor_id, catatan, created_at)
VALUES (
    'a2000000-0000-0000-0000-000000000006',
    'c0000000-0000-0000-0000-000000000006',
    'Pemeriksaan Fisik Khusus & Konsultasi Spesialis',
    150000.00,
    1,
    150000.00,
    'a0000000-0000-0000-0000-000000000005',
    'Pemeriksaan komprehensif kardiovaskular & neurologis dasar',
    NOW()
);

-- Resep Obat Pasien 2 (Sudah diverifikasi apotek: status SELESAI agar tidak memblokir kasir)
DELETE FROM prescription_items WHERE prescription_id IN (SELECT id FROM prescriptions WHERE visit_id = 'c0000000-0000-0000-0000-000000000006');
DELETE FROM prescriptions WHERE visit_id = 'c0000000-0000-0000-0000-000000000006';

INSERT INTO prescriptions (id, nomor_resep, visit_id, patient_id, doctor_id, status, catatan_dokter, catatan_telaah, is_verified, verified_by, verified_at, dispensed_by, dispensed_at, created_at, updated_at)
VALUES (
    'a3000000-0000-0000-0000-000000000006',
    'RSP-20261122-006',
    'c0000000-0000-0000-0000-000000000006',
    'b0000000-0000-0000-0000-000000000006',
    'a0000000-0000-0000-0000-000000000005',
    'SELESAI',
    'Obat diminum teratur sesudah makan',
    'Telaah resep farmasi klinis: tepat pasien, tepat obat, tepat dosis. Aman diberikan.',
    true,
    'a0000000-0000-0000-0000-000000000009', -- verified by farmasi
    NOW(),
    'a0000000-0000-0000-0000-000000000009',
    NOW(),
    NOW(),
    NOW()
);

INSERT INTO prescription_items (id, prescription_id, item_id, signa, aturan_pakai, jumlah, harga_satuan, subtotal, is_racikan, created_at)
VALUES 
(
    'a4000000-0000-0000-0000-000000000001',
    'a3000000-0000-0000-0000-000000000006',
    6, -- Ibuprofen Tablet 400mg
    '2 x 1 tablet',
    'Sesudah Makan',
    2,
    8500.00,
    17000.00,
    false,
    NOW()
),
(
    'a4000000-0000-0000-0000-000000000002',
    'a3000000-0000-0000-0000-000000000006',
    1, -- Paracetamol Sirup 120mg/5ml
    '3 x 1 sendok takar',
    'Bila Perlu / Demam',
    1,
    15000.00,
    15000.00,
    false,
    NOW()
);

-- Tagihan Kasir Pasien 2 (Status: BELUM_BAYAR, Total: 100k admin + 150k tindakan + 32k obat = 282k)
DELETE FROM invoice_items WHERE invoice_id IN (SELECT id FROM invoices WHERE visit_id = 'c0000000-0000-0000-0000-000000000006');
DELETE FROM invoices WHERE visit_id = 'c0000000-0000-0000-0000-000000000006';

INSERT INTO invoices (id, nomor_invoice, visit_id, patient_id, subtotal, diskon, pajak, total_tagihan, total_terbayar, sisa_tagihan, status, jenis_penjamin, catatan, created_at, updated_at)
VALUES (
    'f0000000-0000-0000-0000-000000000006',
    'INV-20261122-006',
    'c0000000-0000-0000-0000-000000000006',
    'b0000000-0000-0000-0000-000000000006',
    282000.00, 0, 0, 282000.00, 0, 282000.00, 'BELUM_BAYAR', 'UMUM',
    'Tagihan Lengkap Pasien Poli Umum (Admin + Konsultasi Dokter + Resep Obat Farmasi)',
    NOW(), NOW()
);

INSERT INTO invoice_items (id, invoice_id, item_type, reference_id, item_name, qty, unit_price, subtotal, created_at)
VALUES 
(
    'f1000000-0000-0000-0000-000000000006',
    'f0000000-0000-0000-0000-000000000006',
    'ADMINISTRASI',
    'c0000000-0000-0000-0000-000000000006',
    'Biaya Administrasi & Registrasi Rawat Jalan',
    1, 100000.00, 100000.00, NOW()
),
(
    'f1000000-0000-0000-0000-000000000007',
    'f0000000-0000-0000-0000-000000000006',
    'TINDAKAN',
    'a2000000-0000-0000-0000-000000000006',
    'Pemeriksaan Fisik Khusus & Konsultasi Spesialis',
    1, 150000.00, 150000.00, NOW()
),
(
    'f1000000-0000-0000-0000-000000000008',
    'f0000000-0000-0000-0000-000000000006',
    'OBAT',
    'a3000000-0000-0000-0000-000000000006',
    'Resep Farmasi: Ibuprofen Tab & Paracetamol Sirup',
    1, 32000.00, 32000.00, NOW()
);

COMMIT;
