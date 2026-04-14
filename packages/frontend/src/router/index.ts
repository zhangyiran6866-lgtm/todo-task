import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/use-auth-store'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/home/index.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/auth/LoginView.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/views/auth/RegisterView.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/tasks',
      name: 'tasks',
      component: () => import('@/views/tasks/TasksView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/tasks/:id',
      name: 'task-detail',
      component: () => import('@/views/tasks/TaskDetailView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/profile',
      name: 'profile',
      component: () => import('@/views/profile/ProfileView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/views/NotFoundView.vue')
    }
  ]
})

// 路由守卫：未登录跳转 /login
router.beforeEach((to) => {
  const authStore = useAuthStore()
  if (to.meta.requiresAuth && !authStore.isLoggedIn) {
    return { name: 'login' }
  }
  // 已登录访问 login/register 跳转到 tasks
  if (!to.meta.requiresAuth && (to.name === 'login' || to.name === 'register') && authStore.isLoggedIn) {
    return { name: 'tasks' }
  }
})

export default router
