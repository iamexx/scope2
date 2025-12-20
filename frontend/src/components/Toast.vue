<script setup>
import { useUiStore } from '../stores/uiStore'
import { computed } from 'vue'

const uiStore = useUiStore()

const toastClasses = (type) => {
  const base = 'px-4 py-3 rounded-lg shadow-lg text-white font-medium flex items-center gap-2'
  switch (type) {
    case 'success':
      return `${base} bg-green-500`
    case 'error':
      return `${base} bg-red-500`
    case 'warning':
      return `${base} bg-amber-500`
    case 'info':
    default:
      return `${base} bg-blue-500`
  }
}

const toastIcon = (type) => {
  switch (type) {
    case 'success':
      return '✓'
    case 'error':
      return '✕'
    case 'warning':
      return '⚠'
    case 'info':
    default:
      return 'ℹ'
  }
}
</script>

<template>
  <div class="fixed bottom-4 right-4 space-y-2 z-50 max-w-sm pointer-events-none">
    <Transition name="toast" v-for="toast in uiStore.toasts" :key="toast.id">
      <div :class="toastClasses(toast.type)" class="pointer-events-auto">
        <span class="text-lg">{{ toastIcon(toast.type) }}</span>
        <span>{{ toast.message }}</span>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}

.toast-enter-from {
  opacity: 0;
  transform: translateX(30px);
}

.toast-leave-to {
  opacity: 0;
  transform: translateX(30px);
}
</style>
