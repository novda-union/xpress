<template>
  <div class="max-w-3xl">
    <div class="surface-card p-6">
      <form class="field-grid" @submit.prevent="saveSettings">
        <div>
          <label class="label">Store Name</label>
          <input v-model="form.name" class="input" />
        </div>
        <div>
          <label class="label">Description</label>
          <textarea v-model="form.description" class="textarea" />
        </div>
        <div class="grid gap-4 md:grid-cols-2">
          <div>
            <label class="label">Address</label>
            <input v-model="form.address" class="input" />
          </div>
          <div>
            <label class="label">Phone</label>
            <input v-model="form.phone" class="input" />
          </div>
        </div>
        <div>
          <label class="label">Logo URL</label>
          <input v-model="form.logo_url" class="input" />
        </div>
        <div class="pt-2">
          <button class="btn-primary" type="submit">Save Store Settings</button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
const { api } = useApi()
const form = reactive({ name: '', description: '', address: '', phone: '', logo_url: '' })

onMounted(async () => {
  const store = await api<any>('/admin/store')
  Object.assign(form, {
    address: store.address,
    description: store.description,
    logo_url: store.logo_url,
    name: store.name,
    phone: store.phone,
  })
})

async function saveSettings() {
  await api('/admin/store', { method: 'PUT', body: { ...form } })
}
</script>
