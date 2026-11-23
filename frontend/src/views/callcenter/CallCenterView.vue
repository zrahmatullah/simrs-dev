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
import {
  Calendar, Clock, UserX, Plus, RefreshCw, Trash2, CheckCircle2,
  AlertTriangle, UserCheck, CalendarDays, UserMinus, ShieldCheck
} from 'lucide-vue-next'

const uiStore = useUiStore()
const patientContext = usePatientContextStore()

type CallCenterTab = 'appointments' | 'schedules' | 'leaves'
const activeTab = ref<CallCenterTab>('appointments')

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

// ----------------------------------------------------
// TAB 2: JADWAL PRAKTIK & KUOTA DOKTER (FASE 1)
// ----------------------------------------------------
const doctorSchedules = ref<any[]>([])
const isLoadingSchedules = ref(false)
const isCreateScheduleModalOpen = ref(false)
const isSubmittingSchedule = ref(false)

const dayNames: Record<number, string> = {
  1: 'Senin',
  2: 'Selasa',
  3: 'Rabu',
  4: 'Kamis',
  5: 'Jumat',
  6: 'Sabtu',
  7: 'Minggu'
}

const scheduleForm = ref({
  doctor_id: '',
  poli_id: 2,
  day_of_week: 1,
  start_time: '08:00',
  end_time: '12:00',
  max_quota: 25
})

async function loadSchedules() {
  isLoadingSchedules.value = true
  try {
    const res = await api.get<any[]>('/appointments/schedules')
    if (res.data) {
      doctorSchedules.value = res.data
    }
  } catch (err: any) {
    console.warn('Gagal memuat jadwal dokter', err)
  } finally {
    isLoadingSchedules.value = false
  }
}

function openCreateScheduleModal() {
  scheduleForm.value = {
    doctor_id: doctorOptions.value[0]?.value || '',
    poli_id: poliOptions.value[0]?.value || 2,
    day_of_week: 1,
    start_time: '08:00',
    end_time: '12:00',
    max_quota: 25
  }
  isCreateScheduleModalOpen.value = true
}

async function submitSchedule() {
  if (!scheduleForm.value.doctor_id) {
    uiStore.showToast('Pilih dokter terlebih dahulu', 'warning')
    return
  }
  isSubmittingSchedule.value = true
  try {
    await api.post('/appointments/schedules', {
      doctor_id: scheduleForm.value.doctor_id,
      poli_id: Number(scheduleForm.value.poli_id),
      day_of_week: Number(scheduleForm.value.day_of_week),
      start_time: scheduleForm.value.start_time,
      end_time: scheduleForm.value.end_time,
      max_quota: Number(scheduleForm.value.max_quota)
    })
    uiStore.showToast('Jadwal dokter berhasil ditambahkan!', 'success')
    isCreateScheduleModalOpen.value = false
    loadSchedules()
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal menyimpan jadwal dokter', 'danger')
  } finally {
    isSubmittingSchedule.value = false
  }
}

function deleteSchedule(sc: any) {
  uiStore.openConfirm({
    title: 'Hapus Jadwal Praktik',
    message: `Hapus jadwal praktik ${sc.doctor_name} hari ${dayNames[sc.day_of_week]} (${sc.start_time} - ${sc.end_time})?`,
    confirmText: 'Hapus Jadwal',
    variant: 'danger',
    onConfirm: async () => {
      try {
        await api.delete(`/appointments/schedules/${sc.id}`)
        uiStore.showToast('Jadwal praktik berhasil dihapus', 'success')
        loadSchedules()
      } catch (err: any) {
        uiStore.showToast(err.message || 'Gagal menghapus jadwal', 'danger')
      }
    }
  })
}

// ----------------------------------------------------
// TAB 3: CUTI & PENGGANTI DOKTER (FASE 1)
// ----------------------------------------------------
const doctorLeaves = ref<any[]>([])
const isLoadingLeaves = ref(false)
const isCreateLeaveModalOpen = ref(false)
const isSubmittingLeave = ref(false)

const leaveForm = ref({
  doctor_id: '',
  start_date: new Date().toISOString().split('T')[0],
  end_date: new Date().toISOString().split('T')[0],
  reason: 'Cuti Tahunan / Tugas Simposium Medis',
  substitute_doctor_id: ''
})

async function loadLeaves() {
  isLoadingLeaves.value = true
  try {
    const res = await api.get<any[]>('/appointments/leaves')
    if (res.data) {
      doctorLeaves.value = res.data
    }
  } catch (err: any) {
    console.warn('Gagal memuat cuti dokter', err)
  } finally {
    isLoadingLeaves.value = false
  }
}

