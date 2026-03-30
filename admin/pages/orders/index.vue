<template>
  <div class="space-y-6">
    <div class="grid gap-4 xl:grid-cols-3">
      <section class="kanban-column border-t-2 border-t-[var(--admin-new)]">
        <div class="mb-4 flex items-center justify-between">
          <h3 class="text-lg font-semibold text-[var(--admin-new)]">New Orders</h3>
          <span class="badge bg-[var(--admin-new-bg)] text-[var(--admin-new)]">{{ newOrders.length }}</span>
        </div>
        <div class="space-y-3">
          <OrderCard
            v-for="order in newOrders"
            :key="order.id"
            :order="order"
            @accept="updateStatus(order.id, order.status === 'pending' ? 'accepted' : 'preparing')"
            @reject="rejectOrder(order.id)"
          />
          <p v-if="!newOrders.length" class="text-sm text-[var(--admin-text-muted)]">No new orders.</p>
        </div>
      </section>

      <section class="kanban-column border-t-2 border-t-[var(--admin-preparing)]">
        <div class="mb-4 flex items-center justify-between">
          <h3 class="text-lg font-semibold text-[var(--admin-preparing)]">Preparing</h3>
          <span class="badge bg-[var(--admin-preparing-bg)] text-[var(--admin-preparing)]">{{ preparingOrders.length }}</span>
        </div>
        <div class="space-y-3">
          <OrderCard v-for="order in preparingOrders" :key="order.id" :order="order" @mark-ready="updateStatus(order.id, 'ready')" />
          <p v-if="!preparingOrders.length" class="text-sm text-[var(--admin-text-muted)]">Nothing in preparation.</p>
        </div>
      </section>

      <section class="kanban-column border-t-2 border-t-[var(--admin-ready)]">
        <div class="mb-4 flex items-center justify-between">
          <h3 class="text-lg font-semibold text-[var(--admin-ready)]">Ready</h3>
          <span class="badge bg-[var(--admin-ready-bg)] text-[var(--admin-ready)]">{{ readyOrders.length }}</span>
        </div>
        <div class="space-y-3">
          <OrderCard v-for="order in readyOrders" :key="order.id" :order="order" @picked-up="updateStatus(order.id, 'picked_up')" />
          <p v-if="!readyOrders.length" class="text-sm text-[var(--admin-text-muted)]">No ready orders.</p>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { AdminOrder } from '~/types/auth'

const { api } = useApi()
const { connect, on } = useAdminWebSocket()
const branchContext = useBranchContext()
const auth = useAuth()

const orders = ref<AdminOrder[]>([])

const scopedOrders = computed(() => {
  if (auth.state.staff?.role === 'director' && branchContext.selectedBranchId.value) {
    return orders.value.filter((order) => order.branch_id === branchContext.selectedBranchId.value)
  }
  return orders.value
})

const newOrders = computed(() => scopedOrders.value.filter((order) => ['pending', 'accepted'].includes(order.status)))
const preparingOrders = computed(() => scopedOrders.value.filter((order) => order.status === 'preparing'))
const readyOrders = computed(() => scopedOrders.value.filter((order) => order.status === 'ready'))

async function loadOrders() {
  orders.value = await api<AdminOrder[]>('/admin/orders')
}

async function updateStatus(orderId: string, status: string, reason = '') {
  await api(`/admin/orders/${orderId}/status`, {
    method: 'PUT',
    body: { status, reason },
  })
  await loadOrders()
}

function rejectOrder(orderId: string) {
  const reason = window.prompt('Rejection reason:')
  if (reason !== null) {
    updateStatus(orderId, 'rejected', reason)
  }
}

onMounted(async () => {
  await loadOrders()
  connect()
  on('order:new', loadOrders)
  on('order:cancelled', loadOrders)
  on('order:status', loadOrders)
  on('order:rejected', loadOrders)
})

watch(() => branchContext.selectedBranchId.value, loadOrders)
</script>
