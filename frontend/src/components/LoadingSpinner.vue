<script setup>
defineProps({
  size: {
    type: String,
    default: 'md', // sm, md, lg
  },
  text: {
    type: String,
    default: 'Loading...',
  },
})

const sizeClasses = {
  sm: 'w-6 h-6',
  md: 'w-8 h-8',
  lg: 'w-12 h-12',
}
</script>

<template>
  <div class="flex flex-col items-center justify-center gap-3">
    <div :class="[sizeClasses[size], 'border-3 border-slate-200 border-t-blue-500 rounded-full animate-spin']"></div>
    <p v-if="text" class="text-slate-600 font-medium">{{ text }}</p>
  </div>
</template>

<style scoped>
</style>
