<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { api } from '@/api/client'
import { useUiStore } from '@/stores/ui'
import { useAuthStore } from '@/stores/auth'
import { usePatientContextStore } from '@/stores/patientContext'
import DataTable, { type Column } from '@/components/common/DataTable.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import SearchableSelect from '@/components/common/SearchableSelect.vue'
import BaseDatePicker from '@/components/common/BaseDatePicker.vue'
import BaseModal from '@/components/common/BaseModal.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import SignaturePad from '@/components/common/SignaturePad.vue'
import {
  Volume2, Stethoscope, ArrowLeft, CheckCircle2, Clock, FileText,
  Activity, AlertCircle, Plus, Trash2, Printer, Search, RefreshCw,
  ChevronDown, ChevronUp, Pill, ClipboardList, Send, Eye, FileCheck,
  Building, Share2, Receipt, Calendar, User, ShieldAlert, ShieldCheck
} from 'lucide-vue-next'
import type {
  Visit, VitalSign, MedicalRecord, Prescription, AncillaryOrder,
  InventoryItem, MasterICD10, MasterICD9, MasterProcedure, VisitProcedure,
  FaskesRujukan, MasterSpesialis, PoliUnit
} from '@/types'

const uiStore = useUiStore()
const authStore = useAuthStore()
const patientContext = usePatientContextStore()

// View Modes: 'queue' (Table Antrean) vs 'emr' (Halaman Pemeriksaan Pasien)
const viewMode = ref<'queue' | 'emr'>('queue')

// ----------------------------------------------------
// 1. QUEUE TABLE MODE STATE & ACTIONS
// ----------------------------------------------------
const queueVisits = ref<Visit[]>([])
const isLoadingQueue = ref(false)
const searchQuery = ref('')
const selectedStatusFilter = ref('ALL')

const queueColumns: Column[] = [
  { key: 'no_antrian', label: 'No. Antrean', sortable: true, width: '120px' },
  { key: 'no_rawat', label: 'No. Rawat', sortable: true, width: '180px' },
  { key: 'patient_name', label: 'Pasien & No. RM', sortable: true },
  { key: 'poli_name', label: 'Poli & Penjamin', sortable: true },
  { key: 'keluhan_utama', label: 'Keluhan Pasien' },
  { key: 'triase', label: 'Triase & TTV', width: '120px' },
  { key: 'status', label: 'Status Pelayanan', width: '160px' },
  { key: 'actions', label: 'Aksi Medis', align: 'right', width: '270px' }
]

const filteredQueue = computed(() => {
  return queueVisits.value.filter(v => {
    const q = searchQuery.value.toLowerCase().trim()
    const matchSearch = !q ||
      (v.patient_name && v.patient_name.toLowerCase().includes(q)) ||
      (v.no_rm && v.no_rm.toLowerCase().includes(q)) ||
      (v.no_rawat && v.no_rawat.toLowerCase().includes(q)) ||
      (v.no_antrian && v.no_antrian.toLowerCase().includes(q)) ||
      (v.nomor_antrean && v.nomor_antrean.toLowerCase().includes(q))

    const matchStatus = selectedStatusFilter.value === 'ALL' || v.status === selectedStatusFilter.value
    return matchSearch && matchStatus
  })
})

const queueStats = computed(() => {
  const waiting = queueVisits.value.filter(v => v.status === 'MENUNGGU_TRIASE' || v.status === 'MENUNGGU_DOKTER').length
  const examining = queueVisits.value.filter(v => v.status === 'DIPERIKSA_DOKTER').length
  const completed = queueVisits.value.filter(v => v.status === 'FARMASI' || v.status === 'KASIR' || v.status === 'SELESAI').length
  return { waiting, examining, completed, total: queueVisits.value.length }
})

async function fetchQueue() {
  isLoadingQueue.value = true
  try {
    const res = await api.get<Visit[]>('/visits')
    if (res.data) {
      queueVisits.value = res.data
    }
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal memuat antrean poliklinik', 'error')
  } finally {
    isLoadingQueue.value = false
  }
}

// Melodic hospital chime & Web Speech synthesis in Indonesian
function playQueueChime() {
  try {
    const AudioContextClass = window.AudioContext || (window as any).webkitAudioContext
    if (!AudioContextClass) return
    const ctx = new AudioContextClass()
    const now = ctx.currentTime

    // Tone 1: C5 (523.25 Hz)
    const osc1 = ctx.createOscillator()
    const gain1 = ctx.createGain()
    osc1.type = 'sine'
    osc1.frequency.setValueAtTime(523.25, now)
    gain1.gain.setValueAtTime(0.2, now)
    gain1.gain.exponentialRampToValueAtTime(0.001, now + 0.45)
    osc1.connect(gain1)
    gain1.connect(ctx.destination)
    osc1.start(now)
    osc1.stop(now + 0.45)

    // Tone 2: G5 (783.99 Hz)
    const osc2 = ctx.createOscillator()
    const gain2 = ctx.createGain()
    osc2.type = 'sine'
    osc2.frequency.setValueAtTime(783.99, now + 0.22)
    gain2.gain.setValueAtTime(0.2, now + 0.22)
    gain2.gain.exponentialRampToValueAtTime(0.001, now + 0.7)
    osc2.connect(gain2)
    gain2.connect(ctx.destination)
    osc2.start(now + 0.22)
    osc2.stop(now + 0.7)
  } catch (e) {
    console.warn('AudioContext chime error:', e)
  }
}

function callPatient(v: Visit) {
  playQueueChime()
  if ('speechSynthesis' in window) {
    const text = `Nomor antrean ${v.nomor_antrean || 'antrean'}, ${v.patient_name || 'Pasien'}, silakan menuju Ruang Periksa Dokter Poliklinik.`
    const utterance = new SpeechSynthesisUtterance(text)
    utterance.lang = 'id-ID'
    utterance.rate = 0.95
    setTimeout(() => {
      window.speechSynthesis.speak(utterance)
    }, 400)
  }
  uiStore.showToast(`Memanggil pasien antrean ${v.nomor_antrean} (${v.patient_name})`, 'info')
}

// ----------------------------------------------------
// 2. ENTER EMR CLINICAL COCKPIT
// ----------------------------------------------------
const currentVisit = ref<Visit | null>(null)
type EMRTab = 'vital' | 'cppt' | 'diagnosa' | 'tindakan' | 'rujuk' | 'penunjang' | 'resep' | 'surat' | 'consent' | 'resume'
const activeTab = ref<EMRTab>('cppt')

// Clinical Data for Current Visit
const activeVitals = ref<VitalSign | null>(null)
const existingEMR = ref<MedicalRecord | null>(null)
const patientHistory = ref<MedicalRecord[]>([])
const visitProcedures = ref<VisitProcedure[]>([])
const visitOrders = ref<AncillaryOrder[]>([])
const activePrescription = ref<Prescription | null>(null)

// Master References
const masterProcedures = ref<MasterProcedure[]>([])
const masterMedicines = ref<InventoryItem[]>([])
const icd10List = ref<MasterICD10[]>([])
const icd9List = ref<MasterICD9[]>([])
const masterFaskes = ref<FaskesRujukan[]>([])
const masterPoli = ref<PoliUnit[]>([])
const masterSpesialis = ref<MasterSpesialis[]>([])

const faskesOptions = computed(() => {
  if (masterFaskes.value.length > 0) {
    return masterFaskes.value.map(f => ({
      value: f.name,
      label: `${f.name} (${f.tipe}) - ${f.kota}`,
      badge: f.tipe
    }))
  }
  return [
    { value: 'RSUP Dr. Sardjito Yogyakarta', label: 'RSUP Dr. Sardjito Yogyakarta (RSUP) - Yogyakarta', badge: 'RSUP' },
    { value: 'RSUP Dr. Kariadi', label: 'RSUP Dr. Kariadi (RSUP) - Semarang', badge: 'RSUP' },
    { value: 'RS Jantung Harapan Kita', label: 'RS Jantung Harapan Kita (RS Khusus) - Jakarta Barat', badge: 'RS Khusus' },
    { value: 'RS Kanker Dharmais', label: 'RS Kanker Dharmais (RS Khusus) - Jakarta Barat', badge: 'RS Khusus' },
    { value: 'RSUD Pasar Minggu', label: 'RSUD Pasar Minggu (RSUD) - Jakarta Selatan', badge: 'RSUD' }
  ]
})

const poliRujukanOptions = computed(() => {
  const defaultReferralPolis = [
    'Poli Spesialis Anak / Tumbuh Kembang',
    'Poli Neurologi Anak (Saraf Anak)',
    'Poli Kardiologi & Vaskular (Jantung)',
    'Poli Bedah Anak',
    'Poli Gastroenterohepatologi',
    'Poli Alergi & Imunologi Anak',
    'Poli Endokrinologi Anak',
    'Poli Respirologi (Paru Anak)',
    'Poli Nefrologi (Ginjal Anak)'
  ]
  const fromMasterPoli = masterPoli.value.map(p => p.name)
  const combined = Array.from(new Set([...defaultReferralPolis, ...fromMasterPoli]))
  return combined.map(p => ({
    value: p,
    label: p
  }))
})

const spesialisOptions = computed(() => {
  if (masterSpesialis.value.length > 0) {
    return masterSpesialis.value.map(s => ({
      value: s.nama_dokter,
      label: `${s.nama_dokter} - ${s.spesialisasi} (${s.rumah_sakit || 'RS Rujukan'})`,
      badge: s.spesialisasi
    }))
  }
  return [
    { value: 'dr. Bambang Sp.A(K)', label: 'dr. Bambang Sp.A(K) - Spesialis Anak (Pediatri Konsultan)', badge: 'Pediatri' },
    { value: 'dr. Ratna Sp.A(K)', label: 'dr. Ratna Sp.A(K) - Spesialis Tumbuh Kembang Pediatri Sosial', badge: 'Tumbuh Kembang' },
    { value: 'dr. Irfan Sp.JP(K)', label: 'dr. Irfan Sp.JP(K) - Spesialis Jantung & Pembuluh Darah', badge: 'Kardiologi' },
    { value: 'dr. Rahmat Sp.PD-KGEH', label: 'dr. Rahmat Sp.PD-KGEH - Spesialis Gastroenterohepatologi', badge: 'Gastro' },
    { value: 'dr. Farida Sp.N', label: 'dr. Farida Sp.N - Spesialis Neurologi (Saraf Anak)', badge: 'Neurologi' }
  ]
})

function onSpesialisChange(val: string) {
  const found = masterSpesialis.value.find(s => s.nama_dokter === val)
  if (found) {
    if (found.rumah_sakit) rujukForm.value.rs_tujuan = found.rumah_sakit
    if (found.spesialisasi) rujukForm.value.poli_tujuan = found.spesialisasi
  }
}

// Forms State
const isSaving = ref(false)

// Tab 1: Vital Sign Form (Nurse / Doctor update)
const vitalsForm = ref({
  td_sistole: 120,
  td_diastole: 80,
  hr: 80,
  rr: 20,
  temp: 36.5,
  spo2: 98,
  bb: 50,
  tb: 160,
  skala_nyeri: 0,
  kesadaran: 'Compos Mentis',
  triase_level: 'HIJAU' as 'HIJAU' | 'KUNING' | 'MERAH' | 'HITAM',
  catatan_keperawatan: ''
})

// Tab 2: CPPT Clean Doctor SOAP (NO PPA Radio Buttons)
const soapForm = ref({
  subjektif: '',
  objektif: '',
  asesmen: '',
  planning: ''
})

// Tab 3: Diagnosa Form
const diagnosaForm = ref({
  icd10_primary: '',
  icd10_primary_desc: '',
  icd10_secondary: [] as string[],
  icd9_code: '',
  icd9_desc: '',
  catatan_diagnosa: ''
})

// Tab 4: Tindakan Form
const selectedProcedureID = ref<number | ''>('')
const procedureQty = ref(1)
const procedureNotes = ref('')

// Tab 5: Surat Rujukan Eksternal Form
const rujukForm = ref({
  rs_tujuan: 'RSUP Dr. Sardjito Yogyakarta',
  poli_tujuan: 'Poli Spesialis Anak / Tumbuh Kembang',
  dokter_tujuan: 'dr. Spesialis Konsultan',
  alasan_rujuk: 'Memerlukan penanganan dan evaluasi lanjutan subspesialistik',
  ringkasan_anamnesa: '',
  terapi_sementara: ''
})

// Tab 6: Rujuk Penunjang Form
const penunjangForm = ref({
  jenis_order: 'LAB' as 'LAB' | 'RAD' | 'FISIOTERAPI' | 'EKG',
  nama_tindakan: '',
  catatan_klinis: ''
})

// Tab 7: e-Resep Form
interface TempPrescriptionItem {
  item_id: number
  item_name: string
  signa: string
  aturan_pakai: string
  jumlah: number
  harga_satuan: number
  subtotal: number
  is_racikan: boolean
  racikan_nama?: string
}
const prescriptionItems = ref<TempPrescriptionItem[]>([])
const selectedMedicineID = ref<number | ''>('')
const recipeSigna = ref('3 x 1 sendok teh')
const recipeAturan = ref('Sesudah Makan')
const recipeQty = ref(1)
const recipeCatatan = ref('')
const isRacikan = ref(false)
const racikanNama = ref('')

// Tab 8: 5 Clinic Letter Templates
const selectedLetterType = ref<'SAKIT' | 'SEHAT' | 'BEBAS_NARKOBA' | 'LAIK_TERBANG' | 'IZIN_KERJA'>('SAKIT')
const letterForm = ref({
  lama_istirahat_hari: 3,
  tgl_mulai: new Date().toISOString().substring(0, 10),
  tgl_selesai: new Date(Date.now() + 2 * 86400000).toISOString().substring(0, 10),
  keperluan: 'Melamar Pekerjaan / Persyaratan Administrasi',
  usia_kehamilan_minggu: 28,
  taksiran_partus: new Date(Date.now() + 84 * 86400000).toISOString().substring(0, 10),
  kondisi_fisik: 'Sehat Jasmani dan Rohani, Tidak Ada Tanda Infeksi Akut',
  hasil_tes_narkoba: 'NEGATIF (Amphetamine, Methamphetamine, THC, Morphine - Non Reaktif)',
  hasil_buta_warna: 'NORMAL (Buku Tes Ishihara 24 Piringan Terbaca Sempurna)'
})

// Tab 9: Resume Medis Form
const resumeForm = ref({
  alasan_dirawat: '',
  ringkasan_penyakit: '',
  pemeriksaan_fisik_penting: '',
  diagnosis_akhir: '',
  prosedur_tindakan: '',
  obat_pulang: '',
  kondisi_pulang: 'Membaik / Sembuh Klinis',
  instruksi_pulang: 'Istirahat cukup, minum obat teratur sesuai anjuran, banyak minum air putih.',
  tgl_kontrol_ulang: new Date(Date.now() + 7 * 86400000).toISOString().substring(0, 10)
})

