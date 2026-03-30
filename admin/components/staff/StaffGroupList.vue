<template>
  <div class="space-y-6">
    <div v-for="group in groups" :key="group.branch_name" class="space-y-3">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-lg font-semibold">{{ group.branch_name }}</p>
          <p class="text-sm text-muted-foreground">{{ group.staff.length }} team member(s)</p>
        </div>
      </div>

      <Card>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b bg-muted/50">
                <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Name</th>
                <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Staff Code</th>
                <th class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground">Role</th>
                <th class="hidden px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted-foreground lg:table-cell">Status</th>
                <th class="px-4 py-3" />
              </tr>
            </thead>
            <tbody>
              <tr v-for="staff in group.staff" :key="staff.id" class="border-b last:border-0 hover:bg-muted/30 transition-colors">
                <td class="px-4 py-3 font-medium">{{ staff.name }}</td>
                <td class="px-4 py-3 font-mono text-xs text-muted-foreground">{{ staff.staff_code }}</td>
                <td class="px-4 py-3">
                  <Badge :class="roleClass(staff.role)" variant="secondary">{{ staff.role }}</Badge>
                </td>
                <td class="hidden px-4 py-3 lg:table-cell">
                  <Badge :variant="staff.is_active === false ? 'secondary' : 'default'">
                    {{ staff.is_active === false ? 'Inactive' : 'Active' }}
                  </Badge>
                </td>
                <td class="px-4 py-3">
                  <div class="flex justify-end gap-2">
                    <Button variant="outline" size="sm" @click="$emit('edit', staff)">Edit</Button>
                    <Button variant="destructive" size="sm" @click="$emit('deactivate', staff)">Deactivate</Button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import type { Staff, StaffGroup } from 'types/auth'

defineProps<{ groups: StaffGroup[] }>()

defineEmits<{
  deactivate: [staff: Staff]
  edit: [staff: Staff]
}>()

function roleClass(role: string) {
  const classes: Record<string, string> = {
    director: 'bg-violet-100 text-violet-700 hover:bg-violet-100',
    manager: 'bg-indigo-100 text-indigo-700 hover:bg-indigo-100',
    barista: 'bg-cyan-100 text-cyan-700 hover:bg-cyan-100',
  }
  return classes[role] || ''
}
</script>
