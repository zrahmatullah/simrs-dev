<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '@/api/client'
import type { InventoryItem, PurchaseOrder, StockMutation, MasterCategory, PurchaseRequisition, GoodsReceiptNote, Supplier } from '@/types'
import { useUiStore } from '@/stores/ui'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import SearchableSelect from '@/components/common/SearchableSelect.vue'
import BaseDatePicker from '@/components/common/BaseDatePicker.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import BaseModal from '@/components/common/BaseModal.vue'
import ConfirmModal from '@/components/common/ConfirmModal.vue'
import {
  Boxes,
  Package,
  ShoppingCart,
  ArrowRightLeft,
  Search,
  Plus,
  RefreshCw,
  AlertTriangle,
  CheckCircle2,
  FileText,
  Clock,
  Trash2,
  Edit,
  ArrowUpRight,
  TrendingDown,
  Warehouse
} from 'lucide-vue-next'

const uiStore = useUiStore()

const activeTab = ref<'pr' | 'po' | 'grn' | 'stok' | 'mutasi'>('pr')
const loading = ref(false)

// ----------------------------------------------------
// PIPELINE 1: PR (Purchase Requisition)
// ----------------------------------------------------
const prs = ref<PurchaseRequisition[]>([])
const showCreatePRModal = ref(false)
const submittingPR = ref(false)
interface PRItemForm {
  item_id: number
  qty: number
  unit_price: number
}
const prForm = ref({
  department: 'FARMASI',
  notes: '',
  items: [] as PRItemForm[]
})

const fetchPRs = async () => {
  loading.value = true
  try {
    const res = await api.get<PurchaseRequisition[]>('/logistics/pr')
    prs.value = res.data || []
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal memuat PR', 'error')
  } finally {
    loading.value = false
  }
}

const openCreatePRModal = () => {
  const firstItem = items.value[0]
  prForm.value = {
    department: 'FARMASI',
    notes: 'Permintaan pengadaan rutin obat depo',
    items: [
      {
        item_id: firstItem?.id || 1,
        qty: 50,
        unit_price: firstItem?.buy_price || 1000
      }
    ]
  }
  showCreatePRModal.value = true
}

const addPRItemRow = () => {
  const firstItem = items.value[0]
  prForm.value.items.push({
    item_id: firstItem?.id || 1,
    qty: 20,
    unit_price: firstItem?.buy_price || 1000
  })
}

const removePRItemRow = (index: number) => {
  prForm.value.items.splice(index, 1)
}

const submitPR = async () => {
  if (prForm.value.items.length === 0) {
    uiStore.showToast('Minimal satu item barang harus ditambahkan ke PR', 'warning')
    return
  }
  submittingPR.value = true
  try {
    await api.post('/logistics/pr', prForm.value)
    uiStore.showToast('Purchase Requisition (PR) berhasil diterbitkan!', 'success')
    showCreatePRModal.value = false
    await fetchPRs()
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal membuat PR', 'error')
  } finally {
    submittingPR.value = false
  }
}

// ----------------------------------------------------
// PIPELINE 3: GRN (Goods Receipt Note)
// ----------------------------------------------------
const grns = ref<GoodsReceiptNote[]>([])
const showCreateGRNModal = ref(false)
const submittingGRN = ref(false)
const grnForm = ref({
  po_id: '',
  receiver_name: 'Staf Gudang & Logistik',
  delivery_order_number: '',
  notes: ''
})

const fetchGRNs = async () => {
  loading.value = true
  try {
    const res = await api.get<GoodsReceiptNote[]>('/logistics/grn')
    grns.value = res.data || []
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal memuat GRN', 'error')
  } finally {
    loading.value = false
  }
}

const openCreateGRNModal = (po?: PurchaseOrder) => {
  grnForm.value = {
    po_id: po?.id || (pos.value[0]?.id || ''),
    receiver_name: 'Staf Gudang Logistik',
    delivery_order_number: `DO-${new Date().getFullYear()}-${Math.floor(1000 + Math.random() * 9000)}`,
    notes: 'Penerimaan fisik barang lengkap sesuai spesifikasi'
  }
  showCreateGRNModal.value = true
}

const submitGRN = async () => {
  if (!grnForm.value.po_id) {
    uiStore.showToast('Pilih Purchase Order yang diterima', 'warning')
    return
  }
  submittingGRN.value = true
  try {
    await api.post('/logistics/grn', grnForm.value)
    uiStore.showToast('Penerimaan barang (GRN) berhasil dicatat! Stok gudang utama bertambah & mutasi tercatat.', 'success')
    showCreateGRNModal.value = false
    await fetchGRNs()
    await fetchPOs()
    await fetchItems()
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal mencatat GRN', 'error')
  } finally {
    submittingGRN.value = false
  }
}

// Tab 1: Stok
const items = ref<InventoryItem[]>([])
const categories = ref<any[]>([])
const searchItem = ref('')
const selectedCategory = ref<string>('0')
const onlyLowStock = ref(false)

// Item modal
const showItemModal = ref(false)
const isEditingItem = ref(false)
const submittingItem = ref(false)
const itemForm = ref({
  id: 0,
  category_id: 1,
  item_code: '',
  item_name: '',
  generic_name: '',
  unit: 'Tablet',
  buy_price: 0,
  sell_price: 0,
  minimum_stock: 10,
  total_stock: 0,
  depo_stock: 0,
  is_medicine: true,
  rack_location: '',
  status_aktif: true,
})

// Tab 2: Purchase Order
const pos = ref<PurchaseOrder[]>([])
const poStatusFilter = ref<string>('SEMUA')
const showCreatePoModal = ref(false)
const submittingPo = ref(false)
const showReceivePoConfirm = ref(false)
const poToReceive = ref<PurchaseOrder | null>(null)

// Master Suppliers
const suppliers = ref<Supplier[]>([])
const fetchSuppliers = async () => {
  try {
    const res = await api.get<Supplier[]>('/master/suppliers')
    if (Array.isArray(res.data) && res.data.length > 0) {
      suppliers.value = res.data
    } else if ((res.data as any)?.data && Array.isArray((res.data as any).data)) {
      suppliers.value = (res.data as any).data
    }
  } catch (err: any) {
    console.warn('Gagal memuat master supplier', err)
  }
}

interface PoItemForm {
  item_id: number
  qty: number
  unit_price: number
}

const poForm = ref({
  supplier_id: '' as number | '',
  supplier_name: '',
  order_date: new Date().toISOString().split('T')[0],
  expected_date: '',
  catatan: '',
  items: [] as PoItemForm[],
})

// Tab 3: Mutasi
const mutations = ref<StockMutation[]>([])
const showMutationModal = ref(false)
const submittingMutation = ref(false)
const mutationForm = ref({
  item_id: 0,
  transaction_type: 'MUTASI_DEPO', // MUTASI_DEPO, ADJUSTMENT
  qty: 1,
  batch_number: '',
  expiry_date: '',
  source_depo: 'GUDANG_UTAMA',
  dest_depo: 'DEPO_FARMASI',
  notes: '',
})

// Unit options
const unitOptions = [
  { label: 'Tablet', value: 'Tablet' },
  { label: 'Kapsul', value: 'Kapsul' },
  { label: 'Botol / Sirup', value: 'Botol' },
  { label: 'Ampul / Vial', value: 'Ampul' },
  { label: 'Tube (Salep)', value: 'Tube' },
  { label: 'Pcs / Box', value: 'Pcs' },
  { label: 'Sachet', value: 'Sachet' },
  { label: 'Roll / Set', value: 'Roll' },
]

