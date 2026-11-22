<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useQueueStore } from '@/stores/queue'
import { api } from '@/api/client'
import BaseButton from '@/components/common/BaseButton.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import {
  LayoutDashboard, Users, UserPlus, Stethoscope, Activity, Pill,
  FlaskConical, Boxes, Receipt, CalendarClock, TrendingUp, AlertTriangle,
  Clock, CheckCircle2, ShieldCheck, ArrowRight, ArrowUpRight, FileText,
  DollarSign, Package, Bell, RefreshCw, FileSpreadsheet, HeartPulse, Sparkles
} from 'lucide-vue-next'
import type { Visit, Invoice, InventoryItem, Appointment } from '@/types'

const router = useRouter()
const authStore = useAuthStore()
const queueStore = useQueueStore()

const isLoading = ref(false)

// For Superadmin, allow previewing dashboards of other roles
const selectedRolePreview = ref<string>('')

const effectiveRole = computed(() => {
  if (authStore.hasRole('SUPERADMIN') && selectedRolePreview.value) {
    return selectedRolePreview.value
  }
  return authStore.user?.role_code || 'SUPERADMIN'
})

// Data state
const visits = ref<Visit[]>([])
const invoices = ref<Invoice[]>([])
const inventoryItems = ref<InventoryItem[]>([])
const appointments = ref<Appointment[]>([])
const activeShiftData = ref<any>(null)

// Current formatted date
const currentDate = computed(() => {
  return new Date().toLocaleDateString('id-ID', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    year: 'numeric'
  })
})

async function fetchDashboardMetrics() {
  isLoading.value = true
  try {
    await queueStore.fetchQueueMetrics()

    const [vRes, invRes, itemRes, apptRes, shiftRes] = await Promise.allSettled([
      api.get<Visit[]>('/visits', { limit: 100 }),
      api.get<Invoice[]>('/billing/invoices', { limit: 100 }),
      api.get<InventoryItem[]>('/logistics/items', { limit: 100 }),
      api.get<Appointment[]>('/appointments', { limit: 100 }),
      api.get('/billing/shifts/active')
    ])

    if (vRes.status === 'fulfilled' && vRes.value.data) visits.value = vRes.value.data
    if (invRes.status === 'fulfilled' && invRes.value.data) invoices.value = invRes.value.data
    if (itemRes.status === 'fulfilled' && itemRes.value.data) inventoryItems.value = itemRes.value.data
    if (apptRes.status === 'fulfilled' && apptRes.value.data) appointments.value = apptRes.value.data
    if (shiftRes.status === 'fulfilled' && shiftRes.value.data) activeShiftData.value = shiftRes.value.data
  } catch (err) {
    console.error('Failed to load dashboard monitoring metrics:', err)
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  fetchDashboardMetrics()
})

// Metrics computed per role
const roleMetrics = computed(() => {
  const allVisits = visits.value
  const allInvoices = invoices.value
  const allItems = inventoryItems.value

  const totalPatientsToday = allVisits.length
  const finishedVisits = allVisits.filter(v => v.status === 'SELESAI').length
  const waitingVisits = allVisits.filter(v => v.status === 'MENUNGGU_TRIASE' || v.status === 'MENUNGGU_DOKTER').length
  const doctorExamining = allVisits.filter(v => v.status === 'DIPERIKSA_DOKTER').length
  const inPharmacy = allVisits.filter(v => v.status === 'FARMASI').length
  const inCashier = allVisits.filter(v => v.status === 'KASIR').length

  const totalRevenue = allInvoices.reduce((acc, inv) => acc + (inv.total_terbayar || 0), 0)
  const pendingInvoices = allInvoices.filter(inv => inv.status === 'BELUM_BAYAR').length
  const totalReceivable = allInvoices.reduce((acc, inv) => acc + (inv.sisa_tagihan || 0), 0)

  const lowStockCount = allItems.filter(i => i.total_stock <= i.minimum_stock).length
  const totalInventoryValuation = allItems.reduce((acc, i) => acc + (i.total_stock * i.buy_price), 0)

  return {
    totalPatientsToday,
    finishedVisits,
    waitingVisits,
    doctorExamining,
    inPharmacy,
    inCashier,
    totalRevenue,
    pendingInvoices,
    totalReceivable,
    lowStockCount,
    totalInventoryValuation
  }
})
</script>

