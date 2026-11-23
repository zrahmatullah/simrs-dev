<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '@/api/client'
import { useUiStore } from '@/stores/ui'
import { exportToExcel } from '@/utils/excel'
import DataTable, { type Column } from '@/components/common/DataTable.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseDatePicker from '@/components/common/BaseDatePicker.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import {
  TrendingUp, DollarSign, Boxes, Users, PieChart, FileSpreadsheet,
  Download, Filter, Calendar, Search, RefreshCw, AlertTriangle,
  ShieldCheck, CheckCircle2, Stethoscope, Activity, ArrowUpRight,
  Pill, Clock, Receipt, UserCheck, HeartPulse
} from 'lucide-vue-next'
import type { Invoice, InventoryItem, Visit, Patient, MasterPoli } from '@/types'

const route = useRoute()
const uiStore = useUiStore()

type ReportTab = 'financial' | 'pharmacy' | 'registrations' | 'demographics' | 'pnl' | 'top_diseases'
const activeTab = ref<ReportTab>('financial')

const validTabs = ['financial', 'pharmacy', 'registrations', 'demographics', 'pnl', 'top_diseases']

// Check query param or route
onMounted(() => {
  if (route.query.tab && validTabs.includes(route.query.tab as string)) {
    activeTab.value = route.query.tab as ReportTab
  }
  loadReportData()
})

watch(() => route.query.tab, (newTab) => {
  if (newTab && validTabs.includes(newTab as string)) {
    activeTab.value = newTab as ReportTab
  }
})

const isLoading = ref(false)

// Global Date Filter Default: Past 30 Days to Today
const dateStart = ref(new Date(Date.now() - 30 * 86400000).toISOString().split('T')[0])
const dateEnd = ref(new Date().toISOString().split('T')[0])

// =========================================================================
// 1. LAPORAN KEUANGAN (FINANCIAL & BILLING REPORT)
// =========================================================================
const invoices = ref<Invoice[]>([])
const finSearch = ref('')
const finStatusFilter = ref('SEMUA')
const finGuarantorFilter = ref('SEMUA')

const financialCols: Column[] = [
  { key: 'nomor_invoice', label: 'No. Invoice', width: '160px' },
  { key: 'created_at', label: 'Tgl Transaksi', width: '130px' },
  { key: 'no_rm', label: 'No. RM', width: '120px' },
  { key: 'patient_name', label: 'Nama Pasien' },
  { key: 'jenis_penjamin', label: 'Penjamin', width: '130px' },
  { key: 'total_tagihan', label: 'Total Tagihan', align: 'right', width: '140px' },
  { key: 'diskon', label: 'Diskon', align: 'right', width: '110px' },
  { key: 'total_terbayar', label: 'Terbayar', align: 'right', width: '140px' },
  { key: 'sisa_tagihan', label: 'Sisa Tagihan', align: 'right', width: '130px' },
  { key: 'status', label: 'Status', align: 'center', width: '120px' },
]

const filteredInvoices = computed(() => {
  return invoices.value.filter(inv => {
    const q = finSearch.value.toLowerCase().trim()
    const matchSearch = !q ||
      (inv.nomor_invoice && inv.nomor_invoice.toLowerCase().includes(q)) ||
      (inv.patient_name && inv.patient_name.toLowerCase().includes(q)) ||
      (inv.no_rm && inv.no_rm.toLowerCase().includes(q)) ||
      (inv.no_rawat && inv.no_rawat.toLowerCase().includes(q))

    const matchStatus = finStatusFilter.value === 'SEMUA' || inv.status === finStatusFilter.value
    const matchGuarantor = finGuarantorFilter.value === 'SEMUA' || inv.jenis_penjamin === finGuarantorFilter.value

    return matchSearch && matchStatus && matchGuarantor
  })
})

const financialMetrics = computed(() => {
  const list = filteredInvoices.value
  const grossRevenue = list.reduce((acc, inv) => acc + (inv.total_tagihan || 0), 0)
  const totalDiscount = list.reduce((acc, inv) => acc + (inv.diskon || 0), 0)
  const netPaid = list.reduce((acc, inv) => acc + (inv.total_terbayar || 0), 0)
  const totalReceivables = list.reduce((acc, inv) => acc + (inv.sisa_tagihan || 0), 0)
  const lunasCount = list.filter(inv => inv.status === 'LUNAS').length
  const unpaidCount = list.filter(inv => inv.status === 'BELUM_BAYAR').length

  return {
    grossRevenue,
    totalDiscount,
    netPaid,
    totalReceivables,
    lunasCount,
    unpaidCount,
    totalCount: list.length
  }
})

function exportFinancialExcel() {
  const exportData = filteredInvoices.value.map((inv, idx) => ({
    'No': idx + 1,
    'No. Invoice': inv.nomor_invoice,
    'No. Rawat': inv.no_rawat || '-',
    'No. RM': inv.no_rm || '-',
    'Nama Pasien': inv.patient_name || '-',
    'Jenis Penjamin': inv.jenis_penjamin || 'Umum Mandiri',
    'Tanggal Transaksi': new Date(inv.created_at).toLocaleDateString('id-ID'),
    'Total Tagihan (Rp)': inv.total_tagihan,
    'Diskon Promo (Rp)': inv.diskon || 0,
    'Total Terbayar (Rp)': inv.total_terbayar || 0,
    'Sisa Tagihan / Piutang (Rp)': inv.sisa_tagihan || 0,
    'Status Pembayaran': inv.status,
    'Catatan': inv.catatan || '-'
  }))
  exportToExcel(`Laporan_Keuangan_SIMRS_${new Date().toISOString().split('T')[0]}`, 'Laporan Keuangan', exportData)
}

// =========================================================================
// 2. LAPORAN STOK FARMASI (PHARMACY STOCK REPORT)
// =========================================================================
const pharmacyItems = ref<InventoryItem[]>([])
const pharmSearch = ref('')
const pharmCategoryFilter = ref('SEMUA')
const pharmStockStatusFilter = ref('SEMUA')

const pharmacyCols: Column[] = [
  { key: 'item_code', label: 'Kode Obat', width: '120px' },
  { key: 'item_name', label: 'Nama Obat & Sediaan' },
  { key: 'category_name', label: 'Kategori', width: '130px' },
  { key: 'unit', label: 'Satuan', width: '90px' },
  { key: 'depo_stock', label: 'Stok Depo', align: 'center', width: '100px' },
  { key: 'total_stock', label: 'Total Stok', align: 'center', width: '110px' },
  { key: 'minimum_stock', label: 'Min. Stok', align: 'center', width: '100px' },
  { key: 'status_stok', label: 'Kondisi Stok', align: 'center', width: '130px' },
  { key: 'buy_price', label: 'Harga Beli', align: 'right', width: '120px' },
  { key: 'sell_price', label: 'Harga Jual', align: 'right', width: '120px' },
  { key: 'valuation', label: 'Valuasi Aset', align: 'right', width: '140px' },
]

