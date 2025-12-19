<script setup>
defineProps({
  variant: {
    type: String,
    default: 'primary',
  },
})
</script>

<template>
  <button
    class="inline-flex items-center justify-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition focus:outline-none focus:ring-2 focus:ring-offset-2"
    :class="
      variant === 'primary'
        ? 'bg-slate-900 text-white hover:bg-slate-800 focus:ring-slate-900'
        : 'bg-white text-slate-900 ring-1 ring-slate-300 hover:bg-slate-50 focus:ring-slate-400'
    "
    type="button"
  >
    <slot />
  </button>
</template>
