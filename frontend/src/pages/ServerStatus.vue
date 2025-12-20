<script setup>
import { ref, onMounted, computed } from 'vue'
import { useServerStore } from '../stores/serverStore'
import { useUiStore } from '../stores/uiStore'
import ServerCard from '../components/ServerCard.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'

const serverStore = useServerStore()
const uiStore = useUiStore()

const refreshing = ref(false)

// Mock data for demonstration (will be replaced by API calls)
const mockServers = [
  {
    id: '1',
    name: 'Main Server',
    status: 'running',
    players: 32,
    maxPlayers: 60,
    port: 2302,
    createdAt: new Date().toISOString(),
  },
  {
    id: '2',
    name: 'Test Server',
    status: 'stopped',
    players: 0,
    maxPlayers: 60,
    port: 2303,
    createdAt: new Date().toISOString(),
  },
  {
    id: '3',
    name: 'Backup Server',
    status: 'running',
    players: 15,
    maxPlayers: 60,
    port: 2304,
    createdAt: new Date().toISOString(),
  },
]

onMounted(async () => {
  await fetchServers()
})

const fetchServers = async () => {
  try {
    refreshing.value = true
    
    // Try to fetch from API
    try {
      await serverStore.fetchServers()
    } catch (error) {
      // Fall back to mock data if API is not ready
      console.log('Using mock server data:', error)
      serverStore.servers = mockServers
    }
  } catch (error) {
    uiStore.showError('Failed to load servers')
    console.error('Error fetching servers:', error)
    // Use mock data as fallback
    serverStore.servers = mockServers
  } finally {
    refreshing.value = false
  }
}

const runningCount = computed(() => {
  return serverStore.servers.filter(s => s.status === 'running').length
})

const stoppedCount = computed(() => {
  return serverStore.servers.filter(s => s.status === 'stopped').length
})

const handleServerAction = async () => {
  // Refresh server list after action
  await fetchServers()
}
</script>

<template>
  <div>
    <!-- Page Header -->
    <div class="mb-8">
      <h1 class="text-4xl font-bold text-slate-900 mb-2">Server Status</h1>
      <p class="text-slate-600">Manage and monitor your DayZ servers</p>
    </div>

    <!-- Stats Section -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
      <div class="bg-white rounded-lg shadow-md p-6 border-l-4 border-blue-500">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-slate-600 text-sm font-medium mb-1">Total Servers</p>
            <p class="text-3xl font-bold text-slate-900">{{ serverStore.servers.length }}</p>
          </div>
          <div class="text-3xl">🖥️</div>
        </div>
      </div>

      <div class="bg-white rounded-lg shadow-md p-6 border-l-4 border-green-500">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-slate-600 text-sm font-medium mb-1">Running</p>
            <p class="text-3xl font-bold text-slate-900">{{ runningCount }}</p>
          </div>
          <div class="text-3xl">●</div>
        </div>
      </div>

      <div class="bg-white rounded-lg shadow-md p-6 border-l-4 border-slate-400">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-slate-600 text-sm font-medium mb-1">Stopped</p>
            <p class="text-3xl font-bold text-slate-900">{{ stoppedCount }}</p>
          </div>
          <div class="text-3xl">◯</div>
        </div>
      </div>
    </div>

    <!-- SteamCMD Status Section -->
    <div class="bg-white rounded-lg shadow-md p-6 mb-8 border-l-4 border-yellow-500">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-xl font-bold text-slate-900 mb-2">SteamCMD Status</h2>
          <p class="text-slate-600">Integration ready for when API is available</p>
        </div>
        <div class="text-4xl">⚙️</div>
      </div>
    </div>

    <!-- Refresh Button -->
    <div class="flex justify-between items-center mb-6">
      <h2 class="text-2xl font-bold text-slate-900">Your Servers</h2>
      <button
        @click="fetchServers"
        :disabled="refreshing || serverStore.loading"
        class="px-4 py-2 rounded-lg bg-blue-500 text-white hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium"
      >
        <span v-if="refreshing || serverStore.loading" class="inline-block w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2"></span>
        {{ refreshing || serverStore.loading ? 'Refreshing...' : 'Refresh' }}
      </button>
    </div>

    <!-- Servers Grid -->
    <div v-if="serverStore.servers.length === 0" class="bg-white rounded-lg shadow-md p-12 text-center">
      <p class="text-slate-500 text-lg">No servers found. Create your first server to get started.</p>
    </div>
    
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <ServerCard
        v-for="server in serverStore.servers"
        :key="server.id"
        :server="server"
        @server-action="handleServerAction"
      />
    </div>
  </div>
</template>

<style scoped>
</style>
