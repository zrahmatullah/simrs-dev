<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api/client'
import DataTable, { type Column } from '@/components/common/DataTable.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseDatePicker from '@/components/common/BaseDatePicker.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseModal from '@/components/common/BaseModal.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { Plus, Tag, Truck, Building2, Stethoscope, Hospital } from 'lucide-vue-next'
import type { FaskesRujukan, MasterSpesialis, PoliUnit } from '@/types'

const activeTab = ref<'poli' | 'guarantor' | 'icd10' | 'icd9' | 'categories' | 'promos' | 'suppliers' | 'faskes' | 'spesialis'>('poli')
const isLoading = ref(false)
const search = ref('')

const poliList = ref<PoliUnit[]>([])
const guarantors = ref<any[]>([])
const icd10List = ref<any[]>([])
const icd9List = ref<any[]>([])
const categories = ref<any[]>([])
const promoList = ref<any[]>([])
const supplierList = ref<any[]>([])
const faskesList = ref<FaskesRujukan[]>([])
const spesialisList = ref<MasterSpesialis[]>([])

// Create Poli Modal
const isCreatePoliOpen = ref(false)
const isSubmittingPoli = ref(false)
const poliForm = ref({
  code: '',
  name: '',
  description: ''
})

// Create Faskes Modal
const isCreateFaskesOpen = ref(false)
const isSubmittingFaskes = ref(false)
const faskesForm = ref({
  code: '',
  name: '',
  tipe: 'RSUD',
  alamat: '',
  kota: '',
  no_telp: ''
})

// Create Spesialis Modal
const isCreateSpesialisOpen = ref(false)
const isSubmittingSpesialis = ref(false)
const spesialisForm = ref({
  code: '',
  nama_dokter: '',
  spesialisasi: '',
  rumah_sakit: ''
})

// Create Supplier Modal
const isCreateSupplierOpen = ref(false)
const isSubmittingSupplier = ref(false)
const supplierForm = ref({
  code: '',
  name: '',
  pic_name: '',
  phone: '',
  email: '',
  address: '',
  city: '',
})

const supplierCols: Column[] = [
  { key: 'code', label: 'Kode PBF', width: '130px' },
  { key: 'name', label: 'Nama Distributor / Supplier' },
  { key: 'pic_name', label: 'Nama PIC / Apoteker', width: '180px' },
  { key: 'phone', label: 'No. Telepon', width: '140px' },
  { key: 'city', label: 'Kota', width: '130px' },
  { key: 'address', label: 'Alamat Kantor / Gudang' },
  { key: 'status_aktif', label: 'Status', align: 'center', width: '110px' },
]

// Create Promo Modal
const isCreatePromoOpen = ref(false)
const isSubmittingPromo = ref(false)
const promoForm = ref({
  code: '',
  name: '',
  discount_type: 'PERCENTAGE',
  discount_value: 10,
  min_transaction: 150000,
  start_date: new Date().toISOString().split('T')[0],
  end_date: new Date(Date.now() + 180 * 86400000).toISOString().split('T')[0],
  status_aktif: true
})

const promoCols: Column[] = [
  { key: 'code', label: 'Kode Promo', width: '150px' },
  { key: 'name', label: 'Nama Program Promo' },
  { key: 'discount', label: 'Besaran Diskon', width: '150px' },
  { key: 'min_transaction', label: 'Min. Transaksi', align: 'right', width: '150px' },
  { key: 'validity', label: 'Periode Berlaku', width: '200px' },
  { key: 'status_aktif', label: 'Status', align: 'center', width: '120px' },
]

const poliCols: Column[] = [
  { key: 'code', label: 'Kode Poli', width: '130px' },
  { key: 'name', label: 'Nama Unit Pelayanan' },
  { key: 'description', label: 'Keterangan Layanan' },
  { key: 'status_aktif', label: 'Status', align: 'center', width: '120px' },
]

const guarantorCols: Column[] = [
  { key: 'code', label: 'Kode Penjamin', width: '140px' },
  { key: 'name', label: 'Nama Penjamin / Asuransi' },
  { key: 'category', label: 'Kategori', width: '130px' },
  { key: 'status_aktif', label: 'Status', align: 'center', width: '120px' },
]

