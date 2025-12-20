import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUiStore = defineStore('ui', () => {
  const toasts = ref([])
  const loading = ref(false)

  let toastId = 0

  const addToast = (message, type = 'info', duration = 3000) => {
    const id = ++toastId
    const toast = {
      id,
      message,
      type, // 'success', 'error', 'info', 'warning'
    }

    toasts.value.push(toast)

    if (duration) {
      setTimeout(() => {
        removeToast(id)
      }, duration)
    }

    return id
  }

  const removeToast = (id) => {
    const index = toasts.value.findIndex(t => t.id === id)
    if (index > -1) {
      toasts.value.splice(index, 1)
    }
  }

  const clearToasts = () => {
    toasts.value = []
  }

  const showSuccess = (message) => {
    return addToast(message, 'success', 3000)
  }

  const showError = (message) => {
    return addToast(message, 'error', 5000)
  }

  const showInfo = (message) => {
    return addToast(message, 'info', 3000)
  }

  const showWarning = (message) => {
    return addToast(message, 'warning', 4000)
  }

  const setLoading = (isLoading) => {
    loading.value = isLoading
  }

  return {
    toasts,
    loading,
    addToast,
    removeToast,
    clearToasts,
    showSuccess,
    showError,
    showInfo,
    showWarning,
    setLoading,
  }
})
