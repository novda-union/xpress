<template>
  <header class="sticky top-0 z-10 border-b border-[var(--admin-border)] bg-white/90 backdrop-blur">
    <div class="flex min-h-16 items-center justify-between gap-4 px-4 py-3 lg:px-6">
      <div>
        <p class="page-title">{{ title }}</p>
        <p class="page-subtitle">{{ subtitle }}</p>
      </div>
      <div class="flex items-center gap-3">
        <div class="hidden rounded-full bg-[var(--admin-accent-bg)] px-3 py-2 text-xs font-semibold text-[var(--admin-accent)] md:block">
          {{ branchContext.selectedBranchLabel.value }}
        </div>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
const route = useRoute()
const branchContext = useBranchContext()

const labels: Record<string, { title: string; subtitle: string }> = {
  '/': { title: 'Dashboard', subtitle: 'Operational overview across your active branch context.' },
  '/orders': { title: 'Orders', subtitle: 'Live order board grouped by service stage.' },
  '/menu': { title: 'Menu', subtitle: 'Branch-aware categories and items.' },
  '/branches': { title: 'Branches', subtitle: 'Manage locations, status, and staffing footprint.' },
  '/staff': { title: 'Staff', subtitle: 'Role and branch assignments for the active store.' },
  '/settings': { title: 'Store Settings', subtitle: 'Global store identity and public information.' },
  '/settings/branch': { title: 'Branch Settings', subtitle: 'Edit the selected branch details and contact data.' },
}

const meta = computed(() => {
  const entry = Object.entries(labels).find(([path]) => route.path === path || route.path.startsWith(`${path}/`))
  return entry?.[1] ?? { title: 'Xpressgo Admin', subtitle: 'Branch-aware operations' }
})

const title = computed(() => meta.value.title)
const subtitle = computed(() => meta.value.subtitle)
</script>
