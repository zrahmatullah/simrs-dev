<template>
  <div class="min-h-screen bg-[#F3F6F9] text-slate-800 flex font-sans">
    <!-- ==================== SIDEBAR ==================== -->
    <aside class="w-64 bg-[#0C2B38] text-slate-300 flex flex-col shrink-0 min-h-screen select-none">
      <!-- Logo SIMRS -->
      <div class="h-16 flex items-center gap-3 px-6 border-b border-white/10">
        <div class="w-8 h-8 rounded-lg bg-teal-500/20 text-teal-400 border border-teal-500/30 flex items-center justify-center font-bold text-lg">
          <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 text-teal-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 6v12M6 12h12"/>
          </svg>
        </div>
        <span class="font-bold text-white text-lg tracking-wider">SIMRS</span>
      </div>

      <!-- Navigation Menu -->
      <nav class="flex-1 py-4 px-3 space-y-1 overflow-y-auto text-sm">
        <button 
          v-for="item in menuItems" 
          :key="item.id"
          @click="activeMenu = item.id"
          :class="[
            'w-full flex items-center gap-3 px-3.5 py-2.5 rounded-xl transition text-left text-xs font-medium',
            activeMenu === item.id 
              ? 'bg-[#17475B] text-white font-semibold shadow-sm border border-cyan-700/40' 
              : 'text-slate-300 hover:text-white hover:bg-white/5'
          ]"
        >
          <span class="w-2 h-2 rounded-full shrink-0" :class="activeMenu === item.id ? 'bg-cyan-300' : 'bg-slate-500'"></span>
          <span>{{ item.label }}</span>
        </button>
      </nav>
    </aside>

    <!-- ==================== MAIN CONTENT ==================== -->
    <div class="flex-1 flex flex-col min-w-0">
      
      <!-- Topbar Header -->
      <header class="h-16 bg-white border-b border-slate-200 px-6 flex items-center justify-between sticky top-0 z-20">
        <h1 class="text-lg font-bold text-slate-800 tracking-tight">Rekam Medis Elektronik (RME)</h1>

        <div class="flex items-center gap-5">
          <!-- Help Icon -->
          <button class="text-slate-500 hover:text-slate-700 transition" title="Bantuan">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </button>

          <!-- Notification Bell -->
          <button class="text-slate-500 hover:text-slate-700 transition relative" title="Notifikasi">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
            </svg>
            <span class="absolute -top-1 -right-1 bg-red-500 text-white text-[10px] font-bold w-4 h-4 rounded-full flex items-center justify-center">1</span>
          </button>

          <!-- User Doctor Profile -->
          <div class="flex items-center gap-3 pl-2 border-l border-slate-200">
            <img :src="currentUser.avatar" alt="Avatar" class="w-9 h-9 rounded-full object-cover border border-slate-300">
            <div class="text-left hidden sm:block">
              <p class="text-xs font-bold text-slate-800 leading-tight">{{ currentUser.name }}</p>
              <p class="text-[11px] text-slate-500">{{ currentUser.role }}</p>
            </div>
          </div>
        </div>
      </header>

      <!-- Workspace Body -->
      <main class="flex-1 p-6 space-y-5 overflow-y-auto max-w-[1600px] w-full mx-auto">

        <!-- ==================== 1. PATIENT HEADER BANNER ==================== -->
        <section class="bg-[#EAF4FA] border border-[#CFE4F1] rounded-2xl p-4 shadow-sm flex flex-col lg:flex-row items-start lg:items-center justify-between gap-4">
          <!-- Left: Identity -->
          <div class="flex items-center gap-4 flex-wrap sm:flex-nowrap">
            <img :src="patient.avatarUrl" alt="Foto Pasien" class="w-14 h-14 rounded-xl object-cover border border-sky-300 shrink-0">
            
            <div class="grid grid-cols-2 sm:grid-cols-5 gap-x-6 gap-y-1 text-xs">
              <div>
                <p class="text-slate-500 font-medium text-[11px]">Nama Pasien</p>
                <p class="font-bold text-slate-800 text-sm">{{ patient.name }}</p>
              </div>
              <div>
                <p class="text-slate-500 font-medium text-[11px]">No. RM</p>
                <p class="font-bold text-slate-800 font-mono text-sm">{{ patient.norm }}</p>
              </div>
              <div>
                <p class="text-slate-500 font-medium text-[11px]">Umur/JK</p>
                <p class="font-semibold text-slate-700">{{ patient.age }} / {{ patient.gender }}</p>
              </div>
              <div>
                <p class="text-slate-500 font-medium text-[11px]">Jaminan</p>
                <p class="font-semibold text-slate-700">{{ patient.insurance }}</p>
              </div>
              <div class="col-span-2 sm:col-span-1">
                <p class="text-slate-500 font-medium text-[11px]">Alergi</p>
                <p class="font-bold text-red-500">{{ patient.allergies.join(', ') }}</p>
              </div>
            </div>
          </div>

          <!-- Right: Action Buttons & Status -->
          <div class="flex items-center gap-2.5 shrink-0 self-end lg:self-center">
            <button @click="$emit('print-wristband')" class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-sky-300 bg-white hover:bg-sky-50 text-slate-700 text-xs font-semibold shadow-xs transition">
              <svg class="w-3.5 h-3.5 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
              </svg>
              <span>Cetak Gelang</span>
            </button>

            <button @click="$emit('view-history')" class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-sky-300 bg-white hover:bg-sky-50 text-slate-700 text-xs font-semibold shadow-xs transition">
              <svg class="w-3.5 h-3.5 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span>Riwayat Kunjungan</span>
            </button>

            <div class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-emerald-50 border border-emerald-200 text-emerald-700 text-xs font-bold">
              <span class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
              <span>Stabil</span>
            </div>
          </div>
        </section>

        <!-- ==================== 2. TWO-COLUMN GRID ==================== -->
        <div class="grid grid-cols-1 lg:grid-cols-12 gap-5 items-start">
          
          <!-- LEFT COLUMN: FORM ASESMEN SOAP (7 COLS) -->
          <section class="lg:col-span-7 bg-white border border-slate-200/90 rounded-2xl p-5 shadow-xs space-y-4">
            <h2 class="text-base font-bold text-slate-800">Form Asesmen SOAP (Pemeriksaan Klinis)</h2>

            <!-- SUBJEKTIF -->
            <div class="space-y-1.5">
              <label class="block text-xs font-bold uppercase tracking-wider text-slate-800">SUBJEKTIF (S)</label>
              <textarea 
                v-model="soap.subjective"
                rows="3" 
                class="w-full text-xs text-slate-700 p-3 rounded-xl border border-slate-300 focus:border-teal-500 focus:ring-1 focus:ring-teal-500 outline-none leading-relaxed transition"
                placeholder="Keluhan utama, riwayat penyakit saat ini, riwayat alergi..."
              ></textarea>
            </div>

            <!-- OBJEKTIF -->
            <div class="space-y-1.5">
              <label class="block text-xs font-bold uppercase tracking-wider text-slate-800">OBJEKTIF (O)</label>
              <textarea 
                v-model="soap.objective"
                rows="3" 
                class="w-full text-xs text-slate-700 p-3 rounded-xl border border-slate-300 focus:border-teal-500 focus:ring-1 focus:ring-teal-500 outline-none leading-relaxed transition"
                placeholder="Hasil pemeriksaan fisik objektif..."
              ></textarea>
            </div>

            <!-- ASESMEN -->
            <div class="space-y-1.5">
              <div class="flex items-center justify-between">
                <label class="text-xs font-bold text-slate-800">Diagnosa utama</label>
                <button type="button" class="text-[11px] font-bold text-teal-600 bg-teal-50 px-2 py-0.5 rounded border border-teal-200 hover:bg-teal-100">ICD-10</button>
              </div>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div class="relative">
                  <input 
                    v-model="soap.primaryDiagnosis"
                    type="text" 
                    placeholder="Cari ICD-10..." 
                    class="w-full text-xs font-medium text-slate-800 pl-3 pr-8 py-2.5 rounded-xl border border-slate-300 bg-slate-50/50 focus:bg-white focus:border-teal-500 outline-none transition"
                  />
                  <svg class="w-4 h-4 text-slate-400 absolute right-2.5 top-2.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                  </svg>
                </div>
                <div>
                  <input 
                    v-model="soap.secondaryDiagnosis"
                    type="text" 
                    placeholder="Diagnosis sekunder..." 
                    class="w-full text-xs font-medium text-slate-800 px-3 py-2.5 rounded-xl border border-slate-300 bg-slate-50/50 focus:bg-white focus:border-teal-500 outline-none transition"
                  />
                </div>
              </div>
            </div>

            <!-- PLANNING -->
            <div class="space-y-1.5">
              <label class="block text-xs font-bold uppercase tracking-wider text-slate-800">PLANNING (P)</label>
              <textarea 
                v-model="soap.planning"
                rows="3" 
                class="w-full text-xs text-slate-700 p-3 rounded-xl border border-slate-300 focus:border-teal-500 focus:ring-1 focus:ring-teal-500 outline-none leading-relaxed transition"
                placeholder="Rencana terapi, edukasi, tindak lanjut..."
              ></textarea>
            </div>

            <!-- Footer Action -->
            <div class="pt-2 flex justify-end gap-2 border-t border-slate-100">
              <button @click="$emit('save-draft', soap)" class="px-4 py-2 rounded-xl border border-slate-300 text-xs font-semibold text-slate-600 hover:bg-slate-50 transition">
                Draft Sementara
              </button>
              <button @click="$emit('finalize', soap)" class="px-4 py-2 rounded-xl bg-teal-600 hover:bg-teal-700 text-white text-xs font-bold shadow-xs transition">
                Simpan & Finalisasi Rekam Medis
              </button>
            </div>
          </section>

          <!-- RIGHT COLUMN: VITALS & INTEGRATED ORDERS (5 COLS) -->
          <div class="lg:col-span-5 space-y-5">

            <!-- 1. KARTU RINGKASAN TANDA-TANDA VITAL -->
            <section class="bg-white border border-slate-200/90 rounded-2xl p-4 shadow-xs">
              <h3 class="text-sm font-bold text-slate-800 mb-3">Kartu Ringkasan Tanda-Tanda Vital</h3>
              <div class="grid grid-cols-5 gap-2">
                <div class="bg-slate-50 border border-slate-200/80 rounded-xl p-2.5 text-center flex flex-col justify-between">
                  <span class="text-[11px] font-bold text-slate-500">TD</span>
                  <span class="text-base sm:text-lg font-bold text-slate-800 my-0.5">{{ vitals.bp }}</span>
                  <span class="text-[10px] text-slate-400 font-medium">mmHg</span>
                </div>
                <div class="bg-slate-50 border border-slate-200/80 rounded-xl p-2.5 text-center flex flex-col justify-between">
                  <span class="text-[11px] font-bold text-slate-500">HR</span>
                  <span class="text-base sm:text-lg font-bold text-slate-800 my-0.5">{{ vitals.hr }}</span>
                  <span class="text-[10px] text-slate-400 font-medium">bpm</span>
                </div>
                <div class="bg-slate-50 border border-slate-200/80 rounded-xl p-2.5 text-center flex flex-col justify-between">
                  <span class="text-[11px] font-bold text-slate-500">Temp</span>
                  <span class="text-base sm:text-lg font-bold text-slate-800 my-0.5">{{ vitals.temp }}</span>
                  <span class="text-[10px] text-slate-400 font-medium">°C</span>
                </div>
                <div class="bg-slate-50 border border-slate-200/80 rounded-xl p-2.5 text-center flex flex-col justify-between">
                  <span class="text-[11px] font-bold text-slate-500">RR</span>
                  <span class="text-base sm:text-lg font-bold text-slate-800 my-0.5">{{ vitals.rr }}</span>
                  <span class="text-[10px] text-slate-400 font-medium">x/m</span>
                </div>
                <div class="bg-slate-50 border border-slate-200/80 rounded-xl p-2.5 text-center flex flex-col justify-between">
                  <span class="text-[11px] font-bold text-slate-500">SpO2</span>
                  <span class="text-base sm:text-lg font-bold text-slate-800 my-0.5">{{ vitals.spo2 }}%</span>
                  <span class="text-[10px] text-slate-400 font-medium">Oksigen</span>
                </div>
              </div>
            </section>

            <!-- 2. TAB TERINTEGRASI E-RESEP & ORDER LAB/RAD -->
            <section class="bg-white border border-slate-200/90 rounded-2xl p-4 shadow-xs">
              <h3 class="text-sm font-bold text-slate-800 mb-2">Tab Terintegrasi 'E-Resep' & "Order Lab/Rad"</h3>

              <!-- Tabs Header -->
              <div class="flex border-b border-slate-200 mb-4">
                <button 
                  @click="activeOrderTab = 'resep'"
                  :class="[
                    'flex-1 py-2 text-xs transition text-center',
                    activeOrderTab === 'resep' 
                      ? 'font-bold text-teal-700 border-b-2 border-teal-600' 
                      : 'font-semibold text-slate-500 hover:text-slate-700 border-b-2 border-transparent'
                  ]"
                >
                  E-Resep
                </button>
                <button 
                  @click="activeOrderTab = 'lab'"
                  :class="[
                    'flex-1 py-2 text-xs transition text-center',
                    activeOrderTab === 'lab' 
                      ? 'font-bold text-teal-700 border-b-2 border-teal-600' 
                      : 'font-semibold text-slate-500 hover:text-slate-700 border-b-2 border-transparent'
                  ]"
                >
                  Lab & Radiologi
                </button>
              </div>

              <!-- Content E-Resep -->
              <div v-show="activeOrderTab === 'resep'" class="space-y-3">
                <div class="overflow-x-auto">
                  <table class="w-full text-left text-xs">
                    <thead>
                      <tr class="text-slate-500 border-b border-slate-100">
                        <th class="py-1.5 font-semibold">Obat</th>
                        <th class="py-1.5 font-semibold text-center w-14">Jumlah</th>
                        <th class="py-1.5 font-semibold text-right">Aturan Pakai</th>
                      </tr>
                    </thead>
                    <tbody class="divide-y divide-slate-100">
                      <tr v-for="(med, idx) in prescriptions" :key="idx">
                        <td class="py-2 font-medium text-slate-800">{{ med.name }}</td>
                        <td class="py-2 text-center">
                          <input type="number" v-model="med.qty" class="w-10 text-center border border-slate-300 rounded p-1 text-xs">
                        </td>
                        <td class="py-2 text-right">
                          <input type="text" v-model="med.signa" class="w-24 text-right border border-slate-300 rounded p-1 text-xs">
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>

                <button @click="addMedicine" class="w-full py-2 rounded-xl border border-dashed border-teal-500 text-teal-600 hover:bg-teal-50 text-xs font-semibold flex items-center justify-center gap-1.5 transition">
                  + Tambah Obat
                </button>
              </div>

              <!-- Content Lab & Radiologi -->
              <div v-show="activeOrderTab === 'lab'" class="space-y-3">
                <div class="overflow-x-auto">
                  <table class="w-full text-left text-xs">
                    <thead>
                      <tr class="text-slate-500 border-b border-slate-100">
                        <th class="py-1.5 font-semibold">Pemeriksaan</th>
                        <th class="py-1.5 font-semibold text-right w-16">Status</th>
                      </tr>
                    </thead>
                    <tbody class="divide-y divide-slate-100">
                      <tr v-for="(lab, idx) in labOrders" :key="idx" class="hover:bg-slate-50">
                        <td class="py-2 font-medium text-slate-800">{{ lab.name }}</td>
                        <td class="py-2 text-right">
                          <input type="checkbox" v-model="lab.checked" class="w-4 h-4 text-teal-600 rounded border-slate-300 focus:ring-teal-500 cursor-pointer">
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>

                <button @click="addLabOrder" class="w-full py-2 rounded-xl border border-dashed border-sky-500 text-sky-600 hover:bg-sky-50 text-xs font-semibold flex items-center justify-center gap-1.5 transition">
                  + Tambah Order Penunjang
                </button>
              </div>
            </section>

          </div>
        </div>

      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';

