<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  status?: string
  variant?: 'success' | 'warning' | 'danger' | 'info' | 'neutral'
  label?: string
}

const props = defineProps<Props>()

const badgeVariant = computed(() => {
  if (props.variant) return props.variant

  const s = (props.status || '').toUpperCase()
  if (['SELESAI', 'LUNAS', 'NORMAL', 'HADIR', 'RECEIVED', 'FINAL', 'TELAAH_OK'].includes(s)) {
    return 'success'
  }
  if (['MENUNGGU_TRIASE', 'MENUNGGU_TELAAH', 'MENUNGGU', 'DRAFT', 'DALAM_PROSES', 'SUBMITTED', 'PENYIAPAN', 'TERJADWAL', 'RESCHEDULE'].includes(s)) {
    return 'warning'
  }
  if (['BATAL', 'KRITIS', 'EXPIRED', 'ALERGI', 'CLOSED', 'MERAH'].includes(s)) {
    return 'danger'
  }
  if (['DIPERIKSA_DOKTER', 'FARMASI', 'KASIR', 'OPEN', 'HIJAU'].includes(s)) {
    return 'info'
  }
  return 'neutral'
})

const badgeClasses = computed(() => {
  switch (badgeVariant.value) {
    case 'success':
      return 'bg-[#DCFCE7] text-[#15803D] border-[#86EFAC]'
    case 'warning':
      return 'bg-[#FEF3C7] text-[#B45309] border-[#FCD34D]'
    case 'danger':
      return 'bg-[#FEE2E2] text-[#B91C1C] border-[#FCA5A5]'
    case 'info':
      return 'bg-[#E0F2FE] text-[#0369A1] border-[#7DD3FC]'
    case 'neutral':
    default:
      return 'bg-[#F1F5F9] text-[#475569] border-[#CBD5E1]'
  }
})
</script>

<template>
  <span
    :class="[
      'inline-flex items-center gap-1 px-2 py-0.5 rounded-[4px] text-[11px] font-medium border leading-tight select-none',
      badgeClasses
    ]"
  >
    <slot>{{ label || status }}</slot>
  </span>
</template>
