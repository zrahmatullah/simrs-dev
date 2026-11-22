<#
.SYNOPSIS
    Automated Git Flow Manager for SIMRS
    Repo: https://github.com/zrahmatullah/simrs-dev
    Flow: dev -> simrs-check -> main
#>

[CmdletBinding()]
param (
    [Parameter(Position = 0, Mandatory = $true)]
    [ValidateSet("setup", "push-dev", "promote-check", "promote-main", "full-sync", "status")]
    [string]$Action,

    [Parameter(Mandatory = $false)]
    [Alias("m")]
    [string]$Message = "",

    [Parameter(Mandatory = $false)]
    [string]$Tag = "",

    [Parameter(Mandatory = $false)]
    [string]$RepoUrl = "https://github.com/zrahmatullah/simrs-dev"
)

$ErrorActionPreference = "Stop"

function Write-Info {
    param([string]$text)
    Write-Host "[INFO] $text" -ForegroundColor Cyan
}

function Write-Success {
    param([string]$text)
    Write-Host "[SUCCESS] $text" -ForegroundColor Green
}

function Write-Warn {
    param([string]$text)
    Write-Host "[WARN] $text" -ForegroundColor Yellow
}

function Write-Err {
    param([string]$text)
    Write-Host "[ERROR] $text" -ForegroundColor Red
}

function Ensure-GitInstalled {
    $gitCmd = Get-Command git -ErrorAction SilentlyContinue
    if (-not $gitCmd) {
        throw "Git tidak ditemukan di sistem PATH. Pastikan Git sudah terinstall."
    }
}

function Ensure-WorkingTreeClean {
    $status = git status --porcelain
    if ($status) {
        throw "Ada perubahan lokal yang belum di-commit. Jalankan push-dev atau commit terlebih dahulu.`n$status"
    }
}

function Run-PreflightChecks {
    Write-Info "Menjalankan Preflight Checks (Backend & Frontend)..."
    
    # 1. Check backend Go compilation if backend exists
    if (Test-Path "backend/go.mod") {
        Write-Info "Verifikasi Backend Go..."
        Push-Location backend
        try {
            $goBuild = go vet ./... 2>&1
            if ($LASTEXITCODE -ne 0) {
                Write-Warn "Go vet melaporkan issue: $goBuild"
            }
        } finally {
            Pop-Location
        }
    }

    # 2. Check frontend build if package.json exists
    if (Test-Path "frontend/package.json") {
        Write-Info "Verifikasi Frontend TypeScript/Vite..."
        Push-Location frontend
        try {
            if (Test-Path "node_modules") {
                npm run build --if-present
            } else {
                Write-Warn "node_modules frontend belum di-install. Melewati uji build frontend."
            }
        } finally {
            Pop-Location
        }
    }

    Write-Success "Preflight Checks selesai."
}

Ensure-GitInstalled

switch ($Action) {
    "setup" {
        Write-Info "Memulai setup Git Flow untuk SIMRS ($RepoUrl)..."

        # 1. Init if needed
        if (-not (Test-Path ".git")) {
            Write-Info "Inisialisasi repositori Git lokal..."
            git init -b main
        }

        # 2. Configure remote origin
        $existingRemote = git remote get-url origin 2>$null
        if (-not $existingRemote) {
            Write-Info "Menambahkan remote origin: $RepoUrl"
            git remote add origin $RepoUrl
        } elseif ($existingRemote -ne $RepoUrl) {
            Write-Info "Memperbarui remote origin ke: $RepoUrl"
            git remote set-url origin $RepoUrl
        } else {
            Write-Info "Remote origin sudah terarah ke: $RepoUrl"
        }

        # 3. Add and commit if root has untracked files
        $status = git status --porcelain
        if ($status) {
            Write-Info "Melakukan staging dan commit awal..."
            git add .
            git commit -m "chore: initial commit structure and git-flow"
        }

        # 4. Ensure main branch exists
        git branch -M main
        Write-Info "Mendorong branch 'main' ke remote..."
        git push -u origin main

        # 5. Create & push simrs-check branch
        Write-Info "Menyiapkan branch 'simrs-check'..."
        git checkout -B simrs-check
        git push -u origin simrs-check

        # 6. Create & push dev branch
        Write-Info "Menyiapkan branch 'dev'..."
        git checkout -B dev
        git push -u origin dev

        Write-Success "Setup selesai! Saat ini Anda berada di branch 'dev'."
        Write-Info "Alur aktif: dev -> simrs-check -> main"
    }

    "push-dev" {
        if (-not $Message) {
            $Message = Read-Host "Masukkan pesan commit untuk dev"
            if (-not $Message) {
                $Message = "feat: update working changes on dev"
            }
        }

        Write-Info "Beralih ke branch 'dev'..."
        git checkout dev

        Write-Info "Staging semua perubahan (sesuai .gitignore)..."
        git add .

        $status = git status --porcelain
        if ($status) {
            Write-Info "Commit ke dev: '$Message'"
            git commit -m "$Message"
        } else {
            Write-Info "Tidak ada perubahan file baru untuk di-commit."
        }

        Write-Info "Mendorong perubahan ke origin/dev..."
        git push -u origin dev
        Write-Success "Berhasil push ke branch 'dev'!"
    }

    "promote-check" {
        Ensure-WorkingTreeClean
        
        Write-Info "Mengambil referensi terbaru..."
        git fetch origin

        Write-Info "Beralih ke branch 'simrs-check'..."
        git checkout simrs-check
        git pull origin simrs-check --rebase 2>$null

        Write-Info "Melakukan merge dari 'dev' ke 'simrs-check'..."
        git merge dev --no-ff -m "merge: dev into simrs-check for verification"

        # Preflight verification
        Run-PreflightChecks

        Write-Info "Mendorong ke origin/simrs-check..."
        git push origin simrs-check
        Write-Success "Berhasil promote ke branch 'simrs-check'!"

        Write-Info "Kembali ke branch 'dev'..."
        git checkout dev
    }

    "promote-main" {
        Ensure-WorkingTreeClean

        Write-Info "Beralih ke branch 'main'..."
        git checkout main
        git pull origin main --rebase 2>$null

        Write-Info "Melakukan merge dari 'simrs-check' ke 'main'..."
        git merge simrs-check --no-ff -m "release: merge simrs-check into main"

        if ($Tag) {
            Write-Info "Membuat Git Tag '$Tag'..."
            git tag -a $Tag -m "Release $Tag"
            git push origin $Tag
        }

        Write-Info "Mendorong ke origin/main..."
        git push origin main
        Write-Success "Berhasil promote & rilis ke branch 'main'!"

        Write-Info "Kembali ke branch 'dev' untuk melanjutkan pengembangan..."
        git checkout dev
    }

    "full-sync" {
        if (-not $Message) {
            $Message = Read-Host "Masukkan pesan commit untuk dev"
            if (-not $Message) {
                $Message = "feat: automated full sync across dev -> simrs-check -> main"
            }
        }

        Write-Info "=== [Step 1/3] Commit & Push ke DEV ==="
        & $PSCommandPath push-dev -Message $Message

        Write-Info "=== [Step 2/3] Promote ke SIMRS-CHECK ==="
        & $PSCommandPath promote-check

        Write-Info "=== [Step 3/3] Promote ke MAIN ==="
        & $PSCommandPath promote-main -Tag $Tag

        Write-Success "=== ALUR LENGKAP SELESAI: dev -> simrs-check -> main ==="
    }

    "status" {
        Write-Info "Status Branch Lokal & Remote:"
        git branch -a -v
        Write-Info "Ringkasan Git Status:"
        git status -s
    }
}
