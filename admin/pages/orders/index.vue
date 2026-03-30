<template>
  <div class="space-y-6">
    <div class="grid gap-4 xl:grid-cols-3">
      <section class="space-y-3 rounded-xl border bg-muted/30 p-4 border-t-4 border-t-blue-500">
        <div class="flex items-center justify-between">
          <h3 class="font-semibold text-blue-700">New Orders</h3>
          <Badge class="bg-blue-100 text-blue-700 hover:bg-blue-100">{{ newOrders.length }}</Badge>
        </div>
        <div class="space-y-3">
          <OrderCard
            v-for="order in newOrders"
            :key="order.id"
            :order="order"
            @accept="updateStatus(order.id, order.status === 'pending' ? 'accepted' : 'preparing')"
            @reject="rejectOrder(order.id)"
          />
          <p v-if="!newOrders.length" class="text-sm text-muted-foreground">No new orders.</p>
        </div>
      </section>

      <section class="space-y-3 rounded-xl border bg-muted/30 p-4 border-t-4 border-t-amber-500">
        <div class="flex items-center justify-between">
          <h3 class="font-semibold text-amber-700">Preparing</h3>
          <Badge class="bg-amber-100 text-amber-700 hover:bg-amber-100">{{ preparingOrders.length }}</Badge>
        </div>
        <div class="space-y-3">
          <OrderCard v-for="order in preparingOrders" :key="order.id" :order="order" @mark-ready="updateStatus(order.id, 'ready')" />
          <p v-if="!preparingOrders.length" class="text-sm text-muted-foreground">Nothing in preparation.</p>
        </div>
      </section>

      <section class="space-y-3 rounded-xl border bg-muted/30 p-4 border-t-4 border-t-green-500">
        <div class="flex items-center justify-between">
          <h3 class="font-semibold text-green-700">Ready</h3>
          <Badge class="bg-green-100 text-green-700 hover:bg-green-100">{{ readyOrders.length }}</Badge>
        </div>
        <div class="space-y-3">
          <OrderCard v-for="order in readyOrders" :key="order.id" :order="order" @picked-up="updateStatus(order.id, 'picked_up')" />
          <p v-if="!readyOrders.length" class="text-sm text-muted-foreground">No ready orders.</p>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Badge } from '@/components/ui/badge'
import type { AdminOrder } from 'types/auth'

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
