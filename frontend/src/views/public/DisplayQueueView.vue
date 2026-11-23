<template>
  <div class="min-h-screen bg-slate-950 text-white flex flex-col justify-between overflow-hidden select-none font-sans">
    <!-- 1. TOP HEADER TV -->
    <header class="bg-slate-900/90 border-b border-slate-800 px-6 py-4 flex items-center justify-between backdrop-blur-md shadow-lg">
      <div class="flex items-center space-x-4">
        <div class="w-12 h-12 rounded-2xl bg-gradient-to-tr from-sky-600 to-teal-400 flex items-center justify-center shadow-lg shadow-sky-500/20">
          <svg class="w-7 h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 4v16m8-8H4" />
          </svg>
        </div>
        <div>
          <h1 class="text-xl md:text-2xl font-black tracking-wide bg-gradient-to-r from-sky-400 via-teal-300 to-white bg-clip-text text-transparent">
            RUMAH SAKIT & KLINIK SIMRS-ASOY
          </h1>
          <p class="text-xs md:text-sm text-slate-400 font-medium">Layar Tampilan Antrean Ruang Tunggu Pasien Rawat Jalan</p>
        </div>
      </div>

      <!-- Jam & Audio Button -->
      <div class="flex items-center space-x-6">
        <button
          @click="toggleSound"
          :class="[
            'px-4 py-2 rounded-xl text-xs font-semibold flex items-center space-x-2 transition-all shadow-md',
            soundEnabled ? 'bg-emerald-600/30 text-emerald-300 border border-emerald-500/40 hover:bg-emerald-600/40' : 'bg-rose-600/30 text-rose-300 border border-rose-500/40 hover:bg-rose-600/40'
          ]"
        >
          <span>{{ soundEnabled ? '🔊 Suara Aktif' : '🔇 Klik Aktifkan Suara' }}</span>
        </button>

        <div class="text-right">
          <div class="text-2xl md:text-3xl font-mono font-black text-sky-400 tracking-wider">{{ currentTime }}</div>
          <div class="text-xs text-slate-400 font-medium">{{ currentDate }}</div>
        </div>
      </div>
    </header>

    <!-- 2. MAIN CALLING HERO & POLI GRID -->
    <main class="flex-1 p-6 grid grid-cols-1 lg:grid-cols-12 gap-6 overflow-hidden">
      <!-- HERO: PEMANGGILAN UTAMA (Left 5 Cols) -->
      <div class="lg:col-span-5 flex flex-col justify-between bg-gradient-to-b from-slate-900 to-slate-900/90 border-2 border-sky-500/50 rounded-3xl p-6 shadow-2xl shadow-sky-500/10 relative overflow-hidden">
        <div class="absolute -right-16 -top-16 w-64 h-64 bg-sky-500/10 rounded-full blur-3xl pointer-events-none"></div>

        <div>
          <div class="flex items-center justify-between mb-4">
            <span class="inline-flex items-center px-4 py-1.5 rounded-full text-xs font-bold bg-sky-500/20 text-sky-300 border border-sky-400/30 animate-pulse">
              <span class="w-2.5 h-2.5 rounded-full bg-sky-400 mr-2 shadow-lg shadow-sky-400"></span>
              PANGGILAN AKTIF SEKARANG
            </span>
            <span class="text-xs font-semibold text-slate-400 font-mono">LOKET / RUANGAN</span>
          </div>

          <div class="text-center py-6">
            <div class="text-xs uppercase tracking-widest text-slate-400 font-bold mb-2">NOMOR ANTREAN</div>
            <div class="font-mono text-7xl md:text-8xl lg:text-9xl font-black text-transparent bg-clip-text bg-gradient-to-b from-white via-sky-200 to-sky-400 drop-shadow-xl tracking-tight">
              {{ activeCall ? activeCall.no_antrian : '---' }}
            </div>

            <div class="mt-6 inline-block bg-slate-800/80 border border-slate-700/60 rounded-2xl px-6 py-3 max-w-full">
              <div class="text-xs uppercase tracking-wider text-slate-400">Nama Pasien</div>
              <div class="text-xl md:text-2xl font-bold text-white truncate max-w-md">
                {{ activeCall ? activeCall.patient_name : 'Menunggu Panggilan Pasien...' }}
              </div>
            </div>
          </div>
        </div>

        <div class="bg-slate-950/60 border border-slate-800 rounded-2xl p-4 flex items-center justify-between mt-4">
          <div>
            <div class="text-xs text-slate-400 uppercase tracking-wide">Poli / Unit Layanan</div>
            <div class="text-base font-bold text-teal-300">{{ activeCall ? activeCall.poli_name : 'Semua Poli' }}</div>
          </div>
          <div class="text-right">
            <div class="text-xs text-slate-400 uppercase tracking-wide">Dokter Pemeriksa</div>
            <div class="text-base font-bold text-white">{{ activeCall ? (activeCall.doctor_name || 'Dokter Jaga') : 'dr. Spesialis' }}</div>
          </div>
        </div>
      </div>

      <!-- GRID POLI CARDS (Right 7 Cols) -->
      <div class="lg:col-span-7 grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4 auto-rows-fr">
        <div
          v-for="unit in poliUnits"
          :key="unit.name"
          class="bg-slate-900/80 border border-slate-800 hover:border-slate-700 rounded-2xl p-4 flex flex-col justify-between transition-all shadow-md"
        >
          <!-- Card Header -->
          <div class="flex items-center justify-between border-b border-slate-800 pb-2 mb-2">
            <span class="text-xs font-bold text-sky-400 tracking-wide truncate max-w-[150px]">{{ unit.name }}</span>
            <span class="text-[10px] font-mono px-2 py-0.5 rounded-full bg-slate-800 text-slate-300 font-semibold">
              {{ unit.waitingCount }} antre
            </span>
          </div>

          <!-- Serving Number -->
          <div class="text-center my-auto py-2">
            <div class="text-[11px] text-slate-400 uppercase font-semibold">Sedang Dilayani</div>
            <div class="text-3xl md:text-4xl font-mono font-black text-emerald-400 tracking-tight">
              {{ unit.serving || '--' }}
            </div>
          </div>

          <!-- Doctor & Next Queue -->
          <div class="bg-slate-950/40 rounded-xl p-2 text-xs border border-slate-800/60 flex items-center justify-between">
            <span class="text-slate-400 truncate text-[11px]">{{ unit.doctor }}</span>
            <span class="font-mono text-amber-400 font-bold text-[11px]">Next: {{ unit.nextQueue || '-' }}</span>
          </div>
        </div>
      </div>
    </main>

    <!-- 3. RUNNING TICKER & FOOTER -->
    <footer class="bg-slate-900 border-t border-slate-800 py-2.5 px-6 flex items-center space-x-4">
      <div class="flex items-center space-x-2 text-xs font-bold text-sky-400 uppercase shrink-0 bg-slate-800 px-3 py-1 rounded-lg">
        <span class="w-2 h-2 rounded-full bg-sky-400 animate-ping"></span>
        <span>INFO RS</span>
      </div>
      <div class="overflow-hidden flex-1 relative whitespace-nowrap">
        <div class="inline-block animate-marquee text-xs md:text-sm text-slate-300 font-medium">
          {{ marqueeText }}
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { api as client } from '@/api/client'

