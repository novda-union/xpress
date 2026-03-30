<template>
  <div class="admin-shell flex min-h-screen items-center justify-center p-6">
    <div class="surface-card w-full max-w-md p-8">
      <div class="mb-8 text-center">
        <div class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-2xl bg-[var(--admin-accent-bg)] text-[var(--admin-accent)]">
          <Store class="h-6 w-6" />
        </div>
        <h1 class="text-2xl font-bold">Xpressgo Admin</h1>
        <p class="mt-2 text-sm text-[var(--admin-text-muted)]">Welcome back. Sign in to manage branches, orders, and staff.</p>
      </div>

      <form class="space-y-4" @submit.prevent="handleLogin">
        <div>
          <label class="label">Store Code</label>
          <input
            v-model="form.storeCode"
            type="text"
            placeholder="e.g. demobar"
            class="input"
            required
          />
        </div>

        <div>
          <label class="label">Staff Code</label>
          <input
            v-model="form.staffCode"
            type="text"
            placeholder="e.g. admin"
            class="input"
            required
          />
        </div>

        <div>
          <label class="label">Password</label>
          <input
            v-model="form.password"
            type="password"
            placeholder="Password"
            class="input"
            required
          />
        </div>

        <p v-if="error" class="text-sm text-[var(--admin-error)]">{{ error }}</p>

        <button
          type="submit"
          :disabled="loading"
          class="btn-primary w-full"
        >
          {{ loading ? 'Signing in...' : 'Sign In' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Store } from 'lucide-vue-next'

definePageMeta({ layout: false })

const { login } = useAuth()
const form = reactive({ storeCode: '', staffCode: '', password: '' })
const error = ref('')
const loading = ref(false)

async function handleLogin() {
  error.value = ''
  loading.value = true
  try {
    await login(form.storeCode, form.staffCode, form.password)
  } catch (e: any) {
    error.value = e?.data?.error || 'Login failed'
  } finally {
    loading.value = false
  }
}
</script>
