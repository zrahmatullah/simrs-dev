<script setup lang="ts">
import { useUiStore } from '@/stores/ui'

const uiStore = useUiStore()
</script>

<template>
  <div class="fixed top-4 right-4 z-50 flex flex-col gap-2 max-w-sm w-full pointer-events-none">
    <div
      v-for="t in uiStore.toasts"
      :key="t.id"
      :class="[
        'pointer-events-auto p-3 rounded-[6px] shadow-lg border flex items-start gap-2.5 transition-all animate-in slide-in-from-top-2 duration-200 text-[13px]',
        t.type === 'success' ? 'bg-[#DCFCE7] text-[#15803D] border-[#86EFAC]' :
        t.type === 'warning' ? 'bg-[#FEF3C7] text-[#B45309] border-[#FCD34D]' :
        t.type === 'danger' ? 'bg-[#FEE2E2] text-[#B91C1C] border-[#FCA5A5]' :
        'bg-[#E0F2FE] text-[#0369A1] border-[#7DD3FC]'
      ]"
    >
      <!-- Icon -->
      <span class="shrink-0 mt-0.5">
        <svg v-if="t.type === 'success'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
        </svg>
        <svg v-else-if="t.type === 'warning'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>
        </svg>
        <svg v-else-if="t.type === 'danger'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
        </svg>
        <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
        </svg>
      </span>

      <span class="flex-1 font-medium leading-tight">{{ t.message }}</span>

      <button
        type="button"
        class="opacity-60 hover:opacity-100 transition-opacity p-0.5 -mr-1"
        @click="uiStore.removeToast(t.id)"
      >
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
        </svg>
      </button>
    </div>
  </div>
</template>
