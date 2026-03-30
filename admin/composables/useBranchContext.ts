import type { BranchSummary } from '~/types/auth'

const selectedBranchIdState = useState<string | null>('admin-selected-branch-id', () => null)
const branchesState = useState<BranchSummary[]>('admin-branches', () => [])
const loadingState = useState<boolean>('admin-branches-loading', () => false)

export function useBranchContext() {
  const { state } = useAuth()
  const { api } = useApi()

  const isDirector = computed(() => state.staff?.role === 'director')
  const selectedBranchId = computed({
    get: () => selectedBranchIdState.value,
    set: (value: string | null) => {
      selectedBranchIdState.value = value
      if (import.meta.client) {
        if (value) {
          localStorage.setItem('admin_selected_branch_id', value)
        } else {
          localStorage.removeItem('admin_selected_branch_id')
        }
      }
    },
  })

  const branches = computed(() => branchesState.value)
  const selectedBranch = computed(() => branchesState.value.find((branch) => branch.id === selectedBranchId.value) || null)
  const selectedBranchLabel = computed(() => {
    if (!isDirector.value && state.staff?.branch_name) return state.staff.branch_name
    if (selectedBranch.value) return selectedBranch.value.name
    return 'All Branches'
  })

  function init() {
    if (!state.staff) return
    if (isDirector.value) {
      if (import.meta.client) {
        selectedBranchIdState.value = localStorage.getItem('admin_selected_branch_id')
      }
      return
    }
    selectedBranchIdState.value = state.staff.branch_id
  }

  async function loadBranches() {
    if (!state.token) return
    loadingState.value = true
    try {
      const result = await api<BranchSummary[]>('/admin/branches')
      branchesState.value = result
      if (!isDirector.value && state.staff?.branch_id) {
        selectedBranchIdState.value = state.staff.branch_id
      }
      if (isDirector.value && selectedBranchIdState.value && !result.some((branch) => branch.id === selectedBranchIdState.value)) {
        selectedBranchIdState.value = null
      }
    } finally {
      loadingState.value = false
    }
  }

  function selectBranch(branchId: string | null) {
    if (!isDirector.value) return
    selectedBranchId.value = branchId
  }

  return {
    branches,
    isDirector,
    loading: readonly(loadingState),
    selectedBranch,
    selectedBranchId,
    selectedBranchLabel,
    init,
    loadBranches,
    selectBranch,
  }
}
