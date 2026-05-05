import { describe, it, expect, beforeEach } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import router from '@/router/index'
import { useUserStore } from '@/stores/userStore'

describe('router guards', () => {
  beforeEach(async () => {
    setActivePinia(createPinia())
    localStorage.clear()
    await router.push({ name: 'login', query: {} })
    await router.isReady()
  })

  it('redirects unauthenticated user to /login with redirect query', async () => {
    await router.push('/blog')
    await router.isReady()

    expect(router.currentRoute.value.name).toBe('login')
    expect(router.currentRoute.value.query.redirect).toBe('/blog')
  })

  it('redirects authenticated user away from /login', async () => {
    const store = useUserStore()
    store.setUserInfo({ token: 'test', userId: 1 })

    await router.push('/blog')
    await router.isReady()
    await router.push('/login')
    await router.isReady()

    expect(router.currentRoute.value.name).toBe('blog')
  })

  it('allows authenticated user to access protected routes', async () => {
    const store = useUserStore()
    store.setUserInfo({ token: 'test', userId: 1 })

    await router.push('/blog')
    await router.isReady()

    expect(router.currentRoute.value.path).toBe('/blog')
  })

  it('allows authenticated user to access fuel routes', async () => {
    const store = useUserStore()
    store.setUserInfo({ token: 'test', userId: 1 })

    await router.push('/fuel')
    await router.isReady()
    expect(router.currentRoute.value.path).toBe('/fuel')

    await router.push('/fuel/42')
    await router.isReady()
    expect(router.currentRoute.value.name).toBe('fuelDetail')
    expect(router.currentRoute.value.params.vehicleId).toBe('42')
  })

  it('initializes store from localStorage when isAuthenticated is false', async () => {
    const stored = { token: 'stored-token', userId: 2 }
    localStorage.setItem('user', JSON.stringify(stored))

    const store = useUserStore()
    expect(store.isAuthenticated).toBe(false)

    await router.push('/debt')
    await router.isReady()

    expect(store.isAuthenticated).toBe(true)
    expect(router.currentRoute.value.path).toBe('/debt')
  })
})
