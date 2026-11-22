<script setup lang="ts">
import { usePatientContextStore } from '@/stores/patientContext'
import BaseButton from '@/components/common/BaseButton.vue'

const patientStore = usePatientContextStore()
</script>

<template>
  <div
    v-if="patientStore.hasActivePatient"
    class="sticky top-0 z-30 bg-white border-b-2 border-[#0284C7] shadow-sm px-4 py-2 flex flex-wrap items-center justify-between gap-3 text-left transition-all"
  >
    <div class="flex flex-wrap items-center gap-3">
      <!-- 1. Patient Name & Gender / Age -->
      <div class="flex items-center gap-2">
        <div class="w-8 h-8 rounded-full bg-[#E0F2FE] text-[#0284C7] flex items-center justify-center font-bold text-sm shrink-0">
          {{ patientStore.patient?.nama_pasien?.charAt(0) || 'P' }}
        </div>
        <div>
          <span class="text-[15px] font-bold text-[#0F172A] leading-tight block">
            {{ patientStore.patient?.nama_pasien }}
          </span>
          <span class="text-[12px] text-[#475569] font-medium">
            {{ patientStore.patient?.jenis_kelamin }} / {{ patientStore.formattedAge }}
          </span>
        </div>
      </div>

      <div class="h-6 w-px bg-[#CBD5E1] hidden sm:block"></div>

      <!-- 2. No. Rekam Medis (Monospace / Tabular in Blue Badge #E0F2FE text #0369A1) -->
      <div class="flex items-center gap-1.5">
        <span class="text-[11px] uppercase tracking-wider text-[#64748B] font-semibold">No. RM:</span>
        <span class="bg-[#E0F2FE] text-[#0369A1] font-mono font-bold text-[13px] px-2 py-0.5 rounded border border-[#BAE6FD] tabular-nums">
          {{ patientStore.patient?.no_rm }}
        </span>
      </div>

      <!-- 3. No. Rawat & Penjamin -->
      <div v-if="patientStore.visit" class="flex items-center gap-2 text-[12px] text-[#334155]">
        <div class="flex items-center gap-1">
          <span class="text-[#64748B]">No. Rawat:</span>
          <span class="font-medium font-mono text-[#0F172A]">{{ patientStore.visit.no_rawat }}</span>
        </div>

        <span class="text-[#94A3B8]">•</span>

        <span class="px-2 py-0.5 rounded text-[11px] font-medium bg-[#F1F5F9] text-[#334155] border border-[#E2E8F0]">
          {{ patientStore.visit.guarantor_name || 'Umum' }}
        </span>

        <span v-if="patientStore.visit.no_antrian" class="bg-[#0284C7] text-white text-[11px] font-bold px-1.5 py-0.5 rounded">
          Antrean {{ patientStore.visit.no_antrian }}
        </span>
      </div>

      <!-- 4. Alert Alergi (Mencolok: #FEE2E2 text #B91C1C) -->
      <div
        v-if="patientStore.allergyAlert"
        class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-[#FEE2E2] text-[#B91C1C] border border-[#FCA5A5] text-[12px] font-semibold animate-pulse"
      >
        <svg class="w-4 h-4 shrink-0 text-[#DC2626]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>
        </svg>
        <span>ALERGI: {{ patientStore.allergyAlert }}</span>
      </div>
    </div>

    <!-- Clear / Release Context Action -->
    <div class="flex items-center gap-2">
      <BaseButton
        variant="ghost"
        size="sm"
        title="Tutup konteks pasien aktif"
        @click="patientStore.clearActivePatient()"
      >
        <svg class="w-4 h-4 mr-1 text-[#64748B]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
        </svg>
        <span>Lepas Pasien</span>
      </BaseButton>
    </div>
  </div>
</template>
