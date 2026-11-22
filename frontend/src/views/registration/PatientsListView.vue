<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/client'
import { useUiStore } from '@/stores/ui'
import { usePatientContextStore } from '@/stores/patientContext'
import DataTable, { type Column } from '@/components/common/DataTable.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseDatePicker from '@/components/common/BaseDatePicker.vue'
import BaseModal from '@/components/common/BaseModal.vue'
import type { Patient } from '@/types'

const uiStore = useUiStore()
const patientContext = usePatientContextStore()
const router = useRouter()

const patients = ref<Patient[]>([])
const isLoading = ref(false)
const search = ref('')
const page = ref(1)
const total = ref(0)
const limit = 15

// Modal state
const isModalOpen = ref(false)
const isEditing = ref(false)
const isSubmitting = ref(false)
const currentPatientId = ref('')

const form = ref({
  nik: '',
  nama_pasien: '',
  jenis_kelamin: 'L',
  tempat_lahir: '',
  tanggal_lahir: '',
  nama_orang_tua: '',
  no_hp: '',
  alamat: '',
  riwayat_alergi: '',
  golongan_darah: 'O',
})

const columns: Column[] = [
  { key: 'no_rm', label: 'No. Rekam Medis', width: '140px' },
  { key: 'nama_pasien', label: 'Nama Pasien' },
  { key: 'jenis_kelamin', label: 'L/P', align: 'center', width: '60px' },
  { key: 'nik', label: 'NIK KTP / KIA', width: '160px' },
  { key: 'tanggal_lahir', label: 'Tanggal Lahir', width: '130px' },
  { key: 'nama_orang_tua', label: 'Nama Orang Tua / Wali' },
  { key: 'riwayat_alergi', label: 'Alert Alergi', width: '180px' },
  { key: 'actions', label: 'Aksi', align: 'right', width: '200px' },
]

async function loadPatients() {
  isLoading.value = true
  try {
    const res = await api.get<Patient[]>('/patients', {
      search: search.value,
      page: page.value,
      limit,
    })
    if (res.success && res.data) {
      patients.value = res.data
      total.value = res.meta?.total_data || 0
    }
  } catch (err: any) {
    uiStore.showToast(err.message, 'danger')
  } finally {
    isLoading.value = false
  }
}

function openAddModal() {
  isEditing.value = false
  currentPatientId.value = ''
  form.value = {
    nik: '',
    nama_pasien: '',
    jenis_kelamin: 'L',
    tempat_lahir: '',
    tanggal_lahir: '',
    nama_orang_tua: '',
    no_hp: '',
    alamat: '',
    riwayat_alergi: '',
    golongan_darah: 'O',
  }
  isModalOpen.value = true
}

function openEditModal(patient: Patient) {
  isEditing.value = true
  currentPatientId.value = patient.id
  form.value = {
    nik: patient.nik,
    nama_pasien: patient.nama_pasien,
    jenis_kelamin: patient.jenis_kelamin,
    tempat_lahir: patient.tempat_lahir || '',
    tanggal_lahir: patient.tanggal_lahir,
    nama_orang_tua: patient.nama_orang_tua || '',
    no_hp: patient.no_hp || '',
    alamat: patient.alamat || '',
    riwayat_alergi: patient.riwayat_alergi || '',
    golongan_darah: patient.golongan_darah || 'O',
  }
  isModalOpen.value = true
}

async function savePatient() {
  isSubmitting.value = true
  try {
    if (isEditing.value) {
      await api.put(`/patients/${currentPatientId.value}`, form.value)
      uiStore.showToast('Data pasien berhasil diperbarui', 'success')
    } else {
      const res = await api.post<Patient>('/patients', form.value)
      uiStore.showToast(`Pasien berhasil didaftarkan: ${res.data?.no_rm}`, 'success')
      if (res.data) {
        patientContext.setActivePatient(res.data)
      }
    }
    isModalOpen.value = false
    loadPatients()
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal menyimpan pasien', 'danger')
  } finally {
    isSubmitting.value = false
  }
}

function selectPatient(patient: Patient) {
  patientContext.setActivePatient(patient)
  uiStore.showToast(`Pasien ${patient.nama_pasien} (${patient.no_rm}) aktif di layar kerja`, 'info')
}

function registerVisitForPatient(patient: Patient) {
  patientContext.setActivePatient(patient)
  router.push('/registration')
}

onMounted(() => {
  loadPatients()
})
</script>

