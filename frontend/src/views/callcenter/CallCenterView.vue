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
import BaseDatePicker from '@/components/common/BaseDatePicker.vue'
import BaseModal from '@/components/common/BaseModal.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import type { Appointment, Patient, MasterPoli, User } from '@/types'

const uiStore = useUiStore()
const patientContext = usePatientContextStore()

const appointments = ref<Appointment[]>([])
const isLoading = ref(false)
const selectedDate = ref('')
const statusFilter = ref('')
const page = ref(1)
const total = ref(0)
const limit = 20

// Options
const patientsList = ref<Patient[]>([])
const poliOptions = ref<{ value: number; label: string }[]>([])
const doctorOptions = ref<{ value: string; label: string }[]>([])

const patientSelectOptions = computed(() => {
  return patientsList.value.map(p => ({
    value: p.id,
    label: `${p.nama_pasien} (RM: ${p.no_rm})`,
    subtitle: `NIK: ${p.nik} • HP: ${p.no_hp || '-'}`,
    badge: p.jenis_kelamin === 'L' ? 'Laki-laki' : 'Perempuan'
  }))
})

// Modals
const isCreateModalOpen = ref(false)
const isRescheduleModalOpen = ref(false)
const isSubmitting = ref(false)
const selectedAppointment = ref<Appointment | null>(null)

const createForm = ref({
  patient_id: '',
  poli_id: 2,
  doctor_id: '',
  appointment_date: new Date().toISOString().split('T')[0],
  start_time: '09:00',
  end_time: '09:30',
  channel: 'WHATSAPP',
  catatan: '',
})

const rescheduleForm = ref({
  appointment_date: '',
  start_time: '',
  end_time: '',
  catatan: '',
})

const columns: Column[] = [
  { key: 'appointment_date', label: 'Tanggal Praktik', width: '130px' },
  { key: 'start_time', label: 'Jam Praktik', width: '110px' },
  { key: 'no_rm', label: 'No. RM', width: '130px' },
  { key: 'patient_name', label: 'Nama Pasien' },
  { key: 'no_hp', label: 'No. HP / WA', width: '140px' },
  { key: 'poli_name', label: 'Poli Unit' },
  { key: 'doctor_name', label: 'Dokter' },
  { key: 'channel', label: 'Saluran Booking', width: '120px' },
  { key: 'status', label: 'Status', align: 'center', width: '130px' },
  { key: 'actions', label: 'Aksi Operasional', align: 'right', width: '220px' },
]

