<script setup>
import { ref } from 'vue'

const props = defineProps({
  isOpen: {
    type: Boolean,
    required: true,
  },
  serverId: {
    type: Number,
    required: true,
  },
  isLoading: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['close', 'submit'])

const username = ref('')
const usernameError = ref('')

const validateUsername = () => {
  if (!username.value.trim()) {
    usernameError.value = 'Username is required'
    return false
  }
  
  if (!/^[a-zA-Z0-9_]+$/.test(username.value)) {
    usernameError.value = 'Username can only contain letters, numbers, and underscores'
    return false
  }
  
  if (username.value.length < 3 || username.value.length > 20) {
    usernameError.value = 'Username must be between 3 and 20 characters'
    return false
  }
  
  usernameError.value = ''
  return true
}

const handleSubmit = () => {
  if (validateUsername()) {
    emit('submit', username.value.trim())
  }
}

const handleClose = () => {
  username.value = ''
  usernameError.value = ''
  emit('close')
}
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="isOpen" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
        <div class="bg-white rounded-lg shadow-xl max-w-md w-full animate-in">
          <div class="p-6">
            <h2 class="text-xl font-bold text-slate-900 mb-4">Create FTP User</h2>
            
            <div class="mb-4">
              <label for="username" class="block text-sm font-medium text-slate-700 mb-2">
                Username
              </label>
              <input
                id="username"
                v-model="username"
                type="text"
                placeholder="e.g., dayz_MyServer"
                :disabled="isLoading"
                class="w-full px-3 py-2 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
                @keyup.enter="handleSubmit"
              />
              <p v-if="usernameError" class="mt-1 text-sm text-red-600">
                {{ usernameError }}
              </p>
              <p class="mt-1 text-sm text-slate-500">
                Username can contain letters, numbers, and underscores only (3-20 characters)
              </p>
            </div>
            
            <div class="flex gap-3 justify-end">
              <button
                @click="handleClose"
                :disabled="isLoading"
                class="px-4 py-2 rounded-lg text-slate-700 bg-slate-100 hover:bg-slate-200 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Cancel
              </button>
              <button
                @click="handleSubmit"
                :disabled="isLoading"
                class="px-4 py-2 rounded-lg bg-blue-500 text-white hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium"
              >
                <span v-if="isLoading" class="inline-block w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2"></span>
                Create User
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