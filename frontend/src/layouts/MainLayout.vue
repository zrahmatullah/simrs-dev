<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useUiStore } from '@/stores/ui'
import { useQueueStore } from '@/stores/queue'
import { usePatientContextStore } from '@/stores/patientContext'
import { api } from '@/api/client'
import type { Patient } from '@/types'
import ToastContainer from '@/components/common/ToastContainer.vue'
import ConfirmModal from '@/components/common/ConfirmModal.vue'
import BaseModal from '@/components/common/BaseModal.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'

import {
  LayoutDashboard,
  Users,
  Database,
  ShieldAlert,
  UserPlus,
  Contact,
  CalendarClock,
  Activity,
  Stethoscope,
  Pill,
  FlaskConical,
  Boxes,
  Receipt,
  Clock,
  Clock3,
  Search,
  User,
  LogOut,
  KeyRound,
  CheckCircle2,
  ChevronDown,
  X,
  Menu,
  ChevronLeft,
  ChevronRight,
  AlertCircle,
  FileSpreadsheet,
  TrendingUp,
  PieChart
} from 'lucide-vue-next'

const authStore = useAuthStore()
const uiStore = useUiStore()
const queueStore = useQueueStore()
const patientStore = usePatientContextStore()
const router = useRouter()
const route = useRoute()

// ==========================================
// 1. SIDEBAR MENU GROUPING & ICONS
// ==========================================
interface MenuItem {
  name: string
  path: string
  roles: string[]
  icon: any
}

interface MenuGroup {
  title: string
  items: MenuItem[]
}

const menuGroups: MenuGroup[] = [
  {
    title: 'Dashboard & Monitoring',
    items: [
      {
        name: 'Dashboard Utama',
        path: '/dashboard',
        roles: ['SUPERADMIN', 'REGISTRATION', 'CALL_CENTER', 'NURSE', 'DOCTOR', 'PHARMACY', 'SUPPORT', 'LOGISTICS', 'CASHIER'],
        icon: LayoutDashboard
      },
    ]
  },
  {
    title: 'Pelayanan & Antrean',
    items: [
      { name: 'Pendaftaran & Antrean', path: '/registration', roles: ['SUPERADMIN', 'REGISTRATION'], icon: UserPlus },
      { name: 'Daftar Pasien', path: '/registration/patients', roles: ['SUPERADMIN', 'REGISTRATION', 'CALL_CENTER', 'NURSE', 'DOCTOR'], icon: Contact },
      { name: 'Jadwal & Appointment', path: '/call-center', roles: ['SUPERADMIN', 'CALL_CENTER'], icon: CalendarClock },
    ]
  },
  {
    title: 'Pelayanan Medis (EMR)',
    items: [
      { name: 'Poli & TTV Keperawatan', path: '/nurse', roles: ['SUPERADMIN', 'NURSE'], icon: Activity },
      { name: 'Rekam Medis (EMR)', path: '/doctor', roles: ['SUPERADMIN', 'DOCTOR'], icon: Stethoscope },
    ]
  },
  {
    title: 'Farmasi & Penunjang',
    items: [
      { name: 'Instalasi Farmasi', path: '/pharmacy', roles: ['SUPERADMIN', 'PHARMACY'], icon: Pill },
      { name: 'Penunjang (Lab/Rad)', path: '/support', roles: ['SUPERADMIN', 'SUPPORT'], icon: FlaskConical },
    ]
  },
  {
    title: 'Billing & Logistik',
    items: [
      { name: 'Kasir & Pembayaran', path: '/cashier', roles: ['SUPERADMIN', 'CASHIER'], icon: Receipt },
      { name: 'Logistik & Gudang', path: '/logistics', roles: ['SUPERADMIN', 'LOGISTICS'], icon: Boxes },
    ]
  },
  {
    title: 'Laporan & Eksekutif',
    items: [
      { name: 'Laporan Keuangan', path: '/reports?tab=financial', roles: ['SUPERADMIN', 'CASHIER'], icon: TrendingUp },
      { name: 'Laporan Stok Farmasi', path: '/reports?tab=pharmacy', roles: ['SUPERADMIN', 'PHARMACY', 'LOGISTICS'], icon: Boxes },
      { name: 'Laporan Pasien Daftar', path: '/reports?tab=registrations', roles: ['SUPERADMIN', 'REGISTRATION', 'CALL_CENTER', 'NURSE', 'DOCTOR'], icon: FileSpreadsheet },
      { name: 'Laporan Demografi', path: '/reports?tab=demographics', roles: ['SUPERADMIN', 'REGISTRATION', 'DOCTOR', 'NURSE'], icon: PieChart },
    ]
  },
  {
    title: 'Pengaturan & Sistem',
    items: [
      { name: 'Dashboard Admin', path: '/admin', roles: ['SUPERADMIN'], icon: LayoutDashboard },
      { name: 'Manajemen User', path: '/admin/users', roles: ['SUPERADMIN'], icon: Users },
      { name: 'Master Data', path: '/admin/master', roles: ['SUPERADMIN'], icon: Database },
      { name: 'Jejak Audit Sistem', path: '/admin/audit', roles: ['SUPERADMIN'], icon: ShieldAlert },
    ]
  }
]

