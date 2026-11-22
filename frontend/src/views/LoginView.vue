<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useUiStore } from '@/stores/ui'

const authStore = useAuthStore()
const uiStore = useUiStore()
const router = useRouter()

const username = ref('superadmin')
const password = ref('password123')
const rememberMe = ref(false)
const showPassword = ref(false)
const errorMsg = ref('')
const isLoading = ref(false)

// Live Time for bottom status bar
const currentTimeStr = ref('')
let timer: any = null

function updateClock() {
  const now = new Date()
  const days = ['Minggu', 'Senin', 'Selasa', 'Rabu', 'Kamis', 'Jumat', 'Sabtu']
  const months = ['Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun', 'Jul', 'Agu', 'Sep', 'Okt', 'Nov', 'Des']
  
  const dayName = days[now.getDay()]
  const date = now.getDate()
  const monthName = months[now.getMonth()]
  const year = now.getFullYear()
  const hours = String(now.getHours()).padStart(2, '0')
  const minutes = String(now.getMinutes()).padStart(2, '0')

  currentTimeStr.value = `${dayName}, ${date} ${monthName} ${year} | ${hours}:${minutes} WIB`
}

onMounted(() => {
  updateClock()
  timer = setInterval(updateClock, 1000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

interface DemoAccount {
  roleCode: string
  roleName: string
  username: string
  icon: string
}

const demoAccounts: DemoAccount[] = [
  { roleCode: 'SUPERADMIN', roleName: 'Superadmin', username: 'superadmin', icon: 'user' },
  { roleCode: 'REGISTRATION', roleName: 'Pendaftaran', username: 'pendaftaran', icon: 'clipboard' },
  { roleCode: 'CALL_CENTER', roleName: 'Call Center', username: 'callcenter', icon: 'phone' },
  { roleCode: 'NURSE', roleName: 'Perawat', username: 'perawat', icon: 'heart' },
  { roleCode: 'DOCTOR', roleName: 'Dokter', username: 'dokter', icon: 'stethoscope' },
  { roleCode: 'PHARMACY', roleName: 'Farmasi', username: 'farmasi', icon: 'pill' },
  { roleCode: 'SUPPORT', roleName: 'Penunjang', username: 'penunjang', icon: 'flask' },
  { roleCode: 'LOGISTICS', roleName: 'Logistik', username: 'logistik', icon: 'box' },
  { roleCode: 'CASHIER', roleName: 'Kasir', username: 'kasir', icon: 'credit-card' },
]

async function handleLogin() {
  errorMsg.value = ''
  if (!username.value || !password.value) {
    errorMsg.value = 'Silakan masukkan username dan kata sandi'
    return
  }

  isLoading.value = true
  try {
    const ok = await authStore.login(username.value, password.value)
    if (ok) {
      uiStore.showToast(`Selamat datang kembali, ${authStore.user?.full_name || username.value}!`, 'success')
      router.push('/')
    } else {
      errorMsg.value = 'Username atau kata sandi tidak sesuai'
    }
  } catch (err: any) {
    errorMsg.value = err.message || 'Login gagal. Silakan periksa koneksi atau kredensial Anda.'
  } finally {
    isLoading.value = false
  }
}

function quickLogin(demo: DemoAccount) {
  username.value = demo.username
  password.value = 'password123'
  handleLogin()
}
</script>

<template>
  <div class="min-h-screen w-full flex flex-col lg:flex-row font-sans bg-slate-50 select-none">
    
    <!-- ==================== LEFT COLUMN: BRANDING & CLINIC SHOWCASE ==================== -->
    <div class="lg:w-1/2 bg-gradient-to-br from-[#024B7A] via-[#02375A] to-[#011C30] text-white p-8 lg:p-12 flex flex-col justify-between relative overflow-hidden shrink-0">
      
      <!-- Subtle Hexagonal Pattern Overlay Background -->
      <div class="absolute inset-0 opacity-10 pointer-events-none bg-[radial-gradient(#38bdf8_1px,transparent_1px)] [background-size:20px_20px]"></div>
      
      <!-- Hexagon vector glow graphics -->
      <svg class="absolute -top-12 -left-12 w-96 h-96 text-sky-500/10 pointer-events-none" viewBox="0 0 100 100" fill="currentColor">
        <polygon points="50 1 95 25 95 75 50 99 5 75 5 25" />
      </svg>
      <svg class="absolute -bottom-24 -right-24 w-96 h-96 text-sky-400/10 pointer-events-none" viewBox="0 0 100 100" fill="currentColor">
        <polygon points="50 1 95 25 95 75 50 99 5 75 5 25" />
      </svg>

      <!-- Brand Header -->
      <div class="relative z-10 text-center space-y-2">
        <div class="w-14 h-14 mx-auto rounded-2xl bg-white text-teal-700 flex items-center justify-center font-bold text-3xl shadow-lg border border-sky-200">
          <svg class="w-8 h-8 text-[#0284C7]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M12 4v16m8-8H4" />
          </svg>
        </div>
        <h1 class="text-xl sm:text-2xl font-extrabold tracking-wide text-white drop-shadow-sm">
          SIMRS-ASOY E-CLINIC
        </h1>
        <p class="text-xs sm:text-sm text-sky-200 font-medium">
          Sistem Informasi Manajemen Rumah Sakit & Klinik Anak
        </p>
      </div>

      <!-- Hero Pediatric Clinic Illustration & Photo Card -->
      <div class="relative z-10 my-6 bg-white/10 backdrop-blur-md rounded-2xl p-4 border border-white/20 shadow-2xl max-w-lg mx-auto w-full">
        <div class="relative rounded-xl overflow-hidden shadow-inner flex flex-col sm:flex-row items-center bg-white/95 text-slate-800 p-3 gap-4">
          <!-- Illustration Left: Pediatric Doctor & Nurse -->
          <div class="flex-1 flex flex-col items-center sm:items-start text-center sm:text-left space-y-1">
            <div class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full bg-sky-100 text-[#0284C7] text-[10px] font-bold">
              <span class="w-1.5 h-1.5 rounded-full bg-sky-500 animate-pulse"></span>
              Pelayanan Terpadu Anak
            </div>
            <p class="font-bold text-xs text-slate-800">Dokter Spesialis & Konsultasi Anak</p>
            <p class="text-[11px] text-slate-500 leading-snug">
              Pelayanan tumbuh kembang, imunisasi, rawat jalan, dan rawat inap anak terintegrasi standar akreditasi KARS.
            </p>
          </div>

          <!-- Doctor & Baby Photo Banner -->
          <div class="w-32 h-28 sm:w-36 sm:h-28 rounded-lg overflow-hidden shrink-0 shadow-sm border border-slate-200">
            <img 
              src="https://images.unsplash.com/photo-1584515979956-d9f6e5d09982?auto=format&fit=crop&q=80&w=300" 
              alt="Dokter dan Pasien Anak" 
              class="w-full h-full object-cover"
            />
          </div>
        </div>
      </div>

      <!-- Live Clinic Statistics Strip (STATISTIK HARI INI) -->
      <div class="relative z-10 space-y-3 max-w-lg mx-auto w-full">
        <div>
          <h2 class="text-[11px] font-bold uppercase tracking-wider text-sky-300">
            STATISTIK HARI INI
          </h2>
          <div class="grid grid-cols-3 gap-4 pt-1.5 text-left border-b border-white/15 pb-3">
            <div>
              <p class="text-[11px] text-sky-200">Jumlah Pasien Aktif:</p>
              <p class="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">142</p>
            </div>
            <div>
              <p class="text-[11px] text-sky-200">Dokter Bertugas:</p>
              <p class="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">18</p>
            </div>
            <div>
              <p class="text-[11px] text-sky-200">Antrean Berjalan:</p>
              <p class="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">24</p>
            </div>
          </div>
        </div>

        <!-- Announcement / Notification Pills -->
        <div class="space-y-1.5 text-xs text-sky-100">
          <div class="bg-black/25 backdrop-blur-xs border border-white/10 rounded-xl px-3.5 py-2 flex items-center gap-2">
            <span class="w-2 h-2 rounded-full bg-amber-400 shrink-0"></span>
            <span class="text-[11px]">
              <strong class="text-white font-semibold">Pengumuman:</strong> Pemeliharaan sistem terjadwal pada hari Sabtu pukul 23:00 WIB.
            </span>
          </div>

          <div class="bg-black/25 backdrop-blur-xs border border-white/10 rounded-xl px-3.5 py-2 flex items-center gap-2">
            <span class="w-2 h-2 rounded-full bg-emerald-400 shrink-0"></span>
            <span class="text-[11px]">
              <strong class="text-white font-semibold">Update:</strong> VClaim BPJS v2.1 telah aktif.
            </span>
          </div>
        </div>
      </div>

      <!-- Left Column Footer -->
      <div class="relative z-10 pt-4 text-center text-[11px] text-sky-300/80 space-y-1 border-t border-white/10 mt-4">
        <p>Butuh Bantuan IT? Hubungi Ext. 112 atau email ke <a href="mailto:it.support@asoyclinic.com" class="text-white hover:underline">it.support@asoyclinic.com</a></p>
        <p class="text-[10px] text-sky-400/60">Hak cipta © 2026 | Versi Sistem Asoy E-clinic</p>
      </div>

    </div>

    <!-- ==================== RIGHT COLUMN: LOGIN FORM & FAST DEMO ROLES ==================== -->
    <div class="lg:w-1/2 flex flex-col justify-between p-6 sm:p-10 lg:p-14 overflow-y-auto">
      
      <!-- Top Spacer -->
      <div></div>

      <!-- Main Form Container -->
      <div class="max-w-md w-full mx-auto my-auto space-y-6">
        
        <!-- Welcome Heading -->
        <div class="text-center space-y-1">
          <h2 class="text-2xl sm:text-3xl font-bold text-slate-800 tracking-tight">
            Selamat Datang Kembali!
          </h2>
          <p class="text-xs sm:text-sm text-slate-500">
            Silakan masuk untuk mengakses sistem operasional.
          </p>
        </div>

        <!-- Login Card -->
        <div class="bg-white border border-slate-200/90 rounded-2xl p-6 sm:p-8 shadow-sm space-y-5">
          
          <!-- Error Alert Banner -->
          <div
            v-if="errorMsg"
            class="p-3 rounded-xl bg-red-50 border border-red-200 text-red-700 text-xs flex items-center gap-2.5"
          >
            <svg class="w-4 h-4 shrink-0 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span class="font-medium">{{ errorMsg }}</span>
          </div>

          <form class="space-y-4" @submit.prevent="handleLogin">
            
            <!-- Field 1: Username Pengguna -->
            <div class="space-y-1.5 text-left">
              <label class="block text-xs font-semibold text-slate-700">
                Username Pengguna <span class="text-red-500">*</span>
              </label>
              <div class="relative">
                <input
                  v-model="username"
                  type="text"
                  placeholder="superadmin"
                  required
                  class="w-full text-xs text-slate-800 px-3.5 py-2.5 rounded-xl border border-slate-300 bg-white focus:border-sky-500 focus:ring-2 focus:ring-sky-200 outline-none transition"
                />
              </div>
            </div>

            <!-- Field 2: Kata Sandi -->
            <div class="space-y-1.5 text-left">
              <label class="block text-xs font-semibold text-slate-700">
                Kata Sandi <span class="text-red-500">*</span>
              </label>
              <div class="relative">
                <input
                  v-model="password"
                  :type="showPassword ? 'text' : 'password'"
                  placeholder="••••••••••••"
                  required
                  class="w-full text-xs text-slate-800 pl-3.5 pr-10 py-2.5 rounded-xl border border-slate-300 bg-white focus:border-sky-500 focus:ring-2 focus:ring-sky-200 outline-none transition"
                />
                <!-- Eye Toggle Icon -->
                <button
                  type="button"
                  @click="showPassword = !showPassword"
                  class="absolute right-3 top-2.5 text-slate-400 hover:text-slate-600 transition"
                  title="Lihat / Sembunyikan Kata Sandi"
                >
                  <svg v-if="!showPassword" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                  </svg>
                  <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l18 18" />
                  </svg>
                </button>
              </div>
            </div>

            <!-- Remember Me & Forgot Password -->
            <div class="flex items-center justify-between text-xs pt-1">
              <label class="flex items-center gap-2 cursor-pointer text-slate-600">
                <input
                  v-model="rememberMe"
                  type="checkbox"
                  class="rounded text-sky-600 focus:ring-sky-500 border-slate-300 w-4 h-4"
                />
                <span>Ingat Saya</span>
              </label>

              <a href="#" @click.prevent="uiStore.showToast('Silakan hubungi administrator SIMRS untuk reset kata sandi.', 'info')" class="text-sky-600 hover:text-sky-700 hover:underline font-semibold">
                Lupa Kata Sandi?
              </a>
            </div>

            <!-- Submit Button: Masuk ke Sistem -->
            <button
              type="submit"
              :disabled="isLoading"
              class="w-full py-2.5 px-4 rounded-xl bg-[#0284C7] hover:bg-sky-700 active:bg-sky-800 text-white font-bold text-xs sm:text-sm shadow-md transition flex items-center justify-center gap-2 disabled:opacity-60 disabled:cursor-not-allowed"
            >
              <svg v-if="isLoading" class="animate-spin w-4 h-4 text-white" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"></path>
              </svg>
              <span>{{ isLoading ? 'Memverifikasi...' : 'Masuk ke Sistem' }}</span>
            </button>
          </form>

          <!-- Divider: ATAU AKSES CEPAT (DEMO ROLE) -->
          <div class="relative py-2">
            <div class="absolute inset-0 flex items-center">
              <div class="w-full border-t border-slate-200"></div>
            </div>
            <div class="relative flex justify-center text-[10px]">
              <span class="bg-white px-3 font-bold uppercase tracking-wider text-slate-400">
                ATAU AKSES CEPAT (DEMO ROLE)
              </span>
            </div>
          </div>

          <!-- 1-Click Fast Demo Login (9 Role) -->
          <div class="space-y-2">
            <div class="flex items-center justify-between text-[11px]">
              <span class="font-bold text-slate-700">1-CLICK FAST DEMO LOGIN (9 ROLE):</span>
              <span class="font-mono text-slate-500 font-semibold">pass: password123</span>
            </div>

            <!-- 9 Role Grid (3x3) -->
            <div class="grid grid-cols-3 gap-2">
              <button
                v-for="demo in demoAccounts"
                :key="demo.roleCode"
                type="button"
                @click="quickLogin(demo)"
                class="p-2 rounded-xl border border-slate-200 bg-slate-50 hover:bg-sky-50 hover:border-sky-300 text-left transition flex flex-col justify-center shadow-2xs group"
              >
                <div class="flex items-center gap-1.5 text-xs font-bold text-slate-800 group-hover:text-sky-700">
                  <!-- Role Icons -->
                  <span v-if="demo.icon === 'user'">👤</span>
                  <span v-else-if="demo.icon === 'clipboard'">📋</span>
                  <span v-else-if="demo.icon === 'phone'">📞</span>
                  <span v-else-if="demo.icon === 'heart'">🩺</span>
                  <span v-else-if="demo.icon === 'stethoscope'">👨‍⚕️</span>
                  <span v-else-if="demo.icon === 'pill'">💊</span>
                  <span v-else-if="demo.icon === 'flask'">🔬</span>
                  <span v-else-if="demo.icon === 'box'">📦</span>
                  <span v-else-if="demo.icon === 'credit-card'">💳</span>
                  <span class="truncate">{{ demo.roleName }}</span>
                </div>
                <span class="text-[10px] text-slate-400 group-hover:text-sky-600 font-mono mt-0.5">
                  @{{ demo.username }}
                </span>
              </button>
            </div>

            <p class="text-[10px] text-slate-400 text-center pt-1">
              Mode demo ini hanya untuk simulasi dan pelatihan internal.
            </p>
          </div>

        </div>

      </div>

      <!-- Bottom Right Status Indicator Bar -->
      <div class="pt-6 flex flex-col sm:flex-row items-center justify-end gap-3 text-[11px] text-slate-500">
        <span class="font-medium">{{ currentTimeStr }}</span>
        <span class="hidden sm:inline text-slate-300">|</span>
        <span class="flex items-center gap-1.5">
          <span>Server Status:</span>
          <span class="font-semibold text-slate-700">Normal</span>
          <span class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
        </span>
        <span class="hidden sm:inline text-slate-300">|</span>
        <span class="flex items-center gap-1.5">
          <span>Bridging BPJS:</span>
          <span class="font-semibold text-slate-700">Online</span>
          <span class="w-2 h-2 rounded-full bg-emerald-500"></span>
        </span>
      </div>

    </div>

  </div>
</template>
