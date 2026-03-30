<template>
  <div class="space-y-6">
    <div v-for="group in groups" :key="group.branch_name" class="space-y-3">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-lg font-semibold">{{ group.branch_name }}</p>
          <p class="text-sm text-[var(--admin-text-muted)]">{{ group.staff.length }} team member(s)</p>
        </div>
      </div>

      <div class="table-shell">
        <table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Staff Code</th>
              <th>Role</th>
              <th class="hide-on-tablet">Status</th>
              <th />
            </tr>
          </thead>
          <tbody>
            <tr v-for="staff in group.staff" :key="staff.id">
              <td>{{ staff.name }}</td>
              <td>{{ staff.staff_code }}</td>
              <td>
                <span class="badge" :class="roleClass(staff.role)">{{ staff.role }}</span>
              </td>
              <td class="hide-on-tablet">
                <span class="badge badge-dot" :class="staff.is_active === false ? 'status-inactive' : 'status-active'">
                  {{ staff.is_active === false ? 'Inactive' : 'Active' }}
                </span>
              </td>
              <td>
                <div class="flex justify-end gap-2">
                  <button class="btn-secondary" @click="$emit('edit', staff)">Edit</button>
                  <button class="btn-danger" @click="$emit('deactivate', staff)">Deactivate</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Staff, StaffGroup } from '~/types/auth'

defineProps<{ groups: StaffGroup[] }>()

defineEmits<{
  deactivate: [staff: Staff]
  edit: [staff: Staff]
}>()

function roleClass(role: string) {
  const classes: Record<string, string> = {
    director: 'role-director',
    manager: 'role-manager',
    barista: 'role-barista',
  }
  return classes[role] || 'status-inactive'
}
</script>
