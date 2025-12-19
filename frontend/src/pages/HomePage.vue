<script setup>
import BaseButton from '../components/BaseButton.vue'
import { useCounterStore } from '../stores/counter'

const counter = useCounterStore()
</script>

<template>
  <main class="mx-auto w-full max-w-5xl px-6 py-10">
    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <h2 class="text-xl font-semibold text-slate-900">Frontend scaffold ready</h2>
      <p class="mt-2 text-sm text-slate-600">
        Tailwind utilities are active. Pinia is initialized. Axios client is configured in
        <code class="rounded bg-slate-100 px-1 py-0.5">src/utils/apiClient.js</code>.
      </p>

      <div class="mt-6 grid gap-4 sm:grid-cols-2">
        <div class="rounded-xl bg-slate-50 p-4">
          <div class="text-sm font-medium text-slate-700">Pinia example store</div>
          <div class="mt-2 text-3xl font-semibold text-slate-900">{{ counter.count }}</div>
          <div class="mt-1 text-sm text-slate-500">Doubled: {{ counter.doubledCount }}</div>
          <div class="mt-4 flex flex-wrap gap-2">
            <BaseButton @click="counter.increment()">Increment</BaseButton>
            <BaseButton variant="secondary" @click="counter.reset()">Reset</BaseButton>
          </div>
        </div>

        <div class="rounded-xl bg-slate-50 p-4">
          <div class="text-sm font-medium text-slate-700">Next steps</div>
          <ul class="mt-2 list-inside list-disc text-sm text-slate-600">
            <li>Create your dashboard pages in <code class="rounded bg-slate-100 px-1">src/pages</code></li>
            <li>Add reusable widgets in <code class="rounded bg-slate-100 px-1">src/components</code></li>
            <li>Centralize API calls via <code class="rounded bg-slate-100 px-1">apiClient</code></li>
          </ul>
        </div>
      </div>
    </div>
  </main>
</template>
