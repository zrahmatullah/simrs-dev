<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '@/api/client'
import type { AncillaryOrder } from '@/types'
import { useUiStore } from '@/stores/ui'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import BaseModal from '@/components/common/BaseModal.vue'
import {
  FlaskConical,
  Activity,
  FileText,
  Search,
  RefreshCw,
  Printer,
  CheckCircle2,
  Clock,
  Eye,
  Radio,
  FileCheck
} from 'lucide-vue-next'

const uiStore = useUiStore()

const loading = ref(false)
const orders = ref<AncillaryOrder[]>([])
const searchQuery = ref('')
const selectedType = ref<string>('SEMUA')
const selectedStatus = ref<string>('SEMUA')

// Modal states
const showResultModal = ref(false)
const showPrintModal = ref(false)
const selectedOrder = ref<AncillaryOrder | null>(null)
const submitting = ref(false)

// Form Input Hasil
const formResult = ref({
  hasil_pemeriksaan: '',
  lampiran_url: '',
  status: 'SELESAI',
})

const typeOptions = [
  { label: 'Semua Layanan', value: 'SEMUA' },
  { label: 'Laboratorium (LAB)', value: 'LAB' },
  { label: 'Radiologi (RAD)', value: 'RAD' },
  { label: 'Fisioterapi', value: 'FISIOTERAPI' },
  { label: 'EKG / Diagnostik', value: 'EKG' },
]

const statusOptions = [
  { label: 'Semua Status', value: 'SEMUA' },
  { label: 'Menunggu', value: 'MENUNGGU' },
  { label: 'Dalam Proses', value: 'DALAM_PROSES' },
  { label: 'Selesai', value: 'SELESAI' },
]

const fetchOrders = async () => {
  loading.value = true
  try {
    const params: Record<string, any> = {
      page: 1,
      limit: 100,
    }
    if (selectedType.value !== 'SEMUA') params.jenis_order = selectedType.value
    if (selectedStatus.value !== 'SEMUA') params.status = selectedStatus.value
    if (searchQuery.value) params.search = searchQuery.value

    const res = await api.get<AncillaryOrder[]>('/ancillary/orders', params)
    orders.value = res.data || []
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal memuat daftar order penunjang', 'error')
  } finally {
    loading.value = false
  }
}

// Metrics
const countTotal = computed(() => orders.value.length)
const countMenunggu = computed(() => orders.value.filter(o => o.status === 'MENUNGGU').length)
const countProses = computed(() => orders.value.filter(o => o.status === 'DALAM_PROSES').length)
const countSelesai = computed(() => orders.value.filter(o => o.status === 'SELESAI').length)

const openInputResultModal = (order: AncillaryOrder) => {
  selectedOrder.value = order
  formResult.value = {
    hasil_pemeriksaan: order.hasil_pemeriksaan || '',
    lampiran_url: order.lampiran_url || '',
    status: 'SELESAI',
  }

  // If empty, prefill friendly template based on service type
  if (!formResult.value.hasil_pemeriksaan) {
    if (order.jenis_order === 'LAB') {
      formResult.value.hasil_pemeriksaan = `Pemeriksaan: ${order.nama_tindakan}\nHasil: NORMAL / DALAM BATAS NORMAL\nNilai Rujukan: Standar Lab SIMRS\nKesan / Interpretasi: Tidak ditemukan kelainan spesifik.`
    } else if (order.jenis_order === 'RAD') {
      formResult.value.hasil_pemeriksaan = `Pemeriksaan: ${order.nama_tindakan}\nCor / Pulmo: Cor dalam batas normal, pulmo tidak tampak infiltrate aktif.\nTulang & Sendi: Struktur skeletal intak, tidak tampak fraktur/dislokasi.\nKesan: Foto Thorax/Radiografi dalam batas normal.`
    } else if (order.jenis_order === 'EKG') {
      formResult.value.hasil_pemeriksaan = `Irama: Sinus Rhytm, HR: 78 bpm, Regular\nGelombang P-QRS-T: Normal\nKesan: Normal Sinus Rhythm (NSR).`
    } else {
      formResult.value.hasil_pemeriksaan = `Hasil tindakan ${order.nama_tindakan} telah dilaksanakan dengan baik. Respon pasien baik dan tidak ada efek samping akut.`
    }
  }

  showResultModal.value = true
}

