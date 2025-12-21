<script setup>
import { ref, onMounted, watch } from 'vue'
import StatusBadge from './StatusBadge.vue'
import CreateFTPUserForm from './CreateFTPUserForm.vue'
import { useUiStore } from '../stores/uiStore'
import { useServerStore } from '../stores/serverStore'

const props = defineProps({
  serverId: {
    type: Number,
    required: true,
  },
  onFTPUpdated: {
    type: Function,
    default: null,
  },
})

const emit = defineEmits(['ftp-updated'])

const uiStore = useUiStore()
const serverStore = useServerStore()

// Component state
const ftpCredentials = ref(null)
const isLoading = ref(false)
const showCreateModal = ref(false)
const showRegenerateModal = ref(false)
const showDeleteModal = ref(false)

// Load FTP credentials on mount
const loadFTPCredentials = async () => {
  try {
    isLoading.value = true
    const response = await serverStore.fetchFTPCredentials(props.serverId)
    ftpCredentials.value = response?.data || null
  } catch (error) {
    console.error('Error loading FTP credentials:', error)
    ftpCredentials.value = null
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  loadFTPCredentials()
})

// Handle FTP creation
const handleCreateFTP = async (username) => {
  try {
    await serverStore.createFTPUser(props.serverId, username)
    
    // Reload FTP credentials to get the complete data including host/port
    await loadFTPCredentials()
    
    uiStore.showSuccess(`FTP user "${username}" created successfully`)
    showCreateModal.value = false
    
    if (props.onFTPUpdated) {
      props.onFTPUpdated(ftpCredentials.value)
    }
    emit('ftp-updated', ftpCredentials.value)
  } catch (error) {
    console.error('Error creating FTP user:', error)
    uiStore.showError(`Failed to create FTP user: ${error.response?.data?.error || error.message}`)
  }
}

// Handle password regeneration
const handleRegeneratePassword = async () => {
  try {
    await serverStore.regenerateFTPPassword(props.serverId)
    uiStore.showSuccess('FTP password regenerated successfully')
    showRegenerateModal.value = false
  } catch (error) {
    console.error('Error regenerating FTP password:', error)
    uiStore.showError(`Failed to regenerate password: ${error.response?.data?.error || error.message}`)
  }
}

// Handle FTP user deletion
const handleDeleteFTPUser = async () => {
  try {
    await serverStore.deleteFTPUser(props.serverId)
    ftpCredentials.value = null
    uiStore.showSuccess('FTP user deleted successfully')
    showDeleteModal.value = false
    
    if (props.onFTPUpdated) {
      props.onFTPUpdated(null)
    }
    emit('ftp-updated', null)
  } catch (error) {
    console.error('Error deleting FTP user:', error)
    uiStore.showError(`Failed to delete FTP user: ${error.response?.data?.error || error.message}`)
  }
}

// Watch for server changes
watch(() => props.serverId, () => {
  loadFTPCredentials()
})
</script>

