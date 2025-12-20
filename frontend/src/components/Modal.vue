<script setup>
const props = defineProps({
  isOpen: {
    type: Boolean,
    required: true,
  },
  title: {
    type: String,
    required: true,
  },
  actionButtonText: {
    type: String,
    default: 'Confirm',
  },
  actionButtonType: {
    type: String,
    default: 'primary', // 'primary', 'danger'
  },
  isLoading: {
    type: Boolean,
    default: false,
  },
  isDangerous: {
    type: Boolean,
    default: false,
  },
})

defineEmits(['close', 'action'])

const getActionButtonClasses = () => {
  const base = 'px-4 py-2 rounded-lg font-medium transition-colors'
  if (props.actionButtonType === 'danger') {
    return `${base} bg-red-500 text-white hover:bg-red-600 disabled:opacity-50 disabled:cursor-not-allowed`
  }
  return `${base} bg-blue-500 text-white hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed`
}
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="isOpen" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
        <div class="bg-white rounded-lg shadow-xl max-w-md w-full animate-in">
          <div class="p-6">
            <h2 class="text-xl font-bold text-slate-900 mb-4">{{ title }}</h2>
            <div class="mb-6 text-slate-600">
              <slot />
            </div>
            
            <div class="flex gap-3 justify-end">
              <button
                @click="$emit('close')"
                :disabled="isLoading"
                class="px-4 py-2 rounded-lg text-slate-700 bg-slate-100 hover:bg-slate-200 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Cancel
              </button>
              <button
                @click="$emit('action')"
                :disabled="isLoading"
                :class="getActionButtonClasses()"
              >
                <span v-if="isLoading" class="inline-block w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2"></span>
                {{ actionButtonText }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.3s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.animate-in {
  animation: slideIn 0.3s ease;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}
</style>
