import { ref, computed } from 'vue'
import { defineStore } from 'pinia'

export interface UserInfo {
  token: string
  userId?: number | string
  username?: string
  [key: string]: any
}

export const useUserStore = defineStore('user', () => {
  const userInfo = ref<UserInfo | null>(null)
  const token = ref<string>('')

  const isAuthenticated = computed(() => {
    return !!userInfo.value && !!token.value
  })

  function setUserInfo(data: UserInfo) {
    userInfo.value = data
    token.value = data.token || ''
    localStorage.setItem('user', JSON.stringify(data))
  }

  function clearUserInfo() {
    userInfo.value = null
    token.value = ''
    localStorage.removeItem('user')
  }

  function initializeFromStorage() {
    const stored = localStorage.getItem('user')
    if (stored) {
      try {
        const parsed: UserInfo = JSON.parse(stored)
        userInfo.value = parsed
        token.value = parsed.token || ''
      } catch (e) {
        localStorage.removeItem('user')
      }
    }
  }

  return {
    userInfo,
    token,
    isAuthenticated,
    setUserInfo,
    clearUserInfo,
    initializeFromStorage,
  }
})