<template>
  <div class="bg-white rounded-lg shadow-md border border-slate-200">
    <!-- Header -->
    <div class="p-6 border-b border-slate-200">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <h2 class="text-lg font-bold text-slate-900">FTP Access</h2>
          <StatusBadge 
            v-if="ftpCredentials"
            status="running" 
          />
          <StatusBadge 
            v-else
            status="stopped" 
          />
        </div>
      </div>
    </div>

    <!-- Content -->
    <div class="p-6">
      <!-- Loading State -->
      <div v-if="isLoading" class="flex items-center justify-center py-8">
        <div class="flex items-center gap-2 text-slate-500">
          <div class="w-5 h-5 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
          <span>Loading FTP credentials...</span>
        </div>
      </div>

      <!-- FTP Credentials Card -->
      <div v-else-if="ftpCredentials" class="space-y-6">
        <div class="bg-slate-50 rounded-lg p-4 border border-slate-200">
          <h3 class="text-sm font-medium text-slate-700 mb-3">FTP Credentials</h3>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
            <div>
              <label class="block text-xs font-medium text-slate-500 uppercase tracking-wide">Username</label>
              <p class="mt-1 text-slate-900 font-mono">{{ ftpCredentials.username }}</p>
            </div>
            <div>
              <label class="block text-xs font-medium text-slate-500 uppercase tracking-wide">Host</label>
              <p class="mt-1 text-slate-900 font-mono">{{ ftpCredentials.host }}</p>
            </div>
            <div>
              <label class="block text-xs font-medium text-slate-500 uppercase tracking-wide">Port</label>
              <p class="mt-1 text-slate-900 font-mono">{{ ftpCredentials.port }}</p>
            </div>
            <div class="md:col-span-2">
              <label class="block text-xs font-medium text-slate-500 uppercase tracking-wide">Home Directory</label>
              <p class="mt-1 text-slate-900 font-mono break-all">{{ ftpCredentials.home_dir }}</p>
            </div>
          </div>
        </div>

        <!-- Action Buttons -->
        <div class="flex gap-3">
          <button
            @click="showRegenerateModal = true"
            :disabled="serverStore.loading"
            class="px-4 py-2 rounded-lg bg-blue-500 text-white hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium text-sm"
          >
            <span v-if="serverStore.loading" class="inline-block w-3 h-3 border-2 border-white border-t-transparent rounded-full animate-spin mr-1"></span>
            Regenerate Password
          </button>
          <button
            @click="showDeleteModal = true"
            :disabled="serverStore.loading"
            class="px-4 py-2 rounded-lg bg-red-500 text-white hover:bg-red-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium text-sm"
          >
            Delete FTP User
          </button>
        </div>
      </div>

      <!-- Empty State -->
      <div v-else class="text-center py-8">
        <div class="mb-4">
          <svg class="mx-auto h-12 w-12 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M9 19l3 3m0 0l3-3m-3 3V10" />
          </svg>
        </div>
        <h3 class="text-sm font-medium text-slate-900 mb-2">No FTP User</h3>
        <p class="text-sm text-slate-500 mb-4">
          Create an FTP user to manage your server files via FTP client.
        </p>
        <button
          @click="showCreateModal = true"
          :disabled="serverStore.loading"
          class="px-4 py-2 rounded-lg bg-green-500 text-white hover:bg-green-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium text-sm"
        >
          <span v-if="serverStore.loading" class="inline-block w-3 h-3 border-2 border-white border-t-transparent rounded-full animate-spin mr-1"></span>
          Create FTP User
        </button>
      </div>
    </div>

    <!-- Create FTP User Modal -->
    <CreateFTPUserForm
      :is-open="showCreateModal"
      :server-id="serverId"
      :is-loading="serverStore.loading"
      @close="showCreateModal = false"
      @submit="handleCreateFTP"
    />

    <!-- Regenerate Password Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showRegenerateModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div class="bg-white rounded-lg shadow-xl max-w-md w-full animate-in">
            <div class="p-6">
              <h2 class="text-xl font-bold text-slate-900 mb-4">Regenerate Password</h2>
              <p class="text-slate-600 mb-6">
                Are you sure you want to regenerate the FTP password for <strong>{{ ftpCredentials?.username }}</strong>? 
                This will invalidate the current password and you will need to use the new password for future connections.
              </p>
              
              <div class="flex gap-3 justify-end">
                <button
                  @click="showRegenerateModal = false"
                  :disabled="serverStore.loading"
                  class="px-4 py-2 rounded-lg text-slate-700 bg-slate-100 hover:bg-slate-200 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Cancel
                </button>
                <button
                  @click="handleRegeneratePassword"
                  :disabled="serverStore.loading"
                  class="px-4 py-2 rounded-lg bg-blue-500 text-white hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium"
                >
                  <span v-if="serverStore.loading" class="inline-block w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2"></span>
                  Regenerate Password
                </button>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Delete FTP User Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showDeleteModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div class="bg-white rounded-lg shadow-xl max-w-md w-full animate-in">
            <div class="p-6">
              <h2 class="text-xl font-bold text-slate-900 mb-4">Delete FTP User</h2>
              <p class="text-slate-600 mb-6">
                Are you sure you want to delete the FTP user <strong>{{ ftpCredentials?.username }}</strong>? 
                This action cannot be undone and you will no longer be able to access your server files via FTP.
              </p>
              
              <div class="flex gap-3 justify-end">
                <button
                  @click="showDeleteModal = false"
                  :disabled="serverStore.loading"
                  class="px-4 py-2 rounded-lg text-slate-700 bg-slate-100 hover:bg-slate-200 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Cancel
                </button>
                <button
                  @click="handleDeleteFTPUser"
                  :disabled="serverStore.loading"
                  class="px-4 py-2 rounded-lg bg-red-500 text-white hover:bg-red-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium"
                >
                  <span v-if="serverStore.loading" class="inline-block w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2"></span>
                  Delete User
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