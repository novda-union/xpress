<template>
  <div class="max-w-3xl">
    <Card>
      <CardContent class="p-6">
        <form class="space-y-4" @submit.prevent="saveSettings">
          <div class="space-y-1.5">
            <Label for="store-name">Store Name</Label>
            <Input id="store-name" v-model="form.name" />
          </div>
          <div class="space-y-1.5">
            <Label for="store-desc">Description</Label>
            <Textarea id="store-desc" v-model="form.description" />
          </div>
          <div class="grid gap-4 md:grid-cols-2">
            <div class="space-y-1.5">
              <Label for="store-address">Address</Label>
              <Input id="store-address" v-model="form.address" />
            </div>
            <div class="space-y-1.5">
              <Label for="store-phone">Phone</Label>
              <Input id="store-phone" v-model="form.phone" />
            </div>
          </div>
          <div class="space-y-1.5">
            <Label for="store-logo">Logo URL</Label>
            <Input id="store-logo" v-model="form.logo_url" placeholder="https://..." />
          </div>
          <div class="pt-2">
            <Button type="submit" :disabled="saving">
              <span v-if="saving" class="flex items-center gap-2">
                <span class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
                Saving...
              </span>
              <span v-else>Save Store Settings</span>
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

const { api } = useApi()
const saving = ref(false)
const form = reactive({ name: '', description: '', address: '', phone: '', logo_url: '' })

interface StoreSettings {
  name: string
  description: string
  address: string
  phone: string
  logo_url: string
}

onMounted(async () => {
  const store = await api<StoreSettings>('/admin/store')
  Object.assign(form, {
    address: store.address,
    description: store.description,
    logo_url: store.logo_url,
    name: store.name,
    phone: store.phone,
  })
})

async function saveSettings() {
  saving.value = true
  try {
    await api('/admin/store', { method: 'PUT', body: { ...form } })
  } finally {
    saving.value = false
  }
}
</script>