const filteredPharmacyItems = computed(() => {
  return pharmacyItems.value.filter(item => {
    const q = pharmSearch.value.toLowerCase().trim()
    const matchSearch = !q ||
      item.item_name.toLowerCase().includes(q) ||
      item.item_code.toLowerCase().includes(q) ||
      item.generic_name.toLowerCase().includes(q)

    const matchCategory = pharmCategoryFilter.value === 'SEMUA' || item.category_name === pharmCategoryFilter.value

    let matchStock = true
    if (pharmStockStatusFilter.value === 'CRITICAL') {
      matchStock = item.total_stock <= item.minimum_stock
    } else if (pharmStockStatusFilter.value === 'SAFE') {
      matchStock = item.total_stock > item.minimum_stock
    } else if (pharmStockStatusFilter.value === 'EMPTY') {
      matchStock = item.total_stock <= 0
    }

    return matchSearch && matchCategory && matchStock
  })
})

const pharmacyMetrics = computed(() => {
  const items = filteredPharmacyItems.value
  const totalValuation = items.reduce((acc, it) => acc + (it.total_stock * it.buy_price), 0)
  const totalDepoValuation = items.reduce((acc, it) => acc + (it.depo_stock * it.buy_price), 0)
  const lowStockCount = items.filter(it => it.total_stock <= it.minimum_stock).length
  const emptyStockCount = items.filter(it => it.total_stock <= 0).length

  return {
    totalItems: items.length,
    totalValuation,
    totalDepoValuation,
    lowStockCount,
    emptyStockCount
  }
})

function exportPharmacyExcel() {
  const exportData = filteredPharmacyItems.value.map((it, idx) => ({
    'No': idx + 1,
    'Kode Obat / BHP': it.item_code,
    'Nama Barang': it.item_name,
    'Nama Generik': it.generic_name,
    'Kategori': it.category_name || 'Obat',
    'Satuan': it.unit,
    'Lokasi Rak': it.rack_location || '-',
    'Stok Depo Rawat Jalan': it.depo_stock,
    'Stok Gudang Total': it.total_stock,
    'Stok Minimum': it.minimum_stock,
    'Status Stok': it.total_stock <= 0 ? 'HABIS' : it.total_stock <= it.minimum_stock ? 'KRITIS / REORDER' : 'AMAN',
    'Harga Beli Satuan (Rp)': it.buy_price,
    'Harga Jual Satuan (Rp)': it.sell_price,
    'Nilai Valuasi Persediaan (Rp)': it.total_stock * it.buy_price
  }))
  exportToExcel(`Laporan_Stok_Farmasi_${new Date().toISOString().split('T')[0]}`, 'Stok Farmasi', exportData)
}

// =========================================================================
// 3. LAPORAN PASIEN DAFTAR (PATIENT VISITS REPORT)
// =========================================================================
const visitsList = ref<Visit[]>([])
const regSearch = ref('')
const regPoliFilter = ref('SEMUA')
const regGuarantorFilter = ref('SEMUA')
const regStatusFilter = ref('SEMUA')

const registrationCols: Column[] = [
  { key: 'no_antrian', label: 'No. Antrean', width: '110px' },
  { key: 'no_rawat', label: 'No. Rawat', width: '160px' },
  { key: 'visit_date', label: 'Tgl Kunjungan', width: '130px' },
  { key: 'no_rm', label: 'No. RM', width: '120px' },
  { key: 'patient_name', label: 'Nama Pasien' },
  { key: 'poli_name', label: 'Poli Unit', width: '150px' },
  { key: 'doctor_name', label: 'Dokter Praktek' },
  { key: 'guarantor_name', label: 'Penjamin', width: '120px' },
  { key: 'status', label: 'Status Layanan', align: 'center', width: '140px' },
]

const filteredVisits = computed(() => {
  return visitsList.value.filter(v => {
    const q = regSearch.value.toLowerCase().trim()
    const matchSearch = !q ||
      (v.patient_name && v.patient_name.toLowerCase().includes(q)) ||
      (v.no_rm && v.no_rm.toLowerCase().includes(q)) ||
      (v.no_rawat && v.no_rawat.toLowerCase().includes(q)) ||
      (v.no_antrian && v.no_antrian.toLowerCase().includes(q))

    const matchPoli = regPoliFilter.value === 'SEMUA' || v.poli_name === regPoliFilter.value
    const matchGuarantor = regGuarantorFilter.value === 'SEMUA' || v.guarantor_name === regGuarantorFilter.value
    const matchStatus = regStatusFilter.value === 'SEMUA' || v.status === regStatusFilter.value

    return matchSearch && matchPoli && matchGuarantor && matchStatus
  })
})

const registrationMetrics = computed(() => {
  const list = filteredVisits.value
  const totalVisits = list.length
  const finishedCount = list.filter(v => v.status === 'SELESAI').length
  const examiningCount = list.filter(v => v.status === 'DIPERIKSA_DOKTER').length
  const waitingCount = list.filter(v => v.status === 'MENUNGGU_TRIASE' || v.status === 'MENUNGGU_DOKTER').length
  const pharmacyCashierCount = list.filter(v => v.status === 'FARMASI' || v.status === 'KASIR').length

  return {
    totalVisits,
    finishedCount,
    examiningCount,
    waitingCount,
    pharmacyCashierCount
  }
})

function exportRegistrationExcel() {
  const exportData = filteredVisits.value.map((v, idx) => ({
    'No': idx + 1,
    'No. Antrean': v.no_antrian || v.nomor_antrean || '-',
    'No. Rawat': v.no_rawat,
    'Tanggal Kunjungan': new Date(v.visit_date || v.created_at).toLocaleDateString('id-ID'),
    'No. RM': v.no_rm || '-',
    'Nama Pasien': v.patient_name || '-',
    'Jenis Kelamin': v.jenis_kelamin || v.patient_gender || '-',
    'Poli Tujuan': v.poli_name || '-',
    'Dokter Pemeriksa': v.doctor_name || '-',
    'Jenis Penjamin': v.guarantor_name || 'Umum',
    'Keluhan Utama': v.keluhan_utama || '-',
    'Status Pelayanan': v.status,
    'Waktu Masuk': v.waktu_masuk ? new Date(v.waktu_masuk).toLocaleTimeString('id-ID') : '-'
  }))
  exportToExcel(`Laporan_Pendaftaran_Pasien_${new Date().toISOString().split('T')[0]}`, 'Kunjungan Pasien', exportData)
}

// =========================================================================
// 4. LAPORAN DEMOGRAFI (PATIENT DEMOGRAPHICS & EPIDEMIOLOGY REPORT)
// =========================================================================
const patients = ref<Patient[]>([])
const demoSearch = ref('')
const demoGenderFilter = ref('SEMUA')
const demoAgeFilter = ref('SEMUA')

const demographicCols: Column[] = [
  { key: 'no_rm', label: 'No. RM', width: '120px' },
  { key: 'nama_pasien', label: 'Nama Lengkap Pasien' },
  { key: 'nik', label: 'NIK / Identitas', width: '160px' },
  { key: 'jenis_kelamin', label: 'Gender', align: 'center', width: '90px' },
  { key: 'tanggal_lahir', label: 'Tgl Lahir', width: '110px' },
  { key: 'age', label: 'Usia', width: '110px' },
  { key: 'golongan_darah', label: 'Gol. Darah', align: 'center', width: '90px' },
  { key: 'alamat', label: 'Alamat Domisili' },
  { key: 'riwayat_alergi', label: 'Alergi', width: '160px' },
]

