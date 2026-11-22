# STANDARISASI KEAMANAN CYBER & IMPLEMENTASI OWASP TOP 10
**Sistem Informasi Manajemen Rumah Sakit (SIMRS-ASOY)**

Dokumen ini memuat spesifikasi dan implementasi teknis kontrol keamanan siber (*cybersecurity controls*) pada sistem SIMRS-ASOY, mengacu pada standar internasional **OWASP Top 10 (2021)**, **OWASP ASVS v4.0 (Application Security Verification Standard)**, dan regulasi perlindungan data rekam medis **Permenkes 24/2022**.

---

## 1. Pemetaan Kontrol Keamanan Terhadap OWASP Top 10

| Kode OWASP | Kategori Kerentanan | Mitigasi & Implementasi Teknis di SIMRS-ASOY | Status |
| :---: | :--- | :--- | :---: |
| **A01:2021** | **Broken Access Control** | Autentikasi JWT Bearer wajib pada seluruh rute API terproteksi, Middleware RBAC granular berbasis peran (`RequireRoles`), penguncian isolasi rekam medis (`is_locked: true`). | **TERLINDUNGI** |
| **A02:2021** | **Cryptographic Failures** | Password di-hash menggunakan **bcrypt** dengan cost 10. Dual-token: Access Token (15 menit) + Refresh Token 256-bit di-hash **SHA-256**. Enforce algoritma `HS256` ketat. | **TERLINDUNGI** |
| **A03:2021** | **Injection** | Seluruh query SQL ke PostgreSQL menggunakan *Parameterized Queries* (`$1, $2, ...` via sqlx) tanpa konkatenasi string. Sanitasi XSS pada teks narasi SOAP klinis. | **TERLINDUNGI** |
| **A04:2021** | **Insecure Design** | Rate limiting anti-brute-force pada `/auth/login` (5 req/menit), global rate limiter (120 req/menit), validasi integer flow & limit transaksi keuangan. | **TERLINDUNGI** |
| **A05:2021** | **Security Misconfiguration** | Injeksi HTTP Security Headers lengkap (CSP, X-Frame-Options DENY, X-Content-Type-Options nosniff, HSTS, Permissions-Policy). Limit ukuran body 2MB. | **TERLINDUNGI** |
| **A06:2021** | **Vulnerable & Outdated Components** | Audit dependensi Go (`go mod`) dan NPM frontend secara berkala. Menghindari library usang. | **TERLINDUNGI** |
| **A07:2021** | **Identification & Authentication Failures** | Refresh Token Rotation (RTR). Deteksi *token theft/reuse* otomatis mencabut seluruh sesi pengguna. Endpoint `/auth/logout` untuk pemutusan sesi resmi. | **TERLINDUNGI** |
| **A08:2021** | **Software & Data Integrity Failures** | Verifikasi integritas tanda tangan JWT. Validasi batas fisiologis TTV dan keabsahan numerik sebelum persistensi data. | **TERLINDUNGI** |
| **A09:2021** | **Security Logging & Monitoring Failures** | Tabel `audit_logs` mencatat seluruh login, login gagal, transaksi pembayaran, dispensing obat, input rekam medis, dan penerbitan surat resmi. | **TERLINDUNGI** |
| **A10:2021** | **Server-Side Request Forgery (SSRF)** | Tidak ada pemanggilan URL eksternal yang dikontrol langsung oleh input pengguna tanpa whitelist ketat. | **TERLINDUNGI** |

---

## 2. Arsitektur Sesi & Token Otentikasi (OWASP ASVS V3)

Sebelumnya, aplikasi hanya menggunakan satu token dengan masa berlaku 24 jam. Hal ini menimbulkan risiko tinggi apabila token disadap. Sistem kini telah ditingkatkan menjadi **Dual-Token Architecture dengan Refresh Token Rotation (RTR)**:

```
[ FRONTEND CLIENT ]                                    [ BACKEND SERVER ]
         |                                                      |
         | --- 1. POST /api/v1/auth/login (User & Pass) ------> |
         |                                                      | Verifikasi Hash Bcrypt
         |                                                      | Buat Access Token (15 Mins)
         |                                                      | Buat Refresh Token (32 Bytes Crypto Random)
         |                                                      | Simpan Hash SHA-256 ke user_refresh_tokens
         | <-- 2. Return { access_token, refresh_token } ------ |
         |                                                      |
         | --- 3. Request API + Header 'Bearer AccessToken' --> | (Otentikasi Berjalan Normal)
         |                                                      |
         | ... (15 Menit Kemudian: Access Token Kedaluwarsa) .. |
         |                                                      |
         | --- 4. Request API dengan Token Kedaluwarsa -------> |
         | <-- 5. Response 401 Unauthorized ------------------- |
         |                                                      |
         | --- 6. Interceptor: POST /api/v1/auth/refresh -----> | Cek Hash Token di DB
         |        { "refresh_token": "..." }                    | Verifikasi Belum Kedaluwarsa / Belum Revoked
         |                                                      | Cabut (Revoke) Token Lama
         |                                                      | Buat Access Token Baru + Refresh Token Baru
         | <-- 7. Return Token Baru (Token Rotated) ----------- |
         |                                                      |
         | --- 8. Replay Otomatis Request Awal ---------------> | Request Sukses! Dokter Tidak Terganggu.
```

