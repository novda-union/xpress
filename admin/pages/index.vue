<template>
  <div class="space-y-6">
    <section class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <StatCard :icon="ClipboardList" label="Today's Orders" :value="stats.totalOrders" />
      <StatCard :icon="Clock3" label="Pending Now" :value="stats.pendingOrders" hint="Waiting for action" hint-class="text-[var(--admin-warning)]" />
      <StatCard :icon="BadgeDollarSign" label="Revenue" :value="`${formatPrice(stats.revenue)} UZS`" />
      <StatCard :icon="Store" label="Active Branches" :value="stats.activeBranches" />
    </section>

    <section v-if="branchCards.length" class="grid gap-4 xl:grid-cols-3">
      <div v-for="branch in branchCards" :key="branch.id" class="surface-card p-5">
        <div class="flex items-start justify-between gap-3">
          <div>
            <p class="text-lg font-semibold">{{ branch.name }}</p>
            <p class="mt-1 text-sm text-[var(--admin-text-muted)]">{{ branch.address }}</p>
          </div>
          <span class="badge badge-dot" :class="branch.is_active ? 'status-active' : 'status-inactive'">
            {{ branch.is_active ? 'Active' : 'Inactive' }}
          </span>
        </div>
        <div class="mt-4 grid grid-cols-2 gap-3">
          <div class="surface-muted px-3 py-3">
            <p class="text-xs uppercase tracking-wide text-[var(--admin-text-muted)]">Orders</p>
            <p class="mt-1 text-xl font-semibold">{{ branch.orderCount }}</p>
          </div>
          <div class="surface-muted px-3 py-3">
            <p class="text-xs uppercase tracking-wide text-[var(--admin-text-muted)]">Revenue</p>
            <p class="mt-1 text-xl font-semibold">{{ formatPrice(branch.revenue) }}</p>
          </div>
        </div>
      </div>
    </section>

    <EmptyState
      v-else
      :icon="LayoutDashboard"
      title="No branch data available"
      description="Once orders start arriving, branch performance cards will appear here."
    />
  </div>
</template>

<script setup lang="ts">
import { BadgeDollarSign, ClipboardList, Clock3, LayoutDashboard, Store } from 'lucide-vue-next'
import type { AdminOrder, BranchSummary } from '~/types/auth'

const { api } = useApi()
const branchContext = useBranchContext()
const auth = useAuth()

const orders = ref<AdminOrder[]>([])
const branches = ref<BranchSummary[]>([])

const filteredOrders = computed(() => {
  if (auth.state.staff?.role !== 'director' || !branchContext.selectedBranchId.value) return orders.value
  return orders.value.filter((order) => order.branch_id === branchContext.selectedBranchId.value)
})

const stats = computed(() => ({
  activeBranches: branches.value.filter((branch) => branch.is_active).length,
  pendingOrders: filteredOrders.value.filter((order) => order.status === 'pending').length,
  revenue: filteredOrders.value
    .filter((order) => !['cancelled', 'rejected'].includes(order.status))
    .reduce((sum, order) => sum + order.total_price, 0),
  totalOrders: filteredOrders.value.length,
}))

const branchCards = computed(() =>
  branches.value
    .filter((branch) => auth.state.staff?.role !== 'director' || !branchContext.selectedBranchId.value || branch.id === branchContext.selectedBranchId.value)
    .map((branch) => ({
      ...branch,
      orderCount: orders.value.filter((order) => order.branch_id === branch.id).length,
      revenue: orders.value
        .filter((order) => order.branch_id === branch.id && !['cancelled', 'rejected'].includes(order.status))
        .reduce((sum, order) => sum + order.total_price, 0),
    })),
)

async function loadData() {
  branches.value = await api<BranchSummary[]>('/admin/branches')
  orders.value = await api<AdminOrder[]>('/admin/orders')
}

function formatPrice(price: number) {
  return price.toLocaleString('en-US')
}

onMounted(async () => {
  branchContext.init()
  await branchContext.loadBranches()
  await loadData()
})

watch(() => branchContext.selectedBranchId.value, loadData)
</script>
