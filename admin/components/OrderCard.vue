<template>
  <Card class="border-l-4 transition-shadow hover:shadow-md" :class="borderClass">
    <CardContent class="p-3">
      <div class="mb-2 flex items-start justify-between gap-2">
        <div>
          <span class="text-base font-bold">#{{ order.order_number }}</span>
          <p class="mt-0.5 text-xs text-muted-foreground">
            {{ formatTime(order.created_at) }} · {{ order.items?.length ?? 0 }} item(s)
          </p>
        </div>
        <Badge :class="statusClass" variant="secondary">{{ order.status.replace('_', ' ') }}</Badge>
      </div>

      <p class="mb-2 text-xs text-muted-foreground">
        ETA ~{{ order.eta_minutes }} min · Branch {{ order.branch_id.slice(0, 8) }} · {{ formatPayment(order.payment_method) }}
      </p>

      <div class="mb-3 space-y-1 text-sm">
        <div v-for="item in (order.items ?? [])" :key="item.id">
          <span class="font-medium">{{ item.quantity }}x {{ item.item_name }}</span>
          <span v-if="item.modifiers?.length" class="text-muted-foreground">
            ({{ item.modifiers.map((m: any) => m.modifier_name).join(', ') }})
          </span>
        </div>
      </div>

      <p class="mb-3 text-sm font-bold">{{ formatPrice(order.total_price) }} UZS</p>

      <div class="flex gap-2">
        <Button
          v-if="order.status === 'pending'"
          class="h-9 w-9"
          size="icon"
          :class="neutralDisabledClass"
          :disabled="loading"
          :title="actionIsLoading('accept') ? 'Loading' : 'Accept order'"
          @click="$emit('accept')"
        >
          <LoaderCircle v-if="actionIsLoading('accept')" class="h-4 w-4 animate-spin" />
          <Check v-else class="h-4 w-4" />
          <span class="sr-only">Accept order</span>
        </Button>

        <Button
          v-if="order.status === 'pending'"
          class="h-9 w-9"
          size="icon"
          variant="destructive"
          :class="neutralDisabledClass"
          :disabled="loading"
          :title="actionIsLoading('reject') ? 'Loading' : 'Reject order'"
          @click="$emit('reject')"
        >
          <LoaderCircle v-if="actionIsLoading('reject')" class="h-4 w-4 animate-spin" />
          <X v-else class="h-4 w-4" />
          <span class="sr-only">Reject order</span>
        </Button>

        <Button
          v-if="order.status === 'accepted'"
          class="h-9 w-9"
          size="icon"
          :class="neutralDisabledClass"
          :disabled="loading"
          :title="actionIsLoading('accept') ? 'Loading' : 'Start preparing'"
          @click="$emit('accept')"
        >
          <LoaderCircle v-if="actionIsLoading('accept')" class="h-4 w-4 animate-spin" />
          <ChefHat v-else class="h-4 w-4" />
          <span class="sr-only">Start preparing</span>
        </Button>

        <Button
          v-if="order.status === 'preparing'"
          class="h-9 w-9"
          size="icon"
          :class="neutralDisabledClass"
          :disabled="loading"
          :title="actionIsLoading('mark-ready') ? 'Loading' : 'Mark ready'"
          @click="$emit('mark-ready')"
        >
          <LoaderCircle v-if="actionIsLoading('mark-ready')" class="h-4 w-4 animate-spin" />
          <PackageCheck v-else class="h-4 w-4" />
          <span class="sr-only">Mark ready</span>
        </Button>

        <Button
          v-if="order.status === 'ready'"
          class="h-9 w-9"
          size="icon"
          variant="outline"
          :class="neutralDisabledClass"
          :disabled="loading"
          :title="actionIsLoading('picked-up') ? 'Loading' : 'Mark picked up'"
          @click="$emit('picked-up')"
        >
          <LoaderCircle v-if="actionIsLoading('picked-up')" class="h-4 w-4 animate-spin" />
          <Check v-else class="h-4 w-4" />
          <span class="sr-only">Mark picked up</span>
        </Button>
      </div>
    </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { Check, ChefHat, LoaderCircle, PackageCheck, X } from 'lucide-vue-next'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import type { AdminOrder } from 'types/auth'

const props = defineProps<{
  order: AdminOrder
  loading?: boolean
  loadingAction?: string | null
}>()
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

function actionIsLoading(action: string) {
  return props.loading && props.loadingAction === action
}

const neutralDisabledClass = 'disabled:border-muted disabled:bg-muted disabled:text-muted-foreground disabled:opacity-100'

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