const icd10Cols: Column[] = [
  { key: 'code', label: 'Kode ICD-10', width: '130px' },
  { key: 'name_id', label: 'Diagnosis (Bahasa Indonesia)' },
  { key: 'name_en', label: 'Diagnosis (English)' },
]

const icd9Cols: Column[] = [
  { key: 'code', label: 'Kode ICD-9', width: '130px' },
  { key: 'name_id', label: 'Nama Prosedur / Tindakan Medis' },
  { key: 'tariff', label: 'Tarif Standar (Rp)', align: 'right', width: '160px' },
]

const categoryCols: Column[] = [
  { key: 'id', label: 'ID', width: '80px' },
  { key: 'name', label: 'Kategori Obat & BHP' },
  { key: 'description', label: 'Deskripsi Kategori' },
]

const faskesCols: Column[] = [
  { key: 'code', label: 'Kode Faskes', width: '130px' },
  { key: 'name', label: 'Nama Rumah Sakit / Faskes Rujukan' },
  { key: 'tipe', label: 'Tipe', width: '120px' },
  { key: 'kota', label: 'Kota / Wilayah', width: '150px' },
  { key: 'no_telp', label: 'No. Hotline / Kontak', width: '150px' },
  { key: 'alamat', label: 'Alamat Faskes' },
  { key: 'status_aktif', label: 'Status', align: 'center', width: '110px' },
]

const spesialisCols: Column[] = [
  { key: 'code', label: 'Kode', width: '130px' },
  { key: 'nama_dokter', label: 'Nama Dokter Spesialis' },
  { key: 'spesialisasi', label: 'Bidang Spesialisasi', width: '260px' },
  { key: 'rumah_sakit', label: 'Rumah Sakit Praktek / Induk', width: '220px' },
  { key: 'status_aktif', label: 'Status', align: 'center', width: '110px' },
]

async function loadData() {
  isLoading.value = true
  try {
    if (activeTab.value === 'poli') {
      const res = await api.get<any[]>('/master/poli')
      if (res.data) poliList.value = res.data
    } else if (activeTab.value === 'guarantor') {
      const res = await api.get<any[]>('/master/guarantors')
      if (res.data) guarantors.value = res.data
    } else if (activeTab.value === 'icd10') {
      const res = await api.get<any[]>('/master/icd10', { search: search.value })
      if (res.data) icd10List.value = res.data
    } else if (activeTab.value === 'icd9') {
      const res = await api.get<any[]>('/master/icd9', { search: search.value })
      if (res.data) icd9List.value = res.data
    } else if (activeTab.value === 'categories') {
      const res = await api.get<any[]>('/master/categories')
      if (res.data) categories.value = res.data
    } else if (activeTab.value === 'promos') {
      const res = await api.get<any[]>('/master/promos')
      if (res.data) promoList.value = res.data
    } else if (activeTab.value === 'suppliers') {
      const res = await api.get<any[]>('/master/suppliers', { search: search.value })
      if (res.data) supplierList.value = res.data
    } else if (activeTab.value === 'faskes') {
      const res = await api.get<FaskesRujukan[]>('/master/faskes', { search: search.value })
      if (res.data) faskesList.value = res.data
    } else if (activeTab.value === 'spesialis') {
      const res = await api.get<MasterSpesialis[]>('/master/spesialis', { search: search.value })
      if (res.data) spesialisList.value = res.data
    }
  } catch (err) {
    console.error('Failed to load master data:', err)
  } finally {
    isLoading.value = false
  }
}

async function createPoli() {
  if (!poliForm.value.code || !poliForm.value.name) return
  isSubmittingPoli.value = true
  try {
    await api.post('/master/poli', poliForm.value)
    isCreatePoliOpen.value = false
    poliForm.value = { code: '', name: '', description: '' }
    await loadData()
  } catch (err) {
    console.error('Failed to create poli:', err)
  } finally {
    isSubmittingPoli.value = false
  }
}

async function createFaskes() {
  if (!faskesForm.value.code || !faskesForm.value.name) return
  isSubmittingFaskes.value = true
  try {
    await api.post('/master/faskes', faskesForm.value)
    isCreateFaskesOpen.value = false
    faskesForm.value = { code: '', name: '', tipe: 'RSUD', alamat: '', kota: '', no_telp: '' }
    await loadData()
  } catch (err) {
    console.error('Failed to create faskes:', err)
  } finally {
    isSubmittingFaskes.value = false
  }
}

