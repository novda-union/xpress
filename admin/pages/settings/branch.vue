<template>
  <div class="max-w-3xl">
    <div v-if="showLoadingState" class="surface-card p-8">
      <p class="text-lg font-semibold">Loading branch settings</p>
      <p class="mt-2 text-sm text-[var(--admin-text-muted)]">Resolving the active branch context for this page.</p>
    </div>

    <div v-else-if="!branchContext.selectedBranchId.value" class="surface-card p-8 text-center">
      <p class="text-lg font-semibold">Select a branch first</p>
      <p class="mt-2 text-sm text-[var(--admin-text-muted)]">Branch settings are only available when a specific branch is active.</p>
    </div>

    <div v-else class="surface-card p-6">
      <form class="field-grid" @submit.prevent="saveSettings">
        <div>
          <label class="label">Branch Name</label>
          <input v-model="form.name" class="input" />
        </div>
        <div>
          <label class="label">Address</label>
          <textarea v-model="form.address" class="textarea" />
        </div>
        <div class="grid gap-4 md:grid-cols-2">
          <div>
            <label class="label">Latitude</label>
            <input v-model.number="form.lat" class="input" type="number" step="0.000001" />
          </div>
          <div>
            <label class="label">Longitude</label>
            <input v-model.number="form.lng" class="input" type="number" step="0.000001" />
          </div>
        </div>
        <div>
          <label class="label">Telegram Group Chat ID</label>
          <input v-model.number="form.telegram_group_chat_id" class="input" type="number" />
        </div>
        <div class="pt-2">
          <button class="btn-primary" type="submit">Save Branch Settings</button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { BranchSummary } from '~/types/auth'

const { api } = useApi()
const branchContext = useBranchContext()
const auth = useAuth()

const form = reactive({
  name: '',
  address: '',
  lat: null as number | null,
  lng: null as number | null,
  telegram_group_chat_id: null as number | null,
  is_active: true,
})

const showLoadingState = computed(() =>
  branchContext.loading.value ||
  (auth.state.staff?.role === 'director' && !branchContext.branches.value.length),
)

async function loadBranch() {
  if (!branchContext.selectedBranchId.value) {
    return
  }

  const branches = await api<BranchSummary[]>('/admin/branches')
  const branch = branches.find((entry) => entry.id === branchContext.selectedBranchId.value)
  if (!branch) {
    return
  }

  form.name = branch.name
  form.address = branch.address
  form.lat = branch.lat
  form.lng = branch.lng
  form.telegram_group_chat_id = branch.telegram_group_chat_id
  form.is_active = branch.is_active
}

async function saveSettings() {
  if (!branchContext.selectedBranchId.value) {
    return
  }

  await api(`/admin/branches/${branchContext.selectedBranchId.value}`, {
    method: 'PUT',
    body: { ...form },
  })
  await branchContext.loadBranches()
}

onMounted(loadBranch)
watch(() => branchContext.selectedBranchId.value, loadBranch)
</script>