// Filter groups and items by user role
const visibleMenuGroups = computed(() => {
  return menuGroups
    .map(group => ({
      title: group.title,
      items: group.items.filter(item => authStore.hasRole(item.roles))
    }))
    .filter(group => group.items.length > 0)
})

// ==========================================
// 2. REALTIME CLOCK
// ==========================================
const currentTime = ref('')
const currentDate = ref('')
let clockTimer: any = null

function updateClock() {
  const now = new Date()
  currentDate.value = now.toLocaleDateString('id-ID', {
    weekday: 'short',
    day: 'numeric',
    month: 'short',
    year: 'numeric'
  })
  const hours = String(now.getHours()).padStart(2, '0')
  const minutes = String(now.getMinutes()).padStart(2, '0')
  const seconds = String(now.getSeconds()).padStart(2, '0')
  currentTime.value = `${hours}:${minutes}:${seconds} WIB`
}

// ==========================================
// 3. STATISTIK PASIEN HARI INI
// ==========================================
const totalSelesai = computed(() => {
  return queueStore.metrics['SELESAI'] || 0
})

const totalBelumSelesai = computed(() => {
  const total = queueStore.metrics['TOTAL'] || 0
  const selesai = queueStore.metrics['SELESAI'] || 0
  return Math.max(0, total - selesai)
})

let metricsTimer: any = null

// ==========================================
// 4. QUICK PATIENT SEARCH IN NAVBAR
// ==========================================
const searchQuery = ref('')
const searchResults = ref<Patient[]>([])
const isSearching = ref(false)
const isSearchDropdownOpen = ref(false)
const searchContainerRef = ref<HTMLElement | null>(null)
let searchDebounce: any = null

function handleSearchInput() {
  if (searchDebounce) clearTimeout(searchDebounce)
  if (!searchQuery.value.trim()) {
    searchResults.value = []
    isSearchDropdownOpen.value = false
    return
  }

  isSearchDropdownOpen.value = true
  searchDebounce = setTimeout(async () => {
    isSearching.value = true
    try {
      const res = await api.get<Patient[]>('/patients', {
        search: searchQuery.value.trim(),
        limit: 5
      })
      if (res.success && res.data) {
        searchResults.value = res.data
      }
    } catch (err) {
      console.error('Navbar patient search error:', err)
    } finally {
      isSearching.value = false
    }
  }, 250)
}

function handleSelectPatient(patient: Patient) {
  patientStore.setActivePatient(patient)
  uiStore.showToast(`Pasien ${patient.nama_pasien} (${patient.no_rm}) berhasil diaktifkan`, 'success')
  searchQuery.value = ''
  searchResults.value = []
  isSearchDropdownOpen.value = false
}

function handleSearchSubmit() {
  if (searchQuery.value.trim()) {
    router.push({
      path: '/registration/patients',
      query: { search: searchQuery.value.trim() }
    })
    isSearchDropdownOpen.value = false
  }
}

