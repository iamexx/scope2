<script setup>
import { useRoute } from 'vue-router'
import { computed } from 'vue'

const route = useRoute()

const isActive = (routeName) => {
  return route.name === routeName
}

const navItems = [
  { name: 'Dashboard', routeName: 'ServerStatus', icon: '📊' },
]
</script>

<template>
  <aside class="w-64 bg-white border-r border-slate-200 shadow-sm">
    <div class="p-6">
      <h2 class="text-sm font-semibold text-slate-400 uppercase tracking-wider mb-6">Navigation</h2>
      <nav class="space-y-2">
        <RouterLink
          v-for="item in navItems"
          :key="item.routeName"
          :to="{ name: item.routeName }"
          :class="[
            'flex items-center gap-3 px-4 py-3 rounded-lg transition-colors',
            isActive(item.routeName)
              ? 'bg-blue-50 text-blue-600 font-medium'
              : 'text-slate-700 hover:bg-slate-50'
          ]"
        >
          <span class="text-xl">{{ item.icon }}</span>
          <span>{{ item.name }}</span>
        </RouterLink>
      </nav>
    </div>
  </aside>
</template>

<style scoped>
</style>
