<script setup>
defineProps({
  title: {
    type: String,
    default: 'Dashboard',
  },
})
</script>

<template>
  <header class="border-b border-slate-200 bg-white">
    <div class="mx-auto flex max-w-5xl items-center justify-between px-6 py-4">
      <h1 class="text-lg font-semibold text-slate-900">{{ title }}</h1>
      <div class="text-sm text-slate-500">Vue 3 · Tailwind v4 · Pinia</div>
    </div>
  </header>
</template>
