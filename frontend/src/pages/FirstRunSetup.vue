<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/authStore'
import { useUiStore } from '../stores/uiStore'
import LoadingSpinner from '../components/LoadingSpinner.vue'

const router = useRouter()
const authStore = useAuthStore()
const uiStore = useUiStore()

const username = ref('')
const password = ref('')
const confirmPassword = ref('')
const errors = ref({})
const checkingFirstRun = ref(true)

onMounted(async () => {
  // Check if this is actually first run
  try {
    const isFirstRun = await authStore.checkFirstRun()
    if (!isFirstRun) {
      router.push('/login')
    }
  } catch (error) {
    // If there's an error, assume it could be first run
  } finally {
    checkingFirstRun.value = false
  }
})

const validateForm = () => {
  errors.value = {}
  
  if (!username.value.trim()) {
    errors.value.username = 'Username is required'
  }
  
  if (!password.value) {
    errors.value.password = 'Password is required'
  } else if (password.value.length < 8) {
    errors.value.password = 'Password must be at least 8 characters'
  }
  
  if (!confirmPassword.value) {
    errors.value.confirmPassword = 'Please confirm your password'
  } else if (password.value !== confirmPassword.value) {
    errors.value.confirmPassword = 'Passwords do not match'
  }
  
  return Object.keys(errors.value).length === 0
}

const handleSubmit = async () => {
  if (!validateForm()) {
    return
  }

  try {
    await authStore.setup(username.value, password.value)
    uiStore.showSuccess('Admin user created successfully')
    router.push('/')
  } catch (error) {
    uiStore.showError(authStore.error || 'Setup failed')
  }
}

const handleKeydown = (e) => {
  if (e.key === 'Enter') {
    handleSubmit()
  }
}
</script>

<template>
  <div v-if="checkingFirstRun" class="w-full max-w-md">
    <div class="bg-white rounded-lg shadow-xl p-8">
      <LoadingSpinner text="Checking setup status..." />
    </div>
  </div>
  
  <div v-else class="w-full max-w-md">
    <div class="bg-white rounded-lg shadow-xl p-8">
      <div class="mb-8 text-center">
        <div class="w-16 h-16 bg-gradient-to-br from-green-500 to-green-600 rounded-lg flex items-center justify-center mx-auto mb-4 shadow-lg">
          <span class="text-white font-bold text-2xl">✓</span>
        </div>
        <h1 class="text-3xl font-bold text-slate-900 mb-2">Welcome to DayZ Manager</h1>
        <p class="text-slate-600">Set up your admin account</p>
      </div>

      <form @submit.prevent="handleSubmit" @keydown="handleKeydown" class="space-y-6">
        <div>
          <label for="username" class="block text-sm font-medium text-slate-700 mb-2">
            Admin Username
          </label>
          <input
            id="username"
            v-model="username"
            type="text"
            placeholder="Choose an admin username"
            :class="[
              'w-full px-4 py-3 rounded-lg border transition-colors',
              errors.username
                ? 'border-red-500 bg-red-50'
                : 'border-slate-300 bg-slate-50 focus:bg-white focus:border-green-500'
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
            placeholder="Create a strong password"
            :class="[
              'w-full px-4 py-3 rounded-lg border transition-colors',
              errors.password
                ? 'border-red-500 bg-red-50'
                : 'border-slate-300 bg-slate-50 focus:bg-white focus:border-green-500'
            ]"
            :disabled="authStore.loading"
            required
          />
          <p class="mt-1 text-xs text-slate-500">Minimum 8 characters</p>
          <p v-if="errors.password" class="mt-2 text-sm text-red-600">{{ errors.password }}</p>
        </div>

        <div>
          <label for="confirmPassword" class="block text-sm font-medium text-slate-700 mb-2">
            Confirm Password
          </label>
          <input
            id="confirmPassword"
            v-model="confirmPassword"
            type="password"
            placeholder="Confirm your password"
            :class="[
              'w-full px-4 py-3 rounded-lg border transition-colors',
              errors.confirmPassword
                ? 'border-red-500 bg-red-50'
                : 'border-slate-300 bg-slate-50 focus:bg-white focus:border-green-500'
            ]"
            :disabled="authStore.loading"
            required
          />
          <p v-if="errors.confirmPassword" class="mt-2 text-sm text-red-600">{{ errors.confirmPassword }}</p>
        </div>

        <button
          type="submit"
          :disabled="authStore.loading"
          class="w-full px-4 py-3 rounded-lg bg-green-500 text-white font-semibold hover:bg-green-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          <span v-if="authStore.loading" class="inline-block w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2"></span>
          {{ authStore.loading ? 'Setting up...' : 'Create Admin Account' }}
        </button>
      </form>

      <div class="mt-6 pt-6 border-t border-slate-200">
        <p class="text-center text-sm text-slate-600">
          Already have an account?
          <RouterLink to="/login" class="text-blue-600 hover:text-blue-700 font-medium">
            Sign in
          </RouterLink>
        </p>
      </div>
    </div>
  </div>
</template>

<style scoped>
</style>