function calculateAge(dobStr: string) {
  if (!dobStr) return { years: 0, text: '-' }
  const dob = new Date(dobStr)
  const diffMs = Date.now() - dob.getTime()
  const ageDt = new Date(diffMs)
  const years = Math.abs(ageDt.getUTCFullYear() - 1970)
  const months = ageDt.getUTCMonth()
  if (years === 0) return { years, text: `${months} Bln` }
  return { years, text: `${years} Thn` }
}

function getAgeGroup(years: number) {
  if (years < 1) return 'Bayi (< 1 Thn)'
  if (years <= 5) return 'Balita (1 - 5 Thn)'
  if (years <= 12) return 'Anak (6 - 12 Thn)'
  if (years <= 18) return 'Remaja (13 - 18 Thn)'
  if (years <= 59) return 'Dewasa (19 - 59 Thn)'
  return 'Lansia (>= 60 Thn)'
}

const enrichedPatients = computed(() => {
  return patients.value.map(p => {
    const ageInfo = calculateAge(p.tanggal_lahir)
    const ageGroup = getAgeGroup(ageInfo.years)
    return {
      ...p,
      ageYears: ageInfo.years,
      ageText: ageInfo.text,
      ageGroup
    }
  })
})

const filteredPatients = computed(() => {
  return enrichedPatients.value.filter(p => {
    const q = demoSearch.value.toLowerCase().trim()
    const matchSearch = !q ||
      p.nama_pasien.toLowerCase().includes(q) ||
      p.no_rm.toLowerCase().includes(q) ||
      p.nik.toLowerCase().includes(q) ||
      (p.alamat && p.alamat.toLowerCase().includes(q))

    const matchGender = demoGenderFilter.value === 'SEMUA' || p.jenis_kelamin === demoGenderFilter.value
    const matchAge = demoAgeFilter.value === 'SEMUA' || p.ageGroup === demoAgeFilter.value

    return matchSearch && matchGender && matchAge
  })
})

const demographicMetrics = computed(() => {
  const list = filteredPatients.value
  const total = list.length
  const maleCount = list.filter(p => p.jenis_kelamin === 'L').length
  const femaleCount = list.filter(p => p.jenis_kelamin === 'P').length
  const allergyCount = list.filter(p => p.riwayat_alergi && p.riwayat_alergi.trim() !== '' && !p.riwayat_alergi.toLowerCase().includes('tidak ada')).length

  // Age group stats
  const ageMap: Record<string, number> = {
    'Bayi (< 1 Thn)': 0,
    'Balita (1 - 5 Thn)': 0,
    'Anak (6 - 12 Thn)': 0,
    'Remaja (13 - 18 Thn)': 0,
    'Dewasa (19 - 59 Thn)': 0,
    'Lansia (>= 60 Thn)': 0
  }
  list.forEach(p => {
    if (ageMap[p.ageGroup] !== undefined) {
      ageMap[p.ageGroup]++
    }
  })

  return {
    total,
    maleCount,
    femaleCount,
    malePct: total > 0 ? Math.round((maleCount / total) * 100) : 0,
    femalePct: total > 0 ? Math.round((femaleCount / total) * 100) : 0,
    allergyCount,
    ageMap
  }
})

function exportDemographicsExcel() {
  const exportData = filteredPatients.value.map((p, idx) => ({
    'No': idx + 1,
    'No. RM': p.no_rm,
    'NIK': p.nik,
    'Nama Pasien': p.nama_pasien,
    'Jenis Kelamin': p.jenis_kelamin === 'L' ? 'Laki-laki' : 'Perempuan',
    'Tanggal Lahir': p.tanggal_lahir,
    'Usia': p.ageText,
    'Kategori Usia': p.ageGroup,
    'Golongan Darah': p.golongan_darah || '-',
    'No. HP / Kontak': p.no_hp || '-',
    'Alamat Domisili': p.alamat || '-',
    'Riwayat Alergi': p.riwayat_alergi || 'Tidak Ada'
  }))
  exportToExcel(`Laporan_Demografi_Pasien_${new Date().toISOString().split('T')[0]}`, 'Demografi Pasien', exportData)
}

// =========================================================================
// 5. LAPORAN LABA RUGI / P&L (PROFIT & LOSS) (FASE 1)
// =========================================================================
const pnlData = ref<any | null>(null)
const isLoadingPnl = ref(false)

async function loadPnLData() {
  isLoadingPnl.value = true
  try {
    const res = await api.get<any>('/reports/pnl', {
      start_date: dateStart.value,
      end_date: dateEnd.value
    })
    if (res.data) {
      pnlData.value = res.data
    }
  } catch (err: any) {
    console.warn('Gagal memuat laporan P&L', err)
  } finally {
    isLoadingPnl.value = false
  }
}

function exportPnLExcel() {
  if (!pnlData.value) return
  const exportData = [
    { 'Komponen Keuangan': 'Total Pendapatan Kotor (Gross Revenue)', 'Nilai (Rp)': pnlData.value.total_revenue },
    { 'Komponen Keuangan': 'Potongan Diskon & Promo', 'Nilai (Rp)': pnlData.value.total_discounts },
    { 'Komponen Keuangan': 'Pendapatan Bersih (Net Revenue)', 'Nilai (Rp)': pnlData.value.net_revenue },
    { 'Komponen Keuangan': 'Estimasi Beban Pokok Pendapatan (HPP Farmasi 70%)', 'Nilai (Rp)': pnlData.value.estimated_hpp },
    { 'Komponen Keuangan': 'Laba Kotor Operasional (Gross Profit)', 'Nilai (Rp)': pnlData.value.gross_profit },
    { 'Komponen Keuangan': 'Margin Keuntungan Kotor (%)', 'Nilai (Rp)': `${pnlData.value.profit_margin_pct}%` },
    ...((pnlData.value.breakdown || []).map((b: any) => ({
      'Komponen Keuangan': `Rincian: ${b.item_type}`,
      'Nilai (Rp)': b.revenue
    })))
  ]
  exportToExcel(`Laporan_Laba_Rugi_${dateStart.value}_sd_${dateEnd.value}`, 'Laba Rugi', exportData)
}

// =========================================================================
// 6. TOP 10 PENYAKIT ICD-10 (FASE 1)
// =========================================================================
const topDiseases = ref<any[]>([])
const isLoadingDiseases = ref(false)

async function loadTopDiseases() {
  isLoadingDiseases.value = true
  try {
    const res = await api.get<any[]>('/reports/top-diseases', {
      start_date: dateStart.value,
      end_date: dateEnd.value,
      limit: 10
    })
    if (res.data) {
      topDiseases.value = res.data
    }
  } catch (err: any) {
    console.warn('Gagal memuat top 10 penyakit', err)
  } finally {
    isLoadingDiseases.value = false
  }
}

function exportTopDiseasesExcel() {
  const exportData = topDiseases.value.map((d, idx) => ({
    'Peringkat': idx + 1,
    'Kode ICD-10': d.icd10_code,
    'Nama Penyakit / Diagnosis': d.icd10_name,
    'Jumlah Kasus Terdata': d.total_cases,
    'Proporsi Kasus (%)': `${d.percentage}%`
  }))
  exportToExcel(`Laporan_Top_10_Penyakit_ICD10_${dateStart.value}_sd_${dateEnd.value}`, 'Top 10 Morbiditas', exportData)
}

// Master references for filters
const poliList = ref<MasterPoli[]>([])

