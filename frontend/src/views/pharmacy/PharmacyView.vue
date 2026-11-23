<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '@/api/client'
import { useUiStore } from '@/stores/ui'
import { usePatientContextStore } from '@/stores/patientContext'
import DataTable, { type Column } from '@/components/common/DataTable.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseModal from '@/components/common/BaseModal.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import {
  Pill, FileText, CheckCircle2, AlertTriangle, BarChart3,
  Clock, Printer, RefreshCw, Search, ArrowRightLeft, ShieldCheck,
  TrendingUp, Boxes, Package, RotateCcw, Plus, Trash2
} from 'lucide-vue-next'
import type { Prescription, PrescriptionItem, InventoryItem, StockMutation } from '@/types'

const uiStore = useUiStore()
const patientContext = usePatientContextStore()

type PharmacyTab = 'resep' | 'kartu_stok' | 'grafik_order' | 'monitoring' | 'retur'
const activeTab = ref<PharmacyTab>('resep')

// ----------------------------------------------------
// TAB 1: PRESCRIPTIONS QUEUE & VERIFIKASI (AUTO POTONG STOK)
// ----------------------------------------------------
const prescriptions = ref<Prescription[]>([])
const isLoadingPrescriptions = ref(false)
const statusFilter = ref('')
const search = ref('')

const isReviewModalOpen = ref(false)
const isEtiketModalOpen = ref(false)
const isSubmitting = ref(false)
const selectedPrescription = ref<Prescription | null>(null)
const selectedItemForEtiket = ref<PrescriptionItem | null>(null)
const catatanTelaah = ref('Resep memenuhi kelayakan klinis dan administratif farmasi')

function printEtiket() {
  window.print()
}

const prescriptionCols: Column[] = [
  { key: 'nomor_resep', label: 'No. Resep', width: '150px' },
  { key: 'no_rm', label: 'No. RM', width: '120px' },
  { key: 'patient_name', label: 'Nama Pasien' },
  { key: 'doctor_name', label: 'Dokter Penulis' },
  { key: 'items_count', label: 'Jumlah Item', align: 'center', width: '110px' },
  { key: 'status', label: 'Status Resep', align: 'center', width: '150px' },
  { key: 'actions', label: 'Aksi Farmasi', align: 'right', width: '250px' },
]

async function loadPrescriptions() {
  isLoadingPrescriptions.value = true
  try {
    const res = await api.get<Prescription[]>('/pharmacy/prescriptions', {
      status: statusFilter.value,
      search: search.value,
    })
    if (res.data) {
      prescriptions.value = res.data
    }
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal memuat resep farmasi', 'error')
  } finally {
    isLoadingPrescriptions.value = false
  }
}

function openReview(p: Prescription) {
  selectedPrescription.value = p
  catatanTelaah.value = p.catatan_telaah || 'Resep memenuhi kelayakan klinis dan administratif farmasi'
  isReviewModalOpen.value = true
}

// Verification automatically cuts stock on backend!
async function submitReview() {
  if (!selectedPrescription.value) return
  isSubmitting.value = true
  try {
    await api.post(`/pharmacy/prescriptions/${selectedPrescription.value.id}/review`, {
      catatan_telaah: catatanTelaah.value,
      status: 'TELAAH_OK',
    })
    uiStore.showToast('Telaah resep disetujui & stok depo farmasi otomatis terpotong!', 'success')
    isReviewModalOpen.value = false
    loadPrescriptions()
    loadItems()
    loadMutations()
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal telaah resep', 'error')
  } finally {
    isSubmitting.value = false
  }
}

function openEtiket(p: Prescription, item: PrescriptionItem) {
  selectedPrescription.value = p
  selectedItemForEtiket.value = item
  isEtiketModalOpen.value = true
}

function dispensePrescription(p: Prescription) {
  uiStore.openConfirm({
    title: 'Penyerahan Obat ke Pasien',
    message: `Konfirmasi penyerahan obat untuk pasien "${p.patient_name}" (${p.nomor_resep})? Status resep akan diselesaikan dan diteruskan ke kasir.`,
    confirmText: 'Serahkan Obat',
    variant: 'primary',
    onConfirm: async () => {
      try {
        await api.post(`/pharmacy/prescriptions/${p.id}/dispense`)
        uiStore.showToast('Obat berhasil diserahkan kepada pasien!', 'success')
        loadPrescriptions()
      } catch (err: any) {
        uiStore.showToast(err.message || 'Gagal menyerahkan obat', 'error')
      }
    }
  })
}

// ----------------------------------------------------
// TAB 2: KARTU STOK SEMUA BARANG
// ----------------------------------------------------
const stockMutations = ref<StockMutation[]>([])
const selectedItemForCard = ref<number | ''>('')
const isLoadingMutations = ref(false)

async function loadMutations() {
  isLoadingMutations.value = true
  try {
    const res = await api.get<StockMutation[]>('/logistics/mutations', {
      item_id: selectedItemForCard.value || undefined,
      limit: 100
    })
    if (res.data) {
      stockMutations.value = res.data
    }
  } catch (err: any) {
    console.warn('Load mutations failed', err)
  } finally {
    isLoadingMutations.value = false
  }
}

