<script setup lang="ts">
import { ref, computed } from 'vue'
import { useUiStore } from '@/stores/ui'
import BaseButton from './BaseButton.vue'

interface Props {
  modelValue?: boolean
  title?: string
  message?: string
  confirmText?: string
  cancelText?: string
  confirmVariant?: 'primary' | 'danger'
  variant?: 'primary' | 'danger'
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: undefined,
  title: undefined,
  message: undefined,
  confirmText: undefined,
  cancelText: undefined,
  confirmVariant: undefined,
  variant: undefined,
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'confirm'): void
  (e: 'cancel'): void
}>()

const uiStore = useUiStore()
const isSubmitting = ref(false)

const isControlled = computed(() => props.modelValue !== undefined)
const isOpen = computed(() => isControlled.value ? props.modelValue : uiStore.confirmModal.isOpen)
const dialogTitle = computed(() => props.title || uiStore.confirmModal.title)
const dialogMessage = computed(() => props.message || uiStore.confirmModal.message)
const dialogConfirmText = computed(() => props.confirmText || uiStore.confirmModal.confirmText || 'Konfirmasi')
const dialogCancelText = computed(() => props.cancelText || uiStore.confirmModal.cancelText || 'Batal')
const dialogVariant = computed(() => props.confirmVariant || props.variant || uiStore.confirmModal.variant || 'primary')

const handleCancel = () => {
  if (isControlled.value) {
    emit('update:modelValue', false)
    emit('cancel')
  } else {
    uiStore.closeConfirm()
  }
}

async function handleConfirm() {
  if (isControlled.value) {
    emit('confirm')
  } else {
    if (uiStore.confirmModal.onConfirm) {
      isSubmitting.value = true
      try {
        await uiStore.confirmModal.onConfirm()
        uiStore.closeConfirm()
      } finally {
        isSubmitting.value = false
      }
    } else {
      uiStore.closeConfirm()
    }
  }
}
</script>

<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <!-- Backdrop -->
    <div
      class="fixed inset-0 bg-slate-900/40 backdrop-blur-sm transition-opacity"
      @click="handleCancel"
    ></div>

    <!-- Confirm Dialog -->
    <div class="relative bg-white rounded-lg shadow-xl border border-[#E2E8F0] overflow-hidden max-w-sm w-full z-10 p-5 space-y-4 animate-in fade-in zoom-in-95 duration-150">
      <div class="flex items-start gap-3">
        <!-- Icon -->
        <div
          :class="[
            'p-2 rounded-full shrink-0',
            dialogVariant === 'danger' ? 'bg-[#FEE2E2] text-[#DC2626]' : 'bg-[#E0F2FE] text-[#0284C7]'
          ]"
        >
          <svg v-if="dialogVariant === 'danger'" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>
          </svg>
          <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
          </svg>
        </div>

        <div class="space-y-1 text-left">
          <h3 class="text-[15px] font-semibold text-[#0F172A]">
            {{ dialogTitle }}
          </h3>
          <p class="text-[13px] text-[#475569] leading-relaxed">
            {{ dialogMessage }}
          </p>
        </div>
      </div>

      <div class="flex items-center justify-end gap-2 pt-2">
        <BaseButton
          variant="secondary"
          size="sm"
          :disabled="isSubmitting"
          @click="handleCancel"
        >
          {{ dialogCancelText }}
        </BaseButton>

        <BaseButton
          :variant="dialogVariant === 'danger' ? 'danger' : 'primary'"
          size="sm"
          :loading="isSubmitting"
          @click="handleConfirm"
        >
          {{ dialogConfirmText }}
        </BaseButton>
      </div>
    </div>
  </div>
</template>
