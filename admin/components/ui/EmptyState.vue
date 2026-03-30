<template>
  <Card>
    <CardContent class="flex flex-col items-center justify-center gap-3 p-8 text-center">
      <div class="flex h-14 w-14 items-center justify-center rounded-2xl bg-muted text-muted-foreground">
        <component :is="icon" class="h-6 w-6" />
      </div>
      <div>
        <p class="text-base font-semibold">{{ title }}</p>
        <p class="mt-1 text-sm text-muted-foreground">{{ description }}</p>
      </div>
      <Button v-if="actionLabel" class="mt-2" @click="$emit('action')">
        {{ actionLabel }}
      </Button>
    </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'

defineProps<{
  icon: unknown
  title: string
  description: string
  actionLabel?: string
}>()

defineEmits<{ action: [] }>()
</script>
