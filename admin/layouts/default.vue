<template>
  <div class="admin-shell flex min-h-screen">
    <LayoutAppSidebar />
    <div class="flex min-h-screen flex-1 flex-col">
      <LayoutAppHeader />
      <main class="flex-1 px-4 py-4 lg:px-6 lg:py-6">
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
const branchContext = useBranchContext()
const auth = useAuth()

auth.init()
branchContext.init()

onMounted(() => {
  branchContext.loadBranches()
})
</script>
