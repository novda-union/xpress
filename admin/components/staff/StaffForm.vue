<template>
  <Teleport to="body">
    <div v-if="open">
      <div class="slide-panel-backdrop" @click="$emit('close')" />
      <aside class="slide-panel">
        <div class="mb-6 flex items-center justify-between">
          <div>
            <p class="text-lg font-semibold">{{ staff ? 'Edit Staff' : 'Add Staff' }}</p>
            <p class="text-sm text-[var(--admin-text-muted)]">Manage credentials and branch assignment.</p>
          </div>
          <button class="btn-ghost" @click="$emit('close')">Close</button>
        </div>

        <form class="field-grid" @submit.prevent="submit">
          <div>
            <label class="label">Full Name</label>
            <input v-model="form.name" class="input" required />
          </div>
          <div>
            <label class="label">Staff Code</label>
            <input v-model="form.staff_code" class="input" required />
          </div>
          <div>
            <label class="label">{{ staff ? 'New Password' : 'Password' }}</label>
            <input v-model="form.password" class="input" type="password" :required="!staff" />
          </div>
          <div>
            <label class="label">Role</label>
            <select v-model="form.role" class="select">
              <option v-for="role in availableRoles" :key="role" :value="role">
                {{ role }}
              </option>
            </select>
          </div>
          <div v-if="showBranchSelect">
            <label class="label">Branch</label>
            <select v-model="form.branch_id" class="select">
              <option disabled value="">Select branch</option>
              <option v-for="branch in branches" :key="branch.id" :value="branch.id">
                {{ branch.name }}
              </option>
            </select>
          </div>

          <div class="flex gap-3 pt-2">
            <button type="submit" class="btn-primary flex-1" :disabled="loading">
              {{ loading ? 'Saving...' : staff ? 'Save Changes' : 'Create Staff' }}
            </button>
            <button type="button" class="btn-secondary" @click="$emit('close')">Cancel</button>
          </div>
        </form>
      </aside>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import type { BranchSummary, Staff, StaffRole } from '~/types/auth'

const props = defineProps<{
  branches: BranchSummary[]
  loading?: boolean
  open: boolean
  role: StaffRole
  staff?: Staff | null
}>()

const emit = defineEmits<{
  close: []
  save: [payload: Record<string, unknown>]
}>()

const form = reactive({
  name: '',
  staff_code: '',
  password: '',
  role: 'barista' as StaffRole,
  branch_id: '',
})

const availableRoles = computed<StaffRole[]>(() => (props.role === 'director' ? ['manager', 'barista'] : ['barista']))
const showBranchSelect = computed(() => props.role === 'director' && form.role !== 'director')

watch(
  () => props.staff,
  (staff) => {
    form.name = staff?.name ?? ''
    form.staff_code = staff?.staff_code ?? ''
    form.password = ''
    form.role = (staff?.role as StaffRole) ?? availableRoles.value[0]
    form.branch_id = staff?.branch_id ?? ''
  },
  { immediate: true },
)

watch(
  availableRoles,
  (roles) => {
    if (!roles.includes(form.role)) {
      form.role = roles[0]
    }
  },
  { immediate: true },
)

function submit() {
  emit('save', {
    ...form,
    branch_id: showBranchSelect.value ? form.branch_id || null : null,
  })
}
</script>
