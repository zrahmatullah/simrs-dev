import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface Toast {
  id: string
  message: string
  type: 'success' | 'warning' | 'danger' | 'info'
  duration?: number
}

export interface ConfirmDialog {
  isOpen: boolean
  title: string
  message: string
  confirmText?: string
  cancelText?: string
  variant?: 'primary' | 'danger'
  onConfirm?: () => Promise<void> | void
}

export const useUiStore = defineStore('ui', () => {
  // Sidebar state (collapsible to 64px from 240px)
  const isSidebarCollapsed = ref<boolean>(false)

  // Global loading
  const isGlobalLoading = ref<boolean>(false)

  // Toasts
  const toasts = ref<Toast[]>([])

  // Confirm Modal
  const confirmModal = ref<ConfirmDialog>({
    isOpen: false,
    title: '',
    message: '',
    confirmText: 'Konfirmasi',
    cancelText: 'Batal',
    variant: 'primary',
  })

  function toggleSidebar() {
    isSidebarCollapsed.value = !isSidebarCollapsed.value
  }

  function setSidebarCollapsed(val: boolean) {
    isSidebarCollapsed.value = val
  }

  function showToast(message: string, type: 'success' | 'warning' | 'danger' | 'info' | 'error' = 'info', duration = 3500) {
    const id = Math.random().toString(36).substring(2, 9)
    const normalizedType = type === 'error' ? 'danger' : type
    toasts.value.push({ id, message, type: normalizedType, duration })

    setTimeout(() => {
      removeToast(id)
    }, duration)
  }

  function removeToast(id: string) {
    toasts.value = toasts.value.filter(t => t.id !== id)
  }

  function openConfirm(options: {
    title: string
    message: string
    confirmText?: string
    cancelText?: string
    variant?: 'primary' | 'danger'
    onConfirm: () => Promise<void> | void
  }) {
    confirmModal.value = {
      isOpen: true,
      title: options.title,
      message: options.message,
      confirmText: options.confirmText || 'Konfirmasi',
      cancelText: options.cancelText || 'Batal',
      variant: options.variant || 'primary',
      onConfirm: options.onConfirm,
    }
  }

  function closeConfirm() {
    confirmModal.value.isOpen = false
    confirmModal.value.onConfirm = undefined
  }

  return {
    isSidebarCollapsed,
    isGlobalLoading,
    toasts,
    confirmModal,
    toggleSidebar,
    setSidebarCollapsed,
    showToast,
    removeToast,
    openConfirm,
    closeConfirm,
  }
})
