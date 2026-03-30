<template>
  <div class="surface-card border-l-4 p-4" :class="borderClass">
    <div class="mb-3 flex items-start justify-between gap-3">
      <div>
        <span class="text-lg font-bold">#{{ order.order_number }}</span>
        <p class="mt-1 text-xs text-[var(--admin-text-muted)]">
          {{ formatTime(order.created_at) }} · {{ order.items.length }} item(s)
        </p>
      </div>
      <span class="badge" :class="statusClass">{{ order.status.replace('_', ' ') }}</span>
    </div>

    <p class="mb-3 text-sm text-[var(--admin-text-muted)]">
      ETA ~{{ order.eta_minutes }} min · Branch {{ order.branch_id.slice(0, 8) }}
    </p>

    <div class="mb-4 space-y-1 text-sm">
      <div v-for="item in order.items" :key="item.id">
        <span class="font-medium">{{ item.quantity }}x {{ item.item_name }}</span>
        <span v-if="item.modifiers?.length" class="text-[var(--admin-text-muted)]">
          ({{ item.modifiers.map((m: any) => m.modifier_name).join(', ') }})
        </span>
      </div>
    </div>

    <p class="mb-4 font-bold">{{ formatPrice(order.total_price) }} UZS</p>

    <div class="flex gap-2">
      <button
        v-if="order.status === 'pending'"
        @click="$emit('accept')"
        class="btn-primary flex-1"
      >
        Accept
      </button>
      <button
        v-if="order.status === 'pending'"
        @click="$emit('reject')"
        class="btn-danger flex-1"
      >
        Reject
      </button>
      <button
        v-if="order.status === 'accepted'"
        @click="$emit('accept')"
        class="btn-primary flex-1"
      >
        Start Preparing
      </button>
      <button
        v-if="order.status === 'preparing'"
        @click="$emit('mark-ready')"
        class="btn-primary flex-1"
      >
        Mark Ready
      </button>
      <button
        v-if="order.status === 'ready'"
        @click="$emit('picked-up')"
        class="btn-secondary flex-1"
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
    pending: 'bg-[var(--admin-new-bg)] text-[var(--admin-new)]',
    accepted: 'bg-[var(--admin-new-bg)] text-[var(--admin-new)]',
    preparing: 'bg-[var(--admin-preparing-bg)] text-[var(--admin-preparing)]',
    ready: 'bg-[var(--admin-ready-bg)] text-[var(--admin-ready)]',
    picked_up: 'status-inactive',
    rejected: 'bg-red-50 text-[var(--admin-error)]',
    cancelled: 'status-inactive',
  }
  return cls[props.order?.status] || ''
})

const borderClass = computed(() => {
  const cls: Record<string, string> = {
    pending: 'border-l-[var(--admin-new)]',
    accepted: 'border-l-[var(--admin-new)]',
    preparing: 'border-l-[var(--admin-preparing)]',
    ready: 'border-l-[var(--admin-ready)]',
  }
  return cls[props.order?.status] || 'border-l-transparent'
})

function formatPrice(price: number) {
  return price?.toLocaleString('en-US') || '0'
}

function formatTime(dateStr: string) {
  return new Date(dateStr).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })
}
</script>
