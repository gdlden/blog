<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/userStore'
import { login, user } from '@/api/Login'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const isLoading = ref(false)
const errorMessage = ref('')

const loginAction = async () => {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const response = await login(user)
    if (response && response.token) {
      userStore.setUserInfo(response)
      const redirect = route.query.redirect as string
      router.push(redirect || '/')
    } else {
      errorMessage.value = '登录失败，请检查用户名或密码后重试。'
    }
  } catch (error) {
    console.error('Login error:', error)
    errorMessage.value = '登录失败，请检查用户名或密码后重试。'
  } finally {
    isLoading.value = false
  }
}

const cancel = () => {
  user.username = ''
  user.password = ''
  errorMessage.value = ''
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center px-4" style="background-color: #f5f5f7;">
    <div class="w-full max-w-sm bg-white rounded-xl p-10" style="border-radius: 12px;">
      <h1 class="text-center mb-8" style="font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', 'SF Pro Icons', 'Helvetica Neue', Helvetica, Arial, sans-serif; font-size: 28px; font-weight: 600; line-height: 1.14; letter-spacing: 0.196px; color: #1d1d1f;">
        登录
      </h1>

      <div class="space-y-5">
        <div>
          <label class="block mb-1.5" style="font-size: 14px; font-weight: 400; letter-spacing: -0.224px; line-height: 1.43; color: rgba(0, 0, 0, 0.8);">
            用户名
          </label>
          <input
            type="text"
            v-model="user.username"
            class="w-full px-3 py-2.5 text-base outline-none transition-colors"
            style="background-color: #fafafc; border-radius: 8px; border: 1px solid rgba(0, 0, 0, 0.04); font-size: 17px; letter-spacing: -0.374px; color: #1d1d1f;"
            @keyup.enter="loginAction"
          />
        </div>

        <div>
          <label class="block mb-1.5" style="font-size: 14px; font-weight: 400; letter-spacing: -0.224px; line-height: 1.43; color: rgba(0, 0, 0, 0.8);">
            密码
          </label>
          <input
            type="password"
            v-model="user.password"
            class="w-full px-3 py-2.5 text-base outline-none transition-colors"
            style="background-color: #fafafc; border-radius: 8px; border: 1px solid rgba(0, 0, 0, 0.04); font-size: 17px; letter-spacing: -0.374px; color: #1d1d1f;"
            @keyup.enter="loginAction"
          />
        </div>

        <p v-if="errorMessage" class="text-sm" style="color: #ff3b30;">
          {{ errorMessage }}
        </p>

        <div class="flex items-center justify-center gap-3 pt-2">
          <button
            type="button"
            @click="cancel"
            class="px-4 py-2 text-white transition-colors"
            style="background-color: #1d1d1f; border-radius: 8px; font-size: 17px; font-weight: 400; line-height: 2.41; letter-spacing: normal;"
          >
            重置
          </button>
          <button
            type="button"
            @click="loginAction"
            :disabled="isLoading"
            class="px-4 py-2 text-white transition-colors disabled:opacity-50"
            style="background-color: #0071e3; border-radius: 8px; font-size: 17px; font-weight: 400; line-height: 2.41; letter-spacing: normal;"
          >
            {{ isLoading ? '登录中...' : '登录' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
