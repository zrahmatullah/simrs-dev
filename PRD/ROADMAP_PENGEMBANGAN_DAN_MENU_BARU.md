# PRD: ROADMAP PENGEMBANGAN & KATALOG MENU BARU (NEXT PHASE)
**Sistem Informasi Manajemen Rumah Sakit & E-Clinic — SIMRS-ASOY**
*Status: Approved for Development & Phased Implementation*

---

## 1. RINGKASAN EKSEKUTIF & TUJUAN ROADMAP

Dokumen ini adalah *Product Requirement Document* (PRD) yang mendefinisikan rincian penambahan menu baru, penyempurnaan fitur klinis eksisting, dan modernisasi arsitektur teknis pada sistem **SIMRS-ASOY**.

Tujuan utama dari roadmap ini:
1. **Kepatuhan Regulasi & Mutu Layanan:** Standardisasi integrasi rekam medis digital (Permenkes No. 24/2022), bridging SATUSEHAT Kemenkes & BPJS VClaim, serta Indeks Kepuasan Masyarakat (IKM).
2. **Otomasi Akuntansi & Proteksi Kebocoran Pendapatan:** Otomasi jasa medis (*fee dokter/perawat*), pembukuan kas kecil (*petty cash*), laporan laba-rugi (*P&L*), dan rekonsiliasi piutang asuransi.
3. **Efisiensi & *Safety* Klinis:** *Clinical Decision Support System* (CDSS) pencegah interaksi obat, *Informed Consent Digital* bertanda tangan elektronik, dan *Stock Opname* berkala.

---

## 2. MATRIKS PRIORITAS & KELAYAKAN IMPLEMENTASI (FEASIBILITY & PHASING)

Fitur diklasifikasikan ke dalam 3 Fase berdasarkan tingkat urgensi klinis dan ketergantungan pihak ketiga (*third-party dependencies*):

```
+--------------------------------------------------------------------------------------------------+
| FASE 1: IMMEDIATE / HIGH-IMPACT (Internal System & Native Implementation - Tanpa Biaya Pihak Ke-3)|
+--------------------------------------------------------------------------------------------------+
| 1. Retur Obat Pasien (Farmasi -> Balik Stok -> Kasir Otomatis)                                    |
| 2. Display Antrean TV Ruang Tunggu (Public Screen URL /display-queue, Suara Otomatis)            |
| 3. Jadwal Praktik & Cuti Dokter (Manajemen Kuota Sesi & Libur)                                   |
| 4. Informed Consent Digital (Canvas Signature Pad & Arsip PDF)                                    |
| 5. Stock Opname (SO) & Pemusnahan Obat Kedaluwarsa                                               |
| 6. Split Payment Kasir (Campuran Tunai + QRIS / EDC)                                             |
| 7. QR Code TTE Dokumen (Surat Sakit, Sehat, Rujukan, Kuitansi)                                   |
| 8. Laporan Laba Rugi (P&L) & Top 10 Penyakit ICD-10                                              |
| 9. Database Row-Level Locking (ACID Locking pemotongan stok & kasir)                              |
+--------------------------------------------------------------------------------------------------+
                                                |
                                                v
+--------------------------------------------------------------------------------------------------+
| FASE 2: ADVANCED CLINICAL & ACCREDITATION (Penyempurnaan Modul Klinis Spesialis & Akuntansi)      |
+--------------------------------------------------------------------------------------------------+
| 1. Odontogram Chart Gigi Interaktif (Poli Gigi)                                                  |
| 2. Kurva Tumbuh Kembang WHO Anak (KIA / Pediatri: Z-Score BB/U, TB/U, LK/U)                       |
| 3. Clinical Decision Support System (CDSS Alergi Obat & Dosis)                                   |
| 4. Template Paket Resep Dokter (1-Klik Input Paket Penyakit Umum)                                |
| 5. Perhitungan Otomatis Jasa Medis Dokter & Perawat (Fee Sharing)                                |
| 6. Piutang & Klaim Asuransi Swasta / BPJS                                                        |
| 7. Kas Kecil (Petty Cash) & Pengeluaran Operasional Klinik                                        |
| 8. Anjungan Pasien Mandiri (Kiosk Self Check-in)                                                 |
| 9. Survei Kepuasan & Keluhan Pasien (IKM Bintang 1-5)                                            |
| 10. Auto-FEFO (Batch Number & Expiry Date Based Dispensing)                                      |
+--------------------------------------------------------------------------------------------------+
                                                |
                                                v
+--------------------------------------------------------------------------------------------------+
| FASE 3: EXTERNAL BRIDGING & CLOUD INFRASTRUCTURE (Memerlukan Kredensial Resmi & Gateway API)      |
+--------------------------------------------------------------------------------------------------+
| 1. Bridging SATUSEHAT Kemenkes RI (HL7 FHIR REST API: Patient, Encounter, Condition, Medication) |
| 2. Bridging BPJS Kesehatan (VClaim, Antrean Online Mobile JKN, & P-Care)                         |
| 3. WhatsApp Gateway Notifications (Karcis Antrean, Reminder H-1, e-Kuitansi PDF)                 |
| 4. WebSocket / SSE Server Real-Time Push Antrean                                                 |
| 5. Automated Encrypted Daily Backup ke S3 / Cloudflare R2                                        |
+--------------------------------------------------------------------------------------------------+
```

