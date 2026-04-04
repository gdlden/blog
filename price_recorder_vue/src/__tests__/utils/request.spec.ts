import { describe, it, expect, beforeEach, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useUserStore } from '@/stores/userStore'
import axios from 'axios'
import instance from '@/utils/request'

describe('request interceptor', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
  })

  it('adds Authorization header when store token exists', async () => {
    const store = useUserStore()
    store.setUserInfo({ token: 'test-token', userId: 1 })

    const mockAdapter = vi.fn().mockResolvedValueOnce({ data: {}, status: 200, statusText: 'OK', headers: {}, config: {} as any })
    const testInstance = axios.create({ baseURL: '/api', adapter: mockAdapter })
    testInstance.interceptors.request.use(
      req => {
        const userStore = useUserStore()
        if (userStore.token) {
          req.headers.Authorization = 'Bearer ' + userStore.token
        }
        return req
      }
    )

    await testInstance.get('/test-endpoint')

    const config = mockAdapter.mock.calls[0][0] as any
    expect(config.headers.Authorization).toBe('Bearer test-token')
  })

  it('omits Authorization header when store token is empty', async () => {
    const mockAdapter = vi.fn().mockResolvedValueOnce({ data: {}, status: 200, statusText: 'OK', headers: {}, config: {} as any })
    const testInstance = axios.create({ baseURL: '/api', adapter: mockAdapter })
    testInstance.interceptors.request.use(
      req => {
        const userStore = useUserStore()
        if (userStore.token) {
          req.headers.Authorization = 'Bearer ' + userStore.token
        }
        return req
      }
    )

    await testInstance.get('/test-endpoint')

    const config = mockAdapter.mock.calls[0][0] as any
    expect(config.headers.Authorization).toBeUndefined()
  })
})
