<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { ChevronDown, Check, Search, X } from 'lucide-vue-next'

export interface SelectOption {
  value: string | number
  label: string
  subtitle?: string
  badge?: string
  disabled?: boolean
  [key: string]: any
}

interface Props {
  modelValue?: string | number | null
  label?: string
  options?: SelectOption[]
  placeholder?: string
  searchPlaceholder?: string
  searchable?: boolean
  required?: boolean
  disabled?: boolean
  loading?: boolean
  loadingText?: string
  error?: string
  hint?: string
  compact?: boolean
  clearable?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  label: '',
  options: () => [],
  placeholder: '-- Pilih opsi --',
  searchPlaceholder: 'Cari opsi...',
  searchable: undefined,
  required: false,
  disabled: false,
  loading: false,
  loadingText: 'Memuat data opsi...',
  error: '',
  hint: '',
  compact: false,
  clearable: false,
})

const emit = defineEmits<{
  (e: 'update:modelValue', val: string | number): void
  (e: 'change', val: string | number): void
}>()

const isOpen = ref(false)
const searchQuery = ref('')
const containerRef = ref<HTMLElement | null>(null)
const searchInputRef = ref<HTMLInputElement | null>(null)
const focusedIndex = ref(-1)

// Determine if search should be enabled
const isSearchEnabled = computed(() => {
  if (props.searchable !== undefined) return props.searchable
  return props.options.length >= 6
})

const selectedOption = computed(() => {
  if (props.modelValue === null || props.modelValue === undefined || props.modelValue === '') {
    return null
  }
  return props.options.find(opt => String(opt.value) === String(props.modelValue)) || null
})

const filteredOptions = computed(() => {
  if (!isSearchEnabled.value || !searchQuery.value.trim()) {
    return props.options
  }
  const q = searchQuery.value.toLowerCase().trim()
  return props.options.filter(opt => {
    const matchLabel = opt.label.toLowerCase().includes(q)
    const matchSub = opt.subtitle ? opt.subtitle.toLowerCase().includes(q) : false
    const matchBadge = opt.badge ? opt.badge.toLowerCase().includes(q) : false
    return matchLabel || matchSub || matchBadge
  })
})

const openDirection = ref<'down' | 'up'>('down')
const dropdownMaxHeight = ref(290)

function updateDropdownPlacement() {
  if (!containerRef.value) return
  const rect = containerRef.value.getBoundingClientRect()
  const spaceBelow = window.innerHeight - rect.bottom
  const spaceAbove = rect.top

  // Check closest scrollable modal body or container
  const scrollParent = containerRef.value.closest('.overflow-y-auto')
  let availableBelow = spaceBelow
  let availableAbove = spaceAbove

  if (scrollParent) {
    const parentRect = scrollParent.getBoundingClientRect()
    availableBelow = Math.min(spaceBelow, parentRect.bottom - rect.bottom)
    availableAbove = Math.min(spaceAbove, rect.top - parentRect.top)
  }

  // If space below is cramped (< 240px) and there is more room above, flip upwards!
  if (availableBelow < 240 && availableAbove > availableBelow) {
    openDirection.value = 'up'
    dropdownMaxHeight.value = Math.min(300, Math.max(160, availableAbove - 16))
  } else {
    openDirection.value = 'down'
    dropdownMaxHeight.value = Math.min(300, Math.max(160, availableBelow - 16))
  }
}

function toggleDropdown() {
  if (props.disabled) return
  isOpen.value = !isOpen.value
  if (isOpen.value) {
    searchQuery.value = ''
    focusedIndex.value = -1
    updateDropdownPlacement()
    if (isSearchEnabled.value) {
      nextTick(() => {
        searchInputRef.value?.focus()
        updateDropdownPlacement()
      })
    }
  }
}

function selectOption(option: SelectOption) {
  if (option.disabled || props.loading) return
  emit('update:modelValue', option.value)
  emit('change', option.value)
  isOpen.value = false
  searchQuery.value = ''
}

function clearSelection(e: MouseEvent) {
  e.stopPropagation()
  emit('update:modelValue', '')
  emit('change', '')
}

