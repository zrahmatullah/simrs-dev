<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api/client'
import { useUiStore } from '@/stores/ui'
import { usePatientContextStore } from '@/stores/patientContext'
import DataTable, { type Column } from '@/components/common/DataTable.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseModal from '@/components/common/BaseModal.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import type { Visit, VitalSign } from '@/types'

const uiStore = useUiStore()
const patientContext = usePatientContextStore()

const visits = ref<Visit[]>([])
const isLoading = ref(false)
const search = ref('')
const selectedPoli = ref<number | ''>('')
const poliList = ref<{ value: number; label: string }[]>([])

// Modal Input TTV & Asesmen
const isTTVModalOpen = ref(false)
const modalTab = ref<'ttv' | 'asesmen'>('ttv')
const isSubmitting = ref(false)
const selectedVisit = ref<Visit | null>(null)
const previousVitals = ref<VitalSign[]>([])

const ttvForm = ref({
  visit_id: '',
  td_sistole: 110,
  td_diastole: 70,
  hr: 88,
  rr: 20,
  temp: 36.6,
  spo2: 99,
  bb: 15.0,
  tb: 98.0,
  lk: 49.0,
  skala_nyeri: 0,
  kesadaran: 'Compos Mentis',
  triase_level: 'HIJAU' as 'MERAH' | 'KUNING' | 'HIJAU' | 'HITAM',
  catatan_keperawatan: '',
})

// Detailed Nursing Assessment Form
const asesmenForm = ref({
  keluhan_keperawatan: '',
  riwayat_penyakit_sekarang: '',
  riwayat_alergi: '',
  skala_nyeri_skor: 0,
  skala_nyeri_lokasi: 'Kepala / Leher',
  skala_nyeri_durasi: 'Hilang timbul (< 30 menit)',
  risiko_jatuh_tingkat: 'RENDAH' as 'RENDAH' | 'SEDANG' | 'TINGGI',
  skrining_gizi_turun_bb: false,
  skrining_gizi_nafsu_makan: false,
  masalah_keperawatan: 'Gangguan kenyamanan / Hipertermia / Nyeri ringan',
  rencana_keperawatan: 'Observasi TTV tiap 4 jam, kompres hangat, edukasi minum cukup',
  evaluasi_keperawatan: 'Pasien kooperatif, kondisi terpantau stabil'
})

const columns: Column[] = [
  { key: 'no_antrian', label: 'No. Antrean', width: '120px' },
  { key: 'no_rawat', label: 'No. Rawat', width: '150px' },
  { key: 'no_rm', label: 'No. RM', width: '130px' },
  { key: 'patient_name', label: 'Nama Pasien' },
  { key: 'poli_name', label: 'Poli Dituju' },
  { key: 'keluhan_utama', label: 'Keluhan Masuk' },
  { key: 'status', label: 'Status Saat Ini', align: 'center', width: '150px' },
  { key: 'actions', label: 'Aksi Keperawatan', align: 'right', width: '200px' },
]

async function loadNurseQueue() {
  isLoading.value = true
  try {
    const res = await api.get<Visit[]>('/visits', {
      search: search.value,
      poli_id: selectedPoli.value,
      // Nurse monitors both waiting triage and ongoing patients
    })
    if (res.success && res.data) {
      // Prioritize MENUNGGU_TRIASE at the top
      visits.value = res.data.sort((a, b) => {
        if (a.status === 'MENUNGGU_TRIASE' && b.status !== 'MENUNGGU_TRIASE') return -1
        if (a.status !== 'MENUNGGU_TRIASE' && b.status === 'MENUNGGU_TRIASE') return 1
        return 0
      })
    }
  } catch (err: any) {
    uiStore.showToast(err.message, 'danger')
  } finally {
    isLoading.value = false
  }
}

async function loadPoli() {
  const res = await api.get<any[]>('/master/poli')
  if (res.data) {
    poliList.value = res.data.map(p => ({ value: p.id, label: p.name }))
  }
}

