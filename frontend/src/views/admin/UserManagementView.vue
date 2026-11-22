<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api/client'
import { useUiStore } from '@/stores/ui'
import DataTable, { type Column } from '@/components/common/DataTable.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseModal from '@/components/common/BaseModal.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import type { User, MasterPoli } from '@/types'

const uiStore = useUiStore()

const users = ref<User[]>([])
const roles = ref<{ value: number; label: string }[]>([])
const poliList = ref<{ value: number; label: string }[]>([])
const isLoading = ref(false)
const search = ref('')
const selectedRole = ref<number | ''>('')
const page = ref(1)
const total = ref(0)
const limit = 15

// Modal state
const isModalOpen = ref(false)
const isEditing = ref(false)
const isSubmitting = ref(false)
const currentUserId = ref('')

const form = ref({
  username: '',
  password: '',
  email: '',
  full_name: '',
  nip: '',
  phone: '',
  role_id: 1,
  poli_id: null as number | null,
  status_aktif: true,
})

const columns: Column[] = [
  { key: 'username', label: 'Username', width: '130px' },
  { key: 'full_name', label: 'Nama Lengkap Pegawai' },
  { key: 'role_name', label: 'Role / Wewenang', width: '160px' },
  { key: 'nip', label: 'NIP' },
  { key: 'poli_name', label: 'Unit / Poli' },
  { key: 'status_aktif', label: 'Status Akun', align: 'center', width: '120px' },
  { key: 'actions', label: 'Aksi', align: 'right', width: '140px' },
]

async function loadUsers() {
  isLoading.value = true
  try {
    const res = await api.get<User[]>('/users', {
      search: search.value,
      role_id: selectedRole.value,
      page: page.value,
      limit,
    })
    if (res.success && res.data) {
      users.value = res.data
      total.value = res.meta?.total_data || 0
    }
  } catch (err: any) {
    uiStore.showToast(err.message, 'danger')
  } finally {
    isLoading.value = false
  }
}

async function loadMasterData() {
  try {
    const [rolesRes, poliRes] = await Promise.all([
      api.get<any[]>('/master/roles'),
      api.get<MasterPoli[]>('/master/poli')
    ])
    if (rolesRes.success && rolesRes.data) {
      roles.value = rolesRes.data.map(r => ({ value: r.id, label: `${r.name} (${r.code})` }))
    }
    if (poliRes.success && poliRes.data) {
      poliList.value = poliRes.data.map(p => ({ value: p.id, label: p.name }))
    }
  } catch (err) {
    console.error('Failed to load master options:', err)
  }
}

function openAddModal() {
  isEditing.value = false
  currentUserId.value = ''
  form.value = {
    username: '',
    password: '',
    email: '',
    full_name: '',
    nip: '',
    phone: '',
    role_id: 1,
    poli_id: null,
    status_aktif: true,
  }
  isModalOpen.value = true
}

function openEditModal(user: User) {
  isEditing.value = true
  currentUserId.value = user.id
  form.value = {
    username: user.username,
    password: '',
    email: user.email,
    full_name: user.full_name,
    nip: user.nip,
    phone: user.phone,
    role_id: user.role_id,
    poli_id: user.poli_id || null,
    status_aktif: user.status_aktif,
  }
  isModalOpen.value = true
}

async function saveUser() {
  isSubmitting.value = true
  try {
    if (isEditing.value) {
      await api.put(`/users/${currentUserId.value}`, form.value)
      uiStore.showToast('Data pengguna berhasil diperbarui', 'success')
    } else {
      await api.post('/users', form.value)
      uiStore.showToast('Pengguna baru berhasil ditambahkan', 'success')
    }
    isModalOpen.value = false
    loadUsers()
  } catch (err: any) {
    uiStore.showToast(err.message || 'Gagal menyimpan pengguna', 'danger')
  } finally {
    isSubmitting.value = false
  }
}

function confirmDelete(user: User) {
  uiStore.openConfirm({
    title: 'Hapus Akun Pengguna',
    message: `Apakah Anda yakin ingin menonaktifkan/menghapus akun "${user.full_name}" (@${user.username})? Tindakan ini dicatat pada audit log.`,
    confirmText: 'Ya, Hapus',
    variant: 'danger',
    onConfirm: async () => {
      try {
        await api.delete(`/users/${user.id}`)
        uiStore.showToast('Pengguna berhasil dihapus', 'success')
        loadUsers()
      } catch (err: any) {
        uiStore.showToast(err.message || 'Gagal menghapus pengguna', 'danger')
      }
    }
  })
}

function onPageChange(newPage: number) {
  page.value = newPage
  loadUsers()
}

onMounted(() => {
  loadUsers()
  loadMasterData()
})
</script>

