import type { ApiResponse } from '@/types'

const API_BASE = '/api/v1'

class ApiClient {
  private isRefreshing = false
  private refreshSubscribers: Array<(token: string) => void> = []

  private getToken(): string | null {
    return localStorage.getItem('simrs_token')
  }

  private getRefreshToken(): string | null {
    return localStorage.getItem('simrs_refresh_token')
  }

  private setTokens(accessToken: string, refreshToken: string) {
    localStorage.setItem('simrs_token', accessToken)
    localStorage.setItem('simrs_refresh_token', refreshToken)
  }

  private clearSession() {
    localStorage.removeItem('simrs_token')
    localStorage.removeItem('simrs_refresh_token')
    localStorage.removeItem('simrs_user')
    localStorage.removeItem('simrs_active_patient')
    if (window.location.pathname !== '/login') {
      window.location.href = '/login'
    }
  }

  private onRefreshed(token: string) {
    this.refreshSubscribers.forEach(cb => cb(token))
    this.refreshSubscribers = []
  }

  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<ApiResponse<T>> {
    // Normalize endpoint to prevent double '/api/v1' prefixes
    const cleanEndpoint = endpoint.startsWith('/api/v1')
      ? endpoint.substring('/api/v1'.length)
      : endpoint
    const url = `${API_BASE}${cleanEndpoint.startsWith('/') ? cleanEndpoint : '/' + cleanEndpoint}`
    const token = this.getToken()

    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
      ...(options.headers as Record<string, string> || {}),
    }

    if (token) {
      headers['Authorization'] = `Bearer ${token}`
    }

    try {
      const res = await fetch(url, {
        ...options,
        headers,
      })

      // OWASP: Handle 401 Unauthorized with Automatic Refresh Token Rotation
      if (res.status === 401 && !endpoint.includes('/auth/login') && !endpoint.includes('/auth/refresh')) {
        const refreshToken = this.getRefreshToken()
        if (refreshToken) {
          if (!this.isRefreshing) {
            this.isRefreshing = true
            try {
              const refreshRes = await fetch(`${API_BASE}/auth/refresh`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ refresh_token: refreshToken }),
              })
              const refreshText = await refreshRes.text()
              let refreshData: any = {}
              try {
                refreshData = refreshText ? JSON.parse(refreshText) : {}
              } catch {
                refreshData = {}
              }

              if (refreshRes.ok && refreshData.success && refreshData.data) {
                const newAccess = refreshData.data.access_token || refreshData.data.token
                const newRefresh = refreshData.data.refresh_token
                this.setTokens(newAccess, newRefresh)
                this.isRefreshing = false
                this.onRefreshed(newAccess)

                // Replay original request with fresh access token
                headers['Authorization'] = `Bearer ${newAccess}`
                const retryRes = await fetch(url, { ...options, headers })
                const retryText = await retryRes.text()
                try {
                  return retryText ? JSON.parse(retryText) : ({} as ApiResponse<T>)
                } catch {
                  throw new Error(retryText || `Permintaan ulang gagal (Status: ${retryRes.status})`)
                }
              } else {
                this.isRefreshing = false
                this.clearSession()
                throw new Error('Sesi Anda telah kedaluwarsa. Silakan login kembali.')
              }
            } catch (refreshErr) {
              this.isRefreshing = false
              this.clearSession()
              throw refreshErr
            }
          } else {
            // Queue request until ongoing refresh finishes
            return new Promise((resolve, reject) => {
              this.refreshSubscribers.push(async (newToken) => {
                try {
                  headers['Authorization'] = `Bearer ${newToken}`
                  const retryRes = await fetch(url, { ...options, headers })
                  const retryText = await retryRes.text()
                  try {
                    const retryJson = retryText ? JSON.parse(retryText) : {}
                    resolve(retryJson)
                  } catch (e) {
                    reject(new Error(retryText || `Gagal memproses balasan (Status: ${retryRes.status})`))
                  }
                } catch (e) {
                  reject(e)
                }
              })
            })
          }
        } else {
          this.clearSession()
        }
      }

      // Safe JSON parsing: Handles plain text (e.g., 404, 500) gracefully without throwing JSON syntax errors
      const text = await res.text()
      let json: ApiResponse<T>
      try {
        json = text ? JSON.parse(text) : ({} as ApiResponse<T>)
      } catch {
        if (!res.ok) {
          throw new Error(text || `Permintaan ke server gagal (Status: ${res.status})`)
        }
        throw new Error(`Respon server bukan format JSON yang valid (Status: ${res.status}): ${text.slice(0, 100)}`)
      }

      if (!res.ok) {
        throw new Error(json.message || `Terjadi kesalahan pada permintaan ke server (Status: ${res.status})`)
      }

      return json
    } catch (err: any) {
      throw err
    }
  }

  get<T>(endpoint: string, params?: Record<string, any>): Promise<ApiResponse<T>> {
    let url = endpoint
    if (params) {
      const filteredParams: Record<string, string> = {}
      for (const [k, v] of Object.entries(params)) {
        if (v !== undefined && v !== null && v !== '') {
          filteredParams[k] = String(v)
        }
      }
      const query = new URLSearchParams(filteredParams).toString()
      if (query) {
        url += (url.includes('?') ? '&' : '?') + query
      }
    }
    return this.request<T>(url, { method: 'GET' })
  }

  post<T>(endpoint: string, body?: any): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: body ? JSON.stringify(body) : undefined,
    })
  }

  put<T>(endpoint: string, body?: any): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: body ? JSON.stringify(body) : undefined,
    })
  }

  patch<T>(endpoint: string, body?: any): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'PATCH',
      body: body ? JSON.stringify(body) : undefined,
    })
  }

  delete<T>(endpoint: string): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { method: 'DELETE' })
  }
}

export const api = new ApiClient()
export default api
