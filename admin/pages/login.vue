<template>
  <div class="flex min-h-screen items-center justify-center p-6">
    <Card class="w-full max-w-md">
      <CardContent class="p-8">
        <div class="mb-8 text-center">
          <div class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-2xl bg-primary/10 text-primary">
            <Store class="h-6 w-6" />
          </div>
          <h1 class="text-2xl font-bold">Xpressgo Admin</h1>
          <p class="mt-2 text-sm text-muted-foreground">Welcome back. Sign in to manage branches, orders, and staff.</p>
        </div>

        <form class="space-y-4" @submit.prevent="handleLogin">
          <div class="space-y-1.5">
            <Label for="store-code">Store Code</Label>
            <Input id="store-code" v-model="form.storeCode" type="text" placeholder="e.g. demobar" required />
          </div>

          <div class="space-y-1.5">
            <Label for="staff-code">Staff Code</Label>
            <Input id="staff-code" v-model="form.staffCode" type="text" placeholder="e.g. admin" required />
          </div>

          <div class="space-y-1.5">
            <Label for="password">Password</Label>
            <Input id="password" v-model="form.password" type="password" placeholder="Password" required />
          </div>

          <p v-if="error" class="text-sm text-destructive">{{ error }}</p>

          <Button type="submit" class="w-full" :disabled="loading">
            <span v-if="loading" class="flex items-center gap-2">
              <span class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
              Signing in...
            </span>
            <span v-else>Sign In</span>
          </Button>
        </form>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { Store } from 'lucide-vue-next'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

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
