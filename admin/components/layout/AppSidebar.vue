<template>
  <Sidebar collapsible="offcanvas">
    <SidebarHeader>
      <div class="flex items-center gap-3 px-2 py-1">
        <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
          <Store class="h-4 w-4" />
        </div>
        <div class="min-w-0">
          <p class="truncate text-sm font-semibold">Xpressgo</p>
          <p class="truncate text-xs text-muted-foreground">Admin Console</p>
        </div>
      </div>
    </SidebarHeader>

    <SidebarContent>
      <SidebarGroup v-if="branchContext.isDirector.value" class="pb-0">
        <SidebarGroupLabel>Branch Context</SidebarGroupLabel>
        <SidebarGroupContent class="px-1">
          <Select :model-value="branchContext.selectedBranchId.value ?? ''" @update:model-value="onBranchChange">
            <SelectTrigger class="w-full">
              <SelectValue placeholder="All Branches" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="">All Branches</SelectItem>
              <SelectItem
                v-for="branch in branchContext.branches.value"
                :key="branch.id"
                :value="branch.id"
              >
                {{ branch.name }}
              </SelectItem>
            </SelectContent>
          </Select>
        </SidebarGroupContent>
      </SidebarGroup>

      <SidebarSeparator v-if="branchContext.isDirector.value" />

      <SidebarGroup>
        <SidebarGroupLabel>Navigation</SidebarGroupLabel>
        <SidebarGroupContent>
          <SidebarMenu>
            <SidebarMenuItem v-for="item in visibleItems" :key="item.to">
              <SidebarMenuButton as-child :is-active="isActive(item.to)" :tooltip="item.label">
                <NuxtLink :to="item.to">
                  <component :is="item.icon" />
                  <span>{{ item.label }}</span>
                </NuxtLink>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarGroupContent>
      </SidebarGroup>
    </SidebarContent>

    <SidebarFooter>
      <div class="flex items-center gap-3 rounded-lg p-2 hover:bg-sidebar-accent transition-colors">
        <Avatar class="h-8 w-8 shrink-0">
          <AvatarFallback class="bg-primary/10 text-primary text-xs font-semibold">
            {{ initials }}
          </AvatarFallback>
        </Avatar>
        <div class="min-w-0 flex-1">
          <p class="truncate text-sm font-semibold">{{ auth.state.staff?.name }}</p>
          <p class="truncate text-xs text-muted-foreground">{{ roleLabel }}</p>
        </div>
        <Button variant="ghost" size="icon" class="h-8 w-8 shrink-0 text-muted-foreground" @click="auth.logout">
          <LogOut class="h-4 w-4" />
          <span class="sr-only">Logout</span>
        </Button>
      </div>
    </SidebarFooter>

    <SidebarRail />
  </Sidebar>
</template>

<script setup lang="ts">
import type { AcceptableValue } from 'reka-ui'
import {
  ClipboardList,
  LayoutDashboard,
  LogOut,
  Settings,
  Store,
  Users,
  UtensilsCrossed,
} from 'lucide-vue-next'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
  SidebarSeparator,
} from '@/components/ui/sidebar'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'

const route = useRoute()
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

const initials = computed(() => {
  const name = auth.state.staff?.name ?? ''
  return name.split(' ').map((n) => n[0]).join('').toUpperCase().slice(0, 2)
})

function isActive(to: string) {
  if (to === '/') return route.path === '/'
  return route.path.startsWith(to)
}

function onBranchChange(value: AcceptableValue) {
  branchContext.selectBranch(typeof value === 'string' && value.length > 0 ? value : null)
}
</script>
