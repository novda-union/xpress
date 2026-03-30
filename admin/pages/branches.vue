<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <Button @click="openCreate">
        <Plus class="h-4 w-4" />
        Add Branch
      </Button>
    </div>

    <EmptyState
      v-if="!branches.length"
      :icon="Store"
      title="No branches yet"
      description="Create the first branch to start assigning staff and publishing a branch menu."
      action-label="Create Branch"
      @action="openCreate"
    />
    <BranchTable
      v-else
      :branches="branches"
      @edit="openEdit"
      @deactivate="deactivateBranch"
    />

    <BranchForm
      :open="drawerOpen"
      :branch="selectedBranch"
      :loading="loading"
      @close="closeDrawer"
      @save="saveBranch"
    />
  </div>
</template>

<script setup lang="ts">
import { Plus, Store } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import BranchForm from 'components/branches/BranchForm.vue'
import BranchTable from 'components/branches/BranchTable.vue'
import EmptyState from 'components/ui/EmptyState.vue'
import type { BranchSummary } from 'types/auth'

const { api } = useApi()
const branchContext = useBranchContext()

const branches = ref<BranchSummary[]>([])
const selectedBranch = ref<BranchSummary | null>(null)
const drawerOpen = ref(false)
const loading = ref(false)

async function loadBranches() {
  branches.value = await api<BranchSummary[]>('/admin/branches')
  await branchContext.loadBranches()
}

function openCreate() {
  selectedBranch.value = null
  drawerOpen.value = true
}

function openEdit(branch: BranchSummary) {
  selectedBranch.value = branch
  drawerOpen.value = true
}

function closeDrawer() {
  drawerOpen.value = false
}

async function saveBranch(payload: Record<string, unknown>) {
  loading.value = true
  try {
    if (selectedBranch.value) {
      await api(`/admin/branches/${selectedBranch.value.id}`, {
        method: 'PUT',
        body: payload,
      })
    } else {
      await api('/admin/branches', {
        method: 'POST',
        body: payload,
      })
    }
    drawerOpen.value = false
    await loadBranches()
  } finally {
    loading.value = false
  }
}

async function deactivateBranch(branch: BranchSummary) {
  await api(`/admin/branches/${branch.id}`, { method: 'DELETE' })
  await loadBranches()
}

onMounted(loadBranches)
</script>
