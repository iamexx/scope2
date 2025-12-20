<script setup>
import StatusBadge from './StatusBadge.vue'
import { ref } from 'vue'
import { useUiStore } from '../stores/uiStore'
import { useServerStore } from '../stores/serverStore'

const props = defineProps({
  server: {
    type: Object,
    required: true,
  },
})

const emit = defineEmits(['server-action'])

const uiStore = useUiStore()
const serverStore = useServerStore()

const isLoading = ref(false)
const showConfirmModal = ref(false)
const pendingAction = ref(null)

const handleAction = async (action) => {
  try {
    isLoading.value = true
    
    switch (action) {
      case 'start':
        await serverStore.startServer(props.server.id)
        uiStore.showSuccess(`Server ${props.server.name} is starting...`)
        break
      case 'stop':
        showConfirmModal.value = true
        pendingAction.value = action
        return
      case 'restart':
        await serverStore.restartServer(props.server.id)
        uiStore.showSuccess(`Server ${props.server.name} is restarting...`)
        break
    }
    
    emit('server-action', action)
  } catch (error) {
    uiStore.showError(`Failed to ${action} server`)
    console.error(`Error ${action} server:`, error)
  } finally {
    isLoading.value = false
  }
}

const confirmAction = async () => {
  if (pendingAction.value === 'stop') {
    try {
      isLoading.value = true
      await serverStore.stopServer(props.server.id)
      uiStore.showSuccess(`Server ${props.server.name} is stopping...`)
      emit('server-action', 'stop')
    } catch (error) {
      uiStore.showError('Failed to stop server')
      console.error('Error stopping server:', error)
    } finally {
      isLoading.value = false
      showConfirmModal.value = false
      pendingAction.value = null
    }
  }
}

const closeConfirmModal = () => {
  showConfirmModal.value = false
  pendingAction.value = null
}
</script>

<template>
  <div class="bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow p-6 border border-slate-200">
    <div class="flex items-start justify-between mb-4">
      <div>
        <h3 class="text-lg font-bold text-slate-900">{{ server.name }}</h3>
        <p class="text-sm text-slate-500">ID: {{ server.id }}</p>
      </div>
      <StatusBadge :status="server.status || 'stopped'" />
    </div>
    
    <div class="space-y-2 mb-6 text-sm">
      <div class="flex justify-between text-slate-600">
        <span>Players:</span>
        <span class="font-medium">{{ server.players || 0 }} / {{ server.maxPlayers || 60 }}</span>
      </div>
      <div class="flex justify-between text-slate-600">
        <span>Port:</span>
        <span class="font-mono font-medium">{{ server.port || 'N/A' }}</span>
      </div>
      <div class="flex justify-between text-slate-600">
        <span>Created:</span>
        <span class="text-xs">{{ new Date(server.createdAt).toLocaleDateString() }}</span>
      </div>
    </div>
    
    <div class="flex gap-2">
      <button
        @click="handleAction('start')"
        :disabled="server.status === 'running' || isLoading"
        class="flex-1 px-3 py-2 rounded-lg bg-green-500 text-white hover:bg-green-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium text-sm"
      >
        <span v-if="isLoading" class="inline-block w-3 h-3 border-2 border-white border-t-transparent rounded-full animate-spin mr-1"></span>
        Start
      </button>
      <button
        @click="handleAction('stop')"
        :disabled="server.status !== 'running' || isLoading"
        class="flex-1 px-3 py-2 rounded-lg bg-red-500 text-white hover:bg-red-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium text-sm"
      >
        Stop
      </button>
      <button
        @click="handleAction('restart')"
        :disabled="server.status !== 'running' || isLoading"
        class="flex-1 px-3 py-2 rounded-lg bg-blue-500 text-white hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium text-sm"
      >
        Restart
      </button>
    </div>

    <!-- Confirm Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showConfirmModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div class="bg-white rounded-lg shadow-xl max-w-md w-full animate-in">
            <div class="p-6">
              <h2 class="text-xl font-bold text-slate-900 mb-4">Confirm Stop</h2>
              <p class="text-slate-600 mb-6">
                Are you sure you want to stop <strong>{{ server.name }}</strong>? This action cannot be undone immediately.
              </p>
              
              <div class="flex gap-3 justify-end">
                <button
                  @click="closeConfirmModal"
                  :disabled="isLoading"
                  class="px-4 py-2 rounded-lg text-slate-700 bg-slate-100 hover:bg-slate-200 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Cancel
                </button>
                <button
                  @click="confirmAction"
                  :disabled="isLoading"
                  class="px-4 py-2 rounded-lg bg-red-500 text-white hover:bg-red-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium"
                >
                  <span v-if="isLoading" class="inline-block w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2"></span>
                  Stop Server
                </button>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
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