<template>
  <div class="space-y-4 text-left">
    <!-- Header -->
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-[20px] font-bold text-[#0F172A] leading-[28px]">
          Direktori Induk Pasien & Rekam Medis (Master Pasien)
        </h1>
        <p class="text-[13px] text-[#475569]">
          Pencarian pasien, penerbitan No. RM otomatis, riwayat alergi obat, dan profil kependudukan NIK / KIA.
        </p>
      </div>

      <BaseButton variant="primary" size="md" @click="openAddModal">
        <template #icon>
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z"></path>
          </svg>
        </template>
        Registrasi Pasien Baru
      </BaseButton>
    </div>

    <!-- Filters Bar -->
    <div class="bg-white p-3 rounded-lg border border-[#E2E8F0] shadow-sm flex flex-wrap items-center gap-3">
      <div class="flex-1 min-w-[200px]">
        <BaseInput
          v-model="search"
          placeholder="Cari No. RM, NIK KTP/KIA, Nama Pasien, atau No. HP..."
          compact
          @keyup.enter="loadPatients"
        />
      </div>

      <BaseButton variant="secondary" size="sm" @click="loadPatients">
        Cari Pasien
      </BaseButton>
    </div>

    <!-- Patients DataTable -->
    <DataTable
      :columns="columns"
      :data="patients"
      :loading="isLoading"
      :page="page"
      :limit="limit"
      :total="total"
      empty-text="Tidak ada data pasien yang cocok"
      @page-change="(p) => { page = p; loadPatients() }"
    >
      <template #no_rm="{ row }">
        <span class="font-mono font-bold text-[#0369A1] bg-[#E0F2FE] px-2 py-0.5 rounded border border-[#BAE6FD] text-[13px] tabular-nums">
          {{ row.no_rm }}
        </span>
      </template>

      <template #nama_pasien="{ row }">
        <span class="font-semibold text-[#0F172A]">{{ row.nama_pasien }}</span>
      </template>

      <template #nik="{ row }">
        <span class="font-mono text-[#475569] text-[12px] tabular-nums">{{ row.nik }}</span>
      </template>

      <template #tanggal_lahir="{ row }">
        <span class="font-mono text-[12px] tabular-nums text-[#334155]">{{ row.tanggal_lahir }}</span>
      </template>

      <template #riwayat_alergi="{ row }">
        <span
          v-if="row.riwayat_alergi"
          class="inline-flex items-center gap-1 text-[11px] font-semibold text-[#B91C1C] bg-[#FEE2E2] px-2 py-0.5 rounded border border-[#FCA5A5]"
        >
          ⚠️ {{ row.riwayat_alergi }}
        </span>
        <span v-else class="text-[#94A3B8] text-[12px]">- Tidak ada -</span>
      </template>

      <template #actions="{ row }">
        <div class="flex items-center justify-end gap-1.5">
          <BaseButton
            variant="soft"
            size="sm"
            title="Jadikan pasien ini aktif pada layar kerja"
            @click="selectPatient(row)"
          >
            Pilih
          </BaseButton>
          <BaseButton
            variant="secondary"
            size="sm"
            title="Daftarkan kunjungan rawat jalan"
            @click="registerVisitForPatient(row)"
          >
            Daftar Poli
          </BaseButton>
          <BaseButton
            variant="ghost"
            size="sm"
            @click="openEditModal(row)"
          >
            Edit
          </BaseButton>
        </div>
      </template>
    </DataTable>

    <!-- Modal Form Pasien Baru / Edit -->
    <BaseModal
      :is-open="isModalOpen"
      :title="isEditing ? 'Ubah Profil Pasien' : 'Registrasi Pasien Baru (Penerbitan No. RM)'"
      max-width="lg"
      @close="isModalOpen = false"
    >
      <form class="space-y-3" @submit.prevent="savePatient">
        <div class="grid grid-cols-2 gap-3">
          <BaseInput
            v-model="form.nama_pasien"
            label="Nama Lengkap Pasien"
            placeholder="contoh: Hilmi Al-Fatih"
            required
          />

          <BaseInput
            v-model="form.nik"
            label="NIK KTP / Kartu Identitas Anak (KIA)"
            placeholder="16 digit NIK"
            required
          />
        </div>

        <div class="grid grid-cols-3 gap-3">
          <BaseSelect
            v-model="form.jenis_kelamin"
            label="Jenis Kelamin"
            :options="[
              { value: 'L', label: 'Laki-Laki (L)' },
              { value: 'P', label: 'Perempuan (P)' }
            ]"
            required
          />

          <BaseInput
            v-model="form.tempat_lahir"
            label="Tempat Lahir"
            placeholder="Bandung"
          />

          <BaseDatePicker
            v-model="form.tanggal_lahir"
            label="Tanggal Lahir"
            required
          />
        </div>

        <div class="grid grid-cols-3 gap-3">
          <BaseInput
            v-model="form.nama_orang_tua"
            label="Nama Orang Tua / Wali"
            placeholder="Ayah & Ibu"
          />

          <BaseInput
            v-model="form.no_hp"
            label="Nomor HP / WhatsApp"
            placeholder="081234567890"
          />

          <BaseSelect
            v-model="form.golongan_darah"
            label="Golongan Darah"
            :options="[
              { value: 'A', label: 'A' },
              { value: 'B', label: 'B' },
              { value: 'AB', label: 'AB' },
              { value: 'O', label: 'O' }
            ]"
          />
        </div>

        <div>
          <BaseInput
            v-model="form.alamat"
            label="Alamat Domisili Lengkap"
            placeholder="Jl. Sukajadi No. 45, Bandung"
          />
        </div>

        <!-- Alert Alergi (Spesifikasi Kritis SIMRS) -->
        <div class="bg-[#FEF3C7] p-3 rounded border border-[#FCD34D] text-left">
          <label class="text-[12px] font-bold text-[#B45309] block mb-1">
            ⚠️ Riwayat Alergi Obat / Makanan (Sangat Penting):
          </label>
          <input
            v-model="form.riwayat_alergi"
            type="text"
            placeholder="contoh: Alergi Penisilin, Amoxicillin, Seafood (Kosongkan jika tidak ada)"
            class="w-full px-3 h-[32px] text-[13px] bg-white border border-[#CBD5E1] rounded focus:border-[#0284C7] outline-none"
          />
          <span class="text-[11px] text-[#92400E] block mt-1">
            Data alergi ini akan otomatis memicu badge merah mencolok di modul Dokter, Perawat, dan Farmasi.
          </span>
        </div>
      </form>

      <template #footer>
        <BaseButton variant="secondary" size="md" @click="isModalOpen = false">
          Batal
        </BaseButton>
        <BaseButton variant="primary" size="md" :loading="isSubmitting" @click="savePatient">
          {{ isEditing ? 'Simpan Perubahan' : 'Terbitkan No. RM & Simpan Pasien' }}
        </BaseButton>
      </template>
    </BaseModal>
  </div>
</template>