async function openTTVModal(v: Visit) {
  selectedVisit.value = v

  // Set active patient context
  patientContext.setActivePatient({
    id: v.patient_id,
    no_rm: v.no_rm || '',
    nik: '',
    nama_pasien: v.patient_name || '',
    jenis_kelamin: v.patient_gender || '',
    tanggal_lahir: v.patient_dob || '',
    riwayat_alergi: v.patient_allergy,
    status_aktif: true,
    created_at: '',
    updated_at: '',
  }, v)

  ttvForm.value = {
    visit_id: v.id,
    td_sistole: 110,
    td_diastole: 70,
    hr: 88,
    rr: 20,
    temp: 36.6,
    spo2: 99,
    bb: 15.0,
    tb: 98.0,
    lk: 49.0,
    skala_nyeri: 0,
    kesadaran: 'Compos Mentis',
    triase_level: 'HIJAU',
    catatan_keperawatan: '',
  }

  // Load previous vitals
  try {
    const res = await api.get<VitalSign[]>(`/clinical/vitals/patient/${v.patient_id}`)
    if (res.data) {
      previousVitals.value = res.data
    }
  } catch {
    previousVitals.value = []
  }

  // Load existing nursing assessment if any
  try {
    const resAsesmen = await api.get<any>(`/clinical/assessments/visit/${v.id}`)
    if (resAsesmen.data) {
      asesmenForm.value = {
        keluhan_keperawatan: resAsesmen.data.keluhan_keperawatan || '',
        riwayat_penyakit_sekarang: resAsesmen.data.riwayat_penyakit_sekarang || '',
        riwayat_alergi: resAsesmen.data.riwayat_alergi || '',
        skala_nyeri_skor: resAsesmen.data.skala_nyeri_skor || 0,
        skala_nyeri_lokasi: resAsesmen.data.skala_nyeri_lokasi || 'Kepala / Leher',
        skala_nyeri_durasi: resAsesmen.data.skala_nyeri_durasi || 'Hilang timbul (< 30 menit)',
        risiko_jatuh_tingkat: resAsesmen.data.risiko_jatuh_tingkat || 'RENDAH',
        skrining_gizi_turun_bb: resAsesmen.data.skrining_gizi_turun_bb || false,
        skrining_gizi_nafsu_makan: resAsesmen.data.skrining_gizi_nafsu_makan || false,
        masalah_keperawatan: resAsesmen.data.masalah_keperawatan || 'Gangguan rasa nyaman / Nyeri ringan',
        rencana_keperawatan: resAsesmen.data.rencana_keperawatan || 'Observasi TTV teratur dan kolaborasi dokter',
        evaluasi_keperawatan: resAsesmen.data.evaluasi_keperawatan || 'Pasien kooperatif dan telah dilakukan pengkajian awal'
      }
    } else {
      asesmenForm.value.keluhan_keperawatan = v.keluhan_utama || 'Keluhan umum datang berobat'
      asesmenForm.value.riwayat_penyakit_sekarang = 'Pasien mengeluhkan gejala sejak 1-2 hari terakhir'
      asesmenForm.value.riwayat_alergi = v.patient_allergy || 'Tidak ada riwayat alergi yang dilaporkan'
    }
  } catch {
    asesmenForm.value.keluhan_keperawatan = v.keluhan_utama || 'Keluhan umum datang berobat'
    asesmenForm.value.riwayat_penyakit_sekarang = 'Pasien mengeluhkan gejala sejak 1-2 hari terakhir'
    asesmenForm.value.riwayat_alergi = v.patient_allergy || 'Tidak ada riwayat alergi yang dilaporkan'
  }

  modalTab.value = 'ttv'
  isTTVModalOpen.value = true
}