<template>
  <div class="space-y-4 text-left">
    <!-- Header -->
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-[20px] font-bold text-[#0F172A] leading-[28px]">
          Manajemen User & Role Access Control (RBAC)
        </h1>
        <p class="text-[13px] text-[#475569]">
          Kelola akun seluruh tenaga medis, staf administrasi, perawat, apoteker, dan kasir sesuai matriks 9 role SIMRS.
        </p>
      </div>

      <BaseButton variant="primary" size="md" @click="openAddModal">
        <template #icon>
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
          </svg>
        </template>
        Tambah Pegawai Baru
      </BaseButton>
    </div>

    <!-- Filters Bar -->
    <div class="bg-white p-3 rounded-lg border border-[#E2E8F0] shadow-sm flex flex-wrap items-center gap-3">
      <div class="flex-1 min-w-[200px]">
        <BaseInput
          v-model="search"
          placeholder="Cari username, nama pegawai, NIP..."
          compact
          @keyup.enter="loadUsers"
        />
      </div>

      <div class="w-[200px]">
        <BaseSelect
          v-model="selectedRole"
          :options="[{ value: '', label: 'Semua Role' }, ...roles]"
          placeholder="Filter Role"
          compact
          @change="loadUsers"
        />
      </div>

      <BaseButton variant="secondary" size="sm" @click="loadUsers">
        Cari
      </BaseButton>
    </div>

    <!-- Users DataTable -->
    <DataTable
      :columns="columns"
      :data="users"
      :loading="isLoading"
      :page="page"
      :limit="limit"
      :total="total"
      @page-change="onPageChange"
    >
      <template #username="{ row }">
        <span class="font-mono font-medium text-[#0284C7]">{{ row.username }}</span>
      </template>

      <template #role_name="{ row }">
        <StatusBadge :status="row.role_code">
          {{ row.role_name }}
        </StatusBadge>
      </template>

      <template #poli_name="{ row }">
        <span class="text-[#475569]">{{ row.poli_name || '-' }}</span>
      </template>

      <template #status_aktif="{ row }">
        <span
          :class="[
            'px-2 py-0.5 rounded text-[11px] font-medium border',
            row.status_aktif ? 'bg-[#DCFCE7] text-[#15803D] border-[#86EFAC]' : 'bg-[#FEE2E2] text-[#B91C1C] border-[#FCA5A5]'
          ]"
        >
          {{ row.status_aktif ? 'Aktif' : 'Nonaktif' }}
        </span>
      </template>

      <template #actions="{ row }">
        <div class="flex items-center justify-end gap-1.5">
          <BaseButton variant="soft" size="sm" @click="openEditModal(row)">
            Edit
          </BaseButton>
          <BaseButton variant="danger" size="sm" @click="confirmDelete(row)">
            Hapus
          </BaseButton>
        </div>
      </template>
    </DataTable>

    <!-- Modal Form User -->
    <BaseModal
      :is-open="isModalOpen"
      :title="isEditing ? 'Ubah Data Pegawai' : 'Tambah Pegawai Baru'"
      max-width="lg"
      @close="isModalOpen = false"
    >
      <form class="space-y-3" @submit.prevent="saveUser">
        <div class="grid grid-cols-2 gap-3">
          <BaseInput
            v-model="form.username"
            label="Username"
            placeholder="contoh: dokter.budi"
            :disabled="isEditing"
            required
          />

          <BaseInput
            v-model="form.password"
            label="Kata Sandi"
            type="password"
            :placeholder="isEditing ? 'Kosongkan jika tidak diubah' : 'Minimal 6 karakter'"
            :required="!isEditing"
          />
        </div>

        <div class="grid grid-cols-2 gap-3">
          <BaseInput
            v-model="form.full_name"
            label="Nama Lengkap & Gelar"
            placeholder="contoh: dr. Budi Santoso, Sp.A"
            required
          />

          <BaseInput
            v-model="form.email"
            label="Alamat Email"
            type="email"
            placeholder="pegawai@simrs-asoy.local"
            required
          />
        </div>

        <div class="grid grid-cols-2 gap-3">
          <BaseInput
            v-model="form.nip"
            label="NIP / Nomor Pegawai"
            placeholder="contoh: DOC-009"
          />

          <BaseInput
            v-model="form.phone"
            label="Nomor Telepon / WhatsApp"
            placeholder="081234567890"
          />
        </div>

        <div class="grid grid-cols-2 gap-3">
          <BaseSelect
            v-model="form.role_id"
            label="Wewenang Role Operasional (RBAC)"
            :options="roles"
            required
          />

          <BaseSelect
            v-model="form.poli_id"
            label="Penempatan Unit / Poli"
            :options="[{ value: '', label: '-- Tidak Ada / Staf Umum --' }, ...poliList]"
          />
        </div>

        <div v-if="isEditing" class="pt-2">
          <label class="flex items-center gap-2 cursor-pointer text-[13px] text-[#334155]">
            <input
              type="checkbox"
              v-model="form.status_aktif"
              class="w-4 h-4 text-[#0284C7] rounded border-[#CBD5E1] focus:ring-[#0284C7]"
            />
            <span>Status Akun Aktif (Dapat Login ke SIMRS)</span>
          </label>
        </div>
      </form>

      <template #footer>
        <BaseButton variant="secondary" size="md" @click="isModalOpen = false">
          Batal
        </BaseButton>
        <BaseButton variant="primary" size="md" :loading="isSubmitting" @click="saveUser">
          {{ isEditing ? 'Simpan Perubahan' : 'Daftarkan Pegawai' }}
        </BaseButton>
      </template>
    </BaseModal>
  </div>
</template>
