<script setup>
import NavBar from '../components/NavBar.vue'
import SideBar from '../components/SideBar.vue'
import Toast from '../components/Toast.vue'
import { useAuthStore } from '../stores/authStore'
import { onMounted } from 'vue'

const authStore = useAuthStore()

onMounted(() => {
  // Fetch user info on mount
  if (authStore.isAuthenticated && !authStore.user) {
    authStore.fetchUser().catch(() => {
      // Silently fail - user might be logged out
    })
  }
})
</script>

<template>
  <div class="min-h-screen bg-slate-50 flex flex-col">
    <NavBar />
    <div class="flex flex-1 overflow-hidden">
      <SideBar />
      <main class="flex-1 overflow-y-auto">
        <div class="p-8">
          <slot />
        </div>
      </main>
    </div>
    <Toast />
  </div>
</template>

<style scoped>
</style>
