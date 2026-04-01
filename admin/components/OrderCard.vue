<template>
  <Card class="border-l-4 transition-shadow hover:shadow-md" :class="borderClass">
    <CardContent class="p-4">
      <div class="mb-3 flex items-start justify-between gap-3">
        <div>
          <span class="text-lg font-bold">#{{ order.order_number }}</span>
          <p class="mt-0.5 text-xs text-muted-foreground">
            {{ formatTime(order.created_at) }} · {{ order.items?.length ?? 0 }} item(s)
          </p>
        </div>
        <Badge :class="statusClass" variant="secondary">{{ order.status.replace('_', ' ') }}</Badge>
      </div>

      <p class="mb-3 text-sm text-muted-foreground">
        ETA ~{{ order.eta_minutes }} min · Branch {{ order.branch_id.slice(0, 8) }} · {{ formatPayment(order.payment_method) }}
      </p>

      <div class="mb-4 space-y-1 text-sm">
        <div v-for="item in (order.items ?? [])" :key="item.id">
          <span class="font-medium">{{ item.quantity }}x {{ item.item_name }}</span>
          <span v-if="item.modifiers?.length" class="text-muted-foreground">
            ({{ item.modifiers.map((m: any) => m.modifier_name).join(', ') }})
          </span>
        </div>
      </div>

      <p class="mb-4 font-bold">{{ formatPrice(order.total_price) }} UZS</p>

      <div class="flex gap-2">
        <Button v-if="order.status === 'pending'" class="flex-1" size="sm" @click="$emit('accept')">
          Accept
        </Button>
        <Button v-if="order.status === 'pending'" class="flex-1" size="sm" variant="destructive" @click="$emit('reject')">
          Reject
        </Button>
        <Button v-if="order.status === 'accepted'" class="flex-1" size="sm" @click="$emit('accept')">
          Start Preparing
        </Button>
        <Button v-if="order.status === 'preparing'" class="flex-1" size="sm" @click="$emit('mark-ready')">
          Mark Ready
        </Button>
        <Button v-if="order.status === 'ready'" class="flex-1" size="sm" variant="outline" @click="$emit('picked-up')">
          Picked Up
        </Button>
      </div>
    </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import type { AdminOrder } from 'types/auth'

const props = defineProps<{ order: AdminOrder }>()
defineEmits(['accept', 'reject', 'mark-ready', 'picked-up'])

const statusClass = computed(() => {
  const cls: Record<string, string> = {
    pending: 'bg-blue-100 text-blue-700 hover:bg-blue-100',
    accepted: 'bg-blue-100 text-blue-700 hover:bg-blue-100',
    preparing: 'bg-amber-100 text-amber-700 hover:bg-amber-100',
    ready: 'bg-green-100 text-green-700 hover:bg-green-100',
    picked_up: 'bg-muted text-muted-foreground hover:bg-muted',
    rejected: 'bg-red-100 text-red-700 hover:bg-red-100',
    cancelled: 'bg-muted text-muted-foreground hover:bg-muted',
  }
  return cls[props.order?.status] || ''
})

const borderClass = computed(() => {
  const cls: Record<string, string> = {
    pending: 'border-l-blue-500',
    accepted: 'border-l-blue-500',
    preparing: 'border-l-amber-500',
    ready: 'border-l-green-500',
  }
  return cls[props.order?.status] || 'border-l-transparent'
})

function formatPrice(price: number) {
  return price?.toLocaleString('en-US') || '0'
}

function formatTime(dateStr: string) {
  return new Date(dateStr).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })
}

function formatPayment(method: string) {
  if (method === 'cash') {
    return 'Cash'
  }
  if (method === 'card') {
    return 'Card'
  }
  return method
    .split(/[\s_-]+/)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1).toLowerCase())
    .join(' ')
}
</script>
