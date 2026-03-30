<template>
  <aside class="w-full border-b border-[var(--admin-border)] bg-[var(--admin-surface)] md:flex md:w-60 md:flex-col md:border-b-0 md:border-r xl:w-64">
    <div class="flex h-16 items-center gap-3 px-5 md:flex-none">
      <div class="flex h-10 w-10 items-center justify-center rounded-2xl bg-[var(--admin-accent-bg)] text-[var(--admin-accent)]">
        <Store class="h-5 w-5" />
      </div>
      <div>
        <p class="text-sm font-semibold text-[var(--admin-text)]">Xpressgo</p>
        <p class="text-xs text-[var(--admin-text-muted)]">Admin Console</p>
      </div>
    </div>

    <div v-if="branchContext.isDirector.value" class="px-4 pb-4 md:flex-none">
      <label class="label">Branch Context</label>
      <select class="select" :value="branchContext.selectedBranchId.value ?? ''" @change="onBranchChange">
        <option value="">All Branches</option>
        <option v-for="branch in branchContext.branches.value" :key="branch.id" :value="branch.id">
          {{ branch.name }}
        </option>
      </select>
    </div>

    <nav class="flex gap-2 overflow-x-auto px-3 pb-4 md:flex-1 md:flex-col md:space-y-1 md:gap-0 md:overflow-visible md:pb-0">
      <NuxtLink
        v-for="item in visibleItems"
        :key="item.to"
        :to="item.to"
        class="flex h-10 shrink-0 items-center gap-3 rounded-xl px-4 text-sm font-medium text-[var(--admin-text-muted)] transition hover:bg-[var(--admin-surface-2)] hover:text-[var(--admin-text)]"
        active-class="bg-[var(--admin-accent-bg)] !text-[var(--admin-accent)]"
      >
        <component :is="item.icon" class="h-5 w-5" />
        <span>{{ item.label }}</span>
      </NuxtLink>
    </nav>

    <div class="m-3 hidden rounded-2xl border border-[var(--admin-border)] bg-[var(--admin-surface-2)] p-4 md:block">
      <div class="flex items-center gap-3">
        <div class="flex h-10 w-10 items-center justify-center rounded-full bg-[var(--admin-accent-bg)] text-[var(--admin-accent)]">
          <User class="h-5 w-5" />
        </div>
        <div class="min-w-0">
          <p class="truncate text-sm font-semibold">{{ auth.state.staff?.name }}</p>
          <p class="text-xs text-[var(--admin-text-muted)]">{{ roleLabel }}</p>
        </div>
      </div>
      <button class="btn-ghost mt-3 w-full justify-start px-0" @click="auth.logout">
        <LogOut class="h-4 w-4" />
        Logout
      </button>
    </div>
  </aside>
</template>

<script setup lang="ts">
import {
  ClipboardList,
  LayoutDashboard,
  LogOut,
  Settings,
  Store,
  User,
  Users,
  UtensilsCrossed,
} from 'lucide-vue-next'

const auth = useAuth()
const permissions = usePermissions()
const branchContext = useBranchContext()

const items = [
  { to: '/', label: 'Dashboard', icon: LayoutDashboard },
  { to: '/orders', label: 'Orders', icon: ClipboardList },
  { to: '/menu', label: 'Menu', icon: UtensilsCrossed },
  { to: '/branches', label: 'Branches', icon: Store },
  { to: '/staff', label: 'Staff', icon: Users },
  { to: '/settings', label: 'Store Settings', icon: Settings },
  { to: '/settings/branch', label: 'Branch Settings', icon: Settings },
]

const visibleItems = computed(() => items.filter((item) => permissions.canVisit(item.to)))
const roleLabel = computed(() => {
  const role = auth.state.staff?.role
  if (!role) return ''
  return role.charAt(0).toUpperCase() + role.slice(1)
})

function onBranchChange(event: Event) {
  const value = (event.target as HTMLSelectElement).value
  branchContext.selectBranch(value || null)
}
</script>