function handleClickOutside(e: MouseEvent) {
  if (containerRef.value && !containerRef.value.contains(e.target as Node)) {
    isOpen.value = false
  }
}

function handleKeyDown(e: KeyboardEvent) {
  if (props.disabled) return

  if (!isOpen.value) {
    if (e.key === 'Enter' || e.key === 'ArrowDown' || e.key === ' ') {
      e.preventDefault()
      toggleDropdown()
    }
    return
  }

  if (e.key === 'Escape') {
    isOpen.value = false
  } else if (e.key === 'ArrowDown') {
    e.preventDefault()
    if (filteredOptions.value.length === 0) return
    focusedIndex.value = (focusedIndex.value + 1) % filteredOptions.value.length
    scrollToFocused()
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    if (filteredOptions.value.length === 0) return
    focusedIndex.value = (focusedIndex.value - 1 + filteredOptions.value.length) % filteredOptions.value.length
    scrollToFocused()
  } else if (e.key === 'Enter') {
    e.preventDefault()
    if (focusedIndex.value >= 0 && focusedIndex.value < filteredOptions.value.length) {
      selectOption(filteredOptions.value[focusedIndex.value])
    }
  }
}

function scrollToFocused() {
  nextTick(() => {
    const list = containerRef.value?.querySelector('.options-scroll-container')
    const item = list?.children[focusedIndex.value] as HTMLElement
    if (item && list) {
      item.scrollIntoView({ block: 'nearest' })
    }
  })
}

function onWindowScrollOrResize() {
  if (isOpen.value) {
    updateDropdownPlacement()
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  window.addEventListener('resize', onWindowScrollOrResize)
  window.addEventListener('scroll', onWindowScrollOrResize, true)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  window.removeEventListener('resize', onWindowScrollOrResize)
  window.removeEventListener('scroll', onWindowScrollOrResize, true)
})
</script>

