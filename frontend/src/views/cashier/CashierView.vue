<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '@/api/client'
import type { Invoice, CashierShift, Visit, Promo, BillingPayment } from '@/types'
import { useUiStore } from '@/stores/ui'
import { useAuthStore } from '@/stores/auth'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import BaseModal from '@/components/common/BaseModal.vue'
import ConfirmModal from '@/components/common/ConfirmModal.vue'
import {
  CreditCard,
  Receipt,
  Clock,
  DollarSign,
  UserCheck,
  Search,
  RefreshCw,
  Printer,
  CheckCircle2,
  AlertCircle,
  FileText,
  Calculator,
  Wallet,
  Building,
  ArrowRight,
  History,
  Lock,
  Plus,
  Trash2,
  QrCode
} from 'lucide-vue-next'

const uiStore = useUiStore()
const authStore = useAuthStore()

const loading = ref(false)
const activeTab = ref<'antrean' | 'riwayat_shift'>('antrean')

// Shift State
const activeShift = ref<CashierShift | null>(null)
const shiftHistory = ref<CashierShift[]>([])
const showOpenShiftModal = ref(false)
const showCloseShiftModal = ref(false)
const submittingShift = ref(false)

const openShiftForm = ref({
  shift_name: 'Shift Pagi (07:00 - 14:00)',
  saldo_awal: 200000,
})

const closeShiftForm = ref({
  saldo_akhir_fisik: 0,
  catatan_shift: '',
})

// Invoices & Queue
const visitsWaiting = ref<Visit[]>([])
const invoices = ref<Invoice[]>([])
const searchQuery = ref('')
const selectedInvoiceStatus = ref<string>('SEMUA')

// Payment Modal State
const showPaymentModal = ref(false)
const selectedInvoice = ref<Invoice | null>(null)
const submittingPayment = ref(false)

// Master Promo State
const promos = ref<Promo[]>([])
const selectedPromoCode = ref('')

// Pending Order Verification Check
const pendingBlockReason = ref<string | null>(null)
const isCheckingPendingOrders = ref(false)

const paymentForm = ref({
  nominal: 0,
  metode_bayar: 'TUNAI',
  referensi_pembayaran: '',
  catatan: '',
})

// ----------------------------------------------------
// SPLIT PAYMENT STATE (FASE 1)
// ----------------------------------------------------
const isSplitPayment = ref(false)
interface SplitLine {
  metode_bayar: string
  nominal: number
  referensi_pembayaran: string
  catatan: string
}
const splitPayments = ref<SplitLine[]>([
  { metode_bayar: 'QRIS', nominal: 0, referensi_pembayaran: '', catatan: '' },
  { metode_bayar: 'TUNAI', nominal: 0, referensi_pembayaran: '', catatan: '' }
])

const totalSplitPaid = computed(() => {
  return splitPayments.value.reduce((acc, p) => acc + (Number(p.nominal) || 0), 0)
})

const splitDifference = computed(() => {
  return totalSplitPaid.value - finalTotalTagihan.value
})

function addSplitLine() {
  splitPayments.value.push({
    metode_bayar: 'TUNAI',
    nominal: 0,
    referensi_pembayaran: '',
    catatan: ''
  })
}

function removeSplitLine(idx: number) {
  if (splitPayments.value.length > 1) {
    splitPayments.value.splice(idx, 1)
  }
}

// Promo Discount & Final Tagihan Calculation
const promoDiscount = computed(() => {
  if (!selectedPromoCode.value || !selectedInvoice.value) return 0
  const promo = promos.value.find(p => p.code === selectedPromoCode.value)
  if (!promo) return 0
  if (selectedInvoice.value.subtotal < promo.min_transaction) return 0
  if (promo.discount_type === 'PERCENTAGE') {
    return (selectedInvoice.value.subtotal * promo.discount_value) / 100
  }
  return promo.discount_value
})

const finalTotalTagihan = computed(() => {
  if (!selectedInvoice.value) return 0
  return Math.max(0, (selectedInvoice.value.total_tagihan || 0) - promoDiscount.value)
})

async function fetchPromos() {
  try {
    const res = await api.get<Promo[]>('/master/promos')
    promos.value = res.data || []
  } catch (e) {
    console.warn('Promos fetch error', e)
  }
}

async function checkPendingOrders(visitId: string) {
  pendingBlockReason.value = null
  isCheckingPendingOrders.value = true
  try {
    // 1. Cek resep farmasi
    const resPresc = await api.get<any>(`/pharmacy/prescriptions/visit/${visitId}`)
    if (resPresc.data && (resPresc.data.status === 'MENUNGGU_TELAAH' || resPresc.data.status === 'PENYIAPAN')) {
      pendingBlockReason.value = `Resep obat pasien masih berstatus "${resPresc.data.status}" di Instalasi Farmasi. Pembayaran belum dapat diproses sampai telaah/penyiapan selesai.`
      return
    }

    // 2. Cek order penunjang (Lab, Rad, Fisio)
    const resOrd = await api.get<any[]>(`/support/orders/visit/${visitId}`)
    if (resOrd.data) {
      const unfinished = resOrd.data.find(o => o.status === 'MENUNGGU' || o.status === 'DALAM_PROSES')
      if (unfinished) {
        pendingBlockReason.value = `Order penunjang ${unfinished.jenis_order} (${unfinished.nama_tindakan}) masih berstatus "${unfinished.status}". Pembayaran belum dapat diproses hingga hasil divalidasi.`
        return
      }
    }
  } catch (e) {
    console.warn('Pending order check warning', e)
  } finally {
    isCheckingPendingOrders.value = false
  }
}

// Receipt Modal State
const showReceiptModal = ref(false)
const invoiceToPrint = ref<Invoice | null>(null)

// Shift Methods
const fetchActiveShift = async () => {
  try {
    const res = await api.get<CashierShift>('/billing/shifts/active')
    activeShift.value = res.data || null
  } catch (err) {
    activeShift.value = null
  }
}

const openShift = async () => {
  if (openShiftForm.value.saldo_awal < 0) {
    uiStore.showToast('Saldo awal tidak boleh negatif', 'warning')
    return
  }

  submittingShift.value = true
  try {
    const res = await api.post<CashierShift>('/billing/shifts/open', openShiftForm.value)
    activeShift.value = res.data
    uiStore.showToast('Shift kasir berhasil dibuka. Selamat bertugas!', 'success')
    showOpenShiftModal.value = false
    await fetchInvoices()
    await fetchVisitsWaiting()
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal membuka shift kasir', 'error')
  } finally {
    submittingShift.value = false
  }
}

const openCloseShiftDialog = () => {
  if (!activeShift.value) return
  closeShiftForm.value = {
    saldo_akhir_fisik: (activeShift.value.saldo_awal + activeShift.value.total_tunai_sistem) || 0,
    catatan_shift: '',
  }
  showCloseShiftModal.value = true
}

const calculatedSelisih = computed(() => {
  if (!activeShift.value) return 0
  const expected = (activeShift.value.saldo_awal || 0) + (activeShift.value.total_tunai_sistem || 0)
  return closeShiftForm.value.saldo_akhir_fisik - expected
})

const closeShift = async () => {
  if (!activeShift.value) return

  submittingShift.value = true
  try {
    await api.post(`/billing/shifts/${activeShift.value.id}/close`, closeShiftForm.value)
    uiStore.showToast('Shift kasir berhasil ditutup dan laporan rekonsiliasi tersimpan', 'success')
    showCloseShiftModal.value = false
    activeShift.value = null
    await fetchShiftHistory()
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal menutup shift kasir', 'error')
  } finally {
    submittingShift.value = false
  }
}

const fetchShiftHistory = async () => {
  try {
    const res = await api.get<CashierShift[]>('/billing/shifts/history', { limit: 15 })
    shiftHistory.value = res.data || []
  } catch (err) {
    console.error('Gagal ambil riwayat shift:', err)
  }
}

