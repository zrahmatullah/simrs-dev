import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Patient, Visit } from '@/types'

export interface PatientContext {
  patient: Patient
  visit?: Visit | null
}

function getStoredPatientContext(): PatientContext | null {
  try {
    const raw = localStorage.getItem('simrs_active_patient')
    return raw ? JSON.parse(raw) : null
  } catch {
    localStorage.removeItem('simrs_active_patient')
    return null
  }
}

export const usePatientContextStore = defineStore('patientContext', () => {
  const currentContext = ref<PatientContext | null>(getStoredPatientContext())

  const hasActivePatient = computed(() => !!currentContext.value?.patient)
  const patient = computed(() => currentContext.value?.patient || null)
  const visit = computed(() => currentContext.value?.visit || null)

  // Calculate age string (e.g. "4 Thn 7 Bln")
  const formattedAge = computed(() => {
    if (!patient.value?.tanggal_lahir) return '-'
    const dob = new Date(patient.value.tanggal_lahir)
    const now = new Date()

    let years = now.getFullYear() - dob.getFullYear()
    let months = now.getMonth() - dob.getMonth()

    if (months < 0) {
      years--
      months += 12
    }

    if (years > 0) {
      return `${years} Thn ${months > 0 ? months + ' Bln' : ''}`.trim()
    }
    return `${months} Bln`
  })

  const allergyAlert = computed(() => patient.value?.riwayat_alergi || null)

  function setActivePatient(p: Patient, v?: Visit | null) {
    currentContext.value = {
      patient: p,
      visit: v || null,
    }
    localStorage.setItem('simrs_active_patient', JSON.stringify(currentContext.value))
  }

  function setPatient(p: any, v?: Visit | null) {
    setActivePatient({
      id: p.id,
      no_rm: p.no_rm,
      nik: p.nik || '',
      nama_pasien: p.nama_lengkap || p.nama_pasien || '',
      jenis_kelamin: p.jenis_kelamin || 'L',
      tempat_lahir: p.tempat_lahir || '',
      tanggal_lahir: p.tanggal_lahir || '',
      nama_orang_tua: p.nama_orang_tua || '',
      no_hp: p.no_hp || '',
      alamat: p.alamat || '',
      riwayat_alergi: p.alergi || p.riwayat_alergi || null,
      golongan_darah: p.golongan_darah || '',
      status_aktif: true,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      ...p
    }, v)
  }

  function clearActivePatient() {
    currentContext.value = null
    localStorage.removeItem('simrs_active_patient')
  }

  return {
    currentContext,
    hasActivePatient,
    patient,
    visit,
    formattedAge,
    allergyAlert,
    setActivePatient,
    setPatient,
    clearActivePatient,
  }
})