async function loadReportData() {
  isLoading.value = true
  try {
    // 1. Load Invoices for Financial Report
    const invRes = await api.get<Invoice[]>('/billing/invoices', { limit: 1000 })
    if (invRes.data) invoices.value = invRes.data

    // 2. Load Logistics Items for Pharmacy Report
    const pharmRes = await api.get<InventoryItem[]>('/logistics/items', { limit: 1000 })
    if (pharmRes.data) pharmacyItems.value = pharmRes.data

    // 3. Load Visits for Registration Report
    const visRes = await api.get<Visit[]>('/visits', { limit: 1000 })
    if (visRes.data) visitsList.value = visRes.data

    // 4. Load Patients for Demographics Report
    const patRes = await api.get<Patient[]>('/patients', { limit: 1000 })
    if (patRes.data) patients.value = patRes.data

    // 5. Load Poli Units
    const poliRes = await api.get<MasterPoli[]>('/master/poli')
    if (poliRes.data) poliList.value = poliRes.data

    // 6. Load PnL and Top Diseases
    loadPnLData()
    loadTopDiseases()
  } catch (err: any) {
    console.error('Failed to load report datasets:', err)
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="w-full max-w-full space-y-4 text-left">
    <!-- Header Banner -->
    <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 flex flex-col md:flex-row md:items-center justify-between gap-4 shadow-xs">
      <div>
        <div class="flex items-center gap-2">
          <div class="w-8 h-8 rounded-[6px] bg-[#0284C7] text-white flex items-center justify-center font-bold">
            <FileSpreadsheet class="w-4 h-4" />
          </div>
          <div>
            <h1 class="text-[18px] font-bold text-[#0F172A] leading-tight">
              Pusat Laporan & Rekapitulasi Eksekutif SIMRS
            </h1>
            <p class="text-[12px] text-[#64748B]">
              Monitoring statistik keuangan, stok farmasi, kunjungan registrasi pasien, dan demografi epidemiologi lengkap dengan ekspor Microsoft Excel.
            </p>
          </div>
        </div>
      </div>

      <div class="flex items-center gap-2">
        <BaseButton variant="outline" size="sm" :loading="isLoading" @click="loadReportData">
          <RefreshCw class="w-3.5 h-3.5 mr-1" />
          Segarkan Data
        </BaseButton>
      </div>
    </div>

    <!-- Tab Navigation -->
    <div class="flex items-center gap-1 border-b border-[#E2E8F0] overflow-x-auto bg-white px-2 pt-2 rounded-t-[8px]">
      <button
        type="button"
        :class="[
          'px-4 py-2.5 text-[13px] font-semibold border-b-2 transition-all flex items-center gap-2 whitespace-nowrap',
          activeTab === 'financial'
            ? 'border-[#0284C7] text-[#0284C7] bg-[#F0F9FF]'
            : 'border-transparent text-[#64748B] hover:text-[#0F172A]'
        ]"
        @click="activeTab = 'financial'"
      >
        <TrendingUp class="w-4 h-4 text-[#0284C7]" />
        Laporan Keuangan
      </button>

      <button
        type="button"
        :class="[
          'px-4 py-2.5 text-[13px] font-semibold border-b-2 transition-all flex items-center gap-2 whitespace-nowrap',
          activeTab === 'pharmacy'
            ? 'border-[#0284C7] text-[#0284C7] bg-[#F0F9FF]'
            : 'border-transparent text-[#64748B] hover:text-[#0F172A]'
        ]"
        @click="activeTab = 'pharmacy'"
      >
        <Boxes class="w-4 h-4 text-[#0284C7]" />
        Laporan Stok Farmasi
      </button>

      <button
        type="button"
        :class="[
          'px-4 py-2.5 text-[13px] font-semibold border-b-2 transition-all flex items-center gap-2 whitespace-nowrap',
          activeTab === 'registrations'
            ? 'border-[#0284C7] text-[#0284C7] bg-[#F0F9FF]'
            : 'border-transparent text-[#64748B] hover:text-[#0F172A]'
        ]"
        @click="activeTab = 'registrations'"
      >
        <Users class="w-4 h-4 text-[#0284C7]" />
        Laporan Pasien Daftar
      </button>

      <button
        type="button"
        :class="[
          'px-4 py-2.5 text-[13px] font-semibold border-b-2 transition-all flex items-center gap-2 whitespace-nowrap',
          activeTab === 'demographics'
            ? 'border-[#0284C7] text-[#0284C7] bg-[#F0F9FF]'
            : 'border-transparent text-[#64748B] hover:text-[#0F172A]'
        ]"
        @click="activeTab = 'demographics'"
      >
        <PieChart class="w-4 h-4 text-[#0284C7]" />
        Laporan Demografi
      </button>

      <button
        type="button"
        :class="[
          'px-4 py-2.5 text-[13px] font-semibold border-b-2 transition-all flex items-center gap-2 whitespace-nowrap',
          activeTab === 'pnl'
            ? 'border-[#0284C7] text-[#0284C7] bg-[#F0F9FF]'
            : 'border-transparent text-[#64748B] hover:text-[#0F172A]'
        ]"
        @click="activeTab = 'pnl'; loadPnLData()"
      >
        <Calculator class="w-4 h-4 text-[#0284C7]" />
        Laba Rugi (P&L)
      </button>

      <button
        type="button"
        :class="[
          'px-4 py-2.5 text-[13px] font-semibold border-b-2 transition-all flex items-center gap-2 whitespace-nowrap',
          activeTab === 'top_diseases'
            ? 'border-[#0284C7] text-[#0284C7] bg-[#F0F9FF]'
            : 'border-transparent text-[#64748B] hover:text-[#0F172A]'
        ]"
        @click="activeTab = 'top_diseases'; loadTopDiseases()"
      >
        <Activity class="w-4 h-4 text-[#0284C7]" />
        Top 10 Penyakit (ICD-10)
      </button>
    </div>

    <!-- ========================================================================= -->
    <!-- TAB 1: LAPORAN KEUANGAN                                                   -->
    <!-- ========================================================================= -->
    <div v-if="activeTab === 'financial'" class="space-y-4">
      <!-- KPI Executive Cards -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-3.5 shadow-xs">
          <span class="text-[11px] font-semibold text-[#64748B] uppercase tracking-wider block">Total Pendapatan Bruto</span>
          <div class="text-[18px] font-bold text-[#0F172A] font-mono mt-1">
            Rp {{ financialMetrics.grossRevenue.toLocaleString('id-ID') }}
          </div>
          <span class="text-[11px] text-[#64748B] mt-0.5 block">Dari {{ financialMetrics.totalCount }} tagihan diterbitkan</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-3.5 shadow-xs">
          <span class="text-[11px] font-semibold text-emerald-700 uppercase tracking-wider block">Total Kas Diterima (Lunas)</span>
          <div class="text-[18px] font-bold text-emerald-700 font-mono mt-1">
            Rp {{ financialMetrics.netPaid.toLocaleString('id-ID') }}
          </div>
          <span class="text-[11px] text-emerald-600 mt-0.5 block">{{ financialMetrics.lunasCount }} tagihan telah lunas</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-3.5 shadow-xs">
          <span class="text-[11px] font-semibold text-amber-700 uppercase tracking-wider block">Sisa Tagihan / Piutang</span>
          <div class="text-[18px] font-bold text-amber-700 font-mono mt-1">
            Rp {{ financialMetrics.totalReceivables.toLocaleString('id-ID') }}
          </div>
          <span class="text-[11px] text-amber-600 mt-0.5 block">{{ financialMetrics.unpaidCount }} invoice belum bayar</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-3.5 shadow-xs">
          <span class="text-[11px] font-semibold text-[#0284C7] uppercase tracking-wider block">Total Diskon Promo</span>
          <div class="text-[18px] font-bold text-[#0284C7] font-mono mt-1">
            Rp {{ financialMetrics.totalDiscount.toLocaleString('id-ID') }}
          </div>
          <span class="text-[11px] text-[#0284C7] mt-0.5 block">Potongan klaim promo klinik</span>
        </div>
      </div>

      <!-- Filters & Action Bar -->
      <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-3.5 shadow-xs space-y-3">
        <div class="flex flex-col md:flex-row md:items-center justify-between gap-3">
          <div class="flex flex-wrap items-center gap-3">
            <div class="w-64">
              <BaseInput
                v-model="finSearch"
                placeholder="Cari No. Invoice, Pasien, RM..."
                compact
              />
            </div>

            <div class="w-48">
              <BaseSelect
                v-model="finStatusFilter"
                :options="[
                  { value: 'SEMUA', label: 'Semua Status' },
                  { value: 'LUNAS', label: 'Lunas' },
                  { value: 'BELUM_BAYAR', label: 'Belum Bayar' },
                  { value: 'BATAL', label: 'Batal' }
                ]"
              />
            </div>

            <div class="w-48">
              <BaseSelect
                v-model="finGuarantorFilter"
                :options="[
                  { value: 'SEMUA', label: 'Semua Penjamin' },
                  { value: 'Umum Mandiri', label: 'Umum Mandiri' },
                  { value: 'BPJS Kesehatan', label: 'BPJS Kesehatan' },
                  { value: 'Asuransi Mandiri Inhealth', label: 'Mandiri Inhealth' },
                  { value: 'Asuransi Prudential', label: 'Prudential' }
                ]"
              />
            </div>
          </div>

          <BaseButton variant="primary" size="md" @click="exportFinancialExcel">
            <Download class="w-4 h-4 mr-1.5" />
            Download Excel (.xlsx)
          </BaseButton>
        </div>
      </div>

      <!-- Financial Table -->
      <DataTable
        :columns="financialCols"
        :data="filteredInvoices"
        :loading="isLoading"
      >
        <template #nomor_invoice="{ row }">
          <span class="font-mono font-bold text-[#0284C7]">{{ row.nomor_invoice }}</span>
        </template>
        <template #created_at="{ row }">
          <span class="text-[12px] text-[#475569]">
            {{ new Date(row.created_at).toLocaleDateString('id-ID') }}
          </span>
        </template>
        <template #total_tagihan="{ row }">
          <span class="font-mono font-semibold tabular-nums text-[#0F172A]">
            Rp {{ Number(row.total_tagihan).toLocaleString('id-ID') }}
          </span>
        </template>
        <template #diskon="{ row }">
          <span class="font-mono tabular-nums text-slate-500">
            {{ row.diskon ? `Rp ${Number(row.diskon).toLocaleString('id-ID')}` : '-' }}
          </span>
        </template>
        <template #total_terbayar="{ row }">
          <span class="font-mono font-bold text-emerald-700 tabular-nums">
            Rp {{ Number(row.total_terbayar || 0).toLocaleString('id-ID') }}
          </span>
        </template>
        <template #sisa_tagihan="{ row }">
          <span class="font-mono tabular-nums" :class="row.sisa_tagihan > 0 ? 'text-amber-700 font-bold' : 'text-slate-400'">
            Rp {{ Number(row.sisa_tagihan || 0).toLocaleString('id-ID') }}
          </span>
        </template>
        <template #status="{ row }">
          <StatusBadge :status="row.status === 'LUNAS' ? 'Lunas' : row.status === 'BELUM_BAYAR' ? 'Belum Bayar' : 'Batal'" :variant="row.status === 'LUNAS' ? 'success' : row.status === 'BELUM_BAYAR' ? 'warning' : 'danger'" />
        </template>
      </DataTable>
    </div>

    <!-- ========================================================================= -->
    <!-- TAB 2: LAPORAN STOK FARMASI                                               -->
    <!-- ========================================================================= -->
    <div v-if="activeTab === 'pharmacy'" class="space-y-4">
      <!-- KPI Cards -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-3.5 shadow-xs">
          <span class="text-[11px] font-semibold text-[#64748B] uppercase tracking-wider block">Total Jenis Barang / Obat</span>
          <div class="text-[20px] font-bold text-[#0F172A] mt-1">{{ pharmacyMetrics.totalItems }} Sediaan</div>
          <span class="text-[11px] text-[#64748B] mt-0.5 block">Formularium klinik terdaftar</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-3.5 shadow-xs">
          <span class="text-[11px] font-semibold text-[#0284C7] uppercase tracking-wider block">Valuasi Persediaan Total</span>
          <div class="text-[18px] font-bold text-[#0284C7] font-mono mt-1">
            Rp {{ pharmacyMetrics.totalValuation.toLocaleString('id-ID') }}
          </div>
          <span class="text-[11px] text-[#0284C7] mt-0.5 block">Aset stok gudang & depo</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-3.5 shadow-xs">
          <span class="text-[11px] font-semibold text-amber-700 uppercase tracking-wider block">Stok Kritis (&lt; Min. Stok)</span>
          <div class="text-[20px] font-bold text-amber-700 mt-1">{{ pharmacyMetrics.lowStockCount }} Item</div>
          <span class="text-[11px] text-amber-600 mt-0.5 block">Segera buat Purchase Order (PO)</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-3.5 shadow-xs">
          <span class="text-[11px] font-semibold text-rose-700 uppercase tracking-wider block">Stok Kosong / Habis</span>
          <div class="text-[20px] font-bold text-rose-700 mt-1">{{ pharmacyMetrics.emptyStockCount }} Item</div>
          <span class="text-[11px] text-rose-600 mt-0.5 block">Ketersediaan 0 di klinik</span>
        </div>
      </div>

      <!-- Filters & Action Bar -->
      <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-3.5 shadow-xs space-y-3">
        <div class="flex flex-col md:flex-row md:items-center justify-between gap-3">
          <div class="flex flex-wrap items-center gap-3">
            <div class="w-64">
              <BaseInput
                v-model="pharmSearch"
                placeholder="Cari kode obat, nama, generik..."
                compact
              />
            </div>

            <div class="w-48">
              <BaseSelect
                v-model="pharmStockStatusFilter"
                :options="[
                  { value: 'SEMUA', label: 'Semua Status Stok' },
                  { value: 'CRITICAL', label: 'Stok Kritis / Reorder' },
                  { value: 'SAFE', label: 'Stok Aman' },
                  { value: 'EMPTY', label: 'Stok Kosong (0)' }
                ]"
              />
            </div>
          </div>

          <BaseButton variant="primary" size="md" @click="exportPharmacyExcel">
            <Download class="w-4 h-4 mr-1.5" />
            Download Excel (.xlsx)
          </BaseButton>
        </div>
      </div>

      <!-- Pharmacy Table -->
      <DataTable
        :columns="pharmacyCols"
        :data="filteredPharmacyItems"
        :loading="isLoading"
      >
        <template #item_code="{ row }">
          <span class="font-mono font-bold text-[#0284C7] bg-[#F0F9FF] px-2 py-0.5 rounded border border-[#BAE6FD] text-[11px]">
            {{ row.item_code }}
          </span>
        </template>
        <template #item_name="{ row }">
          <div class="font-semibold text-slate-800">{{ row.item_name }}</div>
          <div class="text-[11px] text-slate-500 font-mono">{{ row.generic_name }}</div>
        </template>
        <template #depo_stock="{ row }">
          <span class="font-bold text-[#0F172A]">{{ row.depo_stock }}</span>
        </template>
        <template #total_stock="{ row }">
          <span class="font-bold font-mono" :class="row.total_stock <= row.minimum_stock ? 'text-amber-700' : 'text-slate-800'">
            {{ row.total_stock }}
          </span>
        </template>
        <template #status_stok="{ row }">
          <span
            v-if="row.total_stock <= 0"
            class="px-2 py-0.5 text-[10px] font-bold rounded bg-rose-100 text-rose-800 border border-rose-200"
          >
            HABIS
          </span>
          <span
            v-else-if="row.total_stock <= row.minimum_stock"
            class="px-2 py-0.5 text-[10px] font-bold rounded bg-amber-100 text-amber-800 border border-amber-200"
          >
            PERLU ORDER
          </span>
          <span
            v-else
            class="px-2 py-0.5 text-[10px] font-bold rounded bg-emerald-100 text-emerald-800 border border-emerald-200"
          >
            AMAN
          </span>
        </template>
        <template #buy_price="{ row }">
          <span class="font-mono text-slate-600 text-[12px]">Rp {{ Number(row.buy_price).toLocaleString('id-ID') }}</span>
        </template>
        <template #sell_price="{ row }">
          <span class="font-mono text-slate-800 text-[12px] font-semibold">Rp {{ Number(row.sell_price).toLocaleString('id-ID') }}</span>
        </template>
        <template #valuation="{ row }">
          <span class="font-mono font-bold text-[#0284C7] tabular-nums">
            Rp {{ (row.total_stock * row.buy_price).toLocaleString('id-ID') }}
          </span>
        </template>
      </DataTable>
    </div>

    <!-- ========================================================================= -->
    <!-- TAB 3: LAPORAN PASIEN DAFTAR (VISITS)                                     -->
    <!-- ========================================================================= -->
    <div v-if="activeTab === 'registrations'" class="space-y-4">
      <!-- KPI Cards -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-3.5 shadow-xs">
          <span class="text-[11px] font-semibold text-[#64748B] uppercase tracking-wider block">Total Kunjungan Pasien</span>
          <div class="text-[20px] font-bold text-[#0F172A] mt-1">{{ registrationMetrics.totalVisits }} Pasien</div>
          <span class="text-[11px] text-[#64748B] mt-0.5 block">Tercatat di loket pendaftaran</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-3.5 shadow-xs">
          <span class="text-[11px] font-semibold text-emerald-700 uppercase tracking-wider block">Selesai Diperiksa</span>
          <div class="text-[20px] font-bold text-emerald-700 mt-1">{{ registrationMetrics.finishedCount }} Pasien</div>
          <span class="text-[11px] text-emerald-600 mt-0.5 block">Pelayanan klinis komplit</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-3.5 shadow-xs">
          <span class="text-[11px] font-semibold text-[#0284C7] uppercase tracking-wider block">Sedang Diperiksa Dokter</span>
          <div class="text-[20px] font-bold text-[#0284C7] mt-1">{{ registrationMetrics.examiningCount }} Pasien</div>
          <span class="text-[11px] text-[#0284C7] mt-0.5 block">Di dalam ruang poli</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-3.5 shadow-xs">
          <span class="text-[11px] font-semibold text-amber-700 uppercase tracking-wider block">Menunggu Antrean</span>
          <div class="text-[20px] font-bold text-amber-700 mt-1">{{ registrationMetrics.waitingCount }} Pasien</div>
          <span class="text-[11px] text-amber-600 mt-0.5 block">Ruang tunggu poli & triase</span>
        </div>
      </div>

      <!-- Filters & Action Bar -->
      <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-3.5 shadow-xs space-y-3">
        <div class="flex flex-col md:flex-row md:items-center justify-between gap-3">
          <div class="flex flex-wrap items-center gap-3">
            <div class="w-64">
              <BaseInput
                v-model="regSearch"
                placeholder="Cari No. Rawat, Pasien, RM..."
                compact
              />
            </div>

            <div class="w-48">
              <BaseSelect
                v-model="regPoliFilter"
                :options="[
                  { value: 'SEMUA', label: 'Semua Poli' },
                  ...poliList.map(p => ({ value: p.name, label: p.name }))
                ]"
              />
            </div>

            <div class="w-48">
              <BaseSelect
                v-model="regStatusFilter"
                :options="[
                  { value: 'SEMUA', label: 'Semua Status' },
                  { value: 'MENUNGGU_TRIASE', label: 'Menunggu Triase' },
                  { value: 'MENUNGGU_DOKTER', label: 'Menunggu Dokter' },
                  { value: 'DIPERIKSA_DOKTER', label: 'Sedang Diperiksa' },
                  { value: 'FARMASI', label: 'Di Farmasi' },
                  { value: 'KASIR', label: 'Di Kasir' },
                  { value: 'SELESAI', label: 'Selesai' }
                ]"
              />
            </div>
          </div>

          <BaseButton variant="primary" size="md" @click="exportRegistrationExcel">
            <Download class="w-4 h-4 mr-1.5" />
            Download Excel (.xlsx)
          </BaseButton>
        </div>
      </div>

      <!-- Registration Table -->
      <DataTable
        :columns="registrationCols"
        :data="filteredVisits"
        :loading="isLoading"
      >
        <template #no_antrian="{ row }">
          <span class="font-mono font-bold text-[#0284C7] bg-[#F0F9FF] px-2 py-0.5 rounded border border-[#BAE6FD]">
            {{ row.no_antrian || row.nomor_antrean || '-' }}
          </span>
        </template>
        <template #no_rawat="{ row }">
          <span class="font-mono text-[12px] font-semibold text-slate-700">{{ row.no_rawat }}</span>
        </template>
        <template #visit_date="{ row }">
          <span class="text-[12px] text-slate-600">
            {{ new Date(row.visit_date || row.created_at).toLocaleDateString('id-ID') }}
          </span>
        </template>
        <template #status="{ row }">
          <StatusBadge :status="row.status" />
        </template>
      </DataTable>
    </div>

    <!-- ========================================================================= -->
    <!-- TAB 4: LAPORAN DEMOGRAFI PASIEN                                           -->
    <!-- ========================================================================= -->
    <div v-if="activeTab === 'demographics'" class="space-y-4">
      <!-- Demographic Visual Distributions -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
        <!-- Gender Breakdown -->
        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs space-y-3">
          <h3 class="text-[13px] font-bold text-[#0F172A] flex items-center justify-between">
            <span>Rasio Jenis Kelamin</span>
            <span class="text-xs font-semibold text-slate-500">{{ demographicMetrics.total }} Pasien</span>
          </h3>

          <div class="space-y-2 pt-1">
            <div>
              <div class="flex justify-between text-xs font-semibold mb-1">
                <span class="text-blue-700">Laki-laki (L): {{ demographicMetrics.maleCount }}</span>
                <span class="text-blue-700 font-bold">{{ demographicMetrics.malePct }}%</span>
              </div>
              <div class="w-full h-2.5 bg-slate-100 rounded-full overflow-hidden">
                <div class="h-full bg-blue-600 rounded-full" :style="{ width: `${demographicMetrics.malePct}%` }"></div>
              </div>
            </div>

            <div>
              <div class="flex justify-between text-xs font-semibold mb-1">
                <span class="text-pink-700">Perempuan (P): {{ demographicMetrics.femaleCount }}</span>
                <span class="text-pink-700 font-bold">{{ demographicMetrics.femalePct }}%</span>
              </div>
              <div class="w-full h-2.5 bg-slate-100 rounded-full overflow-hidden">
                <div class="h-full bg-pink-500 rounded-full" :style="{ width: `${demographicMetrics.femalePct}%` }"></div>
              </div>
            </div>
          </div>
        </div>

        <!-- Age Distribution -->
        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs space-y-2 md:col-span-2">
          <h3 class="text-[13px] font-bold text-[#0F172A]">Sebaran Kelompok Usia Pasien</h3>
          <div class="grid grid-cols-3 md:grid-cols-6 gap-2 pt-1 text-center">
            <div
              v-for="(count, group) in demographicMetrics.ageMap"
              :key="group"
              class="p-2.5 rounded-[6px] bg-[#F8FAFC] border border-[#E2E8F0]"
            >
              <div class="text-[10px] font-semibold text-[#64748B] truncate">{{ group }}</div>
              <div class="text-[16px] font-bold text-[#0284C7] mt-1">{{ count }}</div>
              <div class="text-[10px] text-slate-400">
                {{ demographicMetrics.total > 0 ? Math.round((count / demographicMetrics.total) * 100) : 0 }}%
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Filters & Action Bar -->
      <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-3.5 shadow-xs space-y-3">
        <div class="flex flex-col md:flex-row md:items-center justify-between gap-3">
          <div class="flex flex-wrap items-center gap-3">
            <div class="w-64">
              <BaseInput
                v-model="demoSearch"
                placeholder="Cari Nama Pasien, NIK, No. RM..."
                compact
              />
            </div>

            <div class="w-48">
              <BaseSelect
                v-model="demoGenderFilter"
                :options="[
                  { value: 'SEMUA', label: 'Semua Gender' },
                  { value: 'L', label: 'Laki-laki (L)' },
                  { value: 'P', label: 'Perempuan (P)' }
                ]"
              />
            </div>

            <div class="w-48">
              <BaseSelect
                v-model="demoAgeFilter"
                :options="[
                  { value: 'SEMUA', label: 'Semua Kelompok Usia' },
                  { value: 'Bayi (< 1 Thn)', label: 'Bayi (< 1 Thn)' },
                  { value: 'Balita (1 - 5 Thn)', label: 'Balita (1 - 5 Thn)' },
                  { value: 'Anak (6 - 12 Thn)', label: 'Anak (6 - 12 Thn)' },
                  { value: 'Remaja (13 - 18 Thn)', label: 'Remaja (13 - 18 Thn)' },
                  { value: 'Dewasa (19 - 59 Thn)', label: 'Dewasa (19 - 59 Thn)' },
                  { value: 'Lansia (>= 60 Thn)', label: 'Lansia (>= 60 Thn)' }
                ]"
              />
            </div>
          </div>

          <BaseButton variant="primary" size="md" @click="exportDemographicsExcel">
            <Download class="w-4 h-4 mr-1.5" />
            Download Excel (.xlsx)
          </BaseButton>
        </div>
      </div>

      <!-- Demographics Table -->
      <DataTable
        :columns="demographicCols"
        :data="filteredPatients"
        :loading="isLoading"
      >
        <template #no_rm="{ row }">
          <span class="font-mono font-bold text-[#0284C7]">{{ row.no_rm }}</span>
        </template>
        <template #jenis_kelamin="{ row }">
          <span
            class="px-2 py-0.5 rounded text-[11px] font-bold"
            :class="row.jenis_kelamin === 'L' ? 'bg-blue-50 text-blue-700 border border-blue-200' : 'bg-pink-50 text-pink-700 border border-pink-200'"
          >
            {{ row.jenis_kelamin === 'L' ? 'Laki-laki' : 'Perempuan' }}
          </span>
        </template>
        <template #age="{ row }">
          <span class="font-semibold text-slate-800 text-[12px]">{{ row.ageText }}</span>
          <div class="text-[10px] text-slate-400">{{ row.ageGroup }}</div>
        </template>
        <template #golongan_darah="{ row }">
          <span class="font-bold text-slate-700 bg-slate-100 px-2 py-0.5 rounded text-xs">{{ row.golongan_darah || '-' }}</span>
        </template>
        <template #riwayat_alergi="{ row }">
          <span v-if="row.riwayat_alergi && !row.riwayat_alergi.toLowerCase().includes('tidak')" class="text-rose-600 font-semibold text-xs flex items-center gap-1">
            <AlertTriangle class="w-3.5 h-3.5" />
            {{ row.riwayat_alergi }}
          </span>
          <span v-else class="text-slate-400 text-xs">-</span>
        </template>
      </DataTable>
    </div>

    <!-- ========================================================================= -->
    <!-- TAB 5: LAPORAN LABA RUGI / P&L (FASE 1)                                   -->
    <!-- ========================================================================= -->
    <div v-if="activeTab === 'pnl'" class="space-y-4">
      <!-- Filter Bar -->
      <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-3.5 shadow-xs">
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div class="flex flex-wrap items-center gap-2">
            <span class="text-[12px] font-bold text-[#475569]">Rentang Periode P&L:</span>
            <div class="w-36">
              <BaseInput v-model="dateStart" type="date" compact />
            </div>
            <span class="text-[12px] text-[#94A3B8]">s/d</span>
            <div class="w-36">
              <BaseInput v-model="dateEnd" type="date" compact />
            </div>
            <BaseButton variant="secondary" size="sm" :loading="isLoadingPnl" @click="loadPnLData">
              <Filter class="w-3.5 h-3.5 mr-1" />
              Terapkan
            </BaseButton>
          </div>

          <BaseButton variant="primary" size="md" @click="exportPnLExcel">
            <Download class="w-4 h-4 mr-1.5" />
            Download Excel (.xlsx)
          </BaseButton>
        </div>
      </div>

      <!-- KPI Executive Cards P&L -->
      <div v-if="pnlData" class="grid grid-cols-2 lg:grid-cols-5 gap-3">
        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-3.5 shadow-xs">
          <span class="text-[11px] font-semibold text-[#64748B] block">Total Pendapatan Kotor</span>
          <div class="text-[18px] font-extrabold font-mono text-[#0F172A] mt-1">
            Rp {{ Number(pnlData.total_revenue || 0).toLocaleString('id-ID') }}
          </div>
          <span class="text-[10px] text-slate-400">Sebelum diskon promo</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-3.5 shadow-xs">
          <span class="text-[11px] font-semibold text-[#64748B] block">Potongan Diskon & Promo</span>
          <div class="text-[18px] font-extrabold font-mono text-rose-600 mt-1">
            -Rp {{ Number(pnlData.total_discounts || 0).toLocaleString('id-ID') }}
          </div>
          <span class="text-[10px] text-slate-400">Total subsidi kupon</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-3.5 shadow-xs">
          <span class="text-[11px] font-semibold text-[#64748B] block">Pendapatan Bersih (Net)</span>
          <div class="text-[18px] font-extrabold font-mono text-[#0284C7] mt-1">
            Rp {{ Number(pnlData.net_revenue || 0).toLocaleString('id-ID') }}
          </div>
          <span class="text-[10px] text-sky-700 font-medium">Omset riil masuk kasir</span>
        </div>

        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-3.5 shadow-xs">
          <span class="text-[11px] font-semibold text-[#64748B] block">Estimasi Beban HPP Obat</span>
          <div class="text-[18px] font-extrabold font-mono text-amber-700 mt-1">
            Rp {{ Number(pnlData.estimated_hpp || 0).toLocaleString('id-ID') }}
          </div>
          <span class="text-[10px] text-amber-700 font-medium">~70% dari omset farmasi</span>
        </div>

        <div class="bg-emerald-50 border border-emerald-200 rounded-[8px] p-3.5 shadow-xs col-span-2 lg:col-span-1">
          <div class="flex items-center justify-between">
            <span class="text-[11px] font-bold text-emerald-900">Laba Kotor Operasional</span>
            <span class="px-1.5 py-0.5 rounded bg-emerald-200 text-emerald-800 text-[10px] font-bold font-mono">
              {{ pnlData.profit_margin_pct }}%
            </span>
          </div>
          <div class="text-[20px] font-extrabold font-mono text-emerald-800 mt-1">
            Rp {{ Number(pnlData.gross_profit || 0).toLocaleString('id-ID') }}
          </div>
          <span class="text-[10px] text-emerald-700 font-semibold">Gross Profit Margin</span>
        </div>
      </div>

      <!-- Breakdown Komponen Pendapatan -->
      <div v-if="pnlData" class="bg-white border border-[#E2E8F0] rounded-[8px] shadow-xs overflow-hidden">
        <div class="p-3.5 border-b border-[#E2E8F0] bg-[#F8FAFC]">
          <h3 class="text-[13px] font-bold text-[#0F172A]">Rincian Pendapatan Berdasarkan Komponen Pelayanan</h3>
        </div>
        <table class="w-full text-left text-[12px]">
          <thead class="bg-[#F8FAFC] text-[#475569] border-b font-bold">
            <tr>
              <th class="p-3">Komponen Pelayanan</th>
              <th class="p-3 text-right">Nilai Pendapatan (Rp)</th>
              <th class="p-3 text-right w-44">Kontribusi (%)</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[#F1F5F9]">
            <tr v-for="b in (pnlData.breakdown || [])" :key="b.item_type" class="hover:bg-[#F8FAFC]">
              <td class="p-3 font-semibold text-[#0F172A] flex items-center gap-2">
                <span class="w-2 h-2 rounded-full bg-[#0284C7]"></span>
                {{ b.item_type }}
              </td>
              <td class="p-3 text-right font-mono font-bold text-slate-800">
                Rp {{ Number(b.revenue).toLocaleString('id-ID') }}
              </td>
              <td class="p-3 text-right font-mono text-[#0284C7] font-semibold">
                {{ pnlData.total_revenue > 0 ? ((b.revenue / pnlData.total_revenue) * 100).toFixed(1) : 0 }}%
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- ========================================================================= -->
    <!-- TAB 6: TOP 10 PENYAKIT ICD-10 (FASE 1)                                    -->
    <!-- ========================================================================= -->
    <div v-if="activeTab === 'top_diseases'" class="space-y-4">
      <div class="p-3 rounded-[6px] bg-[#ECFDF5] border border-[#A7F3D0] text-[#065F46] text-[12px] flex items-center gap-2">
        <CheckCircle2 class="w-4 h-4 text-[#059669] shrink-0" />
        <span>
          <strong>Laporan Morbiditas Resmi (ICD-10):</strong> Menampilkan 10 diagnosis penyakit tertinggi pada pelayanan poliklinik rawat jalan, siap diekspor ke Microsoft Excel untuk pelaporan Dinkes / Kemenkes RI.
        </span>
      </div>

      <!-- Filter Bar -->
      <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-3.5 shadow-xs">
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div class="flex flex-wrap items-center gap-2">
            <span class="text-[12px] font-bold text-[#475569]">Rentang Tanggal Diagnosis:</span>
            <div class="w-36">
              <BaseInput v-model="dateStart" type="date" compact />
            </div>
            <span class="text-[12px] text-[#94A3B8]">s/d</span>
            <div class="w-36">
              <BaseInput v-model="dateEnd" type="date" compact />
            </div>
            <BaseButton variant="secondary" size="sm" :loading="isLoadingDiseases" @click="loadTopDiseases">
              <Filter class="w-3.5 h-3.5 mr-1" />
              Terapkan
            </BaseButton>
          </div>

          <BaseButton variant="primary" size="md" @click="exportTopDiseasesExcel">
            <Download class="w-4 h-4 mr-1.5" />
            Download Excel (.xlsx)
          </BaseButton>
        </div>
      </div>

      <!-- Top 10 Table -->
      <div class="bg-white border border-[#E2E8F0] rounded-[8px] shadow-xs overflow-hidden">
        <table class="w-full text-left text-[12px]">
          <thead class="bg-[#F8FAFC] text-[#475569] border-b font-bold">
            <tr>
              <th class="p-3 text-center w-16">Peringkat</th>
              <th class="p-3 w-32">Kode ICD-10</th>
              <th class="p-3">Nama Diagnosis / Penyakit</th>
              <th class="p-3 text-center w-36">Jumlah Kasus</th>
              <th class="p-3 w-64">Proporsi Kasus (%)</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[#F1F5F9]">
            <tr v-if="topDiseases.length === 0">
              <td colspan="5" class="p-8 text-center text-[#94A3B8]">
                Belum ada data diagnosis penyakit terdata pada rentang tanggal ini.
              </td>
            </tr>
            <tr v-for="(dis, idx) in topDiseases" :key="dis.icd10_code" class="hover:bg-[#F8FAFC]">
              <td class="p-3 text-center">
                <span
                  class="w-6 h-6 rounded-full inline-flex items-center justify-center text-[11px] font-extrabold text-white"
                  :class="idx === 0 ? 'bg-amber-500' : idx === 1 ? 'bg-slate-400' : idx === 2 ? 'bg-amber-700' : 'bg-slate-200 !text-slate-700'"
                >
                  {{ idx + 1 }}
                </span>
              </td>
              <td class="p-3 font-mono font-bold text-[#0284C7]">{{ dis.icd10_code }}</td>
              <td class="p-3 font-semibold text-[#0F172A]">{{ dis.icd10_name }}</td>
              <td class="p-3 text-center font-mono font-bold text-slate-800 text-[13px]">
                {{ dis.total_cases }} pasien
              </td>
              <td class="p-3">
                <div class="flex items-center gap-2">
                  <div class="flex-1 bg-slate-100 rounded-full h-2.5 overflow-hidden">
                    <div
                      class="bg-[#0284C7] h-2.5 rounded-full transition-all"
                      :style="{ width: `${dis.percentage}%` }"
                    ></div>
                  </div>
                  <span class="font-mono text-[11px] font-bold text-slate-600 w-12 text-right">
                    {{ dis.percentage }}%
                  </span>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
