<template>
  <Sheet :open="open" @update:open="(val) => !val && $emit('close')">
    <SheetContent side="right" class="w-full overflow-y-auto sm:max-w-[34rem]">
      <SheetHeader>
        <SheetTitle>{{ branch ? 'Edit Branch' : 'Add Branch' }}</SheetTitle>
        <SheetDescription>Manage branch contact and location data.</SheetDescription>
      </SheetHeader>

      <form class="mt-6 space-y-4" @submit.prevent="submit">
        <div class="space-y-1.5">
          <Label for="branch-name">Name</Label>
          <Input id="branch-name" v-model="form.name" required />
        </div>

        <div class="space-y-1.5">
          <Label for="branch-address">Address</Label>
          <Textarea id="branch-address" v-model="form.address" required />
        </div>

        <div class="grid grid-cols-2 gap-4">
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
          <Label for="branch-banner">Banner Image URL</Label>
          <Input id="branch-banner" v-model="form.banner_image_url" placeholder="https://..." />
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

        <div class="flex items-center gap-3 rounded-lg border bg-muted/30 px-4 py-3">
          <Checkbox id="branch-active" v-model:checked="form.is_active" />
          <Label for="branch-active" class="cursor-pointer font-medium">Branch is active</Label>
        </div>

        <SheetFooter class="pt-2">
          <Button type="button" variant="outline" @click="$emit('close')">Cancel</Button>
          <Button type="submit" :disabled="loading">
            <span v-if="loading" class="flex items-center gap-2">
              <span class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
              Saving...
            </span>
            <span v-else>{{ branch ? 'Save Changes' : 'Create Branch' }}</span>
          </Button>
        </SheetFooter>
      </form>
    </SheetContent>
  </Sheet>
</template>

<script setup lang="ts">
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'
import type { BranchSummary } from 'types/auth'

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
