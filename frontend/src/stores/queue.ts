import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/api/client'
import type { Visit } from '@/types'

export const useQueueStore = defineStore('queue', () => {
  const metrics = ref<Record<string, number>>({
    MENUNGGU_TRIASE: 0,
    DIPERIKSA_DOKTER: 0,
    FARMASI: 0,
    KASIR: 0,
    SELESAI: 0,
    TOTAL: 0,
  })

  const activeVisits = ref<Visit[]>([])
  const isLoading = ref<boolean>(false)

  async function fetchQueueMetrics() {
    try {
      const res = await api.get<Record<string, number>>('/visits/metrics')
      if (res.success && res.data) {
        metrics.value = res.data
      }
    } catch (err) {
      console.error('Failed to fetch queue metrics:', err)
    }
  }

  async function fetchVisits(params: { status?: string; date?: string; poli_id?: number; search?: string } = {}) {
    isLoading.value = true
    try {
      const res = await api.get<Visit[]>('/visits', params)
      if (res.success && res.data) {
        activeVisits.value = res.data
      }
    } catch (err) {
      console.error('Failed to fetch visits queue:', err)
    } finally {
      isLoading.value = false
    }
  }

  return {
    metrics,
    activeVisits,
    isLoading,
    fetchQueueMetrics,
    fetchVisits,
  }
})
