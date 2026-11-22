<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api/client'
import { useQueueStore } from '@/stores/queue'
import DataTable, { type Column } from '@/components/common/DataTable.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import type { Visit } from '@/types'

const queueStore = useQueueStore()
const recentVisits = ref<Visit[]>([])
const isLoading = ref(false)
const totalPatients = ref(0)
const totalUsers = ref(0)

const columns: Column[] = [
  { key: 'no_antrian', label: 'No. Antrean', width: '120px' },
  { key: 'no_rawat', label: 'No. Rawat', width: '160px' },
  { key: 'no_rm', label: 'No. RM', width: '130px' },
  { key: 'patient_name', label: 'Nama Pasien' },
  { key: 'poli_name', label: 'Poli Unit' },
  { key: 'doctor_name', label: 'Dokter Pemeriksa' },
  { key: 'guarantor_name', label: 'Penjamin' },
  { key: 'status', label: 'Status Antrean', align: 'center', width: '140px' },
]

async function loadDashboardData() {
  isLoading.value = true
  try {
    await queueStore.fetchQueueMetrics()

    const visitsRes = await api.get<Visit[]>('/visits', { limit: 10 })
    if (visitsRes.success && visitsRes.data) {
      recentVisits.value = visitsRes.data
    }

    const patRes = await api.get('/patients', { limit: 1 })
    if (patRes.meta) {
      totalPatients.value = patRes.meta.total_data
    }

    const userRes = await api.get('/users', { limit: 1 })
    if (userRes.meta) {
      totalUsers.value = userRes.meta.total_data
    }
  } catch (err) {
    console.error('Failed to load admin stats:', err)
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  loadDashboardData()
})
</script>

<template>
  <div class="space-y-6 text-left">
    <!-- Page Header (Specified: 20px Bold #0F172A) -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-[20px] font-bold text-[#0F172A] leading-[28px]">
          Dashboard Eksekutif & Manajemen SIMRS
        </h1>
        <p class="text-[13px] text-[#475569]">
          Ringkasan operasional klinik, antrean pasien real-time, dan statistik sistem.
        </p>
      </div>

      <div class="flex items-center gap-2">
        <BaseButton variant="secondary" size="sm" @click="loadDashboardData">
          Segarkan Data
        </BaseButton>
        <router-link to="/registration">
          <BaseButton variant="primary" size="sm">
            + Kunjungan Baru
          </BaseButton>
        </router-link>
      </div>
    </div>

    <!-- Quick Stats Cards (4 Cards) -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <!-- Card 1: Total Patients -->
      <div class="bg-white p-4 rounded-lg border border-[#E2E8F0] shadow-sm flex items-center justify-between">
        <div>
          <span class="text-[12px] font-medium text-[#64748B] block">Total Pasien Terdaftar</span>
          <span class="text-[24px] font-bold text-[#0F172A] tabular-nums leading-tight block mt-1">
            {{ totalPatients }}
          </span>
          <span class="text-[11px] text-[#0284C7] font-medium mt-1 block">Rekam medis aktif</span>
        </div>
        <div class="w-10 h-10 rounded-lg bg-[#E0F2FE] text-[#0284C7] flex items-center justify-center font-bold">
          P
        </div>
      </div>

      <!-- Card 2: Today Visits -->
      <div class="bg-white p-4 rounded-lg border border-[#E2E8F0] shadow-sm flex items-center justify-between">
        <div>
          <span class="text-[12px] font-medium text-[#64748B] block">Kunjungan Hari Ini</span>
          <span class="text-[24px] font-bold text-[#0F172A] tabular-nums leading-tight block mt-1">
            {{ queueStore.metrics.TOTAL }}
          </span>
          <span class="text-[11px] text-[#15803D] font-medium mt-1 block">Pasien registrasi</span>
        </div>
        <div class="w-10 h-10 rounded-lg bg-[#DCFCE7] text-[#15803D] flex items-center justify-center font-bold">
          K
        </div>
      </div>

      <!-- Card 3: In Process Clinical -->
      <div class="bg-white p-4 rounded-lg border border-[#E2E8F0] shadow-sm flex items-center justify-between">
        <div>
          <span class="text-[12px] font-medium text-[#64748B] block">Dalam Pemeriksaan</span>
          <span class="text-[24px] font-bold text-[#0F172A] tabular-nums leading-tight block mt-1">
            {{ queueStore.metrics.DIPERIKSA_DOKTER + queueStore.metrics.MENUNGGU_TRIASE }}
          </span>
          <span class="text-[11px] text-[#B45309] font-medium mt-1 block">Poli klinis & triase</span>
        </div>
        <div class="w-10 h-10 rounded-lg bg-[#FEF3C7] text-[#B45309] flex items-center justify-center font-bold">
          T
        </div>
      </div>

      <!-- Card 4: Farmasi & Kasir -->
      <div class="bg-white p-4 rounded-lg border border-[#E2E8F0] shadow-sm flex items-center justify-between">
        <div>
          <span class="text-[12px] font-medium text-[#64748B] block">Farmasi & Kasir</span>
          <span class="text-[24px] font-bold text-[#0F172A] tabular-nums leading-tight block mt-1">
            {{ queueStore.metrics.FARMASI + queueStore.metrics.KASIR }}
          </span>
          <span class="text-[11px] text-[#0369A1] font-medium mt-1 block">Menunggu obat & billing</span>
        </div>
        <div class="w-10 h-10 rounded-lg bg-[#E0F2FE] text-[#0369A1] flex items-center justify-center font-bold">
          F
        </div>
      </div>
    </div>

    <!-- Active Queue Table -->
    <div class="space-y-3">
      <div class="flex items-center justify-between">
        <h2 class="text-[16px] font-semibold text-[#0F172A] leading-[24px]">
          Antrean Kunjungan Pasien Hari Ini
        </h2>
        <span class="text-[12px] text-[#64748B]">
          Menampilkan 10 antrean terkini
        </span>
      </div>

      <DataTable
        :columns="columns"
        :data="recentVisits"
        :loading="isLoading"
        empty-text="Belum ada kunjungan terdaftar hari ini"
      >
        <template #no_antrian="{ row }">
          <span class="font-bold text-[#0284C7] font-mono tabular-nums bg-[#F0F9FF] px-2 py-0.5 rounded border border-[#BAE6FD]">
            {{ row.no_antrian }}
          </span>
        </template>

        <template #no_rawat="{ row }">
          <span class="font-mono text-[#475569]">{{ row.no_rawat }}</span>
        </template>

        <template #no_rm="{ row }">
          <span class="font-mono font-medium text-[#0369A1] bg-[#E0F2FE] px-1.5 py-0.5 rounded text-[12px] tabular-nums">
            {{ row.no_rm }}
          </span>
        </template>

        <template #patient_name="{ row }">
          <div class="font-semibold text-[#0F172A]">{{ row.patient_name }}</div>
          <div v-if="row.patient_allergy" class="text-[11px] text-[#DC2626] font-medium">
            Alergi: {{ row.patient_allergy }}
          </div>
        </template>

        <template #status="{ row }">
          <StatusBadge :status="row.status" />
        </template>
      </DataTable>
    </div>
  </div>
</template>
