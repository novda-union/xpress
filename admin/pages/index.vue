<template>
  <div>
    <h2 class="text-2xl font-bold mb-6">Dashboard</h2>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
      <div class="bg-white p-6 rounded-lg shadow">
        <p class="text-sm text-gray-500">Today's Orders</p>
        <p class="text-3xl font-bold">{{ stats.totalOrders }}</p>
      </div>
      <div class="bg-white p-6 rounded-lg shadow">
        <p class="text-sm text-gray-500">Pending</p>
        <p class="text-3xl font-bold text-yellow-600">{{ stats.pendingOrders }}</p>
      </div>
      <div class="bg-white p-6 rounded-lg shadow">
        <p class="text-sm text-gray-500">Today's Revenue</p>
        <p class="text-3xl font-bold text-green-600">{{ formatPrice(stats.revenue) }} UZS</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const { api } = useApi()

const stats = reactive({ totalOrders: 0, pendingOrders: 0, revenue: 0 })

onMounted(async () => {
  try {
    const orders = await api<any[]>('/admin/orders')
    stats.totalOrders = orders.length
    stats.pendingOrders = orders.filter((o: any) => o.status === 'pending').length
    stats.revenue = orders
      .filter((o: any) => !['cancelled', 'rejected'].includes(o.status))
      .reduce((sum: number, o: any) => sum + o.total_price, 0)
  } catch (e) {
    console.error('Failed to load stats', e)
  }
})

function formatPrice(price: number) {
  return price.toLocaleString('en-US')
}
</script>