interface ActiveCall {
  no_antrian: string
  patient_name: string
  poli_name: string
  doctor_name: string
}

interface PoliStatus {
  name: string
  doctor: string
  serving: string
  nextQueue: string
  waitingCount: number
}

const currentTime = ref('')
const currentDate = ref('')
const soundEnabled = ref(false)
const activeCall = ref<ActiveCall | null>(null)
let lastCalledNumber = ''

const marqueeText = ref(
  'Selamat Datang di RS SIMRS-ASOY • Jam Layanan Poliklinik Rawat Jalan: 08.00 - 21.00 WIB • Harap Menyiapkan KTP / Kartu BPJS / Kartu Asuransi Saat Dipanggil • Informasi Pendaftaran Online & Layanan Call Center: (021) 7890-1234 • Menerima Pembayaran Tunai, QRIS, Debit, Kredit & Split Payment'
)

const poliUnits = ref<PoliStatus[]>([
  { name: 'Poli Umum', doctor: 'dr. Siti Rahma', serving: 'A-001', nextQueue: 'A-002', waitingCount: 3 },
  { name: 'Poli Anak & Tumbuh Kembang', doctor: 'dr. Aris Munandar, Sp.A', serving: 'B-001', nextQueue: 'B-002', waitingCount: 2 },
  { name: 'Poli Gigi & Mulut', doctor: 'drg. Maya Indah', serving: 'C-001', nextQueue: 'C-002', waitingCount: 1 },
  { name: 'Poli Fisioterapi', doctor: 'Ns. Dian Anggraini', serving: 'D-001', nextQueue: '-', waitingCount: 0 },
  { name: 'Instalasi Farmasi Depo', doctor: 'Apt. Rizky Pratama', serving: 'F-003', nextQueue: 'F-004', waitingCount: 4 },
  { name: 'Kasir & Billing', doctor: 'Kasir Utama', serving: 'K-002', nextQueue: 'K-003', waitingCount: 2 },
])

