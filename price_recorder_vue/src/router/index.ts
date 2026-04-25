import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/userStore'
import Login from '@/view/Login.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      name: 'login',
      path: '/login',
      component: Login,
      meta: { requiresAuth: false },
    },
    {
      path: '/',
      component: () => import('@/components/AppLayout.vue'),
      meta: { requiresAuth: true },
      redirect: { name: 'blog' },
      children: [
        {
          name: 'blog',
          path: 'blog',
          component: () => import('@/view/BlogList.vue'),
          meta: { requiresAuth: true },
        },
        {
          name: 'debt',
          path: 'debt',
          component: () => import('@/view/DebtList.vue'),
          meta: { requiresAuth: true },
        },
        {
          name: 'debtDetail',
          path: 'debt/:id',
          component: () => import('@/view/DebtDetail.vue'),
          meta: { requiresAuth: true },
        },
      ],
    },
  ],
})

router.beforeEach((to, from) => {
  const userStore = useUserStore()

  if (!userStore.isAuthenticated) {
    userStore.initializeFromStorage()
  }

  if (to.meta.requiresAuth && !userStore.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  if (to.name === 'login' && userStore.isAuthenticated) {
    return { name: 'blog' }
  }

  return true
})

router.onError((error, to, from) => {
  console.error('error：' + error)
})

export default router