// ----------------------------------------------------
// TAB 3: VISUALISASI GRAFIK PALING BANYAK ORDER
// ----------------------------------------------------
// Simulated / Aggregated top drug dispensed metrics
const topOrderedDrugs = computed(() => {
  const sample = [
    { name: 'Paracetamol Sirup 120mg/5ml', count: 142, category: 'Analgetik & Antipiretik', percentage: 100 },
    { name: 'Amoxicillin Sirup Kering 125mg/5ml', count: 118, category: 'Antibiotik Pediatrik', percentage: 83 },
    { name: 'Cetirizine Sirup 5mg/5ml', count: 96, category: 'Antihistamin Alergi', percentage: 67 },
    { name: 'Ibuprofen Tablet 400mg', count: 84, category: 'Antiinflamasi', percentage: 59 },
    { name: 'Curcuma Plus Sirup Emulsi', count: 72, category: 'Suplemen & Vitamin', percentage: 50 },
    { name: 'Salbutamol Nebules 2.5mg', count: 58, category: 'Bronkodilator', percentage: 40 },
    { name: 'Gentamicin Salep Kulit 0.1%', count: 44, category: 'Antibiotik Topikal', percentage: 31 },
    { name: 'Oralit Sachet Garam Rehidrasi', count: 39, category: 'Elektrolit Cairan', percentage: 27 },
  ]
  return sample
})

// ----------------------------------------------------
// TAB 4: MONITORING STOK & SAFETY STOCK
// ----------------------------------------------------
const allItems = ref<InventoryItem[]>([])
const searchItemMonitoring = ref('')
const onlyLowStock = ref(false)
const isLoadingItems = ref(false)

async function loadItems() {
  isLoadingItems.value = true
  try {
    const res = await api.get<InventoryItem[]>('/logistics/items', { limit: 100 })
    if (Array.isArray(res.data) && res.data.length > 0) {
      allItems.value = res.data
    } else if ((res.data as any)?.data && Array.isArray((res.data as any).data)) {
      allItems.value = (res.data as any).data
    }
  } catch (err: any) {
    console.warn('Load items failed', err)
  } finally {
    isLoadingItems.value = false
  }
}

const filteredMonitoringItems = computed(() => {
  return allItems.value.filter(it => {
    const q = searchItemMonitoring.value.toLowerCase().trim()
    const matchSearch = !q || it.item_name.toLowerCase().includes(q) || it.item_code.toLowerCase().includes(q)
    const isLow = it.depo_stock <= it.minimum_stock || it.total_stock <= it.minimum_stock
    const matchLow = !onlyLowStock.value || isLow
    return matchSearch && matchLow
  })
})

const monitoringSummary = computed(() => {
  const totalItems = allItems.value.length
  const lowStockCount = allItems.value.filter(i => i.depo_stock <= i.minimum_stock).length
  const criticalCount = allItems.value.filter(i => i.depo_stock === 0).length
  return { totalItems, lowStockCount, criticalCount }
})

// ----------------------------------------------------
// TAB 5: RETUR OBAT PASIEN (FASE 1)
// ----------------------------------------------------
const medicineReturns = ref<any[]>([])
const isLoadingReturns = ref(false)
const searchReturns = ref('')
const isReturnModalOpen = ref(false)
const isSubmittingReturn = ref(false)
const selectedPrescriptionForReturn = ref<Prescription | null>(null)
const returnReason = ref('Pasien alergi / penyesuaian dosis terapi dokter')
const returnItems = ref<{
  item_id: number
  item_name: string
  max_qty: number
  qty: number
  unit_price: number
  subtotal: number
}[]>([])

const returnCols: Column[] = [
  { key: 'nomor_retur', label: 'No. Retur', width: '150px' },
  { key: 'created_at', label: 'Waktu Retur', width: '150px' },
  { key: 'no_rawat', label: 'No. Rawat', width: '140px' },
  { key: 'patient_name', label: 'Nama Pasien' },
  { key: 'alasan', label: 'Alasan Retur' },
  { key: 'total_refund', label: 'Nilai Pengembalian', align: 'right', width: '160px' },
  { key: 'created_by', label: 'Petugas Retur', width: '140px' },
  { key: 'status', label: 'Status', align: 'center', width: '120px' },
  { key: 'actions', label: 'Aksi', align: 'right', width: '120px' },
]

async function loadReturns() {
  isLoadingReturns.value = true
  try {
    const res = await api.get<any[]>('/pharmacy/returns', { search: searchReturns.value })
    if (res.data) {
      medicineReturns.value = res.data
    }
  } catch (err: any) {
    console.warn('Gagal memuat retur', err)
  } finally {
    isLoadingReturns.value = false
  }
}

function openReturnModal(p: Prescription) {
  selectedPrescriptionForReturn.value = p
  returnReason.value = 'Pasien alergi / penyesuaian dosis terapi dokter'
  returnItems.value = (p.items || []).map(it => ({
    item_id: it.item_id,
    item_name: it.item_name,
    max_qty: it.jumlah,
    qty: 0,
    unit_price: it.harga_satuan || 0,
    subtotal: 0
  }))
  isReturnModalOpen.value = true
}

const totalReturnAmount = computed(() => {
  return returnItems.value.reduce((acc, it) => acc + (it.qty * it.unit_price), 0)
})

async function submitReturn() {
  if (!selectedPrescriptionForReturn.value) return
  const validItems = returnItems.value.filter(it => it.qty > 0)
  if (validItems.length === 0) {
    uiStore.showToast('Masukkan minimal 1 jumlah obat yang ingin diretur!', 'warning')
    return
  }
  for (const it of validItems) {
    if (it.qty > it.max_qty) {
      uiStore.showToast(`Jumlah retur ${it.item_name} tidak boleh melebihi resep (${it.max_qty})!`, 'error')
      return
    }
  }

  isSubmittingReturn.value = true
  try {
    await api.post('/pharmacy/returns', {
      visit_id: selectedPrescriptionForReturn.value.visit_id,
      prescription_id: selectedPrescriptionForReturn.value.id,
      alasan: returnReason.value,
      items: validItems.map(it => ({
        item_id: it.item_id,
        item_name: it.item_name,
        qty: Number(it.qty),
        unit_price: Number(it.unit_price)
      }))
    })
    uiStore.showToast('Retur obat berhasil! Stok depo bertambah & tagihan kasir berkurang otomatis.', 'success')
    isReturnModalOpen.value = false
    loadReturns()
    loadPrescriptions()
    loadItems()
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal memproses retur obat', 'error')
  } finally {
    isSubmittingReturn.value = false
  }
}

