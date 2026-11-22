<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { api } from '@/api/client'
import { useUiStore } from '@/stores/ui'
import { usePatientContextStore } from '@/stores/patientContext'
import DataTable, { type Column } from '@/components/common/DataTable.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import SearchableSelect from '@/components/common/SearchableSelect.vue'
import BaseModal from '@/components/common/BaseModal.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import type { Visit, Patient, MasterPoli, MasterGuarantor, User } from '@/types'

const uiStore = useUiStore()
const patientContext = usePatientContextStore()

const visits = ref<Visit[]>([])
const isLoading = ref(false)
const statusFilter = ref('')
const search = ref('')
const page = ref(1)
const total = ref(0)
const limit = 20

// Modal Pendaftaran Kunjungan
const isModalOpen = ref(false)
const isSubmitting = ref(false)

// Print Ticket Modal
const isPrintModalOpen = ref(false)
const printedVisit = ref<Visit | null>(null)

// Form options
const patientsList = ref<Patient[]>([])
const poliOptions = ref<{ value: number; label: string }[]>([])
const doctorOptions = ref<{ value: string; label: string }[]>([])
const guarantorOptions = ref<{ value: number; label: string }[]>([])

const form = ref({
  patient_id: '',
  poli_id: 2, // default Poli Anak
  doctor_id: '',
  guarantor_id: 1, // default Umum
  nomor_asuransi: '',
  keluhan_utama: '',
})

const columns: Column[] = [
  { key: 'no_antrian', label: 'No. Antrean', width: '120px' },
  { key: 'no_rawat', label: 'No. Rawat', width: '160px' },
  { key: 'no_rm', label: 'No. RM', width: '130px' },
  { key: 'patient_name', label: 'Nama Pasien' },
  { key: 'poli_name', label: 'Poli Dituju' },
  { key: 'doctor_name', label: 'Dokter Praktik' },
  { key: 'guarantor_name', label: 'Penjamin' },
  { key: 'status', label: 'Status Antrean', align: 'center', width: '140px' },
  { key: 'actions', label: 'Aksi', align: 'right', width: '180px' },
]

async function loadVisits() {
  isLoading.value = true
  try {
    const res = await api.get<Visit[]>('/visits', {
      status: statusFilter.value,
      search: search.value,
      page: page.value,
      limit,
    })
    if (res.success && res.data) {
      visits.value = res.data
      total.value = res.meta?.total_data || 0
    }
  } catch (err: any) {
    uiStore.showToast(err.message, 'danger')
  } finally {
    isLoading.value = false
  }
}

async function loadOptions() {
  try {
    const [patRes, poliRes, docRes, guarRes] = await Promise.all([
      api.get<Patient[]>('/patients', { limit: 50 }),
      api.get<MasterPoli[]>('/master/poli'),
      api.get<User[]>('/users', { role_id: 5, limit: 50 }), // DOCTOR role
      api.get<MasterGuarantor[]>('/master/guarantors'),
    ])

    if (patRes.success && patRes.data) {
      patientsList.value = patRes.data
    }
    if (poliRes.success && poliRes.data) {
      poliOptions.value = poliRes.data.map(p => ({ value: p.id, label: p.name }))
    }
    if (docRes.success && docRes.data) {
      doctorOptions.value = docRes.data.map(d => ({ value: d.id, label: d.full_name }))
    }
    if (guarRes.success && guarRes.data) {
      guarantorOptions.value = guarRes.data.map(g => ({ value: g.id, label: g.name }))
    }
  } catch (err) {
    console.error('Failed to load registration options:', err)
  }
}

const patientSelectOptions = computed(() => {
  return patientsList.value.map(p => ({
    value: p.id,
    label: `${p.nama_pasien} (RM: ${p.no_rm})`,
    subtitle: `NIK: ${p.nik} • TTL: ${p.tempat_lahir || '-'}, ${p.tanggal_lahir}`,
    badge: p.jenis_kelamin === 'L' ? 'Laki-laki' : 'Perempuan'
  }))
})

function openRegistrationModal() {
  form.value = {
    patient_id: patientContext.patient?.id || '',
    poli_id: 2,
    doctor_id: doctorOptions.value[0]?.value || '',
    guarantor_id: 1,
    nomor_asuransi: '',
    keluhan_utama: '',
  }
  isModalOpen.value = true
}

async function submitVisit() {
  if (!form.value.patient_id) {
    uiStore.showToast('Silakan pilih pasien yang akan didaftarkan', 'warning')
    return
  }

  isSubmitting.value = true
  try {
    const res = await api.post<Visit>('/visits', form.value)
    if (res.success && res.data) {
      uiStore.showToast(`Kunjungan berhasil didaftarkan: ${res.data.no_antrian}`, 'success')
      isModalOpen.value = false
      printedVisit.value = res.data
      isPrintModalOpen.value = true

      // Update active context
      const p = patientsList.value.find(x => x.id === form.value.patient_id)
      if (p) {
        patientContext.setActivePatient(p, res.data)
      }

      loadVisits()
    }
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal mendaftarkan kunjungan', 'danger')
  } finally {
    isSubmitting.value = false
  }
}