// Data Fetchers
const fetchCategories = async () => {
  try {
    const res = await api.get<any[]>('/master/categories')
    categories.value = res.data || []
  } catch (err) {
    console.error('Gagal ambil kategori:', err)
  }
}

const fetchItems = async () => {
  loading.value = true
  try {
    const params: Record<string, any> = {
      page: 1,
      limit: 100,
    }
    if (searchItem.value) params.search = searchItem.value
    if (selectedCategory.value !== '0') params.category_id = selectedCategory.value
    if (onlyLowStock.value) params.low_stock = 'true'

    const res = await api.get<InventoryItem[]>('/logistics/items', params)
    items.value = res.data || []
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal memuat katalog barang', 'error')
  } finally {
    loading.value = false
  }
}

const fetchPOs = async () => {
  loading.value = true
  try {
    const params: Record<string, any> = {
      page: 1,
      limit: 50,
    }
    if (poStatusFilter.value !== 'SEMUA') params.status = poStatusFilter.value
    const res = await api.get<PurchaseOrder[]>('/logistics/po', params)
    pos.value = res.data || []
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal memuat Purchase Orders', 'error')
  } finally {
    loading.value = false
  }
}

const fetchMutations = async () => {
  loading.value = true
  try {
    const res = await api.get<StockMutation[]>('/logistics/mutations', { limit: 100 })
    mutations.value = res.data || []
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal memuat riwayat mutasi stok', 'error')
  } finally {
    loading.value = false
  }
}

// Low stock items count
const lowStockCount = computed(() => {
  return items.value.filter(i => (i.total_stock + i.depo_stock) <= i.minimum_stock).length
})

// Item Actions
const openCreateItemModal = () => {
  isEditingItem.value = false
  itemForm.value = {
    id: 0,
    category_id: categories.value[0]?.id || 1,
    item_code: 'OBT-' + Math.floor(1000 + Math.random() * 9000),
    item_name: '',
    generic_name: '',
    unit: 'Tablet',
    buy_price: 0,
    sell_price: 0,
    minimum_stock: 20,
    total_stock: 0,
    depo_stock: 0,
    is_medicine: true,
    rack_location: 'G-01',
    status_aktif: true,
  }
  showItemModal.value = true
}

const openEditItemModal = (item: InventoryItem) => {
  isEditingItem.value = true
  itemForm.value = {
    id: item.id,
    category_id: item.category_id,
    item_code: item.item_code,
    item_name: item.item_name,
    generic_name: item.generic_name,
    unit: item.unit,
    buy_price: item.buy_price,
    sell_price: item.sell_price,
    minimum_stock: item.minimum_stock,
    total_stock: item.total_stock,
    depo_stock: item.depo_stock,
    is_medicine: item.is_medicine,
    rack_location: item.rack_location,
    status_aktif: item.status_aktif,
  }
  showItemModal.value = true
}

const saveItem = async () => {
  if (!itemForm.value.item_name.trim()) {
    uiStore.showToast('Nama barang wajib diisi', 'warning')
    return
  }

  submittingItem.value = true
  try {
    if (isEditingItem.value) {
      await api.put(`/logistics/items/${itemForm.value.id}`, itemForm.value)
      uiStore.showToast('Data barang berhasil diperbarui', 'success')
    } else {
      await api.post('/logistics/items', itemForm.value)
      uiStore.showToast('Barang baru berhasil ditambahkan', 'success')
    }
    showItemModal.value = false
    await fetchItems()
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal menyimpan barang', 'error')
  } finally {
    submittingItem.value = false
  }
}

// PO Actions
const openCreatePoModal = () => {
  const firstSup = suppliers.value[0]
  poForm.value = {
    supplier_id: firstSup?.id || '',
    supplier_name: firstSup?.name || '',
    order_date: new Date().toISOString().split('T')[0],
    expected_date: '',
    catatan: '',
    items: [
      {
        item_id: items.value[0]?.id || 1,
        qty: 50,
        unit_price: items.value[0]?.buy_price || 1000,
      },
    ],
  }
  showCreatePoModal.value = true
}

const onSupplierSelect = (val: string | number) => {
  const selected = suppliers.value.find(s => String(s.id) === String(val))
  if (selected) {
    poForm.value.supplier_id = selected.id
    poForm.value.supplier_name = selected.name
  }
}

const addPoItemRow = () => {
  const firstItem = items.value[0]
  poForm.value.items.push({
    item_id: firstItem?.id || 1,
    qty: 10,
    unit_price: firstItem?.buy_price || 1000,
  })
}

const removePoItemRow = (index: number) => {
  poForm.value.items.splice(index, 1)
}

const onPoItemChange = (row: PoItemForm) => {
  const selected = items.value.find(i => i.id === Number(row.item_id))
  if (selected) {
    row.unit_price = selected.buy_price
  }
}

const poTotalAmount = computed(() => {
  return poForm.value.items.reduce((sum, item) => sum + (item.qty * item.unit_price), 0)
})

const savePo = async () => {
  if (!poForm.value.supplier_id && !poForm.value.supplier_name.trim()) {
    uiStore.showToast('Silakan pilih Supplier / PBF dari Master Supplier', 'warning')
    return
  }
  if (poForm.value.items.length === 0) {
    uiStore.showToast('Minimal masukkan 1 barang dalam PO', 'warning')
    return
  }

  submittingPo.value = true
  try {
    const payload = {
      supplier_id: poForm.value.supplier_id ? Number(poForm.value.supplier_id) : undefined,
      supplier_name: poForm.value.supplier_name,
      order_date: poForm.value.order_date,
      expected_date: poForm.value.expected_date ? poForm.value.expected_date : undefined,
      catatan: poForm.value.catatan,
      items: poForm.value.items.map(it => ({
        item_id: Number(it.item_id),
        qty: Number(it.qty),
        unit_price: Number(it.unit_price),
      })),
    }

    await api.post('/logistics/po', payload)
    uiStore.showToast('Purchase Order berhasil diterbitkan', 'success')
    showCreatePoModal.value = false
    await fetchPOs()
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal menerbitkan PO', 'error')
  } finally {
    submittingPo.value = false
  }
}

const confirmReceivePo = (po: PurchaseOrder) => {
  poToReceive.value = po
  showReceivePoConfirm.value = true
}

const executeReceivePo = async () => {
  if (!poToReceive.value) return
  try {
    await api.post(`/logistics/po/${poToReceive.value.id}/receive`)
    uiStore.showToast('Penerimaan PO berhasil! Stok gudang utama telah bertambah', 'success')
    showReceivePoConfirm.value = false
    await fetchPOs()
    await fetchItems()
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal menerima PO', 'error')
  }
}

// Mutation Actions
const openMutationModal = (defaultItem?: InventoryItem) => {
  mutationForm.value = {
    item_id: defaultItem ? defaultItem.id : (items.value[0]?.id || 1),
    transaction_type: 'MUTASI_DEPO',
    qty: 10,
    batch_number: 'BATCH-' + Math.floor(1000 + Math.random() * 9000),
    expiry_date: '',
    source_depo: 'GUDANG_UTAMA',
    dest_depo: 'DEPO_FARMASI',
    notes: 'Mutasi rutin untuk pelayanan resep farmasi',
  }
  showMutationModal.value = true
}

