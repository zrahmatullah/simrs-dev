<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  isOpen?: boolean
  modelValue?: boolean
  title: string
  maxWidth?: 'sm' | 'md' | 'lg' | 'xl' | '2xl' | '3xl' | '4xl' | '5xl' | '6xl' | '7xl' | 'full' | string
  width?: string
}

const props = withDefaults(defineProps<Props>(), {
  isOpen: undefined,
  modelValue: undefined,
  maxWidth: undefined,
  width: undefined,
})

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'update:modelValue', value: boolean): void
}>()

const visible = computed(() => {
  if (props.modelValue !== undefined) return props.modelValue
  return props.isOpen ?? false
})

const widthClass = computed(() => {
  const chosen = props.width || props.maxWidth

  // If user passed a specific tailwind class or custom width (e.g., "max-w-6xl" or "w-[900px]")
  if (chosen && (chosen.startsWith('max-w-') || chosen.startsWith('w-'))) {
    return chosen
  }

  // Generous, breathable width scale for medical ERP forms
  switch (chosen) {
    case 'sm':
      return 'max-w-xl' // 576px - compact single-column
    case 'md':
      return 'max-w-3xl' // 768px - standard medium
    case 'lg':
      return 'max-w-5xl' // 1024px - standard wide (2-3 columns)
    case 'xl':
      return 'max-w-6xl' // 1152px - broad clinical review
    case '2xl':
      return 'max-w-6xl' // 1152px
    case '3xl':
      return 'max-w-7xl' // 1280px - extensive multi-column clinical TTV & EMR
    case '4xl':
      return 'max-w-[92vw] 2xl:max-w-7xl'
    case '5xl':
    case '6xl':
    case '7xl':
    case 'full':
      return 'max-w-[96vw]'
    default:
      return 'max-w-5xl' // Default is comfortably broad & spacious (1024px)
  }
})

const handleClose = () => {
  emit('close')
  emit('update:modelValue', false)
}
</script>

<template>
  <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-5 md:p-8">
    <!-- Backdrop Blur -->
    <div
      class="fixed inset-0 bg-slate-900/50 backdrop-blur-xs transition-opacity"
      @click="handleClose"
    ></div>

    <!-- Modal Box (Spacious, Fluid, & Legible) -->
    <div
      :class="[
        'relative bg-white rounded-xl shadow-2xl border border-slate-200 overflow-hidden flex flex-col max-h-[92vh] z-10 w-full animate-in fade-in zoom-in-95 duration-150',
        widthClass
      ]"
    >
      <!-- Modal Header -->
      <div class="px-5 sm:px-6 py-3.5 border-b border-[#E2E8F0] bg-[#F8FAFC] flex items-center justify-between shrink-0">
        <h3 class="text-[15px] sm:text-[16px] font-semibold text-[#0F172A] leading-[22px]">
          {{ title }}
        </h3>
        <button
          type="button"
          class="text-[#94A3B8] hover:text-[#0F172A] p-1.5 rounded-md hover:bg-[#F1F5F9] transition-colors"
          title="Tutup Modal"
          @click="handleClose"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
          </svg>
        </button>
      </div>

      <!-- Modal Body -->
      <div class="p-5 sm:p-6 overflow-y-auto space-y-4 text-[13px] text-[#0F172A]">
        <slot></slot>
      </div>

      <!-- Modal Footer -->
      <div v-if="$slots.footer" class="px-5 sm:px-6 py-3.5 bg-[#F8FAFC] border-t border-[#E2E8F0] flex items-center justify-end gap-2.5 shrink-0">
        <slot name="footer"></slot>
      </div>
    </div>
  </div>
</template>
