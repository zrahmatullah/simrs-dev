# STANDAR DESAIN & SPESIFIKASI FRONTEND REKAM MEDIS ELEKTRONIK (RME / EMR)
**Sistem Informasi Manajemen Rumah Sakit (SIMRS)**

Dokumen ini menjadi acuan baku (*Single Source of Truth*) untuk pengembangan tampilan antarmuka (frontend) modul Rekam Medis Elektronik (RME / EMR) pada SIMRS, mengacu pada standar visual *Clinical Cockpit* terpadu.

---

## 1. Filosofi & Prinsip Desain UI/UX Klinis

1. **Glanceable & High-Efficiency (Dapat dipindai cepat)**: Dokter hanya memiliki waktu rata-rata 3–5 menit saat memeriksa pasien poli rawat jalan. Informasi vital (identitas, alergi, TTV) harus terlihat seketika tanpa scrolling berlebih.
2. **Clinical Safety First (Keselamatan Pasien)**: 
   - Banner identitas pasien selalu terpampang (*sticky* / di posisi paling atas).
   - Tanda bahaya/alergi (*Alert Pill*) menggunakan kontras warna merah/oranye tegas untuk mencegah insiden malpraktik atau adverse drug event.
3. **Structured SOAP Workflow**: Alur pemeriksaan mengikuti metodologi standar medis:
   - **S (Subjektif)**: Keluhan utama, RPS, RPD.
   - **O (Objektif)**: Pemeriksaan fisik & data numerik TTV.
   - **A (Asesmen)**: Diagnosis berbasis ICD-10 dengan input cepat & chip.
   - **P (Planning)**: Tindakan, terapi, E-Resep, serta penunjang (Laboratorium & Radiologi).

---

## 2. Design Tokens & Color Palette

### 2.1 Skema Warna Baku

| Token Name | Hex Code | Tailwind Equivalent | Penggunaan |
| :--- | :--- | :--- | :--- |
| **Brand Primary (Dark Teal)** | `#0D3344` / `#0A2533` | `bg-[#0D3344]` | Latar belakang sidebar navigasi utama |
| **Brand Active Teal** | `#164E63` / `#155E75` | `bg-cyan-900 / bg-teal-800` | State menu aktif pada sidebar |
| **Patient Banner Light Sky** | `#EBF5FA` / `#E0F2FE` | `bg-sky-50 / bg-sky-100` | Container banner data pasien |
| **Accent Medical Cyan** | `#0284C7` / `#0D9488` | `text-sky-600 / text-teal-600`| Tombol aksi, border aktif, tab aktif, icon |
| **Status Stable (Green)** | `#10B981` | `text-emerald-600 bg-emerald-50`| Status pasien stabil |
| **Warning / Allergy (Red)** | `#EF4444` | `text-red-600 bg-red-50` | Label dan peringatan alergi obat/makanan |
| **Card Background** | `#FFFFFF` | `bg-white` | Kartu putih form asesmen dan widget |
| **App Canvas Background** | `#F3F6F9` / `#F8FAFC` | `bg-slate-100` | Latar belakang halaman aplikasi |
| **Text Heading** | `#1E293B` | `text-slate-800` | Judul modul, nama pasien, nilai TTV |
| **Text Secondary/Muted** | `#64748B` | `text-slate-500` | Label form, satuan ukuran, deskripsi |
| **Border Neutral** | `#E2E8F0` | `border-slate-200` | Garis batas card, input, dan divider |

### 2.2 Tipografi
- **Primary Font Family**: Inter, Plus Jakarta Sans, atau sistem font modern sans-serif (`font-sans`).
- **Skala Ukuran**:
  - Judul Halaman: `text-lg font-bold text-slate-800` (18px)
  - Judul Section / Card: `text-sm / text-base font-semibold text-slate-800`
  - Nama Pasien: `text-base font-bold text-slate-900`
  - Body Text / Input Form: `text-sm text-slate-700` (14px)
  - Nilai Metrik TTV: `text-lg font-bold text-slate-800` (18–20px)
  - Label Form & Satuan TTV: `text-xs font-medium text-slate-500` (11–12px)

---

## 3. Struktur Layout & Grid

Tampilan menggunakan format **Full-Height Clinical Dashboard** dengan pembagian sebagai berikut:

