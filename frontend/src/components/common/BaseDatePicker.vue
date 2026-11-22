<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Calendar as CalendarIcon, ChevronLeft, ChevronRight, X, RotateCcw } from 'lucide-vue-next'

interface Props {
  modelValue?: string | null // Format: YYYY-MM-DD
  label?: string
  placeholder?: string
  required?: boolean
  disabled?: boolean
  error?: string
  min?: string
  max?: string
  clearable?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  label: '',
  placeholder: 'Pilih Tanggal...',
  required: false,
  disabled: false,
  error: '',
  min: '',
  max: '',
  clearable: true,
})

const emit = defineEmits<{
  (e: 'update:modelValue', val: string): void
  (e: 'change', val: string): void
}>()

const isOpen = ref(false)
const containerRef = ref<HTMLElement | null>(null)

// Current view year & month (0-indexed month)
const today = new Date()
const viewYear = ref(today.getFullYear())
const viewMonth = ref(today.getMonth())

const monthNames = [
  'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
  'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'
]

const dayNames = ['Min', 'Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab']

// Initialize view when modelValue changes or on mount
const initViewDate = () => {
  if (props.modelValue) {
    const parts = props.modelValue.split('-')
    if (parts.length === 3) {
      viewYear.value = parseInt(parts[0], 10)
      viewMonth.value = parseInt(parts[1], 10) - 1
    }
  } else {
    viewYear.value = today.getFullYear()
    viewMonth.value = today.getMonth()
  }
}

// Display Formatted Value
const formattedDisplay = computed(() => {
  if (!props.modelValue) return ''
  const parts = props.modelValue.split('-')
  if (parts.length !== 3) return props.modelValue
  const d = new Date(parseInt(parts[0], 10), parseInt(parts[1], 10) - 1, parseInt(parts[2], 10))
  if (isNaN(d.getTime())) return props.modelValue

  return d.toLocaleDateString('id-ID', {
    weekday: 'short',
    day: 'numeric',
    month: 'short',
    year: 'numeric'
  })
})

// Calendar Days Calculation
interface CalendarDay {
  year: number
  month: number
  date: number
  isoString: string
  isCurrentMonth: boolean
  isToday: boolean
  isSelected: boolean
  isDisabled: boolean
}

const calendarDays = computed<CalendarDay[]>(() => {
  const days: CalendarDay[] = []

  const firstDayOfMonth = new Date(viewYear.value, viewMonth.value, 1).getDay()
  const daysInCurrentMonth = new Date(viewYear.value, viewMonth.value + 1, 0).getDate()
  const daysInPrevMonth = new Date(viewYear.value, viewMonth.value, 0).getDate()

  const todayIso = formatDateIso(today.getFullYear(), today.getMonth() + 1, today.getDate())

  // Previous month trailing days
  for (let i = firstDayOfMonth - 1; i >= 0; i--) {
    const d = daysInPrevMonth - i
    const m = viewMonth.value === 0 ? 11 : viewMonth.value - 1
    const y = viewMonth.value === 0 ? viewYear.value - 1 : viewYear.value
    const iso = formatDateIso(y, m + 1, d)
    days.push({
      year: y,
      month: m,
      date: d,
      isoString: iso,
      isCurrentMonth: false,
      isToday: iso === todayIso,
      isSelected: iso === props.modelValue,
      isDisabled: checkDisabled(iso),
    })
  }

  // Current month days
  for (let d = 1; d <= daysInCurrentMonth; d++) {
    const iso = formatDateIso(viewYear.value, viewMonth.value + 1, d)
    days.push({
      year: viewYear.value,
      month: viewMonth.value,
      date: d,
      isoString: iso,
      isCurrentMonth: true,
      isToday: iso === todayIso,
      isSelected: iso === props.modelValue,
      isDisabled: checkDisabled(iso),
    })
  }

  // Next month leading days (fill up to 42 cells = 6 rows)
  const remainingCells = 42 - days.length
  for (let d = 1; d <= remainingCells; d++) {
    const m = viewMonth.value === 11 ? 0 : viewMonth.value + 1
    const y = viewMonth.value === 11 ? viewYear.value + 1 : viewYear.value
    const iso = formatDateIso(y, m + 1, d)
    days.push({
      year: y,
      month: m,
      date: d,
      isoString: iso,
      isCurrentMonth: false,
      isToday: iso === todayIso,
      isSelected: iso === props.modelValue,
      isDisabled: checkDisabled(iso),
    })
  }

  return days
})