function clearSearch() {
  searchQuery.value = ''
  searchResults.value = []
  isSearchDropdownOpen.value = false
}

// ==========================================
// 5. PROFILE DROPDOWN & PASSWORD MODAL
// ==========================================
const isProfileDropdownOpen = ref(false)
const profileDropdownRef = ref<HTMLElement | null>(null)

const isPasswordModalOpen = ref(false)
const isSubmittingPassword = ref(false)
const passwordError = ref('')
const passwordForm = ref({
  newPassword: '',
  confirmPassword: ''
})

function toggleProfileDropdown() {
  isProfileDropdownOpen.value = !isProfileDropdownOpen.value
}

function openPasswordModal() {
  isProfileDropdownOpen.value = false
  passwordForm.value = {
    newPassword: '',
    confirmPassword: ''
  }
  passwordError.value = ''
  isPasswordModalOpen.value = true
}

async function handleChangePassword() {
  passwordError.value = ''

  if (!passwordForm.value.newPassword) {
    passwordError.value = 'Kata sandi baru wajib diisi'
    return
  }
  if (passwordForm.value.newPassword.length < 6) {
    passwordError.value = 'Kata sandi baru minimal 6 karakter'
    return
  }
  if (passwordForm.value.newPassword !== passwordForm.value.confirmPassword) {
    passwordError.value = 'Konfirmasi kata sandi tidak cocok'
    return
  }

  isSubmittingPassword.value = true
  try {
    const userId = authStore.user?.id
    if (!userId) {
      throw new Error('ID pengguna tidak valid')
    }

    const res = await api.put(`/users/${userId}`, {
      password: passwordForm.value.newPassword,
      email: authStore.user?.email || '',
      full_name: authStore.user?.full_name || '',
      nip: authStore.user?.nip || '',
      phone: authStore.user?.phone || '',
      role_id: authStore.user?.role_id || 1,
      status_aktif: true
    })

    if (res.success) {
      uiStore.showToast('Kata sandi berhasil diperbarui!', 'success')
      isPasswordModalOpen.value = false
    } else {
      passwordError.value = res.message || 'Gagal mengubah kata sandi'
    }
  } catch (err: any) {
    passwordError.value = err.message || 'Gagal mengubah kata sandi'
  } finally {
    isSubmittingPassword.value = false
  }
}

function handleLogout() {
  isProfileDropdownOpen.value = false
  uiStore.openConfirm({
    title: 'Keluar dari Sistem',
    message: 'Apakah Anda yakin ingin keluar dari sesi SIMRS-ASOY?',
    confirmText: 'Ya, Keluar',
    variant: 'danger',
    onConfirm: () => {
      authStore.logout()
    }
  })
}

// Global click outside listener for dropdowns
function handleClickOutside(e: MouseEvent) {
  const target = e.target as Node
  if (profileDropdownRef.value && !profileDropdownRef.value.contains(target)) {
    isProfileDropdownOpen.value = false
  }
  if (searchContainerRef.value && !searchContainerRef.value.contains(target)) {
    isSearchDropdownOpen.value = false
  }
}

onMounted(() => {
  updateClock()
  clockTimer = setInterval(updateClock, 1000)

  queueStore.fetchQueueMetrics()
  metricsTimer = setInterval(() => {
    queueStore.fetchQueueMetrics()
  }, 25000)

  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  if (clockTimer) clearInterval(clockTimer)
  if (metricsTimer) clearInterval(metricsTimer)
  document.removeEventListener('click', handleClickOutside)
})
</script>