```
+---------------------------------------------------------------------------------------------------+
|  SIDEBAR (Dark Teal)  |  TOPBAR: Judul Modul, Help, Notifikasi, Avatar Dokter                     |
|  - Logo SIMRS         |---------------------------------------------------------------------------|
|  - Dashboard          |  BANNER IDENTITAS PASIEN (Light Sky Blue Card)                            |
|  - Jadwal             |  [Foto] [Nama, No.RM, Umur/JK, Jaminan, Alergi]  [Cetak] [Riwayat] [🟢]   |
|  - Pasien (RME) [Aktif]|---------------------------------------------------------------------------|
|  - Rawat Jalan        |  GRID UTAMA (2 Kolom)                                                     |
|  - Rawat Inap         |  +-----------------------------+  +------------------------------------+  |
|  - E-Resep            |  | KOLOM KIRI (60-65%):        |  | KOLOM KANAN (35-40%):              |  |
|  - Laboratorium       |  | FORM ASESMEN SOAP           |  | 1. KARTU TANDA-TANDA VITAL (5 Box) |  |
|  - Radiologi          |  | - Subjektif (S)             |  |    [TD] [HR] [Temp] [RR] [SpO2]    |  |
|  - Tagihan            |  | - Objektif (O)              |  |------------------------------------|  |
|  - Pengaturan         |  | - Asesmen (A) + ICD-10 tags |  | 2. TAB E-RESEP & ORDER LAB/RAD     |  |
|                       |  | - Planning (P)              |  |    [E-Resep] | [Lab & Radiologi]   |  |
|                       |  +-----------------------------+  +------------------------------------+  |
+---------------------------------------------------------------------------------------------------+
```

---

## 4. Spesifikasi Detil Komponen

### 4.1 Sidebar Navigasi
- Lebar: Fixed `w-56` s/d `w-64` (240px).
- Warna Latar: Dark Teal `#0D3344`.
- Brand Logo: Ikon palang medis hijau/cyan terang + Teks "SIMRS" font-bold 18px warna putih.
- Menu Items:
  - Default: Teks & Icon `#94A3B8` (Slate-400), hover `text-white bg-white/10`.
  - Active: Background `#164E63` / `#194E62`, warna teks putih `#FFFFFF`, rounded pill `rounded-xl`, penanda aktif yang jelas.

### 4.2 Topbar Header
- Tinggi: Fixed `h-14` s/d `h-16` (56–64px), latar putih dengan border bawah tipis `border-slate-200`.
- Kiri: Judul Halaman `Rekam Medis Elektronik (RME)` font-bold 18px.
- Kanan:
  - Tombol Help (ikon tanda tanya lingkaran).
  - Tombol Notifikasi (ikon lonceng dengan *badge* merah angka unread).
  - Profil User / Dokter: Foto avatar bulat + Nama lengkap tenaga medis (contoh: *Azhar As Rahmatulloh*).

### 4.3 Banner Identitas Pasien (*Patient Context Card*)
- Latar Belakang: Soft Sky Blue gradient / solid `#EBF5FA` dengan border `border-sky-200` dan radius `rounded-2xl`.
- Layout: Flex row terdistribusi (Kiri: Metadata Pasien, Kanan: Action & Status).
- Metadata Pasien:
  - **Foto Pasien**: Ukuran `w-14 h-14` atau `w-16 h-16`, rounded rectangle (`rounded-xl`), object-cover.
  - **Nama Pasien**: Header bold 16px (contoh: *Azhar As Rahmatulloh*).
  - **No. RM**: Monospace font-semibold (contoh: *00-12-34-56*).
  - **Umur / JK**: Nilai teks (contoh: *26 Th / Laki-laki*).
  - **Jaminan**: Nilai badge / teks (contoh: *BPJS Kesehatan* / *Asuransi Mandiri Inhealth* / *Umum*).
  - **Alergi**: Ditampilkan dengan warna merah tegas (`text-red-600 font-semibold`) untuk keamanan medis (contoh: *Ibuprofen, Debu*).
- Action Buttons (Kanan):
  - `🖨️ Cetak Gelang`: Tombol outline light blue dengan ikon printer.
  - `🕒 Riwayat Kunjungan`: Tombol outline light blue dengan ikon riwayat/jam.
  - `🟢 Stabil`: Badge status kondisi umum pasien dengan dot hijau menyala.

### 4.4 Form Asesmen SOAP (Kolom Kiri)
- Latar Belakang: Card putih `bg-white`, border `border-slate-200`, radius `rounded-xl`, shadow halus.
- Bagian Form:
  1. **SUBJEKTIF (S)**: Textarea multi-baris untuk keluhan utama, riwayat perjalanan penyakit, riwayat penyakit terdahulu, riwayat pengobatan.
  2. **OBJEKTIF (O)**: Textarea untuk temuan fisik objektif dokter, kesadaran, status lokalis.
  3. **ASESMEN (A)**:
     - Label "Diagnosa Utama" dilengkapi tautan atau badge modal "ICD-10".
     - Kolom input chip / autocomplete diagnosis:
       - Input pencarian kode ICD-10 dengan ikon kaca pembesar `🔍 A09.9 - Gastroenteritis`.
       - Diagnosis sekunder/tambahan: `I10 - Essential Hypertension`.
  4. **PLANNING (P)**: Textarea untuk rencana terapi, tindakan medis, edukasi pasien, dan anjuran kontrol.

