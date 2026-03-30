<template>
  <Sheet :open="open" @update:open="(val) => !val && $emit('close')">
    <SheetContent side="right" class="w-full overflow-y-auto sm:max-w-[34rem]">
      <SheetHeader>
        <SheetTitle>{{ staff ? 'Edit Staff' : 'Add Staff' }}</SheetTitle>
        <SheetDescription>Manage credentials and branch assignment.</SheetDescription>
      </SheetHeader>

      <form class="mt-6 space-y-4" @submit.prevent="submit">
        <div class="space-y-1.5">
          <Label for="staff-name">Full Name</Label>
          <Input id="staff-name" v-model="form.name" required />
        </div>

        <div class="space-y-1.5">
          <Label for="staff-code">Staff Code</Label>
          <Input id="staff-code" v-model="form.staff_code" required />
        </div>

        <div class="space-y-1.5">
          <Label for="staff-password">{{ staff ? 'New Password' : 'Password' }}</Label>
          <Input id="staff-password" v-model="form.password" type="password" :required="!staff" />
        </div>

        <div class="space-y-1.5">
          <Label for="staff-role">Role</Label>
          <Select v-model="form.role">
            <SelectTrigger id="staff-role" class="w-full">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="r in availableRoles" :key="r" :value="r">
                {{ r.charAt(0).toUpperCase() + r.slice(1) }}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div v-if="showBranchSelect" class="space-y-1.5">
          <Label for="staff-branch">Branch</Label>
          <Select v-model="form.branch_id">
            <SelectTrigger id="staff-branch" class="w-full">
              <SelectValue placeholder="Select branch" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="branch in branches" :key="branch.id" :value="branch.id">
                {{ branch.name }}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>

        <SheetFooter class="pt-2">
          <Button type="button" variant="outline" @click="$emit('close')">Cancel</Button>
          <Button type="submit" :disabled="loading">
            <span v-if="loading" class="flex items-center gap-2">
              <span class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
              Saving...
            </span>
            <span v-else>{{ staff ? 'Save Changes' : 'Create Staff' }}</span>
          </Button>
        </SheetFooter>
      </form>
    </SheetContent>
  </Sheet>
</template>

<script setup lang="ts">
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import type { BranchSummary, Staff, StaffRole } from 'types/auth'

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