<template>
  <div class="min-h-screen flex bg-[#F8FAFC] text-[#0F172A]">
    <!-- Toast & Confirm Modals -->
    <ToastContainer />
    <ConfirmModal />

    <!-- Modal Setting / Ganti Password -->
    <BaseModal
      :is-open="isPasswordModalOpen"
      title="Setting Password Pengguna"
      max-width="md"
      @close="isPasswordModalOpen = false"
    >
      <form @submit.prevent="handleChangePassword" class="space-y-4">
        <div class="text-xs text-slate-500 mb-2">
          Ubah kata sandi untuk akun <span class="font-semibold text-slate-800">{{ authStore.currentUserName }}</span> ({{ authStore.currentRole }}).
        </div>

        <div v-if="passwordError" class="p-3 bg-red-50 border border-red-200 rounded-md text-xs text-red-600 font-medium flex items-center gap-2">
          <AlertCircle class="w-4 h-4 shrink-0 text-red-500" />
          <span>{{ passwordError }}</span>
        </div>

        <div>
          <label class="block text-xs font-semibold text-slate-700 mb-1">
            Kata Sandi Baru <span class="text-red-500">*</span>
          </label>
          <input
            v-model="passwordForm.newPassword"
            type="password"
            placeholder="Minimal 6 karakter"
            required
            autocomplete="new-password"
            class="w-full px-3 py-2 text-sm border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-[#0284C7] focus:border-transparent transition-all"
          />
        </div>

        <div>
          <label class="block text-xs font-semibold text-slate-700 mb-1">
            Konfirmasi Kata Sandi Baru <span class="text-red-500">*</span>
          </label>
          <input
            v-model="passwordForm.confirmPassword"
            type="password"
            placeholder="Ulangi kata sandi baru"
            required
            autocomplete="new-password"
            class="w-full px-3 py-2 text-sm border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-[#0284C7] focus:border-transparent transition-all"
          />
        </div>

        <div class="flex items-center justify-end gap-2 pt-3 border-t border-slate-100">
          <BaseButton
            type="button"
            variant="ghost"
            size="sm"
            @click="isPasswordModalOpen = false"
          >
            Batal
          </BaseButton>
          <BaseButton
            type="submit"
            variant="primary"
            size="sm"
            :loading="isSubmittingPassword"
          >
            Simpan Kata Sandi
          </BaseButton>
        </div>
      </form>
    </BaseModal>

    <!-- ======================================================== -->
    <!-- COLLAPSIBLE SIDEBAR WITH GROUPED MENUS & CUSTOM ICONS   -->
    <!-- (BOTTOM USER INFO & LOGOUT TAKEN OUT AS REQUESTED)       -->
    <!-- ======================================================== -->
    <aside
      :class="[
        'bg-white border-r border-[#E2E8F0] flex flex-col shrink-0 transition-all duration-200 z-40 select-none sticky top-0 h-screen',
        uiStore.isSidebarCollapsed ? 'w-[68px]' : 'w-[250px]'
      ]"
    >
      <!-- Sidebar Brand Header -->
      <div class="h-14 border-b border-[#E2E8F0] px-3.5 flex items-center justify-between bg-[#F8FAFC]">
        <div v-if="!uiStore.isSidebarCollapsed" class="flex items-center gap-2.5 overflow-hidden">
          <div class="w-8 h-8 rounded-lg bg-[#0284C7] text-white flex items-center justify-center font-bold text-base shadow-sm shrink-0">
            S
          </div>
          <div class="leading-tight truncate">
            <span class="font-bold text-[14px] text-[#0284C7] tracking-tight block truncate">SIMRS-ASOY</span>
            <span class="text-[10px] text-[#64748B] block truncate font-medium">Klinik & Rumah Sakit</span>
          </div>
        </div>

        <div v-else class="mx-auto">
          <div class="w-8 h-8 rounded-lg bg-[#0284C7] text-white flex items-center justify-center font-bold text-base shadow-sm">
            S
          </div>
        </div>

        <!-- Sidebar Collapse Toggle Button -->
        <button
          v-if="!uiStore.isSidebarCollapsed"
          type="button"
          class="p-1 rounded text-[#64748B] hover:text-[#0F172A] hover:bg-[#F1F5F9] transition-colors"
          title="Ciutkan Sidebar"
          @click="uiStore.toggleSidebar()"
        >
          <ChevronLeft class="w-4 h-4" />
        </button>
      </div>

      <!-- Navigation Menu Groups -->
      <nav class="flex-1 py-3 px-2 space-y-4 overflow-y-auto">
        <div
          v-for="group in visibleMenuGroups"
          :key="group.title"
          class="space-y-1"
        >
          <!-- Group Title (Hidden when collapsed) -->
          <div
            v-if="!uiStore.isSidebarCollapsed"
            class="px-2.5 pt-2 pb-1 text-[10px] font-bold text-[#94A3B8] uppercase tracking-wider select-none"
          >
            {{ group.title }}
          </div>
          <!-- Divider when collapsed -->
          <div
            v-else
            class="my-2 border-t border-[#E2E8F0] mx-1"
          ></div>

          <!-- Menu Items within Group -->
          <router-link
            v-for="item in group.items"
            :key="item.path"
            :to="item.path"
            :class="[
              'flex items-center gap-3 px-2.5 py-2 rounded-[6px] text-[13px] font-medium transition-all group relative',
              route.path === item.path || (item.path !== '/' && route.path.startsWith(item.path + '/'))
                ? 'bg-[#E0F2FE] text-[#0369A1] font-semibold shadow-xs'
                : 'text-[#475569] hover:bg-[#F1F5F9] hover:text-[#0F172A]'
            ]"
            :title="uiStore.isSidebarCollapsed ? item.name : undefined"
          >
            <!-- Custom Lucide Icon matched to menu title -->
            <div
              :class="[
                'w-5 h-5 flex items-center justify-center shrink-0 transition-colors',
                route.path === item.path || (item.path !== '/' && route.path.startsWith(item.path + '/'))
                  ? 'text-[#0284C7]'
                  : 'text-[#64748B] group-hover:text-[#0F172A]'
              ]"
            >
              <component :is="item.icon" class="w-4 h-4" />
            </div>

            <span v-if="!uiStore.isSidebarCollapsed" class="truncate">
              {{ item.name }}
            </span>
          </router-link>
        </div>
      </nav>

      <!-- Sidebar Minimal Footer (Expand Toggle Only - No logout / user info) -->
      <div v-if="uiStore.isSidebarCollapsed" class="p-2 border-t border-[#E2E8F0] bg-[#F8FAFC]">
        <button
          type="button"
          class="w-full h-8 flex items-center justify-center text-[#64748B] hover:text-[#0F172A] rounded hover:bg-[#F1F5F9] transition-colors"
          title="Perluas Sidebar"
          @click="uiStore.toggleSidebar()"
        >
          <ChevronRight class="w-4 h-4" />
        </button>
      </div>
    </aside>

    <!-- ======================================================== -->
    <!-- MAIN CONTENT AREA WITH UPGRADED NAVBAR                   -->
    <!-- ======================================================== -->
    <div class="flex-1 flex flex-col min-w-0 overflow-y-auto">
      <!-- Top Navbar Header -->
      <header class="h-14 bg-white border-b border-[#E2E8F0] px-3 sm:px-4 flex items-center justify-between shrink-0 sticky top-0 z-20 gap-3">
        <!-- Left: Sidebar Toggle & Realtime Clock -->
        <div class="flex items-center gap-2.5 sm:gap-3 shrink-0">
          <button
            v-if="uiStore.isSidebarCollapsed"
            type="button"
            class="p-1.5 rounded-md text-[#475569] hover:bg-[#F1F5F9] transition-colors"
            title="Buka Sidebar"
            @click="uiStore.toggleSidebar()"
          >
            <Menu class="w-5 h-5" />
          </button>

          <!-- Realtime Live Clock -->
          <div class="flex items-center gap-2 px-2.5 py-1 bg-slate-50 border border-slate-200/80 rounded-md text-[12px] text-slate-700">
            <Clock class="w-3.5 h-3.5 text-[#0284C7] shrink-0" />
            <span class="font-medium hidden md:inline text-slate-600">{{ currentDate }}</span>
            <span class="text-slate-300 hidden md:inline">•</span>
            <span class="font-mono font-semibold tabular-nums text-slate-800">{{ currentTime }}</span>
          </div>
        </div>

        <!-- Center: Search Pasien -->
        <div ref="searchContainerRef" class="relative flex-1 max-w-xs sm:max-w-sm md:max-w-md">
          <div class="relative flex items-center">
            <Search class="w-4 h-4 text-slate-400 absolute left-3 pointer-events-none" />
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Cari Pasien (No. RM / Nama / NIK)..."
              class="w-full pl-9 pr-8 py-1.5 text-xs bg-slate-50 hover:bg-slate-100/70 focus:bg-white border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-[#0284C7] focus:border-transparent transition-all placeholder:text-slate-400"
              @input="handleSearchInput"
              @keydown.enter="handleSearchSubmit"
              @focus="searchQuery && (isSearchDropdownOpen = true)"
            />
            <button
              v-if="searchQuery"
              type="button"
              class="absolute right-2 p-1 text-slate-400 hover:text-slate-600 rounded"
              title="Bersihkan Pencarian"
              @click="clearSearch"
            >
              <X class="w-3.5 h-3.5" />
            </button>
          </div>

          <!-- Patient Search Autocomplete Dropdown -->
          <div
            v-if="isSearchDropdownOpen"
            class="absolute left-0 right-0 top-full mt-1.5 bg-white border border-slate-200 rounded-lg shadow-xl py-1 z-50 max-h-72 overflow-y-auto"
          >
            <div v-if="isSearching" class="px-3 py-2 text-xs text-slate-500 flex items-center gap-2">
              <span class="w-3 h-3 border-2 border-[#0284C7] border-t-transparent rounded-full animate-spin"></span>
              <span>Mencari data pasien...</span>
            </div>

            <div v-else-if="searchResults.length === 0" class="px-3 py-2 text-xs text-slate-500">
              Tidak ada pasien yang cocok dengan "{{ searchQuery }}"
            </div>

            <div v-else class="divide-y divide-slate-100">
              <div
                v-for="patient in searchResults"
                :key="patient.id"
                class="px-3 py-2 hover:bg-[#F0F9FF] cursor-pointer transition-colors flex items-center justify-between text-left group"
                @click="handleSelectPatient(patient)"
              >
                <div class="flex items-center gap-2.5 truncate">
                  <div class="w-7 h-7 rounded-full bg-[#E0F2FE] text-[#0369A1] font-bold text-xs flex items-center justify-center shrink-0">
                    {{ patient.nama_pasien?.charAt(0) || 'P' }}
                  </div>
                  <div class="truncate">
                    <span class="text-xs font-semibold text-slate-800 block truncate group-hover:text-[#0369A1]">
                      {{ patient.nama_pasien }}
                    </span>
                    <span class="text-[10px] text-slate-500 block truncate">
                      NIK: {{ patient.nik || '-' }} • {{ patient.jenis_kelamin === 'L' ? 'Laki-laki' : 'Perempuan' }}
                    </span>
                  </div>
                </div>

                <span class="font-mono text-[11px] font-bold text-[#0369A1] bg-[#E0F2FE] px-2 py-0.5 rounded shrink-0 border border-[#BAE6FD]">
                  {{ patient.no_rm }}
                </span>
              </div>
            </div>

            <div class="p-1.5 bg-slate-50 border-t border-slate-100 text-center">
              <button
                type="button"
                class="text-[11px] font-medium text-[#0284C7] hover:underline"
                @click="handleSearchSubmit"
              >
                Lihat semua hasil di Daftar Pasien &rarr;
              </button>
            </div>
          </div>
        </div>

        <!-- Right: Today's Patient Stats & Profile Dropdown -->
        <div class="flex items-center gap-2 sm:gap-3 shrink-0">
          <!-- Daily Patient Stats: Selesai & Belum Selesai -->
          <div class="hidden lg:flex items-center gap-2">
            <!-- Badge Pasien Selesai Hari Ini -->
            <div
              class="flex items-center gap-1.5 px-2.5 py-1 bg-emerald-50 text-emerald-700 border border-emerald-200/80 rounded-full text-xs font-medium"
              title="Total Pasien Selesai Hari Ini"
            >
              <CheckCircle2 class="w-3.5 h-3.5 text-emerald-600 shrink-0" />
              <span class="font-bold tabular-nums">{{ totalSelesai }}</span>
              <span class="text-emerald-800">Selesai</span>
            </div>

            <!-- Badge Pasien Belum Selesai Hari Ini -->
            <div
              class="flex items-center gap-1.5 px-2.5 py-1 bg-amber-50 text-amber-700 border border-amber-200/80 rounded-full text-xs font-medium"
              title="Total Pasien Belum Selesai (Dalam Antrean/Pemeriksaan) Hari Ini"
            >
              <Clock3 class="w-3.5 h-3.5 text-amber-600 shrink-0" />
              <span class="font-bold tabular-nums">{{ totalBelumSelesai }}</span>
              <span class="text-amber-800">Belum Selesai</span>
            </div>
          </div>

          <div class="h-4 w-px bg-slate-200 hidden lg:block"></div>

          <!-- Profile Dropdown (Avatar + Dropdown Menu) -->
          <div ref="profileDropdownRef" class="relative">
            <button
              type="button"
              class="flex items-center gap-2 p-1 sm:px-2 sm:py-1 rounded-lg hover:bg-slate-100 text-slate-700 transition-colors focus:outline-none"
              title="Profil Pengguna & Opsi"
              @click="toggleProfileDropdown"
            >
              <div class="w-7 h-7 sm:w-8 sm:h-8 rounded-full bg-[#0284C7] text-white font-bold text-xs sm:text-sm flex items-center justify-center shadow-xs">
                {{ authStore.currentUserName?.charAt(0)?.toUpperCase() || 'U' }}
              </div>

              <div class="hidden md:block text-left text-xs leading-tight">
                <span class="font-semibold text-slate-800 block truncate max-w-[120px]">
                  {{ authStore.currentUserName }}
                </span>
                <span class="text-[10px] text-slate-500 font-medium block">
                  {{ authStore.currentRole }}
                </span>
              </div>

              <ChevronDown
                :class="[
                  'w-3.5 h-3.5 text-slate-400 transition-transform duration-200',
                  isProfileDropdownOpen ? 'rotate-180 text-slate-700' : ''
                ]"
              />
            </button>

            <!-- Dropdown Menu -->
            <div
              v-if="isProfileDropdownOpen"
              class="absolute right-0 top-full mt-2 w-56 bg-white border border-slate-200 rounded-lg shadow-xl py-1.5 z-50 text-left transition-all"
            >
              <!-- User Brief Summary Header -->
              <div class="px-3.5 py-2 border-b border-slate-100 bg-slate-50/70">
                <span class="text-xs font-bold text-slate-900 block truncate">
                  {{ authStore.currentUserName }}
                </span>
                <span class="text-[11px] text-slate-500 block truncate mb-1.5">
                  {{ authStore.user?.email || authStore.user?.username || '-' }}
                </span>
                <StatusBadge :status="authStore.currentRole" variant="info">
                  {{ authStore.currentRole }}
                </StatusBadge>
              </div>

              <!-- Menu Actions -->
              <div class="py-1">
                <!-- Setting Password -->
                <button
                  type="button"
                  class="w-full px-3.5 py-2 text-xs text-slate-700 hover:bg-slate-100 hover:text-slate-900 flex items-center gap-2.5 transition-colors text-left"
                  @click="openPasswordModal"
                >
                  <KeyRound class="w-4 h-4 text-[#0284C7]" />
                  <span>Setting Password</span>
                </button>

                <div class="my-1 border-t border-slate-100"></div>

                <!-- Log Out -->
                <button
                  type="button"
                  class="w-full px-3.5 py-2 text-xs text-red-600 hover:bg-red-50 flex items-center gap-2.5 transition-colors text-left font-medium"
                  @click="handleLogout"
                >
                  <LogOut class="w-4 h-4 text-red-500" />
                  <span>Log Out</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </header>

      <!-- Page Content View -->
      <main class="flex-1 p-4 lg:p-6 overflow-x-hidden">
        <router-view />
      </main>
    </div>
  </div>
</template>