async function createSpesialis() {
  if (!spesialisForm.value.code || !spesialisForm.value.nama_dokter || !spesialisForm.value.spesialisasi) return
  isSubmittingSpesialis.value = true
  try {
    await api.post('/master/spesialis', spesialisForm.value)
    isCreateSpesialisOpen.value = false
    spesialisForm.value = { code: '', nama_dokter: '', spesialisasi: '', rumah_sakit: '' }
    await loadData()
  } catch (err) {
    console.error('Failed to create spesialis:', err)
  } finally {
    isSubmittingSpesialis.value = false
  }
}

async function createSupplier() {
  if (!supplierForm.value.code || !supplierForm.value.name) return
  isSubmittingSupplier.value = true
  try {
    await api.post('/master/suppliers', supplierForm.value)
    isCreateSupplierOpen.value = false
    supplierForm.value = {
      code: '',
      name: '',
      pic_name: '',
      phone: '',
      email: '',
      address: '',
      city: '',
    }
    await loadData()
  } catch (err) {
    console.error('Failed to create supplier:', err)
  } finally {
    isSubmittingSupplier.value = false
  }
}

async function createPromo() {
  if (!promoForm.value.code || !promoForm.value.name) return
  isSubmittingPromo.value = true
  try {
    await api.post('/master/promos', promoForm.value)
    isCreatePromoOpen.value = false
    promoForm.value.code = ''
    promoForm.value.name = ''
    loadData()
  } catch (err) {
    console.error('Failed to create promo:', err)
  } finally {
    isSubmittingPromo.value = false
  }
}