### 2.1 Fitur Deteksi Token Theft (Pencurian Token)
Jika token refresh yang telah dicabut (*revoked*) digunakan kembali:
1. Backend mendeteksi adanya potensi penyadapan (*replay attack*).
2. Sistem secara otomatis **mencabut seluruh sesi aktif milik pengguna tersebut** (`RevokeAllUserTokens`).
3. Dicatat peringatan keamanan tingkat tinggi pada tabel `audit_logs`: `SECURITY_ALERT_TOKEN_REUSE`.
4. Pengguna dipaksa login ulang dengan kredensial utama.

---

## 3. Rate Limiting (Pencegahan Serangan Brute-Force & DoS)

Diterapkan middleware berbasis IP (`IPRateLimiter`) dengan sliding window:
1. **Login Rate Limiter (`/api/v1/auth/login`)**:
   - **Batas**: Maksimal **5 percobaan login per menit** per IP address.
   - **Respons jika Terlampaui**: HTTP 429 Too Many Requests, header `Retry-After: 60`, dan pesan `"Terlalu banyak percobaan login. Silakan tunggu 1 menit demi keamanan."`.
2. **Global API Rate Limiter (`/api/v1/*`)**:
   - **Batas**: Maksimal **120 permintaan per menit** per IP address untuk mencegah *scraping* data pasien dan serangan Denial-of-Service (DoS).
3. **Pembersihan Memori Otomatis**:
   - Goroutine latar belakang berjalan setiap 5 menit untuk menghapus tracker IP yang tidak aktif sehingga terhindar dari kebocoran memori (*memory leak*).

---

## 4. Validasi Integer Flow & Keamanan Logika Bisnis (Arithmetic Security)

Modul `internal/security/validator.go` mencegah manipulasi angka negatif dan integer overflow:
1. **Kuantitas Obat & Barang (`ValidateQuantity`)**:
   - `qty` wajib $\ge 1$ dan $\le 10.000$. Menolak angka negatif, nol, atau angka astronomis.
2. **Nominal Keuangan & Kasir (`ValidatePrice`)**:
   - Mencegah nilai `NaN` (*Not a Number*) atau `+Inf / -Inf`.
   - Nominal pembayaran kasir wajib $> 0$ dan $\le 1.000.000.000$ (1 Miliar Rupiah).
   - Nilai diskon tidak boleh melebihi subtotal faktur tagihan (`ValidateDiscount`).
3. **Tanda-Tanda Vital Klinis (`ValidateVitalSigns`)**:
   - TD Sistole: 30 – 320 mmHg.
   - TD Diastole: 20 – 220 mmHg.
   - Heart Rate: 20 – 300 bpm.
   - Laju Napas (RR): 4 – 100 x/menit.
   - Suhu Tubuh: 25.0 – 46.0 °C.
   - SpO2: 0.0 – 100.0 %.

---

## 5. Pertahanan Sanitasi XSS & Injeksi SQL (Injection Defense)

1. **Sanitasi XSS Narasi Medis (`SanitizeText`)**:
   - Menghapus seluruh tag `<script>`, `<iframe>`, `<style>`.
   - Menghapus event handler berbahaya seperti `onload=`, `onerror=`, `onclick=`.
   - Menghilangkan skema URL palsu seperti `javascript:`.
   - Melindungi isian narasi SOAP dokter, catatan keperawatan, keluhan pasien, dan resep racikan dari serangan *Stored XSS*.
2. **Sanitasi Teks Polos (`SanitizePlainText`)**:
   - Digunakan pada nama pasien, NIK, nomor RM, supplier, dan kode referensi dengan konversi entitas HTML aman (`html.EscapeString`).
3. **Pembatasan Ukuran Muatan Body (`BodyLimit`)**:
   - Middleware `BodyLimit(2 << 20)` membatasi ukuran request body maksimal **2MB** untuk mencegah serangan *Buffer Overflow* atau *Large Payload Memory Exhaustion*.

---

## 6. HTTP Security Headers Baku

Setiap respons dari server SIMRS-ASOY menyertakan header keamanan berikut:
```http
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
Content-Security-Policy: default-src 'self'; frame-ancestors 'none'; object-src 'none';
Permissions-Policy: geolocation=(), camera=(), microphone=()
```