---

## 3. SPESIFIKASI DETAIL MODUL BARU (FASE 1 & FASE 2)

### A. Modul Front Office & Antrean

#### 1. Jadwal Praktik & Cuti Dokter
* **Tujuan:** Menghindari penumpukan antrean dan pendaftaran pasien pada saat dokter berhalangan hadir.
* **Peran Akses:** `SUPERADMIN`, `CALL_CENTER`, `REGISTRATION`.
* **Fitur Utama:**
  * Penentuan hari praktik reguler (Senin–Minggu), jam mulai dan jam selesai.
  * Pembatasan kuota maksimal pasien per sesi (misal: 30 pasien/shift).
  * Fitur permohonan/pencatatan cuti dokter: saat tanggal cuti aktif, sistem secara otomatis menolak/mengunci pendaftaran ke dokter tersebut dan memberi notifikasi ke petugas pendaftaran/call center.
* **Skema Database Tambahan:**
  ```sql
  CREATE TABLE doctor_schedules (
      id SERIAL PRIMARY KEY,
      doctor_id UUID NOT NULL REFERENCES users(id),
      poli_id INT NOT NULL REFERENCES poli_units(id),
      day_of_week INT NOT NULL, -- 0=Minggu, 1=Senin.. 6=Sabtu
      start_time TIME NOT NULL,
      end_time TIME NOT NULL,
      max_quota INT NOT NULL DEFAULT 30,
      status_aktif BOOLEAN NOT NULL DEFAULT TRUE,
      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );

  CREATE TABLE doctor_leaves (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      doctor_id UUID NOT NULL REFERENCES users(id),
      start_date DATE NOT NULL,
      end_date DATE NOT NULL,
      reason TEXT NOT NULL,
      substitute_doctor_id UUID NULL REFERENCES users(id),
      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );
  ```

#### 2. Display Antrean TV (Public Screen)
* **Tujuan:** Tampilan layar lebar di ruang tunggu lobi untuk memanggil antrean dan menampilkan nomor antrean yang sedang dilayani di masing-masing poli.
* **Rute Frontend:** `/display-queue` (Didesain *clean*, fullscreen, tanpa navigasi sidebar/login, *read-only public mode*).
* **Fitur Utama:**
  * Tampilan *High-Contrast* terbaca jelas dari jarak 10 meter.
  * Kotak pemanggilan aktif: No. Antrean, Nama Pasien, Poli Tujuan, dan Dokter.
  * Grid monitoring antrean per poli (Poli Anak, Poli Umum, Gigi, dll.).
  * Web Speech synthesis bahasa Indonesia + audio chime lonceng rumah sakit otomatis saat ada pemanggilan dari dokter/perawat.
  * Ticker teks berjalan (*running text*) informasi jadwal poli & edukasi kesehatan.

#### 3. Anjungan Pasien Mandiri (Kiosk Self Check-In)
* **Tujuan:** Mengurai antrean di loket fisik pendaftaran bagi pasien yang telah melakukan reservasi online/WhatsApp.
* **Rute Frontend:** `/kiosk` (Mode layar sentuh antarmuka besar).
* **Fitur Utama:**
  * Pasien memasukkan No. RM atau scan QR Code bukti booking.
  * Konfirmasi kehadiran (*check-in*) 1 sentuhan.
  * Cetak karcis antrean poli secara otomatis.

