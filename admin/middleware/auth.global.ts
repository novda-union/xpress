export default defineNuxtRouteMiddleware((to) => {
  if (import.meta.server) return

  const { isAuthenticated, init } = useAuth()
  const { canVisit } = usePermissions()
  init()

  if (to.path !== '/login' && !isAuthenticated.value) {
    return navigateTo('/login')
  }

  if (to.path === '/login' && isAuthenticated.value) {
    return navigateTo('/')
  }

  if (to.path !== '/login' && isAuthenticated.value && !canVisit(to.path)) {
    return navigateTo('/')
  }
})