// Queue & Invoices Methods
const fetchVisitsWaiting = async () => {
  try {
    const res = await api.get<Visit[]>('/visits', { status: 'KASIR', limit: 50 })
    visitsWaiting.value = res.data || []
  } catch (err) {
    console.error('Gagal ambil antrean kasir:', err)
  }
}

const fetchInvoices = async () => {
  loading.value = true
  try {
    const params: Record<string, any> = {
      page: 1,
      limit: 50,
    }
    if (selectedInvoiceStatus.value !== 'SEMUA') params.status = selectedInvoiceStatus.value
    if (searchQuery.value) params.search = searchQuery.value

    const res = await api.get<Invoice[]>('/billing/invoices', params)
    invoices.value = res.data || []
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal memuat invoice kasir', 'error')
  } finally {
    loading.value = false
  }
}

// Generate Billing from Visit
const generateBilling = async (visit: Visit) => {
  if (!activeShift.value) {
    uiStore.showToast('Harap buka shift kasir terlebih dahulu sebelum memproses tagihan', 'warning')
    showOpenShiftModal.value = true
    return
  }

  try {
    const res = await api.post<Invoice>(`/billing/generate/${visit.id}`)
    uiStore.showToast(`Tagihan untuk pasien ${visit.patient_name} berhasil dibuat`, 'success')
    await fetchVisitsWaiting()
    await fetchInvoices()
    if (res.data) {
      openPaymentModal(res.data)
    }
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal mengkalkulasi tagihan', 'error')
  }
}

// Open Payment Modal
const openPaymentModal = async (inv: Invoice) => {
  if (!activeShift.value) {
    uiStore.showToast('Buka shift kasir terlebih dahulu untuk memproses pembayaran', 'warning')
    showOpenShiftModal.value = true
    return
  }

  // Fetch full details if items not loaded
  try {
    const res = await api.get<Invoice>(`/billing/invoices/${inv.id}`)
    selectedInvoice.value = res.data
  } catch (err) {
    selectedInvoice.value = inv
  }

  selectedPromoCode.value = ''
  paymentForm.value = {
    nominal: selectedInvoice.value?.total_tagihan || 0,
    metode_bayar: 'TUNAI',
    referensi_pembayaran: '',
    catatan: '',
  }

  // Initialize split payments
  isSplitPayment.value = false
  const half = Math.round((selectedInvoice.value?.total_tagihan || 0) / 2)
  splitPayments.value = [
    { metode_bayar: 'QRIS', nominal: half, referensi_pembayaran: '', catatan: '' },
    { metode_bayar: 'TUNAI', nominal: Math.max(0, (selectedInvoice.value?.total_tagihan || 0) - half), referensi_pembayaran: '', catatan: '' }
  ]

  // Check if any prescription or lab order is unfinished
  if (selectedInvoice.value?.visit_id) {
    await checkPendingOrders(selectedInvoice.value.visit_id)
  }

  showPaymentModal.value = true
}

// Kembalian calculation
const kembalian = computed(() => {
  if (!selectedInvoice.value) return 0
  if (paymentForm.value.metode_bayar !== 'TUNAI') return 0
  const change = paymentForm.value.nominal - finalTotalTagihan.value
  return change > 0 ? change : 0
})

const processPayment = async () => {
  if (!selectedInvoice.value) return
  if (pendingBlockReason.value) {
    uiStore.showToast(pendingBlockReason.value, 'error')
    return
  }

  const currentPromoCode = selectedPromoCode.value
  const promoObj = promos.value.find(p => p.code === currentPromoCode)
  const currentPromoName = promoObj?.name || ''
  const currentPromoDiscount = promoDiscount.value

  if (isSplitPayment.value) {
    if (totalSplitPaid.value < finalTotalTagihan.value) {
      uiStore.showToast(`Total pembayaran split (Rp ${totalSplitPaid.value.toLocaleString('id-ID')}) kurang dari total tagihan (Rp ${finalTotalTagihan.value.toLocaleString('id-ID')})`, 'warning')
      return
    }

    submittingPayment.value = true
    try {
      const payload = {
        payments: splitPayments.value.map(p => ({
          metode_bayar: p.metode_bayar,
          nominal: Number(p.nominal),
          referensi_pembayaran: p.referensi_pembayaran,
          catatan: p.catatan || (currentPromoCode ? `Promo: ${currentPromoCode}` : '')
        })),
        promo_code: currentPromoCode || undefined,
        promo_name: currentPromoName || undefined,
        promo_discount: currentPromoDiscount
      }

      const res = await api.post<Invoice>(`/billing/invoices/${selectedInvoice.value.id}/pay-split`, payload)
      uiStore.showToast('Pembayaran split (campuran) berhasil diproses! Status tagihan: LUNAS', 'success')
      showPaymentModal.value = false

      const updated = res.data || selectedInvoice.value
      invoiceToPrint.value = {
        ...updated,
        promo_code: currentPromoCode || updated.promo_code,
        promo_name: currentPromoName || updated.promo_name,
        diskon: (updated.diskon || 0) > 0 ? updated.diskon : currentPromoDiscount,
        subtotal: updated.subtotal || selectedInvoice.value.subtotal || (selectedInvoice.value.total_tagihan + currentPromoDiscount),
        total_tagihan: finalTotalTagihan.value,
        payments: splitPayments.value.map((p, idx) => ({
          id: `split-${idx}`,
          nomor_transaksi: `SPLIT-${Date.now()}-${idx + 1}`,
          invoice_id: selectedInvoice.value!.id,
          kasir_id: authStore.user?.id || '',
          kasir_name: authStore.user?.full_name || 'Kasir SIMRS',
          nominal: Number(p.nominal),
          metode_bayar: p.metode_bayar as any,
          referensi_pembayaran: p.referensi_pembayaran,
          catatan: p.catatan,
          uang_diterima: Number(p.nominal),
          kembalian: 0,
          created_at: new Date().toISOString()
        }))
      }
      showReceiptModal.value = true

      await fetchActiveShift()
      await fetchInvoices()
      await fetchVisitsWaiting()
    } catch (err: any) {
      uiStore.showToast(err.message || 'Gagal memproses pembayaran split', 'error')
    } finally {
      submittingPayment.value = false
    }
    return
  }

  // Single payment flow
  if (paymentForm.value.nominal < finalTotalTagihan.value) {
    uiStore.showToast('Nominal pembayaran kurang dari total tagihan', 'warning')
    return
  }

  const currentTendered = paymentForm.value.nominal
  const currentChange = kembalian.value
  const currentMethod = paymentForm.value.metode_bayar

  submittingPayment.value = true
  try {
    const payload = {
      ...paymentForm.value,
      nominal: finalTotalTagihan.value,
      uang_diterima: currentTendered,
      kembalian: currentChange,
      promo_code: currentPromoCode || undefined,
      promo_name: currentPromoName || undefined,
      promo_discount: currentPromoDiscount,
      catatan: currentPromoCode ? `Promo: ${currentPromoCode} (Diskon Rp ${currentPromoDiscount.toLocaleString('id-ID')})` : paymentForm.value.catatan
    }

    const res = await api.post<Invoice>(`/billing/invoices/${selectedInvoice.value.id}/pay`, payload)
    uiStore.showToast('Pembayaran berhasil diproses! Status tagihan: LUNAS', 'success')
    showPaymentModal.value = false

    // Open print receipt with full detailed metadata
    const updated = res.data || selectedInvoice.value
    invoiceToPrint.value = {
      ...updated,
      promo_code: currentPromoCode || updated.promo_code,
      promo_name: currentPromoName || updated.promo_name,
      diskon: (updated.diskon || 0) > 0 ? updated.diskon : currentPromoDiscount,
      subtotal: updated.subtotal || selectedInvoice.value.subtotal || (selectedInvoice.value.total_tagihan + currentPromoDiscount),
      total_tagihan: finalTotalTagihan.value,
      payments: [
        {
          id: 'p-latest',
          nomor_transaksi: 'TRX-' + Date.now(),
          invoice_id: selectedInvoice.value.id,
          kasir_id: authStore.user?.id || '',
          kasir_name: authStore.user?.full_name || 'Kasir SIMRS',
          nominal: finalTotalTagihan.value,
          metode_bayar: currentMethod as any,
          referensi_pembayaran: paymentForm.value.referensi_pembayaran,
          catatan: paymentForm.value.catatan,
          uang_diterima: currentTendered,
          kembalian: currentChange,
          created_at: new Date().toISOString()
        }
      ]
    }
    showReceiptModal.value = true

    await fetchActiveShift()
    await fetchInvoices()
    await fetchVisitsWaiting()
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal memproses pembayaran', 'error')
  } finally {
    submittingPayment.value = false
  }
}

