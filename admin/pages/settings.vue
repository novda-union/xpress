<template>
  <div>
    <h2 class="text-2xl font-bold mb-6">Settings</h2>

    <div class="bg-white p-6 rounded-lg shadow max-w-lg">
      <form @submit.prevent="saveSettings" class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Store Name</label>
          <input v-model="form.name" class="w-full px-3 py-2 border rounded" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Description</label>
          <textarea v-model="form.description" class="w-full px-3 py-2 border rounded" rows="3" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Address</label>
          <input v-model="form.address" class="w-full px-3 py-2 border rounded" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Phone</label>
          <input v-model="form.phone" class="w-full px-3 py-2 border rounded" />
        </div>
        <button type="submit" class="bg-blue-600 text-white py-2 px-4 rounded hover:bg-blue-700">
          Save
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
const { api } = useApi()
const form = reactive({ name: '', description: '', address: '', phone: '', logo_url: '' })

onMounted(async () => {
  try {
    const store = await api<any>('/admin/store')
    Object.assign(form, {
      name: store.name,
      description: store.description,
      address: store.address,
      phone: store.phone,
      logo_url: store.logo_url,
    })
  } catch (e) {
    console.error('Failed to load store', e)
  }
})

async function saveSettings() {
  try {
    await api('/admin/store', { method: 'PUT', body: { ...form } })
    alert('Settings saved!')
  } catch (e: any) {
    alert(e?.data?.error || 'Failed to save settings')
  }
}
</script>
