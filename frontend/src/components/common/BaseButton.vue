<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  variant?: 'primary' | 'secondary' | 'soft' | 'danger' | 'ghost' | 'outline'
  size?: 'xs' | 'sm' | 'md' | 'lg'
  type?: 'button' | 'submit' | 'reset'
  loading?: boolean
  disabled?: boolean
  block?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'primary',
  size: 'md',
  type: 'button',
  loading: false,
  disabled: false,
  block: false,
})

const emit = defineEmits<{
  (e: 'click', event: MouseEvent): void
}>()

const sizeClasses = computed(() => {
  switch (props.size) {
    case 'xs':
      return 'h-[26px] px-2 text-[11px]'
    case 'sm':
      return 'h-[30px] px-2.5 text-[12px]'
    case 'lg':
      return 'h-[42px] px-5 text-[14px]'
    case 'md':
    default:
      return 'h-[36px] px-3.5 text-[13px]'
  }
})

const variantClasses = computed(() => {
  switch (props.variant) {
    case 'secondary':
      return 'bg-white text-[#334155] border border-[#CBD5E1] hover:bg-[#F1F5F9] hover:text-[#0F172A]'
    case 'soft':
      return 'bg-[#E0F2FE] text-[#0369A1] border border-transparent hover:bg-[#BAE6FD] hover:text-[#0C4A6E]'
    case 'danger':
      return 'bg-[#DC2626] text-white border border-[#DC2626] hover:bg-[#B91C1C]'
    case 'ghost':
      return 'bg-transparent text-[#475569] border border-transparent hover:bg-[#F1F5F9] hover:text-[#0F172A]'
    case 'outline':
      return 'bg-transparent text-[#0284C7] border border-[#0284C7] hover:bg-[#F0F9FF]'
    case 'primary':
    default:
      return 'bg-[#0284C7] text-white border border-[#0284C7] hover:bg-[#0369A1]'
  }
})

function handleClick(e: MouseEvent) {
  if (props.loading || props.disabled) return
  emit('click', e)
}
</script>

<template>
  <button
    :type="type"
    :disabled="disabled || loading"
    :class="[
      'inline-flex items-center justify-center gap-1.5 font-medium rounded-[6px] transition-all duration-150 select-none outline-none focus:ring-2 focus:ring-[#0EA5E9] focus:ring-offset-1',
      sizeClasses,
      variantClasses,
      block ? 'w-full' : '',
      (disabled || loading) ? 'opacity-55 cursor-not-allowed' : 'cursor-pointer active:scale-[0.98]'
    ]"
    @click="handleClick"
  >
    <!-- Loading Spinner -->
    <svg
      v-if="loading"
      class="animate-spin -ml-0.5 mr-1.5 h-4 w-4 text-current"
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 24 24"
    >
      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
    </svg>

    <slot name="icon"></slot>
    <slot></slot>
  </button>
</template>