const selectedReturnForDetail = ref<any | null>(null)
const isReturnDetailModalOpen = ref(false)

function viewReturnDetail(ret: any) {
  selectedReturnForDetail.value = ret
  isReturnDetailModalOpen.value = true
}

onMounted(() => {
  loadPrescriptions()
  loadItems()
  loadMutations()
  loadReturns()
})
</script>

<template>
  <div class="w-full max-w-full space-y-4 text-left">
    <!-- Header -->
    <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs flex flex-col md:flex-row md:items-center justify-between gap-3">
      <div>
        <div class="flex items-center gap-2">
          <Pill class="w-5 h-5 text-[#0284C7]" />
          <h1 class="text-[18px] font-bold text-[#0F172A] tracking-tight">Instalasi Farmasi & Pengelolaan Obat</h1>
        </div>
        <p class="text-[12px] text-[#64748B] mt-0.5">
          Antrean resep, telaah klinis dengan pemotongan stok otomatis, kartu stok mutasi, grafik obat terlaris, dan monitoring stok minimum.
        </p>
      </div>

      <!-- Quick Metrics -->
      <div class="flex items-center gap-2 flex-wrap">
        <div class="px-3 py-1.5 rounded-[6px] bg-[#E0F2FE] border border-[#BAE6FD] text-[#0369A1] text-[12px]">
          <span class="font-medium">Total Item:</span>
          <strong class="ml-1 font-mono">{{ monitoringSummary.totalItems }}</strong>
        </div>

        <div class="px-3 py-1.5 rounded-[6px] bg-[#FEF3C7] border border-[#FDE68A] text-[#B45309] text-[12px]">
          <span class="font-medium">Stok Menipis:</span>
          <strong class="ml-1 font-mono">{{ monitoringSummary.lowStockCount }}</strong>
        </div>

        <div v-if="monitoringSummary.criticalCount > 0" class="px-3 py-1.5 rounded-[6px] bg-[#FEE2E2] border border-[#FCA5A5] text-[#B91C1C] text-[12px]">
          <span class="font-medium">Stok Habis:</span>
          <strong class="ml-1 font-mono">{{ monitoringSummary.criticalCount }}</strong>
        </div>
      </div>
    </div>

    <!-- 5 Pharmacy Tabs Navigation -->
    <div class="flex items-center gap-2 border-b border-[#E2E8F0] pb-2 overflow-x-auto">
      <button
        v-for="t in [
          { id: 'resep', label: '1. Antrean e-Resep & Telaah', icon: FileText },
          { id: 'kartu_stok', label: '2. Kartu Stok Semua Barang', icon: ArrowRightLeft },
          { id: 'grafik_order', label: '3. Visualisasi Obat Paling Banyak Order', icon: BarChart3 },
          { id: 'monitoring', label: '4. Monitoring Stok & Safety Stock', icon: Boxes },
          { id: 'retur', label: '5. Retur Obat Pasien', icon: RotateCcw }
        ]"
        :key="t.id"
        type="button"
        :class="[
          'px-3.5 py-2 rounded-[6px] text-[13px] font-bold transition-all flex items-center gap-2 shrink-0',
          activeTab === t.id
            ? 'bg-[#0284C7] text-white shadow-xs'
            : 'text-[#475569] hover:bg-[#F1F5F9] hover:text-[#0F172A]'
        ]"
        @click="activeTab = t.id as PharmacyTab"
      >
        <component :is="t.icon" class="w-4 h-4" />
        {{ t.label }}
      </button>
    </div>

    <!-- ========================================================================= -->
    <!-- TAB 1: ANTREAN RESEP & TELAAH (AUTO-CUT STOK)                             -->
    <!-- ========================================================================= -->
    <div v-if="activeTab === 'resep'" class="space-y-4">
      <!-- Notice Alert -->
      <div class="p-3 rounded-[6px] bg-[#ECFDF5] border border-[#A7F3D0] text-[#065F46] text-[12px] flex items-center gap-2">
        <CheckCircle2 class="w-4 h-4 text-[#059669] shrink-0" />
        <span>
          <strong>Sistem Terintegrasi:</strong> Ketika staf farmasi melakukan verifikasi telaah resep (status TELAAH_OK), sistem secara otomatis memotong stok fisik depo farmasi & mencatat kartu stok mutasi.
        </span>
      </div>

      <!-- Filters -->
      <div class="bg-white p-3 rounded-[8px] border border-[#E2E8F0] shadow-xs flex flex-wrap items-center justify-between gap-3">
        <div class="flex-1 min-w-[200px]">
          <BaseInput
            v-model="search"
            placeholder="Cari No. Resep, Nama Pasien, No. RM..."
            compact
            @keyup.enter="loadPrescriptions"
          >
            <template #prefix>
              <Search class="w-4 h-4 text-[#94A3B8]" />
            </template>
          </BaseInput>
        </div>

        <div class="w-56">
          <BaseSelect
            v-model="statusFilter"
            :options="[
              { value: '', label: 'Semua Status Resep' },
              { value: 'MENUNGGU_TELAAH', label: 'Menunggu Telaah' },
              { value: 'TELAAH_OK', label: 'Telaah Disetujui' },
              { value: 'PENYIAPAN', label: 'Dalam Penyiapan' },
              { value: 'SELESAI', label: 'Selesai Diserahkan' }
            ]"
            compact
            @change="loadPrescriptions"
          />
        </div>

        <BaseButton variant="outline" size="sm" @click="loadPrescriptions">
          <RefreshCw class="w-3.5 h-3.5 mr-1" />
          Segarkan
        </BaseButton>
      </div>

      <!-- Prescriptions Table -->
      <div class="bg-white border border-[#E2E8F0] rounded-[8px] shadow-xs overflow-hidden">
        <DataTable
          :columns="prescriptionCols"
          :data="prescriptions"
          :loading="isLoadingPrescriptions"
        >
          <template #cell-nomor_resep="{ row }">
            <span class="font-mono font-bold text-[12px] text-[#0284C7] bg-[#F0F9FF] px-2 py-0.5 rounded border border-[#BAE6FD]">
              {{ row.nomor_resep }}
            </span>
          </template>

          <template #cell-no_rm="{ row }">
            <span class="font-mono font-semibold text-[12px] text-[#334155]">{{ row.no_rm }}</span>
          </template>

          <template #cell-patient_name="{ row }">
            <span class="font-bold text-[13px] text-[#0F172A]">{{ row.patient_name }}</span>
          </template>

          <template #cell-doctor_name="{ row }">
            <span class="text-[12px] text-[#475569]">{{ row.doctor_name || 'dr. Spesialis Poliklinik' }}</span>
          </template>

          <template #cell-items_count="{ row }">
            <span class="font-bold font-mono text-[12px] text-[#0F172A]">{{ row.items?.length || 0 }} item</span>
          </template>

          <template #cell-status="{ row }">
            <StatusBadge :status="row.status" />
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center justify-end gap-1.5">
              <BaseButton
                v-if="row.status === 'MENUNGGU_TELAAH'"
                variant="primary"
                size="sm"
                @click="openReview(row)"
              >
                <CheckCircle2 class="w-3.5 h-3.5 mr-1" />
                Verifikasi & Potong Stok
              </BaseButton>

              <BaseButton
                v-else-if="row.status === 'TELAAH_OK' || row.status === 'PENYIAPAN'"
                variant="primary"
                size="sm"
                class="!bg-[#059669]"
                @click="dispensePrescription(row)"
              >
                Serahkan Obat
              </BaseButton>

              <BaseButton
                v-if="row.items && row.items.length > 0"
                variant="outline"
                size="sm"
                @click="openEtiket(row, row.items[0])"
              >
                <Printer class="w-3.5 h-3.5 mr-1" />
                Etiket
              </BaseButton>

              <BaseButton
                v-if="row.status === 'SELESAI' || row.status === 'TELAAH_OK'"
                variant="outline"
                size="sm"
                class="!text-amber-700 !border-amber-300 hover:!bg-amber-50"
                @click="openReturnModal(row)"
              >
                <RotateCcw class="w-3.5 h-3.5 mr-1" />
                Retur Obat
              </BaseButton>
            </div>
          </template>
        </DataTable>
      </div>
    </div>

    <!-- ========================================================================= -->
    <!-- TAB 2: KARTU STOK SEMUA BARANG                                            -->
    <!-- ========================================================================= -->
    <div v-if="activeTab === 'kartu_stok'" class="space-y-4">
      <div class="bg-white p-4 rounded-[8px] border border-[#E2E8F0] shadow-xs space-y-3">
        <div class="flex flex-col md:flex-row md:items-center justify-between gap-3">
          <div class="w-full md:w-96">
            <label class="text-[11px] font-bold text-[#475569] block mb-1">Pilih Barang / Obat untuk Kartu Stok</label>
            <BaseSelect
              v-model="selectedItemForCard"
              :options="[
                { value: '', label: 'Semua Mutasi Barang Farmasi' },
                ...allItems.map(m => ({
                  value: m.id,
                  label: `${m.item_name} (Depo: ${m.depo_stock} | Gudang: ${m.total_stock} ${m.unit})`,
                  badge: m.category_name
                }))
              ]"
              placeholder="-- Pilih Barang --"
              searchable
              @change="loadMutations"
            />
          </div>

          <BaseButton variant="outline" size="sm" @click="loadMutations">
            <RefreshCw class="w-3.5 h-3.5 mr-1" />
            Segarkan Mutasi
          </BaseButton>
        </div>

        <!-- Stock Mutations Table -->
        <div class="border border-[#E2E8F0] rounded-[6px] overflow-hidden">
          <table class="w-full text-left text-[12px]">
            <thead class="bg-[#F8FAFC] text-[#475569] border-b border-[#E2E8F0]">
              <tr>
                <th class="p-2.5 font-bold">Waktu Mutasi</th>
                <th class="p-2.5 font-bold">Nama Barang</th>
                <th class="p-2.5 font-bold text-center">Jenis Transaksi</th>
                <th class="p-2.5 font-bold text-center">Jumlah (Qty)</th>
                <th class="p-2.5 font-bold">Alur Depo (Asal → Tujuan)</th>
                <th class="p-2.5 font-bold">Keterangan / Dokumen</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-[#F1F5F9]">
              <tr v-if="stockMutations.length === 0">
                <td colspan="6" class="p-6 text-center text-[#94A3B8]">
                  Tidak ada riwayat mutasi stok untuk kriteria yang dipilih.
                </td>
              </tr>
              <tr v-for="m in stockMutations" :key="m.id" class="hover:bg-[#F8FAFC]">
                <td class="p-2.5 font-mono text-[11px] text-[#64748B]">
                  {{ new Date(m.created_at).toLocaleString('id-ID') }}
                </td>
                <td class="p-2.5 font-semibold text-[#0F172A]">{{ m.item_name }}</td>
                <td class="p-2.5 text-center">
                  <span
                    class="px-2 py-0.5 rounded text-[10px] font-bold"
                    :class="m.transaction_type.startsWith('IN') ? 'bg-emerald-100 text-emerald-800' :
                            m.transaction_type.startsWith('OUT') ? 'bg-red-100 text-red-800' :
                            'bg-blue-100 text-blue-800'"
                  >
                    {{ m.transaction_type }}
                  </span>
                </td>
                <td class="p-2.5 text-center font-bold font-mono text-[13px]" :class="m.transaction_type.startsWith('OUT') ? 'text-red-600' : 'text-emerald-600'">
                  {{ m.transaction_type.startsWith('OUT') ? '-' : '+' }}{{ m.qty }}
                </td>
                <td class="p-2.5 text-[11px] font-medium text-[#475569]">
                  {{ m.source_depo }} → {{ m.dest_depo }}
                </td>
                <td class="p-2.5 text-[11px] text-[#334155]">{{ m.notes }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- ========================================================================= -->
    <!-- TAB 3: VISUALISASI GRAFIK PALING BANYAK ORDER                             -->
    <!-- ========================================================================= -->
    <div v-if="activeTab === 'grafik_order'" class="space-y-4">
      <div class="bg-white p-4 rounded-[8px] border border-[#E2E8F0] shadow-xs space-y-4">
        <div>
          <h2 class="text-[15px] font-bold text-[#0F172A]">Visualisasi Ranking Obat Paling Banyak Dipesan</h2>
          <p class="text-[12px] text-[#64748B]">Statistik penggunaan obat e-resep poliklinik untuk perencanaan pengadaan logistik.</p>
        </div>

        <!-- High-density bar visualization -->
        <div class="space-y-3 pt-2">
          <div
            v-for="(drug, index) in topOrderedDrugs"
            :key="drug.name"
            class="p-3 rounded-[6px] border border-[#E2E8F0] bg-[#F8FAFC] space-y-1.5"
          >
            <div class="flex items-center justify-between text-[13px]">
              <div class="flex items-center gap-2">
                <span
                  class="w-5 h-5 rounded-full flex items-center justify-center text-[10px] font-bold text-white shrink-0"
                  :class="index === 0 ? 'bg-amber-500' : index === 1 ? 'bg-slate-400' : index === 2 ? 'bg-amber-700' : 'bg-slate-300 text-slate-700'"
                >
                  {{ index + 1 }}
                </span>
                <span class="font-bold text-[#0F172A]">{{ drug.name }}</span>
                <span class="text-[11px] text-[#64748B]">({{ drug.category }})</span>
              </div>

              <div class="text-right font-mono">
                <span class="font-extrabold text-[#0284C7] text-[14px]">{{ drug.count }}</span>
                <span class="text-[11px] text-[#64748B] ml-1">kali diresepkan</span>
              </div>
            </div>

            <!-- Progress bar -->
            <div class="w-full h-2.5 bg-[#E2E8F0] rounded-full overflow-hidden">
              <div
                class="h-full rounded-full transition-all duration-500"
                :class="index === 0 ? 'bg-[#0284C7]' : index < 3 ? 'bg-[#0369A1]' : 'bg-[#38BDF8]'"
                :style="{ width: `${drug.percentage}%` }"
              ></div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- ========================================================================= -->
    <!-- TAB 4: MONITORING STOK & SAFETY STOCK                                     -->
    <!-- ========================================================================= -->
    <div v-if="activeTab === 'monitoring'" class="space-y-4">
      <div class="bg-white p-3 rounded-[8px] border border-[#E2E8F0] shadow-xs flex flex-wrap items-center justify-between gap-3">
        <div class="flex-1 min-w-[200px]">
          <BaseInput
            v-model="searchItemMonitoring"
            placeholder="Cari nama obat, kode barang..."
            compact
          >
            <template #prefix>
              <Search class="w-4 h-4 text-[#94A3B8]" />
            </template>
          </BaseInput>
        </div>

        <label class="flex items-center gap-2 text-[12px] font-semibold text-[#B45309] cursor-pointer">
          <input v-model="onlyLowStock" type="checkbox" class="rounded text-[#B45309]" />
          <span>Hanya tampilkan stok menipis / kritis</span>
        </label>

        <BaseButton variant="outline" size="sm" @click="loadItems">
          <RefreshCw class="w-3.5 h-3.5 mr-1" />
          Segarkan Stok
        </BaseButton>
      </div>

      <!-- Monitoring Table -->
      <div class="bg-white border border-[#E2E8F0] rounded-[8px] shadow-xs overflow-hidden">
        <table class="w-full text-left text-[12px]">
          <thead class="bg-[#F8FAFC] text-[#475569] border-b border-[#E2E8F0]">
            <tr>
              <th class="p-2.5 font-bold">Kode Barang</th>
              <th class="p-2.5 font-bold">Nama Obat / Sediaan</th>
              <th class="p-2.5 font-bold">Kategori</th>
              <th class="p-2.5 font-bold text-center">Stok Depo</th>
              <th class="p-2.5 font-bold text-center">Stok Gudang</th>
              <th class="p-2.5 font-bold text-center">Safety Stock</th>
              <th class="p-2.5 font-bold text-center">Status</th>
              <th class="p-2.5 font-bold">Rak</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[#F1F5F9]">
            <tr v-if="filteredMonitoringItems.length === 0">
              <td colspan="8" class="p-6 text-center text-[#94A3B8]">
                Tidak ada data barang yang sesuai dengan filter.
              </td>
            </tr>
            <tr v-for="it in filteredMonitoringItems" :key="it.id" class="hover:bg-[#F8FAFC]">
              <td class="p-2.5 font-mono text-[11px] text-[#0284C7] font-semibold">{{ it.item_code }}</td>
              <td class="p-2.5 font-semibold text-[#0F172A]">{{ it.item_name }}</td>
              <td class="p-2.5 text-[#475569]">{{ it.category_name || 'Obat' }}</td>
              <td class="p-2.5 text-center font-bold font-mono text-[13px]" :class="it.depo_stock <= it.minimum_stock ? 'text-amber-600' : 'text-slate-800'">
                {{ it.depo_stock }} {{ it.unit }}
              </td>
              <td class="p-2.5 text-center font-bold font-mono text-[13px] text-slate-800">
                {{ it.total_stock }} {{ it.unit }}
              </td>
              <td class="p-2.5 text-center font-mono text-[12px] text-[#64748B]">
                {{ it.minimum_stock }} {{ it.unit }}
              </td>
              <td class="p-2.5 text-center">
                <span
                  v-if="it.depo_stock === 0"
                  class="px-2 py-0.5 rounded text-[10px] font-bold bg-red-100 text-red-800"
                >
                  HABIS
                </span>
                <span
                  v-else-if="it.depo_stock <= it.minimum_stock"
                  class="px-2 py-0.5 rounded text-[10px] font-bold bg-amber-100 text-amber-800"
                >
                  MENIPIS
                </span>
                <span
                  v-else
                  class="px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-100 text-emerald-800"
                >
                  AMAN
                </span>
              </td>
              <td class="p-2.5 text-[11px] font-mono text-[#64748B]">{{ it.rack_location || '-' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- ========================================================================= -->
    <!-- TAB 5: RETUR OBAT PASIEN (FASE 1)                                         -->
    <!-- ========================================================================= -->
    <div v-if="activeTab === 'retur'" class="space-y-4">
      <div class="p-3 rounded-[6px] bg-[#FEF3C7] border border-[#FDE68A] text-[#92400E] text-[12px] flex items-center gap-2">
        <RotateCcw class="w-4 h-4 text-[#D97706] shrink-0" />
        <span>
          <strong>Retur Terintegrasi ACID:</strong> Pengembalian obat pasien secara otomatis memulihkan stok fisik di Depo Farmasi, mencatat riwayat kartu stok mutasi (IN_RETUR), serta memotong sisa tagihan di kasir secara langsung dengan row-level locking.
        </span>
      </div>

      <div class="bg-white p-3 rounded-[8px] border border-[#E2E8F0] shadow-xs flex flex-wrap items-center justify-between gap-3">
        <div class="flex-1 min-w-[240px]">
          <BaseInput
            v-model="searchReturns"
            placeholder="Cari No. Retur, Nama Pasien, No. RM..."
            compact
            @keyup.enter="loadReturns"
          >
            <template #prefix>
              <Search class="w-4 h-4 text-[#94A3B8]" />
            </template>
          </BaseInput>
        </div>

        <div class="flex items-center gap-2">
          <BaseButton variant="outline" size="sm" @click="loadReturns">
            <RefreshCw class="w-3.5 h-3.5 mr-1" />
            Segarkan
          </BaseButton>
        </div>
      </div>

      <div class="bg-white border border-[#E2E8F0] rounded-[8px] shadow-xs overflow-hidden">
        <DataTable
          :columns="returnCols"
          :data="medicineReturns"
          :loading="isLoadingReturns"
        >
          <template #cell-nomor_retur="{ row }">
            <span class="font-mono font-bold text-[12px] text-[#0284C7] bg-[#F0F9FF] px-2 py-0.5 rounded border border-[#BAE6FD]">
              {{ row.nomor_retur }}
            </span>
          </template>

          <template #cell-created_at="{ row }">
            <span class="text-[12px] text-[#475569]">
              {{ new Date(row.created_at).toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'short' }) }}
            </span>
          </template>

          <template #cell-no_rawat="{ row }">
            <span class="font-mono text-[12px] font-semibold text-[#334155]">{{ row.no_rawat || '-' }}</span>
          </template>

          <template #cell-patient_name="{ row }">
            <span class="font-bold text-[13px] text-[#0F172A]">{{ row.patient_name }}</span>
          </template>

          <template #cell-alasan="{ row }">
            <span class="text-[12px] text-[#475569] italic">{{ row.alasan || '-' }}</span>
          </template>

          <template #cell-total_refund="{ row }">
            <span class="font-mono font-bold text-[13px] text-emerald-600">
              Rp {{ Number(row.total_refund || 0).toLocaleString('id-ID') }}
            </span>
          </template>

          <template #cell-created_by="{ row }">
            <span class="text-[12px] text-[#64748B]">{{ row.created_by || 'Apoteker' }}</span>
          </template>

          <template #cell-status="{ row }">
            <span class="px-2 py-0.5 rounded text-[11px] font-bold bg-emerald-100 text-emerald-800">
              {{ row.status || 'SELESAI' }}
            </span>
          </template>

          <template #cell-actions="{ row }">
            <BaseButton variant="outline" size="sm" @click="viewReturnDetail(row)">
              <FileText class="w-3.5 h-3.5 mr-1" />
              Detail
            </BaseButton>
          </template>
        </DataTable>
      </div>
    </div>

    <!-- MODAL: Buat Retur Obat Pasien -->
    <BaseModal
      v-model="isReturnModalOpen"
      title="Form Pengembalian (Retur) Obat Pasien"
      max-width="2xl"
    >
      <div v-if="selectedPrescriptionForReturn" class="space-y-4 text-[13px]">
        <div class="p-3 bg-[#F8FAFC] rounded border border-[#E2E8F0] grid grid-cols-2 gap-2 text-[12px]">
          <div><span class="text-[#64748B]">No. Resep:</span> <strong class="font-mono text-[#0284C7]">{{ selectedPrescriptionForReturn.nomor_resep }}</strong></div>
          <div><span class="text-[#64748B]">Nama Pasien:</span> <strong>{{ selectedPrescriptionForReturn.patient_name }} ({{ selectedPrescriptionForReturn.no_rm }})</strong></div>
          <div><span class="text-[#64748B]">Dokter Penulis:</span> <strong>{{ selectedPrescriptionForReturn.doctor_name }}</strong></div>
          <div><span class="text-[#64748B]">Status Resep:</span> <strong class="text-emerald-700">{{ selectedPrescriptionForReturn.status }}</strong></div>
        </div>

        <div>
          <label class="text-[12px] font-bold text-[#334155] block mb-1">Alasan Retur Obat *</label>
          <input
            v-model="returnReason"
            type="text"
            placeholder="Misal: Pasien alergi, dokter mengganti regimen terapi, dll."
            class="w-full p-2.5 text-[12px] border border-[#CBD5E1] rounded outline-none focus:border-[#0284C7]"
          />
        </div>

        <div>
          <label class="text-[12px] font-bold text-[#334155] block mb-1.5">Pilih Item & Jumlah Obat yang Diretur</label>
          <div class="border border-[#E2E8F0] rounded overflow-hidden">
            <table class="w-full text-left text-[12px]">
              <thead class="bg-[#F8FAFC] text-[#475569] border-b">
                <tr>
                  <th class="p-2.5 font-bold">Nama Obat</th>
                  <th class="p-2.5 font-bold text-center">Jumlah e-Resep</th>
                  <th class="p-2.5 font-bold text-center w-28">Jumlah Retur</th>
                  <th class="p-2.5 font-bold text-right">Harga Satuan</th>
                  <th class="p-2.5 font-bold text-right">Subtotal Refund</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-[#F1F5F9]">
                <tr v-for="it in returnItems" :key="it.item_id" class="hover:bg-[#F8FAFC]">
                  <td class="p-2.5 font-semibold text-[#0F172A]">{{ it.item_name }}</td>
                  <td class="p-2.5 text-center font-mono font-bold text-[#475569]">{{ it.max_qty }}</td>
                  <td class="p-2.5 text-center">
                    <input
                      v-model.number="it.qty"
                      type="number"
                      min="0"
                      :max="it.max_qty"
                      class="w-20 p-1 text-center font-mono font-bold border border-[#CBD5E1] rounded text-[13px] text-[#0284C7] focus:border-[#0284C7]"
                    />
                  </td>
                  <td class="p-2.5 text-right font-mono text-[#64748B]">
                    Rp {{ Number(it.unit_price).toLocaleString('id-ID') }}
                  </td>
                  <td class="p-2.5 text-right font-mono font-bold text-emerald-700">
                    Rp {{ Number(it.qty * it.unit_price).toLocaleString('id-ID') }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div class="flex items-center justify-between p-3 rounded bg-emerald-50 border border-emerald-200">
          <span class="text-[13px] font-bold text-emerald-900">Total Nilai Pengembalian (Refund):</span>
          <span class="text-[16px] font-extrabold font-mono text-emerald-800">
            Rp {{ Number(totalReturnAmount).toLocaleString('id-ID') }}
          </span>
        </div>

        <div class="p-2.5 rounded bg-amber-50 border border-amber-200 text-amber-800 text-[11px] space-y-1">
          <p>⚠️ <strong>Catatan Row Locking ACID:</strong></p>
          <p>• Stok fisik di Depo Farmasi akan bertambah secara instan.</p>
          <p>• Jika invoice kasir sudah terbit, total tagihan kasir pasien otomatis dipotong sejumlah refund di atas.</p>
        </div>
      </div>

      <template #footer>
        <BaseButton variant="outline" @click="isReturnModalOpen = false">Batal</BaseButton>
        <BaseButton
          variant="primary"
          :loading="isSubmittingReturn"
          :disabled="totalReturnAmount <= 0"
          @click="submitReturn"
        >
          <RotateCcw class="w-4 h-4 mr-1.5" />
          Proses Retur & Kembalikan Stok
        </BaseButton>
      </template>
    </BaseModal>

    <!-- MODAL: Detail / Cetak Bukti Retur -->
    <BaseModal
      v-model="isReturnDetailModalOpen"
      title="Bukti Pengembalian Obat Pasien (Retur)"
      max-width="lg"
    >
      <div v-if="selectedReturnForDetail" class="p-4 bg-white border border-[#CBD5E1] rounded space-y-3 font-sans text-[12px]">
        <div class="border-b border-[#0F172A] pb-2 text-center">
          <h3 class="font-bold text-[14px] uppercase text-[#0284C7]">BUKTI RETUR FARMASI RESMI</h3>
          <p class="text-[10px] text-[#475569]">KLINIK & RUMAH SAKIT SIMRS-ASOY • Jl. Kesehatan No. 10 Yogyakarta</p>
        </div>

        <div class="grid grid-cols-2 gap-2 text-[11px] pt-1">
          <div><span class="text-[#64748B]">No. Retur:</span> <strong class="font-mono text-[#0284C7]">{{ selectedReturnForDetail.nomor_retur }}</strong></div>
          <div><span class="text-[#64748B]">Waktu:</span> <strong>{{ new Date(selectedReturnForDetail.created_at).toLocaleString('id-ID') }}</strong></div>
          <div><span class="text-[#64748B]">Pasien:</span> <strong>{{ selectedReturnForDetail.patient_name }}</strong></div>
          <div><span class="text-[#64748B]">No. Rawat:</span> <strong class="font-mono">{{ selectedReturnForDetail.no_rawat }}</strong></div>
          <div><span class="text-[#64748B]">Petugas Retur:</span> <strong>{{ selectedReturnForDetail.created_by }}</strong></div>
          <div><span class="text-[#64748B]">Alasan:</span> <strong class="italic">{{ selectedReturnForDetail.alasan }}</strong></div>
        </div>

        <div class="border border-[#CBD5E1] rounded overflow-hidden">
          <table class="w-full text-left text-[11px]">
            <thead class="bg-[#F8FAFC] border-b">
              <tr>
                <th class="p-2">Item Obat</th>
                <th class="p-2 text-center">Qty</th>
                <th class="p-2 text-right">Harga Satuan</th>
                <th class="p-2 text-right">Subtotal</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-[#F1F5F9]">
              <tr v-for="it in (selectedReturnForDetail.items || [])" :key="it.id">
                <td class="p-2 font-medium">{{ it.item_name }}</td>
                <td class="p-2 text-center font-mono font-bold">{{ it.qty }}</td>
                <td class="p-2 text-right font-mono">Rp {{ Number(it.unit_price).toLocaleString('id-ID') }}</td>
                <td class="p-2 text-right font-mono font-bold">Rp {{ Number(it.subtotal).toLocaleString('id-ID') }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="flex justify-between items-center pt-2 border-t">
          <span class="font-bold text-[13px]">Total Nilai Refund:</span>
          <span class="font-extrabold font-mono text-[15px] text-emerald-700">
            Rp {{ Number(selectedReturnForDetail.total_refund).toLocaleString('id-ID') }}
          </span>
        </div>
      </div>

      <template #footer>
        <BaseButton variant="outline" @click="isReturnDetailModalOpen = false">Tutup</BaseButton>
        <BaseButton variant="primary" @click="printEtiket">
          <Printer class="w-4 h-4 mr-1.5" />
          Cetak Bukti Retur
        </BaseButton>
      </template>
    </BaseModal>

    <!-- MODAL: Telaah Resep Farmasi -->
    <BaseModal
      v-model="isReviewModalOpen"
      title="Telaah Resep Obat & Verifikasi Klinis"
      max-width="lg"
    >
      <div v-if="selectedPrescription" class="space-y-3 text-[13px]">
        <div class="p-3 bg-[#F8FAFC] rounded border border-[#E2E8F0] space-y-1 text-[12px]">
          <div><span class="text-[#64748B]">No. Resep:</span> <strong class="font-mono text-[#0284C7]">{{ selectedPrescription.nomor_resep }}</strong></div>
          <div><span class="text-[#64748B]">Pasien:</span> <strong>{{ selectedPrescription.patient_name }} (RM: {{ selectedPrescription.no_rm }})</strong></div>
          <div><span class="text-[#64748B]">Dokter:</span> <strong>{{ selectedPrescription.doctor_name }}</strong></div>
        </div>

        <div class="border border-[#E2E8F0] rounded overflow-hidden">
          <table class="w-full text-left text-[12px]">
            <thead class="bg-[#F8FAFC] text-[#475569] border-b">
              <tr>
                <th class="p-2">Item Obat</th>
                <th class="p-2">Signa</th>
                <th class="p-2 text-center">Jumlah</th>
              </tr>
            </thead>
            <tbody class="divide-y">
              <tr v-for="it in selectedPrescription.items" :key="it.id">
                <td class="p-2 font-semibold">{{ it.item_name }}</td>
                <td class="p-2">{{ it.signa }}</td>
                <td class="p-2 text-center font-bold font-mono">{{ it.jumlah }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div>
          <label class="text-[12px] font-semibold text-[#334155] block mb-1">Catatan Telaah Klinis & Administratif Farmasi</label>
          <textarea
            v-model="catatanTelaah"
            rows="2"
            class="w-full p-2.5 text-[12px] border border-[#CBD5E1] rounded outline-none focus:border-[#0284C7]"
          ></textarea>
        </div>

        <div class="p-2.5 rounded bg-amber-50 border border-amber-200 text-amber-800 text-[11px]">
          ⚠️ Mengklik "Setujui Telaah" akan otomatis mengurangi stok depo farmasi untuk seluruh item di atas.
        </div>
      </div>

      <template #footer>
        <BaseButton variant="outline" @click="isReviewModalOpen = false">Batal</BaseButton>
        <BaseButton variant="primary" :loading="isSubmitting" @click="submitReview">
          Setujui Telaah & Potong Stok
        </BaseButton>
      </template>
    </BaseModal>

    <!-- MODAL: Cetak Etiket Obat -->
    <BaseModal
      v-model="isEtiketModalOpen"
      title="Cetak Etiket Aturan Pakai Obat"
      max-width="md"
    >
      <div v-if="selectedPrescription && selectedItemForEtiket" class="p-4 bg-white border border-[#0F172A] rounded space-y-2 text-center font-sans">
        <div class="border-b border-[#0F172A] pb-1">
          <h3 class="font-bold text-[13px] uppercase text-[#0284C7]">INSTALASI FARMASI KLINIK SIMRS-ASOY</h3>
          <p class="text-[10px] text-[#475569]">Jl. Kesehatan No. 10 Yogyakarta • Telp: (0274) 555-889</p>
        </div>

        <div class="text-[11px] text-left pt-1 space-y-0.5">
          <p>No. Resep: <strong>{{ selectedPrescription.nomor_resep }}</strong></p>
          <p>Tanggal: <strong>{{ new Date().toLocaleDateString('id-ID') }}</strong></p>
          <p>Pasien: <strong>{{ selectedPrescription.patient_name }} ({{ selectedPrescription.no_rm }})</strong></p>
        </div>

        <div class="my-2 p-2 rounded bg-[#F8FAFC] border border-[#CBD5E1]">
          <h4 class="font-bold text-[14px] text-[#0F172A]">{{ selectedItemForEtiket.item_name }}</h4>
          <p class="text-[14px] font-extrabold text-[#0284C7] mt-1">{{ selectedItemForEtiket.signa }}</p>
          <p class="text-[12px] font-semibold text-[#475569]">{{ selectedItemForEtiket.aturan_pakai }}</p>
        </div>

        <p class="text-[10px] text-[#64748B]">Semoga lekas sembuh • Jauhkan dari jangkauan anak-anak</p>
      </div>

      <template #footer>
        <BaseButton variant="secondary" @click="isEtiketModalOpen = false">Tutup</BaseButton>
        <BaseButton variant="primary" @click="printEtiket">
          <Printer class="w-4 h-4 mr-1" />
          Cetak Etiket
        </BaseButton>
      </template>
    </BaseModal>
  </div>
</template>
