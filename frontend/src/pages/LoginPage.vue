<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/authStore'
import { useUiStore } from '../stores/uiStore'

const router = useRouter()
const authStore = useAuthStore()
const uiStore = useUiStore()

const username = ref('')
const password = ref('')
const errors = ref({})

const validateForm = () => {
  errors.value = {}
  
  if (!username.value.trim()) {
    errors.value.username = 'Username is required'
  }
  
  if (!password.value) {
    errors.value.password = 'Password is required'
  }
  
  return Object.keys(errors.value).length === 0
}

const handleSubmit = async () => {
  if (!validateForm()) {
    return
  }

  try {
    await authStore.login(username.value, password.value)
    uiStore.showSuccess('Login successful')
    router.push('/')
  } catch (error) {
    uiStore.showError(authStore.error || 'Login failed')
    password.value = ''
  }
}

const handleKeydown = (e) => {
  if (e.key === 'Enter') {
    handleSubmit()
  }
}
</script>

<template>
  <div class="w-full max-w-md">
    <div class="bg-white rounded-lg shadow-xl p-8">
      <div class="mb-8 text-center">
        <div class="w-16 h-16 bg-gradient-to-br from-blue-500 to-blue-600 rounded-lg flex items-center justify-center mx-auto mb-4 shadow-lg">
          <span class="text-white font-bold text-2xl">D</span>
        </div>
        <h1 class="text-3xl font-bold text-slate-900 mb-2">DayZ Manager</h1>
        <p class="text-slate-600">Sign in to your dashboard</p>
      </div>

      <form @submit.prevent="handleSubmit" @keydown="handleKeydown" class="space-y-6">
        <div>
          <label for="username" class="block text-sm font-medium text-slate-700 mb-2">
            Username
          </label>
          <input
            id="username"
            v-model="username"
            type="text"
            placeholder="Enter your username"
            :class="[
              'w-full px-4 py-3 rounded-lg border transition-colors',
              errors.username
                ? 'border-red-500 bg-red-50'
                : 'border-slate-300 bg-slate-50 focus:bg-white focus:border-blue-500'
            ]"
            :disabled="authStore.loading"
            required
          />
          <p v-if="errors.username" class="mt-2 text-sm text-red-600">{{ errors.username }}</p>
        </div>

        <div>
          <label for="password" class="block text-sm font-medium text-slate-700 mb-2">
            Password
          </label>
          <input
            id="password"
            v-model="password"
            type="password"
            placeholder="Enter your password"
            :class="[
              'w-full px-4 py-3 rounded-lg border transition-colors',
              errors.password
                ? 'border-red-500 bg-red-50'
                : 'border-slate-300 bg-slate-50 focus:bg-white focus:border-blue-500'
            ]"
            :disabled="authStore.loading"
            required
          />
          <p v-if="errors.password" class="mt-2 text-sm text-red-600">{{ errors.password }}</p>
        </div>

        <button
          type="submit"
          :disabled="authStore.loading"
          class="w-full px-4 py-3 rounded-lg bg-blue-500 text-white font-semibold hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          <span v-if="authStore.loading" class="inline-block w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2"></span>
          {{ authStore.loading ? 'Signing in...' : 'Sign In' }}
        </button>
      </form>

      <div class="mt-6 pt-6 border-t border-slate-200">
        <p class="text-center text-sm text-slate-600">
          Don't have an account?
          <RouterLink to="/setup" class="text-blue-600 hover:text-blue-700 font-medium">
            Create one
          </RouterLink>
        </p>
      </div>
    </div>
  </div>
</template>

<style scoped>
</style>
