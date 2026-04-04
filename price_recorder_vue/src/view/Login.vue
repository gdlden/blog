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
  <div class="flex items-center justify-center h-screen bg-gray-100">
    <div class="bg-white rounded-lg shadow-md p-8 w-full max-w-sm">
      <h1 class="text-display font-semibold text-center mb-6">登录</h1>
      <div class="space-y-4">
        <div>
          <label class="block text-label mb-1">用户名</label>
          <input
            type="text"
            v-model="user.username"
            class="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-600"
          />
        </div>
        <div>
          <label class="block text-label mb-1">密码</label>
          <input
            type="password"
            v-model="user.password"
            class="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-600"
          />
        </div>
        <p v-if="errorMessage" class="text-sm text-red-600">{{ errorMessage }}</p>
        <div class="flex items-center justify-center gap-4 pt-2">
          <button
            type="button"
            @click="cancel"
            class="px-4 py-2 rounded border border-gray-300 text-gray-700 hover:bg-gray-50"
          >
            重置
          </button>
          <button
            type="button"
            @click="loginAction"
            :disabled="isLoading"
            class="px-4 py-2 rounded bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50"
          >
            {{ isLoading ? '登录中...' : '登录' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.text-display {
  font-size: 28px;
  line-height: 1.2;
}
.text-label {
  font-size: 14px;
  line-height: 1.4;
}
</style>