function setTab(tab: typeof activeTab.value) {
  activeTab.value = tab
  search.value = ''
  loadData()
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="space-y-4 text-left">
    <!-- Header -->
    <div>
      <h1 class="text-[20px] font-bold text-[#0F172A] leading-[28px]">
        Master Data & Pengaturan Tarif Klinis
      </h1>
      <p class="text-[13px] text-[#475569]">
        Konfigurasi master referensi internal, poli klinik, penjamin BPJS/Asuransi, kode diagnosa ICD-10, dan tarif prosedur ICD-9.
      </p>
    </div>

    <!-- Tab Buttons -->
    <div class="flex items-center gap-1 border-b border-[#E2E8F0]">
      <button
        type="button"
        :class="[
          'px-4 py-2 text-[13px] font-semibold border-b-2 transition-all',
          activeTab === 'poli'
            ? 'border-[#0284C7] text-[#0284C7] bg-[#F0F9FF]'
            : 'border-transparent text-[#64748B] hover:text-[#0F172A]'
        ]"
        @click="setTab('poli')"
      >
        Unit Pelayanan & Poli
      </button>

      <button
        type="button"
        :class="[
          'px-4 py-2 text-[13px] font-semibold border-b-2 transition-all',
          activeTab === 'guarantor'
            ? 'border-[#0284C7] text-[#0284C7] bg-[#F0F9FF]'
            : 'border-transparent text-[#64748B] hover:text-[#0F172A]'
        ]"
        @click="setTab('guarantor')"
      >
        Penjamin & Asuransi
      </button>

      <button
        type="button"
        :class="[
          'px-4 py-2 text-[13px] font-semibold border-b-2 transition-all',
          activeTab === 'icd10'
            ? 'border-[#0284C7] text-[#0284C7] bg-[#F0F9FF]'
            : 'border-transparent text-[#64748B] hover:text-[#0F172A]'
        ]"
        @click="setTab('icd10')"
      >
        ICD-10 (Diagnosis)
      </button>

      <button
        type="button"
        :class="[
          'px-4 py-2 text-[13px] font-semibold border-b-2 transition-all',
          activeTab === 'icd9'
            ? 'border-[#0284C7] text-[#0284C7] bg-[#F0F9FF]'
            : 'border-transparent text-[#64748B] hover:text-[#0F172A]'
        ]"
        @click="setTab('icd9')"
      >
        ICD-9-CM & Tarif Tindakan
      </button>

      <button
        type="button"
        :class="[
          'px-4 py-2 text-[13px] font-semibold border-b-2 transition-all',
          activeTab === 'categories'
            ? 'border-[#0284C7] text-[#0284C7] bg-[#F0F9FF]'
            : 'border-transparent text-[#64748B] hover:text-[#0F172A]'
        ]"
        @click="setTab('categories')"
      >
        Kategori Obat & BHP
      </button>

      <button
        type="button"
        :class="[
          'px-4 py-2 text-[13px] font-semibold border-b-2 transition-all flex items-center gap-1.5',
          activeTab === 'promos'
            ? 'border-[#0284C7] text-[#0284C7] bg-[#F0F9FF]'
            : 'border-transparent text-[#64748B] hover:text-[#0F172A]'
        ]"
        @click="setTab('promos')"
      >
        <Tag class="w-3.5 h-3.5" />
        Master Promo & Diskon
      </button>

      <button
        type="button"
        :class="[
          'px-4 py-2 text-[13px] font-semibold border-b-2 transition-all flex items-center gap-1.5',
          activeTab === 'suppliers'
            ? 'border-[#0284C7] text-[#0284C7] bg-[#F0F9FF]'
            : 'border-transparent text-[#64748B] hover:text-[#0F172A]'
        ]"
        @click="setTab('suppliers')"
      >
        <Truck class="w-3.5 h-3.5" />
        Master Supplier (PBF)
      </button>

      <button
        type="button"
        :class="[
          'px-4 py-2 text-[13px] font-semibold border-b-2 transition-all flex items-center gap-1.5',
          activeTab === 'faskes'
            ? 'border-[#0284C7] text-[#0284C7] bg-[#F0F9FF]'
            : 'border-transparent text-[#64748B] hover:text-[#0F172A]'
        ]"
        @click="setTab('faskes')"
      >
        <Hospital class="w-3.5 h-3.5" />
        Faskes & RS Rujukan
      </button>

      <button
        type="button"
        :class="[
          'px-4 py-2 text-[13px] font-semibold border-b-2 transition-all flex items-center gap-1.5',
          activeTab === 'spesialis'
            ? 'border-[#0284C7] text-[#0284C7] bg-[#F0F9FF]'
            : 'border-transparent text-[#64748B] hover:text-[#0F172A]'
        ]"
        @click="setTab('spesialis')"
      >
        <Stethoscope class="w-3.5 h-3.5" />
        Dokter Spesialis Rujukan
      </button>
    </div>

    <!-- Search for ICD, Supplier, Faskes, and Spesialis tabs -->
    <div v-if="['icd10', 'icd9', 'suppliers', 'faskes', 'spesialis'].includes(activeTab)" class="max-w-md">
      <BaseInput
        v-model="search"
        :placeholder="activeTab === 'suppliers' ? 'Cari nama supplier, kode, kota, atau PIC...' :
                     activeTab === 'faskes' ? 'Cari nama rumah sakit, kode, kota, atau tipe...' :
                     activeTab === 'spesialis' ? 'Cari nama dokter spesialis atau bidang...' :
                     'Cari kode atau nama tindakan/diagnosa...'"
        compact
        @keyup.enter="loadData"
      />
    </div>

    <!-- Tab 1: Poli Units -->
    <div v-if="activeTab === 'poli'" class="space-y-3">
      <div class="flex items-center justify-between">
        <p class="text-xs text-slate-500">Unit pelayanan klinis, poliklinik rawat jalan, dan instalasi medis klinik.</p>
        <BaseButton variant="primary" size="sm" @click="isCreatePoliOpen = true">
          <Plus class="w-4 h-4 mr-1" />
          Tambah Unit Poli Baru
        </BaseButton>
      </div>

      <DataTable
        :columns="poliCols"
        :data="poliList"
        :loading="isLoading"
      >
        <template #code="{ row }">
          <span class="font-mono font-semibold text-[#0284C7]">{{ row.code }}</span>
        </template>
        <template #status_aktif="{ row }">
          <StatusBadge :status="row.status_aktif ? 'Aktif' : 'Nonaktif'" variant="success" />
        </template>
      </DataTable>
    </div>

    <!-- Tab 2: Guarantors -->
    <DataTable
      v-if="activeTab === 'guarantor'"
      :columns="guarantorCols"
      :data="guarantors"
      :loading="isLoading"
    >
      <template #code="{ row }">
        <span class="font-mono font-semibold text-[#0284C7]">{{ row.code }}</span>
      </template>
      <template #status_aktif="{ row }">
        <StatusBadge :status="row.status_aktif ? 'Aktif' : 'Nonaktif'" variant="success" />
      </template>
    </DataTable>

    <!-- Tab 3: ICD-10 -->
    <DataTable
      v-if="activeTab === 'icd10'"
      :columns="icd10Cols"
      :data="icd10List"
      :loading="isLoading"
    >
      <template #code="{ row }">
        <span class="font-mono font-bold text-[#0369A1] bg-[#E0F2FE] px-2 py-0.5 rounded border border-[#BAE6FD]">
          {{ row.code }}
        </span>
      </template>
    </DataTable>

    <!-- Tab 4: ICD-9 -->
    <DataTable
      v-if="activeTab === 'icd9'"
      :columns="icd9Cols"
      :data="icd9List"
      :loading="isLoading"
    >
      <template #code="{ row }">
        <span class="font-mono font-bold text-[#0369A1] bg-[#E0F2FE] px-2 py-0.5 rounded border border-[#BAE6FD]">
          {{ row.code }}
        </span>
      </template>
      <template #tariff="{ row }">
        <span class="font-mono font-semibold tabular-nums text-[#0F172A]">
          Rp {{ Number(row.tariff).toLocaleString('id-ID') }}
        </span>
      </template>
    </DataTable>

    <!-- Tab 5: Categories -->
    <DataTable
      v-if="activeTab === 'categories'"
      :columns="categoryCols"
      :data="categories"
      :loading="isLoading"
    />

    <!-- Tab 6: Master Promo -->
    <div v-if="activeTab === 'promos'" class="space-y-3">
      <div class="flex items-center justify-between">
        <p class="text-xs text-slate-500">Program diskon, kupon promo, dan voucher potongan biaya rawat jalan.</p>
        <BaseButton variant="primary" size="sm" @click="isCreatePromoOpen = true">
          <Plus class="w-4 h-4 mr-1" />
          Tambah Promo Baru
        </BaseButton>
      </div>

      <DataTable
        :columns="promoCols"
        :data="promoList"
        :loading="isLoading"
      >
        <template #code="{ row }">
          <span class="font-mono font-bold text-[#0284C7] bg-[#F0F9FF] px-2 py-0.5 rounded border border-[#BAE6FD]">
            {{ row.code }}
          </span>
        </template>
        <template #discount="{ row }">
          <span class="font-bold text-emerald-700 font-mono">
            {{ row.discount_type === 'PERCENTAGE' ? `${row.discount_value}%` : `Rp ${Number(row.discount_value).toLocaleString('id-ID')}` }}
          </span>
        </template>
        <template #min_transaction="{ row }">
          <span class="font-mono text-slate-700">
            Rp {{ Number(row.min_transaction).toLocaleString('id-ID') }}
          </span>
        </template>
        <template #validity="{ row }">
          <span class="text-[11px] text-slate-600">
            {{ new Date(row.start_date).toLocaleDateString('id-ID') }} s/d {{ new Date(row.end_date).toLocaleDateString('id-ID') }}
          </span>
        </template>
        <template #status_aktif="{ row }">
          <StatusBadge :status="row.status_aktif ? 'Aktif' : 'Nonaktif'" variant="success" />
        </template>
      </DataTable>
    </div>

    <!-- Tab 7: Master Supplier / PBF -->
    <div v-if="activeTab === 'suppliers'" class="space-y-3">
      <div class="flex items-center justify-between">
        <p class="text-xs text-slate-500">Daftar Pedagang Besar Farmasi (PBF) dan Rekanan Distributor Logistik Medis.</p>
        <BaseButton variant="primary" size="sm" @click="isCreateSupplierOpen = true">
          <Plus class="w-4 h-4 mr-1" />
          Tambah Supplier Baru
        </BaseButton>
      </div>

      <DataTable
        :columns="supplierCols"
        :data="supplierList"
        :loading="isLoading"
      >
        <template #code="{ row }">
          <span class="font-mono font-bold text-[#0284C7] bg-[#F0F9FF] px-2 py-0.5 rounded border border-[#BAE6FD] text-[11px]">
            {{ row.code }}
          </span>
        </template>
        <template #name="{ row }">
          <div class="font-semibold text-slate-800">{{ row.name }}</div>
          <div v-if="row.email" class="text-[11px] text-slate-500">{{ row.email }}</div>
        </template>
        <template #city="{ row }">
          <span class="text-xs font-medium text-slate-700 bg-slate-100 px-2 py-0.5 rounded">{{ row.city || '-' }}</span>
        </template>
        <template #status_aktif="{ row }">
          <StatusBadge :status="row.status_aktif ? 'Aktif' : 'Nonaktif'" variant="success" />
        </template>
      </DataTable>
    </div>

    <!-- Tab 8: Master Faskes & RS Rujukan -->
    <div v-if="activeTab === 'faskes'" class="space-y-3">
      <div class="flex items-center justify-between">
        <p class="text-xs text-slate-500">Master direktori Rumah Sakit Umum Pusat, RS Daerah, dan Fasilitas Kesehatan jejaring rujukan eksternal.</p>
        <BaseButton variant="primary" size="sm" @click="isCreateFaskesOpen = true">
          <Plus class="w-4 h-4 mr-1" />
          Tambah Faskes Rujukan Baru
        </BaseButton>
      </div>

      <DataTable
        :columns="faskesCols"
        :data="faskesList"
        :loading="isLoading"
      >
        <template #code="{ row }">
          <span class="font-mono font-bold text-[#0284C7] bg-[#F0F9FF] px-2 py-0.5 rounded border border-[#BAE6FD] text-[11px]">
            {{ row.code }}
          </span>
        </template>
        <template #name="{ row }">
          <div class="font-semibold text-slate-800">{{ row.name }}</div>
          <div v-if="row.alamat" class="text-[11px] text-slate-500 truncate max-w-xs">{{ row.alamat }}</div>
        </template>
        <template #tipe="{ row }">
          <span class="text-xs font-semibold px-2 py-0.5 rounded bg-blue-50 text-blue-700 border border-blue-200">
            {{ row.tipe }}
          </span>
        </template>
        <template #status_aktif="{ row }">
          <StatusBadge :status="row.status_aktif ? 'Aktif' : 'Nonaktif'" variant="success" />
        </template>
      </DataTable>
    </div>

    <!-- Tab 9: Master Dokter Spesialis Rujukan -->
    <div v-if="activeTab === 'spesialis'" class="space-y-3">
      <div class="flex items-center justify-between">
        <p class="text-xs text-slate-500">Master dokter spesialis konsultan dan subspesialis tujuan rujukan klinis pasien.</p>
        <BaseButton variant="primary" size="sm" @click="isCreateSpesialisOpen = true">
          <Plus class="w-4 h-4 mr-1" />
          Tambah Dokter Spesialis Baru
        </BaseButton>
      </div>

      <DataTable
        :columns="spesialisCols"
        :data="spesialisList"
        :loading="isLoading"
      >
        <template #code="{ row }">
          <span class="font-mono font-bold text-[#0284C7] bg-[#F0F9FF] px-2 py-0.5 rounded border border-[#BAE6FD] text-[11px]">
            {{ row.code }}
          </span>
        </template>
        <template #nama_dokter="{ row }">
          <div class="font-semibold text-slate-800 flex items-center gap-1.5">
            <Stethoscope class="w-3.5 h-3.5 text-[#0284C7]" />
            {{ row.nama_dokter }}
          </div>
        </template>
        <template #spesialisasi="{ row }">
          <span class="text-xs font-medium text-emerald-800 bg-emerald-50 px-2 py-0.5 rounded border border-emerald-200">
            {{ row.spesialisasi }}
          </span>
        </template>
        <template #rumah_sakit="{ row }">
          <span class="text-xs text-slate-700 font-medium">
            {{ row.rumah_sakit || '-' }}
          </span>
        </template>
        <template #status_aktif="{ row }">
          <StatusBadge :status="row.status_aktif ? 'Aktif' : 'Nonaktif'" variant="success" />
        </template>
      </DataTable>
    </div>

    <!-- MODAL: Tambah Promo Baru -->
    <BaseModal
      v-model="isCreatePromoOpen"
      title="Tambah Master Promo / Kupon Baru"
      width="md"
    >
      <div class="space-y-3">
        <BaseInput
          v-model="promoForm.code"
          label="Kode Promo (Huruf Kapital, Unik)"
          placeholder="CONTOH: PROMOSEHAT20"
          required
        />
        <BaseInput
          v-model="promoForm.name"
          label="Nama Program Promo"
          placeholder="Diskon Hari Anak Nasional 20%"
          required
        />
        <div class="grid grid-cols-2 gap-3">
          <BaseSelect
            v-model="promoForm.discount_type"
            label="Tipe Potongan"
            :options="[
              { value: 'PERCENTAGE', label: 'Persentase (%)' },
              { value: 'FIXED', label: 'Nominal Tetap (Rp)' }
            ]"
            required
          />
          <BaseInput
            v-model.number="promoForm.discount_value"
            type="number"
            :label="promoForm.discount_type === 'PERCENTAGE' ? 'Nilai Diskon (%)' : 'Nilai Diskon (Rp)'"
            required
          />
        </div>
        <BaseInput
          v-model.number="promoForm.min_transaction"
          type="number"
          label="Minimal Transaksi Tagihan (Rp)"
          required
        />
        <div class="grid grid-cols-2 gap-3">
          <BaseDatePicker
            v-model="promoForm.start_date"
            label="Tanggal Mulai"
            required
          />
          <BaseDatePicker
            v-model="promoForm.end_date"
            label="Tanggal Berakhir"
            required
          />
        </div>

        <div class="flex items-center justify-end gap-2 pt-3 border-t border-slate-100">
          <BaseButton variant="ghost" size="sm" @click="isCreatePromoOpen = false">
            Batal
          </BaseButton>
          <BaseButton
            variant="primary"
            size="sm"
            :loading="isSubmittingPromo"
            @click="createPromo"
          >
            Simpan Promo
          </BaseButton>
        </div>
      </div>
    </BaseModal>

    <!-- MODAL: Tambah Master Supplier Baru -->
    <BaseModal
      v-model="isCreateSupplierOpen"
      title="Tambah Master Supplier / PBF Baru"
      width="md"
    >
      <div class="space-y-3">
        <div class="grid grid-cols-2 gap-3">
          <BaseInput
            v-model="supplierForm.code"
            label="Kode PBF / Supplier *"
            placeholder="PBF-XXX-01"
            required
          />
          <BaseInput
            v-model="supplierForm.city"
            label="Kota / Lokasi"
            placeholder="Jakarta Pusat"
          />
        </div>
        <BaseInput
          v-model="supplierForm.name"
          label="Nama Distributor / PBF *"
          placeholder="PT. Kimia Farma Trading & Distribution"
          required
        />
        <div class="grid grid-cols-2 gap-3">
          <BaseInput
            v-model="supplierForm.pic_name"
            label="Nama PIC / Kontak"
            placeholder="Budi Santoso, S.Si"
          />
          <BaseInput
            v-model="supplierForm.phone"
            label="No. Telepon / HP"
            placeholder="021-3847709"
          />
        </div>
        <BaseInput
          v-model="supplierForm.email"
          label="Email Supplier"
          placeholder="kftd@distributor.com"
        />
        <BaseInput
          v-model="supplierForm.address"
          label="Alamat Kantor / Gudang"
          placeholder="Jl. Veteran No. 9, Jakarta"
        />

        <div class="flex items-center justify-end gap-2 pt-3 border-t border-slate-100">
          <BaseButton variant="ghost" size="sm" @click="isCreateSupplierOpen = false">
            Batal
          </BaseButton>
          <BaseButton
            variant="primary"
            size="sm"
            :loading="isSubmittingSupplier"
            @click="createSupplier"
          >
            Simpan Supplier
          </BaseButton>
        </div>
      </div>
    </BaseModal>

    <!-- MODAL: Tambah Poli Baru -->
    <BaseModal
      v-model="isCreatePoliOpen"
      title="Tambah Unit Pelayanan & Poli Baru"
      width="md"
    >
      <div class="space-y-3">
        <BaseInput
          v-model="poliForm.code"
          label="Kode Poli *"
          placeholder="POLI-006"
          required
        />
        <BaseInput
          v-model="poliForm.name"
          label="Nama Unit Pelayanan / Poli *"
          placeholder="Poli Tumbuh Kembang & Fisioterapi"
          required
        />
        <BaseInput
          v-model="poliForm.description"
          label="Keterangan / Deskripsi Layanan"
          placeholder="Layanan observasi dan terapi tumbuh kembang anak"
        />

        <div class="flex items-center justify-end gap-2 pt-3 border-t border-slate-100">
          <BaseButton variant="ghost" size="sm" @click="isCreatePoliOpen = false">
            Batal
          </BaseButton>
          <BaseButton
            variant="primary"
            size="sm"
            :loading="isSubmittingPoli"
            @click="createPoli"
          >
            Simpan Unit Poli
          </BaseButton>
        </div>
      </div>
    </BaseModal>

    <!-- MODAL: Tambah Faskes Rujukan Baru -->
    <BaseModal
      v-model="isCreateFaskesOpen"
      title="Tambah Master Faskes / RS Rujukan Baru"
      width="md"
    >
      <div class="space-y-3">
        <div class="grid grid-cols-2 gap-3">
          <BaseInput
            v-model="faskesForm.code"
            label="Kode Faskes *"
            placeholder="RS-008"
            required
          />
          <BaseSelect
            v-model="faskesForm.tipe"
            label="Tipe Fasilitas Kesehatan *"
            :options="[
              { value: 'RSUP', label: 'RS Umum Pusat (RSUP)' },
              { value: 'RSUD', label: 'RS Umum Daerah (RSUD)' },
              { value: 'RS Swasta', label: 'Rumah Sakit Swasta' },
              { value: 'RS Khusus', label: 'Rumah Sakit Khusus' },
              { value: 'Klinik Utama', label: 'Klinik Utama' },
              { value: 'Puskesmas', label: 'Puskesmas' }
            ]"
            required
          />
        </div>
        <BaseInput
          v-model="faskesForm.name"
          label="Nama Fasilitas / Rumah Sakit *"
          placeholder="RSUP Fatmawati"
          required
        />
        <div class="grid grid-cols-2 gap-3">
          <BaseInput
            v-model="faskesForm.kota"
            label="Kota / Kabupaten"
            placeholder="Jakarta Selatan"
          />
          <BaseInput
            v-model="faskesForm.no_telp"
            label="No. Hotline / Telepon"
            placeholder="021-7501524"
          />
        </div>
        <BaseInput
          v-model="faskesForm.alamat"
          label="Alamat Lengkap Faskes"
          placeholder="Jl. RS. Fatmawati Raya No. 4, Cilandak"
        />

        <div class="flex items-center justify-end gap-2 pt-3 border-t border-slate-100">
          <BaseButton variant="ghost" size="sm" @click="isCreateFaskesOpen = false">
            Batal
          </BaseButton>
          <BaseButton
            variant="primary"
            size="sm"
            :loading="isSubmittingFaskes"
            @click="createFaskes"
          >
            Simpan Faskes Rujukan
          </BaseButton>
        </div>
      </div>
    </BaseModal>

    <!-- MODAL: Tambah Dokter Spesialis Baru -->
    <BaseModal
      v-model="isCreateSpesialisOpen"
      title="Tambah Master Dokter Spesialis Rujukan"
      width="md"
    >
      <div class="space-y-3">
        <BaseInput
          v-model="spesialisForm.code"
          label="Kode Dokter / ID Spesialis *"
          placeholder="DRSP-008"
          required
        />
        <BaseInput
          v-model="spesialisForm.nama_dokter"
          label="Nama Lengkap Dokter & Gelar *"
          placeholder="dr. Anindya Putri, Sp.A, Subsp.Neo(K)"
          required
        />
        <BaseInput
          v-model="spesialisForm.spesialisasi"
          label="Bidang Spesialisasi / Subspesialisasi *"
          placeholder="Spesialis Anak (Konsultan Neonatologi)"
          required
        />
        <BaseInput
          v-model="spesialisForm.rumah_sakit"
          label="Rumah Sakit Praktek / Induk"
          placeholder="RSUP Dr. Sardjito Yogyakarta"
        />

        <div class="flex items-center justify-end gap-2 pt-3 border-t border-slate-100">
          <BaseButton variant="ghost" size="sm" @click="isCreateSpesialisOpen = false">
            Batal
          </BaseButton>
          <BaseButton
            variant="primary"
            size="sm"
            :loading="isSubmittingSpesialis"
            @click="createSpesialis"
          >
            Simpan Dokter Spesialis
          </BaseButton>
        </div>
      </div>
    </BaseModal>
  </div>
</template>
