<template>
  <div class="space-y-6">
    <div v-if="requiresSpecificBranch" class="surface-card p-8 text-center">
      <p class="text-lg font-semibold">Select a branch to manage its menu.</p>
      <p class="mt-2 text-sm text-[var(--admin-text-muted)]">Store-wide menu editing is disabled because categories and items are now branch scoped.</p>
    </div>

    <template v-else>
      <div class="flex flex-wrap items-center gap-3">
        <button class="btn-primary" @click="showCategoryForm = !showCategoryForm">
          {{ showCategoryForm ? 'Close Category Form' : 'Add Category' }}
        </button>
      </div>

      <div v-if="showCategoryForm" class="surface-card p-5">
        <form class="grid gap-4 md:grid-cols-[1fr_140px_auto]" @submit.prevent="addCategory">
          <input v-model="newCategory.name" class="input" placeholder="Category name" required />
          <input v-model.number="newCategory.sort_order" class="input" type="number" placeholder="Order" />
          <button class="btn-primary" type="submit">Save Category</button>
        </form>
      </div>

      <div v-if="menu?.categories?.length" class="space-y-4">
        <div v-for="category in menu.categories" :key="category.id" class="surface-card p-5">
          <div class="flex flex-wrap items-center justify-between gap-3 border-b border-[var(--admin-border)] pb-4">
            <div>
              <h3 class="text-lg font-semibold">{{ category.name }}</h3>
              <p class="text-sm text-[var(--admin-text-muted)]">{{ category.items.length }} item(s)</p>
            </div>
            <div class="flex gap-2">
              <button class="btn-secondary" @click="openItemForm(category)">Add Item</button>
              <button class="btn-danger" @click="deleteCategory(category.id)">Delete</button>
            </div>
          </div>

          <div class="mt-4 space-y-3">
            <div v-for="item in category.items" :key="item.id" class="surface-muted flex items-center justify-between gap-3 px-4 py-4">
              <div>
                <p class="font-semibold">{{ item.name }}</p>
                <p class="mt-1 text-sm text-[var(--admin-text-muted)]">{{ item.description }}</p>
                <p class="mt-1 text-xs text-[var(--admin-text-subtle)]">{{ item.modifier_groups?.length || 0 }} modifier group(s)</p>
              </div>
              <div class="flex items-center gap-3">
                <span class="font-semibold">{{ formatPrice(item.base_price) }} UZS</span>
                <button class="btn-danger" @click="deleteItem(item.id)">Delete</button>
              </div>
            </div>
          </div>
        </div>
      </div>
      <EmptyState
        v-else
        :icon="UtensilsCrossed"
        title="No menu categories yet"
        description="Create the first category for the active branch to start building its menu."
        action-label="Add Category"
        @action="showCategoryForm = true"
      />
    </template>

    <Teleport to="body">
      <div v-if="showItemForm">
        <div class="slide-panel-backdrop" @click="showItemForm = false" />
        <aside class="slide-panel">
          <div class="mb-6 flex items-center justify-between">
            <div>
              <p class="text-lg font-semibold">Add Item</p>
              <p class="text-sm text-[var(--admin-text-muted)]">New item for {{ selectedCategory?.name }}</p>
            </div>
            <button class="btn-ghost" @click="showItemForm = false">Close</button>
          </div>
          <form class="field-grid" @submit.prevent="addItem">
            <div>
              <label class="label">Item Name</label>
              <input v-model="newItem.name" class="input" required />
            </div>
            <div>
              <label class="label">Description</label>
              <textarea v-model="newItem.description" class="textarea" />
            </div>
            <div>
              <label class="label">Base Price</label>
              <input v-model.number="newItem.base_price" class="input" type="number" required />
            </div>
            <button class="btn-primary" type="submit">Create Item</button>
          </form>
        </aside>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { UtensilsCrossed } from 'lucide-vue-next'

interface MenuCategory {
  id: string
  name: string
  items: Array<{
    id: string
    name: string
    description: string
    base_price: number
    modifier_groups?: Array<unknown>
  }>
}

interface MenuPayload {
  categories: MenuCategory[]
}

const { api } = useApi()
const branchContext = useBranchContext()
const auth = useAuth()

const menu = ref<MenuPayload | null>(null)
const selectedCategory = ref<MenuCategory | null>(null)
const showCategoryForm = ref(false)
const showItemForm = ref(false)

const newCategory = reactive({ name: '', sort_order: 0 })
const newItem = reactive({ name: '', description: '', base_price: 0 })

const requiresSpecificBranch = computed(() => auth.state.staff?.role === 'director' && !branchContext.selectedBranchId.value)

function menuPath(path: string) {
  if (branchContext.selectedBranchId.value) {
    const glue = path.includes('?') ? '&' : '?'
    return `${path}${glue}branch_id=${branchContext.selectedBranchId.value}`
  }
  return path
}

async function loadMenu() {
  if (requiresSpecificBranch.value) {
    menu.value = null
    return
  }
  menu.value = await api<MenuPayload>(menuPath('/admin/menu'))
}

async function addCategory() {
  await api('/admin/menu/categories', {
    method: 'POST',
    body: {
      ...newCategory,
      branch_id: branchContext.selectedBranchId.value,
    },
  })
  newCategory.name = ''
  newCategory.sort_order = 0
  showCategoryForm.value = false
  await loadMenu()
}

async function deleteCategory(id: string) {
  if (!window.confirm('Delete this category and all its items?')) return
  await api(`/admin/menu/categories/${id}`, { method: 'DELETE' })
  await loadMenu()
}

function openItemForm(category: MenuCategory) {
  selectedCategory.value = category
  showItemForm.value = true
}

async function addItem() {
  if (!selectedCategory.value) return
  await api('/admin/menu/items', {
    method: 'POST',
    body: {
      branch_id: branchContext.selectedBranchId.value,
      category_id: selectedCategory.value.id,
      name: newItem.name,
      description: newItem.description,
      base_price: newItem.base_price,
    },
  })
  newItem.name = ''
  newItem.description = ''
  newItem.base_price = 0
  showItemForm.value = false
  await loadMenu()
}

async function deleteItem(id: string) {
  if (!window.confirm('Delete this item?')) return
  await api(`/admin/menu/items/${id}`, { method: 'DELETE' })
  await loadMenu()
}

function formatPrice(price: number) {
  return price.toLocaleString('en-US')
}

onMounted(loadMenu)
watch(() => branchContext.selectedBranchId.value, loadMenu)
</script>
