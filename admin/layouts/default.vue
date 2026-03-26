<template>
  <div class="flex min-h-screen">
    <aside class="w-64 bg-gray-900 text-white p-4 flex flex-col">
      <h1 class="text-xl font-bold mb-8">Xpressgo Admin</h1>
      <nav class="space-y-1 flex-1">
        <NuxtLink
          to="/"
          class="block px-3 py-2 rounded text-gray-300 hover:bg-gray-800 hover:text-white"
          active-class="bg-gray-800 text-white"
        >
          Dashboard
        </NuxtLink>
        <NuxtLink
          to="/orders"
          class="block px-3 py-2 rounded text-gray-300 hover:bg-gray-800 hover:text-white"
          active-class="bg-gray-800 text-white"
        >
          Orders
        </NuxtLink>
        <NuxtLink
          to="/menu"
          class="block px-3 py-2 rounded text-gray-300 hover:bg-gray-800 hover:text-white"
          active-class="bg-gray-800 text-white"
        >
          Menu
        </NuxtLink>
        <NuxtLink
          to="/settings"
          class="block px-3 py-2 rounded text-gray-300 hover:bg-gray-800 hover:text-white"
          active-class="bg-gray-800 text-white"
        >
          Settings
        </NuxtLink>
      </nav>
      <button @click="logout()" class="w-full px-3 py-2 text-left text-gray-400 hover:text-white hover:bg-gray-800 rounded">
        Logout
      </button>
    </aside>
    <main class="flex-1 bg-gray-50 p-6">
      <slot />
    </main>
  </div>
</template>

<script setup lang="ts">
const { logout } = useAuth()
</script>
