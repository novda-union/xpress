export default defineNuxtRouteMiddleware((to) => {
  if (import.meta.server) return

  const { isAuthenticated, init, state } = useAuth()
  const { canVisit } = usePermissions()
  init()

  if (to.path !== '/login' && !isAuthenticated.value) {
    return navigateTo('/login')
  }

  if (to.path === '/login' && isAuthenticated.value) {
    return navigateTo('/')
  }

  if (to.path !== '/login' && isAuthenticated.value && !canVisit(to.path)) {
    const fallback = state.staff?.role === 'director'
      ? '/'
      : canVisit('/orders')
          ? '/orders'
          : canVisit('/settings/branch')
              ? '/settings/branch'
              : '/login'

    if (fallback !== to.path) {
      return navigateTo(fallback)
    }
  }
})
