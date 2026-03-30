<template>
  <Card>
    <CardContent class="p-5">
      <div class="flex items-start justify-between gap-3">
        <div>
          <p class="text-sm font-medium text-muted-foreground">{{ label }}</p>
          <p class="mt-3 text-[2rem] font-bold leading-none">{{ value }}</p>
          <p v-if="hint" class="mt-3 text-xs font-semibold" :class="hintClass">
            {{ hint }}
          </p>
        </div>
        <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl bg-primary/10 text-primary">
          <component :is="icon" class="h-5 w-5" />
        </div>
      </div>
    </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { Card, CardContent } from '@/components/ui/card'

defineProps<{
  icon: unknown
  label: string
  value: string | number
  hint?: string
  hintClass?: string
}>()
</script>
