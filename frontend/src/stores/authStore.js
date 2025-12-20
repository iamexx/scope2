import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { apiClient } from '../utils/apiClient'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || null)
  const user = ref(null)
  const isFirstRun = ref(false)
  const loading = ref(false)
  const error = ref(null)

  const isAuthenticated = computed(() => !!token.value)

  const setToken = (newToken) => {
    token.value = newToken
    if (newToken) {
      localStorage.setItem('token', newToken)
    } else {
      localStorage.removeItem('token')
    }
  }

  const setUser = (userData) => {
    user.value = userData
  }

  const clearAuth = () => {
    token.value = null
    user.value = null
    localStorage.removeItem('token')
  }

  const checkFirstRun = async () => {
    try {
      loading.value = true
      error.value = null
      
      // Try to fetch /api/auth/me - if we get 401, it's likely first run
      // If we can't make any authenticated request, assume first run
      const response = await apiClient.get('/auth/me')
      isFirstRun.value = false
      setUser(response.data)
    } catch (err) {
      // On error, check if there's a token - if not, could be first run
      isFirstRun.value = !token.value
    } finally {
      loading.value = false
    }
  }

  const setup = async (username, password) => {
    try {
      loading.value = true
      error.value = null
      
      const response = await apiClient.post('/auth/setup', {
        username,
        password,
      })
      
      setToken(response.data.token)
      setUser({ username })
      isFirstRun.value = false
      
      return response.data
    } catch (err) {
      error.value = err.response?.data?.error || 'Setup failed'
      throw err
    } finally {
      loading.value = false
    }
  }

  const login = async (username, password) => {
    try {
      loading.value = true
      error.value = null
      
      const response = await apiClient.post('/auth/login', {
        username,
        password,
      })
      
      setToken(response.data.token)
      setUser({ username })
      
      return response.data
    } catch (err) {
      error.value = err.response?.data?.error || 'Login failed'
      throw err
    } finally {
      loading.value = false
    }
  }

  const fetchUser = async () => {
    try {
      loading.value = true
      error.value = null
      
      const response = await apiClient.get('/auth/me')
      setUser(response.data)
      
      return response.data
    } catch (err) {
      error.value = err.response?.data?.error || 'Failed to fetch user'
      clearAuth()
      throw err
    } finally {
      loading.value = false
    }
  }

  const logout = () => {
    clearAuth()
    isFirstRun.value = false
  }

  return {
    token,
    user,
    isFirstRun,
    loading,
    error,
    isAuthenticated,
    setToken,
    setUser,
    clearAuth,
    checkFirstRun,
    setup,
    login,
    fetchUser,
    logout,
  }
})
