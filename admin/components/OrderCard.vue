<template>
  <div class="bg-white rounded-lg shadow p-4">
    <div class="flex justify-between items-start mb-2">
      <span class="font-bold text-lg">#{{ order.order_number }}</span>
      <span class="text-xs px-2 py-1 rounded" :class="statusClass">{{ order.status }}</span>
    </div>

    <p class="text-sm text-gray-500 mb-2">
      ETA: ~{{ order.eta_minutes }} min | {{ formatTime(order.created_at) }}
    </p>

    <div class="text-sm mb-3 space-y-1">
      <div v-for="item in order.items" :key="item.id">
        <span class="font-medium">{{ item.quantity }}x {{ item.item_name }}</span>
        <span v-if="item.modifiers?.length" class="text-gray-500">
          ({{ item.modifiers.map((m: any) => m.modifier_name).join(', ') }})
        </span>
      </div>
    </div>

    <p class="font-bold mb-3">{{ formatPrice(order.total_price) }} UZS</p>

    <div class="flex gap-2">
      <button
        v-if="order.status === 'pending'"
        @click="$emit('accept')"
        class="flex-1 bg-green-600 text-white py-1.5 px-3 rounded text-sm hover:bg-green-700"
      >
        Accept
      </button>
      <button
        v-if="order.status === 'pending'"
        @click="$emit('reject')"
        class="flex-1 bg-red-600 text-white py-1.5 px-3 rounded text-sm hover:bg-red-700"
      >
        Reject
      </button>
      <button
        v-if="order.status === 'accepted'"
        @click="$emit('accept')"
        class="flex-1 bg-blue-600 text-white py-1.5 px-3 rounded text-sm hover:bg-blue-700"
      >
        Start Preparing
      </button>
      <button
        v-if="order.status === 'preparing'"
        @click="$emit('mark-ready')"
        class="flex-1 bg-green-600 text-white py-1.5 px-3 rounded text-sm hover:bg-green-700"
      >
        Mark Ready
      </button>
      <button
        v-if="order.status === 'ready'"
        @click="$emit('picked-up')"
        class="flex-1 bg-gray-600 text-white py-1.5 px-3 rounded text-sm hover:bg-gray-700"
      >
        Picked Up
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{ order: any }>()
defineEmits(['accept', 'reject', 'mark-ready', 'picked-up'])

const statusClass = computed(() => {
  const cls: Record<string, string> = {
    pending: 'bg-yellow-100 text-yellow-800',
    accepted: 'bg-blue-100 text-blue-800',
    preparing: 'bg-blue-100 text-blue-800',
    ready: 'bg-green-100 text-green-800',
    picked_up: 'bg-gray-100 text-gray-800',
    rejected: 'bg-red-100 text-red-800',
    cancelled: 'bg-gray-100 text-gray-800',
  }
  return cls[props.order?.status] || ''
})

function formatPrice(price: number) {
  return price?.toLocaleString('en-US') || '0'
}

function formatTime(dateStr: string) {
  return new Date(dateStr).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })
}
</script>