async function submitTTV() {
  if (!selectedVisit.value) return
  isSubmitting.value = true
  try {
    // 1. Submit TTV
    await api.post('/clinical/vitals', ttvForm.value)

    // 2. Submit Detailed Nursing Assessment
    await api.post('/clinical/assessments', {
      visit_id: selectedVisit.value.id,
      keluhan_keperawatan: asesmenForm.value.keluhan_keperawatan,
      riwayat_penyakit_sekarang: asesmenForm.value.riwayat_penyakit_sekarang,
      riwayat_alergi: asesmenForm.value.riwayat_alergi,
      skala_nyeri_skor: Number(ttvForm.value.skala_nyeri || asesmenForm.value.skala_nyeri_skor),
      skala_nyeri_lokasi: asesmenForm.value.skala_nyeri_lokasi,
      skala_nyeri_durasi: asesmenForm.value.skala_nyeri_durasi,
      risiko_jatuh_tingkat: asesmenForm.value.risiko_jatuh_tingkat,
      skrining_gizi_turun_bb: asesmenForm.value.skrining_gizi_turun_bb,
      skrining_gizi_nafsu_makan: asesmenForm.value.skrining_gizi_nafsu_makan,
      masalah_keperawatan: asesmenForm.value.masalah_keperawatan,
      rencana_keperawatan: asesmenForm.value.rencana_keperawatan,
      evaluasi_keperawatan: asesmenForm.value.evaluasi_keperawatan
    })

    uiStore.showToast(`Pengkajian TTV & Asesmen Keperawatan berhasil disimpan! Pasien ${selectedVisit.value?.patient_name} siap diperiksa Dokter.`, 'success')
    isTTVModalOpen.value = false
    loadNurseQueue()
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal menyimpan pengkajian keperawatan', 'danger')
  } finally {
    isSubmitting.value = false
  }
}

onMounted(() => {
  loadNurseQueue()
  loadPoli()
})
</script>