function formatDateIso(y: number, m: number, d: number): string {
  const mm = m < 10 ? `0${m}` : `${m}`
  const dd = d < 10 ? `0${d}` : `${d}`
  return `${y}-${mm}-${dd}`
}

function checkDisabled(iso: string): boolean {
  if (props.min && iso < props.min) return true
  if (props.max && iso > props.max) return true
  return false
}

function prevMonth() {
  if (viewMonth.value === 0) {
    viewMonth.value = 11
    viewYear.value--
  } else {
    viewMonth.value--
  }
}

function nextMonth() {
  if (viewMonth.value === 11) {
    viewMonth.value = 0
    viewYear.value++
  } else {
    viewMonth.value++
  }
}

function selectDay(day: CalendarDay) {
  if (day.isDisabled) return
  emit('update:modelValue', day.isoString)
  emit('change', day.isoString)
  isOpen.value = false
}

function selectPreset(type: 'today' | 'yesterday' | 'tomorrow') {
  const target = new Date()
  if (type === 'yesterday') target.setDate(target.getDate() - 1)
  if (type === 'tomorrow') target.setDate(target.getDate() + 1)

  const iso = formatDateIso(target.getFullYear(), target.getMonth() + 1, target.getDate())
  viewYear.value = target.getFullYear()
  viewMonth.value = target.getMonth()

  emit('update:modelValue', iso)
  emit('change', iso)
  isOpen.value = false
}

function clearValue(e: MouseEvent) {
  e.stopPropagation()
  emit('update:modelValue', '')
  emit('change', '')
}

function togglePicker() {
  if (props.disabled) return
  isOpen.value = !isOpen.value
  if (isOpen.value) {
    initViewDate()
  }
}