let clockTimer: any = null
let queueTimer: any = null

function updateClock() {
  const now = new Date()
  currentTime.value = now.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' }) + ' WIB'
  currentDate.value = now.toLocaleDateString('id-ID', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })
}

function toggleSound() {
  soundEnabled.value = !soundEnabled.value
  if (soundEnabled.value) {
    playHospitalChime()
  }
}

function playHospitalChime() {
  try {
    const AudioCtx = window.AudioContext || (window as any).webkitAudioContext
    if (!AudioCtx) return
    const ctx = new AudioCtx()
    const now = ctx.currentTime

    const osc1 = ctx.createOscillator()
    const osc2 = ctx.createOscillator()
    const gainNode = ctx.createGain()

    osc1.type = 'sine'
    osc1.frequency.setValueAtTime(587.33, now) // D5
    osc1.frequency.setValueAtTime(880, now + 0.3) // A5

    gainNode.gain.setValueAtTime(0.3, now)
    gainNode.gain.exponentialRampToValueAtTime(0.001, now + 0.9)

    osc1.connect(gainNode)
    gainNode.connect(ctx.destination)

    osc1.start(now)
    osc1.stop(now + 0.9)
  } catch (e) {
    console.error('Audio chime error:', e)
  }
}

function announceQueue(call: ActiveCall) {
  if (!soundEnabled.value || !('speechSynthesis' in window)) return
  window.speechSynthesis.cancel()
  playHospitalChime()

  setTimeout(() => {
    const spelledNumber = call.no_antrian.replace('-', ' ')
    const text = `Nomor antrean, ${spelledNumber}, ${call.patient_name}, silakan menuju ke ${call.poli_name}`
    const utterance = new SpeechSynthesisUtterance(text)
    utterance.lang = 'id-ID'
    utterance.rate = 0.88
    utterance.pitch = 1.05
    window.speechSynthesis.speak(utterance)
  }, 750)
}

async function fetchQueueData() {
  try {
    const res = await client.get<any>('/public/queue/display?limit=10')
    const list = Array.isArray(res.data) ? res.data : (res.data?.data || [])

    if (list.length > 0) {
      // Find active serving or DIPERIKSA_DOKTER or KASIR
      const active = list.find((v: any) => v.status === 'DIPERIKSA_DOKTER' || v.status === 'FARMASI' || v.status === 'KASIR') || list[0]
      if (active) {
        const newCall: ActiveCall = {
          no_antrian: active.no_antrian || 'A-001',
          patient_name: active.patient_name || 'Pasien Rawat Jalan',
          poli_name: active.poli_name || 'Poliklinik Rawat Jalan',
          doctor_name: active.doctor_name || 'Dokter Jaga',
        }

        if (newCall.no_antrian !== lastCalledNumber) {
          activeCall.value = newCall
          lastCalledNumber = newCall.no_antrian
          announceQueue(newCall)
        }
      }

      // Update unit cards based on actual poli list
      const groups: Record<string, any[]> = {}
      list.forEach((v: any) => {
        const pName = v.poli_name || 'Poli Umum'
        if (!groups[pName]) groups[pName] = []
        groups[pName].push(v)
      })

      poliUnits.value = Object.keys(groups).slice(0, 6).map((pName) => {
        const items = groups[pName]
        const servingItem = items.find((it: any) => it.status === 'DIPERIKSA_DOKTER' || it.status === 'FARMASI') || items[0]
        const nextItem = items.find((it: any) => it.id !== servingItem?.id)
        return {
          name: pName,
          doctor: servingItem?.doctor_name || 'dr. Jaga',
          serving: servingItem?.no_antrian || '-',
          nextQueue: nextItem?.no_antrian || '-',
          waitingCount: items.filter((it: any) => it.status.includes('MENUNGGU')).length,
        }
      })
    }
  } catch (err) {
    console.warn('Display queue sync warning:', err)
  }
}

onMounted(() => {
  updateClock()
  clockTimer = setInterval(updateClock, 1000)
  fetchQueueData()
  queueTimer = setInterval(fetchQueueData, 4000)
})

onUnmounted(() => {
  if (clockTimer) clearInterval(clockTimer)
  if (queueTimer) clearInterval(queueTimer)
  if ('speechSynthesis' in window) window.speechSynthesis.cancel()
})
</script>

<style scoped>
@keyframes marquee {
  0% { transform: translateX(100%); }
  100% { transform: translateX(-100%); }
}

.animate-marquee {
  display: inline-block;
  animation: marquee 35s linear infinite;
}
</style>