---

### B. Modul Klinis & Rekam Medis Elektronik

#### 4. Informed Consent Digital (Persetujuan Tindakan Medis)
* **Tujuan:** Perlindungan medikolegal klinik dan pemenuhan rekam medis elektronik berstandar akreditasi.
* **Peran Akses:** `DOCTOR`, `NURSE`.
* **Fitur Utama:**
  * Pilihan jenis formulir: Persetujuan Tindakan Medis Invasif, Penolakan Tindakan, Persetujuan Anestesi/Sedasi, atau Persetujuan Tindakan Bedah Minor.
  * Komponen *HTML5 Canvas Signature Pad*: Pasien atau keluarga penanggung jawab menandatangani langsung di layar tablet/PC monitor dokter.
  * Tanda tangan digital dokter pemeriksa dan saksi tenaga kesehatan.
  * Otomatis mengompilasi lembar informed consent resmi ber-QR Code validasi ke format PDF terarsip di riwayat EMR pasien.
* **Skema Database Tambahan:**
  ```sql
  CREATE TABLE informed_consents (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      visit_id UUID NOT NULL REFERENCES visits(id),
      patient_id UUID NOT NULL REFERENCES patients(id),
      doctor_id UUID NOT NULL REFERENCES users(id),
      consent_type VARCHAR(50) NOT NULL, -- PERSETUJUAN_TINDAKAN, PENOLAKAN_TINDAKAN
      action_name VARCHAR(255) NOT NULL,
      risks_explained TEXT NOT NULL,
      signer_name VARCHAR(150) NOT NULL,
      signer_relation VARCHAR(50) NOT NULL, -- DIRI_SENDIRI, ORANG_TUA, SUAMI_ISTRI, ANAK
      patient_signature_data TEXT NOT NULL, -- Base64 PNG Canvas Signature
      doctor_signature_data TEXT NOT NULL,
      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );
  ```

#### 5. Odontogram Chart Interaktif (Poli Gigi & Mulut)
* **Tujuan:** Standarisasi pemeriksaan gigi dan mulut pada poli gigi.
* **Fitur Utama:**
  * Skema grid 32 gigi dewasa dan 20 gigi susu anak (Metode FDI Two-Digit World Dental Federation).
  * Simbol anatomi kondisi gigi: *Car (Caries/Karies)*, *Mis (Missing/Gigi Hilang)*, *Amf (Amalgam Filling)*, *Gif (GIC Filling)*, *Ext (Extraction/Pencabutan)*, *Imp (Impacted)*, *Crown*.
  * Klik langsung pada mahkota/akar gigi untuk menetapkan kondisi klinis.

#### 6. Kurva Pertumbuhan WHO Interaktif (Poli Anak & Tumbuh Kembang)
* **Tujuan:** Pemantauan status gizi dan stunting pada poli spesialis anak.
* **Fitur Utama:**
  * Mengambil data TTV historis anak: Usia (Bulan), Berat Badan (kg), Tinggi Badan (cm), dan Lingkar Kepala (cm).
  * Plotting otomatis pada grafik standar WHO Multicentre Growth Reference Study:
    - Kurva Berat Badan menurut Umur (BB/U)
    - Kurva Panjang/Tinggi Badan menurut Umur (TB/U)
    - Kurva Berat Badan menurut Panjang/Tinggi Badan (BB/TB)
    - Kurva Indeks Massa Tubuh menurut Umur (IMT/U)
  * Otomatis mengkalkulasi kategori Z-Score: Gizi Buruk, Gizi Kurang, Gizi Baik, Berisiko Gizi Lebih, Obesitas, serta Stunted / Normal.

#### 7. Clinical Decision Support System (CDSS) Pencegah Interaksi Obat
* **Tujuan:** *Patient safety* dalam penulisan resep obat.
* **Fitur Utama:**
  * *Allergy Cross-Check:* Sistem membandingkan item obat yang dipilih dokter dengan data `patients.riwayat_alergi`.
  * *Drug-Drug Interaction Alert:* Memberikan dialog peringatan apabila dokter meresepkan kombinasi obat yang kontraindikasi mayor (misal: NSAID ganda atau interaksi antikoagulan).
  * *Age/Dose Checking:* Peringatan otomatis jika obat dewasa diresepkan pada pasien usia bayi/balita tanpa penyesuaian dosis.

