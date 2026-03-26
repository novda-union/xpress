<template>
  <div>
    <h2 class="text-2xl font-bold mb-6">Orders</h2>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
      <!-- New Orders -->
      <div>
        <h3 class="text-lg font-semibold mb-3 text-yellow-600">New</h3>
        <div class="space-y-3">
          <OrderCard
            v-for="order in newOrders"
            :key="order.id"
            :order="order"
            @accept="updateStatus(order.id, 'accepted')"
            @reject="rejectOrder(order.id)"
          />
          <p v-if="!newOrders.length" class="text-gray-400 text-sm">No new orders</p>
        </div>
      </div>

      <!-- Preparing -->
      <div>
        <h3 class="text-lg font-semibold mb-3 text-blue-600">Preparing</h3>
        <div class="space-y-3">
          <OrderCard
            v-for="order in preparingOrders"
            :key="order.id"
            :order="order"
            @mark-ready="updateStatus(order.id, 'ready')"
          />
          <p v-if="!preparingOrders.length" class="text-gray-400 text-sm">None preparing</p>
        </div>
      </div>

      <!-- Ready -->
      <div>
        <h3 class="text-lg font-semibold mb-3 text-green-600">Ready</h3>
        <div class="space-y-3">
          <OrderCard
            v-for="order in readyOrders"
            :key="order.id"
            :order="order"
            @picked-up="updateStatus(order.id, 'picked_up')"
          />
          <p v-if="!readyOrders.length" class="text-gray-400 text-sm">None ready</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const { api } = useApi()
const { connect, on } = useAdminWebSocket()

const orders = ref<any[]>([])

const newOrders = computed(() => orders.value.filter(o => ['pending', 'accepted'].includes(o.status)))
const preparingOrders = computed(() => orders.value.filter(o => o.status === 'preparing'))
const readyOrders = computed(() => orders.value.filter(o => o.status === 'ready'))

async function loadOrders() {
  try {
    orders.value = await api<any[]>('/admin/orders')
  } catch (e) {
    console.error('Failed to load orders', e)
  }
}

async function updateStatus(orderId: string, status: string, reason = '') {
  try {
    await api(`/admin/orders/${orderId}/status`, {
      method: 'PUT',
      body: { status, reason },
    })
    await loadOrders()
  } catch (e: any) {
    alert(e?.data?.error || 'Failed to update status')
  }
}

function rejectOrder(orderId: string) {
  const reason = prompt('Rejection reason:')
  if (reason !== null) {
    updateStatus(orderId, 'rejected', reason)
  }
}

onMounted(() => {
  loadOrders()
  connect()

  on('order:new', () => {
    loadOrders()
    // Play sound alert
    try {
      new Audio('/notification.mp3').play()
    } catch {}
  })

  on('order:cancelled', () => loadOrders())
})
</script>
