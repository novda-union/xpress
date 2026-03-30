<template>
  <div class="max-w-3xl">
    <Card v-if="showLoadingState">
      <CardContent class="p-8">
        <div class="flex items-center gap-3">
          <span class="h-5 w-5 animate-spin rounded-full border-2 border-primary border-t-transparent" />
          <div>
            <p class="text-lg font-semibold">Loading branch settings</p>
            <p class="mt-1 text-sm text-muted-foreground">Resolving the active branch context for this page.</p>
          </div>
        </div>
      </CardContent>
    </Card>

    <Card v-else-if="!branchContext.selectedBranchId.value">
      <CardContent class="p-8 text-center">
        <p class="text-lg font-semibold">Select a branch first</p>
        <p class="mt-2 text-sm text-muted-foreground">Branch settings are only available when a specific branch is active.</p>
      </CardContent>
    </Card>

    <Card v-else>
      <CardContent class="p-6">
        <form class="space-y-4" @submit.prevent="saveSettings">
          <div class="space-y-1.5">
            <Label for="branch-name">Branch Name</Label>
            <Input id="branch-name" v-model="form.name" />
          </div>
          <div class="space-y-1.5">
            <Label for="branch-address">Address</Label>
            <Textarea id="branch-address" v-model="form.address" />
          </div>
          <div class="grid gap-4 md:grid-cols-2">
            <div class="space-y-1.5">
              <Label for="branch-lat">Latitude</Label>
              <Input
                id="branch-lat"
                :model-value="form.lat ?? undefined"
                type="number"
                step="0.000001"
                @update:model-value="updateLat"
              />
            </div>
            <div class="space-y-1.5">
              <Label for="branch-lng">Longitude</Label>
              <Input
                id="branch-lng"
                :model-value="form.lng ?? undefined"
                type="number"
                step="0.000001"
                @update:model-value="updateLng"
              />
            </div>
          </div>
          <div class="space-y-1.5">
            <Label for="branch-tg">Telegram Group Chat ID</Label>
            <Input
              id="branch-tg"
              :model-value="form.telegram_group_chat_id ?? undefined"
              type="number"
              @update:model-value="updateTelegramGroupChatId"
            />
          </div>
          <div class="pt-2">
            <Button type="submit" :disabled="saving">
              <span v-if="saving" class="flex items-center gap-2">
                <span class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
                Saving...
              </span>
              <span v-else>Save Branch Settings</span>
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import type { BranchSummary } from 'types/auth'

const { api } = useApi()
const branchContext = useBranchContext()
const auth = useAuth()
const saving = ref(false)

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
  if (!branchContext.selectedBranchId.value) return

  const branches = await api<BranchSummary[]>('/admin/branches')
  const branch = branches.find((entry) => entry.id === branchContext.selectedBranchId.value)
  if (!branch) return

  form.name = branch.name
  form.address = branch.address
  form.lat = branch.lat
  form.lng = branch.lng
  form.telegram_group_chat_id = branch.telegram_group_chat_id
  form.is_active = branch.is_active
}

async function saveSettings() {
  if (!branchContext.selectedBranchId.value) return

  saving.value = true
  try {
    await api(`/admin/branches/${branchContext.selectedBranchId.value}`, {
      method: 'PUT',
      body: { ...form },
    })
    await branchContext.loadBranches()
  } finally {
    saving.value = false
  }
}

onMounted(loadBranch)
watch(() => branchContext.selectedBranchId.value, loadBranch)

function updateLat(value: string | number) {
  form.lat = toNullableNumber(value)
}

function updateLng(value: string | number) {
  form.lng = toNullableNumber(value)
}

function updateTelegramGroupChatId(value: string | number) {
  form.telegram_group_chat_id = toNullableNumber(value)
}

function toNullableNumber(value: string | number) {
  if (value === '') return null
  const parsed = typeof value === 'number' ? value : Number(value)
  return Number.isFinite(parsed) ? parsed : null
}
</script>