function handleClickOutside(e: MouseEvent) {
  if (containerRef.value && !containerRef.value.contains(e.target as Node)) {
    isOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  initViewDate()
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<template>
  <div ref="containerRef" class="relative flex flex-col gap-1 w-full text-left">
    <label v-if="label" class="text-[13px] font-medium text-[#334155] flex items-center gap-1">
      {{ label }}
      <span v-if="required" class="text-[#DC2626] font-bold">*</span>
    </label>

    <!-- Trigger Input -->
    <div
      tabindex="0"
      :class="[
        'w-full px-3 h-[38px] bg-white border rounded-lg transition-all cursor-pointer flex items-center justify-between gap-2 select-none outline-none',
        disabled ? 'bg-[#F1F5F9] text-[#94A3B8] cursor-not-allowed border-[#CBD5E1]' : 'hover:border-[#94A3B8]',
        isOpen ? 'border-[#0284C7] ring-3 ring-[rgba(2,132,199,0.15)] shadow-xs' : 'border-[#CBD5E1]',
        error ? 'border-[#DC2626] focus:ring-2 focus:ring-[#FEE2E2]' : ''
      ]"
      @click="togglePicker"
    >
      <div class="flex items-center gap-2.5 min-w-0">
        <CalendarIcon class="w-4 h-4 text-sky-600 shrink-0" />
        <span v-if="formattedDisplay" class="text-[13px] font-semibold text-slate-800 truncate">
          {{ formattedDisplay }}
        </span>
        <span v-else class="text-[13px] text-slate-400">
          {{ placeholder }}
        </span>
      </div>

      <!-- Clear & Indicator -->
      <div class="flex items-center gap-1 text-slate-400 shrink-0">
        <button
          v-if="clearable && modelValue && !disabled"
          type="button"
          class="p-0.5 hover:text-slate-600 rounded-full hover:bg-slate-100 transition-colors"
          title="Hapus Tanggal"
          @click="clearValue"
        >
          <X class="w-3.5 h-3.5" />
        </button>
      </div>
    </div>

    <!-- Error message -->
    <span v-if="error" class="text-[12px] text-[#DC2626] font-normal leading-tight">
      {{ error }}
    </span>

    <!-- Calendar Dropdown Popover -->
    <div
      v-if="isOpen"
      class="absolute left-0 mt-1.5 z-50 bg-white border border-slate-200 rounded-2xl shadow-xl p-3.5 w-72 animate-in fade-in zoom-in-95 duration-150 select-none"
    >
      <!-- Quick Presets -->
      <div class="flex items-center gap-1.5 mb-3 pb-2.5 border-b border-slate-100 text-[11px]">
        <button
          type="button"
          class="px-2.5 py-1 rounded-md bg-slate-100 hover:bg-sky-50 hover:text-sky-700 text-slate-600 font-medium transition-colors"
          @click="selectPreset('today')"
        >
          Hari Ini
        </button>
        <button
          type="button"
          class="px-2.5 py-1 rounded-md bg-slate-100 hover:bg-sky-50 hover:text-sky-700 text-slate-600 font-medium transition-colors"
          @click="selectPreset('yesterday')"
        >
          Kemarin
        </button>
        <button
          type="button"
          class="px-2.5 py-1 rounded-md bg-slate-100 hover:bg-sky-50 hover:text-sky-700 text-slate-600 font-medium transition-colors"
          @click="selectPreset('tomorrow')"
        >
          Besok
        </button>
      </div>

      <!-- Month Navigation Header -->
      <div class="flex items-center justify-between mb-3 px-1">
        <button
          type="button"
          class="p-1 rounded-lg text-slate-500 hover:text-slate-800 hover:bg-slate-100 transition-colors"
          @click="prevMonth"
        >
          <ChevronLeft class="w-4 h-4" />
        </button>

        <div class="font-bold text-xs text-slate-800 flex items-center gap-1.5">
          <span>{{ monthNames[viewMonth] }}</span>
          <span class="text-sky-700 font-mono">{{ viewYear }}</span>
        </div>

        <button
          type="button"
          class="p-1 rounded-lg text-slate-500 hover:text-slate-800 hover:bg-slate-100 transition-colors"
          @click="nextMonth"
        >
          <ChevronRight class="w-4 h-4" />
        </button>
      </div>

      <!-- Weekday Headers -->
      <div class="grid grid-cols-7 gap-1 text-center text-[10px] font-bold text-slate-400 uppercase tracking-wider mb-1">
        <div v-for="d in dayNames" :key="d" class="py-0.5">
          {{ d }}
        </div>
      </div>

      <!-- Day Cells Grid -->
      <div class="grid grid-cols-7 gap-1 text-center">
        <button
          v-for="day in calendarDays"
          :key="day.isoString + day.month"
          type="button"
          :disabled="day.isDisabled"
          :class="[
            'h-7 w-7 mx-auto rounded-lg text-xs font-medium flex items-center justify-center transition-all relative',
            day.isSelected ? 'bg-sky-600 text-white font-bold shadow-xs' : '',
            !day.isSelected && day.isCurrentMonth ? 'text-slate-700 hover:bg-sky-50 hover:text-sky-700' : '',
            !day.isSelected && !day.isCurrentMonth ? 'text-slate-300 hover:bg-slate-50' : '',
            day.isDisabled ? 'opacity-30 cursor-not-allowed hover:bg-transparent' : '',
            day.isToday && !day.isSelected ? 'border border-sky-500 text-sky-600 font-bold' : ''
          ]"
          @click="selectDay(day)"
        >
          {{ day.date }}
          <!-- Small dot for today if not selected -->
          <span
            v-if="day.isToday && !day.isSelected"
            class="absolute bottom-0.5 w-1 h-1 rounded-full bg-sky-500"
          ></span>
        </button>
      </div>

      <!-- Bottom Jump to Today -->
      <div class="mt-3 pt-2.5 border-t border-slate-100 flex items-center justify-between text-[11px] px-1">
        <span class="text-slate-400">Pilih tanggal dari kalender</span>
        <button
          type="button"
          class="font-semibold text-sky-600 hover:text-sky-700 flex items-center gap-1"
          @click="selectPreset('today')"
        >
          <RotateCcw class="w-3 h-3" />
          Hari Ini
        </button>
      </div>
    </div>
  </div>
</template>