function openCreateLeaveModal() {
  leaveForm.value = {
    doctor_id: doctorOptions.value[0]?.value || '',
    start_date: new Date().toISOString().split('T')[0],
    end_date: new Date().toISOString().split('T')[0],
    reason: 'Cuti Tahunan / Simposium Kedokteran',
    substitute_doctor_id: doctorOptions.value[1]?.value || ''
  }
  isCreateLeaveModalOpen.value = true
}

async function submitLeave() {
  if (!leaveForm.value.doctor_id) {
    uiStore.showToast('Pilih dokter yang mengajukan cuti', 'warning')
    return
  }
  isSubmittingLeave.value = true
  try {
    await api.post('/appointments/leaves', {
      doctor_id: leaveForm.value.doctor_id,
      start_date: leaveForm.value.start_date,
      end_date: leaveForm.value.end_date,
      reason: leaveForm.value.reason,
      substitute_doctor_id: leaveForm.value.substitute_doctor_id ? leaveForm.value.substitute_doctor_id : undefined
    })
    uiStore.showToast('Permohonan cuti dokter berhasil disimpan!', 'success')
    isCreateLeaveModalOpen.value = false
    loadLeaves()
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal menyimpan cuti dokter', 'danger')
  } finally {
    isSubmittingLeave.value = false
  }
}

onMounted(() => {
  loadAppointments()
  loadOptions()
  loadSchedules()
  loadLeaves()
})
</script>

