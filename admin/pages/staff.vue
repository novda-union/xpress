<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <Button @click="openCreate">
        <Plus class="h-4 w-4" />
        Add Staff
      </Button>
    </div>

    <EmptyState
      v-if="!groups.length"
      :icon="Users"
      title="No staff members found"
      description="Create managers and baristas to handle operations inside each branch."
      action-label="Add Staff"
      @action="openCreate"
    />
    <StaffGroupList
      v-else
      :groups="groups"
      @edit="openEdit"
      @deactivate="deactivateStaff"
    />

    <StaffForm
      :open="drawerOpen"
      :staff="selectedStaff"
      :branches="branches"
      :role="auth.state.staff?.role ?? 'barista'"
      :loading="loading"
      @close="closeDrawer"
      @save="saveStaff"
    />
  </div>
</template>

<script setup lang="ts">
import { Plus, Users } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import StaffForm from '@/components/staff/StaffForm.vue'
import StaffGroupList from '@/components/staff/StaffGroupList.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import type { BranchSummary, Staff, StaffGroup } from '@/types/auth'

const { api } = useApi()
const auth = useAuth()

const groups = ref<StaffGroup[]>([])
const branches = ref<BranchSummary[]>([])
const selectedStaff = ref<Staff | null>(null)
const drawerOpen = ref(false)
const loading = ref(false)

async function loadData() {
  groups.value = await api<StaffGroup[]>('/admin/staff')
  branches.value = await api<BranchSummary[]>('/admin/branches').catch(() => [])
}

function openCreate() {
  selectedStaff.value = null
  drawerOpen.value = true
}

function openEdit(staff: Staff) {
  selectedStaff.value = staff
  drawerOpen.value = true
}

function closeDrawer() {
  drawerOpen.value = false
}

async function saveStaff(payload: Record<string, unknown>) {
  loading.value = true
  try {
    if (selectedStaff.value) {
      await api(`/admin/staff/${selectedStaff.value.id}`, {
        method: 'PUT',
        body: payload,
      })
    } else {
      await api('/admin/staff', {
        method: 'POST',
        body: payload,
      })
    }
    drawerOpen.value = false
    await loadData()
  } finally {
    loading.value = false
  }
}

async function deactivateStaff(staff: Staff) {
  await api(`/admin/staff/${staff.id}`, { method: 'DELETE' })
  await loadData()
}

onMounted(loadData)
</script>
