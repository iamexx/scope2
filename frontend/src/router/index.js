import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/authStore'

// Import pages
import LoginPage from '../pages/LoginPage.vue'
import FirstRunSetup from '../pages/FirstRunSetup.vue'
import ServerStatus from '../pages/ServerStatus.vue'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: LoginPage,
    meta: { layout: 'auth' },
  },
  {
    path: '/setup',
    name: 'FirstRunSetup',
    component: FirstRunSetup,
    meta: { layout: 'auth' },
  },
  {
    path: '/',
    name: 'ServerStatus',
    component: ServerStatus,
    meta: { requiresAuth: true, layout: 'main' },
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: () => {
      const authStore = useAuthStore()
      return authStore.isAuthenticated ? '/' : '/login'
    },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()

  // If trying to access protected route without authentication
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next('/login')
    return
  }

  // If logged in and trying to access login/setup, redirect to dashboard
  if ((to.name === 'Login' || to.name === 'FirstRunSetup') && authStore.isAuthenticated) {
    next('/')
    return
  }

  next()
})

export default router