// Tab 10: Informed Consent Digital (Fase 1)
const consentList = ref<any[]>([])
const isConsentModalOpen = ref(false)
const isPrintConsentModalOpen = ref(false)
const selectedConsent = ref<any | null>(null)
const signaturePadRef = ref<any | null>(null)
const consentForm = ref({
  consent_type: 'PERSETUJUAN_TINDAKAN',
  action_name: '',
  risks_explained: 'Telah dijelaskan secara lengkap mengenai indikasi medis, tata cara tindakan, risiko komplikasi yang mungkin timbul, serta alternatif terapi lainnya.',
  signer_name: '',
  signer_relation: 'DIRI_SENDIRI',
  patient_signature_data: ''
})

const consentTypeOptions = [
  { value: 'PERSETUJUAN_TINDAKAN', label: 'Persetujuan Tindakan Medis Invasif / Operatif' },
  { value: 'PENOLAKAN_TINDAKAN', label: 'Penolakan Tindakan Medis (Informed Refusal)' },
  { value: 'PERSETUJUAN_ANESTESI', label: 'Persetujuan Tindakan Anestesi / Sedasi' },
  { value: 'BEDAH_MINOR', label: 'Persetujuan Bedah Minor / Ekstirpasi' }
]

const relationOptions = [
  { value: 'DIRI_SENDIRI', label: 'Pasien Sendiri' },
  { value: 'ORANG_TUA', label: 'Orang Tua / Wali' },
  { value: 'SUAMI_ISTRI', label: 'Suami / Istri' },
  { value: 'ANAK', label: 'Anak Kandung' },
  { value: 'KELUARGA', label: 'Keluarga Lainnya' }
]

function openConsentModal() {
  if (currentVisit.value) {
    consentForm.value.signer_name = currentVisit.value.patient_name
  }
  consentForm.value.action_name = ''
  consentForm.value.patient_signature_data = ''
  if (signaturePadRef.value) {
    signaturePadRef.value.clear()
  }
  isConsentModalOpen.value = true
}

async function saveInformedConsent() {
  if (!consentForm.value.action_name.trim()) {
    uiStore.showToast('Nama tindakan medis wajib diisi', 'error')
    return
  }
  if (!consentForm.value.signer_name.trim()) {
    uiStore.showToast('Nama penandatangan wajib diisi', 'error')
    return
  }
  if (!consentForm.value.patient_signature_data) {
    uiStore.showToast('Tanda tangan digital pasien / wali wajib digoreskan pada canvas', 'error')
    return
  }
  if (!currentVisit.value) return

  isSaving.value = true
  try {
    await api.post('/clinical/consents', {
      visit_id: currentVisit.value.id,
      consent_type: consentForm.value.consent_type,
      action_name: consentForm.value.action_name,
      risks_explained: consentForm.value.risks_explained,
      signer_name: consentForm.value.signer_name,
      signer_relation: consentForm.value.signer_relation,
      patient_signature_data: consentForm.value.patient_signature_data,
      doctor_signature_data: ''
    })

    uiStore.showToast('Informed consent digital berhasil diterbitkan ber-QR Code', 'success')
    isConsentModalOpen.value = false

    const res = await api.get<any[]>(`/clinical/consents/visit/${currentVisit.value.id}`)
    consentList.value = res.data || []
  } catch (err: any) {
    uiStore.showToast(err.response?.data?.message || 'Gagal menyimpan informed consent', 'error')
  } finally {
    isSaving.value = false
  }
}

function previewConsent(c: any) {
  selectedConsent.value = c
  isPrintConsentModalOpen.value = true
}

function printConsentDoc() {
  window.print()
}

// Modals
const isPreviewLetterModalOpen = ref(false)
const isPreviewResumeModalOpen = ref(false)
const isBillingBreakdownOpen = ref(false)

// ----------------------------------------------------
// REAL-TIME TOTAL BILLING CALCULATION
// ----------------------------------------------------
const calculatedBilling = computed(() => {
  // 1. Mandatory Admin Poli (Specified Rp 100.000)
  const biayaAdmin = 100000
  // 2. Jasa Dokter Konsultasi
  const jasaDokter = 150000
  // 3. Tindakan di Ruangan
  const totalTindakan = visitProcedures.value.reduce((acc, p) => acc + (p.subtotal || 0), 0)
  // 4. Penunjang Lab/Rad
  const totalPenunjang = visitOrders.value.reduce((acc, o) => acc + (o.tarif || 0), 0)
  // 5. e-Resep
  const totalResep = prescriptionItems.value.reduce((acc, it) => acc + (it.subtotal || 0), 0)

  const grandTotal = biayaAdmin + jasaDokter + totalTindakan + totalPenunjang + totalResep
  return {
    biayaAdmin,
    jasaDokter,
    totalTindakan,
    totalPenunjang,
    totalResep,
    grandTotal
  }
})

// Enter Patient EMR Cockpit
async function enterEMR(v: Visit) {
  currentVisit.value = v
  patientContext.setPatient({
    id: v.patient_id,
    no_rm: v.no_rm,
    nama_lengkap: v.patient_name,
    jenis_kelamin: v.jenis_kelamin || 'L',
    tanggal_lahir: v.tanggal_lahir || '2000-01-01',
    alergi: v.alergi || 'Tidak ada riwayat alergi yang dilaporkan',
    penjamin: v.guarantor_name,
    no_rawat: v.no_rawat
  })

  // Update visit status to DIPERIKSA_DOKTER if not yet
  if (v.status === 'MENUNGGU_TRIASE' || v.status === 'MENUNGGU_DOKTER') {
    try {
      await api.patch(`/visits/${v.id}/status`, { status: 'DIPERIKSA_DOKTER' })
      v.status = 'DIPERIKSA_DOKTER'
    } catch (e) {
      console.warn('Status update warning:', e)
    }
  }

  viewMode.value = 'emr'
  await loadPatientClinicalData(v)
}

function leaveEMR() {
  viewMode.value = 'queue'
  currentVisit.value = null
  fetchQueue()
}

async function loadPatientClinicalData(v: Visit) {
  // Reset local items
  prescriptionItems.value = []

  // 1. Load Vitals
  try {
    const res = await api.get<VitalSign[]>(`/clinical/vitals/visit/${v.id}`)
    if (res.data && res.data.length > 0) {
      activeVitals.value = res.data[0]
      vitalsForm.value = {
        td_sistole: activeVitals.value.td_sistole || 120,
        td_diastole: activeVitals.value.td_diastole || 80,
        hr: activeVitals.value.hr || 80,
        rr: activeVitals.value.rr || 20,
        temp: activeVitals.value.temp || 36.5,
        spo2: activeVitals.value.spo2 || 98,
        bb: activeVitals.value.bb || 50,
        tb: activeVitals.value.tb || 160,
        skala_nyeri: activeVitals.value.skala_nyeri || 0,
        kesadaran: activeVitals.value.kesadaran || 'Compos Mentis',
        triase_level: activeVitals.value.triase_level || 'HIJAU',
        catatan_keperawatan: activeVitals.value.catatan_keperawatan || ''
      }
    }
  } catch (e) {
    console.warn('No vitals yet', e)
  }

  // 2. Load EMR / CPPT
  try {
    const res = await api.get<MedicalRecord>(`/clinical/emr/visit/${v.id}`)
    if (res.data) {
      existingEMR.value = res.data
      soapForm.value = {
        subjektif: res.data.subjektif || '',
        objektif: res.data.objektif || '',
        asesmen: res.data.asesmen || '',
        planning: res.data.planning || ''
      }
      diagnosaForm.value.icd10_primary = res.data.icd10_code || ''
      diagnosaForm.value.icd10_primary_desc = res.data.diagnosis_deskripsi || ''
      diagnosaForm.value.icd9_code = res.data.icd9_code || ''
      diagnosaForm.value.icd9_desc = res.data.prosedur_deskripsi || ''
    } else {
      soapForm.value = {
        subjektif: `Pasien datang dengan keluhan ${v.keluhan_utama || 'tidak spesifik'}.`,
        objektif: 'Keadaan umum: Baik, Compos mentis. Tanda vital dalam batas terkontrol.',
        asesmen: 'Observasi keluhan klinis.',
        planning: 'Terapi medikamentosa simtomatik dan edukasi pola hidup.'
      }
    }
  } catch (e) {
    soapForm.value = {
      subjektif: `Pasien datang dengan keluhan ${v.keluhan_utama || 'tidak spesifik'}.`,
      objektif: 'Keadaan umum: Baik, Compos mentis.',
      asesmen: 'Observasi keluhan klinis.',
      planning: 'Terapi medikamentosa simtomatik dan edukasi.'
    }
  }

  // 3. Load Past CPPT History per visit
  try {
    const res = await api.get<MedicalRecord[]>(`/clinical/emr/patient/${v.patient_id}`)
    if (res.data) {
      patientHistory.value = res.data.filter(rec => rec.visit_id !== v.id)
    }
  } catch (e) {
    console.warn('History load failed', e)
  }

  // 4. Load Procedures performed
  try {
    const res = await api.get<VisitProcedure[]>(`/clinical/procedures/visit/${v.id}`)
    if (res.data) {
      visitProcedures.value = res.data
    }
  } catch (e) {
    visitProcedures.value = []
  }

  // 5. Load Ancillary Orders
  try {
    const res = await api.get<AncillaryOrder[]>(`/support/orders/visit/${v.id}`)
    if (res.data) {
      visitOrders.value = res.data
    }
  } catch (e) {
    visitOrders.value = []
  }

  // 6. Load Prescription if exists
  try {
    const res = await api.get<Prescription>(`/pharmacy/prescriptions/visit/${v.id}`)
    if (res.data && res.data.items) {
      activePrescription.value = res.data
      prescriptionItems.value = res.data.items.map(it => ({
        item_id: it.item_id,
        item_name: it.item_name || 'Obat',
        signa: it.signa,
        aturan_pakai: it.aturan_pakai,
        jumlah: it.jumlah,
        harga_satuan: it.harga_satuan,
        subtotal: it.subtotal,
        is_racikan: it.is_racikan,
        racikan_nama: it.racikan_nama
      }))
    }
  } catch (e) {
    activePrescription.value = null
  }

  // 7. Load Informed Consents (Fase 1)
  try {
    const res = await api.get<any[]>(`/clinical/consents/visit/${v.id}`)
    consentList.value = res.data || []
  } catch (e) {
    consentList.value = []
  }

  // Pre-fill Resume Medis default values
  resumeForm.value.alasan_dirawat = v.keluhan_utama || 'Keluhan umum rawat jalan'
  resumeForm.value.ringkasan_penyakit = soapForm.value.subjektif
  resumeForm.value.diagnosis_akhir = diagnosaForm.value.icd10_primary_desc || 'Observasi Klinis Rawat Jalan'
  resumeForm.value.prosedur_tindakan = visitProcedures.value.map(p => p.procedure_name).join(', ') || 'Pemeriksaan Fisik & Konsultasi Spesialis'
  resumeForm.value.obat_pulang = prescriptionItems.value.map(i => `${i.item_name} (${i.signa})`).join(', ') || 'Terapi simtomatik'
}

// ----------------------------------------------------
// CLINICAL ACTIONS IN TABS
// ----------------------------------------------------

// Tab 1: Save Vitals
async function saveVitals() {
  if (!currentVisit.value) return
  isSaving.value = true
  try {
    await api.post('/clinical/vitals', {
      visit_id: currentVisit.value.id,
      td_sistole: Number(vitalsForm.value.td_sistole),
      td_diastole: Number(vitalsForm.value.td_diastole),
      hr: Number(vitalsForm.value.hr),
      rr: Number(vitalsForm.value.rr),
      temp: Number(vitalsForm.value.temp),
      spo2: Number(vitalsForm.value.spo2),
      bb: Number(vitalsForm.value.bb),
      tb: Number(vitalsForm.value.tb),
      skala_nyeri: Number(vitalsForm.value.skala_nyeri),
      kesadaran: vitalsForm.value.kesadaran,
      triase_level: vitalsForm.value.triase_level,
      catatan_keperawatan: vitalsForm.value.catatan_keperawatan
    })
    uiStore.showToast('Tanda-tanda vital (TTV) berhasil diperbarui', 'success')
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal menyimpan TTV', 'error')
  } finally {
    isSaving.value = false
  }
}

// Tab 2: Save CPPT SOAP
async function saveSOAP() {
  if (!currentVisit.value) return
  isSaving.value = true
  try {
    await api.post('/clinical/emr', {
      visit_id: currentVisit.value.id,
      subjektif: soapForm.value.subjektif,
      objektif: soapForm.value.objektif,
      asesmen: soapForm.value.asesmen,
      planning: soapForm.value.planning,
      icd10_code: diagnosaForm.value.icd10_primary,
      icd9_code: diagnosaForm.value.icd9_code,
      diagnosis_deskripsi: diagnosaForm.value.icd10_primary_desc,
      prosedur_deskripsi: diagnosaForm.value.icd9_desc,
      status: 'FINAL'
    })
    uiStore.showToast('Catatan SOAP CPPT dokter berhasil disimpan', 'success')
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal menyimpan SOAP EMR', 'error')
  } finally {
    isSaving.value = false
  }
}

// Tab 4: Add Room Procedure with tariff
async function addProcedure() {
  if (!currentVisit.value || !selectedProcedureID.value) {
    uiStore.showToast('Pilih tindakan yang dilakukan terlebih dahulu', 'warning')
    return
  }
  const proc = masterProcedures.value.find(p => p.id === Number(selectedProcedureID.value))
  if (!proc) return

  isSaving.value = true
  try {
    const res = await api.post<VisitProcedure>('/clinical/procedures', {
      visit_id: currentVisit.value.id,
      procedure_name: proc.name,
      tariff: proc.tariff,
      qty: Number(procedureQty.value) || 1,
      catatan: procedureNotes.value
    })
    if (res.data) {
      visitProcedures.value.push(res.data)
      selectedProcedureID.value = ''
      procedureQty.value = 1
      procedureNotes.value = ''
      uiStore.showToast(`Tindakan ${proc.name} berhasil ditambahkan ke billing`, 'success')
    }
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal menambahkan tindakan poli', 'error')
  } finally {
    isSaving.value = false
  }
}

// Tab 6: Add Ancillary Order
async function addAncillaryOrder() {
  if (!currentVisit.value || !penunjangForm.value.nama_tindakan) {
    uiStore.showToast('Isi nama pemeriksaan penunjang terlebih dahulu', 'warning')
    return
  }
  isSaving.value = true
  try {
    let tarif = 120000
    if (penunjangForm.value.jenis_order === 'RAD') tarif = 220000
    if (penunjangForm.value.jenis_order === 'FISIOTERAPI') tarif = 150000
    if (penunjangForm.value.jenis_order === 'EKG') tarif = 90000

    const res = await api.post<AncillaryOrder>('/support/orders', {
      visit_id: currentVisit.value.id,
      jenis_order: penunjangForm.value.jenis_order,
      nama_tindakan: penunjangForm.value.nama_tindakan,
      tarif: tarif,
      catatan_klinis: penunjangForm.value.catatan_klinis
    })
    if (res.data) {
      visitOrders.value.push(res.data)
      penunjangForm.value.nama_tindakan = ''
      penunjangForm.value.catatan_klinis = ''
      uiStore.showToast('Order pemeriksaan penunjang berhasil dikirimkan', 'success')
    }
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal mengirim order penunjang', 'error')
  } finally {
    isSaving.value = false
  }
}

