<script setup>
defineProps({
  status: {
    type: String,
    required: true,
    // 'running', 'stopped', 'loading', 'error'
  },
})

const statusConfig = {
  running: {
    bg: 'bg-green-100',
    text: 'text-green-700',
    label: 'Running',
    icon: '●',
  },
  stopped: {
    bg: 'bg-slate-100',
    text: 'text-slate-700',
    label: 'Stopped',
    icon: '●',
  },
  loading: {
    bg: 'bg-blue-100',
    text: 'text-blue-700',
    label: 'Loading',
    icon: '◌',
  },
  error: {
    bg: 'bg-red-100',
    text: 'text-red-700',
    label: 'Error',
    icon: '●',
  },
}

const config = statusConfig[status] || statusConfig.stopped
</script>

<template>
  <span :class="[config.bg, config.text, 'inline-flex items-center gap-2 px-3 py-1 rounded-full text-sm font-medium']">
    <span>{{ config.icon }}</span>
    <span>{{ config.label }}</span>
  </span>
</template>

<style scoped>
</style>
