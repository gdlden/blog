import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/userStore'
import Article from '@/view/Article.vue'
import Api from '@/view/Api.vue'
import Login from '@/view/Login.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      component: Api,
      meta: { requiresAuth: true },
    },
    {
      path: '/test',
      component: Article,
      meta: { requiresAuth: true },
    },
    {
      name: 'login',
      path: '/login',
      component: Login,
      meta: { requiresAuth: false },
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
    return { path: '/' }
  }

  return true
})

router.onError((error, to, from) => {
  console.error('error：' + error)
})

export default router
