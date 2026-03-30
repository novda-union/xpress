<template>
  <div class="space-y-6">
    <Card v-if="requiresSpecificBranch">
      <CardContent class="p-8 text-center">
        <p class="text-lg font-semibold">Select a branch to manage its menu.</p>
        <p class="mt-2 text-sm text-muted-foreground">Store-wide menu editing is disabled because categories and items are now branch scoped.</p>
      </CardContent>
    </Card>

    <template v-else>
      <div class="flex flex-wrap items-center gap-3">
        <Button @click="showCategoryForm = !showCategoryForm">
          {{ showCategoryForm ? 'Close Category Form' : 'Add Category' }}
        </Button>
      </div>

      <Card v-if="showCategoryForm">
        <CardContent class="p-5">
          <form class="grid gap-4 md:grid-cols-[1fr_140px_auto]" @submit.prevent="addCategory">
            <Input v-model="newCategory.name" placeholder="Category name" required />
            <Input v-model.number="newCategory.sort_order" type="number" placeholder="Order" />
            <Button type="submit">Save Category</Button>
          </form>
        </CardContent>
      </Card>

      <div v-if="menu?.categories?.length" class="space-y-4">
        <Card v-for="category in menu.categories" :key="category.id">
          <CardContent class="p-5">
            <div class="flex flex-wrap items-center justify-between gap-3 border-b pb-4">
              <div>
                <h3 class="text-lg font-semibold">{{ category.name }}</h3>
                <p class="text-sm text-muted-foreground">{{ category.items.length }} item(s)</p>
              </div>
              <div class="flex gap-2">
                <Button variant="outline" size="sm" @click="openItemForm(category)">Add Item</Button>
                <Button variant="destructive" size="sm" @click="deleteCategory(category.id)">Delete</Button>
              </div>
            </div>

            <div class="mt-4 space-y-3">
              <div
                v-for="item in category.items"
                :key="item.id"
                class="flex items-center justify-between gap-3 rounded-lg bg-muted/50 px-4 py-4"
              >
                <div>
                  <p class="font-semibold">{{ item.name }}</p>
                  <p class="mt-1 text-sm text-muted-foreground">{{ item.description }}</p>
                  <p class="mt-1 text-xs text-muted-foreground/70">{{ item.modifier_groups?.length || 0 }} modifier group(s)</p>
                </div>
                <div class="flex items-center gap-3">
                  <span class="font-semibold">{{ formatPrice(item.base_price) }} UZS</span>
                  <Button variant="destructive" size="sm" @click="deleteItem(item.id)">Delete</Button>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
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

    <Sheet :open="showItemForm" @update:open="(val) => !val && (showItemForm = false)">
      <SheetContent side="right" class="w-full overflow-y-auto sm:max-w-[34rem]">
        <SheetHeader>
          <SheetTitle>Add Item</SheetTitle>
          <SheetDescription>New item for {{ selectedCategory?.name }}</SheetDescription>
        </SheetHeader>
        <form class="mt-6 space-y-4" @submit.prevent="addItem">
          <div class="space-y-1.5">
            <Label for="item-name">Item Name</Label>
            <Input id="item-name" v-model="newItem.name" required />
          </div>
          <div class="space-y-1.5">
            <Label for="item-desc">Description</Label>
            <Textarea id="item-desc" v-model="newItem.description" />
          </div>
          <div class="space-y-1.5">
            <Label for="item-price">Base Price</Label>
            <Input id="item-price" v-model.number="newItem.base_price" type="number" required />
          </div>
          <SheetFooter>
            <Button type="button" variant="outline" @click="showItemForm = false">Cancel</Button>
            <Button type="submit">Create Item</Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  </div>
</template>

<script setup lang="ts">
import { UtensilsCrossed } from 'lucide-vue-next'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from '@/components/ui/sheet'

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
    body: { ...newCategory, branch_id: branchContext.selectedBranchId.value },
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
