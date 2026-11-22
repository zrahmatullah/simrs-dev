---
name: simrs-git-flow
description: >-
  Alur Git push dan branch management berjenjang untuk repository SIMRS
  (https://github.com/zrahmatullah/simrs-dev) dengan skema alur branch: dev -> simrs-check -> main.
  Gunakan skill ini saat pengguna ingin melakukan git commit, push ke dev, promote ke simrs-check,
  rilis ke main, atau menjalankan otomasi sinkronisasi Git SIMRS.
---

# SIMRS Git Flow: dev > simrs-check > main

Skill ini mengatur siklus kerja Git terstruktur untuk repositori:
**`https://github.com/zrahmatullah/simrs-dev`**

> [!NOTE]
> Git melarang karakter spasi pada nama branch (berdasarkan spesifikasi `git-check-ref-format`).
> Oleh karena itu, nama branch **"simrs check"** distandarkan menjadi **`simrs-check`** (kebab-case).

---

## 1. Arsitektur Branch

```text
Feature/Fix (Local)
       │
       ▼
 ┌───────────┐      Push & Test Feature
 │    dev    │ ◄────────────────────────────── (Pekerjaan harian / Active dev)
 └─────┬─────┘
       │  Merge setelah fitur selesai diuji lokal
       ▼
 ┌───────────────┐  Verifikasi Integrasi SIMRS
 │  simrs-check  │ ◄────────────────────────── (Staging / QA / Healthcheck)
 └─────┬─────────┘
       │  Merge setelah lolos QA & checklist validasi
       ▼
 ┌───────────┐      Deployment / Stable Release
 │   main    │ ◄────────────────────────────── (Production-ready / Stable)
 └───────────┘
```

| Branch | Tujuan | Kriteria Lolos |
| :--- | :--- | :--- |
| **`dev`** | Pengembangan aktif harian, perbaikan bug, penambahan modul baru | Kode jalan di lokal, tidak ada syntax error |
| **`simrs-check`** | Pengujian integrasi SIMRS, verifikasi modul EMR/Antrian/Billing, DB migration | Lolos pre-flight build (Go backend & Frontend Vite) |
| **`main`** | Rilis stabil production | Teruji penuh di `simrs-check`, siap dideploy |

---

## 2. Otomasi Cepat dengan Helper Script

Helper script PowerShell telah disediakan di:
`./.agents/skills/simrs-git-flow/scripts/simrs-flow.ps1`

### Perintah Utama:

1. **Inisialisasi & Setup Pertama Kali**:
   ```powershell
   powershell -ExecutionPolicy Bypass -File ./.agents/skills/simrs-git-flow/scripts/simrs-flow.ps1 setup
   ```
   *Menginisialisasi git (jika belum), memasang remote origin `https://github.com/zrahmatullah/simrs-dev`, membuat branch `main`, `simrs-check`, `dev`, dan push ke GitHub.*

2. **Push Pekerjaan ke `dev`**:
   ```powershell
   powershell -ExecutionPolicy Bypass -File ./.agents/skills/simrs-git-flow/scripts/simrs-flow.ps1 push-dev -m "feat(emr): tambah integrasi cppt"
   ```

3. **Promote dari `dev` ke `simrs-check`**:
   ```powershell
   powershell -ExecutionPolicy Bypass -File ./.agents/skills/simrs-git-flow/scripts/simrs-flow.ps1 promote-check
   ```
   *Merge `dev` ke `simrs-check`, memverifikasi build backend & frontend, lalu push ke `origin/simrs-check`.*

4. **Promote dari `simrs-check` ke `main` (Production Release)**:
   ```powershell
   powershell -ExecutionPolicy Bypass -File ./.agents/skills/simrs-git-flow/scripts/simrs-flow.ps1 promote-main -tag "v1.0.0"
   ```
   *Merge `simrs-check` ke `main`, tag versi (opsional), push ke `origin/main`, dan otomatis kembali ke branch `dev`.*

5. **Full Pipeline (Dev -> simrs-check -> main sekaligus)**:
   ```powershell
   powershell -ExecutionPolicy Bypass -File ./.agents/skills/simrs-git-flow/scripts/simrs-flow.ps1 full-sync -m "feat: rilis modul baru"
   ```

6. **Cek Status Semua Branch**:
   ```powershell
   powershell -ExecutionPolicy Bypass -File ./.agents/skills/simrs-git-flow/scripts/simrs-flow.ps1 status
   ```

---

## 3. Prosedur Manual (Step-by-Step)

Jika ingin menjalankan secara manual via CLI:

### Tahap 1: Commit & Push ke `dev`
```powershell
git checkout dev
git status
git add .
git commit -m "feat(modul): keterangan perubahan"
git push -u origin dev
```

### Tahap 2: Promosi ke `simrs-check`
```powershell
git checkout simrs-check
git merge dev --no-ff -m "merge: dev into simrs-check for verification"

# Jalankan Verifikasi Pra-Rilis
cd backend; go test ./... -run=^$; cd ..
cd frontend; npm run build; cd ..

git push -u origin simrs-check
```

### Tahap 3: Promosi ke `main`
```powershell
git checkout main
git merge simrs-check --no-ff -m "release: merge simrs-check into main"
git push -u origin main

# Kembali ke dev untuk lanjut coding
git checkout dev
```

---

## 4. Keamanan & Proteksi File Sensitif
Pastikan file berikut **TIDAK PERNAH** ter-commit:
- `backend/.env` (Berisi kredensial database dan JWT secret)
- `backend/server.exe` atau file binary executable lainnya
- `node_modules/` (baik di frontend maupun angular viewer)
- `frontend/dist/`
Gunakan `.gitignore` yang sudah terpasang di root workspace.