<template>
  <div class="space-y-4 text-left">
    <!-- Header -->
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-[20px] font-bold text-[#0F172A] leading-[28px]">
          Call Center & Penjadwalan Praktik Dokter
        </h1>
        <p class="text-[13px] text-[#475569]">
          Pemesanan janji temu via WhatsApp/Telepon, manajemen jadwal rutin poli, kuota sesi periksa, dan cuti dokter.
        </p>
      </div>

      <div class="flex items-center gap-2">
        <BaseButton
          v-if="activeTab === 'appointments'"
          variant="primary"
          size="md"
          @click="openCreateModal"
        >
          <Plus class="w-4 h-4 mr-1.5" />
          Jadwalkan Janji Temu
        </BaseButton>
        <BaseButton
          v-else-if="activeTab === 'schedules'"
          variant="primary"
          size="md"
          @click="openCreateScheduleModal"
        >
          <Plus class="w-4 h-4 mr-1.5" />
          Tambah Jadwal Praktik
        </BaseButton>
        <BaseButton
          v-else-if="activeTab === 'leaves'"
          variant="primary"
          size="md"
          class="!bg-amber-600 hover:!bg-amber-700"
          @click="openCreateLeaveModal"
        >
          <UserMinus class="w-4 h-4 mr-1.5" />
          Pengajuan Cuti Dokter
        </BaseButton>
      </div>
    </div>

    <!-- 3 Call Center Tabs Navigation -->
    <div class="flex items-center gap-2 border-b border-[#E2E8F0] pb-2">
      <button
        v-for="t in [
          { id: 'appointments', label: '1. Janji Temu Pasien', icon: Calendar },
          { id: 'schedules', label: '2. Jadwal Praktik & Kuota Dokter', icon: Clock },
          { id: 'leaves', label: '3. Cuti & Dokter Pengganti', icon: UserMinus }
        ]"
        :key="t.id"
        type="button"
        :class="[
          'px-3.5 py-2 rounded-[6px] text-[13px] font-bold transition-all flex items-center gap-2',
          activeTab === t.id
            ? 'bg-[#0284C7] text-white shadow-xs'
            : 'text-[#475569] hover:bg-[#F1F5F9] hover:text-[#0F172A]'
        ]"
        @click="activeTab = t.id as CallCenterTab"
      >
        <component :is="t.icon" class="w-4 h-4" />
        {{ t.label }}
      </button>
    </div>

    <!-- ========================================================================= -->
    <!-- TAB 1: JANJI TEMU PASIEN                                                  -->
    <!-- ========================================================================= -->
    <div v-if="activeTab === 'appointments'" class="space-y-4">
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
    </div>

    <!-- ========================================================================= -->
    <!-- TAB 2: JADWAL PRAKTIK & KUOTA DOKTER                                      -->
    <!-- ========================================================================= -->
    <div v-if="activeTab === 'schedules'" class="space-y-4">
      <div class="bg-white p-3.5 rounded-lg border border-[#E2E8F0] shadow-sm flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div class="flex items-center gap-2">
          <Clock class="w-4 h-4 text-[#0284C7]" />
          <span class="text-[13px] font-bold text-[#0F172A]">Daftar Jadwal Praktik Mingguan & Batas Kuota Pasien</span>
        </div>
        <div class="flex items-center gap-2">
          <BaseButton variant="outline" size="sm" @click="loadSchedules">
            <RefreshCw class="w-3.5 h-3.5 mr-1" />
            Segarkan
          </BaseButton>
          <BaseButton variant="primary" size="sm" @click="openCreateScheduleModal">
            <Plus class="w-4 h-4 mr-1" />
            Tambah Jadwal Praktik
          </BaseButton>
        </div>
      </div>

      <div class="bg-white rounded-lg border border-[#E2E8F0] shadow-sm overflow-hidden">
        <table class="w-full text-left text-[12px]">
          <thead class="bg-[#F8FAFC] text-[#475569] border-b font-bold">
            <tr>
              <th class="p-3">Hari Praktik</th>
              <th class="p-3">Nama Dokter</th>
              <th class="p-3">Poliklinik</th>
              <th class="p-3 text-center">Jam Praktik</th>
              <th class="p-3 text-center">Maks. Kuota / Sesi</th>
              <th class="p-3 text-center">Status Sesi</th>
              <th class="p-3 text-right">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[#F1F5F9]">
            <tr v-if="doctorSchedules.length === 0">
              <td colspan="7" class="p-8 text-center text-[#94A3B8]">
                Belum ada jadwal praktik dokter yang tersimpan.
              </td>
            </tr>
            <tr v-for="sc in doctorSchedules" :key="sc.id" class="hover:bg-[#F8FAFC]">
              <td class="p-3 font-bold text-[#0284C7]">
                <span class="px-2.5 py-1 rounded bg-[#E0F2FE] border border-[#BAE6FD]">
                  {{ dayNames[sc.day_of_week] || `Hari ${sc.day_of_week}` }}
                </span>
              </td>
              <td class="p-3 font-semibold text-[#0F172A]">{{ sc.doctor_name }}</td>
              <td class="p-3 text-[#475569]">{{ sc.poli_name }}</td>
              <td class="p-3 text-center font-mono font-bold text-[#334155]">
                {{ sc.start_time.substring(0, 5) }} - {{ sc.end_time.substring(0, 5) }}
              </td>
              <td class="p-3 text-center font-mono font-bold text-emerald-700">
                <span class="px-2 py-0.5 rounded bg-emerald-50 border border-emerald-200">
                  {{ sc.max_quota }} Pasien
                </span>
              </td>
              <td class="p-3 text-center">
                <span class="px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-100 text-emerald-800">
                  {{ sc.is_active ? 'AKTIF' : 'NONAKTIF' }}
                </span>
              </td>
              <td class="p-3 text-right">
                <button
                  class="text-rose-600 hover:text-rose-800 p-1.5 rounded hover:bg-rose-50 transition"
                  title="Hapus Jadwal"
                  @click="deleteSchedule(sc)"
                >
                  <Trash2 class="w-4 h-4" />
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- ========================================================================= -->
    <!-- TAB 3: CUTI & DOKTER PENGGANTI                                            -->
    <!-- ========================================================================= -->
    <div v-if="activeTab === 'leaves'" class="space-y-4">
      <div class="bg-white p-3.5 rounded-lg border border-[#E2E8F0] shadow-sm flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div class="flex items-center gap-2">
          <UserMinus class="w-4 h-4 text-amber-600" />
          <span class="text-[13px] font-bold text-[#0F172A]">Daftar Libur, Izin, & Cuti Dokter dengan Pelimpahan Pasien</span>
        </div>
        <div class="flex items-center gap-2">
          <BaseButton variant="outline" size="sm" @click="loadLeaves">
            <RefreshCw class="w-3.5 h-3.5 mr-1" />
            Segarkan
          </BaseButton>
          <BaseButton variant="primary" size="sm" class="!bg-amber-600 hover:!bg-amber-700" @click="openCreateLeaveModal">
            <Plus class="w-4 h-4 mr-1" />
            Pengajuan Cuti Dokter
          </BaseButton>
        </div>
      </div>

      <div class="bg-white rounded-lg border border-[#E2E8F0] shadow-sm overflow-hidden">
        <table class="w-full text-left text-[12px]">
          <thead class="bg-[#F8FAFC] text-[#475569] border-b font-bold">
            <tr>
              <th class="p-3">Dokter yang Mengajukan</th>
              <th class="p-3">Tanggal Mulai Cuti</th>
              <th class="p-3">Tanggal Selesai</th>
              <th class="p-3">Alasan / Keterangan</th>
              <th class="p-3">Dokter Pengganti (Substitute)</th>
              <th class="p-3 text-center">Status</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[#F1F5F9]">
            <tr v-if="doctorLeaves.length === 0">
              <td colspan="6" class="p-8 text-center text-[#94A3B8]">
                Belum ada permohonan cuti dokter yang tercatat.
              </td>
            </tr>
            <tr v-for="lv in doctorLeaves" :key="lv.id" class="hover:bg-[#F8FAFC]">
              <td class="p-3 font-semibold text-[#0F172A]">{{ lv.doctor_name }}</td>
              <td class="p-3 text-rose-700 font-medium">{{ lv.start_date }}</td>
              <td class="p-3 text-rose-700 font-medium">{{ lv.end_date }}</td>
              <td class="p-3 text-[#475569] italic">{{ lv.reason }}</td>
              <td class="p-3">
                <span v-if="lv.substitute_doctor_name" class="font-medium text-[#0284C7] bg-[#F0F9FF] px-2 py-0.5 rounded border border-[#BAE6FD]">
                  {{ lv.substitute_doctor_name }}
                </span>
                <span v-else class="text-[#94A3B8] italic">-</span>
              </td>
              <td class="p-3 text-center">
                <span class="px-2 py-0.5 rounded text-[10px] font-bold bg-amber-100 text-amber-800">
                  {{ lv.status || 'DISETUJUI' }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

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

    <!-- Modal Tambah Jadwal Praktik Dokter -->
    <BaseModal
      :is-open="isCreateScheduleModalOpen"
      title="Tambah Jadwal Praktik Rutin Dokter"
      max-width="md"
      @close="isCreateScheduleModalOpen = false"
    >
      <form class="space-y-3" @submit.prevent="submitSchedule">
        <BaseSelect
          v-model="scheduleForm.doctor_id"
          label="Pilih Dokter Praktik *"
          :options="doctorOptions"
          required
        />

        <BaseSelect
          v-model="scheduleForm.poli_id"
          label="Poli Pelayanan *"
          :options="poliOptions"
          required
        />

        <BaseSelect
          v-model="scheduleForm.day_of_week"
          label="Hari Praktik *"
          :options="[
            { value: 1, label: 'Senin' },
            { value: 2, label: 'Selasa' },
            { value: 3, label: 'Rabu' },
            { value: 4, label: 'Kamis' },
            { value: 5, label: 'Jumat' },
            { value: 6, label: 'Sabtu' },
            { value: 7, label: 'Minggu' }
          ]"
          required
        />

        <div class="grid grid-cols-2 gap-3">
          <BaseInput
            v-model="scheduleForm.start_time"
            label="Jam Mulai"
            type="time"
            required
          />
          <BaseInput
            v-model="scheduleForm.end_time"
            label="Jam Selesai"
            type="time"
            required
          />
        </div>

        <BaseInput
          v-model.number="scheduleForm.max_quota"
          type="number"
          label="Batas Maksimal Kuota Pasien / Sesi"
          placeholder="25"
          required
        />
      </form>

      <template #footer>
        <BaseButton variant="secondary" size="sm" @click="isCreateScheduleModalOpen = false">
          Batal
        </BaseButton>
        <BaseButton variant="primary" size="sm" :loading="isSubmittingSchedule" @click="submitSchedule">
          Simpan Jadwal Dokter
        </BaseButton>
      </template>
    </BaseModal>

    <!-- Modal Permohonan Cuti Dokter -->
    <BaseModal
      :is-open="isCreateLeaveModalOpen"
      title="Pengajuan Izin / Cuti Praktik Dokter"
      max-width="md"
      @close="isCreateLeaveModalOpen = false"
    >
      <form class="space-y-3" @submit.prevent="submitLeave">
        <BaseSelect
          v-model="leaveForm.doctor_id"
          label="Dokter yang Mengajukan Cuti *"
          :options="doctorOptions"
          required
        />

        <div class="grid grid-cols-2 gap-3">
          <BaseDatePicker
            v-model="leaveForm.start_date"
            label="Tanggal Mulai *"
            required
          />
          <BaseDatePicker
            v-model="leaveForm.end_date"
            label="Tanggal Selesai *"
            required
          />
        </div>

        <BaseSelect
          v-model="leaveForm.substitute_doctor_id"
          label="Dokter Pengganti (Substitute / Pelimpahan)"
          :options="[
            { value: '', label: '-- Tanpa Dokter Pengganti (Poli Libur) --' },
            ...doctorOptions
          ]"
        />

        <BaseInput
          v-model="leaveForm.reason"
          label="Alasan / Keperluan Cuti *"
          placeholder="Misal: Cuti tahunan, simposium klinis IDI, dll."
          required
        />
      </form>

      <template #footer>
        <BaseButton variant="secondary" size="sm" @click="isCreateLeaveModalOpen = false">
          Batal
        </BaseButton>
        <BaseButton variant="primary" size="sm" class="!bg-amber-600 hover:!bg-amber-700" :loading="isSubmittingLeave" @click="submitLeave">
          Kirim Pengajuan Cuti
        </BaseButton>
      </template>
    </BaseModal>
  </div>
</template>
