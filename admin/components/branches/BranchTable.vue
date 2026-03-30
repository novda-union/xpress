<template>
  <div class="grid gap-4 xl:grid-cols-2">
    <Card v-for="branch in branches" :key="branch.id" class="overflow-hidden">
      <div class="h-24 bg-gradient-to-br from-primary/8 via-background to-muted" />
      <CardContent class="space-y-4 p-5">
        <div class="flex items-start justify-between gap-3">
          <div>
            <p class="text-lg font-semibold">{{ branch.name }}</p>
            <p class="mt-1 text-sm text-muted-foreground">{{ branch.address }}</p>
          </div>
          <Badge :variant="branch.is_active ? 'default' : 'secondary'">
            {{ branch.is_active ? 'Active' : 'Inactive' }}
          </Badge>
        </div>
        <div class="grid grid-cols-2 gap-3 text-sm">
          <div class="rounded-lg bg-muted px-3 py-3">
            <p class="text-xs font-medium uppercase tracking-wide text-muted-foreground">Staff</p>
            <p class="mt-1 text-base font-semibold">{{ branch.staff_count ?? 0 }}</p>
          </div>
          <div class="rounded-lg bg-muted px-3 py-3">
            <p class="text-xs font-medium uppercase tracking-wide text-muted-foreground">Coordinates</p>
            <p class="mt-1 text-base font-semibold">
              {{ branch.lat ?? '—' }}, {{ branch.lng ?? '—' }}
            </p>
          </div>
        </div>
        <div class="flex gap-3">
          <Button variant="outline" class="flex-1" @click="$emit('edit', branch)">Edit</Button>
          <Button variant="destructive" class="flex-1" @click="$emit('deactivate', branch)">
            {{ branch.is_active ? 'Deactivate' : 'Delete' }}
          </Button>
        </div>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import type { BranchSummary } from 'types/auth'

defineProps<{ branches: BranchSummary[] }>()

defineEmits<{
  deactivate: [branch: BranchSummary]
  edit: [branch: BranchSummary]
}>()
</script>
