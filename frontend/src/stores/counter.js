import { defineStore } from 'pinia'

export const useCounterStore = defineStore('counter', {
  state: () => ({
    count: 0,
  }),
  getters: {
    doubledCount: (state) => state.count * 2,
  },
  actions: {
    increment() {
      this.count += 1
    },
    reset() {
      this.count = 0
    },
  },
})