const saveResult = async () => {
  if (!selectedOrder.value) return
  if (!formResult.value.hasil_pemeriksaan.trim()) {
    uiStore.showToast('Hasil pemeriksaan wajib diisi', 'warning')
    return
  }

  submitting.value = true
  try {
    await api.put(`/ancillary/orders/${selectedOrder.value.id}/result`, formResult.value)
    uiStore.showToast('Hasil pemeriksaan penunjang berhasil divalidasi dan disimpan', 'success')
    showResultModal.value = false
    await fetchOrders()
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal menyimpan hasil pemeriksaan', 'error')
  } finally {
    submitting.value = false
  }
}

const openPrintModal = (order: AncillaryOrder) => {
  selectedOrder.value = order
  showPrintModal.value = true
}

const printDirect = () => {
  window.print()
}

const formatDate = (isoString?: string) => {
  if (!isoString) return '-'
  return new Date(isoString).toLocaleString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const formatCurrency = (val: number) => {
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(val)
}

onMounted(() => {
  fetchOrders()
})
</script>

<template>
  <div class="space-y-4">
    <!-- Header Page -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-white p-4 rounded-xl border border-slate-200 shadow-sm">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-lg bg-sky-50 text-sky-600 flex items-center justify-center border border-sky-100">
          <FlaskConical class="w-5 h-5" />
        </div>
        <div>
          <h1 class="text-xl font-bold text-slate-800">Pelayanan Penunjang Medis</h1>
          <p class="text-xs text-slate-500">Laboratorium, Radiologi, Fisioterapi, EKG & Diagnostik Khusus</p>
        </div>
      </div>
      <div class="flex items-center gap-2">
        <BaseButton variant="secondary" size="sm" @click="fetchOrders">
          <RefreshCw class="w-4 h-4 mr-1.5" :class="{ 'animate-spin': loading }" />
          Segarkan Data
        </BaseButton>
      </div>
    </div>

    <!-- Metrics Cards -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
      <div class="bg-white p-3.5 rounded-xl border border-slate-200 shadow-sm flex items-center gap-3">
        <div class="w-10 h-10 rounded-lg bg-sky-50 text-sky-600 flex items-center justify-center shrink-0">
          <Activity class="w-5 h-5" />
        </div>
        <div>
          <div class="text-xs text-slate-500 font-medium">Total Permintaan</div>
          <div class="text-xl font-bold text-slate-800 tabular-nums">{{ countTotal }}</div>
        </div>
      </div>

      <div class="bg-white p-3.5 rounded-xl border border-slate-200 shadow-sm flex items-center gap-3">
        <div class="w-10 h-10 rounded-lg bg-amber-50 text-amber-600 flex items-center justify-center shrink-0">
          <Clock class="w-5 h-5" />
        </div>
        <div>
          <div class="text-xs text-slate-500 font-medium">Menunggu Tindakan</div>
          <div class="text-xl font-bold text-amber-600 tabular-nums">{{ countMenunggu }}</div>
        </div>
      </div>

      <div class="bg-white p-3.5 rounded-xl border border-slate-200 shadow-sm flex items-center gap-3">
        <div class="w-10 h-10 rounded-lg bg-blue-50 text-blue-600 flex items-center justify-center shrink-0">
          <Radio class="w-5 h-5" />
        </div>
        <div>
          <div class="text-xs text-slate-500 font-medium">Dalam Proses</div>
          <div class="text-xl font-bold text-blue-600 tabular-nums">{{ countProses }}</div>
        </div>
      </div>

      <div class="bg-white p-3.5 rounded-xl border border-slate-200 shadow-sm flex items-center gap-3">
        <div class="w-10 h-10 rounded-lg bg-emerald-50 text-emerald-600 flex items-center justify-center shrink-0">
          <CheckCircle2 class="w-5 h-5" />
        </div>
        <div>
          <div class="text-xs text-slate-500 font-medium">Hasil Selesai</div>
          <div class="text-xl font-bold text-emerald-600 tabular-nums">{{ countSelesai }}</div>
        </div>
      </div>
    </div>

    <!-- Filter & Search Bar -->
    <div class="bg-white p-4 rounded-xl border border-slate-200 shadow-sm">
      <div class="grid grid-cols-1 md:grid-cols-12 gap-3 items-center">
        <div class="md:col-span-5 relative">
          <Search class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Cari No. Order, No. RM, atau Nama Pasien..."
            class="w-full pl-9 pr-3 py-2 text-sm border border-slate-200 rounded-lg focus:outline-none focus:border-sky-500 focus:ring-1 focus:ring-sky-500 transition-colors"
            @keyup.enter="fetchOrders"
          />
        </div>

        <div class="md:col-span-4 flex items-center gap-2">
          <BaseSelect
            v-model="selectedType"
            :options="typeOptions"
            placeholder="Pilih Layanan"
            @change="fetchOrders"
          />
        </div>

        <div class="md:col-span-3 flex items-center gap-2">
          <BaseSelect
            v-model="selectedStatus"
            :options="statusOptions"
            placeholder="Pilih Status"
            @change="fetchOrders"
          />
        </div>
      </div>
    </div>

    <!-- Orders Table -->
    <div class="bg-white rounded-xl border border-slate-200 shadow-sm overflow-hidden">
      <div class="px-4 py-3 border-b border-slate-200 flex items-center justify-between bg-slate-50/50">
        <div class="flex items-center gap-2">
          <FileText class="w-4 h-4 text-sky-600" />
          <span class="text-sm font-semibold text-slate-700">Daftar Order Pemeriksaan Penunjang</span>
          <span class="text-xs bg-sky-100 text-sky-700 font-semibold px-2 py-0.5 rounded-full tabular-nums">{{ orders.length }}</span>
        </div>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs text-slate-600">
          <thead class="bg-slate-50 text-slate-700 font-semibold border-b border-slate-200 uppercase tracking-wider">
            <tr>
              <th class="py-3 px-4">No. Order / Waktu</th>
              <th class="py-3 px-4">Pasien & Penjamin</th>
              <th class="py-3 px-4">Layanan & Tindakan</th>
              <th class="py-3 px-4">Dokter Pengirim</th>
              <th class="py-3 px-4">Catatan Klinis</th>
              <th class="py-3 px-4 text-right">Tarif</th>
              <th class="py-3 px-4 text-center">Status</th>
              <th class="py-3 px-4 text-center">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="loading">
              <td colspan="8" class="text-center py-10 text-slate-400">
                <RefreshCw class="w-6 h-6 animate-spin mx-auto mb-2 text-sky-500" />
                Memuat data order pemeriksaan...
              </td>
            </tr>
            <tr v-else-if="orders.length === 0">
              <td colspan="8" class="text-center py-12 text-slate-400">
                <FlaskConical class="w-10 h-10 mx-auto mb-2 text-slate-300" />
                Tidak ada order pemeriksaan penunjang yang ditemukan.
              </td>
            </tr>
            <tr
              v-for="order in orders"
              :key="order.id"
              class="hover:bg-slate-50 transition-colors"
            >
              <td class="py-3 px-4">
                <div class="font-bold text-sky-700 font-mono">{{ order.nomor_order }}</div>
                <div class="text-[11px] text-slate-400 flex items-center gap-1 mt-0.5">
                  <Clock class="w-3 h-3" />
                  {{ formatDate(order.created_at) }}
                </div>
              </td>

              <td class="py-3 px-4">
                <div class="font-semibold text-slate-800">{{ order.patient_name || '-' }}</div>
                <div class="text-[11px] text-slate-500">
                  <span class="font-mono font-medium text-slate-600">RM: {{ order.no_rm || '-' }}</span>
                </div>
              </td>

              <td class="py-3 px-4">
                <div class="flex items-center gap-1.5 mb-1">
                  <span
                    class="px-2 py-0.5 rounded text-[10px] font-bold tracking-wider uppercase"
                    :class="{
                      'bg-purple-100 text-purple-700 border border-purple-200': order.jenis_order === 'LAB',
                      'bg-indigo-100 text-indigo-700 border border-indigo-200': order.jenis_order === 'RAD',
                      'bg-teal-100 text-teal-700 border border-teal-200': order.jenis_order === 'FISIOTERAPI',
                      'bg-rose-100 text-rose-700 border border-rose-200': order.jenis_order === 'EKG',
                    }"
                  >
                    {{ order.jenis_order }}
                  </span>
                </div>
                <div class="font-medium text-slate-800">{{ order.nama_tindakan }}</div>
              </td>

              <td class="py-3 px-4">
                <div class="text-slate-700 font-medium">{{ order.doctor_name || 'Dokter Pemeriksa' }}</div>
              </td>

              <td class="py-3 px-4 max-w-[200px]">
                <div class="truncate text-slate-500 text-[11px]" :title="order.catatan_klinis">
                  {{ order.catatan_klinis || '-' }}
                </div>
              </td>

              <td class="py-3 px-4 text-right font-medium text-slate-700 tabular-nums">
                {{ formatCurrency(order.tarif) }}
              </td>

              <td class="py-3 px-4 text-center">
                <StatusBadge :status="order.status" />
              </td>

              <td class="py-3 px-4 text-center">
                <div class="flex items-center justify-center gap-1">
                  <!-- Input / Update Result -->
                  <BaseButton
                    v-if="order.status !== 'SELESAI'"
                    variant="primary"
                    size="xs"
                    @click="openInputResultModal(order)"
                  >
                    <FileCheck class="w-3.5 h-3.5 mr-1" />
                    Input Hasil
                  </BaseButton>

                  <!-- View & Print Result -->
                  <template v-else>
                    <BaseButton
                      variant="soft"
                      size="xs"
                      @click="openInputResultModal(order)"
                      title="Edit / Lihat Hasil"
                    >
                      <Eye class="w-3.5 h-3.5 mr-1" />
                      Detail
                    </BaseButton>
                    <BaseButton
                      variant="secondary"
                      size="xs"
                      @click="openPrintModal(order)"
                      title="Cetak Lembar Hasil"
                    >
                      <Printer class="w-3.5 h-3.5 text-slate-600" />
                    </BaseButton>
                  </template>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Modal Input / Edit Hasil Pemeriksaan -->
    <BaseModal
      v-model="showResultModal"
      :title="selectedOrder?.status === 'SELESAI' ? 'Detail & Edit Hasil Pemeriksaan' : 'Validasi & Input Hasil Pemeriksaan'"
      width="lg"
    >
      <div v-if="selectedOrder" class="space-y-4">
        <!-- Order Context Header -->
        <div class="bg-sky-50 border border-sky-100 p-3 rounded-lg text-xs space-y-1">
          <div class="flex justify-between items-center">
            <span class="font-bold text-sky-800 text-sm">{{ selectedOrder.nama_tindakan }} ({{ selectedOrder.jenis_order }})</span>
            <span class="font-mono text-sky-700">{{ selectedOrder.nomor_order }}</span>
          </div>
          <div class="text-slate-600 flex flex-wrap gap-x-4">
            <span>Pasien: <strong>{{ selectedOrder.patient_name }}</strong> (RM: {{ selectedOrder.no_rm }})</span>
            <span>Dokter: <strong>{{ selectedOrder.doctor_name || 'Dokter Pemeriksa' }}</strong></span>
          </div>
          <div v-if="selectedOrder.catatan_klinis" class="text-slate-500 pt-1 border-t border-sky-200/60">
            <em>Catatan Klinis Dokter: {{ selectedOrder.catatan_klinis }}</em>
          </div>
        </div>

        <!-- Input Form -->
        <div class="space-y-3">
          <div>
            <label class="block text-xs font-semibold text-slate-700 mb-1">
              Hasil Eksaminasi & Interpretasi Klinis <span class="text-rose-500">*</span>
            </label>
            <textarea
              v-model="formResult.hasil_pemeriksaan"
              rows="6"
              class="w-full px-3 py-2 text-xs border border-slate-200 rounded-lg focus:outline-none focus:border-sky-500 focus:ring-1 focus:ring-sky-500 font-mono leading-relaxed"
              placeholder="Tuliskan hasil pengujian laboratorium, bacaan radiologi, atau evaluasi fisioterapi..."
            ></textarea>
          </div>

          <div>
            <BaseInput
              v-model="formResult.lampiran_url"
              label="URL Berkas / Lampiran Gambar PACS / DICOM / Dokumen PDF"
              placeholder="https://storage.simrs.id/pacs/rad-10293.dcm atau drive link..."
              hint="Opsional: tautan ke file digital hasil pemindaian penunjang medis"
            />
          </div>
        </div>

        <div class="flex items-center justify-end gap-2 pt-3 border-t border-slate-100">
          <BaseButton variant="ghost" size="sm" @click="showResultModal = false">
            Batal
          </BaseButton>
          <BaseButton
            variant="primary"
            size="sm"
            :loading="submitting"
            @click="saveResult"
          >
            <CheckCircle2 class="w-4 h-4 mr-1.5" />
            Simpan & Validasi Hasil
          </BaseButton>
        </div>
      </div>
    </BaseModal>

    <!-- Modal Cetak Hasil Pemeriksaan -->
    <BaseModal
      v-model="showPrintModal"
      title="Lembar Hasil Pemeriksaan Penunjang Medis"
      width="lg"
    >
      <div v-if="selectedOrder" class="space-y-4">
        <!-- Printable Area -->
        <div id="print-area" class="p-6 bg-white border border-slate-300 rounded-lg shadow-sm text-slate-800 font-sans">
          <!-- Kop Surat SIMRS -->
          <div class="border-b-2 border-slate-800 pb-3 mb-4 text-center">
            <h2 class="text-base font-bold uppercase tracking-wider text-slate-900">INSTALASI PENUNJANG MEDIS SIMRS-ASOY</h2>
            <p class="text-xs text-slate-600">Jl. Sehat Bersama No. 12, Jakarta Selatan | Telp: (021) 7890-1234</p>
            <p class="text-[10px] text-slate-500">Izin Operasional Kemenkes RI No: 440/123/SIMRS/2026</p>
          </div>

          <div class="text-center font-bold text-sm uppercase underline decoration-1 underline-offset-4 mb-4">
            HASIL PEMERIKSAAN {{ selectedOrder.jenis_order }}
          </div>

          <!-- Metadata Pasien -->
          <div class="grid grid-cols-2 gap-2 text-xs mb-4 pb-3 border-b border-slate-200">
            <div>
              <div class="grid grid-cols-3">
                <span class="text-slate-500">No. Order</span>
                <span class="col-span-2 font-mono font-bold">: {{ selectedOrder.nomor_order }}</span>
              </div>
              <div class="grid grid-cols-3">
                <span class="text-slate-500">No. RM</span>
                <span class="col-span-2 font-mono font-bold">: {{ selectedOrder.no_rm }}</span>
              </div>
              <div class="grid grid-cols-3">
                <span class="text-slate-500">Nama Pasien</span>
                <span class="col-span-2 font-bold">: {{ selectedOrder.patient_name }}</span>
              </div>
            </div>
            <div>
              <div class="grid grid-cols-3">
                <span class="text-slate-500">Tgl Periksa</span>
                <span class="col-span-2">: {{ formatDate(selectedOrder.created_at) }}</span>
              </div>
              <div class="grid grid-cols-3">
                <span class="text-slate-500">Dokter Perujuk</span>
                <span class="col-span-2">: {{ selectedOrder.doctor_name || 'Dokter DPJP' }}</span>
              </div>
              <div class="grid grid-cols-3">
                <span class="text-slate-500">Pemeriksaan</span>
                <span class="col-span-2 font-bold text-sky-800">: {{ selectedOrder.nama_tindakan }}</span>
              </div>
            </div>
          </div>

          <!-- Isi Hasil -->
          <div class="mb-6">
            <h4 class="text-xs font-bold text-slate-700 uppercase mb-2">Hasil Pemeriksaan & Catatan Klinis:</h4>
            <div class="bg-slate-50 p-3 rounded border border-slate-200 text-xs font-mono whitespace-pre-line leading-relaxed">
              {{ selectedOrder.hasil_pemeriksaan || 'Belum ada hasil yang diunggah.' }}
            </div>
            <div v-if="selectedOrder.lampiran_url" class="mt-2 text-xs text-sky-700">
              Lampiran Digital: <a :href="selectedOrder.lampiran_url" target="_blank" class="underline">{{ selectedOrder.lampiran_url }}</a>
            </div>
          </div>

          <!-- Tanda Tangan -->
          <div class="grid grid-cols-2 pt-6 text-xs text-center border-t border-slate-200">
            <div>
              <p class="text-slate-500">Dokter DPJP / Perujuk,</p>
              <div class="h-14"></div>
              <p class="font-bold border-t border-slate-300 mx-8 pt-1">{{ selectedOrder.doctor_name || 'Dokter DPJP' }}</p>
            </div>
            <div>
              <p class="text-slate-500">Petugas Penunjang Medis / Pemeriksa,</p>
              <div class="h-14"></div>
              <p class="font-bold border-t border-slate-300 mx-8 pt-1">Analis / Radiografer SIMRS</p>
            </div>
          </div>
        </div>

        <!-- Footer Actions -->
        <div class="flex items-center justify-between pt-2">
          <span class="text-xs text-slate-400">Dokumen rekam medis penunjang tersimpan permanen di basis data.</span>
          <div class="flex items-center gap-2">
            <BaseButton variant="ghost" size="sm" @click="showPrintModal = false">
              Tutup
            </BaseButton>
            <BaseButton variant="primary" size="sm" @click="printDirect">
              <Printer class="w-4 h-4 mr-1.5" />
              Cetak Dokumen
            </BaseButton>
          </div>
        </div>
      </div>
    </BaseModal>
  </div>
</template>

<style scoped>
@media print {
  body * {
    visibility: hidden;
  }
  #print-area, #print-area * {
    visibility: visible;
  }
  #print-area {
    position: absolute;
    left: 0;
    top: 0;
    width: 100%;
    border: none;
    box-shadow: none;
    padding: 0;
  }
}
</style>
