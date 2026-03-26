<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h2 class="text-2xl font-bold">Menu</h2>
      <button @click="showAddCategory = true" class="bg-blue-600 text-white py-2 px-4 rounded hover:bg-blue-700">
        Add Category
      </button>
    </div>

    <!-- Add Category Form -->
    <div v-if="showAddCategory" class="bg-white p-4 rounded-lg shadow mb-6">
      <form @submit.prevent="addCategory" class="flex gap-3">
        <input
          v-model="newCategory.name"
          placeholder="Category name"
          class="flex-1 px-3 py-2 border rounded"
          required
        />
        <input
          v-model.number="newCategory.sort_order"
          type="number"
          placeholder="Order"
          class="w-20 px-3 py-2 border rounded"
        />
        <button type="submit" class="bg-green-600 text-white px-4 py-2 rounded">Save</button>
        <button type="button" @click="showAddCategory = false" class="bg-gray-300 px-4 py-2 rounded">Cancel</button>
      </form>
    </div>

    <!-- Categories -->
    <div class="space-y-4">
      <div v-for="cat in menu?.categories" :key="cat.id" class="bg-white rounded-lg shadow">
        <div class="flex justify-between items-center p-4 border-b">
          <h3 class="font-semibold text-lg">{{ cat.name }}</h3>
          <div class="flex gap-2">
            <button
              @click="selectedCategory = cat; showAddItem = true"
              class="text-sm bg-blue-100 text-blue-700 px-3 py-1 rounded hover:bg-blue-200"
            >
              Add Item
            </button>
            <button
              @click="deleteCategory(cat.id)"
              class="text-sm bg-red-100 text-red-700 px-3 py-1 rounded hover:bg-red-200"
            >
              Delete
            </button>
          </div>
        </div>

        <div class="p-4 space-y-3">
          <div v-for="item in cat.items" :key="item.id" class="flex justify-between items-center py-2 border-b last:border-0">
            <div>
              <p class="font-medium">{{ item.name }}</p>
              <p class="text-sm text-gray-500">{{ item.description }}</p>
              <p class="text-sm text-gray-400" v-if="item.modifier_groups?.length">
                {{ item.modifier_groups.length }} modifier group(s)
              </p>
            </div>
            <div class="flex items-center gap-3">
              <span class="font-semibold">{{ formatPrice(item.base_price) }} UZS</span>
              <button @click="deleteItem(item.id)" class="text-red-500 text-sm hover:text-red-700">Delete</button>
            </div>
          </div>
          <p v-if="!cat.items?.length" class="text-gray-400 text-sm">No items in this category</p>
        </div>
      </div>
    </div>

    <!-- Add Item Modal -->
    <div v-if="showAddItem" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div class="bg-white p-6 rounded-lg w-full max-w-md">
        <h3 class="text-lg font-bold mb-4">Add Item to {{ selectedCategory?.name }}</h3>
        <form @submit.prevent="addItem" class="space-y-3">
          <input v-model="newItem.name" placeholder="Item name" class="w-full px-3 py-2 border rounded" required />
          <input v-model="newItem.description" placeholder="Description" class="w-full px-3 py-2 border rounded" />
          <input v-model.number="newItem.base_price" type="number" placeholder="Price (UZS)" class="w-full px-3 py-2 border rounded" required />
          <div class="flex gap-3">
            <button type="submit" class="flex-1 bg-green-600 text-white py-2 rounded">Save</button>
            <button type="button" @click="showAddItem = false" class="flex-1 bg-gray-300 py-2 rounded">Cancel</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const { api } = useApi()

const menu = ref<any>(null)
const showAddCategory = ref(false)
const showAddItem = ref(false)
const selectedCategory = ref<any>(null)
const newCategory = reactive({ name: '', sort_order: 0 })
const newItem = reactive({ name: '', description: '', base_price: 0 })

async function loadMenu() {
  try {
    menu.value = await api<any>('/admin/menu')
  } catch (e) {
    console.error('Failed to load menu', e)
  }
}

async function addCategory() {
  try {
    await api('/admin/menu/categories', { method: 'POST', body: { ...newCategory } })
    newCategory.name = ''
    newCategory.sort_order = 0
    showAddCategory.value = false
    await loadMenu()
  } catch (e: any) {
    alert(e?.data?.error || 'Failed to add category')
  }
}

async function deleteCategory(id: string) {
  if (!confirm('Delete this category and all its items?')) return
  try {
    await api(`/admin/menu/categories/${id}`, { method: 'DELETE' })
    await loadMenu()
  } catch (e: any) {
    alert(e?.data?.error || 'Failed to delete category')
  }
}

async function addItem() {
  if (!selectedCategory.value) return
  try {
    await api('/admin/menu/items', {
      method: 'POST',
      body: {
        category_id: selectedCategory.value.id,
        name: newItem.name,
        description: newItem.description,
        base_price: newItem.base_price,
      },
    })
    newItem.name = ''
    newItem.description = ''
    newItem.base_price = 0
    showAddItem.value = false
    await loadMenu()
  } catch (e: any) {
    alert(e?.data?.error || 'Failed to add item')
  }
}

async function deleteItem(id: string) {
  if (!confirm('Delete this item?')) return
  try {
    await api(`/admin/menu/items/${id}`, { method: 'DELETE' })
    await loadMenu()
  } catch (e: any) {
    alert(e?.data?.error || 'Failed to delete item')
  }
}

function formatPrice(price: number) {
  return price?.toLocaleString('en-US') || '0'
}

onMounted(loadMenu)
</script>
