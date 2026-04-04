import { describe, it, expect, beforeEach } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useUserStore } from '@/stores/userStore'

describe('userStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
  })

  it('initializeFromStorage restores userInfo and token from localStorage', () => {
    const stored = { token: 'stored-token', userId: 1, username: 'alice' }
    localStorage.setItem('user', JSON.stringify(stored))

    const store = useUserStore()
    store.initializeFromStorage()

    expect(store.userInfo).toEqual(stored)
    expect(store.token).toBe('stored-token')
  })

  it('setUserInfo updates userInfo, token, and persists to localStorage', () => {
    const store = useUserStore()
    const data = { token: 'new-token', userId: 2, username: 'bob' }

    store.setUserInfo(data)

    expect(store.userInfo).toEqual(data)
    expect(store.token).toBe('new-token')
    expect(localStorage.getItem('user')).toBe(JSON.stringify(data))
  })

  it('clearUserInfo nulls userInfo, clears token, and removes localStorage key', () => {
    const store = useUserStore()
    store.setUserInfo({ token: 'temp-token', userId: 3 })

    store.clearUserInfo()

    expect(store.userInfo).toBeNull()
    expect(store.token).toBe('')
    expect(localStorage.getItem('user')).toBeNull()
  })

  it('isAuthenticated is true only when both userInfo and token are present', () => {
    const store = useUserStore()

    expect(store.isAuthenticated).toBe(false)

    store.setUserInfo({ token: '', userId: 1 })
    expect(store.isAuthenticated).toBe(false)

    store.setUserInfo({ token: 'valid-token', userId: 1 })
    expect(store.isAuthenticated).toBe(true)

    store.clearUserInfo()
    expect(store.isAuthenticated).toBe(false)
  })
})
