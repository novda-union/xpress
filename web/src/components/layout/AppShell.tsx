import type { PropsWithChildren, ReactNode } from 'react'

interface AppShellProps extends PropsWithChildren {
  header?: ReactNode
  bottomBar?: ReactNode
}

export function AppShell({ header, bottomBar, children }: AppShellProps) {
  return (
    <div className="app-shell">
      {header}
      <main className="pb-safe">{children}</main>
      {bottomBar}
    </div>
  )
}
