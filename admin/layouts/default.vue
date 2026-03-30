<template>
  <SidebarProvider>
    <LayoutAppSidebar />
    <SidebarInset class="flex min-h-svh flex-col">
      <LayoutAppHeader />
      <main class="flex-1 px-4 py-4 lg:px-6 lg:py-6">
        <slot />
      </main>
    </SidebarInset>
  </SidebarProvider>
</template>

<script setup lang="ts">
import { SidebarProvider, SidebarInset } from '@/components/ui/sidebar'

const branchContext = useBranchContext()
const auth = useAuth()

auth.init()
branchContext.init()

onMounted(() => {
  branchContext.loadBranches()
})
</script>