const saveMutation = async () => {
  if (!mutationForm.value.item_id) {
    uiStore.showToast('Pilih barang terlebih dahulu', 'warning')
    return
  }
  if (mutationForm.value.qty <= 0) {
    uiStore.showToast('Jumlah mutasi harus lebih besar dari 0', 'warning')
    return
  }

  submittingMutation.value = true
  try {
    const payload = {
      item_id: Number(mutationForm.value.item_id),
      transaction_type: mutationForm.value.transaction_type,
      qty: Number(mutationForm.value.qty),
      batch_number: mutationForm.value.batch_number,
      expiry_date: mutationForm.value.expiry_date ? mutationForm.value.expiry_date : undefined,
      source_depo: mutationForm.value.source_depo,
      dest_depo: mutationForm.value.dest_depo,
      notes: mutationForm.value.notes,
    }
    await api.post('/logistics/mutations', payload)
    uiStore.showToast('Mutasi stok berhasil dicatat', 'success')
    showMutationModal.value = false
    await fetchMutations()
    await fetchItems()
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal mencatat mutasi stok', 'error')
  } finally {
    submittingMutation.value = false
  }
}

const formatCurrency = (val: number) => {
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(val)
}

const formatDate = (isoString?: string | null) => {
  if (!isoString) return '-'
  return new Date(isoString).toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}

onMounted(async () => {
  await fetchCategories()
  await fetchItems()
  await fetchSuppliers()
})
</script>