---

### C. Modul Farmasi, Depo & Gudang

#### 8. Retur Obat Pasien
* **Tujuan:** Menangani pembatalan atau perubahan dosis obat setelah resep diverifikasi farmasi.
* **Peran Akses:** `PHARMACY`, `CASHIER`.
* **Alur Bisnis Retur:**
  1. Pasien/Dokter mengajukan pembatalan item obat.
  2. Petugas Farmasi membuka formulir **Retur Resep**, memilih item dan jumlah yang dikembalikan, serta mencatat alasan retur (contoh: alergi, salah dosis, pasien menolak).
  3. Sistem secara otomatis:
     - Mengembalikan (*re-stock*) saldo kuantitas obat ke Depo/Gudang.
     - Menghasilkan entri mutasi stok tipe `IN_RETUR`.
     - Mengurangi tagihan pada invoice kasir atau menerbitkan *Credit Note / Refund Kasir* jika invoice sudah lunas.

#### 9. Stock Opname (SO) & Pemusnahan Obat
* **Tujuan:** Validasi fisik berkala persediaan obat dan penghapusan obat kedaluwarsa.
* **Peran Akses:** `LOGISTICS`, `PHARMACY`, `SUPERADMIN`.
* **Fitur Utama:**
  * Mode *Freeze Mutasi*: Mengunci pergerakan stok saat pencatatan fisik berlangsung.
  * Form input selisih fisik vs sistem: Otomatis mencatat nilai deviasi (rugi/untung stok).
  * Berita Acara Pemusnahan Obat: Pencatatan obat rusak/kedaluwarsa lengkap dengan No. Batch, tanggal kedaluwarsa, nilai kerugian HPP, saksi apoteker, dan unduh PDF resmi.

---

### D. Modul Keuangan, Kasir & Akuntansi Klinik

#### 10. Split Payment (Metode Pembayaran Ganda) di Kasir
* **Tujuan:** Fleksibilitas pembayaran pasien yang membayar dengan kombinasi metode bayar.
* **Fitur Utama:**
  * Dalam satu invoice, kasir dapat memecah nominal pembayaran:
    - Contoh: Total tagihan Rp 450.000 $\rightarrow$ Rp 300.000 dibayar via QRIS / Asuransi, Rp 150.000 dibayar Tunai.
  * Menghasilkan multi-rekaman di tabel `billing_payments` dengan referensi nomor transaksi masing-masing.

#### 11. Perhitungan Jasa Medis Dokter & Perawat (Fee Sharing)
* **Tujuan:** Menghitung hak komisi/jasa pelayanan tenaga medis secara adil, transparan, dan otomatis.
* **Peran Akses:** `SUPERADMIN`, `DOCTOR`, `NURSE`.
* **Fitur Utama:**
  * Pengaturan persentase bagi hasil jasa konsultasi & tindakan poli (misal: 70% Dokter, 30% Klinik).
  * Perhitungan hanya dieksekusi dari tindakan & konsultasi yang invoice-nya berstatus **LUNAS**.
  * Laporan rincian jasa medis per dokter periode bulanan siap unduh Excel/PDF untuk slip gaji dokter.

#### 12. Kas Kecil (Petty Cash) & Biaya Operasional
* **Tujuan:** Pencatatan arus kas keluar harian klinik non-medis (ATK, konsumsi, kebersihan, listrik, maintenance).
* **Fitur Utama:**
  * Pengajuan & persetujuan kas kecil.
  * Kategori beban pengeluaran operasional.
  * Integrasi otomatis ke Laporan Laba Rugi Bersih (*Net Profit & Loss*).

---

### E. Modul Integrasi Eksternal (Fase 3)

#### 13. Bridging SATUSEHAT Kemenkes RI (HL7 FHIR API)
* **Prasyarat:** Organisasi Faskes ID, Client ID, dan Client Secret resmi dari DTO Kemenkes.
* **Resource FHIR yang Di-bridging:**
  - `Encounter`: Kunjungan rawat jalan & waktu pelayanan poli.
  - `Condition`: Kode diagnosis ICD-10 klinis.
  - `Observation`: Hasil pemeriksaan tanda-tanda vital (TTV) & hasil laboratorium.
  - `MedicationRequest` & `MedicationDispense`: Resep dan penyerahan obat farmasi.
  - `Procedure`: Tindakan medis berbasis ICD-9-CM.