<template>
  <div
    ref="containerRef"
    :class="[
      'relative flex flex-col gap-1 w-full text-left',
      isOpen ? 'z-50' : 'z-auto'
    ]"
    @keydown="handleKeyDown"
  >
    <label v-if="label" class="text-[12px] font-medium text-[#475569] flex items-center gap-1 select-none">
      {{ label }}
      <span v-if="required" class="text-[#DC2626] font-bold">*</span>
    </label>

    <!-- Custom Select Trigger -->
    <div
      tabindex="0"
      :class="[
        'relative w-full px-3 text-[13px] bg-white border rounded-[10px] transition-all duration-150 cursor-pointer flex items-center justify-between gap-2 select-none outline-none overflow-hidden',
        compact ? 'h-[32px]' : 'h-[38px]',
        disabled ? 'bg-[#F8FAFC] text-[#94A3B8] cursor-not-allowed border-[#E2E8F0]' : 'hover:border-[#94A3B8]',
        isOpen ? 'border-[#0284C7] ring-2 ring-[rgba(2,132,199,0.2)] shadow-xs' : 'border-[#CBD5E1]',
        error ? 'border-[#DC2626] focus:border-[#DC2626] focus:ring-2 focus:ring-[#FEE2E2]' : ''
      ]"
      @click="toggleDropdown"
    >
      <!-- Label Display -->
      <div class="flex-1 min-w-0 flex items-center gap-2">
        <template v-if="loading && !selectedOption">
          <div class="flex items-center gap-2 w-full">
            <span class="text-[#0284C7] text-[12px] font-medium animate-pulse">
              {{ loadingText }}
            </span>
          </div>
        </template>
        <template v-else>
          <span
            v-if="selectedOption"
            class="font-medium text-[#0F172A] truncate"
          >
            {{ selectedOption.label }}
          </span>
          <span v-else class="text-[#94A3B8] truncate">
            {{ placeholder }}
          </span>

          <span
            v-if="selectedOption?.badge"
            class="px-1.5 py-0.5 rounded text-[10px] font-bold bg-[#E0F2FE] text-[#0369A1] shrink-0"
          >
            {{ selectedOption.badge }}
          </span>
        </template>
      </div>

      <!-- Actions (Clear, Modern Loading Spinner, & Chevron) -->
      <div class="flex items-center gap-1.5 shrink-0 text-[#64748B]">
        <!-- Clear Button -->
        <button
          v-if="clearable && selectedOption && !disabled && !loading"
          type="button"
          class="p-0.5 hover:text-[#0F172A] rounded-full hover:bg-[#F1F5F9] transition-colors"
          title="Reset"
          @click="clearSelection"
        >
          <X class="w-3.5 h-3.5" />
        </button>

        <!-- Modern Orbital Dual-Ring Spinner when loading -->
        <div v-if="loading" class="relative w-4 h-4 flex items-center justify-center shrink-0">
          <div class="absolute inset-0 rounded-full border-2 border-[#0284C7]/20"></div>
          <div class="w-4 h-4 rounded-full border-2 border-transparent border-t-[#0284C7] border-r-[#0284C7] animate-spin"></div>
        </div>

        <!-- Chevron Icon with 180deg smooth rotation -->
        <ChevronDown
          class="w-4 h-4 transition-transform duration-200 shrink-0"
          :class="[
            isOpen ? 'rotate-180 text-[#0284C7]' : 'text-[#64748B]',
            loading ? 'opacity-40' : ''
          ]"
        />
      </div>

      <!-- Linear Glowing Bottom Progress Bar when loading -->
      <div v-if="loading" class="absolute bottom-0 left-0 right-0 h-[2px] overflow-hidden">
        <div class="h-full w-full bg-gradient-to-r from-sky-400 via-[#0284C7] to-indigo-500 animate-gradient-flow"></div>
      </div>
    </div>

    <!-- Error message or Hint -->
    <span v-if="error" class="text-[11px] text-[#DC2626] font-normal leading-tight">
      {{ error }}
    </span>
    <span v-else-if="hint" class="text-[11px] text-[#64748B] font-normal leading-tight">
      {{ hint }}
    </span>

    <!-- Dropdown Menu Popover with Smart Auto-Flip -->
    <Transition name="dropdown">
      <div
        v-if="isOpen"
        :class="[
          'absolute left-0 right-0 z-50 bg-white border border-[#CBD5E1] rounded-[14px] overflow-hidden flex flex-col transition-all duration-150',
          openDirection === 'up'
            ? 'bottom-full mb-1.5 shadow-[0_-12px_28px_-6px_rgba(15,23,42,0.14),0_-6px_12px_-4px_rgba(15,23,42,0.06)]'
            : 'top-full mt-1.5 shadow-[0_12px_28px_-6px_rgba(15,23,42,0.14),0_6px_12px_-4px_rgba(15,23,42,0.06)]'
        ]"
        :style="{ maxHeight: `${dropdownMaxHeight}px` }"
      >
        <!-- Optional Search Bar if options >= 6 or searchable -->
        <div v-if="isSearchEnabled" class="p-2 border-b border-[#F1F5F9] bg-[#F8FAFC]">
          <div class="relative">
            <Search class="w-3.5 h-3.5 absolute left-2.5 top-1/2 -translate-y-1/2 text-[#94A3B8]" />
            <input
              ref="searchInputRef"
              v-model="searchQuery"
              type="text"
              :placeholder="searchPlaceholder"
              class="w-full pl-8 pr-7 py-1 text-[12px] bg-white border border-[#CBD5E1] rounded-lg outline-none focus:border-[#0284C7] focus:ring-2 focus:ring-[#0284C7]/20 text-[#0F172A]"
              @click.stop
            />
            <button
              v-if="searchQuery"
              type="button"
              class="absolute right-2 top-1/2 -translate-y-1/2 text-[#94A3B8] hover:text-[#0F172A] p-0.5"
              @click.stop="searchQuery = ''"
            >
              <X class="w-3 h-3" />
            </button>
          </div>
        </div>

        <!-- Modern Loading State inside Dropdown (when loading === true) -->
        <div v-if="loading" class="p-2 space-y-2">
          <div class="flex items-center justify-between px-3 py-1.5 text-[11px] font-semibold text-[#0284C7] bg-[#F0F9FF] rounded-[8px] border border-[#BAE6FD]/60">
            <span class="flex items-center gap-2">
              <span class="relative flex h-2 w-2">
                <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-sky-400 opacity-75"></span>
                <span class="relative inline-flex rounded-full h-2 w-2 bg-[#0284C7]"></span>
              </span>
              {{ loadingText }}
            </span>
            <div class="w-3.5 h-3.5 rounded-full border-2 border-transparent border-t-[#0284C7] animate-spin"></div>
          </div>

          <!-- Shimmering Skeleton Rows -->
          <div v-for="n in 3" :key="n" class="p-2.5 rounded-[8px] bg-slate-50/80 border border-slate-100 flex items-center justify-between relative overflow-hidden my-1">
            <div class="space-y-1.5 flex-1">
              <div class="h-3.5 bg-slate-200/80 rounded-md w-3/5"></div>
              <div class="h-2.5 bg-slate-100 rounded-md w-2/5"></div>
            </div>
            <div class="h-3 w-8 bg-slate-200/80 rounded-full"></div>
            <!-- Shimmer Gradient Sweep -->
            <div class="absolute inset-0 -translate-x-full bg-gradient-to-r from-transparent via-white/80 to-transparent animate-shimmer"></div>
          </div>
        </div>

        <!-- Options Container -->
        <div v-else class="options-scroll-container flex-1 overflow-y-auto p-1.5">
          <div
            v-if="filteredOptions.length === 0"
            class="py-6 px-3 text-center text-[#94A3B8] text-[12px] flex flex-col items-center justify-center gap-1"
          >
            <span class="text-[13px] font-medium text-[#64748B]">Tidak ada opsi yang sesuai</span>
            <span class="text-[11px]">Silakan periksa kata kunci pencarian Anda</span>
          </div>

          <div
            v-for="(opt, idx) in filteredOptions"
            :key="opt.value"
            :class="[
              'px-3.5 py-2.5 rounded-[8px] cursor-pointer transition-all duration-150 flex items-center justify-between gap-2 text-[13px] select-none my-0.5',
              String(opt.value) === String(modelValue)
                ? 'bg-[#F0F9FF] text-[#0284C7] font-semibold border border-transparent'
                : focusedIndex === idx
                  ? 'border border-[#0284C7] bg-[#F8FAFC] text-[#0F172A]'
                  : 'border border-transparent text-[#334155] hover:bg-[#F8FAFC] hover:text-[#0F172A]',
              opt.disabled ? 'opacity-40 cursor-not-allowed bg-slate-50 border-transparent' : ''
            ]"
            @click="selectOption(opt)"
            @mouseenter="focusedIndex = idx"
          >
            <div class="flex-1 min-w-0">
              <div class="truncate leading-normal">{{ opt.label }}</div>
              <div v-if="opt.subtitle" class="text-[11px] text-[#64748B] truncate mt-0.5 font-normal">
                {{ opt.subtitle }}
              </div>
            </div>

            <div class="flex items-center gap-1.5 shrink-0">
              <span
                v-if="opt.badge"
                class="px-1.5 py-0.5 rounded text-[10px] font-bold bg-[#F1F5F9] text-[#475569]"
              >
                {{ opt.badge }}
              </span>

              <!-- Selected Checkmark matching reference image -->
              <Check
                v-if="String(opt.value) === String(modelValue)"
                class="w-4 h-4 text-[#0284C7] stroke-[2.5] shrink-0"
              />
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
@keyframes shimmer {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}

@keyframes gradientMove {
  0% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
  100% { background-position: 0% 50%; }
}

.animate-shimmer {
  animation: shimmer 1.5s infinite cubic-bezier(0.4, 0, 0.2, 1);
}

.animate-gradient-flow {
  background-size: 200% 200%;
  animation: gradientMove 2s ease infinite;
}

.dropdown-enter-active {
  transition: all 0.16s cubic-bezier(0.16, 1, 0.3, 1);
}

.dropdown-leave-active {
  transition: all 0.1s cubic-bezier(0.7, 0, 0.84, 0);
}

.dropdown-enter-from {
  opacity: 0;
  transform: translateY(-5px) scale(0.985);
}

.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-3px) scale(0.985);
}
</style>