<template>
  <div class="space-y-4 text-left">
    <!-- Header -->
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-[20px] font-bold text-[#0F172A] leading-[28px]">
          Stasiun Perawat: Pengkajian Awal & Tanda-Tanda Vital (TTV)
        </h1>
        <p class="text-[13px] text-[#475569]">
          Antrean triase perawat, input parameter TTV (Tekanan Darah, Nadi, Suhu, RR, SpO2, Antropometri Anak, Skala Nyeri), dan triase klinis.
        </p>
      </div>

      <BaseButton variant="secondary" size="sm" @click="loadNurseQueue">
        Segarkan Antrean
      </BaseButton>
    </div>

    <!-- Filters Bar -->
    <div class="bg-white p-3 rounded-lg border border-[#E2E8F0] shadow-sm flex flex-wrap items-center gap-3">
      <div class="flex-1 min-w-[200px]">
        <BaseInput
          v-model="search"
          placeholder="Cari No. Antrean, Nama Pasien, No. RM..."
          compact
          @keyup.enter="loadNurseQueue"
        />
      </div>

      <div class="w-[200px]">
        <BaseSelect
          v-model="selectedPoli"
          :options="[{ value: '', label: 'Semua Poli' }, ...poliList]"
          placeholder="Filter Poli"
          compact
          @change="loadNurseQueue"
        />
      </div>

      <BaseButton variant="secondary" size="sm" @click="loadNurseQueue">
        Filter
      </BaseButton>
    </div>

    <!-- Antrean DataTable -->
    <DataTable
      :columns="columns"
      :data="visits"
      :loading="isLoading"
      empty-text="Tidak ada pasien dalam antrean perawat"
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
        <div v-if="row.patient_allergy" class="text-[11px] text-[#DC2626] font-semibold">
          ⚠️ ALERGI: {{ row.patient_allergy }}
        </div>
      </template>

      <template #status="{ row }">
        <StatusBadge :status="row.status" />
      </template>

      <template #actions="{ row }">
        <div class="flex items-center justify-end gap-1.5">
          <BaseButton
            :variant="row.status === 'MENUNGGU_TRIASE' ? 'primary' : 'soft'"
            size="sm"
            @click="openTTVModal(row)"
          >
            {{ row.status === 'MENUNGGU_TRIASE' ? 'Input TTV & Triase' : 'Lihat / Update TTV' }}
          </BaseButton>
        </div>
      </template>
    </DataTable>

    <!-- Modal Form TTV & Asesmen Detail -->
    <BaseModal
      :is-open="isTTVModalOpen"
      :title="`Pengkajian Perawat: ${selectedVisit?.patient_name} (${selectedVisit?.no_rm})`"
      max-width="3xl"
      @close="isTTVModalOpen = false"
    >
      <!-- Modal Navigation Tabs -->
      <div class="flex items-center gap-2 border-b border-[#E2E8F0] pb-2 mb-3">
        <button
          type="button"
          :class="[
            'px-3 py-1.5 rounded-[6px] text-[13px] font-bold transition-all',
            modalTab === 'ttv' ? 'bg-[#0284C7] text-white shadow-2xs' : 'text-[#64748B] hover:bg-[#F1F5F9]'
          ]"
          @click="modalTab = 'ttv'"
        >
          1. Tanda-Tanda Vital & Triase (TTV)
        </button>

        <button
          type="button"
          :class="[
            'px-3 py-1.5 rounded-[6px] text-[13px] font-bold transition-all',
            modalTab === 'asesmen' ? 'bg-[#0284C7] text-white shadow-2xs' : 'text-[#64748B] hover:bg-[#F1F5F9]'
          ]"
          @click="modalTab = 'asesmen'"
        >
          2. Form Asesmen Keperawatan Detail
        </button>
      </div>

      <!-- Patient Mini Bar -->
      <div class="bg-[#F8FAFC] p-3 rounded-lg border border-[#E2E8F0] flex flex-wrap items-center justify-between gap-2 text-[12px] mb-3">
        <div><span class="text-[#64748B]">No. Rawat:</span> <span class="font-mono font-semibold">{{ selectedVisit?.no_rawat }}</span></div>
        <div><span class="text-[#64748B]">Poli Tujuan:</span> <span class="font-semibold text-[#0284C7]">{{ selectedVisit?.poli_name }}</span></div>
        <div><span class="text-[#64748B]">Penjamin:</span> <span class="font-semibold">{{ selectedVisit?.guarantor_name }}</span></div>
      </div>

      <!-- TAB 1: TTV & Triase -->
      <form v-show="modalTab === 'ttv'" class="space-y-4" @submit.prevent="submitTTV">
        <!-- 1. Tanda Tanda Vital (TTV) Grid -->
        <div>
          <h4 class="text-[14px] font-semibold text-[#0F172A] mb-2 flex items-center gap-1.5">
            <span>Parameter Tanda-Tanda Vital Utama</span>
          </h4>

          <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 bg-white p-3 rounded-lg border border-[#E2E8F0]">
            <div>
              <BaseInput
                v-model.number="ttvForm.td_sistole"
                label="TD Sistole (mmHg)"
                type="number"
                placeholder="110"
              />
            </div>
            <div>
              <BaseInput
                v-model.number="ttvForm.td_diastole"
                label="TD Diastole (mmHg)"
                type="number"
                placeholder="70"
              />
            </div>
            <div>
              <BaseInput
                v-model.number="ttvForm.hr"
                label="Nadi / HR (x/mnt)"
                type="number"
                placeholder="88"
              />
            </div>
            <div>
              <BaseInput
                v-model.number="ttvForm.rr"
                label="Laju Napas / RR (x/mnt)"
                type="number"
                placeholder="20"
              />
            </div>
            <div>
              <BaseInput
                v-model.number="ttvForm.temp"
                label="Suhu Tubuh (°C)"
                type="number"
                step="0.1"
                placeholder="36.5"
                required
              />
            </div>
            <div>
              <BaseInput
                v-model.number="ttvForm.spo2"
                label="SpO2 Saturasi (%)"
                type="number"
                placeholder="99"
              />
            </div>
            <div>
              <BaseSelect
                v-model="ttvForm.kesadaran"
                label="Tingkat Kesadaran"
                :options="[
                  { value: 'Compos Mentis', label: 'Compos Mentis (Sadar Penuh)' },
                  { value: 'Apatis', label: 'Apatis (Acuh tak acuh)' },
                  { value: 'Somnolen', label: 'Somnolen (Mengantuk)' },
                  { value: 'Sopor', label: 'Sopor' },
                  { value: 'Koma', label: 'Koma' }
                ]"
              />
            </div>
            <div>
              <BaseSelect
                v-model.number="ttvForm.skala_nyeri"
                label="Skala Nyeri (0-10)"
                :options="[
                  { value: 0, label: '0 - Tidak Nyeri' },
                  { value: 2, label: '2 - Ringan' },
                  { value: 4, label: '4 - Sedang' },
                  { value: 6, label: '6 - Sangat Mengganggu' },
                  { value: 8, label: '8 - Hebat' },
                  { value: 10, label: '10 - Nyeri Terberat' }
                ]"
              />
            </div>
          </div>
        </div>

        <!-- 2. Antropometri Anak & Pediatrik -->
        <div>
          <h4 class="text-[14px] font-semibold text-[#0F172A] mb-2">
            Pertumbuhan & Antropometri (Pediatrik / Anak)
          </h4>

          <div class="grid grid-cols-3 gap-3 bg-white p-3 rounded-lg border border-[#E2E8F0]">
            <div>
              <BaseInput
                v-model.number="ttvForm.bb"
                label="Berat Badan (kg)"
                type="number"
                step="0.01"
                placeholder="15.5"
                required
              />
            </div>
            <div>
              <BaseInput
                v-model.number="ttvForm.tb"
                label="Tinggi / Panjang (cm)"
                type="number"
                step="0.1"
                placeholder="98.0"
                required
              />
            </div>
            <div>
              <BaseInput
                v-model.number="ttvForm.lk"
                label="Lingkar Kepala (cm)"
                type="number"
                step="0.1"
                placeholder="48.5"
              />
            </div>
          </div>
        </div>

        <!-- 3. Triase Klinis & Catatan Keperawatan -->
        <div>
          <h4 class="text-[14px] font-semibold text-[#0F172A] mb-2">
            Triase Pasien & Catatan Observasi Keperawatan
          </h4>

          <div class="space-y-3 bg-white p-3 rounded-lg border border-[#E2E8F0]">
            <div>
              <label class="text-[13px] font-medium text-[#334155] block mb-1">
                Kategori Triase Klinis <span class="text-[#DC2626] font-bold">*</span>
              </label>
              <div class="grid grid-cols-4 gap-2">
                <label
                  :class="[
                    'p-2.5 rounded-lg border text-center cursor-pointer transition-all flex flex-col items-center justify-center',
                    ttvForm.triase_level === 'HIJAU' ? 'bg-[#DCFCE7] border-[#15803D] text-[#15803D] font-bold ring-2 ring-[#86EFAC]' : 'border-[#CBD5E1] hover:bg-[#F8FAFC]'
                  ]"
                >
                  <input type="radio" v-model="ttvForm.triase_level" value="HIJAU" class="sr-only" />
                  <span class="text-[13px]">HIJAU</span>
                  <span class="text-[10px] font-normal">Tidak Gawat / Rutin</span>
                </label>

                <label
                  :class="[
                    'p-2.5 rounded-lg border text-center cursor-pointer transition-all flex flex-col items-center justify-center',
                    ttvForm.triase_level === 'KUNING' ? 'bg-[#FEF3C7] border-[#B45309] text-[#B45309] font-bold ring-2 ring-[#FCD34D]' : 'border-[#CBD5E1] hover:bg-[#F8FAFC]'
                  ]"
                >
                  <input type="radio" v-model="ttvForm.triase_level" value="KUNING" class="sr-only" />
                  <span class="text-[13px]">KUNING</span>
                  <span class="text-[10px] font-normal">Mendesak / Observasi</span>
                </label>

                <label
                  :class="[
                    'p-2.5 rounded-lg border text-center cursor-pointer transition-all flex flex-col items-center justify-center',
                    ttvForm.triase_level === 'MERAH' ? 'bg-[#FEE2E2] border-[#B91C1C] text-[#B91C1C] font-bold ring-2 ring-[#FCA5A5]' : 'border-[#CBD5E1] hover:bg-[#F8FAFC]'
                  ]"
                >
                  <input type="radio" v-model="ttvForm.triase_level" value="MERAH" class="sr-only" />
                  <span class="text-[13px]">MERAH</span>
                  <span class="text-[10px] font-normal">Gawat Darurat Segera</span>
                </label>

                <label
                  :class="[
                    'p-2.5 rounded-lg border text-center cursor-pointer transition-all flex flex-col items-center justify-center',
                    ttvForm.triase_level === 'HITAM' ? 'bg-[#334155] border-black text-white font-bold ring-2 ring-slate-400' : 'border-[#CBD5E1] hover:bg-[#F8FAFC]'
                  ]"
                >
                  <input type="radio" v-model="ttvForm.triase_level" value="HITAM" class="sr-only" />
                  <span class="text-[13px]">HITAM</span>
                  <span class="text-[10px] font-normal">Meninggal Dunia</span>
                </label>
              </div>
            </div>

            <div>
              <label class="text-[13px] font-medium text-[#334155] block mb-1">
                Catatan Pengkajian Keperawatan
              </label>
              <textarea
                v-model="ttvForm.catatan_keperawatan"
                rows="2"
                placeholder="contoh: Pasien sadar, anak aktif, tidak ada retraksi dinding dada, tidak ada dehidrasi berat..."
                class="w-full p-2.5 text-[13px] text-[#0F172A] bg-white border border-[#CBD5E1] rounded-[6px] focus:border-[#0284C7] outline-none"
              ></textarea>
            </div>
          </div>
        </div>
      </form>

      <!-- TAB 2: Form Asesmen Keperawatan Detail -->
      <div v-show="modalTab === 'asesmen'" class="space-y-4">
        <!-- Anamnesa & Alergi -->
        <div class="bg-white p-3 rounded-lg border border-[#E2E8F0] space-y-3">
          <h4 class="text-[14px] font-semibold text-[#0F172A]">Anamnesa Keperawatan & Riwayat Alergi</h4>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div>
              <label class="text-[12px] font-medium text-[#334155] block mb-1">Keluhan Keperawatan Pasien</label>
              <textarea
                v-model="asesmenForm.keluhan_keperawatan"
                rows="2"
                class="w-full p-2 text-[12px] border border-[#CBD5E1] rounded-[6px] outline-none focus:border-[#0284C7]"
                placeholder="Keluhan keperawatan yang dirasakan pasien..."
              ></textarea>
            </div>

            <div>
              <label class="text-[12px] font-medium text-[#334155] block mb-1">Riwayat Alergi (Obat / Makanan / Lainnya)</label>
              <textarea
                v-model="asesmenForm.riwayat_alergi"
                rows="2"
                class="w-full p-2 text-[12px] border border-[#CBD5E1] rounded-[6px] outline-none focus:border-[#0284C7]"
                placeholder="Sebutkan riwayat alergi jika ada..."
              ></textarea>
            </div>
          </div>
        </div>

        <!-- Skrining Nyeri & Risiko Jatuh -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div class="bg-white p-3 rounded-lg border border-[#E2E8F0] space-y-2">
            <h4 class="text-[13px] font-semibold text-[#0F172A]">Skrining Nyeri Komprehensif</h4>
            <div>
              <label class="text-[11px] font-medium text-[#475569] block mb-1">Lokasi Nyeri</label>
              <BaseInput v-model="asesmenForm.skala_nyeri_lokasi" placeholder="Kepala / Perut / Dada / Ekstremitas" compact />
            </div>
            <div>
              <label class="text-[11px] font-medium text-[#475569] block mb-1">Durasi / Frekuensi Nyeri</label>
              <BaseInput v-model="asesmenForm.skala_nyeri_durasi" placeholder="Terus-menerus / Hilang timbul" compact />
            </div>
          </div>

          <div class="bg-white p-3 rounded-lg border border-[#E2E8F0] space-y-2">
            <h4 class="text-[13px] font-semibold text-[#0F172A]">Skrining Risiko Jatuh (Morse / Humpty Dumpty)</h4>
            <div>
              <label class="text-[11px] font-medium text-[#475569] block mb-1">Tingkat Risiko Jatuh</label>
              <BaseSelect
                v-model="asesmenForm.risiko_jatuh_tingkat"
                :options="[
                  { value: 'RENDAH', label: 'Risiko Rendah (0 - 24)' },
                  { value: 'SEDANG', label: 'Risiko Sedang (25 - 44)' },
                  { value: 'TINGGI', label: 'Risiko Tinggi (>= 45)' }
                ]"
                compact
              />
            </div>
            <div class="text-[11px] text-[#64748B] pt-1">
              Catatan: Pasang gelang kuning pengaman jika pasien berisiko tinggi jatuh.
            </div>
          </div>
        </div>

        <!-- Skrining Gizi MST (Malnutrition Screening Tool) -->
        <div class="bg-white p-3 rounded-lg border border-[#E2E8F0] space-y-2">
          <h4 class="text-[13px] font-semibold text-[#0F172A]">Skrining Gizi Sederhana (MST)</h4>
          <div class="space-y-1.5 text-[12px]">
            <label class="flex items-center gap-2 cursor-pointer">
              <input v-model="asesmenForm.skrining_gizi_turun_bb" type="checkbox" class="rounded text-[#0284C7]" />
              <span>Apakah pasien mengalami penurunan berat badan yang tidak diinginkan dalam 6 bulan terakhir?</span>
            </label>
            <label class="flex items-center gap-2 cursor-pointer">
              <input v-model="asesmenForm.skrining_gizi_nafsu_makan" type="checkbox" class="rounded text-[#0284C7]" />
              <span>Apakah asupan makan berkurang karena berkurangnya nafsu makan?</span>
            </label>
          </div>
        </div>

        <!-- Masalah & Rencana Keperawatan -->
        <div class="bg-white p-3 rounded-lg border border-[#E2E8F0] space-y-3">
          <h4 class="text-[14px] font-semibold text-[#0F172A]">Masalah & Rencana Asuhan Keperawatan</h4>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div>
              <label class="text-[12px] font-medium text-[#334155] block mb-1">Masalah Keperawatan (SDKI / NANDA)</label>
              <BaseInput v-model="asesmenForm.masalah_keperawatan" placeholder="Nyeri akut, Hipertermia, Bersihan jalan napas..." compact />
            </div>

            <div>
              <label class="text-[12px] font-medium text-[#334155] block mb-1">Rencana Tindakan Keperawatan (SIKI)</label>
              <BaseInput v-model="asesmenForm.rencana_keperawatan" placeholder="Observasi TTV, posisikan semi-fowler, edukasi..." compact />
            </div>
          </div>

          <div>
            <label class="text-[12px] font-medium text-[#334155] block mb-1">Evaluasi Keperawatan Awal</label>
            <BaseInput v-model="asesmenForm.evaluasi_keperawatan" placeholder="Kondisi terkontrol, pasien kooperatif..." compact />
          </div>
        </div>
      </div>

      <template #footer>
        <BaseButton variant="secondary" size="md" @click="isTTVModalOpen = false">
          Batal
        </BaseButton>
        <BaseButton variant="primary" size="md" :loading="isSubmitting" @click="submitTTV">
          Simpan TTV & Teruskan ke Dokter
        </BaseButton>
      </template>
    </BaseModal>
  </div>
</template>