// Menu Items Definition
const activeMenu = ref('pasien_rme');
const menuItems = [
  { id: 'dashboard', label: 'Dashboard' },
  { id: 'jadwal', label: 'Jadwal' },
  { id: 'pasien_rme', label: 'Pasien (RME)' },
  { id: 'rawat_jalan', label: 'Rawat Jalan' },
  { id: 'rawat_inap', label: 'Rawat Inap' },
  { id: 'e_resep', label: 'E-Resep' },
  { id: 'laboratorium', label: 'Laboratorium' },
  { id: 'radiologi', label: 'Radiologi' },
  { id: 'tagihan', label: 'Tagihan' },
  { id: 'pengaturan', label: 'Pengaturan' },
];

const currentUser = ref({
  name: 'Azhar As Rahmatulloh',
  role: 'Dokter Umum',
  avatar: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=120'
});

const patient = ref({
  name: 'Azhar As Rahmatulloh',
  norm: '00-12-34-56',
  age: '26 Th',
  gender: 'Laki-laki',
  insurance: 'BPJS Kesehatan',
  allergies: ['Ibuprofen', 'Debu'],
  avatarUrl: 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&q=80&w=160'
});

const soap = ref({
  subjective: 'Keluhan utama: Azhar As Rahmatulloh, untuk kontrol berkala dan penyakit. Riwayat penyakit dahulu: Gastritis kronis. Riwayat alergi: Ibuprofen dan alergi debu. Pemeriksaan fisik subjektif pasien.',
  objective: 'Hasil pemeriksaan objektif (pasien):\nTensi: 120/80 mmHg, Nadi: 82 bpm, Suhu: 36.5 °C, Pernapasan, kesadaran: Compos Mentis.',
  primaryDiagnosis: 'A09.9 - Gastroenteritis',
  secondaryDiagnosis: 'I10 - Essential Hypertension',
  planning: 'Rencana terapi medikamentosa, edukasi pola diet seimbang, istirahat cukup, dan tindak lanjut evaluasi keluhan.'
});

const vitals = ref({
  bp: '120/80',
  hr: 82,
  temp: 36.5,
  rr: 18,
  spo2: 98
});

const activeOrderTab = ref<'resep' | 'lab'>('resep');

const prescriptions = ref([
  { name: 'Paracetamol 500mg', qty: 1, signa: '3x1 500mg' },
  { name: 'Amoxicillin 500mg', qty: 1, signa: '3x1 500mg' },
]);

const labOrders = ref([
  { name: 'Darah Lengkap', checked: true },
  { name: 'Rontgen Thorax', checked: true },
  { name: 'Darah Rutin / Lamsu', checked: false },
  { name: 'Rontgen Thorax AP', checked: false },
  { name: 'TSSA', checked: false },
  { name: 'Medial Darah', checked: false },
]);

const addMedicine = () => {
  prescriptions.value.push({ name: '', qty: 1, signa: '' });
};

const addLabOrder = () => {
  labOrders.value.push({ name: 'Pemeriksaan Baru', checked: false });
};
</script>
