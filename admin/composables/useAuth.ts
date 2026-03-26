interface Staff {
  id: string
  store_id: string
  staff_code: string
  name: string
  role: string
}

interface AuthState {
  token: string | null
  staff: Staff | null
}

const state = reactive<AuthState>({
  token: null,
  staff: null,
})

export function useAuth() {
  const config = useRuntimeConfig()
  const router = useRouter()

  const isAuthenticated = computed(() => !!state.token)

  function init() {
    if (import.meta.client) {
      state.token = localStorage.getItem('admin_token')
      const staffJson = localStorage.getItem('admin_staff')
      if (staffJson) {
        state.staff = JSON.parse(staffJson)
      }
    }
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
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_staff')
    router.push('/login')
  }

  function getAuthHeaders() {
    return state.token ? { Authorization: `Bearer ${state.token}` } : {}
  }

  return {
    state: readonly(state),
    isAuthenticated,
    init,
    login,
    logout,
    getAuthHeaders,
  }
}
