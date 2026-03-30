<template>
  <div class="surface-card flex flex-col items-center justify-center gap-3 p-8 text-center">
    <div class="flex h-14 w-14 items-center justify-center rounded-2xl bg-[var(--admin-surface-2)] text-[var(--admin-text-muted)]">
      <component :is="icon" class="h-6 w-6" />
    </div>
    <div>
      <p class="text-base font-semibold">{{ title }}</p>
      <p class="mt-1 text-sm text-[var(--admin-text-muted)]">{{ description }}</p>
    </div>
    <button v-if="actionLabel" class="btn-primary mt-2" @click="$emit('action')">{{ actionLabel }}</button>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  icon: unknown
  title: string
  description: string
  actionLabel?: string
}>()

defineEmits<{ action: [] }>()
</script>