// Tab 7: Prescription Add & Submit
function addPrescriptionItem() {
  if (!selectedMedicineID.value) {
    uiStore.showToast('Pilih obat atau sediaan formularium terlebih dahulu', 'warning')
    return
  }
  const item = masterMedicines.value.find(m => m.id === Number(selectedMedicineID.value))
  if (!item) return

  const qty = Number(recipeQty.value) || 1
  const subtotal = item.sell_price * qty

  prescriptionItems.value.push({
    item_id: item.id,
    item_name: item.item_name,
    signa: recipeSigna.value,
    aturan_pakai: recipeAturan.value,
    jumlah: qty,
    harga_satuan: item.sell_price,
    subtotal: subtotal,
    is_racikan: isRacikan.value,
    racikan_nama: isRacikan.value ? racikanNama.value : undefined
  })

  // Reset item form
  selectedMedicineID.value = ''
  recipeQty.value = 1
  recipeCatatan.value = ''
  isRacikan.value = false
  racikanNama.value = ''
  uiStore.showToast('Obat ditambahkan ke daftar resep', 'info')
}

function removePrescriptionItem(index: number) {
  prescriptionItems.value.splice(index, 1)
}

async function submitPrescription() {
  if (!currentVisit.value || prescriptionItems.value.length === 0) {
    uiStore.showToast('Daftar resep obat masih kosong', 'warning')
    return
  }
  isSaving.value = true
  try {
    const res = await api.post<Prescription>('/pharmacy/prescriptions', {
      visit_id: currentVisit.value.id,
      catatan_dokter: 'Resep dokter rawat jalan poli',
      items: prescriptionItems.value.map(it => ({
        item_id: it.item_id,
        signa: it.signa,
        aturan_pakai: it.aturan_pakai,
        jumlah: it.jumlah,
        is_racikan: it.is_racikan,
        racikan_nama: it.racikan_nama
      }))
    })
    if (res.data) {
      activePrescription.value = res.data
      uiStore.showToast('e-Resep obat berhasil diterbitkan dan masuk ke antrean Farmasi', 'success')
    }
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal menerbitkan e-resep', 'error')
  } finally {
    isSaving.value = false
  }
}

// Complete Examination Action ("Selesai Pemeriksaan Dokter")
async function finalizeExamination() {
  if (!currentVisit.value) return

  uiStore.openConfirm({
    title: 'Selesai Pemeriksaan Dokter',
    message: `Selesaikan pemeriksaan untuk pasien ${currentVisit.value.patient_name}? Status kunjungan akan dialihkan ke Farmasi (jika ada resep) atau Kasir Pembayaran.`,
    confirmText: 'Ya, Selesaikan',
    variant: 'primary',
    onConfirm: async () => {
      try {
        // Auto-save SOAP if filled
        await saveSOAP()

        // Submit prescription if not yet saved but items exist
        if (prescriptionItems.value.length > 0 && !activePrescription.value) {
          await submitPrescription()
        }

        // Determine next status: FARMASI if medicines prescribed, else KASIR
        const nextStatus = prescriptionItems.value.length > 0 ? 'FARMASI' : 'KASIR'
        await api.patch(`/visits/${currentVisit.value!.id}/status`, { status: nextStatus })

        uiStore.showToast(`Pemeriksaan dokter selesai! Pasien diarahkan ke antrean ${nextStatus}.`, 'success')
        leaveEMR()
      } catch (err: any) {
        uiStore.showToast(err.message || 'Gagal menyelesaikan pemeriksaan', 'error')
      }
    }
  })
}

// Print trigger
function triggerPrint() {
  window.print()
}

// ----------------------------------------------------
// INITIALIZATION ON MOUNT
// ----------------------------------------------------
onMounted(async () => {
  await fetchQueue()

  // Load Master Procedures
  try {
    const res = await api.get<MasterProcedure[]>('/master/procedures')
    if (res.data) masterProcedures.value = res.data
  } catch (e) {
    console.warn('Procedures fetch failed', e)
  }

  // Load Master Medicines
  const fallbackMedicines: InventoryItem[] = [
    { id: 1, category_id: 1, category_name: 'Obat Sirup', item_code: 'OBAT-001', item_name: 'Paracetamol Sirup 120mg/5ml 60ml', generic_name: 'Paracetamol', unit: 'Botol', buy_price: 8000, sell_price: 15000, minimum_stock: 20, total_stock: 150, depo_stock: 45, is_medicine: true, rack_location: 'RAK-A1', status_aktif: true, created_at: '' },
    { id: 2, category_id: 1, category_name: 'Antibiotik', item_code: 'OBAT-002', item_name: 'Amoxicillin Sirup Kering 125mg/5ml', generic_name: 'Amoxicillin Trihydrate', unit: 'Botol', buy_price: 9500, sell_price: 18000, minimum_stock: 15, total_stock: 80, depo_stock: 25, is_medicine: true, rack_location: 'RAK-A2', status_aktif: true, created_at: '' },
    { id: 3, category_id: 1, category_name: 'Antihistamin', item_code: 'OBAT-003', item_name: 'Cetirizine Sirup 5mg/5ml 60ml', generic_name: 'Cetirizine HCl', unit: 'Botol', buy_price: 18000, sell_price: 32000, minimum_stock: 10, total_stock: 60, depo_stock: 20, is_medicine: true, rack_location: 'RAK-A3', status_aktif: true, created_at: '' },
    { id: 4, category_id: 5, category_name: 'Vitamin', item_code: 'OBAT-004', item_name: 'Curcuma Plus Sirup Emulsi 200ml', generic_name: 'Curcuma + Multivitamin', unit: 'Botol', buy_price: 22000, sell_price: 38000, minimum_stock: 10, total_stock: 50, depo_stock: 15, is_medicine: true, rack_location: 'RAK-B1', status_aktif: true, created_at: '' },
    { id: 5, category_id: 3, category_name: 'Obat Luar', item_code: 'OBAT-005', item_name: 'Gentamicin Salep Kulit 0.1% 5g', generic_name: 'Gentamicin Sulfate', unit: 'Tube', buy_price: 7000, sell_price: 14000, minimum_stock: 10, total_stock: 40, depo_stock: 12, is_medicine: true, rack_location: 'RAK-B2', status_aktif: true, created_at: '' },
    { id: 6, category_id: 1, category_name: 'Analgesik', item_code: 'OBAT-006', item_name: 'Ibuprofen Tablet 400mg', generic_name: 'Ibuprofen', unit: 'Tablet', buy_price: 5000, sell_price: 8500, minimum_stock: 20, total_stock: 200, depo_stock: 50, is_medicine: true, rack_location: 'RAK-A4', status_aktif: true, created_at: '' },
    { id: 7, category_id: 1, category_name: 'Bronkodilator', item_code: 'OBAT-007', item_name: 'Salbutamol Sirup 2mg/5ml 100ml', generic_name: 'Salbutamol', unit: 'Botol', buy_price: 15000, sell_price: 26000, minimum_stock: 10, total_stock: 45, depo_stock: 18, is_medicine: true, rack_location: 'RAK-A5', status_aktif: true, created_at: '' },
    { id: 8, category_id: 5, category_name: 'Mineral', item_code: 'OBAT-008', item_name: 'Zinc Sirup 20mg/5ml 60ml', generic_name: 'Zinc Sulfate', unit: 'Botol', buy_price: 11000, sell_price: 20000, minimum_stock: 10, total_stock: 90, depo_stock: 35, is_medicine: true, rack_location: 'RAK-B3', status_aktif: true, created_at: '' }
  ]

  try {
    const res = await api.get<InventoryItem[]>('/logistics/items', { limit: 100 })
    if (Array.isArray(res.data) && res.data.length > 0) {
      masterMedicines.value = res.data
    } else if ((res.data as any)?.data && Array.isArray((res.data as any).data) && (res.data as any).data.length > 0) {
      masterMedicines.value = (res.data as any).data
    } else {
      masterMedicines.value = fallbackMedicines
    }
  } catch (e) {
    console.warn('Medicines fetch failed, using fallback catalogue', e)
    masterMedicines.value = fallbackMedicines
  }

  // Load ICD10
  try {
    const res = await api.get<MasterICD10[]>('/master/icd10')
    if (res.data) icd10List.value = res.data
  } catch (e) {
    console.warn('ICD10 fetch failed', e)
  }

  // Load ICD9
  try {
    const res = await api.get<MasterICD9[]>('/master/icd9')
    if (res.data) icd9List.value = res.data
  } catch (e) {
    console.warn('ICD9 fetch failed', e)
  }

  // Load Master Faskes
  try {
    const res = await api.get<FaskesRujukan[]>('/master/faskes')
    if (res.data && res.data.length > 0) masterFaskes.value = res.data
  } catch (e) {
    console.warn('Faskes fetch failed', e)
  }

  // Load Master Poli
  try {
    const res = await api.get<PoliUnit[]>('/master/poli')
    if (res.data && res.data.length > 0) masterPoli.value = res.data
  } catch (e) {
    console.warn('Poli fetch failed', e)
  }

  // Load Master Spesialis
  try {
    const res = await api.get<MasterSpesialis[]>('/master/spesialis')
    if (res.data && res.data.length > 0) masterSpesialis.value = res.data
  } catch (e) {
    console.warn('Spesialis fetch failed', e)
  }
})
</script>

