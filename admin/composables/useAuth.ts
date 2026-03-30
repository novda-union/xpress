import type { AuthState, Staff } from '~/types/auth'

const state = reactive<AuthState>({
  token: null,
  staff: null,
  initialized: false,
})

function isValidStaff(value: unknown): value is Staff {
  if (!value || typeof value !== 'object') return false

  const candidate = value as Partial<Staff>
  return typeof candidate.id === 'string' &&
    typeof candidate.store_id === 'string' &&
    typeof candidate.staff_code === 'string' &&
    typeof candidate.name === 'string' &&
    (candidate.branch_id === undefined || candidate.branch_id === null || typeof candidate.branch_id === 'string') &&
    (candidate.branch_name === undefined || candidate.branch_name === null || typeof candidate.branch_name === 'string') &&
    ['director', 'manager', 'barista'].includes(candidate.role ?? '')
}

export function useAuth() {
  const config = useRuntimeConfig()
  const router = useRouter()

  const isAuthenticated = computed(() => !!state.token && !!state.staff)
  const role = computed(() => state.staff?.role ?? null)

  function init() {
    if (!import.meta.client || state.initialized) return

    state.token = localStorage.getItem('admin_token')
    const staffJson = localStorage.getItem('admin_staff')

    try {
      const parsed = staffJson ? JSON.parse(staffJson) : null
      state.staff = isValidStaff(parsed) ? parsed : null
    } catch {
      state.staff = null
    }

    // Clear partial or invalid persisted auth to avoid redirect loops.
    if (!state.token || !state.staff) {
      state.token = null
      state.staff = null
      localStorage.removeItem('admin_token')
      localStorage.removeItem('admin_staff')
      localStorage.removeItem('admin_selected_branch_id')
    }

    state.initialized = true
  }

  async function login(storeCode: string, staffCode: string, password: string) {
    const res = await $fetch<{ token: string; staff: Staff }>(`${config.public.apiBase}/admin/auth`, {
      method: 'POST',
      body: { store_code: storeCode, staff_code: staffCode, password },
    })

    state.token = res.token
    state.staff = res.staff
    localStorage.setItem('admin_token', res.token)
    localStorage.setItem('admin_staff', JSON.stringify(res.staff))

    router.push('/')
  }

  function logout() {
    state.token = null
    state.staff = null
    state.initialized = true
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_staff')
    localStorage.removeItem('admin_selected_branch_id')
    router.push('/login')
  }

  function getAuthHeaders(): Record<string, string> {
    return state.token ? { Authorization: `Bearer ${state.token}` } : {}
  }

  return {
    state: readonly(state),
    isAuthenticated,
    role,
    init,
    login,
    logout,
    getAuthHeaders,
  }
}