<template>
  <div class="w-full max-w-full space-y-4 text-left">
    <!-- Role Switcher for Superadmin Demo / Testing -->
    <div v-if="authStore.hasRole('SUPERADMIN')" class="bg-[#F8FAFC] border border-[#CBD5E1] rounded-[8px] p-3 flex flex-wrap items-center justify-between gap-3 text-xs">
      <div class="flex items-center gap-2">
        <Sparkles class="w-4 h-4 text-[#0284C7]" />
        <span class="font-bold text-[#0F172A]">Mode Preview Role Dashboard (Superadmin Tool):</span>
      </div>
      <div class="flex flex-wrap items-center gap-1.5">
        <button
          v-for="r in [
            { code: '', label: 'Executive Superadmin' },
            { code: 'REGISTRATION', label: 'Pendaftaran' },
            { code: 'CALL_CENTER', label: 'Call Center' },
            { code: 'NURSE', label: 'Perawat / Triase' },
            { code: 'DOCTOR', label: 'Dokter Spesialis' },
            { code: 'PHARMACY', label: 'Farmasi' },
            { code: 'SUPPORT', label: 'Penunjang Lab/Rad' },
            { code: 'LOGISTICS', label: 'Logistik Medis' },
            { code: 'CASHIER', label: 'Kasir & Billing' }
          ]"
          :key="r.code"
          type="button"
          :class="[
            'px-2.5 py-1 rounded-[6px] font-semibold transition-all',
            selectedRolePreview === r.code
              ? 'bg-[#0284C7] text-white shadow-xs'
              : 'bg-white border border-[#E2E8F0] text-[#475569] hover:bg-slate-50'
          ]"
          @click="selectedRolePreview = r.code"
        >
          {{ r.label }}
        </button>
      </div>
    </div>

    <!-- ========================================================================= -->
    <!-- WELCOME HERO BANNER                                                       -->
    <!-- ========================================================================= -->
    <div class="bg-gradient-to-r from-[#0369A1] via-[#0284C7] to-[#0EA5E9] rounded-[10px] p-5 text-white shadow-md relative overflow-hidden">
      <div class="relative z-10 flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div class="space-y-1">
          <div class="flex items-center gap-2">
            <span class="px-2 py-0.5 rounded-full bg-white/20 text-white text-[11px] font-bold backdrop-blur-xs flex items-center gap-1">
              <span class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
              Sistem Aktif & Terintegrasi
            </span>
            <span class="text-white/80 text-[12px] font-medium">{{ currentDate }}</span>
          </div>

          <h1 class="text-[22px] font-extrabold tracking-tight">
            Selamat Datang, {{ authStore.user?.nama_lengkap || authStore.user?.username }}!
          </h1>

          <p class="text-[13px] text-white/90 max-w-2xl">
            <span v-if="effectiveRole === 'SUPERADMIN'">
              Pusat kendali eksekutif operasional klinik dan rumah sakit. Pantau indikator kinerja seluruh unit pelayanan, antrean real-time, dan pendapatan.
            </span>
            <span v-else-if="effectiveRole === 'REGISTRATION'">
              Dashboard pelayanan pendaftaran pasien dan manajemen loket antrean poliklinik rawat jalan.
            </span>
            <span v-else-if="effectiveRole === 'CALL_CENTER'">
              Pusat konfirmasi janji temu (appointment), reservasi pasien poliklinik, dan koordinasi jadwal dokter.
            </span>
            <span v-else-if="effectiveRole === 'NURSE'">
              Stasiun kerja perawat poliklinik, monitoring triase gawat darurat, dan kepatuhan pengkajian tanda-tanda vital (TTV).
            </span>
            <span v-else-if="effectiveRole === 'DOCTOR'">
              Ruang kerja dokter spesialis, antrean pemeriksaan pasien aktif, telaah e-Resep, dan pengisian resume medis elektronik.
            </span>
            <span v-else-if="effectiveRole === 'PHARMACY'">
              Instalasi farmasi rawat jalan, telaah kelayakan klinis resep dokter, penyiapan obat, dan monitoring persediaan depo.
            </span>
            <span v-else-if="effectiveRole === 'SUPPORT'">
              Instalasi penunjang medis diagnostik (Laboratorium Medis, Radiologi X-Ray, Fisioterapi & EKG).
            </span>
            <span v-else-if="effectiveRole === 'LOGISTICS'">
              Gudang medis sentral, manajemen Purchase Order (PO), penerimaan barang (GRN), dan kartu mutasi persediaan.
            </span>
            <span v-else-if="effectiveRole === 'CASHIER'">
              Loket kasir rawat jalan, penyelesaian pembayaran invoice tagihan, dan rekonsiliasi kas shift berjalan.
            </span>
          </p>
        </div>

        <div class="flex items-center gap-2 self-start md:self-auto">
          <BaseButton variant="outline" size="sm" class="bg-white/10 hover:bg-white/20 text-white border-white/30" @click="fetchDashboardMetrics">
            <RefreshCw class="w-3.5 h-3.5 mr-1" />
            Refresh Data
          </BaseButton>
        </div>
      </div>
    </div>

    <!-- ========================================================================= -->
    <!-- 1. DASHBOARD SUPERADMIN (EXECUTIVE COMMAND CENTER)                         -->
    <!-- ========================================================================= -->
    <div v-if="effectiveRole === 'SUPERADMIN'" class="space-y-4">
      <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs">
          <span class="text-[11px] font-semibold text-[#64748B] uppercase tracking-wider block">Kunjungan Pasien Hari Ini</span>
          <div class="text-[22px] font-bold text-[#0F172A] mt-1">{{ roleMetrics.totalPatientsToday }} Pasien</div>
          <span class="text-[11px] text-emerald-600 font-semibold mt-0.5 block flex items-center gap-1">
            <CheckCircle2 class="w-3.5 h-3.5" /> {{ roleMetrics.finishedVisits }} telah selesai
          </span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs">
          <span class="text-[11px] font-semibold text-[#0284C7] uppercase tracking-wider block">Total Penerimaan Kas</span>
          <div class="text-[20px] font-bold text-[#0284C7] font-mono mt-1">
            Rp {{ roleMetrics.totalRevenue.toLocaleString('id-ID') }}
          </div>
          <span class="text-[11px] text-[#0284C7] mt-0.5 block">Pendapatan tagihan terbayar</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs">
          <span class="text-[11px] font-semibold text-amber-700 uppercase tracking-wider block">Pasien Dalam Proses</span>
          <div class="text-[22px] font-bold text-amber-700 mt-1">{{ roleMetrics.waitingVisits + roleMetrics.doctorExamining }} Pasien</div>
          <span class="text-[11px] text-amber-600 mt-0.5 block">Triase & Ruang Poli</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs">
          <span class="text-[11px] font-semibold text-rose-700 uppercase tracking-wider block">Stok Farmasi Menipis</span>
          <div class="text-[22px] font-bold text-rose-700 mt-1">{{ roleMetrics.lowStockCount }} Item</div>
          <span class="text-[11px] text-rose-600 mt-0.5 block">Di bawah batas minimum</span>
        </div>
      </div>

      <!-- Quick Action Launchers -->
      <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs space-y-3">
        <h3 class="text-[13px] font-bold text-[#0F172A]">Akses Cepat Modul Utama SIMRS</h3>
        <div class="grid grid-cols-2 md:grid-cols-4 gap-2.5">
          <router-link to="/reports/financial" class="p-3 rounded-[6px] border border-[#E2E8F0] bg-[#F8FAFC] hover:bg-[#F0F9FF] hover:border-[#0284C7] transition-all flex items-center justify-between">
            <div class="flex items-center gap-2.5">
              <TrendingUp class="w-4 h-4 text-[#0284C7]" />
              <span class="text-xs font-bold text-[#0F172A]">Laporan Keuangan</span>
            </div>
            <ArrowRight class="w-3.5 h-3.5 text-slate-400" />
          </router-link>

          <router-link to="/reports/pharmacy-stock" class="p-3 rounded-[6px] border border-[#E2E8F0] bg-[#F8FAFC] hover:bg-[#F0F9FF] hover:border-[#0284C7] transition-all flex items-center justify-between">
            <div class="flex items-center gap-2.5">
              <Boxes class="w-4 h-4 text-[#0284C7]" />
              <span class="text-xs font-bold text-[#0F172A]">Laporan Stok Farmasi</span>
            </div>
            <ArrowRight class="w-3.5 h-3.5 text-slate-400" />
          </router-link>

          <router-link to="/reports/patient-registrations" class="p-3 rounded-[6px] border border-[#E2E8F0] bg-[#F8FAFC] hover:bg-[#F0F9FF] hover:border-[#0284C7] transition-all flex items-center justify-between">
            <div class="flex items-center gap-2.5">
              <Users class="w-4 h-4 text-[#0284C7]" />
              <span class="text-xs font-bold text-[#0F172A]">Laporan Kunjungan</span>
            </div>
            <ArrowRight class="w-3.5 h-3.5 text-slate-400" />
          </router-link>

          <router-link to="/reports/demographics" class="p-3 rounded-[6px] border border-[#E2E8F0] bg-[#F8FAFC] hover:bg-[#F0F9FF] hover:border-[#0284C7] transition-all flex items-center justify-between">
            <div class="flex items-center gap-2.5">
              <PieChart class="w-4 h-4 text-[#0284C7]" />
              <span class="text-xs font-bold text-[#0F172A]">Laporan Demografi</span>
            </div>
            <ArrowRight class="w-3.5 h-3.5 text-slate-400" />
          </router-link>
        </div>
      </div>
    </div>

    <!-- ========================================================================= -->
    <!-- 2. DASHBOARD REGISTRATION (FRONT OFFICE)                                   -->
    <!-- ========================================================================= -->
    <div v-else-if="effectiveRole === 'REGISTRATION'" class="space-y-4">
      <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs">
          <span class="text-[11px] font-semibold text-[#64748B] uppercase tracking-wider block">Registrasi Hari Ini</span>
          <div class="text-[22px] font-bold text-[#0F172A] mt-1">{{ roleMetrics.totalPatientsToday }} Pasien</div>
          <span class="text-[11px] text-[#0284C7] font-semibold mt-0.5 block">Tercatat di loket antrean</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs">
          <span class="text-[11px] font-semibold text-amber-700 uppercase tracking-wider block">Menunggu Triase</span>
          <div class="text-[22px] font-bold text-amber-700 mt-1">{{ roleMetrics.waitingVisits }} Pasien</div>
          <span class="text-[11px] text-amber-600 mt-0.5 block">Di ruang tunggu</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs">
          <span class="text-[11px] font-semibold text-emerald-700 uppercase tracking-wider block">Pelayanan Selesai</span>
          <div class="text-[22px] font-bold text-emerald-700 mt-1">{{ roleMetrics.finishedVisits }} Pasien</div>
          <span class="text-[11px] text-emerald-600 mt-0.5 block">Pasien telah pulang</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs">
          <span class="text-[11px] font-semibold text-[#0284C7] uppercase tracking-wider block">Total Antrean Aktif</span>
          <div class="text-[22px] font-bold text-[#0284C7] mt-1">{{ visits.filter(v => v.status !== 'SELESAI' && v.status !== 'BATAL').length }} Pasien</div>
          <span class="text-[11px] text-[#0284C7] mt-0.5 block">Dalam alur pelayanan</span>
        </div>
      </div>

      <!-- Quick Action Buttons -->
      <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs space-y-3">
        <div class="flex items-center justify-between">
          <h3 class="text-[13px] font-bold text-[#0F172A]">Tindakan Cepat Pendaftaran</h3>
          <router-link to="/registration" class="text-xs font-semibold text-[#0284C7] hover:underline flex items-center gap-1">
            Buka Loket Pendaftaran <ArrowRight class="w-3.5 h-3.5" />
          </router-link>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
          <router-link to="/registration" class="p-3.5 rounded-[8px] border border-[#BAE6FD] bg-[#F0F9FF] hover:bg-[#E0F2FE] transition-all flex items-center gap-3">
            <div class="w-10 h-10 rounded-full bg-[#0284C7] text-white flex items-center justify-center font-bold">
              <UserPlus class="w-5 h-5" />
            </div>
            <div>
              <div class="text-xs font-bold text-[#0F172A]">Daftarkan Pasien Baru</div>
              <div class="text-[11px] text-slate-500">Input NIK, data identitas & cetak No. RM</div>
            </div>
          </router-link>

          <router-link to="/registration/patients" class="p-3.5 rounded-[8px] border border-[#E2E8F0] bg-white hover:bg-slate-50 transition-all flex items-center gap-3">
            <div class="w-10 h-10 rounded-full bg-slate-100 text-slate-700 flex items-center justify-center font-bold">
              <Users class="w-5 h-5" />
            </div>
            <div>
              <div class="text-xs font-bold text-[#0F172A]">Cari Database Pasien</div>
              <div class="text-[11px] text-slate-500">Pencarian No. RM, NIK, dan riwayat</div>
            </div>
          </router-link>

          <router-link to="/reports/patient-registrations" class="p-3.5 rounded-[8px] border border-[#E2E8F0] bg-white hover:bg-slate-50 transition-all flex items-center gap-3">
            <div class="w-10 h-10 rounded-full bg-emerald-50 text-emerald-700 flex items-center justify-center font-bold">
              <FileSpreadsheet class="w-5 h-5" />
            </div>
            <div>
              <div class="text-xs font-bold text-[#0F172A]">Laporan Kunjungan</div>
              <div class="text-[11px] text-slate-500">Rekapitulasi kunjungan & download Excel</div>
            </div>
          </router-link>
        </div>
      </div>
    </div>

    <!-- ========================================================================= -->
    <!-- 3. DASHBOARD NURSE (PERAWAT & TRIASE)                                      -->
    <!-- ========================================================================= -->
    <div v-else-if="effectiveRole === 'NURSE'" class="space-y-4">
      <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs">
          <span class="text-[11px] font-semibold text-amber-700 uppercase tracking-wider block">Menunggu Triase & TTV</span>
          <div class="text-[22px] font-bold text-amber-700 mt-1">
            {{ visits.filter(v => v.status === 'MENUNGGU_TRIASE').length }} Pasien
          </div>
          <span class="text-[11px] text-amber-600 mt-0.5 block">Perlu pengukuran tanda vital</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs">
          <span class="text-[11px] font-semibold text-emerald-700 uppercase tracking-wider block">Siap Diperiksa Dokter</span>
          <div class="text-[22px] font-bold text-emerald-700 mt-1">
            {{ visits.filter(v => v.status === 'MENUNGGU_DOKTER').length }} Pasien
          </div>
          <span class="text-[11px] text-emerald-600 mt-0.5 block">TTV telah lengkap</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs">
          <span class="text-[11px] font-semibold text-[#0284C7] uppercase tracking-wider block">Sedang Diperiksa</span>
          <div class="text-[22px] font-bold text-[#0284C7] mt-1">{{ roleMetrics.doctorExamining }} Pasien</div>
          <span class="text-[11px] text-[#0284C7] mt-0.5 block">Di dalam ruang poli</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs">
          <span class="text-[11px] font-semibold text-slate-700 uppercase tracking-wider block">Total Pasien Poli Hari Ini</span>
          <div class="text-[22px] font-bold text-slate-800 mt-1">{{ roleMetrics.totalPatientsToday }} Pasien</div>
          <span class="text-[11px] text-slate-500 mt-0.5 block">Rekap harian keperawatan</span>
        </div>
      </div>

      <!-- Quick Action Buttons -->
      <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs space-y-3">
        <div class="flex items-center justify-between">
          <h3 class="text-[13px] font-bold text-[#0F172A]">Stasiun Kerja Keperawatan</h3>
          <router-link to="/nurse" class="text-xs font-semibold text-[#0284C7] hover:underline flex items-center gap-1">
            Buka Ruang Triase <ArrowRight class="w-3.5 h-3.5" />
          </router-link>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
          <router-link to="/nurse" class="p-3.5 rounded-[8px] border border-[#BAE6FD] bg-[#F0F9FF] hover:bg-[#E0F2FE] transition-all flex items-center gap-3">
            <div class="w-10 h-10 rounded-full bg-[#0284C7] text-white flex items-center justify-center font-bold">
              <Activity class="w-5 h-5" />
            </div>
            <div>
              <div class="text-xs font-bold text-[#0F172A]">Input TTV & Asesmen Keperawatan</div>
              <div class="text-[11px] text-slate-500">Ukur tensi, nadi, suhu, SpO2, skala nyeri & risiko jatuh</div>
            </div>
          </router-link>

          <router-link to="/reports/demographics" class="p-3.5 rounded-[8px] border border-[#E2E8F0] bg-white hover:bg-slate-50 transition-all flex items-center gap-3">
            <div class="w-10 h-10 rounded-full bg-slate-100 text-slate-700 flex items-center justify-center font-bold">
              <PieChart class="w-5 h-5" />
            </div>
            <div>
              <div class="text-xs font-bold text-[#0F172A]">Laporan Demografi Pasien</div>
              <div class="text-[11px] text-slate-500">Statistik usia anak, gender, dan riwayat alergi</div>
            </div>
          </router-link>
        </div>
      </div>
    </div>

    <!-- ========================================================================= -->
    <!-- 4. DASHBOARD DOCTOR (DOKTER SPESIALIS)                                     -->
    <!-- ========================================================================= -->
    <div v-else-if="effectiveRole === 'DOCTOR'" class="space-y-4">
      <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs">
          <span class="text-[11px] font-semibold text-[#0284C7] uppercase tracking-wider block">Menunggu Pemeriksaan</span>
          <div class="text-[22px] font-bold text-[#0284C7] mt-1">
            {{ visits.filter(v => v.status === 'MENUNGGU_DOKTER').length }} Pasien
          </div>
          <span class="text-[11px] text-[#0284C7] mt-0.5 block">Siap masuk ruang periksa</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs">
          <span class="text-[11px] font-semibold text-emerald-700 uppercase tracking-wider block">Selesai Diperiksa Hari Ini</span>
          <div class="text-[22px] font-bold text-emerald-700 mt-1">{{ roleMetrics.finishedVisits }} Pasien</div>
          <span class="text-[11px] text-emerald-600 mt-0.5 block">SOAP & e-Resep lengkap</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs">
          <span class="text-[11px] font-semibold text-amber-700 uppercase tracking-wider block">Di Farmasi & Kasir</span>
          <div class="text-[22px] font-bold text-amber-700 mt-1">
            {{ roleMetrics.inPharmacy + roleMetrics.inCashier }} Pasien
          </div>
          <span class="text-[11px] text-amber-600 mt-0.5 block">Menunggu obat & pembayaran</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs">
          <span class="text-[11px] font-semibold text-slate-700 uppercase tracking-wider block">Total Antrean Poli Hari Ini</span>
          <div class="text-[22px] font-bold text-slate-800 mt-1">{{ roleMetrics.totalPatientsToday }} Pasien</div>
          <span class="text-[11px] text-slate-500 mt-0.5 block">Beban konsultasi poliklinik</span>
        </div>
      </div>

      <!-- Quick Action Buttons -->
      <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs space-y-3">
        <div class="flex items-center justify-between">
          <h3 class="text-[13px] font-bold text-[#0F172A]">Workspace Klinis Dokter</h3>
          <router-link to="/doctor" class="text-xs font-semibold text-[#0284C7] hover:underline flex items-center gap-1">
            Buka Antrean & EMR Dokter <ArrowRight class="w-3.5 h-3.5" />
          </router-link>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
          <router-link to="/doctor" class="p-3.5 rounded-[8px] border border-[#BAE6FD] bg-[#F0F9FF] hover:bg-[#E0F2FE] transition-all flex items-center gap-3">
            <div class="w-10 h-10 rounded-full bg-[#0284C7] text-white flex items-center justify-center font-bold">
              <Stethoscope class="w-5 h-5" />
            </div>
            <div>
              <div class="text-xs font-bold text-[#0F172A]">Buka Rekam Medis Elektronik (EMR)</div>
              <div class="text-[11px] text-slate-500">Panggil pasien, isi SOAP, input ICD-10/9 & e-Resep</div>
            </div>
          </router-link>

          <router-link to="/reports/demographics" class="p-3.5 rounded-[8px] border border-[#E2E8F0] bg-white hover:bg-slate-50 transition-all flex items-center gap-3">
            <div class="w-10 h-10 rounded-full bg-slate-100 text-slate-700 flex items-center justify-center font-bold">
              <PieChart class="w-5 h-5" />
            </div>
            <div>
              <div class="text-xs font-bold text-[#0F172A]">Laporan Demografi & Diagnosa Pasien</div>
              <div class="text-[11px] text-slate-500">Lihat sebaran usia, jenis kelamin, dan prevalensi penyakit</div>
            </div>
          </router-link>
        </div>
      </div>
    </div>

    <!-- ========================================================================= -->
    <!-- 5. DASHBOARD PHARMACY (FARMASI & OBAT)                                     -->
    <!-- ========================================================================= -->
    <div v-else-if="effectiveRole === 'PHARMACY'" class="space-y-4">
      <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs">
          <span class="text-[11px] font-semibold text-amber-700 uppercase tracking-wider block">Resep Masuk Hari Ini</span>
          <div class="text-[22px] font-bold text-amber-700 mt-1">{{ roleMetrics.inPharmacy }} Resep</div>
          <span class="text-[11px] text-amber-600 mt-0.5 block">Menunggu telaah klinis & dispensing</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs">
          <span class="text-[11px] font-semibold text-rose-700 uppercase tracking-wider block">Stok Obat Kritis (&lt; Min)</span>
          <div class="text-[22px] font-bold text-rose-700 mt-1">{{ roleMetrics.lowStockCount }} Item</div>
          <span class="text-[11px] text-rose-600 mt-0.5 block">Perlu order ke PBF / gudang</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs">
          <span class="text-[11px] font-semibold text-[#0284C7] uppercase tracking-wider block">Total Jenis Obat Aktif</span>
          <div class="text-[22px] font-bold text-[#0284C7] mt-1">{{ inventoryItems.length }} Sediaan</div>
          <span class="text-[11px] text-[#0284C7] mt-0.5 block">Formularium klinik</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs">
          <span class="text-[11px] font-semibold text-emerald-700 uppercase tracking-wider block">Valuasi Persediaan</span>
          <div class="text-[18px] font-bold text-emerald-700 font-mono mt-1">
            Rp {{ roleMetrics.totalInventoryValuation.toLocaleString('id-ID') }}
          </div>
          <span class="text-[11px] text-emerald-600 mt-0.5 block">Aset persediaan obat</span>
        </div>
      </div>

      <!-- Quick Action Buttons -->
      <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs space-y-3">
        <div class="flex items-center justify-between">
          <h3 class="text-[13px] font-bold text-[#0F172A]">Tindakan Cepat Farmasi</h3>
          <router-link to="/pharmacy" class="text-xs font-semibold text-[#0284C7] hover:underline flex items-center gap-1">
            Buka Antrean Resep <ArrowRight class="w-3.5 h-3.5" />
          </router-link>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
          <router-link to="/pharmacy" class="p-3.5 rounded-[8px] border border-[#BAE6FD] bg-[#F0F9FF] hover:bg-[#E0F2FE] transition-all flex items-center gap-3">
            <div class="w-10 h-10 rounded-full bg-[#0284C7] text-white flex items-center justify-center font-bold">
              <Pill class="w-5 h-5" />
            </div>
            <div>
              <div class="text-xs font-bold text-[#0F172A]">Verifikasi Resep & Dispensing Obat</div>
              <div class="text-[11px] text-slate-500">Telaah klinis, potong stok otomatis, dan cetak etiket</div>
            </div>
          </router-link>

          <router-link to="/reports/pharmacy-stock" class="p-3.5 rounded-[8px] border border-[#E2E8F0] bg-white hover:bg-slate-50 transition-all flex items-center gap-3">
            <div class="w-10 h-10 rounded-full bg-emerald-50 text-emerald-700 flex items-center justify-center font-bold">
              <Boxes class="w-5 h-5" />
            </div>
            <div>
              <div class="text-xs font-bold text-[#0F172A]">Laporan Stok Farmasi & Kartu Mutasi</div>
              <div class="text-[11px] text-slate-500">Pantau valuasi aset obat & download Excel</div>
            </div>
          </router-link>
        </div>
      </div>
    </div>

    <!-- ========================================================================= -->
    <!-- 6. DASHBOARD CASHIER (KASIR & BILLING)                                     -->
    <!-- ========================================================================= -->
    <div v-else-if="effectiveRole === 'CASHIER'" class="space-y-4">
      <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs">
          <span class="text-[11px] font-semibold text-[#0284C7] uppercase tracking-wider block">Penerimaan Kas Lunas</span>
          <div class="text-[20px] font-bold text-[#0284C7] font-mono mt-1">
            Rp {{ roleMetrics.totalRevenue.toLocaleString('id-ID') }}
          </div>
          <span class="text-[11px] text-[#0284C7] mt-0.5 block">Total pembayaran lunas hari ini</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs">
          <span class="text-[11px] font-semibold text-amber-700 uppercase tracking-wider block">Tagihan Belum Lunas</span>
          <div class="text-[22px] font-bold text-amber-700 mt-1">{{ roleMetrics.pendingInvoices }} Invoice</div>
          <span class="text-[11px] text-amber-600 mt-0.5 block">Pasien dalam antrean kasir</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs">
          <span class="text-[11px] font-semibold text-rose-700 uppercase tracking-wider block">Total Piutang Berjalan</span>
          <div class="text-[20px] font-bold text-rose-700 font-mono mt-1">
            Rp {{ roleMetrics.totalReceivable.toLocaleString('id-ID') }}
          </div>
          <span class="text-[11px] text-rose-600 mt-0.5 block">Sisa tagihan belum dibayar</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs">
          <span class="text-[11px] font-semibold text-emerald-700 uppercase tracking-wider block">Status Sesi Shift</span>
          <div class="text-[16px] font-bold text-emerald-700 mt-1">
            {{ activeShiftData ? activeShiftData.shift_name : 'Shift Aktif' }}
          </div>
          <span class="text-[11px] text-emerald-600 mt-0.5 block">Loket kasir siap memproses</span>
        </div>
      </div>

      <!-- Quick Action Buttons -->
      <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs space-y-3">
        <div class="flex items-center justify-between">
          <h3 class="text-[13px] font-bold text-[#0F172A]">Tindakan Cepat Kasir</h3>
          <router-link to="/cashier" class="text-xs font-semibold text-[#0284C7] hover:underline flex items-center gap-1">
            Buka Meja Kasir <ArrowRight class="w-3.5 h-3.5" />
          </router-link>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
          <router-link to="/cashier" class="p-3.5 rounded-[8px] border border-[#BAE6FD] bg-[#F0F9FF] hover:bg-[#E0F2FE] transition-all flex items-center gap-3">
            <div class="w-10 h-10 rounded-full bg-[#0284C7] text-white flex items-center justify-center font-bold">
              <Receipt class="w-5 h-5" />
            </div>
            <div>
              <div class="text-xs font-bold text-[#0F172A]">Proses Pembayaran & Kuitansi</div>
              <div class="text-[11px] text-slate-500">Pelunasan Tunai, QRIS, EDC, cetak kuitansi resmi</div>
            </div>
          </router-link>

          <router-link to="/reports/financial" class="p-3.5 rounded-[8px] border border-[#E2E8F0] bg-white hover:bg-slate-50 transition-all flex items-center gap-3">
            <div class="w-10 h-10 rounded-full bg-emerald-50 text-emerald-700 flex items-center justify-center font-bold">
              <TrendingUp class="w-5 h-5" />
            </div>
            <div>
              <div class="text-xs font-bold text-[#0F172A]">Laporan Keuangan & Rekonsiliasi Kasir</div>
              <div class="text-[11px] text-slate-500">Rincian pendapatan per penjamin & unduh Excel</div>
            </div>
          </router-link>
        </div>
      </div>
    </div>

    <!-- ========================================================================= -->
    <!-- 7. DASHBOARD DEFAULT / OTHER ROLES (LOGISTICS, SUPPORT, CALL CENTER)       -->
    <!-- ========================================================================= -->
    <div v-else class="space-y-4">
      <div class="grid grid-cols-2 md:grid-cols-3 gap-3">
        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs">
          <span class="text-[11px] font-semibold text-[#64748B] uppercase tracking-wider block">Kunjungan Hari Ini</span>
          <div class="text-[22px] font-bold text-[#0F172A] mt-1">{{ roleMetrics.totalPatientsToday }} Pasien</div>
          <span class="text-[11px] text-emerald-600 mt-0.5 block">{{ roleMetrics.finishedVisits }} telah selesai</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs">
          <span class="text-[11px] font-semibold text-[#0284C7] uppercase tracking-wider block">Total Persediaan Medis</span>
          <div class="text-[22px] font-bold text-[#0284C7] mt-1">{{ inventoryItems.length }} Item</div>
          <span class="text-[11px] text-[#0284C7] mt-0.5 block">Gudang & Depo</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs">
          <span class="text-[11px] font-semibold text-amber-700 uppercase tracking-wider block">Janji Temu Terdaftar</span>
          <div class="text-[22px] font-bold text-amber-700 mt-1">{{ appointments.length }} Appointment</div>
          <span class="text-[11px] text-amber-600 mt-0.5 block">Pasien terkonfirmasi</span>
        </div>
      </div>

      <!-- Quick Action Buttons -->
      <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs space-y-3">
        <h3 class="text-[13px] font-bold text-[#0F172A]">Akses Cepat Layanan</h3>
        <div class="grid grid-cols-2 md:grid-cols-4 gap-2.5">
          <router-link to="/reports/financial" class="p-3 rounded-[6px] border border-[#E2E8F0] bg-[#F8FAFC] hover:bg-[#F0F9FF] hover:border-[#0284C7] transition-all flex items-center justify-between">
            <span class="text-xs font-bold text-[#0F172A]">Laporan Keuangan</span>
            <ArrowRight class="w-3.5 h-3.5 text-slate-400" />
          </router-link>
          <router-link to="/reports/pharmacy-stock" class="p-3 rounded-[6px] border border-[#E2E8F0] bg-[#F8FAFC] hover:bg-[#F0F9FF] hover:border-[#0284C7] transition-all flex items-center justify-between">
            <span class="text-xs font-bold text-[#0F172A]">Laporan Stok</span>
            <ArrowRight class="w-3.5 h-3.5 text-slate-400" />
          </router-link>
          <router-link to="/reports/patient-registrations" class="p-3 rounded-[6px] border border-[#E2E8F0] bg-[#F8FAFC] hover:bg-[#F0F9FF] hover:border-[#0284C7] transition-all flex items-center justify-between">
            <span class="text-xs font-bold text-[#0F172A]">Laporan Pendaftaran</span>
            <ArrowRight class="w-3.5 h-3.5 text-slate-400" />
          </router-link>
          <router-link to="/reports/demographics" class="p-3 rounded-[6px] border border-[#E2E8F0] bg-[#F8FAFC] hover:bg-[#F0F9FF] hover:border-[#0284C7] transition-all flex items-center justify-between">
            <span class="text-xs font-bold text-[#0F172A]">Laporan Demografi</span>
            <ArrowRight class="w-3.5 h-3.5 text-slate-400" />
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>
