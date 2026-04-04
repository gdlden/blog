import { describe, it, expect, beforeEach, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useUserStore } from '@/stores/userStore'
import axios, { type InternalAxiosRequestConfig } from 'axios'

describe('request interceptor', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    vi.resetModules()
  })

  it('adds Authorization header when store token exists', async () => {
    const store = useUserStore()
    store.setUserInfo({ token: 'test-token', userId: 1 })

    const instance = axios.create({ baseURL: '/api' })
    const { useUserStore: getUserStore } = await import('@/stores/userStore')
    instance.interceptors.request.use(
      req => {
        const userStore = getUserStore()
        if (userStore.token) {
          req.headers.Authorization = 'Bearer ' + userStore.token
        }
        return req
      }
    )

    const dummyConfig = { headers: {} } as InternalAxiosRequestConfig
    const handler = instance.interceptors.request.handlers[0]
    const result = handler.fulfilled(dummyConfig)

    expect(result.headers.Authorization).toBe('Bearer test-token')
  })

  it('omits Authorization header when store token is empty', async () => {
    const instance = axios.create({ baseURL: '/api' })
    const { useUserStore: getUserStore } = await import('@/stores/userStore')
    instance.interceptors.request.use(
      req => {
        const userStore = getUserStore()
        if (userStore.token) {
          req.headers.Authorization = 'Bearer ' + userStore.token
        }
        return req
      }
    )

    const dummyConfig = { headers: {} } as InternalAxiosRequestConfig
    const handler = instance.interceptors.request.handlers[0]
    const result = handler.fulfilled(dummyConfig)

    expect(result.headers.Authorization).toBeUndefined()
  })
})
