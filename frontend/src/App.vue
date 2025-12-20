<script setup>
import { computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from './stores/authStore'
import MainLayout from './layouts/MainLayout.vue'
import AuthLayout from './layouts/AuthLayout.vue'

const route = useRoute()
const authStore = useAuthStore()

const layoutComponent = computed(() => {
  const layout = route.meta.layout
  if (layout === 'auth') {
    return AuthLayout
  }
  return MainLayout
})

onMounted(async () => {
  // Check if user is already authenticated
  if (authStore.token && !authStore.user) {
    try {
      await authStore.fetchUser()
    } catch (error) {
      // Token is invalid, clear auth
      authStore.clearAuth()
    }
  }
})
</script>

<template>
  <component :is="layoutComponent">
    <RouterView />
  </component>
</template>
