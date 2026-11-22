import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '@/api/client'
import type { User } from '@/types'

function getStoredUser(): User | null {
  try {
    const raw = localStorage.getItem('simrs_user')
    return raw ? JSON.parse(raw) : null
  } catch {
    localStorage.removeItem('simrs_user')
    return null
  }
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('simrs_token'))
  const refreshToken = ref<string | null>(localStorage.getItem('simrs_refresh_token'))
  const user = ref<User | null>(getStoredUser())
  const isAuthenticating = ref<boolean>(false)

  const isAuthenticated = computed(() => !!token.value && !!user.value)
  const currentRole = computed(() => user.value?.role_code || '')
  const currentUserName = computed(() => user.value?.full_name || user.value?.username || '')

  async function login(username: string, password: string): Promise<boolean> {
    isAuthenticating.value = true
    try {
      const res = await api.post<{ 
        access_token?: string
        refresh_token?: string
        token?: string
        user: User 
      }>('/auth/login', { username, password })

      if (res.success && res.data) {
        const accessToken = res.data.access_token || res.data.token || ''
        const refToken = res.data.refresh_token || ''

        token.value = accessToken
        refreshToken.value = refToken
        user.value = res.data.user

        localStorage.setItem('simrs_token', accessToken)
        if (refToken) {
          localStorage.setItem('simrs_refresh_token', refToken)
        }
        localStorage.setItem('simrs_user', JSON.stringify(res.data.user))
        return true
      }
      return false
    } finally {
      isAuthenticating.value = false
    }
  }

  async function fetchMe(): Promise<void> {
    if (!token.value) return
    try {
      const res = await api.get<User>('/auth/me')
      if (res.success && res.data) {
        user.value = res.data
        localStorage.setItem('simrs_user', JSON.stringify(res.data))
      }
    } catch {
      logout()
    }
  }

  async function logout(): Promise<void> {
    try {
      if (refreshToken.value) {
        await api.post('/auth/logout', { refresh_token: refreshToken.value })
      }
    } catch {
      // Ignore network error on logout
    } finally {
      token.value = null
      refreshToken.value = null
      user.value = null
      localStorage.removeItem('simrs_token')
      localStorage.removeItem('simrs_refresh_token')
      localStorage.removeItem('simrs_user')
      localStorage.removeItem('simrs_active_patient')
      window.location.href = '/login'
    }
  }

  function hasRole(allowedRoles: string | string[]): boolean {
    if (!user.value) return false
    if (user.value.role_code === 'SUPERADMIN') return true // Superadmin has access to everything
    const roles = Array.isArray(allowedRoles) ? allowedRoles : [allowedRoles]
    return roles.includes(user.value.role_code)
  }

  return {
    token,
    refreshToken,
    user,
    isAuthenticating,
    isAuthenticated,
    currentRole,
    currentUserName,
    login,
    fetchMe,
    logout,
    hasRole,
  }
})