<template>
  <div class="w-full max-w-full space-y-4">
    <!-- ========================================================================= -->
    <!-- MODE 1: DAFTAR TABEL ANTREAN PASIEN DOKTER                                -->
    <!-- ========================================================================= -->
    <div v-if="viewMode === 'queue'" class="space-y-4">
      <!-- Header Banner & Operational Metrics -->
      <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 flex flex-col md:flex-row md:items-center justify-between gap-4 shadow-xs">
        <div>
          <div class="flex items-center gap-2">
            <span class="w-2.5 h-2.5 rounded-full bg-[#0284C7] animate-pulse"></span>
            <h1 class="text-[18px] font-bold text-[#0F172A] tracking-tight">Antrean Pasien Poliklinik Dokter</h1>
          </div>
          <p class="text-[12px] text-[#64748B] mt-0.5">
            Panggil pasien dengan pengeras suara terintegrasi dan pilih "Periksa / Masuk EMR" untuk membuka rekam medis.
          </p>
        </div>

        <!-- Metric Badges -->
        <div class="flex items-center gap-2 flex-wrap">
          <div class="px-3 py-1.5 rounded-[6px] bg-[#E0F2FE] border border-[#BAE6FD] text-[#0369A1] flex items-center gap-2">
            <Clock class="w-4 h-4" />
            <span class="text-[12px] font-medium">Menunggu:</span>
            <span class="text-[14px] font-bold font-mono">{{ queueStats.waiting }}</span>
          </div>

          <div class="px-3 py-1.5 rounded-[6px] bg-[#FEF3C7] border border-[#FDE68A] text-[#B45309] flex items-center gap-2">
            <Stethoscope class="w-4 h-4" />
            <span class="text-[12px] font-medium">Sedang Diperiksa:</span>
            <span class="text-[14px] font-bold font-mono">{{ queueStats.examining }}</span>
          </div>

          <div class="px-3 py-1.5 rounded-[6px] bg-[#DCFCE7] border border-[#BBF7D0] text-[#15803D] flex items-center gap-2">
            <CheckCircle2 class="w-4 h-4" />
            <span class="text-[12px] font-medium">Selesai:</span>
            <span class="text-[14px] font-bold font-mono">{{ queueStats.completed }}</span>
          </div>

          <BaseButton variant="outline" size="sm" :loading="isLoadingQueue" @click="fetchQueue">
            <RefreshCw class="w-3.5 h-3.5 mr-1" />
            Muat Ulang
          </BaseButton>
        </div>
      </div>

      <!-- Filters & Search Toolbar -->
      <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-3 flex flex-col md:flex-row items-center justify-between gap-3 shadow-xs">
        <div class="flex-1 w-full md:w-auto relative">
          <BaseInput
            v-model="searchQuery"
            placeholder="Cari berdasarkan No. RM, Nama Pasien, No. Rawat, atau Antrean..."
            class="w-full"
          >
            <template #prefix>
              <Search class="w-4 h-4 text-[#94A3B8]" />
            </template>
          </BaseInput>
        </div>

        <div class="w-full md:w-64">
          <BaseSelect
            v-model="selectedStatusFilter"
            :options="[
              { value: 'ALL', label: 'Semua Status Antrean' },
              { value: 'MENUNGGU_TRIASE', label: 'Menunggu Triase & TTV' },
              { value: 'MENUNGGU_DOKTER', label: 'Siap Diperiksa Dokter' },
              { value: 'DIPERIKSA_DOKTER', label: 'Sedang Diperiksa' },
              { value: 'FARMASI', label: 'Menuju Farmasi' },
              { value: 'KASIR', label: 'Menuju Kasir' },
              { value: 'SELESAI', label: 'Pelayanan Selesai' }
            ]"
            placeholder="Filter Status..."
            compact
          />
        </div>
      </div>

      <!-- Queue Table -->
      <div class="bg-white border border-[#E2E8F0] rounded-[8px] shadow-xs overflow-hidden">
        <DataTable
          :columns="queueColumns"
          :data="filteredQueue"
          :loading="isLoadingQueue"
          hoverable
        >
          <!-- No. Antrean -->
          <template #no_antrian="{ row }">
            <div class="flex items-center gap-1.5">
              <span class="px-2.5 py-1 rounded bg-[#0284C7] text-white font-bold text-[13px] font-mono tracking-wider shadow-2xs">
                {{ row.no_antrian || row.nomor_antrean || '-' }}
              </span>
            </div>
          </template>

          <!-- No. Rawat -->
          <template #no_rawat="{ row }">
            <div class="text-left font-mono text-[12px] text-[#334155]">
              <span class="font-semibold text-[#0F172A] block">{{ row.no_rawat }}</span>
              <span v-if="row.waktu_masuk || row.waktu_kunjungan" class="text-[11px] text-[#64748B] block">
                {{ new Date(row.waktu_masuk || row.waktu_kunjungan).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' }) }} WIB
              </span>
            </div>
          </template>

          <!-- Pasien & No. RM -->
          <template #patient_name="{ row }">
            <div class="text-left">
              <div class="flex items-center gap-1.5">
                <span class="font-bold text-[13px] text-[#0F172A]">{{ row.patient_name }}</span>
                <span class="px-1.5 py-0.2 rounded text-[10px] font-bold" :class="row.jenis_kelamin === 'L' ? 'bg-blue-100 text-blue-700' : 'bg-pink-100 text-pink-700'">
                  {{ row.jenis_kelamin === 'L' ? 'L' : 'P' }}
                </span>
              </div>
              <span class="text-[11px] font-mono text-[#0284C7] font-semibold block">RM: {{ row.no_rm }}</span>
            </div>
          </template>

          <!-- Poli & Penjamin -->
          <template #poli_name="{ row }">
            <div class="text-left">
              <span class="text-[12px] font-semibold text-[#0F172A] block">{{ row.poli_name }}</span>
              <span class="text-[11px] text-[#64748B] block">{{ row.guarantor_name || 'Umum' }}</span>
            </div>
          </template>

          <!-- Keluhan Pasien -->
          <template #keluhan_utama="{ row }">
            <div class="text-left max-w-[200px] truncate text-[12px] text-[#475569]" :title="row.keluhan_utama">
              {{ row.keluhan_utama || 'Tidak ada catatan keluhan' }}
            </div>
          </template>

          <!-- Triase & TTV -->
          <template #triase="{ row }">
            <div class="flex items-center gap-1">
              <span
                class="px-2 py-0.5 rounded text-[11px] font-bold tracking-tight"
                :class="row.triase_level === 'MERAH' ? 'bg-red-100 text-red-700 border border-red-300' :
                        row.triase_level === 'KUNING' ? 'bg-amber-100 text-amber-800 border border-amber-300' :
                        'bg-emerald-100 text-emerald-700 border border-emerald-300'"
              >
                {{ row.triase_level || 'HIJAU' }}
              </span>
            </div>
          </template>

          <!-- Status Pelayanan -->
          <template #status="{ row }">
            <StatusBadge :status="row.status" />
          </template>

          <!-- Actions -->
          <template #actions="{ row }">
            <div class="flex items-center justify-end gap-1.5">
              <!-- Button Panggil Pasien -->
              <BaseButton
                variant="soft"
                size="sm"
                title="Panggil pasien via speaker antrean"
                @click="callPatient(row)"
              >
                <template #icon>
                  <Volume2 class="w-3.5 h-3.5" />
                </template>
                Panggil
              </BaseButton>

              <!-- Button Masuk EMR -->
              <BaseButton
                variant="primary"
                size="sm"
                title="Buka rekam medis pasien (EMR)"
                @click="enterEMR(row)"
              >
                <template #icon>
                  <Stethoscope class="w-3.5 h-3.5" />
                </template>
                EMR
              </BaseButton>
            </div>
          </template>
        </DataTable>
      </div>
    </div>

    <!-- ========================================================================= -->
    <!-- MODE 2: EMR CLINICAL COCKPIT (9 TABS DETAIL)                               -->
    <!-- ========================================================================= -->
    <div v-else-if="viewMode === 'emr' && currentVisit" class="space-y-4">
      <!-- Top Action Bar & Patient Metadata Bar -->
      <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-3.5 shadow-xs space-y-3">
        <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-3">
          <!-- Back Button & Patient Identity -->
          <div class="flex items-center gap-3">
            <BaseButton
              variant="secondary"
              size="sm"
              @click="leaveEMR"
            >
              <template #icon>
                <ArrowLeft class="w-4 h-4" />
              </template>
              Kembali ke Antrean
            </BaseButton>

            <div class="h-6 w-px bg-[#E2E8F0] hidden sm:block"></div>

            <div>
              <div class="flex items-center gap-2 flex-wrap">
                <span class="text-[16px] font-bold text-[#0F172A]">{{ currentVisit.patient_name }}</span>
                <span class="px-1.5 py-0.5 rounded text-[10px] font-bold bg-[#E0F2FE] text-[#0369A1] font-mono">
                  RM: {{ currentVisit.no_rm }}
                </span>
                <span class="px-1.5 py-0.5 rounded text-[10px] font-bold bg-[#F1F5F9] text-[#475569] font-mono">
                  RAWAT: {{ currentVisit.no_rawat }}
                </span>
                <span class="px-1.5 py-0.5 rounded text-[10px] font-bold bg-amber-50 text-amber-700 border border-amber-200">
                  {{ currentVisit.guarantor_name }}
                </span>
                <span v-if="currentVisit.alergi" class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-red-100 text-red-700 border border-red-200 flex items-center gap-1">
                  <ShieldAlert class="w-3 h-3" />
                  Alergi: {{ currentVisit.alergi }}
                </span>
              </div>
              <p class="text-[11px] text-[#64748B] mt-0.5">
                Keluhan: <span class="text-[#334155] font-medium">{{ currentVisit.keluhan_utama || 'Pemeriksaan rutin' }}</span>
              </p>
            </div>
          </div>

          <!-- Real-Time Total Billing & Examination Complete Button -->
          <div class="flex items-center gap-2.5 flex-wrap">
            <!-- Total Billing Card -->
            <div
              class="px-3 py-1.5 rounded-[6px] bg-[#F8FAFC] border border-[#CBD5E1] text-right cursor-pointer hover:border-[#0284C7] transition-all"
              title="Klik untuk melihat rincian billing"
              @click="isBillingBreakdownOpen = true"
            >
              <span class="text-[10px] text-[#64748B] uppercase font-bold tracking-wider block">Total Tagihan Sementara</span>
              <span class="text-[15px] font-extrabold text-[#0284C7] font-mono tabular-nums block">
                Rp {{ calculatedBilling.grandTotal.toLocaleString('id-ID') }}
              </span>
            </div>

            <!-- Button Selesai Pemeriksaan Dokter -->
            <BaseButton
              variant="primary"
              size="md"
              class="!bg-[#059669] hover:!bg-[#047857] shadow-sm flex items-center gap-1.5"
              @click="finalizeExamination"
            >
              <CheckCircle2 class="w-4 h-4" />
              Selesai Pemeriksaan Dokter
            </BaseButton>
          </div>
        </div>

        <!-- 9 Operational EMR Tabs Navigation -->
        <div class="flex items-center gap-1 overflow-x-auto border-t border-[#F1F5F9] pt-2 scrollbar-none">
          <button
            v-for="t in [
              { id: 'vital', label: 'VITAL SIGN', icon: Activity },
              { id: 'cppt', label: 'CPPT', icon: FileText },
              { id: 'diagnosa', label: 'DIAGNOSA', icon: AlertCircle },
              { id: 'tindakan', label: 'TINDAKAN', icon: Stethoscope },
              { id: 'rujuk', label: 'RUJUK', icon: Share2 },
              { id: 'penunjang', label: 'RUJUK PENUNJANG', icon: Building },
              { id: 'resep', label: 'RESEP', icon: Pill },
              { id: 'surat', label: 'SURAT', icon: FileCheck },
              { id: 'consent', label: 'INFORMED CONSENT', icon: ShieldCheck },
              { id: 'resume', label: 'RESUME MEDIS', icon: ClipboardList }
            ]"
            :key="t.id"
            type="button"
            :class="[
              'px-3 py-1.5 rounded-[6px] text-[12px] font-bold transition-all flex items-center gap-1.5 whitespace-nowrap shrink-0',
              activeTab === t.id
                ? 'bg-[#0284C7] text-white shadow-xs'
                : 'text-[#475569] hover:bg-[#F1F5F9] hover:text-[#0F172A]'
            ]"
            @click="activeTab = t.id as EMRTab"
          >
            <component :is="t.icon" class="w-3.5 h-3.5" />
            {{ t.label }}
          </button>
        </div>
      </div>

      <!-- ========================================================================= -->
      <!-- TAB 1: VITAL SIGN (TTV)                                                   -->
      <!-- ========================================================================= -->
      <div v-if="activeTab === 'vital'" class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs space-y-4">
        <div class="flex items-center justify-between border-b border-[#F1F5F9] pb-3">
          <div>
            <h2 class="text-[15px] font-bold text-[#0F172A]">Tanda-Tanda Vital & Triase (TTV)</h2>
            <p class="text-[12px] text-[#64748B]">Parameter klinis vital hasil pengukuran awal perawat / dokter.</p>
          </div>
          <BaseButton variant="primary" size="sm" :loading="isSaving" @click="saveVitals">
            Simpan Perubahan TTV
          </BaseButton>
        </div>

        <!-- 8 Vital Sign Metric Cards -->
        <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
          <div class="p-3 rounded-[6px] bg-[#F8FAFC] border border-[#E2E8F0]">
            <span class="text-[11px] text-[#64748B] font-medium block">Tekanan Darah (TD)</span>
            <div class="flex items-center gap-1 mt-1">
              <input
                v-model.number="vitalsForm.td_sistole"
                type="number"
                class="w-16 h-8 text-center font-bold text-[14px] border border-[#CBD5E1] rounded bg-white"
              />
              <span class="text-[#64748B] font-bold">/</span>
              <input
                v-model.number="vitalsForm.td_diastole"
                type="number"
                class="w-16 h-8 text-center font-bold text-[14px] border border-[#CBD5E1] rounded bg-white"
              />
              <span class="text-[11px] text-[#64748B] ml-1">mmHg</span>
            </div>
          </div>

          <div class="p-3 rounded-[6px] bg-[#F8FAFC] border border-[#E2E8F0]">
            <span class="text-[11px] text-[#64748B] font-medium block">Denyut Nadi (HR)</span>
            <div class="flex items-center gap-1 mt-1">
              <input
                v-model.number="vitalsForm.hr"
                type="number"
                class="w-20 h-8 text-center font-bold text-[14px] border border-[#CBD5E1] rounded bg-white"
              />
              <span class="text-[11px] text-[#64748B] ml-1">bpm</span>
            </div>
          </div>

          <div class="p-3 rounded-[6px] bg-[#F8FAFC] border border-[#E2E8F0]">
            <span class="text-[11px] text-[#64748B] font-medium block">Laju Napas (RR)</span>
            <div class="flex items-center gap-1 mt-1">
              <input
                v-model.number="vitalsForm.rr"
                type="number"
                class="w-20 h-8 text-center font-bold text-[14px] border border-[#CBD5E1] rounded bg-white"
              />
              <span class="text-[11px] text-[#64748B] ml-1">x/mnt</span>
            </div>
          </div>

          <div class="p-3 rounded-[6px] bg-[#F8FAFC] border border-[#E2E8F0]">
            <span class="text-[11px] text-[#64748B] font-medium block">Suhu Tubuh</span>
            <div class="flex items-center gap-1 mt-1">
              <input
                v-model.number="vitalsForm.temp"
                type="number"
                step="0.1"
                class="w-20 h-8 text-center font-bold text-[14px] border border-[#CBD5E1] rounded bg-white"
              />
              <span class="text-[11px] text-[#64748B] ml-1">°C</span>
            </div>
          </div>

          <div class="p-3 rounded-[6px] bg-[#F8FAFC] border border-[#E2E8F0]">
            <span class="text-[11px] text-[#64748B] font-medium block">Saturasi O2 (SpO2)</span>
            <div class="flex items-center gap-1 mt-1">
              <input
                v-model.number="vitalsForm.spo2"
                type="number"
                class="w-20 h-8 text-center font-bold text-[14px] border border-[#CBD5E1] rounded bg-white"
              />
              <span class="text-[11px] text-[#64748B] ml-1">%</span>
            </div>
          </div>

          <div class="p-3 rounded-[6px] bg-[#F8FAFC] border border-[#E2E8F0]">
            <span class="text-[11px] text-[#64748B] font-medium block">Berat Badan (BB)</span>
            <div class="flex items-center gap-1 mt-1">
              <input
                v-model.number="vitalsForm.bb"
                type="number"
                step="0.5"
                class="w-20 h-8 text-center font-bold text-[14px] border border-[#CBD5E1] rounded bg-white"
              />
              <span class="text-[11px] text-[#64748B] ml-1">kg</span>
            </div>
          </div>

          <div class="p-3 rounded-[6px] bg-[#F8FAFC] border border-[#E2E8F0]">
            <span class="text-[11px] text-[#64748B] font-medium block">Tinggi Badan (TB)</span>
            <div class="flex items-center gap-1 mt-1">
              <input
                v-model.number="vitalsForm.tb"
                type="number"
                class="w-20 h-8 text-center font-bold text-[14px] border border-[#CBD5E1] rounded bg-white"
              />
              <span class="text-[11px] text-[#64748B] ml-1">cm</span>
            </div>
          </div>

          <div class="p-3 rounded-[6px] bg-[#F8FAFC] border border-[#E2E8F0]">
            <span class="text-[11px] text-[#64748B] font-medium block">Skala Nyeri (NRS 0-10)</span>
            <div class="flex items-center gap-1 mt-1">
              <input
                v-model.number="vitalsForm.skala_nyeri"
                type="number"
                min="0"
                max="10"
                class="w-20 h-8 text-center font-bold text-[14px] border border-[#CBD5E1] rounded bg-white"
              />
              <span class="text-[11px] text-[#64748B] ml-1">/ 10</span>
            </div>
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label class="text-[12px] font-semibold text-[#334155] block mb-1">Tingkat Kesadaran</label>
            <BaseSelect
              v-model="vitalsForm.kesadaran"
              :options="[
                { value: 'Compos Mentis', label: 'Compos Mentis (Sadar Penuh)' },
                { value: 'Apatis', label: 'Apatis' },
                { value: 'Somnolen', label: 'Somnolen (Mengantuk)' },
                { value: 'Sopor', label: 'Sopor' },
                { value: 'Koma', label: 'Koma' }
              ]"
            />
          </div>

          <div>
            <label class="text-[12px] font-semibold text-[#334155] block mb-1">Kategori Triase</label>
            <BaseSelect
              v-model="vitalsForm.triase_level"
              :options="[
                { value: 'HIJAU', label: 'HIJAU - Tidak Gawat Darurat (Rawat Jalan Standar)' },
                { value: 'KUNING', label: 'KUNING - Memerlukan Penanganan Segera' },
                { value: 'MERAH', label: 'MERAH - Gawat Darurat Mengancam Jiwa' }
              ]"
            />
          </div>
        </div>
      </div>

      <!-- ========================================================================= -->
      <!-- TAB 2: CPPT (SOAP FORM & RIWAYAT CPPT PER KUNJUNGAN)                       -->
      <!-- ========================================================================= -->
      <div v-if="activeTab === 'cppt'" class="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <!-- Left 2 Cols: Doctor SOAP Form -->
        <div class="lg:col-span-2 bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs space-y-4">
          <div class="flex items-center justify-between border-b border-[#F1F5F9] pb-3">
            <div>
              <h2 class="text-[15px] font-bold text-[#0F172A]">Catatan Perkembangan Pasien Terintegrasi (CPPT Dokter)</h2>
              <p class="text-[12px] text-[#64748B]">Format dokumentasi klinis SOAP dokter spesialis rawat jalan.</p>
            </div>
            <BaseButton variant="primary" size="sm" :loading="isSaving" @click="saveSOAP">
              Simpan CPPT
            </BaseButton>
          </div>

          <!-- S (Subjektif) -->
          <div>
            <label class="text-[12px] font-bold text-[#0F172A] flex items-center gap-1.5 mb-1">
              <span class="w-5 h-5 rounded bg-[#0284C7] text-white flex items-center justify-center text-[11px] font-bold">S</span>
              Subjektif (Anamnesa & Keluhan Pasien)
            </label>
            <textarea
              v-model="soapForm.subjektif"
              rows="3"
              class="w-full text-[13px] border border-[#CBD5E1] rounded-[6px] p-2.5 outline-none focus:border-[#0284C7] focus:ring-2 focus:ring-[rgba(2,132,199,0.15)] transition-all resize-none"
              placeholder="Keluhan utama, riwayat penyakit sekarang, onset, keluhan penyerta..."
            ></textarea>
          </div>

          <!-- O (Objektif) -->
          <div>
            <label class="text-[12px] font-bold text-[#0F172A] flex items-center gap-1.5 mb-1">
              <span class="w-5 h-5 rounded bg-[#0284C7] text-white flex items-center justify-center text-[11px] font-bold">O</span>
              Objektif (Pemeriksaan Fisik & Penunjang)
            </label>
            <textarea
              v-model="soapForm.objektif"
              rows="3"
              class="w-full text-[13px] border border-[#CBD5E1] rounded-[6px] p-2.5 outline-none focus:border-[#0284C7] focus:ring-2 focus:ring-[rgba(2,132,199,0.15)] transition-all resize-none"
              placeholder="Pemeriksaan kepala, leher, thorax, cor, pulmo, abdomen, ekstremitas..."
            ></textarea>
          </div>

          <!-- A (Asesmen) -->
          <div>
            <label class="text-[12px] font-bold text-[#0F172A] flex items-center gap-1.5 mb-1">
              <span class="w-5 h-5 rounded bg-[#0284C7] text-white flex items-center justify-center text-[11px] font-bold">A</span>
              Asesmen (Diagnosis Kerja / Diferensial)
            </label>
            <textarea
              v-model="soapForm.asesmen"
              rows="2"
              class="w-full text-[13px] border border-[#CBD5E1] rounded-[6px] p-2.5 outline-none focus:border-[#0284C7] focus:ring-2 focus:ring-[rgba(2,132,199,0.15)] transition-all resize-none"
              placeholder="Diagnosis kerja utama klinis..."
            ></textarea>
          </div>

          <!-- P (Planning) -->
          <div>
            <label class="text-[12px] font-bold text-[#0F172A] flex items-center gap-1.5 mb-1">
              <span class="w-5 h-5 rounded bg-[#0284C7] text-white flex items-center justify-center text-[11px] font-bold">P</span>
              Planning (Rencana Tindakan, Terapi & Edukasi)
            </label>
            <textarea
              v-model="soapForm.planning"
              rows="2"
              class="w-full text-[13px] border border-[#CBD5E1] rounded-[6px] p-2.5 outline-none focus:border-[#0284C7] focus:ring-2 focus:ring-[rgba(2,132,199,0.15)] transition-all resize-none"
              placeholder="Rencana medikamentosa, tindakan medis, edukasi pasien, tanggal kontrol..."
            ></textarea>
          </div>
        </div>

        <!-- Right 1 Col: Riwayat CPPT (Per Kunjungan Sebelumnya) -->
        <div class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs space-y-3">
          <div class="flex items-center justify-between border-b border-[#F1F5F9] pb-2">
            <h3 class="text-[13px] font-bold text-[#0F172A] flex items-center gap-1.5">
              <Calendar class="w-4 h-4 text-[#0284C7]" />
              Riwayat CPPT Pasien
            </h3>
            <span class="text-[11px] font-bold px-1.5 py-0.5 rounded bg-[#F1F5F9] text-[#64748B]">
              {{ patientHistory.length }} Riwayat
            </span>
          </div>

          <div v-if="patientHistory.length === 0" class="text-center py-8 text-[#94A3B8]">
            <FileText class="w-8 h-8 mx-auto mb-1.5 text-[#CBD5E1]" />
            <p class="text-[12px]">Belum ada riwayat CPPT pada kunjungan sebelumnya.</p>
          </div>

          <div v-else class="space-y-3 max-h-[520px] overflow-y-auto pr-1">
            <div
              v-for="rec in patientHistory"
              :key="rec.id"
              class="p-3 rounded-[6px] border border-[#E2E8F0] bg-[#F8FAFC] text-left space-y-1.5 text-[12px]"
            >
              <div class="flex items-center justify-between">
                <span class="font-bold text-[#0284C7]">
                  {{ new Date(rec.created_at).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' }) }}
                </span>
                <span class="text-[11px] font-medium text-[#64748B]">{{ rec.doctor_name || 'Dokter Spesialis' }}</span>
              </div>

              <div class="text-[#334155] border-t border-[#E2E8F0] pt-1.5 space-y-1">
                <p><span class="font-bold text-[#0F172A]">S:</span> {{ rec.subjektif }}</p>
                <p><span class="font-bold text-[#0F172A]">O:</span> {{ rec.objektif }}</p>
                <p><span class="font-bold text-[#0F172A]">A:</span> {{ rec.asesmen }} <span v-if="rec.icd10_code" class="font-mono text-[#0369A1]">({{ rec.icd10_code }})</span></p>
                <p><span class="font-bold text-[#0F172A]">P:</span> {{ rec.planning }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- ========================================================================= -->
      <!-- TAB 3: DIAGNOSA (ICD-10 & ICD-9)                                          -->
      <!-- ========================================================================= -->
      <div v-if="activeTab === 'diagnosa'" class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs space-y-4">
        <div class="flex items-center justify-between border-b border-[#F1F5F9] pb-3">
          <div>
            <h2 class="text-[15px] font-bold text-[#0F172A]">Kodifikasi Diagnosis Medis (ICD-10 & ICD-9-CM)</h2>
            <p class="text-[12px] text-[#64748B]">Standar kode diagnosa internasional untuk klaim asuransi dan rekam medis.</p>
          </div>
          <BaseButton variant="primary" size="sm" :loading="isSaving" @click="saveSOAP">
            Simpan Diagnosis
          </BaseButton>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <!-- ICD-10 Diagnosis Utama -->
          <div class="space-y-1.5">
            <label class="text-[12px] font-bold text-[#0F172A]">ICD-10 Diagnosis Primer</label>
            <BaseSelect
              v-model="diagnosaForm.icd10_primary"
              :options="icd10List.map(icd => ({
                value: icd.code,
                label: `${icd.code} - ${icd.name_id}`,
                badge: icd.code
              }))"
              placeholder="Cari kode ICD-10 primer..."
              searchable
              @change="(val) => {
                const found = icd10List.find(i => i.code === val)
                if (found) diagnosaForm.icd10_primary_desc = found.name_id
              }"
            />
            <p v-if="diagnosaForm.icd10_primary_desc" class="text-[12px] text-[#0369A1] font-medium bg-[#E0F2FE] p-2 rounded">
              Deskripsi: {{ diagnosaForm.icd10_primary_desc }}
            </p>
          </div>

          <!-- ICD-9 Prosedur Klinis -->
          <div class="space-y-1.5">
            <label class="text-[12px] font-bold text-[#0F172A]">ICD-9-CM Prosedur Diagnostik / Tindakan</label>
            <BaseSelect
              v-model="diagnosaForm.icd9_code"
              :options="icd9List.map(icd => ({
                value: icd.code,
                label: `${icd.code} - ${icd.name_id}`,
                badge: icd.code
              }))"
              placeholder="Cari kode ICD-9 tindakan..."
              searchable
              @change="(val) => {
                const found = icd9List.find(i => i.code === val)
                if (found) diagnosaForm.icd9_desc = found.name_id
              }"
            />
            <p v-if="diagnosaForm.icd9_desc" class="text-[12px] text-[#059669] font-medium bg-[#ECFDF5] p-2 rounded">
              Deskripsi Prosedur: {{ diagnosaForm.icd9_desc }}
            </p>
          </div>
        </div>

        <!-- Catatan Diagnosis Dokter -->
        <div>
          <label class="text-[12px] font-bold text-[#0F172A] block mb-1">Catatan Tambahan Diagnosis Klinis / Prognosis</label>
          <textarea
            v-model="diagnosaForm.catatan_diagnosa"
            rows="2"
            class="w-full text-[13px] border border-[#CBD5E1] rounded-[6px] p-2.5 outline-none focus:border-[#0284C7] focus:ring-2 focus:ring-[rgba(2,132,199,0.15)]"
            placeholder="Catatan diferensial diagnosa, derajat keparahan, prognosis pasien..."
          ></textarea>
        </div>
      </div>

      <!-- ========================================================================= -->
      <!-- TAB 4: TINDAKAN DI RUANGAN POLI & TARIF                                   -->
      <!-- ========================================================================= -->
      <div v-if="activeTab === 'tindakan'" class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs space-y-4">
        <div class="flex items-center justify-between border-b border-[#F1F5F9] pb-3">
          <div>
            <h2 class="text-[15px] font-bold text-[#0F172A]">Tindakan Medis di Ruangan Poli</h2>
            <p class="text-[12px] text-[#64748B]">Tindakan klinis yang dilakukan dokter di poli dan otomatis tercatat ke billing pasien.</p>
          </div>
        </div>

        <!-- Add Procedure Form -->
        <div class="p-3.5 bg-[#F8FAFC] border border-[#E2E8F0] rounded-[8px] space-y-3">
          <span class="text-[12px] font-bold text-[#0F172A] block">Tambah Tindakan Medis</span>
          <div class="grid grid-cols-1 md:grid-cols-4 gap-3 items-end">
            <div class="md:col-span-2">
              <label class="text-[11px] font-semibold text-[#475569] block mb-1">Pilih Tindakan & Tarif</label>
              <BaseSelect
                v-model="selectedProcedureID"
                :options="masterProcedures.map(p => ({
                  value: p.id,
                  label: `${p.name} - Rp ${p.tariff.toLocaleString('id-ID')}`,
                  badge: p.category
                }))"
                placeholder="-- Pilih Tindakan Medis Poli --"
                searchable
              />
            </div>

            <div>
              <label class="text-[11px] font-semibold text-[#475569] block mb-1">Jumlah (Qty)</label>
              <input
                v-model.number="procedureQty"
                type="number"
                min="1"
                class="w-full h-[36px] px-3 text-[13px] border border-[#CBD5E1] rounded-[6px] bg-white font-bold"
              />
            </div>

            <div>
              <BaseButton
                variant="primary"
                size="md"
                class="w-full h-[36px] flex items-center justify-center gap-1"
                :loading="isSaving"
                @click="addProcedure"
              >
                <Plus class="w-4 h-4" />
                Tambah Tindakan
              </BaseButton>
            </div>
          </div>
        </div>

        <!-- Performed Procedures Table -->
        <div class="border border-[#E2E8F0] rounded-[6px] overflow-hidden">
          <table class="w-full text-left text-[12px]">
            <thead class="bg-[#F8FAFC] text-[#475569] border-b border-[#E2E8F0]">
              <tr>
                <th class="p-2.5 font-bold">Nama Tindakan Medis</th>
                <th class="p-2.5 font-bold text-right">Tarif Satuan</th>
                <th class="p-2.5 font-bold text-center">Qty</th>
                <th class="p-2.5 font-bold text-right">Subtotal</th>
                <th class="p-2.5 font-bold text-center">Waktu Input</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-[#F1F5F9]">
              <tr v-if="visitProcedures.length === 0">
                <td colspan="5" class="p-4 text-center text-[#94A3B8]">
                  Belum ada tindakan medis yang ditambahkan pada kunjungan ini.
                </td>
              </tr>
              <tr v-for="vp in visitProcedures" :key="vp.id" class="hover:bg-[#F8FAFC]">
                <td class="p-2.5 font-semibold text-[#0F172A]">{{ vp.procedure_name }}</td>
                <td class="p-2.5 text-right font-mono tabular-nums">Rp {{ vp.tariff.toLocaleString('id-ID') }}</td>
                <td class="p-2.5 text-center font-bold font-mono">{{ vp.qty }}</td>
                <td class="p-2.5 text-right font-bold text-[#0284C7] font-mono tabular-nums">Rp {{ vp.subtotal.toLocaleString('id-ID') }}</td>
                <td class="p-2.5 text-center text-[#64748B] text-[11px]">{{ new Date(vp.created_at).toLocaleTimeString('id-ID') }} WIB</td>
              </tr>
            </tbody>
            <tfoot v-if="visitProcedures.length > 0" class="bg-[#F8FAFC] font-bold border-t border-[#E2E8F0]">
              <tr>
                <td colspan="3" class="p-2.5 text-right text-[#0F172A]">Total Biaya Tindakan:</td>
                <td class="p-2.5 text-right text-[#0284C7] font-mono text-[13px] tabular-nums">
                  Rp {{ calculatedBilling.totalTindakan.toLocaleString('id-ID') }}
                </td>
                <td></td>
              </tr>
            </tfoot>
          </table>
        </div>
      </div>

      <!-- ========================================================================= -->
      <!-- TAB 5: RUJUK (SURAT RUJUKAN EKSTERNAL KE RS LAIN / SPESIALIS LAIN)       -->
      <!-- ========================================================================= -->
      <div v-if="activeTab === 'rujuk'" class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs space-y-4">
        <div class="flex items-center justify-between border-b border-[#F1F5F9] pb-3">
          <div>
            <h2 class="text-[15px] font-bold text-[#0F172A]">Surat Rujukan Eksternal</h2>
            <p class="text-[12px] text-[#64748B]">Penerbitan surat rujukan ke Rumah Sakit Umum Pusat atau Dokter Spesialis lain.</p>
          </div>
          <BaseButton variant="primary" size="sm" @click="isPreviewLetterModalOpen = true">
            <Eye class="w-3.5 h-3.5 mr-1" />
            Preview & Cetak Rujukan
          </BaseButton>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label class="text-[12px] font-semibold text-[#334155] block mb-1">Rumah Sakit / Faskes Tujuan *</label>
            <BaseSelect
              v-model="rujukForm.rs_tujuan"
              :options="faskesOptions"
              placeholder="-- Pilih RS / Faskes Rujukan --"
              searchable
            />
          </div>

          <div>
            <label class="text-[12px] font-semibold text-[#334155] block mb-1">Poli / Spesialisasi Rujukan *</label>
            <BaseSelect
              v-model="rujukForm.poli_tujuan"
              :options="poliRujukanOptions"
              placeholder="-- Pilih Poli / Spesialisasi Rujukan --"
              searchable
            />
          </div>

          <div>
            <label class="text-[12px] font-semibold text-[#334155] block mb-1">Dokter Spesialis Tujuan</label>
            <BaseSelect
              v-model="rujukForm.dokter_tujuan"
              :options="spesialisOptions"
              placeholder="-- Pilih Dokter Spesialis Konsultan --"
              searchable
              @change="onSpesialisChange"
            />
          </div>

          <div>
            <label class="text-[12px] font-semibold text-[#334155] block mb-1">Alasan / Indikasi Rujukan</label>
            <BaseInput v-model="rujukForm.alasan_rujuk" placeholder="Alasan rujukan pasien..." />
          </div>
        </div>

        <div>
          <label class="text-[12px] font-semibold text-[#334155] block mb-1">Ringkasan Klinis & Terapi Sementara</label>
          <textarea
            v-model="rujukForm.ringkasan_anamnesa"
            rows="3"
            class="w-full text-[13px] border border-[#CBD5E1] rounded-[6px] p-2.5 outline-none focus:border-[#0284C7]"
            placeholder="Ringkasan temuan pemeriksaan fisik dan terapi sementara yang telah diberikan di klinik..."
          ></textarea>
        </div>
      </div>

      <!-- ========================================================================= -->
      <!-- TAB 6: RUJUK PENUNJANG (ORDER PEMERIKSAAN LAB, RAD, FISIOTERAPI, EKG)     -->
      <!-- ========================================================================= -->
      <div v-if="activeTab === 'penunjang'" class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs space-y-4">
        <div class="flex items-center justify-between border-b border-[#F1F5F9] pb-3">
          <div>
            <h2 class="text-[15px] font-bold text-[#0F172A]">Rujuk Pemeriksaan Penunjang (Lab / Radiologi / Fisio / EKG)</h2>
            <p class="text-[12px] text-[#64748B]">Order pemeriksaan diagnostik penunjang ke instalasi laboratorium medis dan radiologi.</p>
          </div>
        </div>

        <!-- Add Order Form -->
        <div class="p-3.5 bg-[#F8FAFC] border border-[#E2E8F0] rounded-[8px] space-y-3">
          <span class="text-[12px] font-bold text-[#0F172A] block">Form Permintaan Pemeriksaan Penunjang</span>
          <div class="grid grid-cols-1 md:grid-cols-3 gap-3 items-end">
            <div>
              <label class="text-[11px] font-semibold text-[#475569] block mb-1">Instalasi Penunjang</label>
              <BaseSelect
                v-model="penunjangForm.jenis_order"
                :options="[
                  { value: 'LAB', label: 'Laboratorium Medis (Darah, Urin)' },
                  { value: 'RAD', label: 'Radiologi / X-Ray (Foto Thorax, USG)' },
                  { value: 'FISIOTERAPI', label: 'Fisioterapi & Tumbuh Kembang' },
                  { value: 'EKG', label: 'Elektrokardiografi (EKG 12 Lead)' }
                ]"
              />
            </div>

            <div>
              <label class="text-[11px] font-semibold text-[#475569] block mb-1">Nama Pemeriksaan / Tindakan</label>
              <BaseInput
                v-model="penunjangForm.nama_tindakan"
                placeholder="Contoh: Darah Lengkap, Rontgen Thorax..."
              />
            </div>

            <div>
              <BaseButton
                variant="primary"
                size="md"
                class="w-full h-[36px] flex items-center justify-center gap-1"
                :loading="isSaving"
                @click="addAncillaryOrder"
              >
                <Send class="w-3.5 h-3.5" />
                Kirim Order Penunjang
              </BaseButton>
            </div>
          </div>

          <div>
            <label class="text-[11px] font-semibold text-[#475569] block mb-1">Indikasi Klinis / Catatan Petugas Penunjang</label>
            <input
              v-model="penunjangForm.catatan_klinis"
              type="text"
              class="w-full h-[34px] px-3 text-[12px] border border-[#CBD5E1] rounded-[6px] bg-white"
              placeholder="Contoh: Curiga ISPA / Evaluasi respon terapi..."
            />
          </div>
        </div>

        <!-- Orders Table -->
        <div class="border border-[#E2E8F0] rounded-[6px] overflow-hidden">
          <table class="w-full text-left text-[12px]">
            <thead class="bg-[#F8FAFC] text-[#475569] border-b border-[#E2E8F0]">
              <tr>
                <th class="p-2.5 font-bold">No. Order</th>
                <th class="p-2.5 font-bold">Instalasi</th>
                <th class="p-2.5 font-bold">Pemeriksaan</th>
                <th class="p-2.5 font-bold text-right">Tarif</th>
                <th class="p-2.5 font-bold text-center">Status</th>
                <th class="p-2.5 font-bold">Hasil / Catatan</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-[#F1F5F9]">
              <tr v-if="visitOrders.length === 0">
                <td colspan="6" class="p-4 text-center text-[#94A3B8]">
                  Belum ada permintaan pemeriksaan penunjang untuk kunjungan ini.
                </td>
              </tr>
              <tr v-for="ord in visitOrders" :key="ord.id" class="hover:bg-[#F8FAFC]">
                <td class="p-2.5 font-mono text-[11px] text-[#0284C7] font-semibold">{{ ord.nomor_order }}</td>
                <td class="p-2.5 font-bold text-[#0F172A]">{{ ord.jenis_order }}</td>
                <td class="p-2.5 text-[#334155]">{{ ord.nama_tindakan }}</td>
                <td class="p-2.5 text-right font-mono tabular-nums">Rp {{ ord.tarif.toLocaleString('id-ID') }}</td>
                <td class="p-2.5 text-center">
                  <span
                    class="px-2 py-0.5 rounded text-[10px] font-bold"
                    :class="ord.status === 'SELESAI' ? 'bg-emerald-100 text-emerald-700' :
                            ord.status === 'DALAM_PROSES' ? 'bg-blue-100 text-blue-700' :
                            'bg-amber-100 text-amber-800'"
                  >
                    {{ ord.status }}
                  </span>
                </td>
                <td class="p-2.5 text-[11px] text-[#475569]">
                  {{ ord.hasil_pemeriksaan || ord.catatan_klinis || '-' }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- ========================================================================= -->
      <!-- TAB 7: RESEP (ORDER FORMULARIUM OBAT & RACIKAN)                          -->
      <!-- ========================================================================= -->
      <div v-if="activeTab === 'resep'" class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs space-y-4">
        <div class="flex items-center justify-between border-b border-[#F1F5F9] pb-3">
          <div>
            <h2 class="text-[15px] font-bold text-[#0F172A]">Order Resep Obat Elektronik (e-Resep)</h2>
            <p class="text-[12px] text-[#64748B]">Formularium obat klinis rawat jalan, signa aturan pakai, dan perhitungan harga resep.</p>
          </div>
          <BaseButton
            variant="primary"
            size="sm"
            class="!bg-[#0284C7]"
            :loading="isSaving"
            @click="submitPrescription"
          >
            <Send class="w-3.5 h-3.5 mr-1" />
            Kirim e-Resep ke Farmasi
          </BaseButton>
        </div>

        <!-- Add Medicine to Prescription Form -->
        <div class="p-3.5 bg-[#F8FAFC] border border-[#E2E8F0] rounded-[8px] space-y-3">
          <div class="flex items-center justify-between">
            <span class="text-[12px] font-bold text-[#0F172A]">Tambah Obat ke Formularium e-Resep</span>
            <label class="flex items-center gap-1.5 text-[12px] font-semibold text-[#0369A1] cursor-pointer">
              <input v-model="isRacikan" type="checkbox" class="rounded text-[#0284C7]" />
              Sediaan Racikan (Puyer / Kapsul)
            </label>
          </div>

          <div v-if="isRacikan" class="p-2 rounded bg-sky-50 border border-sky-200">
            <label class="text-[11px] font-bold text-[#0369A1] block mb-1">Nama Racikan</label>
            <BaseInput v-model="racikanNama" placeholder="Contoh: Puyer Batuk Pilek No. 1" />
          </div>

          <div class="grid grid-cols-1 md:grid-cols-4 gap-3 items-end">
            <!-- Searchable Dropdown for Medicines -->
            <div class="md:col-span-2">
              <label class="text-[11px] font-semibold text-[#475569] block mb-1">Pilih Obat / Sediaan</label>
              <BaseSelect
                v-model="selectedMedicineID"
                :options="masterMedicines.map(m => ({
                  value: m.id,
                  label: `${m.item_name} (Stok: ${m.total_stock ?? 0} ${m.unit || 'pcs'}) - Rp ${(m.sell_price || 0).toLocaleString('id-ID')}`,
                  badge: m.category_name || (m.is_medicine ? 'Obat' : 'BHP')
                }))"
                placeholder="-- Pilih Obat / Sediaan --"
                searchable
              />
            </div>

            <div>
              <label class="text-[11px] font-semibold text-[#475569] block mb-1">Signa / Dosis</label>
              <BaseInput v-model="recipeSigna" placeholder="3 x 1 sendok teh" />
            </div>

            <div>
              <label class="text-[11px] font-semibold text-[#475569] block mb-1">Jumlah (Qty)</label>
              <input
                v-model.number="recipeQty"
                type="number"
                min="1"
                class="w-full h-[36px] px-3 text-[13px] border border-[#CBD5E1] rounded-[6px] bg-white font-bold"
              />
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-3 items-end">
            <div>
              <label class="text-[11px] font-semibold text-[#475569] block mb-1">Aturan Pakai</label>
              <BaseSelect
                v-model="recipeAturan"
                :options="[
                  { value: 'Sesudah Makan', label: 'Sesudah Makan' },
                  { value: 'Sebelum Makan', label: 'Sebelum Makan' },
                  { value: 'Bersama Makan', label: 'Bersama Makan' },
                  { value: 'Malam Sebelum Tidur', label: 'Malam Sebelum Tidur' },
                  { value: 'Bila Perlu / Demam', label: 'Bila Perlu / Demam' }
                ]"
              />
            </div>

            <div class="md:col-span-2 flex justify-end">
              <BaseButton variant="outline" size="md" class="h-[36px]" @click="addPrescriptionItem">
                <Plus class="w-4 h-4 mr-1" />
                Tambahkan ke Daftar Resep
              </BaseButton>
            </div>
          </div>
        </div>

        <!-- Prescribed Items Table -->
        <div class="border border-[#E2E8F0] rounded-[6px] overflow-hidden">
          <table class="w-full text-left text-[12px]">
            <thead class="bg-[#F8FAFC] text-[#475569] border-b border-[#E2E8F0]">
              <tr>
                <th class="p-2.5 font-bold">Nama Obat / Sediaan</th>
                <th class="p-2.5 font-bold">Signa & Aturan</th>
                <th class="p-2.5 font-bold text-center">Jumlah</th>
                <th class="p-2.5 font-bold text-right">Harga Satuan</th>
                <th class="p-2.5 font-bold text-right">Subtotal</th>
                <th class="p-2.5 font-bold text-center w-12">Aksi</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-[#F1F5F9]">
              <tr v-if="prescriptionItems.length === 0">
                <td colspan="6" class="p-4 text-center text-[#94A3B8]">
                  Belum ada obat yang dimasukkan ke resep.
                </td>
              </tr>
              <tr v-for="(it, idx) in prescriptionItems" :key="idx" class="hover:bg-[#F8FAFC]">
                <td class="p-2.5 font-semibold text-[#0F172A]">
                  {{ it.item_name }}
                  <span v-if="it.is_racikan" class="ml-1 text-[10px] px-1.5 py-0.2 rounded bg-purple-100 text-purple-700 font-bold">
                    {{ it.racikan_nama || 'Racikan' }}
                  </span>
                </td>
                <td class="p-2.5 text-[#334155]">{{ it.signa }} - {{ it.aturan_pakai }}</td>
                <td class="p-2.5 text-center font-bold font-mono">{{ it.jumlah }}</td>
                <td class="p-2.5 text-right font-mono tabular-nums">Rp {{ it.harga_satuan.toLocaleString('id-ID') }}</td>
                <td class="p-2.5 text-right font-bold text-[#0284C7] font-mono tabular-nums">Rp {{ it.subtotal.toLocaleString('id-ID') }}</td>
                <td class="p-2.5 text-center">
                  <button
                    type="button"
                    class="p-1 text-red-500 hover:text-red-700 rounded transition-colors"
                    title="Hapus Obat"
                    @click="removePrescriptionItem(idx)"
                  >
                    <Trash2 class="w-3.5 h-3.5" />
                  </button>
                </td>
              </tr>
            </tbody>
            <tfoot v-if="prescriptionItems.length > 0" class="bg-[#F8FAFC] font-bold border-t border-[#E2E8F0]">
              <tr>
                <td colspan="4" class="p-2.5 text-right text-[#0F172A]">Total Harga Obat e-Resep:</td>
                <td class="p-2.5 text-right text-[#0284C7] font-mono text-[13px] tabular-nums">
                  Rp {{ calculatedBilling.totalResep.toLocaleString('id-ID') }}
                </td>
                <td></td>
              </tr>
            </tfoot>
          </table>
        </div>
      </div>

      <!-- ========================================================================= -->
      <!-- TAB 8: SURAT KLINIS (5 TEMPLATE STANDAR KLINIK)                           -->
      <!-- ========================================================================= -->
      <div v-if="activeTab === 'surat'" class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs space-y-4">
        <div class="flex items-center justify-between border-b border-[#F1F5F9] pb-3">
          <div>
            <h2 class="text-[15px] font-bold text-[#0F172A]">Surat Keterangan Medis (5 Template Klinik)</h2>
            <p class="text-[12px] text-[#64748B]">Pilih template surat resmi klinik, sesuaikan parameter, lalu cetak surat.</p>
          </div>
          <BaseButton variant="primary" size="sm" @click="isPreviewLetterModalOpen = true">
            <Eye class="w-3.5 h-3.5 mr-1" />
            Preview & Cetak Surat
          </BaseButton>
        </div>

        <!-- 5 Template Selector Chips -->
        <div class="flex items-center gap-2 flex-wrap">
          <button
            v-for="st in [
              { id: 'SAKIT', label: '1. Surat Keterangan Sakit' },
              { id: 'SEHAT', label: '2. Surat Keterangan Sehat' },
              { id: 'BEBAS_NARKOBA', label: '3. Bebas Narkoba & Buta Warna' },
              { id: 'LAIK_TERBANG', label: '4. Surat Laik Terbang / Hamil' },
              { id: 'IZIN_KERJA', label: '5. Surat Izin Kerja / Sekolah' }
            ]"
            :key="st.id"
            type="button"
            :class="[
              'px-3 py-1.5 rounded-[6px] text-[12px] font-bold transition-all border',
              selectedLetterType === st.id
                ? 'bg-[#E0F2FE] border-[#0284C7] text-[#0369A1] shadow-2xs'
                : 'bg-white border-[#CBD5E1] text-[#475569] hover:bg-[#F8FAFC]'
            ]"
            @click="selectedLetterType = st.id as any"
          >
            {{ st.label }}
          </button>
        </div>

        <!-- Template Form Parameter Fields -->
        <div class="p-4 bg-[#F8FAFC] border border-[#E2E8F0] rounded-[8px] space-y-3">
          <!-- Template 1: Sakit -->
          <div v-if="selectedLetterType === 'SAKIT'" class="grid grid-cols-1 md:grid-cols-3 gap-3">
            <div>
              <label class="text-[12px] font-semibold text-[#334155] block mb-1">Lama Istirahat (Hari)</label>
              <input v-model.number="letterForm.lama_istirahat_hari" type="number" min="1" class="w-full h-[36px] px-3 text-[13px] border border-[#CBD5E1] rounded-[6px] bg-white font-bold" />
            </div>
            <div>
              <label class="text-[12px] font-semibold text-[#334155] block mb-1">Tanggal Mulai Istirahat</label>
              <BaseDatePicker v-model="letterForm.tgl_mulai" />
            </div>
            <div>
              <label class="text-[12px] font-semibold text-[#334155] block mb-1">Tanggal Selesai Istirahat</label>
              <BaseDatePicker v-model="letterForm.tgl_selesai" />
            </div>
          </div>

          <!-- Template 2: Sehat -->
          <div v-else-if="selectedLetterType === 'SEHAT'" class="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div>
              <label class="text-[12px] font-semibold text-[#334155] block mb-1">Keperluan Surat Keterangan Sehat</label>
              <BaseInput v-model="letterForm.keperluan" placeholder="Melamar Pekerjaan / Masuk Perguruan Tinggi..." />
            </div>
            <div>
              <label class="text-[12px] font-semibold text-[#334155] block mb-1">Kondisi Fisik Pasien</label>
              <BaseInput v-model="letterForm.kondisi_fisik" placeholder="Sehat jasmani dan rohani..." />
            </div>
          </div>

          <!-- Template 3: Bebas Narkoba & Buta Warna -->
          <div v-else-if="selectedLetterType === 'BEBAS_NARKOBA'" class="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div>
              <label class="text-[12px] font-semibold text-[#334155] block mb-1">Hasil Tes Skrining Narkoba Urin (5 Parameter)</label>
              <BaseInput v-model="letterForm.hasil_tes_narkoba" />
            </div>
            <div>
              <label class="text-[12px] font-semibold text-[#334155] block mb-1">Hasil Tes Persepsi Warna (Ishihara)</label>
              <BaseInput v-model="letterForm.hasil_buta_warna" />
            </div>
          </div>

          <!-- Template 4: Laik Terbang / Hamil -->
          <div v-else-if="selectedLetterType === 'LAIK_TERBANG'" class="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div>
              <label class="text-[12px] font-semibold text-[#334155] block mb-1">Usia Kehamilan (Minggu)</label>
              <input v-model.number="letterForm.usia_kehamilan_minggu" type="number" class="w-full h-[36px] px-3 text-[13px] border border-[#CBD5E1] rounded-[6px] bg-white font-bold" />
            </div>
            <div>
              <label class="text-[12px] font-semibold text-[#334155] block mb-1">Taksiran Tanggal Persalinan (HPL)</label>
              <BaseDatePicker v-model="letterForm.taksiran_partus" />
            </div>
          </div>

          <!-- Template 5: Izin Kerja / Sekolah -->
          <div v-else-if="selectedLetterType === 'IZIN_KERJA'" class="grid grid-cols-1 md:grid-cols-3 gap-3">
            <div>
              <label class="text-[12px] font-semibold text-[#334155] block mb-1">Dispensasi Tidak Masuk (Hari)</label>
              <input v-model.number="letterForm.lama_istirahat_hari" type="number" min="1" class="w-full h-[36px] px-3 text-[13px] border border-[#CBD5E1] rounded-[6px] bg-white font-bold" />
            </div>
            <div>
              <label class="text-[12px] font-semibold text-[#334155] block mb-1">Tanggal Mulai</label>
              <BaseDatePicker v-model="letterForm.tgl_mulai" />
            </div>
            <div>
              <label class="text-[12px] font-semibold text-[#334155] block mb-1">Tanggal Selesai</label>
              <BaseDatePicker v-model="letterForm.tgl_selesai" />
            </div>
          </div>
        </div>
      </div>

      <!-- ========================================================================= -->
      <!-- TAB: INFORMED CONSENT DIGITAL (FASE 1)                                    -->
      <!-- ========================================================================= -->
      <div v-if="activeTab === 'consent'" class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs space-y-4">
        <div class="flex items-center justify-between border-b border-[#F1F5F9] pb-3">
          <div>
            <h2 class="text-[15px] font-bold text-[#0F172A] flex items-center gap-2">
              <ShieldCheck class="w-4 h-4 text-emerald-600" />
              Informed Consent Digital (Persetujuan Tindakan Medis)
            </h2>
            <p class="text-[12px] text-[#64748B]">Perlindungan medikolegal dengan tanda tangan digital canvas dan validasi QR Code resmi.</p>
          </div>
          <BaseButton variant="primary" size="sm" @click="openConsentModal">
            <Plus class="w-3.5 h-3.5 mr-1" />
            + Buat Informed Consent Baru
          </BaseButton>
        </div>

        <!-- Daftar Dokumen Consent Terbit -->
        <div v-if="consentList.length > 0" class="divide-y divide-slate-100">
          <div
            v-for="c in consentList"
            :key="c.id"
            class="py-3 flex items-center justify-between hover:bg-slate-50/60 p-2 rounded-xl transition-all"
          >
            <div>
              <div class="flex items-center space-x-2">
                <span
                  :class="[
                    'px-2.5 py-0.5 rounded-full text-[11px] font-bold',
                    c.consent_type.includes('PENOLAKAN')
                      ? 'bg-rose-50 text-rose-700 border border-rose-200'
                      : 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                  ]"
                >
                  {{ c.consent_type }}
                </span>
                <span class="font-bold text-slate-800 text-sm">{{ c.action_name }}</span>
              </div>
              <div class="text-xs text-slate-500 mt-1">
                Penandatangan: <strong class="text-slate-700">{{ c.signer_name }}</strong> ({{ c.signer_relation }}) • Dokter: {{ c.doctor_name || 'dr. Pemeriksa' }} • Token: <code class="text-sky-600 font-mono">{{ c.qr_code_token }}</code>
              </div>
            </div>

            <div class="flex items-center space-x-2">
              <BaseButton variant="outline" size="sm" @click="previewConsent(c)">
                <Printer class="w-3.5 h-3.5 mr-1" />
                Cetak Lembar Informed Consent
              </BaseButton>
            </div>
          </div>
        </div>

        <div v-else class="text-center py-10 bg-slate-50 border border-dashed border-slate-200 rounded-2xl">
          <ShieldCheck class="w-10 h-10 text-slate-300 mx-auto mb-2" />
          <p class="text-sm font-semibold text-slate-700">Belum Ada Formulir Informed Consent</p>
          <p class="text-xs text-slate-400 mt-0.5">Klik tombol di atas untuk membuat persetujuan / penolakan tindakan medis baru.</p>
        </div>
      </div>

      <!-- ========================================================================= -->
      <!-- TAB 9: RESUME MEDIS RAWAT JALAN                                           -->
      <!-- ========================================================================= -->
      <div v-if="activeTab === 'resume'" class="bg-white border border-[#E2E8F0] rounded-[8px] p-4 shadow-xs space-y-4">
        <div class="flex items-center justify-between border-b border-[#F1F5F9] pb-3">
          <div>
            <h2 class="text-[15px] font-bold text-[#0F172A]">Ringkasan Pasien Pulang (Resume Medis Rawat Jalan)</h2>
            <p class="text-[12px] text-[#64748B]">Dokumen resmi ringkasan akhir perawatan untuk arsip rekam medis dan pasien.</p>
          </div>
          <BaseButton variant="primary" size="sm" @click="isPreviewResumeModalOpen = true">
            <Eye class="w-3.5 h-3.5 mr-1" />
            Preview & Cetak Resume Medis
          </BaseButton>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label class="text-[12px] font-semibold text-[#334155] block mb-1">Keluhan & Alasan Kunjungan</label>
            <textarea v-model="resumeForm.alasan_dirawat" rows="2" class="w-full text-[13px] border border-[#CBD5E1] rounded-[6px] p-2.5 outline-none focus:border-[#0284C7]"></textarea>
          </div>

          <div>
            <label class="text-[12px] font-semibold text-[#334155] block mb-1">Diagnosis Akhir (Utama & Sekunder)</label>
            <textarea v-model="resumeForm.diagnosis_akhir" rows="2" class="w-full text-[13px] border border-[#CBD5E1] rounded-[6px] p-2.5 outline-none focus:border-[#0284C7]"></textarea>
          </div>

          <div>
            <label class="text-[12px] font-semibold text-[#334155] block mb-1">Prosedur / Tindakan yang Dilakukan</label>
            <textarea v-model="resumeForm.prosedur_tindakan" rows="2" class="w-full text-[13px] border border-[#CBD5E1] rounded-[6px] p-2.5 outline-none focus:border-[#0284C7]"></textarea>
          </div>

          <div>
            <label class="text-[12px] font-semibold text-[#334155] block mb-1">Terapi Obat saat Pulang</label>
            <textarea v-model="resumeForm.obat_pulang" rows="2" class="w-full text-[13px] border border-[#CBD5E1] rounded-[6px] p-2.5 outline-none focus:border-[#0284C7]"></textarea>
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
          <div>
            <label class="text-[12px] font-semibold text-[#334155] block mb-1">Kondisi saat Pulang</label>
            <BaseSelect
              v-model="resumeForm.kondisi_pulang"
              :options="[
                { value: 'Membaik / Sembuh Klinis', label: 'Membaik / Sembuh Klinis' },
                { value: 'Memerlukan Kontrol Rutin', label: 'Memerlukan Kontrol Rutin' },
                { value: 'Dirujuk ke RS Lebih Tinggi', label: 'Dirujuk ke RS Lebih Tinggi' },
                { value: 'Pulang atas Permintaan Sendiri', label: 'Pulang atas Permintaan Sendiri' }
              ]"
            />
          </div>

          <div class="md:col-span-2">
            <label class="text-[12px] font-semibold text-[#334155] block mb-1">Tanggal Rencana Kontrol Ulang</label>
            <BaseDatePicker v-model="resumeForm.tgl_kontrol_ulang" />
          </div>
        </div>
      </div>
    </div>

    <!-- ========================================================================= -->
    <!-- MODAL 1: PREVIEW SURAT KLINIS (CETAK RESMI)                                -->
    <!-- ========================================================================= -->
    <BaseModal
      v-model="isPreviewLetterModalOpen"
      title="Preview Surat Keterangan Medis"
      max-width="2xl"
    >
      <div id="printable-letter" class="p-6 bg-white text-[#0F172A] space-y-4 border border-[#E2E8F0] rounded">
        <!-- Letterhead / Kop Surat -->
        <div class="border-b-2 border-[#0F172A] pb-3 text-center">
          <h2 class="text-[16px] font-black uppercase tracking-wider text-[#0284C7]">KLINIK PRATAMA SIMRS-ASOY</h2>
          <p class="text-[11px] text-[#475569]">Jl. Kesehatan No. 10, DI Yogyakarta • Telp: (0274) 555-889 • Email: info@simrs-asoy.med.id</p>
          <p class="text-[10px] text-[#64748B]">Izin Operasional Dinkes No: 445/1029/SK/2024</p>
        </div>

        <!-- Title -->
        <div class="text-center pt-2">
          <h3 class="text-[15px] font-black uppercase underline tracking-wide">
            {{ selectedLetterType === 'SAKIT' ? 'SURAT KETERANGAN SAKIT' :
               selectedLetterType === 'SEHAT' ? 'SURAT KETERANGAN KESEHATAN' :
               selectedLetterType === 'BEBAS_NARKOBA' ? 'SURAT KETERANGAN BEBAS NARKOBA' :
               selectedLetterType === 'LAIK_TERBANG' ? 'SURAT KETERANGAN LAIK TERBANG (FIT TO FLY)' :
               'SURAT KETERANGAN IZIN DISPENSASI' }}
          </h3>
          <span class="text-[11px] font-mono text-[#64748B]">Nomor: SKM/{{ new Date().getFullYear() }}/{{ Math.floor(1000 + Math.random() * 9000) }}</span>
        </div>

        <!-- Body -->
        <div class="text-[12px] space-y-2 text-justify leading-relaxed">
          <p>Yang bertanda tangan di bawah ini, Dokter Pemeriksa Klinik SIMRS-ASOY, menerangkan dengan sebenarnya bahwa:</p>
          <div class="pl-4 space-y-1 font-mono text-[11px]">
            <p><span class="inline-block w-36 font-sans">Nama Lengkap</span>: <strong class="font-sans text-[12px]">{{ currentVisit?.patient_name }}</strong></p>
            <p><span class="inline-block w-36 font-sans">Nomor Rekam Medis</span>: {{ currentVisit?.no_rm }}</p>
            <p><span class="inline-block w-36 font-sans">Jenis Kelamin</span>: {{ currentVisit?.jenis_kelamin === 'L' ? 'Laki-laki' : 'Perempuan' }}</p>
            <p><span class="inline-block w-36 font-sans">Penjamin Pasien</span>: {{ currentVisit?.guarantor_name }}</p>
          </div>

          <div v-if="selectedLetterType === 'SAKIT'" class="space-y-1 pt-1">
            <p>
              Berhubung sedang sakit, pasien memerlukan istirahat selama
              <strong>{{ letterForm.lama_istirahat_hari }} hari</strong>, terhitung mulai tanggal
              <strong>{{ letterForm.tgl_mulai }}</strong> s/d <strong>{{ letterForm.tgl_selesai }}</strong>.
            </p>
            <p class="text-[11px] text-[#64748B]">Diagnosis kerja: {{ soapForm.asesmen || 'Observasi Klinis Rawat Jalan' }}</p>
          </div>

          <div v-else-if="selectedLetterType === 'SEHAT'" class="space-y-1 pt-1">
            <p>
              Telah dilakukan pemeriksaan fisik lengkap dan saat ini dinyatakan dalam kondisi
              <strong>{{ letterForm.kondisi_fisik }}</strong>. Surat ini dipergunakan untuk keperluan:
              <strong>{{ letterForm.keperluan }}</strong>.
            </p>
          </div>

          <div v-else-if="selectedLetterType === 'BEBAS_NARKOBA'" class="space-y-1 pt-1">
            <p>
              Telah menjalani skrining toksikologi urin dan dinyatakan:
              <strong>{{ letterForm.hasil_tes_narkoba }}</strong>. Hasil pemeriksaan buta warna:
              <strong>{{ letterForm.hasil_buta_warna }}</strong>.
            </p>
          </div>

          <div v-else-if="selectedLetterType === 'LAIK_TERBANG'" class="space-y-1 pt-1">
            <p>
              Pasien dengan usia kehamilan <strong>{{ letterForm.usia_kehamilan_minggu }} minggu</strong> (Taksiran Partus: {{ letterForm.taksiran_partus }}),
              dinyatakan dalam keadaan stabil, tidak ada komplikasi obstetrik, dan <strong>LAIK TERBANG</strong> menggunakan transportasi udara.
            </p>
          </div>

          <div v-else-if="selectedLetterType === 'IZIN_KERJA'" class="space-y-1 pt-1">
            <p>
              Diberikan izin dispensasi tidak dapat masuk kerja / sekolah selama
              <strong>{{ letterForm.lama_istirahat_hari }} hari</strong> terhitung mulai {{ letterForm.tgl_mulai }} s/d {{ letterForm.tgl_selesai }}.
            </p>
          </div>

          <p class="pt-2">Demikian surat keterangan ini dibuat dengan sebenarnya untuk dipergunakan sebagaimana mestinya.</p>
        </div>

        <!-- Signature -->
        <div class="flex justify-between items-end pt-6">
          <div class="text-[10px] text-[#94A3B8] font-mono">
            Validitas: Dokumen Digital Terotentikasi SIMRS
          </div>
          <div class="text-center text-[12px]">
            <p class="text-[11px] text-[#64748B]">Yogyakarta, {{ new Date().toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' }) }}</p>
            <p class="font-bold text-[#0F172A] mt-1">Dokter Pemeriksa,</p>
            <div class="h-14 flex items-center justify-center">
              <span class="text-[11px] font-mono text-[#0284C7] border border-[#0284C7] px-2 py-0.5 rounded font-bold">DIGITALLY SIGNED</span>
            </div>
            <p class="font-bold text-[#0F172A] underline">{{ authStore.user?.nama_lengkap || 'dr. Spesialis Poliklinik' }}</p>
            <p class="text-[10px] text-[#64748B]">SIP: 503/445/SIP-D/2023</p>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="flex items-center justify-end gap-2">
          <BaseButton variant="outline" @click="isPreviewLetterModalOpen = false">Tutup</BaseButton>
          <BaseButton variant="primary" @click="triggerPrint">
            <Printer class="w-4 h-4 mr-1" />
            Cetak Surat
          </BaseButton>
        </div>
      </template>
    </BaseModal>

    <!-- ========================================================================= -->
    <!-- MODAL 2: PREVIEW RESUME MEDIS RAWAT JALAN                                  -->
    <!-- ========================================================================= -->
    <BaseModal
      v-model="isPreviewResumeModalOpen"
      title="Preview Ringkasan Pasien Pulang (Resume Medis)"
      max-width="2xl"
    >
      <div id="printable-resume" class="p-6 bg-white text-[#0F172A] space-y-4 border border-[#E2E8F0] rounded">
        <div class="border-b-2 border-[#0F172A] pb-3 text-center">
          <h2 class="text-[16px] font-black uppercase tracking-wider text-[#0284C7]">KLINIK PRATAMA SIMRS-ASOY</h2>
          <h3 class="text-[14px] font-bold uppercase underline mt-1">RESUME MEDIS RAWAT JALAN</h3>
          <span class="text-[10px] text-[#64748B]">No. Rawat: {{ currentVisit?.no_rawat }} • No. RM: {{ currentVisit?.no_rm }}</span>
        </div>

        <div class="grid grid-cols-2 gap-2 text-[12px] bg-[#F8FAFC] p-2.5 rounded border border-[#E2E8F0]">
          <div><span class="text-[#64748B]">Nama Pasien:</span> <strong>{{ currentVisit?.patient_name }}</strong></div>
          <div><span class="text-[#64748B]">Penjamin:</span> <strong>{{ currentVisit?.guarantor_name }}</strong></div>
          <div><span class="text-[#64748B]">Poli Tujuan:</span> <strong>{{ currentVisit?.poli_name }}</strong></div>
          <div><span class="text-[#64748B]">Tanggal Kunjungan:</span> <strong>{{ new Date().toLocaleDateString('id-ID') }}</strong></div>
        </div>

        <div class="space-y-2 text-[12px]">
          <div>
            <strong class="text-[#0284C7] block">1. Keluhan & Ringkasan Penyakit:</strong>
            <p class="text-[#334155]">{{ resumeForm.alasan_dirawat }}</p>
          </div>

          <div>
            <strong class="text-[#0284C7] block">2. Diagnosis Utama & Sekunder:</strong>
            <p class="text-[#334155] font-semibold">{{ resumeForm.diagnosis_akhir }}</p>
          </div>

          <div>
            <strong class="text-[#0284C7] block">3. Tindakan & Prosedur Medis:</strong>
            <p class="text-[#334155]">{{ resumeForm.prosedur_tindakan }}</p>
          </div>

          <div>
            <strong class="text-[#0284C7] block">4. Terapi Obat Pulang:</strong>
            <p class="text-[#334155]">{{ resumeForm.obat_pulang }}</p>
          </div>

          <div class="grid grid-cols-2 gap-2 bg-[#F1F5F9] p-2 rounded">
            <div>
              <span class="text-[#64748B] block text-[11px]">Kondisi Saat Pulang:</span>
              <strong class="text-[#0F172A]">{{ resumeForm.kondisi_pulang }}</strong>
            </div>
            <div>
              <span class="text-[#64748B] block text-[11px]">Rencana Kontrol Berikutnya:</span>
              <strong class="text-[#0F172A]">{{ resumeForm.tgl_kontrol_ulang }}</strong>
            </div>
          </div>
        </div>

        <div class="flex justify-end pt-4">
          <div class="text-center text-[12px]">
            <p class="text-[11px] text-[#64748B]">Dokter Penanggung Jawab Pelayanan,</p>
            <div class="h-12 flex items-center justify-center">
              <span class="text-[10px] font-mono text-[#0284C7] border border-[#0284C7] px-2 py-0.5 rounded font-bold">TERVERIFIKASI DPJP</span>
            </div>
            <p class="font-bold text-[#0F172A] underline">{{ authStore.user?.nama_lengkap || 'dr. Spesialis DPJP' }}</p>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="flex items-center justify-end gap-2">
          <BaseButton variant="outline" @click="isPreviewResumeModalOpen = false">Tutup</BaseButton>
          <BaseButton variant="primary" @click="triggerPrint">
            <Printer class="w-4 h-4 mr-1" />
            Cetak Resume Medis
          </BaseButton>
        </div>
      </template>
    </BaseModal>

    <!-- ========================================================================= -->
    <!-- MODAL 3: RINCIAN TOTAL BILLING PASIEN                                     -->
    <!-- ========================================================================= -->
    <BaseModal
      v-model="isBillingBreakdownOpen"
      title="Rincian Total Tagihan Pasien (Real-Time)"
      max-width="lg"
    >
      <div class="space-y-3 text-[13px]">
        <div class="flex items-center justify-between p-2.5 rounded bg-[#F8FAFC] border border-[#E2E8F0]">
          <span>Biaya Administrasi Pendaftaran Poli:</span>
          <span class="font-bold font-mono">Rp {{ calculatedBilling.biayaAdmin.toLocaleString('id-ID') }}</span>
        </div>

        <div class="flex items-center justify-between p-2.5 rounded bg-[#F8FAFC] border border-[#E2E8F0]">
          <span>Jasa Pemeriksaan & Konsultasi Dokter:</span>
          <span class="font-bold font-mono">Rp {{ calculatedBilling.jasaDokter.toLocaleString('id-ID') }}</span>
        </div>

        <div class="flex items-center justify-between p-2.5 rounded bg-[#F8FAFC] border border-[#E2E8F0]">
          <span>Tindakan Medis di Ruangan ({{ visitProcedures.length }} item):</span>
          <span class="font-bold font-mono text-[#0284C7]">Rp {{ calculatedBilling.totalTindakan.toLocaleString('id-ID') }}</span>
        </div>

        <div class="flex items-center justify-between p-2.5 rounded bg-[#F8FAFC] border border-[#E2E8F0]">
          <span>Pemeriksaan Penunjang Lab/Rad ({{ visitOrders.length }} item):</span>
          <span class="font-bold font-mono text-[#0284C7]">Rp {{ calculatedBilling.totalPenunjang.toLocaleString('id-ID') }}</span>
        </div>

        <div class="flex items-center justify-between p-2.5 rounded bg-[#F8FAFC] border border-[#E2E8F0]">
          <span>Obat & Alkes e-Resep ({{ prescriptionItems.length }} item):</span>
          <span class="font-bold font-mono text-[#0284C7]">Rp {{ calculatedBilling.totalResep.toLocaleString('id-ID') }}</span>
        </div>

        <div class="flex items-center justify-between p-3 rounded bg-[#E0F2FE] border border-[#BAE6FD] text-[#0369A1] font-bold text-[15px]">
          <span>Grand Total Billing Sementara:</span>
          <span class="font-mono tabular-nums">Rp {{ calculatedBilling.grandTotal.toLocaleString('id-ID') }}</span>
        </div>
      </div>

      <template #footer>
        <BaseButton variant="outline" class="w-full" @click="isBillingBreakdownOpen = false">
          Tutup Rincian
        </BaseButton>
      </template>
    </BaseModal>

    <!-- ========================================================================= -->
    <!-- MODAL 4: BUAT INFORMED CONSENT DIGITAL (FASE 1)                           -->
    <!-- ========================================================================= -->
    <BaseModal
      v-model="isConsentModalOpen"
      title="Penerbitan Informed Consent Digital (Persetujuan Tindakan Medis)"
      max-width="3xl"
    >
      <div class="space-y-4">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <BaseSelect
            v-model="consentForm.consent_type"
            label="Jenis Formulir Persetujuan / Penolakan *"
            :options="consentTypeOptions"
          />

          <BaseInput
            v-model="consentForm.action_name"
            label="Nama Tindakan / Prosedur Medis *"
            placeholder="Contoh: Ekstirpasi Kista Ateroma / Jahit Luka / Bedah Minor"
          />
        </div>

        <div>
          <label class="text-xs font-semibold text-slate-700 block mb-1">
            Penjelasan Indikasi, Prosedur & Risiko Komplikasi Klinis *
          </label>
          <textarea
            v-model="consentForm.risks_explained"
            rows="3"
            class="w-full p-2.5 text-xs border border-slate-300 rounded-xl focus:ring-2 focus:ring-sky-500 bg-white"
            placeholder="Tuliskan penjelasan risiko, komplikasi, dan tata cara tindakan yang telah diinformasikan ke pasien..."
          ></textarea>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4 bg-slate-50 p-3 rounded-2xl border border-slate-200">
          <BaseInput
            v-model="consentForm.signer_name"
            label="Nama Lengkap Penandatangan *"
            placeholder="Nama Pasien atau Keluarga Penanggung Jawab"
          />

          <BaseSelect
            v-model="consentForm.signer_relation"
            label="Hubungan dengan Pasien *"
            :options="relationOptions"
          />
        </div>

        <!-- HTML5 Canvas Signature Pad -->
        <SignaturePad
          ref="signaturePadRef"
          label="Goreskan Tanda Tangan Pasien / Wali Penanggung Jawab *"
          @update:signature="val => consentForm.patient_signature_data = val"
        />
      </div>

      <template #footer>
        <div class="flex items-center justify-end space-x-3 w-full">
          <BaseButton variant="outline" @click="isConsentModalOpen = false">
            Batal
          </BaseButton>
          <BaseButton variant="primary" :loading="isSaving" @click="saveInformedConsent">
            Simpan & Terbitkan Informed Consent
          </BaseButton>
        </div>
      </template>
    </BaseModal>

    <!-- ========================================================================= -->
    <!-- MODAL 5: CETAK LEMBAR INFORMED CONSENT RESMI BER-QR                       -->
    <!-- ========================================================================= -->
    <BaseModal
      v-model="isPrintConsentModalOpen"
      title="Lembar Informed Consent Resmi (Dokumen Medikolegal)"
      max-width="3xl"
    >
      <div v-if="selectedConsent" id="printable-consent" class="p-6 bg-white text-slate-900 space-y-4 border border-slate-200 rounded-2xl">
        <!-- KOP SURAT KLINIK -->
        <div class="border-b-2 border-slate-900 pb-3 text-center">
          <h2 class="text-lg font-black uppercase tracking-wider text-sky-700">RUMAH SAKIT & KLINIK UTAMA SIMRS-ASOY</h2>
          <p class="text-xs text-slate-500">Jl. Sehat Bersama No. 12, Jakarta Selatan • Telp: (021) 7890-1234 • Izin Faskes: 445/RS-ASOY/2026</p>
          <h3 class="text-sm font-bold uppercase underline mt-2 tracking-wide text-slate-800">
            LEMBAR {{ selectedConsent.consent_type }}
          </h3>
          <span class="text-[11px] text-slate-500 font-mono">Kode Validasi: {{ selectedConsent.qr_code_token }}</span>
        </div>

        <!-- IDENTITAS PASIEN -->
        <div class="grid grid-cols-2 gap-2 text-xs bg-slate-50 p-3 rounded-xl border border-slate-200">
          <div><span class="text-slate-500">No. Rekam Medis:</span> <strong class="font-mono text-sky-700">{{ selectedConsent.no_rm }}</strong></div>
          <div><span class="text-slate-500">No. Rawat Kunjungan:</span> <strong class="font-mono">{{ selectedConsent.no_rawat }}</strong></div>
          <div><span class="text-slate-500">Nama Pasien:</span> <strong>{{ selectedConsent.patient_name }}</strong></div>
          <div><span class="text-slate-500">Dokter Penanggung Jawab:</span> <strong>{{ selectedConsent.doctor_name || 'dr. Spesialis' }}</strong></div>
        </div>

        <!-- ISI PERNYATAAN PERSETUJUAN -->
        <div class="space-y-3 text-xs text-slate-700 leading-relaxed">
          <p>
            Yang bertanda tangan di bawah ini, <strong>{{ selectedConsent.signer_name }}</strong>, bertindak sebagai <strong>{{ selectedConsent.signer_relation }}</strong> dari pasien di atas, menyatakan telah menerima penjelasan lengkap mengenai:
          </p>
          <div class="bg-sky-50 border border-sky-100 p-3 rounded-xl">
            <div class="font-bold text-slate-800">Tindakan Medis yang Diberikan:</div>
            <div class="text-sky-800 font-semibold text-sm">{{ selectedConsent.action_name }}</div>
            <div class="mt-2 font-bold text-slate-800">Penjelasan Risiko & Komplikasi:</div>
            <div class="text-slate-600 italic">{{ selectedConsent.risks_explained }}</div>
          </div>
          <p>
            Dengan ini menyatakan <strong>SETUJU / MEMAHAMI SEPENUHNYA</strong> untuk dilakukan tindakan medis tersebut di atas demi keselamatan dan pemulihan kesehatan pasien.
          </p>
        </div>

        <!-- TANDA TANGAN PASIEN & DOKTER + QR CODE -->
        <div class="grid grid-cols-3 gap-4 pt-4 border-t border-slate-200 text-center text-xs">
          <div>
            <div class="text-slate-500 mb-1">Pasien / Wali Penanggung Jawab</div>
            <div class="h-20 flex items-center justify-center border border-dashed border-slate-200 rounded-xl bg-slate-50/50 p-1">
              <img
                v-if="selectedConsent.patient_signature_data"
                :src="selectedConsent.patient_signature_data"
                alt="Tanda Tangan Pasien"
                class="max-h-16 max-w-full object-contain"
              />
              <span v-else class="text-[10px] text-slate-400 italic">TTE Tercatat</span>
            </div>
            <div class="font-bold text-slate-800 mt-1">{{ selectedConsent.signer_name }}</div>
          </div>

          <div class="flex flex-col items-center justify-center p-2 border border-slate-200 rounded-xl bg-slate-50">
            <div class="text-[10px] font-bold text-slate-500 uppercase mb-1">Scan QR Verifikasi Resmi</div>
            <div class="w-16 h-16 bg-white border border-slate-300 rounded-lg flex items-center justify-center p-1 shadow-xs">
              <div class="text-[8px] font-mono text-slate-400 text-center leading-tight">
                [QR CODE]<br>{{ selectedConsent.qr_code_token.substring(0, 12) }}
              </div>
            </div>
            <div class="text-[9px] text-emerald-600 font-bold mt-1">SAH TTE SIMRS</div>
          </div>

          <div>
            <div class="text-slate-500 mb-1">Dokter Penanggung Jawab</div>
            <div class="h-20 flex items-center justify-center border border-dashed border-slate-200 rounded-xl bg-slate-50/50 p-1">
              <div class="text-xs font-serif italic text-sky-800">Signed Digitally by E-System</div>
            </div>
            <div class="font-bold text-slate-800 mt-1">{{ selectedConsent.doctor_name || 'dr. Pemeriksa' }}</div>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="flex items-center justify-between w-full">
          <span class="text-xs text-slate-500">Dokumen memenuhi standar rekam medis Permenkes No. 24/2022</span>
          <div class="flex items-center space-x-3">
            <BaseButton variant="outline" @click="isPrintConsentModalOpen = false">
              Tutup
            </BaseButton>
            <BaseButton variant="primary" @click="printConsentDoc">
              <Printer class="w-4 h-4 mr-1.5" />
              Cetak Dokumen
            </BaseButton>
          </div>
        </div>
      </template>
    </BaseModal>
  </div>
</template>