function printTicket(visit: Visit) {
  printedVisit.value = visit
  isPrintModalOpen.value = true
}

function selectPatientContext(visit: Visit) {
  const p: Patient = {
    id: visit.patient_id,
    no_rm: visit.no_rm || '',
    nik: '',
    nama_pasien: visit.patient_name || '',
    jenis_kelamin: visit.patient_gender || '',
    tanggal_lahir: visit.patient_dob || '',
    riwayat_alergi: visit.patient_allergy,
    status_aktif: true,
    created_at: '',
    updated_at: '',
  }
  patientContext.setActivePatient(p, visit)
  uiStore.showToast(`Pasien ${visit.patient_name} aktif di layar kerja`, 'info')
}

onMounted(() => {
  loadVisits()
  loadOptions()
})
</script>

<template>
  <div class="space-y-4 text-left">
    <!-- Header -->
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-[20px] font-bold text-[#0F172A] leading-[28px]">
          Pendaftaran Kunjungan & Antrean Pasien Rawat Jalan
        </h1>
        <p class="text-[13px] text-[#475569]">
          Registrasi kunjungan baru/lama, penerbitan nomor antrean loket/poli, verifikasi penjamin, dan cetak karcis antrean.
        </p>
      </div>

      <div class="flex items-center gap-2">
        <router-link to="/registration/patients">
          <BaseButton variant="secondary" size="md">
            + Registrasi Pasien Baru
          </BaseButton>
        </router-link>

        <BaseButton variant="primary" size="md" @click="openRegistrationModal">
          <template #icon>
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
            </svg>
          </template>
          Daftarkan Pasien ke Antrean Poli
        </BaseButton>
      </div>
    </div>

    <!-- Filters Bar -->
    <div class="bg-white p-3 rounded-lg border border-[#E2E8F0] shadow-sm flex flex-wrap items-center gap-3">
      <div class="flex-1 min-w-[200px]">
        <BaseInput
          v-model="search"
          placeholder="Cari No. Antrean, No. Rawat, No. RM, atau Nama Pasien..."
          compact
          @keyup.enter="loadVisits"
        />
      </div>

      <div class="w-[200px]">
        <BaseSelect
          v-model="statusFilter"
          :options="[
            { value: '', label: 'Semua Status Antrean' },
            { value: 'MENUNGGU_TRIASE', label: 'Menunggu Triase (Perawat)' },
            { value: 'DIPERIKSA_DOKTER', label: 'Diperiksa Dokter' },
            { value: 'FARMASI', label: 'Farmasi' },
            { value: 'KASIR', label: 'Kasir & Billing' },
            { value: 'SELESAI', label: 'Selesai Dilayani' },
          ]"
          placeholder="Filter Status"
          compact
          @change="loadVisits"
        />
      </div>

      <BaseButton variant="secondary" size="sm" @click="loadVisits">
        Filter
      </BaseButton>
    </div>

    <!-- Antrean DataTable -->
    <DataTable
      :columns="columns"
      :data="visits"
      :loading="isLoading"
      :page="page"
      :limit="limit"
      :total="total"
      empty-text="Belum ada kunjungan terdaftar pada kriteria ini"
      @page-change="(p) => { page = p; loadVisits() }"
    >
      <template #no_antrian="{ row }">
        <span class="font-bold text-[#0284C7] font-mono tabular-nums bg-[#F0F9FF] px-2 py-0.5 rounded border border-[#BAE6FD] text-[13px]">
          {{ row.no_antrian }}
        </span>
      </template>

      <template #no_rawat="{ row }">
        <span class="font-mono text-[12px] text-[#475569]">{{ row.no_rawat }}</span>
      </template>

      <template #no_rm="{ row }">
        <span class="font-mono font-medium text-[#0369A1] bg-[#E0F2FE] px-1.5 py-0.5 rounded text-[12px] tabular-nums">
          {{ row.no_rm }}
        </span>
      </template>

      <template #patient_name="{ row }">
        <div class="font-semibold text-[#0F172A]">{{ row.patient_name }}</div>
        <div v-if="row.patient_allergy" class="text-[11px] text-[#DC2626] font-semibold flex items-center gap-1">
          <span>⚠️ ALERGI: {{ row.patient_allergy }}</span>
        </div>
      </template>

      <template #status="{ row }">
        <StatusBadge :status="row.status" />
      </template>

      <template #actions="{ row }">
        <div class="flex items-center justify-end gap-1.5">
          <BaseButton
            variant="soft"
            size="sm"
            title="Pilih pasien ini untuk dilayani di layar aktif"
            @click="selectPatientContext(row)"
          >
            Pilih Pasien
          </BaseButton>
          <BaseButton
            variant="secondary"
            size="sm"
            title="Cetak Karcis Antrean"
            @click="printTicket(row)"
          >
            Cetak Karcis
          </BaseButton>
        </div>
      </template>
    </DataTable>

    <!-- Modal Form Pendaftaran Kunjungan -->
    <BaseModal
      :is-open="isModalOpen"
      title="Pendaftaran Kunjungan Rawat Jalan & Poli"
      max-width="lg"
      @close="isModalOpen = false"
    >
      <form class="space-y-3" @submit.prevent="submitVisit">
        <SearchableSelect
          v-model="form.patient_id"
          label="Pilih Pasien Terdaftar"
          :options="patientSelectOptions"
          placeholder="-- Cari Rekam Medis / NIK / Nama Pasien --"
          search-placeholder="Ketik No. RM, NIK, atau Nama..."
          required
        />

        <div class="grid grid-cols-2 gap-3">
          <BaseSelect
            v-model="form.poli_id"
            label="Poli Tujuan Pelayanan"
            :options="poliOptions"
            required
          />

          <BaseSelect
            v-model="form.doctor_id"
            label="Dokter Pemeriksa Praktik"
            :options="doctorOptions"
          />
        </div>

        <div class="grid grid-cols-2 gap-3">
          <BaseSelect
            v-model="form.guarantor_id"
            label="Jenis Penjamin"
            :options="guarantorOptions"
            required
          />

          <BaseInput
            v-model="form.nomor_asuransi"
            label="Nomor Kartu BPJS / Polis"
            placeholder="Wajib jika BPJS/Asuransi"
          />
        </div>

        <div>
          <label class="text-[13px] font-medium text-[#334155] block mb-1">
            Keluhan Utama Pasien / Indikasi Berobat
          </label>
          <textarea
            v-model="form.keluhan_utama"
            rows="3"
            placeholder="contoh: Demam 3 hari naik turun disertai batuk pilek..."
            class="w-full p-2.5 text-[13px] text-[#0F172A] bg-white border border-[#CBD5E1] rounded-[6px] focus:border-[#0284C7] focus:ring-3 focus:ring-[rgba(2,132,199,0.15)] outline-none"
          ></textarea>
        </div>
      </form>

      <template #footer>
        <BaseButton variant="secondary" size="md" @click="isModalOpen = false">
          Batal
        </BaseButton>
        <BaseButton variant="primary" size="md" :loading="isSubmitting" @click="submitVisit">
          Terbitkan Karcis Antrean
        </BaseButton>
      </template>
    </BaseModal>

    <!-- Modal Cetak Karcis Antrean (Spesifikasi Klinis SIMRS) -->
    <BaseModal
      :is-open="isPrintModalOpen"
      title="Cetak Karcis Nomor Antrean Poli"
      max-width="sm"
      @close="isPrintModalOpen = false"
    >
      <div v-if="printedVisit" class="bg-white border-2 border-dashed border-[#CBD5E1] p-4 rounded text-center space-y-2 font-mono">
        <div class="text-[14px] font-bold text-[#0F172A]">SIMRS-ASOY E-CLINIC</div>
        <div class="text-[11px] text-[#64748B]">Klinik Pratama & Terapi Tumbuh Kembang</div>
        <div class="border-b border-[#E2E8F0] my-2"></div>

        <div class="text-[12px] text-[#475569]">NOMOR ANTREAN POLI</div>
        <div class="text-[36px] font-bold text-[#0284C7] leading-none my-1 tracking-wider">
          {{ printedVisit.no_antrian }}
        </div>
        <div class="text-[13px] font-bold text-[#0F172A]">{{ printedVisit.poli_name }}</div>

        <div class="border-b border-[#E2E8F0] my-2"></div>

        <div class="text-left text-[12px] space-y-1 text-[#334155]">
          <div><span class="text-[#64748B]">No. Rawat :</span> {{ printedVisit.no_rawat }}</div>
          <div><span class="text-[#64748B]">No. RM    :</span> {{ printedVisit.no_rm }}</div>
          <div><span class="text-[#64748B]">Pasien    :</span> {{ printedVisit.patient_name }}</div>
          <div><span class="text-[#64748B]">Penjamin  :</span> {{ printedVisit.guarantor_name }}</div>
        </div>

        <div class="border-b border-[#E2E8F0] my-2"></div>
        <div class="text-[10px] text-[#94A3B8]">
          Harap menunggu nomor antrean dipanggil oleh perawat di ruang tunggu.
        </div>
      </div>

      <template #footer>
        <BaseButton variant="primary" size="sm" @click="isPrintModalOpen = false">
          Selesai / Cetak Karcis
        </BaseButton>
      </template>
    </BaseModal>
  </div>
</template>
