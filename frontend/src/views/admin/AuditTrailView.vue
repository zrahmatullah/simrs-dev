<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api/client'
import DataTable, { type Column } from '@/components/common/DataTable.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseModal from '@/components/common/BaseModal.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import type { AuditLog } from '@/types'

const logs = ref<AuditLog[]>([])
const isLoading = ref(false)
const search = ref('')
const actionFilter = ref('')
const page = ref(1)
const total = ref(0)
const limit = 20

const isDetailOpen = ref(false)
const selectedLog = ref<AuditLog | null>(null)

const columns: Column[] = [
  { key: 'created_at', label: 'Waktu / Timestamp', width: '170px' },
  { key: 'username', label: 'Pegawai / User', width: '130px' },
  { key: 'action', label: 'Aksi Sistem', width: '150px' },
  { key: 'entity_type', label: 'Entitas Medis', width: '140px' },
  { key: 'entity_id', label: 'ID Entitas', width: '160px' },
  { key: 'ip_address', label: 'IP Address', width: '130px' },
  { key: 'actions', label: 'Detail Payload', align: 'right', width: '110px' },
]

async function loadAuditLogs() {
  isLoading.value = true
  try {
    const res = await api.get<AuditLog[]>('/audit/logs', {
      search: search.value,
      action: actionFilter.value,
      page: page.value,
      limit,
    })
    if (res.success && res.data) {
      logs.value = res.data
      total.value = res.meta?.total_data || 0
    }
  } catch (err) {
    console.error('Failed to load audit logs:', err)
  } finally {
    isLoading.value = false
  }
}

function viewDetail(log: AuditLog) {
  selectedLog.value = log
  isDetailOpen.value = true
}

function formatJson(str: string): string {
  try {
    return JSON.stringify(JSON.parse(str), null, 2)
  } catch {
    return str
  }
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

function onPageChange(newPage: number) {
  page.value = newPage
  loadAuditLogs()
}

onMounted(() => {
  loadAuditLogs()
})
</script>

<template>
  <div class="space-y-4 text-left">
    <!-- Header -->
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-[20px] font-bold text-[#0F172A] leading-[28px]">
          Jejak Audit Sistem (Audit Trail Medis)
        </h1>
        <p class="text-[13px] text-[#475569]">
          Pencatatan kepatuhan hukum (*immutable medical audit trail*) atas seluruh mutasi rekam medis, e-resep, stok obat, dan pembayaran kasir.
        </p>
      </div>

      <BaseButton variant="secondary" size="sm" @click="loadAuditLogs">
        Segarkan Log
      </BaseButton>
    </div>

    <!-- Filters -->
    <div class="bg-white p-3 rounded-lg border border-[#E2E8F0] shadow-sm flex flex-wrap items-center gap-3">
      <div class="flex-1 min-w-[200px]">
        <BaseInput
          v-model="search"
          placeholder="Cari user, ID entitas, atau kata kunci payload..."
          compact
          @keyup.enter="loadAuditLogs"
        />
      </div>

      <BaseButton variant="primary" size="sm" @click="loadAuditLogs">
        Filter
      </BaseButton>
    </div>

    <!-- DataTable -->
    <DataTable
      :columns="columns"
      :data="logs"
      :loading="isLoading"
      :page="page"
      :limit="limit"
      :total="total"
      @page-change="onPageChange"
    >
      <template #created_at="{ row }">
        <span class="font-mono text-[12px] text-[#475569] tabular-nums">
          {{ formatDate(row.created_at) }}
        </span>
      </template>

      <template #username="{ row }">
        <span class="font-mono font-semibold text-[#0284C7]">{{ row.username }}</span>
      </template>

      <template #action="{ row }">
        <StatusBadge :status="row.action" />
      </template>

      <template #entity_type="{ row }">
        <span class="font-mono text-[12px] bg-[#F1F5F9] px-1.5 py-0.5 rounded text-[#334155] border border-[#CBD5E1]">
          {{ row.entity_type }}
        </span>
      </template>

      <template #entity_id="{ row }">
        <span class="font-mono text-[11px] text-[#64748B] truncate block max-w-[150px]">
          {{ row.entity_id }}
        </span>
      </template>

      <template #ip_address="{ row }">
        <span class="font-mono text-[11px] text-[#94A3B8]">{{ row.ip_address || '127.0.0.1' }}</span>
      </template>

      <template #actions="{ row }">
        <BaseButton variant="soft" size="sm" @click="viewDetail(row)">
          Detail
        </BaseButton>
      </template>
    </DataTable>

    <!-- Detail Payload Modal -->
    <BaseModal
      :is-open="isDetailOpen"
      title="Rincian Jejak Audit (Payload Audit Log)"
      max-width="lg"
      @close="isDetailOpen = false"
    >
      <div v-if="selectedLog" class="space-y-3">
        <div class="grid grid-cols-2 gap-2 text-[12px] bg-[#F8FAFC] p-3 rounded border border-[#E2E8F0]">
          <div><span class="text-[#64748B]">Waktu:</span> <span class="font-semibold">{{ formatDate(selectedLog.created_at) }}</span></div>
          <div><span class="text-[#64748B]">Pengguna:</span> <span class="font-semibold text-[#0284C7]">@{{ selectedLog.username }}</span></div>
          <div><span class="text-[#64748B]">Aksi:</span> <span class="font-semibold">{{ selectedLog.action }}</span></div>
          <div><span class="text-[#64748B]">Entitas:</span> <span class="font-semibold">{{ selectedLog.entity_type }}</span></div>
          <div class="col-span-2 truncate"><span class="text-[#64748B]">Entity ID:</span> <span class="font-mono text-[11px]">{{ selectedLog.entity_id }}</span></div>
        </div>

        <div>
          <label class="text-[12px] font-semibold text-[#334155] block mb-1">
            Data Payload JSON:
          </label>
          <pre class="bg-slate-900 text-emerald-400 p-3 rounded font-mono text-[12px] overflow-x-auto max-h-[300px]">{{ formatJson(selectedLog.details) }}</pre>
        </div>
      </div>

      <template #footer>
        <BaseButton variant="secondary" size="sm" @click="isDetailOpen = false">
          Tutup
        </BaseButton>
      </template>
    </BaseModal>
  </div>
</template>
