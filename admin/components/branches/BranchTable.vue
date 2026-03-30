<template>
  <div class="grid gap-4 xl:grid-cols-2">
    <div v-for="branch in branches" :key="branch.id" class="surface-card overflow-hidden">
      <div class="h-28 bg-gradient-to-br from-[var(--admin-accent-bg)] via-white to-[var(--admin-surface-2)]" />
      <div class="space-y-4 p-5">
        <div class="flex items-start justify-between gap-3">
          <div>
            <p class="text-lg font-semibold">{{ branch.name }}</p>
            <p class="mt-1 text-sm text-[var(--admin-text-muted)]">{{ branch.address }}</p>
          </div>
          <span class="badge badge-dot" :class="branch.is_active ? 'status-active' : 'status-inactive'">
            {{ branch.is_active ? 'Active' : 'Inactive' }}
          </span>
        </div>
        <div class="grid grid-cols-2 gap-3 text-sm text-[var(--admin-text-muted)]">
          <div class="surface-muted px-3 py-3">
            <p class="text-xs uppercase tracking-wide">Staff</p>
            <p class="mt-1 text-base font-semibold text-[var(--admin-text)]">{{ branch.staff_count ?? 0 }}</p>
          </div>
          <div class="surface-muted px-3 py-3">
            <p class="text-xs uppercase tracking-wide">Coordinates</p>
            <p class="mt-1 text-base font-semibold text-[var(--admin-text)]">
              {{ branch.lat ?? '—' }}, {{ branch.lng ?? '—' }}
            </p>
          </div>
        </div>
        <div class="flex gap-3">
          <button class="btn-secondary flex-1" @click="$emit('edit', branch)">Edit</button>
          <button class="btn-danger flex-1" @click="$emit('deactivate', branch)">
            {{ branch.is_active ? 'Deactivate' : 'Delete' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { BranchSummary } from '~/types/auth'

defineProps<{ branches: BranchSummary[] }>()

defineEmits<{
  deactivate: [branch: BranchSummary]
  edit: [branch: BranchSummary]
}>()
</script>