const openReceiptModal = async (inv: Invoice) => {
  try {
    const res = await api.get<Invoice>(`/billing/invoices/${inv.id}`)
    invoiceToPrint.value = res.data || inv
  } catch (err) {
    invoiceToPrint.value = inv
  }
  showReceiptModal.value = true
}

const printDirect = () => {
  window.print()
}

// Receipt helpers for promo and change
const getReceiptPromoCode = (inv?: Invoice | null): string => {
  if (!inv) return ''
  if (inv.promo_code) return inv.promo_code
  if (inv.catatan && inv.catatan.includes('Promo:')) {
    const match = inv.catatan.match(/Promo:\s*([^\s(]+)/i)
    if (match) return match[1]
  }
  return ''
}

const getReceiptPromoDiscount = (inv?: Invoice | null): number => {
  if (!inv) return 0
  if (inv.diskon && inv.diskon > 0) return inv.diskon
  if (inv.catatan && inv.catatan.includes('Diskon Rp')) {
    const match = inv.catatan.match(/Diskon Rp\s*([\d.]+)/i)
    if (match) {
      const cleanNum = Number(match[1].replace(/\./g, ''))
      if (!isNaN(cleanNum)) return cleanNum
    }
  }
  return 0
}

const getReceiptPayment = (inv?: Invoice | null): BillingPayment | null => {
  if (!inv || !inv.payments || inv.payments.length === 0) return null
  return inv.payments[inv.payments.length - 1]
}

const getReceiptPaymentMethod = (inv?: Invoice | null): string => {
  const p = getReceiptPayment(inv)
  if (!p) return 'TUNAI'
  if (p.metode_bayar === 'TUNAI') return 'Uang Tunai (Cash)'
  if (p.metode_bayar === 'QRIS') return 'QRIS'
  if (p.metode_bayar === 'DEBIT') return 'Kartu Debit (EDC)'
  if (p.metode_bayar === 'KREDIT') return 'Kartu Kredit (EDC)'
  if (p.metode_bayar === 'TRANSFER') return 'Transfer Bank'
  return p.metode_bayar
}

const getReceiptTendered = (inv?: Invoice | null): number => {
  const p = getReceiptPayment(inv)
  if (!p) return inv?.total_terbayar || inv?.total_tagihan || 0
  return p.uang_diterima && p.uang_diterima > 0 ? p.uang_diterima : (p.nominal || inv?.total_tagihan || 0)
}

const getReceiptChange = (inv?: Invoice | null): number => {
  const p = getReceiptPayment(inv)
  if (!p) return 0
  return p.kembalian || 0
}

const formatCurrency = (val?: number) => {
  if (val === undefined || val === null) return 'Rp 0'
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(val)
}

const formatDate = (isoString?: string | null) => {
  if (!isoString) return '-'
  return new Date(isoString).toLocaleString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

// Terbilang Rupiah Helper
const terbilang = (n?: number): string => {
  if (!n) return 'Nol Rupiah'
  const angka = ['', 'Satu', 'Dua', 'Tiga', 'Empat', 'Lima', 'Enam', 'Tujuh', 'Delapan', 'Sembilan', 'Sepuluh', 'Sebelas']
  const konversi = (num: number): string => {
    if (num < 12) return angka[num]
    if (num < 20) return konversi(num - 10) + ' Belas'
    if (num < 100) return konversi(Math.floor(num / 10)) + ' Puluh ' + konversi(num % 10)
    if (num < 200) return 'Seratus ' + konversi(num - 100)
    if (num < 1000) return konversi(Math.floor(num / 100)) + ' Ratus ' + konversi(num % 100)
    if (num < 2000) return 'Seribu ' + konversi(num - 1000)
    if (num < 1000000) return konversi(Math.floor(num / 1000)) + ' Ribu ' + konversi(num % 1000)
    if (num < 1000000000) return konversi(Math.floor(num / 1000000)) + ' Juta ' + konversi(num % 1000000)
    return ''
  }
  return konversi(Math.floor(n)).trim() + ' Rupiah'
}

onMounted(async () => {
  await fetchActiveShift()
  await fetchVisitsWaiting()
  await fetchInvoices()
  await fetchPromos()
})
</script>

<template>
  <div class="space-y-4">
    <!-- Header Page -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-white p-4 rounded-xl border border-slate-200 shadow-sm">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-lg bg-sky-50 text-sky-600 flex items-center justify-center border border-sky-100">
          <Receipt class="w-5 h-5" />
        </div>
        <div>
          <h1 class="text-xl font-bold text-slate-800">Kasir & Manajemen Tagihan (Billing)</h1>
          <p class="text-xs text-slate-500">Penyelesaian Pembayaran Pasien, Rincian Jasa Pelayanan, & Kuitansi Resmi</p>
        </div>
      </div>

      <!-- Shift Status Action -->
      <div>
        <BaseButton
          v-if="!activeShift"
          variant="primary"
          size="sm"
          @click="showOpenShiftModal = true"
        >
          <Lock class="w-4 h-4 mr-1.5" />
          Buka Shift Kasir Baru
        </BaseButton>
        <BaseButton
          v-else
          variant="secondary"
          size="sm"
          @click="openCloseShiftDialog"
        >
          <CheckCircle2 class="w-4 h-4 mr-1.5 text-emerald-600" />
          Tutup Shift & Rekonsiliasi
        </BaseButton>
      </div>
    </div>

    <!-- Active Shift Status Card -->
    <div
      v-if="activeShift"
      class="bg-gradient-to-r from-sky-900 to-sky-700 text-white rounded-xl p-4 shadow-sm"
    >
      <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div class="space-y-1">
          <div class="flex items-center gap-2">
            <span class="px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-500/30 text-emerald-300 border border-emerald-400/40">
              SHIFT AKTIF
            </span>
            <span class="font-bold text-sm text-sky-100">{{ activeShift.shift_name }}</span>
          </div>
          <div class="text-xs text-sky-200 flex items-center gap-4">
            <span>Petugas: <strong>{{ activeShift.kasir_name || authStore.user?.full_name }}</strong></span>
            <span>Waktu Buka: <strong>{{ formatDate(activeShift.waktu_buka) }}</strong></span>
          </div>
        </div>

        <div class="grid grid-cols-3 gap-3 bg-white/10 backdrop-blur-sm p-3 rounded-lg border border-white/10 text-xs">
          <div>
            <div class="text-sky-200 text-[11px]">Modal Saldo Awal</div>
            <div class="font-bold text-sm text-white tabular-nums">{{ formatCurrency(activeShift.saldo_awal) }}</div>
          </div>
          <div>
            <div class="text-sky-200 text-[11px]">Total Tunai Masuk</div>
            <div class="font-bold text-sm text-emerald-300 tabular-nums">{{ formatCurrency(activeShift.total_tunai_sistem) }}</div>
          </div>
          <div>
            <div class="text-sky-200 text-[11px]">Total Non-Tunai</div>
            <div class="font-bold text-sm text-amber-300 tabular-nums">{{ formatCurrency(activeShift.total_non_tunai_sistem) }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- If No Active Shift Warning -->
    <div
      v-else
      class="bg-amber-50 border border-amber-200 rounded-xl p-4 flex items-center justify-between gap-4 text-amber-800"
    >
      <div class="flex items-center gap-3">
        <AlertCircle class="w-6 h-6 text-amber-600 shrink-0" />
        <div>
          <h3 class="font-bold text-sm">Shift Kasir Belum Dibuka</h3>
          <p class="text-xs text-amber-700">Anda harus membuka shift kasir dengan saldo awal untuk dapat memproses transaksi pembayaran pasien.</p>
        </div>
      </div>
      <BaseButton variant="primary" size="sm" @click="showOpenShiftModal = true">
        Buka Shift Sekarang
      </BaseButton>
    </div>

    <!-- Navigation Tabs -->
    <div class="flex items-center border-b border-slate-200 gap-2 bg-white px-4 pt-2 rounded-t-xl">
      <button
        class="flex items-center gap-2 px-4 py-2.5 text-xs font-semibold border-b-2 transition-colors"
        :class="activeTab === 'antrean' ? 'border-sky-600 text-sky-600' : 'border-transparent text-slate-500 hover:text-slate-700'"
        @click="activeTab = 'antrean'"
      >
        <Receipt class="w-4 h-4" />
        Antrean Pelayanan & Tagihan Kasir
        <span class="text-[10px] px-1.5 py-0.5 rounded-full bg-slate-100 text-slate-600 tabular-nums">{{ visitsWaiting.length }} antrean</span>
      </button>

      <button
        class="flex items-center gap-2 px-4 py-2.5 text-xs font-semibold border-b-2 transition-colors"
        :class="activeTab === 'riwayat_shift' ? 'border-sky-600 text-sky-600' : 'border-transparent text-slate-500 hover:text-slate-700'"
        @click="() => { activeTab = 'riwayat_shift'; fetchShiftHistory() }"
      >
        <History class="w-4 h-4" />
        Riwayat & Rekonsiliasi Shift Kasir
      </button>
    </div>

    <!-- TAB 1: ANTREAN & TAGIHAN KASIR -->
    <div v-if="activeTab === 'antrean'" class="space-y-4">
      <!-- Section A: Pasien Menunggu Tagihan (Status KASIR) -->
      <div v-if="visitsWaiting.length > 0" class="bg-white rounded-xl border border-sky-200 shadow-sm overflow-hidden">
        <div class="px-4 py-3 bg-sky-50/70 border-b border-sky-100 flex items-center justify-between">
          <div class="flex items-center gap-2">
            <Clock class="w-4 h-4 text-sky-600" />
            <span class="text-xs font-bold text-sky-900">Pasien Menunggu Penyelesaian Tagihan di Kasir</span>
            <span class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-sky-600 text-white tabular-nums">{{ visitsWaiting.length }}</span>
          </div>
          <BaseButton variant="ghost" size="xs" @click="fetchVisitsWaiting">
            <RefreshCw class="w-3.5 h-3.5 mr-1" />
            Perbarui Antrean
          </BaseButton>
        </div>

        <div class="divide-y divide-slate-100">
          <div
            v-for="v in visitsWaiting"
            :key="v.id"
            class="p-3 hover:bg-slate-50 flex items-center justify-between gap-4 transition-colors text-xs"
          >
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 rounded-full bg-sky-100 text-sky-700 flex items-center justify-center font-bold">
                {{ v.no_antrian }}
              </div>
              <div>
                <div class="font-bold text-slate-800">{{ v.patient_name }}</div>
                <div class="text-[11px] text-slate-500 flex items-center gap-2">
                  <span class="font-mono">RM: {{ v.no_rm }}</span>
                  <span>•</span>
                  <span>Poli: {{ v.poli_name }}</span>
                  <span>•</span>
                  <span>Penjamin: {{ v.guarantor_name || 'Umum Mandiri' }}</span>
                </div>
              </div>
            </div>

            <div class="flex items-center gap-2">
              <BaseButton
                variant="primary"
                size="xs"
                @click="generateBilling(v)"
              >
                <Calculator class="w-3.5 h-3.5 mr-1" />
                Hitung & Bayar Tagihan
              </BaseButton>
            </div>
          </div>
        </div>
      </div>

      <!-- Section B: Daftar Seluruh Tagihan / Invoice Kasir -->
      <div class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
        <!-- Filter & Search Bar -->
        <div class="p-3 border-b border-slate-200 flex flex-col md:flex-row items-center gap-3 bg-slate-50/50">
          <div class="w-full md:w-1/3 relative">
            <Search class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Cari No. Invoice, RM, atau Nama Pasien..."
              class="w-full pl-9 pr-3 py-1.5 text-xs border border-slate-200 rounded-lg focus:outline-none focus:border-sky-500"
              @keyup.enter="fetchInvoices"
            />
          </div>

          <div class="w-full md:w-1/4">
            <BaseSelect
              v-model="selectedInvoiceStatus"
              :options="[
                { value: 'SEMUA', label: 'Semua Status Invoice' },
                { value: 'BELUM_BAYAR', label: 'Belum Bayar' },
                { value: 'LUNAS', label: 'Lunas' },
              ]"
              compact
              @change="fetchInvoices"
            />
          </div>

          <div class="ml-auto flex items-center gap-2">
            <BaseButton variant="secondary" size="xs" @click="fetchInvoices">
              <RefreshCw class="w-3.5 h-3.5 mr-1" :class="{ 'animate-spin': loading }" />
              Segarkan
            </BaseButton>
          </div>
        </div>

        <!-- Invoices Table -->
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs text-slate-600">
            <thead class="bg-slate-50 text-slate-700 font-semibold border-b border-slate-200 uppercase tracking-wider">
              <tr>
                <th class="py-2.5 px-3">No. Invoice</th>
                <th class="py-2.5 px-3">Pasien & No. RM</th>
                <th class="py-2.5 px-3">Penjamin</th>
                <th class="py-2.5 px-3 text-right">Subtotal</th>
                <th class="py-2.5 px-3 text-right">Diskon</th>
                <th class="py-2.5 px-3 text-right">Total Tagihan</th>
                <th class="py-2.5 px-3 text-center">Status</th>
                <th class="py-2.5 px-3 text-center">Aksi</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-if="loading">
                <td colspan="8" class="text-center py-10 text-slate-400">
                  <RefreshCw class="w-6 h-6 animate-spin mx-auto mb-2 text-sky-500" />
                  Memuat data tagihan kasir...
                </td>
              </tr>
              <tr v-else-if="invoices.length === 0">
                <td colspan="8" class="text-center py-12 text-slate-400">
                  <Receipt class="w-10 h-10 mx-auto mb-2 text-slate-300" />
                  Belum ada tagihan kasir yang tercatat.
                </td>
              </tr>
              <tr
                v-for="inv in invoices"
                :key="inv.id"
                class="hover:bg-slate-50 transition-colors"
              >
                <td class="py-2.5 px-3">
                  <div class="font-mono font-bold text-sky-800">{{ inv.nomor_invoice }}</div>
                  <div class="text-[10px] text-slate-400">{{ formatDate(inv.created_at) }}</div>
                </td>

                <td class="py-2.5 px-3">
                  <div class="font-semibold text-slate-800">{{ inv.patient_name || '-' }}</div>
                  <div class="text-[10px] text-slate-500 font-mono">RM: {{ inv.no_rm || '-' }}</div>
                </td>

                <td class="py-2.5 px-3">
                  <span class="px-2 py-0.5 rounded text-[10px] font-semibold bg-slate-100 text-slate-700">
                    {{ inv.jenis_penjamin || 'Umum' }}
                  </span>
                </td>

                <td class="py-2.5 px-3 text-right tabular-nums">
                  {{ formatCurrency(inv.subtotal) }}
                </td>

                <td class="py-2.5 px-3 text-right tabular-nums text-rose-600">
                  {{ inv.diskon > 0 ? '-' + formatCurrency(inv.diskon) : 'Rp 0' }}
                </td>

                <td class="py-2.5 px-3 text-right font-bold text-slate-800 tabular-nums">
                  {{ formatCurrency(inv.total_tagihan) }}
                </td>

                <td class="py-2.5 px-3 text-center">
                  <StatusBadge :status="inv.status" />
                </td>

                <td class="py-2.5 px-3 text-center">
                  <div class="flex items-center justify-center gap-1">
                    <BaseButton
                      v-if="inv.status === 'BELUM_BAYAR'"
                      variant="primary"
                      size="xs"
                      @click="openPaymentModal(inv)"
                    >
                      <DollarSign class="w-3.5 h-3.5 mr-1" />
                      Bayar
                    </BaseButton>

                    <BaseButton
                      v-else
                      variant="secondary"
                      size="xs"
                      @click="openReceiptModal(inv)"
                    >
                      <Printer class="w-3.5 h-3.5 mr-1 text-slate-600" />
                      Kuitansi
                    </BaseButton>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- TAB 2: RIWAYAT SHIFT KASIR -->
    <div v-else-if="activeTab === 'riwayat_shift'" class="space-y-3">
      <div class="bg-white p-3 rounded-xl border border-slate-200 shadow-sm flex items-center justify-between">
        <div class="text-xs font-semibold text-slate-700 flex items-center gap-2">
          <History class="w-4 h-4 text-sky-600" />
          Riwayat Rekonsiliasi Shift Kasir
        </div>
        <BaseButton variant="secondary" size="xs" @click="fetchShiftHistory">
          <RefreshCw class="w-3.5 h-3.5 mr-1" />
          Segarkan
        </BaseButton>
      </div>

      <div class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs text-slate-600">
            <thead class="bg-slate-50 text-slate-700 font-semibold border-b border-slate-200 uppercase tracking-wider">
              <tr>
                <th class="py-2.5 px-3">Nama Shift</th>
                <th class="py-2.5 px-3">Kasir</th>
                <th class="py-2.5 px-3">Waktu Buka / Tutup</th>
                <th class="py-2.5 px-3 text-right">Saldo Awal</th>
                <th class="py-2.5 px-3 text-right">Tunai Masuk</th>
                <th class="py-2.5 px-3 text-right">Non-Tunai Masuk</th>
                <th class="py-2.5 px-3 text-right">Fisik Akhir</th>
                <th class="py-2.5 px-3 text-right">Selisih</th>
                <th class="py-2.5 px-3 text-center">Status</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-if="shiftHistory.length === 0">
                <td colspan="9" class="text-center py-10 text-slate-400">
                  Belum ada riwayat penutupan shift kasir.
                </td>
              </tr>
              <tr
                v-for="s in shiftHistory"
                :key="s.id"
                class="hover:bg-slate-50 transition-colors"
              >
                <td class="py-2.5 px-3 font-semibold text-slate-800">
                  {{ s.shift_name }}
                </td>

                <td class="py-2.5 px-3 text-slate-700">
                  {{ s.kasir_name || '-' }}
                </td>

                <td class="py-2.5 px-3 text-[11px] text-slate-500">
                  <div>Buka: {{ formatDate(s.waktu_buka) }}</div>
                  <div>Tutup: {{ formatDate(s.waktu_tutup) }}</div>
                </td>

                <td class="py-2.5 px-3 text-right tabular-nums">
                  {{ formatCurrency(s.saldo_awal) }}
                </td>

                <td class="py-2.5 px-3 text-right font-medium text-emerald-700 tabular-nums">
                  {{ formatCurrency(s.total_tunai_sistem) }}
                </td>

                <td class="py-2.5 px-3 text-right font-medium text-amber-700 tabular-nums">
                  {{ formatCurrency(s.total_non_tunai_sistem) }}
                </td>

                <td class="py-2.5 px-3 text-right font-bold text-slate-800 tabular-nums">
                  {{ formatCurrency(s.saldo_akhir_fisik || 0) }}
                </td>

                <td class="py-2.5 px-3 text-right font-bold tabular-nums"
                  :class="(s.selisih || 0) === 0 ? 'text-emerald-600' : 'text-rose-600'"
                >
                  {{ (s.selisih || 0) > 0 ? '+' : '' }}{{ formatCurrency(s.selisih || 0) }}
                </td>

                <td class="py-2.5 px-3 text-center">
                  <span
                    class="px-2 py-0.5 rounded text-[10px] font-bold"
                    :class="s.status === 'OPEN' ? 'bg-emerald-100 text-emerald-800' : 'bg-slate-100 text-slate-700'"
                  >
                    {{ s.status }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- MODAL: Buka Shift Kasir -->
    <BaseModal
      v-model="showOpenShiftModal"
      title="Buka Shift Kasir Baru"
      width="md"
    >
      <div class="space-y-3">
        <BaseSelect
          v-model="openShiftForm.shift_name"
          label="Nama Sesi Shift"
          :options="[
            { value: 'Shift Pagi (07:00 - 14:00)', label: 'Shift Pagi (07:00 - 14:00)' },
            { value: 'Shift Siang (14:00 - 21:00)', label: 'Shift Siang (14:00 - 21:00)' },
            { value: 'Shift Malam (21:00 - 07:00)', label: 'Shift Malam (21:00 - 07:00)' },
          ]"
          required
        />

        <BaseInput
          v-model.number="openShiftForm.saldo_awal"
          type="number"
          label="Saldo Awal Kas / Modal Kembalian (Rp)"
          placeholder="200000"
          required
          hint="Uang tunai kas kecil yang diserahkan untuk modal kembalian awal"
        />

        <div class="flex items-center justify-end gap-2 pt-3 border-t border-slate-100">
          <BaseButton variant="ghost" size="sm" @click="showOpenShiftModal = false">
            Batal
          </BaseButton>
          <BaseButton
            variant="primary"
            size="sm"
            :loading="submittingShift"
            @click="openShift"
          >
            <Lock class="w-4 h-4 mr-1.5" />
            Mulai Shift Kasir
          </BaseButton>
        </div>
      </div>
    </BaseModal>

    <!-- MODAL: Tutup Shift Kasir -->
    <BaseModal
      v-model="showCloseShiftModal"
      title="Tutup Shift & Rekonsiliasi Kasir"
      width="md"
    >
      <div v-if="activeShift" class="space-y-3">
        <div class="bg-slate-50 p-3 rounded-lg border border-slate-200 text-xs space-y-1.5">
          <div class="flex justify-between">
            <span class="text-slate-500">Saldo Awal Modal:</span>
            <span class="font-bold tabular-nums">{{ formatCurrency(activeShift.saldo_awal) }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-slate-500">Total Tunai Sistem:</span>
            <span class="font-bold text-emerald-700 tabular-nums">+{{ formatCurrency(activeShift.total_tunai_sistem) }}</span>
          </div>
          <div class="flex justify-between pt-1 border-t border-slate-200 font-semibold">
            <span>Uang Fisik Kas Seharusnya:</span>
            <span class="font-bold text-sky-800 tabular-nums">
              {{ formatCurrency(activeShift.saldo_awal + activeShift.total_tunai_sistem) }}
            </span>
          </div>
        </div>

        <BaseInput
          v-model.number="closeShiftForm.saldo_akhir_fisik"
          type="number"
          label="Hitungan Uang Tunai Fisik di Laci Kasir (Rp)"
          required
        />

        <!-- Selisih preview -->
        <div
          class="p-2.5 rounded-lg border text-xs flex justify-between items-center"
          :class="calculatedSelisih === 0 ? 'bg-emerald-50 border-emerald-200 text-emerald-800' : 'bg-rose-50 border-rose-200 text-rose-800'"
        >
          <span>Selisih (Fisik - Sistem):</span>
          <span class="font-bold font-mono text-sm tabular-nums">
            {{ calculatedSelisih > 0 ? '+' : '' }}{{ formatCurrency(calculatedSelisih) }}
          </span>
        </div>

        <BaseInput
          v-model="closeShiftForm.catatan_shift"
          label="Catatan Rekonsiliasi (Bila Ada Selisih)"
          placeholder="Catatan kendala atau penyerahan kas ke supervisor..."
        />

        <div class="flex items-center justify-end gap-2 pt-3 border-t border-slate-100">
          <BaseButton variant="ghost" size="sm" @click="showCloseShiftModal = false">
            Batal
          </BaseButton>
          <BaseButton
            variant="danger"
            size="sm"
            :loading="submittingShift"
            @click="closeShift"
          >
            <CheckCircle2 class="w-4 h-4 mr-1.5" />
            Tutup Shift & Serahkan Kas
          </BaseButton>
        </div>
      </div>
    </BaseModal>

    <!-- MODAL: Proses Pembayaran Pasien -->
    <BaseModal
      v-model="showPaymentModal"
      title="Penyelesaian Pembayaran Tagihan"
      width="lg"
    >
      <div v-if="selectedInvoice" class="space-y-4">
        <!-- PENDING ORDER WARNING BANNER (PAYMENT LOCK) -->
        <div v-if="pendingBlockReason" class="p-3.5 rounded-lg bg-rose-50 border border-rose-300 text-rose-800 text-xs flex items-start gap-2.5">
          <Lock class="w-5 h-5 text-rose-600 shrink-0 mt-0.5" />
          <div class="space-y-0.5">
            <strong class="font-bold text-[13px] text-rose-900 block">PEMBAYARAN DIKUNCI: Ada Orderan yang Belum Selesai</strong>
            <p class="leading-relaxed">{{ pendingBlockReason }}</p>
            <p class="text-[11px] text-rose-700 font-medium pt-1">
              * Sesuai standar operasional klinis, kasir tidak dapat menerima pelunasan tagihan sebelum seluruh tindakan dan obat telah selesai divalidasi.
            </p>
          </div>
        </div>

        <!-- Patient Context & Invoice Metadata -->
        <div class="bg-sky-50 border border-sky-100 p-3 rounded-lg text-xs space-y-1">
          <div class="flex justify-between items-center">
            <span class="font-bold text-sky-900 text-sm">Invoice: {{ selectedInvoice.nomor_invoice }}</span>
            <span class="px-2 py-0.5 rounded text-[10px] font-bold bg-sky-200/80 text-sky-800">
              {{ selectedInvoice.jenis_penjamin || 'Umum' }}
            </span>
          </div>
          <div class="text-slate-600 flex flex-wrap gap-x-4">
            <span>Pasien: <strong>{{ selectedInvoice.patient_name }}</strong></span>
            <span>No. RM: <strong>{{ selectedInvoice.no_rm }}</strong></span>
          </div>
        </div>

        <!-- Rincian Item Tagihan -->
        <div class="border border-slate-200 rounded-lg overflow-hidden text-xs">
          <table class="w-full text-left">
            <thead class="bg-slate-50 text-slate-700 font-semibold border-b border-slate-200">
              <tr>
                <th class="py-2 px-3">Komponen Pelayanan</th>
                <th class="py-2 px-3 text-center">Qty</th>
                <th class="py-2 px-3 text-right">Harga Satuan</th>
                <th class="py-2 px-3 text-right">Subtotal</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-for="item in (selectedInvoice.items || [])" :key="item.id">
                <td class="py-2 px-3">
                  <div class="font-medium text-slate-800">{{ item.item_name }}</div>
                  <div class="text-[10px] text-slate-400">{{ item.item_type }}</div>
                </td>
                <td class="py-2 px-3 text-center tabular-nums">{{ item.qty }}</td>
                <td class="py-2 px-3 text-right tabular-nums">{{ formatCurrency(item.unit_price) }}</td>
                <td class="py-2 px-3 text-right font-semibold text-slate-800 tabular-nums">{{ formatCurrency(item.subtotal) }}</td>
              </tr>
            </tbody>
          </table>

          <div class="bg-slate-50 p-3 border-t border-slate-200 space-y-1 text-xs">
            <div class="flex justify-between text-slate-600">
              <span>Subtotal Pelayanan:</span>
              <span class="tabular-nums">{{ formatCurrency(selectedInvoice.subtotal) }}</span>
            </div>
            <div v-if="selectedInvoice.diskon > 0 || promoDiscount > 0" class="flex justify-between text-rose-600 font-medium">
              <span>Potongan Diskon / Promo:</span>
              <span class="tabular-nums">-{{ formatCurrency((selectedInvoice.diskon || 0) + promoDiscount) }}</span>
            </div>
            <div class="flex justify-between font-bold text-sm text-slate-800 pt-1 border-t border-slate-200">
              <span>Total Tagihan Akhir:</span>
              <span class="text-sky-700 tabular-nums">{{ formatCurrency(finalTotalTagihan) }}</span>
            </div>
          </div>
        </div>

        <!-- Master Promo Selector -->
        <div class="p-3 bg-white border border-slate-200 rounded-lg space-y-2 text-xs">
          <div class="flex items-center justify-between">
            <span class="font-bold text-slate-700">Master Promo & Kupon Potongan</span>
            <span v-if="promoDiscount > 0" class="font-bold text-emerald-600 font-mono">
              Hemat {{ formatCurrency(promoDiscount) }}
            </span>
          </div>
          <BaseSelect
            v-model="selectedPromoCode"
            :options="[
              { value: '', label: '-- Tanpa Kode Promo --' },
              ...promos.filter(p => p.status_aktif).map(p => ({
                value: p.code,
                label: `${p.code} - ${p.name} (Diskon ${p.discount_type === 'PERCENTAGE' ? p.discount_value + '%' : 'Rp ' + p.discount_value.toLocaleString('id-ID')})`,
                badge: p.code
              }))
            ]"
            placeholder="Pilih Kupon Promo..."
            compact
            @change="paymentForm.nominal = finalTotalTagihan"
          />
        </div>

        <!-- Toggle Split Payment (Fase 1) -->
        <div class="flex items-center justify-between p-3 rounded-lg bg-sky-50 border border-sky-200">
          <div class="flex items-center gap-2">
            <CreditCard class="w-4 h-4 text-sky-700" />
            <div>
              <span class="text-xs font-bold text-sky-900 block">Metode Pembayaran Campuran (Split Payment)</span>
              <span class="text-[11px] text-sky-700">Kombinasikan Tunai + Non-Tunai / EDC / QRIS dalam 1 tagihan</span>
            </div>
          </div>
          <label class="relative inline-flex items-center cursor-pointer">
            <input type="checkbox" v-model="isSplitPayment" class="sr-only peer" />
            <div class="w-9 h-5 bg-slate-300 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-slate-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-sky-600"></div>
          </label>
        </div>

        <!-- SPLIT PAYMENT FORM -->
        <div v-if="isSplitPayment" class="space-y-3 bg-slate-50/70 p-3.5 rounded-lg border border-slate-200">
          <div class="flex items-center justify-between">
            <span class="text-xs font-bold text-slate-800">Daftar Pembayaran Campuran</span>
            <BaseButton variant="outline" size="xs" @click="addSplitLine">
              <Plus class="w-3.5 h-3.5 mr-1" />
              Tambah Jalur Bayar
            </BaseButton>
          </div>

          <div class="space-y-2">
            <div
              v-for="(sp, idx) in splitPayments"
              :key="idx"
              class="p-2.5 bg-white border border-slate-200 rounded-lg grid grid-cols-12 gap-2 items-center text-xs"
            >
              <div class="col-span-3">
                <label class="block text-[10px] font-semibold text-slate-500 mb-0.5">Metode</label>
                <select
                  v-model="sp.metode_bayar"
                  class="w-full p-1.5 border border-slate-200 rounded text-xs bg-white font-medium"
                >
                  <option value="TUNAI">Uang Tunai</option>
                  <option value="QRIS">QRIS</option>
                  <option value="DEBIT">Kartu Debit</option>
                  <option value="KREDIT">Kartu Kredit</option>
                  <option value="TRANSFER">Transfer Bank</option>
                </select>
              </div>

              <div class="col-span-4">
                <label class="block text-[10px] font-semibold text-slate-500 mb-0.5">Nominal (Rp)</label>
                <input
                  v-model.number="sp.nominal"
                  type="number"
                  min="0"
                  class="w-full p-1.5 border border-slate-200 rounded text-xs font-mono font-bold text-sky-800"
                />
              </div>

              <div class="col-span-4">
                <label class="block text-[10px] font-semibold text-slate-500 mb-0.5">No. Ref / Approval</label>
                <input
                  v-model="sp.referensi_pembayaran"
                  type="text"
                  placeholder="Opsional / No. Ref"
                  class="w-full p-1.5 border border-slate-200 rounded text-[11px]"
                />
              </div>

              <div class="col-span-1 text-center pt-3">
                <button
                  v-if="splitPayments.length > 1"
                  class="text-rose-500 hover:text-rose-700 p-1"
                  title="Hapus baris pembayaran"
                  @click="removeSplitLine(idx)"
                >
                  <Trash2 class="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>

          <!-- Summary Split -->
          <div class="p-3 bg-white rounded border border-slate-200 space-y-1.5 text-xs">
            <div class="flex justify-between">
              <span class="text-slate-600">Total Tagihan:</span>
              <strong class="font-mono text-slate-800">{{ formatCurrency(finalTotalTagihan) }}</strong>
            </div>
            <div class="flex justify-between">
              <span class="text-slate-600">Total Bayar Split:</span>
              <strong class="font-mono text-sky-700">{{ formatCurrency(totalSplitPaid) }}</strong>
            </div>
            <div class="flex justify-between pt-1 border-t font-bold">
              <span>Status Saldo Pembayaran:</span>
              <span
                v-if="splitDifference >= 0"
                class="font-mono text-emerald-700"
              >
                LUNAS {{ splitDifference > 0 ? `(Lebih: ${formatCurrency(splitDifference)})` : '' }}
              </span>
              <span
                v-else
                class="font-mono text-rose-600"
              >
                Kurang: {{ formatCurrency(Math.abs(splitDifference)) }}
              </span>
            </div>
          </div>
        </div>

        <!-- SINGLE PAYMENT FORM -->
        <div v-else class="space-y-3 bg-slate-50/70 p-3.5 rounded-lg border border-slate-200">
          <div class="grid grid-cols-2 gap-3">
            <BaseSelect
              v-model="paymentForm.metode_bayar"
              label="Metode Pembayaran"
              :options="[
                { value: 'TUNAI', label: 'Uang Tunai (Cash)' },
                { value: 'QRIS', label: 'QRIS Dinamis / Statis' },
                { value: 'DEBIT', label: 'Kartu Debit (EDC)' },
                { value: 'KREDIT', label: 'Kartu Kredit (EDC)' },
                { value: 'TRANSFER', label: 'Transfer Bank' },
              ]"
              required
            />

            <BaseInput
              v-model.number="paymentForm.nominal"
              type="number"
              label="Nominal Dibayarkan (Rp)"
              required
            />
          </div>

          <!-- Kembalian if TUNAI -->
          <div
            v-if="paymentForm.metode_bayar === 'TUNAI'"
            class="p-2.5 rounded-lg bg-emerald-50 border border-emerald-200 text-emerald-800 text-xs flex justify-between items-center"
          >
            <span class="font-medium">Uang Kembalian Pasien:</span>
            <span class="text-base font-bold font-mono tabular-nums">{{ formatCurrency(kembalian) }}</span>
          </div>

          <!-- Referensi if Non Tunai -->
          <BaseInput
            v-if="paymentForm.metode_bayar !== 'TUNAI'"
            v-model="paymentForm.referensi_pembayaran"
            label="Nomor Referensi / Approval Code EDC / ID Transaksi"
            placeholder="Contoh: REF-92817263 / Approval 1827"
            required
          />

          <BaseInput
            v-model="paymentForm.catatan"
            label="Catatan Pembayaran"
            placeholder="Catatan tambahan bila ada..."
          />
        </div>

        <div class="flex items-center justify-end gap-2 pt-2 border-t border-slate-100">
          <BaseButton variant="ghost" size="sm" @click="showPaymentModal = false">
            Batal
          </BaseButton>
          <BaseButton
            variant="primary"
            size="sm"
            :disabled="!!pendingBlockReason"
            :loading="submittingPayment"
            :title="pendingBlockReason || undefined"
            @click="processPayment"
          >
            <CheckCircle2 class="w-4 h-4 mr-1.5" />
            Konfirmasi Bayar & Cetak Kuitansi
          </BaseButton>
        </div>
      </div>
    </BaseModal>

    <!-- MODAL: Cetak Kuitansi Resmi SIMRS -->
    <BaseModal
      v-model="showReceiptModal"
      title="Kuitansi Pembayaran Pelayanan SIMRS"
      width="lg"
    >
      <div v-if="invoiceToPrint" class="space-y-4">
        <!-- Printable Receipt -->
        <div id="receipt-area" class="p-6 bg-white border border-slate-300 rounded-lg shadow-sm text-slate-800 font-sans">
          <!-- Kop Surat -->
          <div class="border-b-2 border-slate-800 pb-3 mb-4 text-center">
            <h2 class="text-base font-bold uppercase tracking-wider text-slate-900">RUMAH SAKIT & KLINIK SIMRS-ASOY</h2>
            <p class="text-xs text-slate-600">Jl. Sehat Bersama No. 12, Jakarta Selatan | Telp: (021) 7890-1234</p>
            <p class="text-[10px] text-slate-500">Bukti Pembayaran Kasir Pelayanan Sah & Diakui</p>
          </div>

          <div class="flex justify-between items-center mb-4">
            <div>
              <span class="text-xs text-slate-500">Nomor Kuitansi:</span>
              <div class="font-mono font-bold text-sm text-slate-900">{{ invoiceToPrint.nomor_invoice }}</div>
            </div>
            <div class="text-right">
              <span class="text-xs text-slate-500">Tanggal:</span>
              <div class="text-xs font-semibold text-slate-800">{{ formatDate(invoiceToPrint.created_at) }}</div>
            </div>
          </div>

          <!-- Metadata Pasien -->
          <div class="bg-slate-50 p-3 rounded border border-slate-200 text-xs grid grid-cols-2 gap-2 mb-4">
            <div>
              <div class="grid grid-cols-3">
                <span class="text-slate-500">Nama Pasien</span>
                <span class="col-span-2 font-bold">: {{ invoiceToPrint.patient_name }}</span>
              </div>
              <div class="grid grid-cols-3">
                <span class="text-slate-500">No. Rekam Medis</span>
                <span class="col-span-2 font-mono font-bold">: {{ invoiceToPrint.no_rm }}</span>
              </div>
            </div>
            <div>
              <div class="grid grid-cols-3">
                <span class="text-slate-500">Penjamin</span>
                <span class="col-span-2 font-bold">: {{ invoiceToPrint.jenis_penjamin || 'Umum' }}</span>
              </div>
              <div class="grid grid-cols-3">
                <span class="text-slate-500">Kasir Bertugas</span>
                <span class="col-span-2">: {{ getReceiptPayment(invoiceToPrint)?.kasir_name || authStore.user?.full_name || 'Kasir SIMRS' }}</span>
              </div>
            </div>
          </div>

          <!-- Rincian Pembayaran Table -->
          <table class="w-full text-xs text-left border-collapse mb-4">
            <thead>
              <tr class="border-b-2 border-slate-300 font-bold text-slate-700">
                <th class="py-1.5">No</th>
                <th class="py-1.5">Rincian Komponen Pelayanan</th>
                <th class="py-1.5 text-center">Banyaknya</th>
                <th class="py-1.5 text-right">Tarif</th>
                <th class="py-1.5 text-right">Jumlah</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-200">
              <tr v-for="(it, idx) in (invoiceToPrint.items || [])" :key="it.id">
                <td class="py-1.5">{{ idx + 1 }}</td>
                <td class="py-1.5 font-medium">{{ it.item_name }}</td>
                <td class="py-1.5 text-center tabular-nums">{{ it.qty }}</td>
                <td class="py-1.5 text-right tabular-nums">{{ formatCurrency(it.unit_price) }}</td>
                <td class="py-1.5 text-right tabular-nums">{{ formatCurrency(it.subtotal) }}</td>
              </tr>
            </tbody>
            <tfoot>
              <!-- Subtotal jika ada diskon atau promo -->
              <tr v-if="(invoiceToPrint.diskon && invoiceToPrint.diskon > 0) || getReceiptPromoDiscount(invoiceToPrint) > 0" class="border-t-2 border-slate-300 text-slate-600">
                <td colspan="4" class="py-1 text-right font-medium">Subtotal Pelayanan:</td>
                <td class="py-1 text-right tabular-nums font-mono">
                  {{ formatCurrency(invoiceToPrint.subtotal || ((invoiceToPrint.total_tagihan || 0) + getReceiptPromoDiscount(invoiceToPrint))) }}
                </td>
              </tr>

              <!-- Potongan Diskon Promo -->
              <tr v-if="(invoiceToPrint.diskon && invoiceToPrint.diskon > 0) || getReceiptPromoDiscount(invoiceToPrint) > 0" class="text-emerald-700 bg-emerald-50/70">
                <td colspan="4" class="py-1.5 text-right font-semibold">
                  <div class="inline-flex items-center gap-1.5">
                    <span class="px-1.5 py-0.5 rounded bg-emerald-100 text-emerald-800 text-[10px] font-bold tracking-wide border border-emerald-200">
                      PROMO: {{ invoiceToPrint.promo_code || getReceiptPromoCode(invoiceToPrint) || 'KUPON' }}
                    </span>
                    <span>Potongan Diskon Promo:</span>
                  </div>
                </td>
                <td class="py-1.5 text-right tabular-nums font-mono text-emerald-700 font-bold">
                  -{{ formatCurrency(invoiceToPrint.diskon || getReceiptPromoDiscount(invoiceToPrint)) }}
                </td>
              </tr>

              <!-- Total Pembayaran / Tagihan Akhir -->
              <tr class="border-t-2 border-slate-800 font-bold text-slate-900 text-[13px]">
                <td colspan="4" class="py-1.5 text-right">Total Pembayaran:</td>
                <td class="py-1.5 text-right tabular-nums font-mono text-[14px]">
                  {{ formatCurrency(invoiceToPrint.total_tagihan) }}
                </td>
              </tr>

              <!-- Rincian Pembayaran Kasir (Single atau Split) -->
              <template v-if="invoiceToPrint.payments && invoiceToPrint.payments.length > 1">
                <tr v-for="p in invoiceToPrint.payments" :key="p.id" class="border-t border-dashed border-slate-200 text-slate-700">
                  <td colspan="4" class="py-1 text-right">
                    <span>Metode: {{ p.metode_bayar }} {{ p.referensi_pembayaran ? `(${p.referensi_pembayaran})` : '' }}:</span>
                  </td>
                  <td class="py-1 text-right tabular-nums font-mono font-medium">
                    {{ formatCurrency(p.nominal) }}
                  </td>
                </tr>
              </template>
              <tr v-else class="border-t border-dashed border-slate-300 text-slate-700">
                <td colspan="4" class="py-1 text-right">
                  <span>Jumlah Diterima ({{ getReceiptPaymentMethod(invoiceToPrint) }}):</span>
                </td>
                <td class="py-1 text-right tabular-nums font-mono font-medium">
                  {{ formatCurrency(getReceiptTendered(invoiceToPrint)) }}
                </td>
              </tr>

              <!-- Uang Kembalian Pasien (jika ada) -->
              <tr v-if="getReceiptChange(invoiceToPrint) > 0" class="text-sky-900 bg-sky-50 font-bold">
                <td colspan="4" class="py-1 text-right">
                  <div class="inline-flex items-center gap-1.5">
                    <span class="w-1.5 h-1.5 rounded-full bg-sky-600"></span>
                    <span>Uang Kembalian Pasien:</span>
                  </div>
                </td>
                <td class="py-1 text-right tabular-nums font-mono text-sky-800 font-bold">
                  {{ formatCurrency(getReceiptChange(invoiceToPrint)) }}
                </td>
              </tr>
            </tfoot>
          </table>

          <!-- Terbilang Box -->
          <div class="bg-slate-100 p-2.5 rounded text-xs mb-6">
            <span class="text-slate-500 block text-[10px] uppercase font-bold tracking-wider">Terbilang:</span>
            <span class="font-bold italic text-slate-800">"{{ terbilang(invoiceToPrint.total_tagihan) }}"</span>
          </div>

          <!-- Tanda Tangan Kasir -->
          <div class="grid grid-cols-2 text-xs text-center pt-4">
            <div>
              <p class="text-slate-500">Pasien / Keluarga Pasien,</p>
              <div class="h-16"></div>
              <p class="font-bold border-t border-slate-300 mx-8 pt-1">{{ invoiceToPrint.patient_name }}</p>
            </div>
            <div>
              <p class="text-slate-500">Kasir Penerima,</p>
              <div class="h-16"></div>
              <p class="font-bold border-t border-slate-300 mx-8 pt-1">{{ authStore.user?.full_name || 'Kasir SIMRS' }}</p>
            </div>
          </div>

          <!-- Digital Verification QR Code (Fase 1) -->
          <div class="mt-4 pt-3 border-t border-slate-200 flex items-center justify-between">
            <div class="flex items-center gap-3">
              <img
                :src="`https://api.qrserver.com/v1/create-qr-code/?size=60x60&data=${encodeURIComponent('http://localhost:5173/verify-doc?type=INVOICE&id=' + invoiceToPrint.id + '&token=' + invoiceToPrint.nomor_invoice)}`"
                alt="QR TTE"
                class="w-12 h-12 rounded border border-slate-300 bg-white p-0.5"
              />
              <div class="text-[10px] text-slate-500 text-left">
                <div class="font-bold text-slate-800 flex items-center gap-1">
                  <span>TTE & Dokumen Digital Terverifikasi</span>
                </div>
                <div>Pindai QR untuk validasi keaslian dokumen di Portal Publik SIMRS</div>
                <div class="font-mono text-sky-700 font-semibold">{{ invoiceToPrint.nomor_invoice }}</div>
              </div>
            </div>
            <div class="text-[9px] text-slate-400 text-right">
              Dokumen ini dicetak otomatis dari SIMRS-ASOY<br/>
              dan sah tanpa memerlukan cap basah.
            </div>
          </div>
        </div>

        <div class="flex items-center justify-between pt-2">
          <span class="text-xs text-slate-400">Bukti pembayaran sah, cetak untuk tanda bukti ke pasien.</span>
          <div class="flex items-center gap-2">
            <BaseButton variant="ghost" size="sm" @click="showReceiptModal = false">
              Tutup
            </BaseButton>
            <BaseButton variant="primary" size="sm" @click="printDirect">
              <Printer class="w-4 h-4 mr-1.5" />
              Cetak Kuitansi
            </BaseButton>
          </div>
        </div>
      </div>
    </BaseModal>
  </div>
</template>

<style scoped>
@media print {
  body * {
    visibility: hidden;
  }
  #receipt-area, #receipt-area * {
    visibility: visible;
  }
  #receipt-area {
    position: absolute;
    left: 0;
    top: 0;
    width: 100%;
    border: none;
    box-shadow: none;
    padding: 0;
  }
}
</style>
