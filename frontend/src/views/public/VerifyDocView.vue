<template>
  <div class="min-h-screen bg-slate-50 flex flex-col justify-between font-sans">
    <header class="bg-white border-b border-slate-200 px-6 py-4 shadow-sm">
      <div class="max-w-4xl mx-auto flex items-center justify-between">
        <div class="flex items-center space-x-3">
          <div class="w-10 h-10 rounded-xl bg-sky-600 flex items-center justify-center text-white font-bold text-xl">
            +
          </div>
          <div>
            <h1 class="text-lg font-bold text-slate-800">SIMRS-ASOY Verification Portal</h1>
            <p class="text-xs text-slate-500">Verifikasi Keabsahan Tanda Tangan Elektronik & Dokumen Medis</p>
          </div>
        </div>
        <router-link to="/login" class="text-xs font-semibold text-sky-600 hover:text-sky-700">
          Masuk ke Aplikasi &rarr;
        </router-link>
      </div>
    </header>

    <main class="max-w-xl mx-auto w-full p-6 my-auto">
      <div class="bg-white border border-slate-200 rounded-3xl p-8 shadow-xl relative overflow-hidden">
        <div class="w-16 h-16 rounded-2xl bg-emerald-50 text-emerald-600 flex items-center justify-center mx-auto mb-4 border border-emerald-100 shadow-inner">
          <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
          </svg>
        </div>

        <h2 class="text-xl font-bold text-center text-slate-800 mb-1">DOKUMEN RESMI TERVERIFIKASI</h2>
        <p class="text-xs text-center text-emerald-600 font-semibold mb-6 flex items-center justify-center space-x-1">
          <span class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
          <span>Keabsahan Terverifikasi dalam Basis Data SIMRS</span>
        </p>

        <div v-if="loading" class="text-center py-6 text-sm text-slate-500">
          Memeriksa keabsahan dokumen medis...
        </div>

        <div v-else-if="error" class="bg-rose-50 border border-rose-200 text-rose-700 rounded-2xl p-4 text-xs text-center">
          {{ error }}
        </div>

        <div v-else class="space-y-4">
          <div class="bg-slate-50 rounded-2xl p-4 border border-slate-100 space-y-2 text-xs">
            <div class="flex justify-between py-1 border-b border-slate-200/60">
              <span class="text-slate-500">Instansi Penerbit:</span>
              <span class="font-bold text-slate-800">{{ verification.clinic_name }}</span>
            </div>
            <div class="flex justify-between py-1 border-b border-slate-200/60">
              <span class="text-slate-500">Tipe Dokumen:</span>
              <span class="font-bold text-slate-800">{{ verification.document_type }}</span>
            </div>
            <div class="flex justify-between py-1 border-b border-slate-200/60">
              <span class="text-slate-500">Nomor Registrasi:</span>
              <span class="font-mono font-bold text-sky-600">{{ verification.document_number }}</span>
            </div>
            <div class="flex justify-between py-1 border-b border-slate-200/60">
              <span class="text-slate-500">Nama Pasien:</span>
              <span class="font-bold text-slate-800">{{ verification.patient_name }} ({{ verification.no_rm }})</span>
            </div>
            <div v-if="verification.doctor_name" class="flex justify-between py-1 border-b border-slate-200/60">
              <span class="text-slate-500">Dokter Penanggung Jawab:</span>
              <span class="font-bold text-slate-800">{{ verification.doctor_name }}</span>
            </div>
            <div class="flex justify-between py-1">
              <span class="text-slate-500">Waktu Terbit:</span>
              <span class="font-medium text-slate-700">{{ verification.issue_date }}</span>
            </div>
          </div>

          <div class="p-3 bg-sky-50 border border-sky-100 rounded-xl text-[11px] text-sky-800">
            <strong>Keterangan Resmi:</strong> {{ verification.details }}
          </div>
        </div>
      </div>
    </main>

    <footer class="text-center py-4 text-xs text-slate-400">
      &copy; 2026 RS & KLINIK SIMRS-ASOY • Sistem Informasi Manajemen Rumah Sakit Terakreditasi
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { api as client } from '@/api/client'

const route = useRoute()
const loading = ref(true)
const error = ref('')

const verification = ref<any>({
  clinic_name: 'RUMAH SAKIT & KLINIK SIMRS-ASOY',
  document_type: 'SURAT_MEDIS',
  document_number: 'DOC-2026-001',
  patient_name: 'Pasien Terdaftar',
  no_rm: 'RM-0001',
  doctor_name: 'dr. Spesialis',
  issue_date: '23 Nov 2026',
  details: 'Dokumen tercatat sah dalam sistem rekam medis elektronik.',
})

onMounted(async () => {
  const type = (route.query.type as string) || 'INVOICE'
  const code = (route.query.code as string) || ''

  if (!code) {
    loading.value = false
    return
  }

  try {
    const res = await client.get<any>(`/public/verify-doc?type=${encodeURIComponent(type)}&code=${encodeURIComponent(code)}`)
    if (res.data) {
      verification.value = res.data.data || res.data
    }
  } catch (err: any) {
    error.value = err.response?.data?.message || 'Dokumen medis tidak ditemukan atau kode verifikasi tidak valid.'
  } finally {
    loading.value = false
  }
})
</script>