* **Monitor Log Bridging:** Dashboard status pengiriman payload JSON, waktu kirim, HTTP response code (200/201 vs 400/500), serta tombol *Resend Failure*.

#### 14. Bridging BPJS Kesehatan (VClaim & Antrean Online)
* **Prasyarat:** Consumer ID, Consumer Secret, User Key, dan Kode PPK Faskes dari BPJS.
* **Fitur Utama:**
  - Validasi Nomor Rujukan & Surat Eligibilitas Peserta (SEP) rawat jalan.
  - Sinkronisasi status antrean poli ke Mobile JKN Kemenkes/BPJS (*Waktu Tunggu Admisi, Waktu Layan Dokter, Waktu Tunggu Farmasi*).

#### 15. WhatsApp Gateway Notifications
* **Prasyarat:** API Provider WhatsApp (Fonnte / Wablas / Twilio Business API).
* **Fitur Otomatisasi:**
  - Notifikasi karcis antrean & link pantau nomor antrean saat pasien selesai registrasi.
  - Pengingat jadwal kontrol (H-1) bagi pasien yang memiliki jadwal kontrol ulang dokter.
  - Pengiriman file PDF kuitansi resmi dan e-Resume Medis setelah pelunasan kasir.

---

## 4. PENYEMPURNAAN ARSITEKTUR TEKNIS & KEAMANAN

### 1. Database Row-Level Locking (`SELECT ... FOR UPDATE`)
Untuk mencegah *race condition* (stok minus saat dua resep menelaah obat yang sama secara bersamaan):
```go
// Transaksi ACID dengan Row-Level Locking
tx, err := r.db.Beginx()
if err != nil { return err }
defer tx.Rollback()

// Kunci baris item persediaan selama proses evaluasi stok
query := `SELECT total_stock, depo_stock FROM inventory_items WHERE id = $1 FOR UPDATE`
var currentStock domain.InventoryItem
if err := tx.Get(&currentStock, query, itemID); err != nil {
    return err
}

if currentStock.depo_stock < qty {
    return errors.New("stok persediaan di depo tidak mencukupi")
}

// Lakukan pemotongan dan catat mutasi dalam transaksi yang sama
...
return tx.Commit()
```

### 2. QR Code Verifikasi Dokumen TTE (Tanda Tangan Elektronik)
Setiap dokumen keluar (Surat Sakit, Sehat, Rujukan, Kuitansi, dan Resume Medis) disematkan QR Code dinamis:
* URL Validasi Publik: `https://simrs.klinik.com/verify-doc?code=DOC-YYYY-UUID`
* Saat discan oleh pihak luar (kantor/asuransi), browser membuka halaman verifikasi keabsahan dokumen berstempel resmi rumah sakit, nama dokter penerbit, dan waktu penerbitan.

### 3. Arsitektur Real-Time Antrean (Server-Sent Events / SSE)
Mengganti *polling* AJAX periodik dengan Server-Sent Events (SSE) berlatar belakang Go HTTP Channel untuk menghemat konsumsi *bandwidth* server hingga 85% dan memastikan perpindahan antrean di Display TV dan meja dokter sinkron secara instan.

---

## 5. REKOMENDASI TAHAPAN PENGERJAAN BERIKUTNYA

Direkomendasikan mengeksekusi **Fase 1** terlebih dahulu:
1. **Fitur Retur Obat Farmasi & Penyesuaian Kasir** (Menyempurnakan alur farmasi & kasir).
2. **Display Antrean TV Ruang Tunggu (`/display-queue`)** (Meningkatkan pengalaman ruang tunggu pasien).
3. **Informed Consent Digital Canvas Signature** (Penguatan aspek medikolegal tindakan medis).
4. **Jadwal Praktik & Manajemen Kuota Dokter** (Menertibkan kuota pelayanan harian).
5. **Split Payment di Kasir & Laporan Laba Rugi P&L** (Menyempurnakan tata kelola akuntansi).
