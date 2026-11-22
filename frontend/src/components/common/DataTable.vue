<script setup lang="ts">
import { computed } from 'vue'
import BaseButton from './BaseButton.vue'

export interface Column {
  key: string
  label: string
  align?: 'left' | 'center' | 'right'
  width?: string
  sortable?: boolean
}

interface Props {
  columns: Column[]
  data: any[]
  loading?: boolean
  emptyText?: string
  // Pagination
  page?: number
  limit?: number
  total?: number
}

const props = withDefaults(defineProps<Props>(), {
  data: () => [],
  loading: false,
  emptyText: 'Tidak ada data ditemukan',
  page: 1,
  limit: 20,
  total: 0,
})

const emit = defineEmits<{
  (e: 'page-change', page: number): void
}>()

const totalPages = computed(() => Math.ceil(props.total / props.limit) || 1)
const fromIndex = computed(() => (props.page - 1) * props.limit + (props.data.length > 0 ? 1 : 0))
const toIndex = computed(() => Math.min(props.page * props.limit, props.total))

function prevPage() {
  if (props.page > 1) {
    emit('page-change', props.page - 1)
  }
}

function nextPage() {
  if (props.page < totalPages.value) {
    emit('page-change', props.page + 1)
  }
}
</script>

<template>
  <div class="w-full bg-white border border-[#E2E8F0] rounded-[6px] shadow-sm flex flex-col overflow-hidden">
    <!-- Table Container -->
    <div class="overflow-x-auto w-full">
      <table class="w-full text-left border-collapse">
        <!-- Table Header (Specified: #F1F5F9, 12px uppercase semibold) -->
        <thead class="bg-[#F1F5F9] border-b border-[#E2E8F0]">
          <tr>
            <th
              v-for="col in columns"
              :key="col.key"
              :style="{ width: col.width }"
              :class="[
                'py-2 px-3 text-[12px] font-semibold text-[#334155] uppercase tracking-wider',
                col.align === 'right' || col.key === 'actions' ? 'text-right' : col.align === 'center' ? 'text-center' : 'text-left'
              ]"
            >
              {{ col.label }}
            </th>
          </tr>
        </thead>

        <!-- Table Body (Specified: min-height 42px, padding 8px 12px, hover #F8FAFC) -->
        <tbody class="divide-y divide-[#E2E8F0] text-[13px] text-[#0F172A]">
          <tr v-if="loading">
            <td :colspan="columns.length" class="py-8 text-center text-[#94A3B8]">
              <div class="inline-flex items-center gap-2">
                <svg class="animate-spin h-5 w-5 text-[#0284C7]" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                <span>Memuat data...</span>
              </div>
            </td>
          </tr>

          <tr v-else-if="data.length === 0">
            <td :colspan="columns.length" class="py-8 text-center text-[#94A3B8]">
              {{ emptyText }}
            </td>
          </tr>

          <tr
            v-for="(row, idx) in data"
            v-else
            :key="row.id || idx"
            class="hover:bg-[#F8FAFC] transition-colors min-h-[42px]"
          >
            <td
              v-for="col in columns"
              :key="col.key"
              :class="[
                'py-2 px-3 align-middle',
                col.align === 'right' || col.key === 'actions' ? 'text-right' : col.align === 'center' ? 'text-center' : 'text-left'
              ]"
            >
              <slot :name="col.key" :row="row" :index="idx">
                <slot :name="'cell-' + col.key" :row="row" :index="idx">
                  {{ row[col.key] }}
                </slot>
              </slot>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Table Pagination Footer -->
    <div
      v-if="total > 0"
      class="px-3 py-2 bg-[#F8FAFC] border-t border-[#E2E8F0] flex items-center justify-between text-[12px] text-[#64748B]"
    >
      <div>
        Menampilkan <span class="font-medium text-[#0F172A]">{{ fromIndex }}</span> -
        <span class="font-medium text-[#0F172A]">{{ toIndex }}</span> dari
        <span class="font-medium text-[#0F172A]">{{ total }}</span> data
      </div>

      <div class="flex items-center gap-2">
        <BaseButton
          variant="secondary"
          size="sm"
          :disabled="page <= 1"
          @click="prevPage"
        >
          Sebelumnya
        </BaseButton>

        <span class="px-2 font-medium text-[#0F172A]">
          Hal {{ page }} / {{ totalPages }}
        </span>

        <BaseButton
          variant="secondary"
          size="sm"
          :disabled="page >= totalPages"
          @click="nextPage"
        >
          Selanjutnya
        </BaseButton>
      </div>
    </div>
  </div>
</template>
