<template>
  <header class="sticky top-0 z-10 flex h-16 shrink-0 items-center border-b bg-background/90 backdrop-blur">
    <div class="flex flex-1 items-center justify-between gap-4 px-4 lg:px-6">
      <div class="flex items-center gap-3">
        <SidebarTrigger class="-ml-1" />
        <Separator orientation="vertical" class="h-4" />
        <div>
          <p class="text-base font-semibold leading-tight">{{ title }}</p>
          <p class="hidden text-xs text-muted-foreground sm:block">{{ subtitle }}</p>
        </div>
      </div>
      <Badge variant="secondary" class="hidden md:inline-flex font-medium">
        {{ branchContext.selectedBranchLabel.value }}
      </Badge>
    </div>
  </header>
</template>

<script setup lang="ts">
import { SidebarTrigger } from '@/components/ui/sidebar'
import { Separator } from '@/components/ui/separator'
import { Badge } from '@/components/ui/badge'

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
