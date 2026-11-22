<script setup lang="ts">
import BaseDatePicker from './BaseDatePicker.vue'

interface Props {
  modelValue?: string | number | null
  label?: string
  placeholder?: string
  type?: string
  required?: boolean
  disabled?: boolean
  error?: string
  compact?: boolean
  helper?: string
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  label: '',
  placeholder: '',
  type: 'text',
  required: false,
  disabled: false,
  error: '',
  compact: false,
  helper: '',
})

const emit = defineEmits<{
  (e: 'update:modelValue', val: string | number): void
  (e: 'change', val: string | number): void
}>()

function onInput(e: Event) {
  const target = e.target as HTMLInputElement
  emit('update:modelValue', target.value)
}

function onChange(e: Event) {
  const target = e.target as HTMLInputElement
  emit('change', target.value)
}

function onDateChange(val: string) {
  emit('update:modelValue', val)
  emit('change', val)
}
</script>

<template>
  <div class="flex flex-col gap-1 w-full text-left">
    <template v-if="type === 'date'">
      <BaseDatePicker
        :model-value="String(modelValue || '')"
        :label="label"
        :placeholder="placeholder || 'Pilih Tanggal...'"
        :disabled="disabled"
        :required="required"
        :error="error"
        @update:model-value="onDateChange"
        @change="onDateChange"
      />
    </template>

    <template v-else>
      <label v-if="label" class="text-[13px] font-medium text-[#334155] flex items-center gap-1">
        {{ label }}
        <span v-if="required" class="text-[#DC2626] font-bold">*</span>
      </label>

      <div class="relative w-full">
        <input
          :type="type"
          :value="modelValue"
          :placeholder="placeholder"
          :disabled="disabled"
          :required="required"
          :class="[
            'w-full px-3 text-[13px] text-[#0F172A] bg-white border rounded-[6px] transition-all outline-none',
            compact ? 'h-[30px]' : 'h-[36px]',
            disabled ? 'bg-[#F1F5F9] text-[#94A3B8] cursor-not-allowed border-[#CBD5E1]' : 'hover:border-[#94A3B8]',
            error ? 'border-[#DC2626] focus:border-[#DC2626] focus:ring-2 focus:ring-[#FEE2E2]' : 'border-[#CBD5E1] focus:border-[#0284C7] focus:ring-3 focus:ring-[rgba(2,132,199,0.15)]',
          ]"
          @input="onInput"
          @change="onChange"
        />
      </div>

      <span v-if="error" class="text-[12px] text-[#DC2626] font-normal leading-tight">
        {{ error }}
      </span>
      <span v-else-if="helper" class="text-[11px] text-[#94A3B8] font-normal">
        {{ helper }}
      </span>
    </template>
  </div>
</template>