<template>
  <div class="space-y-4">
    <!-- Header Page -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-white p-4 rounded-xl border border-slate-200 shadow-sm">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-lg bg-sky-50 text-sky-600 flex items-center justify-center border border-sky-100">
          <Warehouse class="w-5 h-5" />
        </div>
        <div>
          <h1 class="text-xl font-bold text-slate-800">Gudang & Logistik Farmasi</h1>
          <p class="text-xs text-slate-500">Manajemen Pengadaan PO, Kartu Stok Multi-Depo, & Penyesuaian Barang Medis</p>
        </div>
      </div>

      <!-- Action Buttons based on active tab -->
      <div class="flex items-center gap-2">
        <template v-if="activeTab === 'pr'">
          <BaseButton variant="primary" size="sm" @click="openCreatePRModal">
            <Plus class="w-4 h-4 mr-1.5" />
            Buat Permintaan (PR)
          </BaseButton>
        </template>
        <template v-else-if="activeTab === 'po'">
          <BaseButton variant="primary" size="sm" @click="openCreatePoModal">
            <Plus class="w-4 h-4 mr-1.5" />
            Buat Purchase Order (PO)
          </BaseButton>
        </template>
        <template v-else-if="activeTab === 'grn'">
          <BaseButton variant="primary" size="sm" @click="() => openCreateGRNModal()">
            <Plus class="w-4 h-4 mr-1.5" />
            Penerimaan Barang (GRN)
          </BaseButton>
        </template>
        <template v-else-if="activeTab === 'stok'">
          <BaseButton variant="primary" size="sm" @click="openCreateItemModal">
            <Plus class="w-4 h-4 mr-1.5" />
            Tambah Barang Baru
          </BaseButton>
        </template>
        <template v-else-if="activeTab === 'mutasi'">
          <BaseButton variant="primary" size="sm" @click="() => openMutationModal()">
            <ArrowRightLeft class="w-4 h-4 mr-1.5" />
            Catat Mutasi / Koreksi Stok
          </BaseButton>
        </template>
      </div>
    </div>

    <!-- Pipeline Flow Stepper Banner -->
    <div class="bg-white p-3 rounded-xl border border-slate-200 shadow-sm flex items-center justify-between text-xs overflow-x-auto gap-2">
      <div class="flex items-center gap-2 cursor-pointer" :class="activeTab === 'pr' ? 'text-sky-700 font-bold' : 'text-slate-500'" @click="activeTab = 'pr'; fetchPRs()">
        <span class="w-5 h-5 rounded-full flex items-center justify-center text-[10px] text-white shrink-0" :class="activeTab === 'pr' ? 'bg-sky-600' : 'bg-slate-300'">1</span>
        <span>PR (Permintaan Pengadaan)</span>
      </div>
      <span class="text-slate-300">➔</span>
      <div class="flex items-center gap-2 cursor-pointer" :class="activeTab === 'po' ? 'text-sky-700 font-bold' : 'text-slate-500'" @click="activeTab = 'po'; fetchPOs()">
        <span class="w-5 h-5 rounded-full flex items-center justify-center text-[10px] text-white shrink-0" :class="activeTab === 'po' ? 'bg-sky-600' : 'bg-slate-300'">2</span>
        <span>PO (Pemesanan Supplier)</span>
      </div>
      <span class="text-slate-300">➔</span>
      <div class="flex items-center gap-2 cursor-pointer" :class="activeTab === 'grn' ? 'text-sky-700 font-bold' : 'text-slate-500'" @click="activeTab = 'grn'; fetchGRNs()">
        <span class="w-5 h-5 rounded-full flex items-center justify-center text-[10px] text-white shrink-0" :class="activeTab === 'grn' ? 'bg-sky-600' : 'bg-slate-300'">3</span>
        <span>GRN (Penerimaan Barang)</span>
      </div>
      <span class="text-slate-300">➔</span>
      <div class="flex items-center gap-2 cursor-pointer" :class="activeTab === 'stok' ? 'text-sky-700 font-bold' : 'text-slate-500'" @click="activeTab = 'stok'; fetchItems()">
        <span class="w-5 h-5 rounded-full flex items-center justify-center text-[10px] text-white shrink-0" :class="activeTab === 'stok' ? 'bg-sky-600' : 'bg-slate-300'">4</span>
        <span>Update Stok Otomatis & Mutasi</span>
      </div>
    </div>

    <!-- Alert Banner Low Stock -->
    <div
      v-if="lowStockCount > 0"
      class="bg-amber-50 border border-amber-200 rounded-xl p-3 flex items-center justify-between gap-3 text-amber-800 text-xs"
    >
      <div class="flex items-center gap-2">
        <AlertTriangle class="w-4 h-4 text-amber-600 shrink-0" />
        <span>
          Terdapat <strong>{{ lowStockCount }} item barang</strong> dengan stok berada di bawah batas minimum. Segera buat Purchase Order pengadaan!
        </span>
      </div>
      <BaseButton
        variant="secondary"
        size="xs"
        @click="() => { activeTab = 'stok'; onlyLowStock = !onlyLowStock; fetchItems() }"
      >
        {{ onlyLowStock ? 'Tampilkan Semua Barang' : 'Filter Stok Menipis' }}
      </BaseButton>
    </div>

    <!-- Navigation Tabs -->
    <div class="flex items-center border-b border-slate-200 gap-2 bg-white px-4 pt-2 rounded-t-xl overflow-x-auto">
      <button
        class="flex items-center gap-2 px-3 py-2 text-xs font-semibold border-b-2 transition-colors shrink-0"
        :class="activeTab === 'pr' ? 'border-sky-600 text-sky-600' : 'border-transparent text-slate-500 hover:text-slate-700'"
        @click="() => { activeTab = 'pr'; fetchPRs() }"
      >
        <FileText class="w-4 h-4" />
        1. Purchase Requisition (PR)
        <span class="text-[10px] px-1.5 py-0.5 rounded-full bg-slate-100 text-slate-600 tabular-nums">{{ prs.length }}</span>
      </button>

      <button
        class="flex items-center gap-2 px-3 py-2 text-xs font-semibold border-b-2 transition-colors shrink-0"
        :class="activeTab === 'po' ? 'border-sky-600 text-sky-600' : 'border-transparent text-slate-500 hover:text-slate-700'"
        @click="() => { activeTab = 'po'; fetchPOs() }"
      >
        <ShoppingCart class="w-4 h-4" />
        2. Purchase Order (PO)
        <span class="text-[10px] px-1.5 py-0.5 rounded-full bg-slate-100 text-slate-600 tabular-nums">{{ pos.length }}</span>
      </button>

      <button
        class="flex items-center gap-2 px-3 py-2 text-xs font-semibold border-b-2 transition-colors shrink-0"
        :class="activeTab === 'grn' ? 'border-sky-600 text-sky-600' : 'border-transparent text-slate-500 hover:text-slate-700'"
        @click="() => { activeTab = 'grn'; fetchGRNs() }"
      >
        <Package class="w-4 h-4" />
        3. Goods Receipt Note (GRN)
        <span class="text-[10px] px-1.5 py-0.5 rounded-full bg-slate-100 text-slate-600 tabular-nums">{{ grns.length }}</span>
      </button>

      <button
        class="flex items-center gap-2 px-3 py-2 text-xs font-semibold border-b-2 transition-colors shrink-0"
        :class="activeTab === 'stok' ? 'border-sky-600 text-sky-600' : 'border-transparent text-slate-500 hover:text-slate-700'"
        @click="() => { activeTab = 'stok'; fetchItems() }"
      >
        <Boxes class="w-4 h-4" />
        4. Stok Gudang & Depo
        <span class="text-[10px] px-1.5 py-0.5 rounded-full bg-slate-100 text-slate-600 tabular-nums">{{ items.length }}</span>
      </button>

      <button
        class="flex items-center gap-2 px-3 py-2 text-xs font-semibold border-b-2 transition-colors shrink-0"
        :class="activeTab === 'mutasi' ? 'border-sky-600 text-sky-600' : 'border-transparent text-slate-500 hover:text-slate-700'"
        @click="() => { activeTab = 'mutasi'; fetchMutations() }"
      >
        <ArrowRightLeft class="w-4 h-4" />
        5. Kartu Stok Mutasi
      </button>
    </div>

    <!-- ========================================================================= -->
    <!-- PIPELINE TAB 1: PURCHASE REQUISITION (PR)                                  -->
    <!-- ========================================================================= -->
    <div v-if="activeTab === 'pr'" class="space-y-3">
      <div class="bg-white p-3 rounded-xl border border-slate-200 shadow-sm flex items-center justify-between">
        <div class="flex items-center gap-2">
          <FileText class="w-4 h-4 text-sky-600" />
          <span class="text-xs font-bold text-slate-800">Daftar Permintaan Pengadaan Barang (Purchase Requisition)</span>
        </div>
        <BaseButton variant="primary" size="sm" @click="openCreatePRModal">
          <Plus class="w-4 h-4 mr-1" />
          Buat PR Baru
        </BaseButton>
      </div>

      <div class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
        <table class="w-full text-left text-xs">
          <thead class="bg-slate-50 text-slate-600 border-b border-slate-200 font-bold">
            <tr>
              <th class="p-3">No. PR</th>
              <th class="p-3">Tanggal Permintaan</th>
              <th class="p-3">Pemohon / Departemen</th>
              <th class="p-3 text-right">Estimasi Total</th>
              <th class="p-3 text-center">Status</th>
              <th class="p-3">Catatan</th>
              <th class="p-3 text-right">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="prs.length === 0">
              <td colspan="7" class="p-6 text-center text-slate-400">
                Belum ada Purchase Requisition (PR) yang diterbitkan.
              </td>
            </tr>
            <tr v-for="pr in prs" :key="pr.id" class="hover:bg-slate-50/80">
              <td class="p-3 font-mono font-bold text-sky-700">{{ pr.pr_number }}</td>
              <td class="p-3 text-slate-600">{{ new Date(pr.request_date).toLocaleDateString('id-ID') }}</td>
              <td class="p-3">
                <span class="font-semibold text-slate-800 block">{{ pr.requester_name || 'Staff Farmasi' }}</span>
                <span class="text-[10px] text-slate-500 block">{{ pr.department }}</span>
              </td>
              <td class="p-3 text-right font-mono font-bold text-slate-800 tabular-nums">
                {{ formatCurrency(pr.total_estimated) }}
              </td>
              <td class="p-3 text-center">
                <StatusBadge :status="pr.status" />
              </td>
              <td class="p-3 text-slate-600 max-w-xs truncate">{{ pr.notes || '-' }}</td>
              <td class="p-3 text-right">
                <BaseButton
                  variant="outline"
                  size="xs"
                  @click="() => { activeTab = 'po'; openCreatePoModal() }"
                >
                  Proses ke PO ➔
                </BaseButton>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- ========================================================================= -->
    <!-- PIPELINE TAB 3: GOODS RECEIPT NOTE (GRN / PENERIMAAN FISIK BARANG)        -->
    <!-- ========================================================================= -->
    <div v-if="activeTab === 'grn'" class="space-y-3">
      <div class="bg-white p-3 rounded-xl border border-slate-200 shadow-sm flex items-center justify-between">
        <div class="flex items-center gap-2">
          <Package class="w-4 h-4 text-emerald-600" />
          <span class="text-xs font-bold text-slate-800">Daftar Penerimaan Fisik Barang dari Supplier (GRN)</span>
        </div>
        <BaseButton variant="primary" size="sm" class="!bg-emerald-600 hover:!bg-emerald-700" @click="() => openCreateGRNModal()">
          <Plus class="w-4 h-4 mr-1" />
          Penerimaan Barang (GRN)
        </BaseButton>
      </div>

      <div class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
        <table class="w-full text-left text-xs">
          <thead class="bg-slate-50 text-slate-600 border-b border-slate-200 font-bold">
            <tr>
              <th class="p-3">No. GRN</th>
              <th class="p-3">Tanggal Terima</th>
              <th class="p-3">No. PO & Supplier</th>
              <th class="p-3">No. Surat Jalan / DO</th>
              <th class="p-3">Petugas Penerima</th>
              <th class="p-3 text-center">Status</th>
              <th class="p-3">Keterangan</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="grns.length === 0">
              <td colspan="7" class="p-6 text-center text-slate-400">
                Belum ada Goods Receipt Note (GRN) penerimaan barang yang dicatat.
              </td>
            </tr>
            <tr v-for="grn in grns" :key="grn.id" class="hover:bg-slate-50/80">
              <td class="p-3 font-mono font-bold text-emerald-700">{{ grn.grn_number }}</td>
              <td class="p-3 text-slate-600">{{ new Date(grn.received_date).toLocaleDateString('id-ID') }}</td>
              <td class="p-3">
                <span class="font-mono font-semibold text-slate-800 block">{{ grn.po_number }}</span>
                <span class="text-[10px] text-slate-500 block">{{ grn.supplier_name }}</span>
              </td>
              <td class="p-3 font-mono text-slate-700">{{ grn.delivery_order_number || '-' }}</td>
              <td class="p-3 font-semibold text-slate-800">{{ grn.receiver_name }}</td>
              <td class="p-3 text-center">
                <span class="px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-100 text-emerald-800 border border-emerald-200">
                  {{ grn.status }}
                </span>
              </td>
              <td class="p-3 text-slate-600 max-w-xs truncate">{{ grn.notes || '-' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- TAB 1: STOK BARANG -->
    <div v-if="activeTab === 'stok'" class="space-y-3">
      <!-- Filter Bar -->
      <div class="bg-white p-3 rounded-xl border border-slate-200 shadow-sm flex flex-col md:flex-row items-center gap-3">
        <div class="w-full md:w-1/3 relative">
          <Search class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
          <input
            v-model="searchItem"
            type="text"
            placeholder="Cari kode, nama obat, atau generik..."
            class="w-full pl-9 pr-3 py-1.5 text-xs border border-slate-200 rounded-lg focus:outline-none focus:border-sky-500"
            @keyup.enter="fetchItems"
          />
        </div>

        <div class="w-full md:w-1/4">
          <BaseSelect
            v-model="selectedCategory"
            :options="[
              { value: '0', label: 'Semua Kategori' },
              ...categories.map(c => ({ value: String(c.id), label: c.name }))
            ]"
            compact
            @change="fetchItems"
          />
        </div>

        <div class="flex items-center gap-2 ml-auto">
          <label class="flex items-center gap-1.5 text-xs text-slate-600 cursor-pointer select-none">
            <input
              v-model="onlyLowStock"
              type="checkbox"
              class="rounded text-sky-600 focus:ring-sky-500 h-3.5 w-3.5"
              @change="fetchItems"
            />
            Hanya Stok Menipis (<= Min)
          </label>

          <BaseButton variant="secondary" size="xs" @click="fetchItems">
            <RefreshCw class="w-3.5 h-3.5 mr-1" :class="{ 'animate-spin': loading }" />
            Segarkan
          </BaseButton>
        </div>
      </div>

      <!-- Table Items -->
      <div class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs text-slate-600">
            <thead class="bg-slate-50 text-slate-700 font-semibold border-b border-slate-200 uppercase tracking-wider">
              <tr>
                <th class="py-2.5 px-3">Kode / Rak</th>
                <th class="py-2.5 px-3">Nama Obat & Generik</th>
                <th class="py-2.5 px-3">Kategori & Satuan</th>
                <th class="py-2.5 px-3 text-right">Harga Beli</th>
                <th class="py-2.5 px-3 text-right">Harga Jual</th>
                <th class="py-2.5 px-3 text-center">Gudang Utama</th>
                <th class="py-2.5 px-3 text-center">Depo Farmasi</th>
                <th class="py-2.5 px-3 text-center">Total Stok</th>
                <th class="py-2.5 px-3 text-center">Min. Stok</th>
                <th class="py-2.5 px-3 text-center">Status</th>
                <th class="py-2.5 px-3 text-center">Aksi</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-if="loading">
                <td colspan="11" class="text-center py-10 text-slate-400">
                  <RefreshCw class="w-6 h-6 animate-spin mx-auto mb-2 text-sky-500" />
                  Memuat data stok barang...
                </td>
              </tr>
              <tr v-else-if="items.length === 0">
                <td colspan="11" class="text-center py-12 text-slate-400">
                  <Package class="w-10 h-10 mx-auto mb-2 text-slate-300" />
                  Tidak ada barang medis yang sesuai filter.
                </td>
              </tr>
              <tr
                v-for="item in items"
                :key="item.id"
                class="hover:bg-slate-50 transition-colors"
                :class="{ 'bg-amber-50/40': (item.total_stock + item.depo_stock) <= item.minimum_stock }"
              >
                <td class="py-2.5 px-3">
                  <div class="font-mono font-bold text-sky-800">{{ item.item_code }}</div>
                  <div class="text-[10px] text-slate-400">Rak: {{ item.rack_location || '-' }}</div>
                </td>

                <td class="py-2.5 px-3">
                  <div class="font-semibold text-slate-800">{{ item.item_name }}</div>
                  <div class="text-[10px] text-slate-400 italic">{{ item.generic_name || '-' }}</div>
                </td>

                <td class="py-2.5 px-3">
                  <div class="text-slate-700">{{ item.category_name || 'Obat' }}</div>
                  <div class="text-[10px] text-slate-400">Satuan: {{ item.unit }}</div>
                </td>

                <td class="py-2.5 px-3 text-right tabular-nums">
                  {{ formatCurrency(item.buy_price) }}
                </td>

                <td class="py-2.5 px-3 text-right font-semibold text-slate-800 tabular-nums">
                  {{ formatCurrency(item.sell_price) }}
                </td>

                <td class="py-2.5 px-3 text-center font-mono font-semibold text-sky-700 tabular-nums">
                  {{ item.total_stock }}
                </td>

                <td class="py-2.5 px-3 text-center font-mono font-semibold text-emerald-700 tabular-nums">
                  {{ item.depo_stock }}
                </td>

                <td class="py-2.5 px-3 text-center">
                  <span
                    class="font-mono font-bold px-2 py-0.5 rounded text-xs tabular-nums"
                    :class="(item.total_stock + item.depo_stock) <= item.minimum_stock ? 'bg-amber-100 text-amber-800' : 'bg-slate-100 text-slate-800'"
                  >
                    {{ item.total_stock + item.depo_stock }}
                  </span>
                </td>

                <td class="py-2.5 px-3 text-center font-mono text-slate-500 tabular-nums">
                  {{ item.minimum_stock }}
                </td>

                <td class="py-2.5 px-3 text-center">
                  <span
                    v-if="(item.total_stock + item.depo_stock) === 0"
                    class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-rose-100 text-rose-700"
                  >
                    HABIS
                  </span>
                  <span
                    v-else-if="(item.total_stock + item.depo_stock) <= item.minimum_stock"
                    class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-amber-100 text-amber-700"
                  >
                    MENIPIS
                  </span>
                  <span
                    v-else
                    class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-100 text-emerald-700"
                  >
                    AMAN
                  </span>
                </td>

                <td class="py-2.5 px-3 text-center">
                  <div class="flex items-center justify-center gap-1">
                    <button
                      class="p-1 hover:bg-slate-200 rounded text-slate-600 hover:text-sky-600"
                      title="Mutasi ke Depo"
                      @click="openMutationModal(item)"
                    >
                      <ArrowRightLeft class="w-3.5 h-3.5" />
                    </button>
                    <button
                      class="p-1 hover:bg-slate-200 rounded text-slate-600 hover:text-sky-600"
                      title="Edit Barang"
                      @click="openEditItemModal(item)"
                    >
                      <Edit class="w-3.5 h-3.5" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- TAB 2: PURCHASE ORDER -->
    <div v-else-if="activeTab === 'po'" class="space-y-3">
      <!-- Filter Bar -->
      <div class="bg-white p-3 rounded-xl border border-slate-200 shadow-sm flex items-center justify-between gap-3">
        <div class="flex items-center gap-2">
          <span class="text-xs font-semibold text-slate-700 shrink-0">Filter Status PO:</span>
          <div class="w-60">
            <BaseSelect
              v-model="poStatusFilter"
              :options="[
                { value: 'SEMUA', label: 'Semua Status' },
                { value: 'SUBMITTED', label: 'Diajukan / Menunggu Barang' },
                { value: 'RECEIVED', label: 'Diterima & Selesai' },
                { value: 'CANCELLED', label: 'Dibatalkan' },
              ]"
              compact
              @change="fetchPOs"
            />
          </div>
        </div>

        <BaseButton variant="secondary" size="xs" @click="fetchPOs">
          <RefreshCw class="w-3.5 h-3.5 mr-1" :class="{ 'animate-spin': loading }" />
          Segarkan PO
        </BaseButton>
      </div>

      <!-- PO Table -->
      <div class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs text-slate-600">
            <thead class="bg-slate-50 text-slate-700 font-semibold border-b border-slate-200 uppercase tracking-wider">
              <tr>
                <th class="py-2.5 px-4">No. PO</th>
                <th class="py-2.5 px-4">Supplier Pengadaan</th>
                <th class="py-2.5 px-4">Tgl Order</th>
                <th class="py-2.5 px-4">Estimasi Tiba</th>
                <th class="py-2.5 px-4 text-right">Total Nilai PO</th>
                <th class="py-2.5 px-4">Catatan</th>
                <th class="py-2.5 px-4 text-center">Status</th>
                <th class="py-2.5 px-4 text-center">Aksi</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-if="loading">
                <td colspan="8" class="text-center py-10 text-slate-400">
                  <RefreshCw class="w-6 h-6 animate-spin mx-auto mb-2 text-sky-500" />
                  Memuat daftar Purchase Order...
                </td>
              </tr>
              <tr v-else-if="pos.length === 0">
                <td colspan="8" class="text-center py-12 text-slate-400">
                  <ShoppingCart class="w-10 h-10 mx-auto mb-2 text-slate-300" />
                  Belum ada Purchase Order pengadaan.
                </td>
              </tr>
              <tr
                v-for="po in pos"
                :key="po.id"
                class="hover:bg-slate-50 transition-colors"
              >
                <td class="py-2.5 px-4 font-mono font-bold text-sky-800">
                  {{ po.po_number }}
                </td>

                <td class="py-2.5 px-4 font-semibold text-slate-800">
                  {{ po.supplier_name }}
                </td>

                <td class="py-2.5 px-4 tabular-nums">
                  {{ formatDate(po.order_date) }}
                </td>

                <td class="py-2.5 px-4 tabular-nums">
                  {{ formatDate(po.expected_date) }}
                </td>

                <td class="py-2.5 px-4 text-right font-bold text-slate-800 tabular-nums">
                  {{ formatCurrency(po.total_amount) }}
                </td>

                <td class="py-2.5 px-4 max-w-[180px] truncate text-slate-500">
                  {{ po.catatan || '-' }}
                </td>

                <td class="py-2.5 px-4 text-center">
                  <StatusBadge :status="po.status" />
                </td>

                <td class="py-2.5 px-4 text-center">
                  <BaseButton
                    v-if="po.status === 'SUBMITTED' || po.status === 'DRAFT'"
                    variant="primary"
                    size="xs"
                    @click="confirmReceivePo(po)"
                  >
                    <CheckCircle2 class="w-3.5 h-3.5 mr-1" />
                    Terima Barang
                  </BaseButton>
                  <span v-else class="text-[11px] text-slate-400 font-medium">
                    Telah Diterima
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- TAB 3: MUTASI & KARTU STOK -->
    <div v-else-if="activeTab === 'mutasi'" class="space-y-3">
      <!-- Filter Bar -->
      <div class="bg-white p-3 rounded-xl border border-slate-200 shadow-sm flex items-center justify-between">
        <div class="text-xs font-semibold text-slate-700 flex items-center gap-2">
          <Clock class="w-4 h-4 text-sky-600" />
          Riwayat Transaksi Keluar / Masuk Kartu Stok
        </div>
        <BaseButton variant="secondary" size="xs" @click="fetchMutations">
          <RefreshCw class="w-3.5 h-3.5 mr-1" :class="{ 'animate-spin': loading }" />
          Segarkan Riwayat
        </BaseButton>
      </div>

      <!-- Mutations Table -->
      <div class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs text-slate-600">
            <thead class="bg-slate-50 text-slate-700 font-semibold border-b border-slate-200 uppercase tracking-wider">
              <tr>
                <th class="py-2.5 px-4">Waktu Transaksi</th>
                <th class="py-2.5 px-4">Nama Barang</th>
                <th class="py-2.5 px-4 text-center">Jenis Mutasi</th>
                <th class="py-2.5 px-4 text-center">Jumlah (Qty)</th>
                <th class="py-2.5 px-4">Batch & Expiry</th>
                <th class="py-2.5 px-4">Alur Perpindahan</th>
                <th class="py-2.5 px-4">Keterangan</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-if="loading">
                <td colspan="7" class="text-center py-10 text-slate-400">
                  <RefreshCw class="w-6 h-6 animate-spin mx-auto mb-2 text-sky-500" />
                  Memuat kartu mutasi stok...
                </td>
              </tr>
              <tr v-else-if="mutations.length === 0">
                <td colspan="7" class="text-center py-12 text-slate-400">
                  <ArrowRightLeft class="w-10 h-10 mx-auto mb-2 text-slate-300" />
                  Belum ada mutasi stok yang tercatat.
                </td>
              </tr>
              <tr
                v-for="mut in mutations"
                :key="mut.id"
                class="hover:bg-slate-50 transition-colors"
              >
                <td class="py-2.5 px-4 tabular-nums text-slate-500">
                  {{ formatDate(mut.created_at) }}
                </td>

                <td class="py-2.5 px-4 font-semibold text-slate-800">
                  {{ mut.item_name || 'Barang ID: ' + mut.item_id }}
                </td>

                <td class="py-2.5 px-4 text-center">
                  <span
                    class="px-2 py-0.5 rounded text-[10px] font-bold"
                    :class="{
                      'bg-emerald-100 text-emerald-800': mut.transaction_type === 'IN_PO',
                      'bg-rose-100 text-rose-800': mut.transaction_type === 'OUT_DISPENSE',
                      'bg-sky-100 text-sky-800': mut.transaction_type === 'MUTASI_DEPO',
                      'bg-amber-100 text-amber-800': mut.transaction_type === 'ADJUSTMENT',
                    }"
                  >
                    {{ mut.transaction_type }}
                  </span>
                </td>

                <td class="py-2.5 px-4 text-center font-mono font-bold tabular-nums"
                  :class="mut.transaction_type === 'OUT_DISPENSE' ? 'text-rose-600' : 'text-emerald-600'"
                >
                  {{ mut.transaction_type === 'OUT_DISPENSE' ? '-' : '+' }}{{ mut.qty }}
                </td>

                <td class="py-2.5 px-4">
                  <div class="font-mono text-slate-700">{{ mut.batch_number || '-' }}</div>
                  <div class="text-[10px] text-slate-400">Exp: {{ formatDate(mut.expiry_date) }}</div>
                </td>

                <td class="py-2.5 px-4">
                  <div class="flex items-center gap-1.5 font-medium text-slate-700">
                    <span>{{ mut.source_depo }}</span>
                    <span class="text-slate-400">→</span>
                    <span class="text-sky-700">{{ mut.dest_depo }}</span>
                  </div>
                </td>

                <td class="py-2.5 px-4 text-slate-500">
                  {{ mut.notes || '-' }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- MODAL: Tambah / Edit Barang -->
    <BaseModal
      v-model="showItemModal"
      :title="isEditingItem ? 'Edit Data Barang Medis' : 'Tambah Barang Medis / BHP Baru'"
      width="lg"
    >
      <div class="space-y-3">
        <div class="grid grid-cols-2 gap-3">
          <BaseInput
            v-model="itemForm.item_code"
            label="Kode Barang"
            placeholder="OBT-001"
            required
          />
          <BaseInput
            v-model="itemForm.rack_location"
            label="Lokasi Rak Gudang"
            placeholder="Rak A-02"
          />
        </div>

        <BaseInput
          v-model="itemForm.item_name"
          label="Nama Barang / Obat Dagang"
          placeholder="Paracetamol Sirup 120mg/5ml"
          required
        />

        <BaseInput
          v-model="itemForm.generic_name"
          label="Nama Generik / Kandungan Zat Aktif"
          placeholder="Paracetamol"
        />

        <div class="grid grid-cols-2 gap-3">
          <BaseSelect
            v-model="itemForm.category_id"
            label="Kategori Barang"
            :options="categories.map(c => ({ value: c.id, label: c.name }))"
            required
          />

          <BaseSelect
            v-model="itemForm.unit"
            label="Satuan Kemasan"
            :options="unitOptions"
            required
          />
        </div>

        <div class="grid grid-cols-3 gap-3">
          <BaseInput
            v-model.number="itemForm.buy_price"
            type="number"
            label="Harga Beli (Rp)"
            required
          />
          <BaseInput
            v-model.number="itemForm.sell_price"
            type="number"
            label="Harga Jual (Rp)"
            required
          />
          <BaseInput
            v-model.number="itemForm.minimum_stock"
            type="number"
            label="Batas Min. Stok"
            required
          />
        </div>

        <div class="flex items-center gap-4 pt-2">
          <label class="flex items-center gap-1.5 text-xs text-slate-700 cursor-pointer">
            <input
              v-model="itemForm.is_medicine"
              type="checkbox"
              class="rounded text-sky-600 focus:ring-sky-500 h-4 w-4"
            />
            Termasuk Golongan Obat Medis (Resep)
          </label>

          <label class="flex items-center gap-1.5 text-xs text-slate-700 cursor-pointer">
            <input
              v-model="itemForm.status_aktif"
              type="checkbox"
              class="rounded text-sky-600 focus:ring-sky-500 h-4 w-4"
            />
            Status Aktif Digunakan
          </label>
        </div>

        <div class="flex items-center justify-end gap-2 pt-3 border-t border-slate-100">
          <BaseButton variant="ghost" size="sm" @click="showItemModal = false">
            Batal
          </BaseButton>
          <BaseButton
            variant="primary"
            size="sm"
            :loading="submittingItem"
            @click="saveItem"
          >
            <CheckCircle2 class="w-4 h-4 mr-1.5" />
            {{ isEditingItem ? 'Simpan Perubahan' : 'Tambah Barang' }}
          </BaseButton>
        </div>
      </div>
    </BaseModal>

    <!-- MODAL: Buat Purchase Order -->
    <BaseModal
      v-model="showCreatePoModal"
      title="Penerbitan Purchase Order (PO) Pengadaan"
      width="xl"
    >
      <div class="space-y-4">
        <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
          <div>
            <label class="block text-xs font-semibold text-slate-700 mb-1">
              Nama Supplier / PBF <span class="text-red-500">*</span>
            </label>
            <SearchableSelect
              v-model="poForm.supplier_id"
              :options="suppliers.map(s => ({
                value: s.id,
                label: s.name,
                badge: s.code,
                subtitle: `${s.city ? s.city + ' • ' : ''}${s.phone || 'PBF Resmi'}`
              }))"
              placeholder="-- Pilih Master Supplier / PBF --"
              class="w-full"
              @update:model-value="onSupplierSelect"
            />
          </div>
          <BaseDatePicker
            v-model="poForm.order_date"
            label="Tanggal Pemesanan"
            required
          />
          <BaseDatePicker
            v-model="poForm.expected_date"
            label="Estimasi Kedatangan"
          />
        </div>

        <BaseInput
          v-model="poForm.catatan"
          label="Catatan Pengadaan"
          placeholder="Instruksi pengiriman batch exp date minimal 2 tahun..."
        />

        <!-- Items Table in PO -->
        <div class="border border-slate-200 rounded-lg overflow-visible">
          <div class="bg-slate-50 px-3 py-2 border-b border-slate-200 flex items-center justify-between rounded-t-lg">
            <span class="text-xs font-bold text-slate-700">Daftar Barang yang Dipesan</span>
            <BaseButton variant="secondary" size="xs" @click="addPoItemRow">
              <Plus class="w-3.5 h-3.5 mr-1" />
              Tambah Baris
            </BaseButton>
          </div>

          <div class="p-3 space-y-2 overflow-visible">
            <div
              v-for="(row, idx) in poForm.items"
              :key="idx"
              class="grid grid-cols-12 gap-2 items-center bg-slate-50/60 p-2 rounded border border-slate-200/80 relative"
              :style="{ zIndex: poForm.items.length - idx }"
            >
              <div class="col-span-5">
                <SearchableSelect
                  v-model="row.item_id"
                  :options="items.map(it => ({
                    value: it.id,
                    label: it.item_name,
                    subtitle: `Stok: ${it.total_stock} ${it.unit}`,
                    stock: it.total_stock,
                    unit: it.unit,
                    price: it.buy_price,
                    badge: it.category_name
                  }))"
                  placeholder="-- Pilih Barang --"
                  compact
                  @change="() => onPoItemChange(row)"
                />
              </div>

              <div class="col-span-2">
                <input
                  v-model.number="row.qty"
                  type="number"
                  min="1"
                  placeholder="Qty"
                  class="w-full px-2 py-1.5 text-xs border border-slate-200 rounded focus:outline-none text-center"
                />
              </div>

              <div class="col-span-2">
                <input
                  v-model.number="row.unit_price"
                  type="number"
                  placeholder="Harga"
                  class="w-full px-2 py-1.5 text-xs border border-slate-200 rounded focus:outline-none text-right tabular-nums"
                />
              </div>

              <div class="col-span-2 text-right text-xs font-bold text-slate-700 tabular-nums">
                {{ formatCurrency(row.qty * row.unit_price) }}
              </div>

              <div class="col-span-1 text-center">
                <button
                  class="text-rose-500 hover:text-rose-700 p-1"
                  title="Hapus baris"
                  @click="removePoItemRow(idx)"
                >
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
          </div>

          <div class="bg-slate-100 p-3 flex justify-between items-center text-xs font-bold text-slate-800">
            <span>Total Nilai Pengadaan:</span>
            <span class="text-sm font-bold text-sky-700 tabular-nums">{{ formatCurrency(poTotalAmount) }}</span>
          </div>
        </div>

        <div class="flex items-center justify-end gap-2 pt-2 border-t border-slate-100">
          <BaseButton variant="ghost" size="sm" @click="showCreatePoModal = false">
            Batal
          </BaseButton>
          <BaseButton
            variant="primary"
            size="sm"
            :loading="submittingPo"
            @click="savePo"
          >
            <CheckCircle2 class="w-4 h-4 mr-1.5" />
            Terbitkan Purchase Order
          </BaseButton>
        </div>
      </div>
    </BaseModal>

    <!-- MODAL: Mutasi / Penyesuaian Stok -->
    <BaseModal
      v-model="showMutationModal"
      title="Catat Mutasi / Koreksi Stok Gudang"
      width="md"
    >
      <div class="space-y-3">
        <SearchableSelect
          v-model="mutationForm.item_id"
          label="Pilih Barang Medis"
          :options="items.map(it => ({
            value: it.id,
            label: it.item_name,
            subtitle: `Gudang: ${it.total_stock} • Depo: ${it.depo_stock} ${it.unit}`,
            stock: it.total_stock,
            unit: it.unit,
            badge: it.category_name
          }))"
          placeholder="-- Cari Barang Medis --"
          required
        />

        <div class="grid grid-cols-2 gap-3">
          <BaseSelect
            v-model="mutationForm.transaction_type"
            label="Jenis Transaksi"
            :options="[
              { value: 'MUTASI_DEPO', label: 'Mutasi Gudang → Depo Farmasi' },
              { value: 'ADJUSTMENT', label: 'Koreksi Fisik (Stock Opname)' },
            ]"
            required
          />

          <BaseInput
            v-model.number="mutationForm.qty"
            type="number"
            label="Jumlah (Qty)"
            required
          />
        </div>

        <div class="grid grid-cols-2 gap-3">
          <BaseInput
            v-model="mutationForm.batch_number"
            label="Nomor Batch"
            placeholder="BATCH-1092"
          />
          <BaseDatePicker
            v-model="mutationForm.expiry_date"
            label="Tanggal Kadaluarsa (Exp)"
          />
        </div>

        <div class="grid grid-cols-2 gap-3">
          <BaseInput
            v-model="mutationForm.source_depo"
            label="Lokasi Asal"
            placeholder="GUDANG_UTAMA"
          />
          <BaseInput
            v-model="mutationForm.dest_depo"
            label="Lokasi Tujuan"
            placeholder="DEPO_FARMASI"
          />
        </div>

        <BaseInput
          v-model="mutationForm.notes"
          label="Keterangan / Catatan Alasan"
          placeholder="Mutasi stok mingguan atau koreksi selisih fisik..."
        />

        <div class="flex items-center justify-end gap-2 pt-3 border-t border-slate-100">
          <BaseButton variant="ghost" size="sm" @click="showMutationModal = false">
            Batal
          </BaseButton>
          <BaseButton
            variant="primary"
            size="sm"
            :loading="submittingMutation"
            @click="saveMutation"
          >
            <CheckCircle2 class="w-4 h-4 mr-1.5" />
            Simpan Mutasi
          </BaseButton>
        </div>
      </div>
    </BaseModal>

    <!-- MODAL: Buat Purchase Requisition (PR) -->
    <BaseModal
      v-model="showCreatePRModal"
      title="Penerbitan Permintaan Pengadaan (Purchase Requisition)"
      width="lg"
    >
      <div class="space-y-4">
        <div class="grid grid-cols-2 gap-3">
          <BaseSelect
            v-model="prForm.department"
            label="Departemen Pemohon"
            :options="[
              { value: 'FARMASI', label: 'Instalasi Farmasi' },
              { value: 'RAWAT_JALAN', label: 'Poliklinik Rawat Jalan' },
              { value: 'LABORATORIUM', label: 'Laboratorium Medis' },
              { value: 'RADIOLOGI', label: 'Radiologi' }
            ]"
            required
          />
          <BaseInput
            v-model="prForm.notes"
            label="Catatan Keperluan"
            placeholder="Pengadaan rutin safety stock..."
          />
        </div>

        <div class="border border-slate-200 rounded-lg overflow-visible">
          <div class="bg-slate-50 px-3 py-2 border-b border-slate-200 flex items-center justify-between rounded-t-lg">
            <span class="text-xs font-bold text-slate-700">Daftar Barang yang Diminta</span>
            <BaseButton variant="secondary" size="xs" @click="addPRItemRow">
              <Plus class="w-3.5 h-3.5 mr-1" />
              Tambah Baris
            </BaseButton>
          </div>

          <div class="p-3 space-y-2 overflow-visible">
            <div
              v-for="(row, idx) in prForm.items"
              :key="idx"
              class="grid grid-cols-12 gap-2 items-center bg-slate-50/60 p-2 rounded border border-slate-200/80 relative"
              :style="{ zIndex: prForm.items.length - idx }"
            >
              <div class="col-span-6">
                <SearchableSelect
                  v-model="row.item_id"
                  :options="items.map(it => ({
                    value: it.id,
                    label: it.item_name,
                    subtitle: `Stok: ${it.total_stock} ${it.unit}`,
                    stock: it.total_stock,
                    unit: it.unit,
                    price: it.buy_price,
                    badge: it.category_name
                  }))"
                  placeholder="-- Pilih Barang --"
                  compact
                  @change="() => {
                    const sel = items.find(i => i.id === Number(row.item_id))
                    if (sel) row.unit_price = sel.buy_price
                  }"
                />
              </div>

              <div class="col-span-3">
                <input
                  v-model.number="row.qty"
                  type="number"
                  min="1"
                  placeholder="Qty"
                  class="w-full px-2 py-1.5 text-xs border border-slate-200 rounded focus:outline-none text-center"
                />
              </div>

              <div class="col-span-2 text-right text-xs font-bold text-slate-700 tabular-nums">
                {{ formatCurrency(row.qty * row.unit_price) }}
              </div>

              <div class="col-span-1 text-center">
                <button
                  class="text-rose-500 hover:text-rose-700 p-1"
                  title="Hapus baris"
                  @click="removePRItemRow(idx)"
                >
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
          </div>
        </div>

        <div class="flex items-center justify-end gap-2 pt-2 border-t border-slate-100">
          <BaseButton variant="ghost" size="sm" @click="showCreatePRModal = false">
            Batal
          </BaseButton>
          <BaseButton
            variant="primary"
            size="sm"
            :loading="submittingPR"
            @click="submitPR"
          >
            <CheckCircle2 class="w-4 h-4 mr-1.5" />
            Terbitkan PR
          </BaseButton>
        </div>
      </div>
    </BaseModal>

    <!-- MODAL: Penerimaan Barang Fisik (GRN) -->
    <BaseModal
      v-model="showCreateGRNModal"
      title="Penerimaan Barang Fisik dari Supplier (GRN)"
      width="md"
    >
      <div class="space-y-3">
        <BaseSelect
          v-model="grnForm.po_id"
          label="Pilih Purchase Order (PO) yang Diterima"
          :options="pos.map(p => ({
            value: p.id,
            label: `${p.po_number} - ${p.supplier_name} (${formatCurrency(p.total_amount)})`,
            badge: p.status
          }))"
          required
        />

        <BaseInput
          v-model="grnForm.delivery_order_number"
          label="Nomor Surat Jalan / Delivery Order (DO)"
          placeholder="DO/2026/0912"
          required
        />

        <BaseInput
          v-model="grnForm.receiver_name"
          label="Nama Petugas Penerima Gudang"
          required
        />

        <BaseInput
          v-model="grnForm.notes"
          label="Catatan Fisik Penerimaan"
          placeholder="Barang diterima lengkap dalam kondisi tersegel baik..."
        />

        <div class="p-2.5 rounded bg-emerald-50 border border-emerald-200 text-emerald-800 text-[11px]">
          📦 Penerimaan GRN ini akan otomatis menambahkan stok fisik ke Gudang Utama dan mencatat riwayat kartu stok mutasi.
        </div>

        <div class="flex items-center justify-end gap-2 pt-3 border-t border-slate-100">
          <BaseButton variant="ghost" size="sm" @click="showCreateGRNModal = false">
            Batal
          </BaseButton>
          <BaseButton
            variant="primary"
            size="sm"
            class="!bg-emerald-600 hover:!bg-emerald-700"
            :loading="submittingGRN"
            @click="submitGRN"
          >
            <CheckCircle2 class="w-4 h-4 mr-1.5" />
            Konfirmasi Terima & Tambah Stok
          </BaseButton>
        </div>
      </div>
    </BaseModal>

    <!-- CONFIRM MODAL: Receive PO -->
    <ConfirmModal
      v-model="showReceivePoConfirm"
      title="Verifikasi Penerimaan Barang (GRN)"
      :message="`Konfirmasi penerimaan barang untuk PO ${poToReceive?.po_number} dari supplier ${poToReceive?.supplier_name}? Seluruh stok barang akan otomatis ditambahkan ke Gudang Utama.`"
      confirm-text="Terima & Masukkan ke Stok"
      confirm-variant="primary"
      @confirm="executeReceivePo"
    />
  </div>
</template>