### 4.5 Kartu Ringkasan Tanda-Tanda Vital (Kolom Kanan Atas)
- Wadah card putih berisi 5 kartu metrik horizontal:
  1. **TD (Tekanan Darah)**: Nilai `120/80`, Satuan `mmHg`.
  2. **HR (Heart Rate / Nadi)**: Nilai `82`, Satuan `bpm`.
  3. **Temp (Suhu Tubuh)**: Nilai `36.5`, Satuan `°C`.
  4. **RR (Respiratory Rate)**: Nilai `18`, Satuan `x/m`.
  5. **SpO2 (Saturasi O2)**: Nilai `98`, Satuan `%`.
- Desain Metrik:
  - Latar box kecil: Soft light gray `bg-slate-50 border border-slate-200 rounded-xl p-2.5 text-center`.
  - Label atas: Huruf kapital semi-bold abu-abu (11px).
  - Angka tengah: Tebal 18px warna gelap `font-bold text-slate-800`.
  - Satuan bawah: Abu-abu redup (10–11px).

### 4.6 Tab Terintegrasi 'E-Resep' & 'Order Lab/Rad' (Kolom Kanan Bawah)
- Tab Header:
  - Tab 1: `E-Resep` (Indikator aktif: garis bawah teal/cyan & teks berwarna).
  - Tab 2: `Lab & Radiologi` (State inactive atau switchable view).
- Konten E-Resep:
  - Tabel ringkas 3 kolom:
    - **Obat**: Nama obat & sediaan (contoh: *Paracetamol 500mg*, *Amoxicillin 500mg*).
    - **Jumlah**: Input kuantitas (contoh: *1*, *10*).
    - **Aturan Pakai / Signa**: Input aturan minum (contoh: *3x1 tablet*, *500mg*).
  - Tombol `+ Tambah Obat`: Tombol outline hijau/teal untuk menambah baris resep baru.
- Konten Order Penunjang (Lab & Radiologi):
  - Daftar periksa (*Checklist*) dengan checkbox interaktif:
    - `[✓] Darah Lengkap` (Lab)
    - `[✓] Rontgen Thorax` (Radiologi)
    - `[ ] Kimia Darah / SGOT SGPT`
    - `[ ] Urin Rutin`
    - `[ ] USG Abdomen`

---

## 5. Kontrak Data Frontend (JSON Schema Interface)

```typescript
export interface PatientClinicalContext {
  id: string;
  norm: string;
  name: string;
  age: string;
  gender: 'Laki-laki' | 'Perempuan';
  insurance: string;
  allergies: string[];
  status: 'Stabil' | 'Perhatian' | 'Kritis';
  avatarUrl?: string;
}

export interface VitalSigns {
  bloodPressure: string; // e.g. "120/80"
  heartRate: number;      // e.g. 82
  temperature: number;    // e.g. 36.5
  respiratoryRate: number;// e.g. 18
  spO2: number;           // e.g. 98
}

export interface DiagnosisItem {
  code: string;           // e.g. "A09.9"
  description: string;    // e.g. "Gastroenteritis and colitis of unspecified origin"
  isPrimary: boolean;
}

export interface PrescriptionItem {
  id: string;
  medicineName: string;
  quantity: number;
  signa: string;          // e.g. "3x1 tablet sesudah makan"
}

export interface DiagnosticOrderItem {
  id: string;
  name: string;
  category: 'LAB' | 'RADIOLOGY';
  checked: boolean;
}

export interface SoapRecord {
  encounterId: string;
  patientId: string;
  subjective: string;
  objective: string;
  assessmentDiagnoses: DiagnosisItem[];
  planning: string;
  vitals: VitalSigns;
  prescriptions: PrescriptionItem[];
  diagnosticOrders: DiagnosticOrderItem[];
}
```

---

## 6. Checklist Implementasi Standar Pengembang Frontend

- [x] Banner identitas pasien menggunakan warna latar `bg-sky-50` dengan peringatan alergi warna merah tegas.
- [x] Struktur kolom kiri (SOAP) dan kolom kanan (Vitals + Orders) responsif (12-column grid: 7-col kiri, 5-col kanan di desktop).
- [x] Komponen 5 Box TTV ringkas, bersih, dan kontras angka jelas.
- [x] Fitur auto-save debounce (misal: 1500ms) pada kolom SOAP agar dokter tidak kehilangan input saat koneksi terputus.
- [x] Navigasi keyboard cepat (Tab, Enter) untuk entri data resep dan penunjang medis.
