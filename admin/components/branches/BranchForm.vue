<template>
  <Teleport to="body">
    <div v-if="open">
      <div class="slide-panel-backdrop" @click="$emit('close')" />
      <aside class="slide-panel">
        <div class="mb-6 flex items-center justify-between">
          <div>
            <p class="text-lg font-semibold">{{ branch ? 'Edit Branch' : 'Add Branch' }}</p>
            <p class="text-sm text-[var(--admin-text-muted)]">Manage branch contact and location data.</p>
          </div>
          <button class="btn-ghost" @click="$emit('close')">Close</button>
        </div>

        <form class="field-grid" @submit.prevent="submit">
          <div>
            <label class="label">Name</label>
            <input v-model="form.name" class="input" required />
          </div>
          <div>
            <label class="label">Address</label>
            <textarea v-model="form.address" class="textarea" required />
          </div>
          <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
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
            <label class="label">Banner Image URL</label>
            <input v-model="form.banner_image_url" class="input" placeholder="https://..." />
          </div>
          <div>
            <label class="label">Telegram Group Chat ID</label>
            <input v-model.number="form.telegram_group_chat_id" class="input" type="number" />
          </div>
          <label class="flex items-center gap-3 rounded-2xl border border-[var(--admin-border)] bg-white px-4 py-3">
            <input v-model="form.is_active" type="checkbox" />
            <span class="text-sm font-medium">Branch is active</span>
          </label>

          <div class="flex gap-3 pt-2">
            <button type="submit" class="btn-primary flex-1" :disabled="loading">
              {{ loading ? 'Saving...' : branch ? 'Save Changes' : 'Create Branch' }}
            </button>
            <button type="button" class="btn-secondary" @click="$emit('close')">Cancel</button>
          </div>
        </form>
      </aside>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import type { BranchSummary } from '~/types/auth'

const props = defineProps<{
  branch?: BranchSummary | null
  loading?: boolean
  open: boolean
}>()

const emit = defineEmits<{
  close: []
  save: [payload: Record<string, unknown>]
}>()

const form = reactive({
  name: '',
  address: '',
  lat: null as number | null,
  lng: null as number | null,
  banner_image_url: '',
  telegram_group_chat_id: null as number | null,
  is_active: true,
})

watch(
  () => props.branch,
  (branch) => {
    form.name = branch?.name ?? ''
    form.address = branch?.address ?? ''
    form.lat = branch?.lat ?? null
    form.lng = branch?.lng ?? null
    form.banner_image_url = branch?.banner_image_url ?? ''
    form.telegram_group_chat_id = branch?.telegram_group_chat_id ?? null
    form.is_active = branch?.is_active ?? true
  },
  { immediate: true },
)

function submit() {
  emit('save', {
    ...form,
    lat: Number.isFinite(form.lat) ? form.lat : null,
    lng: Number.isFinite(form.lng) ? form.lng : null,
    telegram_group_chat_id: Number.isFinite(form.telegram_group_chat_id) ? form.telegram_group_chat_id : null,
  })
}
</script>
