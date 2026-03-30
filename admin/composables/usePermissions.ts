import type { StaffRole } from 'types/auth'

const routeRules: Record<string, StaffRole[]> = {
  '/': ['director', 'manager', 'barista'],
  '/orders': ['director', 'manager', 'barista'],
  '/menu': ['director', 'manager'],
  '/branches': ['director'],
  '/staff': ['director', 'manager'],
  '/settings': ['director'],
  '/settings/branch': ['director', 'manager'],
}

export function usePermissions() {
  const { state } = useAuth()

  function can(action: string, targetRole?: StaffRole) {
    const role = state.staff?.role
    const branchId = state.staff?.branch_id
    if (!role) return false
    if (role === 'director') return true

    switch (action) {
      case 'branch:create':
      case 'branch:edit':
      case 'branch:delete':
      case 'staff:create:manager':
      case 'settings:store':
      case 'orders:view:all':
      case 'dashboard:all':
        return false
      case 'staff:create:barista':
      case 'menu:manage':
      case 'settings:branch':
        return role === 'manager' && !!branchId
      case 'orders:view':
      case 'dashboard:branch':
        return !!branchId
      case 'staff:edit':
        return role === 'manager' && !!branchId && targetRole === 'barista'
      default:
        return false
    }
  }

  function canVisit(path: string) {
    const role = state.staff?.role
    if (!role) return false
    const match = Object.keys(routeRules)
      .sort((a, b) => b.length - a.length)
      .find((rule) => path === rule || path.startsWith(`${rule}/`))
    return match ? routeRules[match].includes(role) : true
  }

  return { can, canVisit }
}