async function loadAppointments() {
  isLoading.value = true
  try {
    const res = await api.get<Appointment[]>('/appointments', {
      date: selectedDate.value,
      status: statusFilter.value,
      page: page.value,
      limit,
    })
    if (res.success && res.data) {
      appointments.value = res.data
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
    const [patRes, poliRes, docRes] = await Promise.all([
      api.get<Patient[]>('/patients', { limit: 50 }),
      api.get<MasterPoli[]>('/master/poli'),
      api.get<User[]>('/users', { role_id: 5, limit: 50 }),
    ])

    if (patRes.data) patientsList.value = patRes.data
    if (poliRes.data) poliOptions.value = poliRes.data.map(p => ({ value: p.id, label: p.name }))
    if (docRes.data) doctorOptions.value = docRes.data.map(d => ({ value: d.id, label: d.full_name }))
  } catch (err) {
    console.error('Failed to load options:', err)
  }
}

function openCreateModal() {
  createForm.value = {
    patient_id: patientContext.patient?.id || '',
    poli_id: 2,
    doctor_id: doctorOptions.value[0]?.value || '',
    appointment_date: new Date().toISOString().split('T')[0],
    start_time: '09:00',
    end_time: '09:30',
    channel: 'WHATSAPP',
    catatan: '',
  }
  isCreateModalOpen.value = true
}

async function submitAppointment() {
  if (!createForm.value.patient_id) {
    uiStore.showToast('Silakan pilih pasien', 'warning')
    return
  }

  isSubmitting.value = true
  try {
    await api.post('/appointments', createForm.value)
    uiStore.showToast('Janji temu berhasil dijadwalkan', 'success')
    isCreateModalOpen.value = false
    loadAppointments()
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal membuat appointment', 'danger')
  } finally {
    isSubmitting.value = false
  }
}

function openReschedule(apt: Appointment) {
  selectedAppointment.value = apt
  rescheduleForm.value = {
    appointment_date: apt.appointment_date,
    start_time: apt.start_time.substring(0, 5),
    end_time: apt.end_time.substring(0, 5),
    catatan: apt.catatan,
  }
  isRescheduleModalOpen.value = true
}

async function submitReschedule() {
  if (!selectedAppointment.value) return
  isSubmitting.value = true
  try {
    await api.post(`/appointments/${selectedAppointment.value.id}/reschedule`, rescheduleForm.value)
    uiStore.showToast('Jadwal temu berhasil diperbarui', 'success')
    isRescheduleModalOpen.value = false
    loadAppointments()
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal menjadwalkan ulang', 'danger')
  } finally {
    isSubmitting.value = false
  }
}

function handleCheckIn(apt: Appointment) {
  uiStore.openConfirm({
    title: 'Konfirmasi Check-in Pasien',
    message: `Lakukan check-in untuk pasien "${apt.patient_name}"? Pasien akan langsung diterbitkan nomor antrean dan masuk ke antrean perawat / poli.`,
    confirmText: 'Check-in Pasien',
    variant: 'primary',
    onConfirm: async () => {
      try {
        const res = await api.post<{ no_antrian?: string }>(`/appointments/${apt.id}/checkin`)
        uiStore.showToast(`Pasien ${apt.patient_name} berhasil check-in: Antrean ${res.data?.no_antrian}`, 'success')
        loadAppointments()
      } catch (err: any) {
        uiStore.showToast(err.message || 'Gagal check-in', 'danger')
      }
    }
  })
}

function cancelAppointment(apt: Appointment) {
  uiStore.openConfirm({
    title: 'Batalkan Janji Temu',
    message: `Apakah Anda yakin ingin membatalkan janji temu untuk pasien "${apt.patient_name}"?`,
    confirmText: 'Ya, Batalkan',
    variant: 'danger',
    onConfirm: async () => {
      try {
        await api.post(`/appointments/${apt.id}/cancel`, { notes: 'Dibatalkan via Call Center' })
        uiStore.showToast('Janji temu telah dibatalkan', 'success')
        loadAppointments()
      } catch (err: any) {
        uiStore.showToast(err.message || 'Gagal membatalkan', 'danger')
      }
    }
  })
}

onMounted(() => {
  loadAppointments()
  loadOptions()
})
</script>

<template>
  <div class="space-y-4 text-left">
    <!-- Header -->
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-[20px] font-bold text-[#0F172A] leading-[28px]">
          Call Center & Manajemen Janji Temu (Appointments)
        </h1>
        <p class="text-[13px] text-[#475569]">
          Pemesanan janji temu via WhatsApp/Telepon, penjadwalan dokter, konfirmasi kehadiran (Check-in), dan reschedule.
        </p>
      </div>

      <BaseButton variant="primary" size="md" @click="openCreateModal">
        <template #icon>
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"></path>
          </svg>
        </template>
        + Jadwalkan Janji Temu Baru
      </BaseButton>
    </div>

    <!-- Filters Bar -->
    <div class="bg-white p-3 rounded-lg border border-[#E2E8F0] shadow-sm flex flex-wrap items-center gap-3">
      <div class="w-[210px]">
        <BaseDatePicker
          v-model="selectedDate"
          placeholder="Filter Tanggal..."
          @change="loadAppointments"
        />
      </div>

      <div class="w-[200px]">
        <BaseSelect
          v-model="statusFilter"
          :options="[
            { value: '', label: 'Semua Status' },
            { value: 'TERJADWAL', label: 'Terjadwal' },
            { value: 'HADIR', label: 'Hadir (Check-in)' },
            { value: 'RESCHEDULE', label: 'Reschedule' },
            { value: 'BATAL', label: 'Batal' },
          ]"
          placeholder="Filter Status"
          compact
          @change="loadAppointments"
        />
      </div>

      <BaseButton variant="secondary" size="sm" @click="loadAppointments">
        Filter
      </BaseButton>
    </div>

    <!-- DataTable -->
    <DataTable
      :columns="columns"
      :data="appointments"
      :loading="isLoading"
      :page="page"
      :limit="limit"
      :total="total"
      empty-text="Belum ada janji temu pada tanggal / kriteria ini"
      @page-change="(p) => { page = p; loadAppointments() }"
    >
      <template #appointment_date="{ row }">
        <span class="font-mono text-[12px] font-semibold text-[#0F172A] tabular-nums">
          {{ row.appointment_date }}
        </span>
      </template>

      <template #start_time="{ row }">
        <span class="font-mono text-[12px] text-[#0284C7] bg-[#F0F9FF] px-1.5 py-0.5 rounded border border-[#BAE6FD] tabular-nums">
          {{ row.start_time.substring(0, 5) }} - {{ row.end_time.substring(0, 5) }}
        </span>
      </template>

      <template #no_rm="{ row }">
        <span class="font-mono text-[12px] font-medium text-[#0369A1] bg-[#E0F2FE] px-1.5 py-0.5 rounded tabular-nums">
          {{ row.no_rm }}
        </span>
      </template>

      <template #patient_name="{ row }">
        <span class="font-semibold text-[#0F172A]">{{ row.patient_name }}</span>
      </template>

      <template #channel="{ row }">
        <span class="text-[11px] font-medium px-2 py-0.5 rounded bg-[#F1F5F9] text-[#334155] border border-[#CBD5E1]">
          {{ row.channel }}
        </span>
      </template>

      <template #status="{ row }">
        <StatusBadge :status="row.status" />
      </template>

      <template #actions="{ row }">
        <div class="flex items-center justify-end gap-1.5">
          <BaseButton
            v-if="row.status !== 'HADIR' && row.status !== 'BATAL'"
            variant="primary"
            size="sm"
            title="Konfirmasi kedatangan pasien ke klinik"
            @click="handleCheckIn(row)"
          >
            Check-in
          </BaseButton>
          <BaseButton
            v-if="row.status !== 'HADIR' && row.status !== 'BATAL'"
            variant="soft"
            size="sm"
            @click="openReschedule(row)"
          >
            Reschedule
          </BaseButton>
          <BaseButton
            v-if="row.status !== 'HADIR' && row.status !== 'BATAL'"
            variant="ghost"
            size="sm"
            @click="cancelAppointment(row)"
          >
            Batal
          </BaseButton>
          <span v-if="row.status === 'HADIR'" class="text-[12px] text-[#15803D] font-medium">
            Telah Masuk Antrean
          </span>
        </div>
      </template>
    </DataTable>

    <!-- Modal Form Appointment Baru -->
    <BaseModal
      :is-open="isCreateModalOpen"
      title="Pemesanan Janji Temu (Appointment Booking)"
      max-width="lg"
      @close="isCreateModalOpen = false"
    >
      <form class="space-y-3" @submit.prevent="submitAppointment">
        <SearchableSelect
          v-model="createForm.patient_id"
          label="Pilih Pasien Terdaftar"
          :options="patientSelectOptions"
          placeholder="-- Cari Pasien / No. RM / No. WA --"
          search-placeholder="Ketik Nama, No. RM, atau HP..."
          required
        />

        <div class="grid grid-cols-2 gap-3">
          <BaseSelect
            v-model="createForm.poli_id"
            label="Poli Pelayanan"
            :options="poliOptions"
            required
          />

          <BaseSelect
            v-model="createForm.doctor_id"
            label="Dokter Praktik"
            :options="doctorOptions"
          />
        </div>

        <div class="grid grid-cols-3 gap-3">
          <BaseDatePicker
            v-model="createForm.appointment_date"
            label="Tanggal Janji Temu"
            required
          />

          <BaseInput
            v-model="createForm.start_time"
            label="Jam Mulai"
            type="time"
            required
          />

          <BaseInput
            v-model="createForm.end_time"
            label="Jam Selesai"
            type="time"
            required
          />
        </div>

        <div class="grid grid-cols-2 gap-3">
          <BaseSelect
            v-model="createForm.channel"
            label="Saluran Komunikasi Pemesanan"
            :options="[
              { value: 'WHATSAPP', label: 'WhatsApp' },
              { value: 'TELEPON', label: 'Telepon Langsung' },
              { value: 'WEBSITE', label: 'Portal / Web' },
              { value: 'LOKET', label: 'Loket Pendaftaran' },
            ]"
            required
          />

          <BaseInput
            v-model="createForm.catatan"
            label="Catatan / Keluhan Awal"
            placeholder="contoh: Sesi evaluasi terapi okupasi"
          />
        </div>
      </form>

      <template #footer>
        <BaseButton variant="secondary" size="md" @click="isCreateModalOpen = false">
          Batal
        </BaseButton>
        <BaseButton variant="primary" size="md" :loading="isSubmitting" @click="submitAppointment">
          Simpan Jadwal Temu
        </BaseButton>
      </template>
    </BaseModal>

    <!-- Modal Reschedule -->
    <BaseModal
      :is-open="isRescheduleModalOpen"
      title="Jadwalkan Ulang Janji Temu (Reschedule)"
      max-width="md"
      @close="isRescheduleModalOpen = false"
    >
      <form class="space-y-3" @submit.prevent="submitReschedule">
        <BaseDatePicker
          v-model="rescheduleForm.appointment_date"
          label="Tanggal Baru"
          required
        />

        <div class="grid grid-cols-2 gap-3">
          <BaseInput
            v-model="rescheduleForm.start_time"
            label="Jam Mulai Baru"
            type="time"
            required
          />

          <BaseInput
            v-model="rescheduleForm.end_time"
            label="Jam Selesai Baru"
            type="time"
            required
          />
        </div>

        <BaseInput
          v-model="rescheduleForm.catatan"
          label="Alasan Perubahan Jadwal"
          placeholder="contoh: Orang tua berhalangan hadir di pagi hari"
          required
        />
      </form>

      <template #footer>
        <BaseButton variant="secondary" size="sm" @click="isRescheduleModalOpen = false">
          Batal
        </BaseButton>
        <BaseButton variant="primary" size="sm" :loading="isSubmitting" @click="submitReschedule">
          Simpan Jadwal Baru
        </BaseButton>
      </template>
    </BaseModal>
  </div>
</template>
