import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { apiClient } from '../utils/apiClient'

export const useServerStore = defineStore('servers', () => {
  const servers = ref([])
  const currentServer = ref(null)
  const loading = ref(false)
  const error = ref(null)

  const serverCount = computed(() => servers.value.length)
  const runningServers = computed(() => servers.value.filter(s => s.status === 'running').length)

  const fetchServers = async () => {
    try {
      loading.value = true
      error.value = null
      
      const response = await apiClient.get('/servers')
      servers.value = response.data || []
      
      return servers.value
    } catch (err) {
      error.value = err.response?.data?.error || 'Failed to fetch servers'
      console.error('Error fetching servers:', err)
    } finally {
      loading.value = false
    }
  }

  const fetchServerDetails = async (serverId) => {
    try {
      loading.value = true
      error.value = null
      
      const response = await apiClient.get(`/servers/${serverId}`)
      currentServer.value = response.data
      
      return response.data
    } catch (err) {
      error.value = err.response?.data?.error || 'Failed to fetch server details'
      throw err
    } finally {
      loading.value = false
    }
  }

  const fetchServerStatus = async (serverId) => {
    try {
      const response = await apiClient.get(`/servers/${serverId}/status`)
      
      // Update server status in list
      const server = servers.value.find(s => s.id === serverId)
      if (server) {
        server.status = response.data.status
      }
      
      // Update current server if it matches
      if (currentServer.value?.id === serverId) {
        currentServer.value.status = response.data.status
      }
      
      return response.data
    } catch (err) {
      error.value = err.response?.data?.error || 'Failed to fetch server status'
      throw err
    }
  }

  const startServer = async (serverId) => {
    try {
      loading.value = true
      error.value = null
      
      const response = await apiClient.post(`/servers/${serverId}/start`)
      
      // Update server status
      const server = servers.value.find(s => s.id === serverId)
      if (server) {
        server.status = 'running'
      }
      if (currentServer.value?.id === serverId) {
        currentServer.value.status = 'running'
      }
      
      return response.data
    } catch (err) {
      error.value = err.response?.data?.error || 'Failed to start server'
      throw err
    } finally {
      loading.value = false
    }
  }

  const stopServer = async (serverId) => {
    try {
      loading.value = true
      error.value = null
      
      const response = await apiClient.post(`/servers/${serverId}/stop`)
      
      // Update server status
      const server = servers.value.find(s => s.id === serverId)
      if (server) {
        server.status = 'stopped'
      }
      if (currentServer.value?.id === serverId) {
        currentServer.value.status = 'stopped'
      }
      
      return response.data
    } catch (err) {
      error.value = err.response?.data?.error || 'Failed to stop server'
      throw err
    } finally {
      loading.value = false
    }
  }

  const restartServer = async (serverId) => {
    try {
      loading.value = true
      error.value = null
      
      const response = await apiClient.post(`/servers/${serverId}/restart`)
      
      // Update server status
      const server = servers.value.find(s => s.id === serverId)
      if (server) {
        server.status = 'running'
      }
      if (currentServer.value?.id === serverId) {
        currentServer.value.status = 'running'
      }
      
      return response.data
    } catch (err) {
      error.value = err.response?.data?.error || 'Failed to restart server'
      throw err
    } finally {
      loading.value = false
    }
  }

  const setCurrentServer = (server) => {
    currentServer.value = server
  }

  // FTP Methods
  const fetchFTPCredentials = async (serverId) => {
    try {
      const response = await apiClient.get(`/servers/${serverId}/ftp/credentials`)
      return response.data
    } catch (err) {
      if (err.response?.status === 404) {
        return null // FTP user doesn't exist
      }
      error.value = err.response?.data?.error || 'Failed to fetch FTP credentials'
      throw err
    }
  }

  const createFTPUser = async (serverId, username) => {
    try {
      loading.value = true
      error.value = null
      
      const response = await apiClient.post(`/servers/${serverId}/ftp/create`, {
        username
      })
      
      return response.data
    } catch (err) {
      error.value = err.response?.data?.error || 'Failed to create FTP user'
      throw err
    } finally {
      loading.value = false
    }
  }

  const regenerateFTPPassword = async (serverId) => {
    try {
      loading.value = true
      error.value = null
      
      const response = await apiClient.post(`/servers/${serverId}/ftp/regenerate-password`)
      
      return response.data
    } catch (err) {
      error.value = err.response?.data?.error || 'Failed to regenerate FTP password'
      throw err
    } finally {
      loading.value = false
    }
  }

  const deleteFTPUser = async (serverId) => {
    try {
      loading.value = true
      error.value = null
      
      const response = await apiClient.delete(`/servers/${serverId}/ftp/user`)
      
      return response.data
    } catch (err) {
      error.value = err.response?.data?.error || 'Failed to delete FTP user'
      throw err
    } finally {
      loading.value = false
    }
  }

  return {
    servers,
    currentServer,
    loading,
    error,
    serverCount,
    runningServers,
    fetchServers,
    fetchServerDetails,
    fetchServerStatus,
    startServer,
    stopServer,
    restartServer,
    setCurrentServer,
    fetchFTPCredentials,
    createFTPUser,
    regenerateFTPPassword,
    deleteFTPUser,
  }
})
