<script setup>
import { useAuthStore } from '../stores/authStore'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}
</script>

<template>
  <nav class="bg-white border-b border-slate-200 shadow-sm">
    <div class="max-w-full px-6 py-4 flex items-center justify-between">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 bg-gradient-to-br from-blue-500 to-blue-600 rounded-lg flex items-center justify-center">
          <span class="text-white font-bold text-lg">D</span>
        </div>
        <h1 class="text-2xl font-bold text-slate-900">DayZ Manager</h1>
      </div>
      
      <div class="flex items-center gap-4">
        <div class="text-sm text-slate-600">
          <span v-if="authStore.user">{{ authStore.user.username }}</span>
          <span v-else>Guest</span>
        </div>
        <button
          @click="handleLogout"
          class="px-4 py-2 rounded-lg bg-slate-100 text-slate-700 hover:bg-slate-200 transition-colors font-medium"
        >
          Logout
        </button>
      </div>
    </div>
  </nav>
</template>

<style scoped>
</style>
