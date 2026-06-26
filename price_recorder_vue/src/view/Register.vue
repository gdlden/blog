<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/userStore'
import { sendEmailCode, registerWithEmail } from '@/api/Register'

const router = useRouter()
const userStore = useUserStore()

const email = ref('')
const code = ref('')
const username = ref('')
const password = ref('')
const confirmPassword = ref('')

const isLoading = ref(false)
const isSendingCode = ref(false)
const countdown = ref(0)
const errorMessage = ref('')
const step = ref<'email' | 'register'>('email')

const codeBtnText = computed(() => {
  if (isSendingCode.value) return '发送中...'
  if (countdown.value > 0) return `${countdown.value}s`
  return '获取验证码'
})

const canSendCode = computed(() => {
  return email.value.trim() !== '' && !isSendingCode.value && countdown.value === 0
})

let countdownTimer: ReturnType<typeof setInterval> | null = null

const startCountdown = () => {
  countdown.value = 60
  if (countdownTimer) clearInterval(countdownTimer)
  countdownTimer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      if (countdownTimer) clearInterval(countdownTimer)
      countdownTimer = null
    }
  }, 1000)
}

const handleSendCode = async () => {
  if (!email.value.trim()) {
    errorMessage.value = '请输入邮箱'
    return
  }
  isSendingCode.value = true
  errorMessage.value = ''
  try {
    const res = await sendEmailCode(email.value.trim())
    if (res.flag) {
      startCountdown()
      step.value = 'register'
      errorMessage.value = ''
    } else {
      errorMessage.value = res.message || '发送失败'
    }
  } catch (error) {
    errorMessage.value = '发送验证码失败，请检查邮箱格式'
  } finally {
    isSendingCode.value = false
  }
}

const handleRegister = async () => {
  if (!email.value.trim() || !code.value.trim() || !username.value.trim() || !password.value.trim()) {
    errorMessage.value = '请填写所有字段'
    return
  }
  if (password.value !== confirmPassword.value) {
    errorMessage.value = '两次密码输入不一致'
    return
  }
  if (password.value.length < 6) {
    errorMessage.value = '密码长度不能少于 6 位'
    return
  }

  isLoading.value = true
  errorMessage.value = ''
  try {
    const response = await registerWithEmail({
      email: email.value.trim(),
      code: code.value.trim(),
      username: username.value.trim(),
      password: password.value,
    })
    if (response && response.token) {
      userStore.setUserInfo(response)
      router.push('/')
    } else {
      errorMessage.value = '注册失败，请重试'
    }
  } catch (error) {
    console.error('Register error:', error)
    errorMessage.value = '注册失败，请检查信息后重试'
  } finally {
    isLoading.value = false
  }
}

const goToLogin = () => {
  router.push('/login')
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center px-4" style="background-color: #f5f5f7;">
    <div class="w-full max-w-sm bg-white rounded-xl p-10" style="border-radius: 12px;">
      <h1 class="text-center mb-8" style="font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', 'SF Pro Icons', 'Helvetica Neue', Helvetica, Arial, sans-serif; font-size: 28px; font-weight: 600; line-height: 1.14; letter-spacing: 0.196px; color: #1d1d1f;">
        注册
      </h1>

      <div class="space-y-5">
        <!-- 邮箱输入 -->
        <div>
          <label class="block mb-1.5" style="font-size: 14px; font-weight: 400; letter-spacing: -0.224px; line-height: 1.43; color: rgba(0, 0, 0, 0.8);">
            邮箱
          </label>
          <div class="flex gap-2">
            <input
              type="email"
              v-model="email"
              placeholder="your@email.com"
              class="flex-1 px-3 py-2.5 text-base outline-none transition-colors"
              style="background-color: #fafafc; border-radius: 8px; border: 1px solid rgba(0, 0, 0, 0.04); font-size: 17px; letter-spacing: -0.374px; color: #1d1d1f;"
              :disabled="step === 'register'"
            />
            <button
              type="button"
              @click="handleSendCode"
              :disabled="!canSendCode"
              class="px-3 py-2.5 text-white transition-colors whitespace-nowrap disabled:opacity-50"
              style="background-color: #0071e3; border-radius: 8px; font-size: 14px; font-weight: 400; line-height: 1;"
            >
              {{ codeBtnText }}
            </button>
          </div>
        </div>

        <!-- 验证码 -->
        <div v-if="step === 'register'">
          <label class="block mb-1.5" style="font-size: 14px; font-weight: 400; letter-spacing: -0.224px; line-height: 1.43; color: rgba(0, 0, 0, 0.8);">
            验证码
          </label>
          <input
            type="text"
            v-model="code"
            placeholder="请输入 6 位验证码"
            class="w-full px-3 py-2.5 text-base outline-none transition-colors"
            style="background-color: #fafafc; border-radius: 8px; border: 1px solid rgba(0, 0, 0, 0.04); font-size: 17px; letter-spacing: -0.374px; color: #1d1d1f;"
          />
        </div>

        <!-- 用户名 (注册阶段显示) -->
        <div v-if="step === 'register'">
          <label class="block mb-1.5" style="font-size: 14px; font-weight: 400; letter-spacing: -0.224px; line-height: 1.43; color: rgba(0, 0, 0, 0.8);">
            用户名
          </label>
          <input
            type="text"
            v-model="username"
            placeholder="请输入用户名"
            class="w-full px-3 py-2.5 text-base outline-none transition-colors"
            style="background-color: #fafafc; border-radius: 8px; border: 1px solid rgba(0, 0, 0, 0.04); font-size: 17px; letter-spacing: -0.374px; color: #1d1d1f;"
          />
        </div>

        <!-- 密码 (注册阶段显示) -->
        <div v-if="step === 'register'">
          <label class="block mb-1.5" style="font-size: 14px; font-weight: 400; letter-spacing: -0.224px; line-height: 1.43; color: rgba(0, 0, 0, 0.8);">
            密码
          </label>
          <input
            type="password"
            v-model="password"
            placeholder="至少 6 位"
            class="w-full px-3 py-2.5 text-base outline-none transition-colors"
            style="background-color: #fafafc; border-radius: 8px; border: 1px solid rgba(0, 0, 0, 0.04); font-size: 17px; letter-spacing: -0.374px; color: #1d1d1f;"
          />
        </div>

        <!-- 确认密码 (注册阶段显示) -->
        <div v-if="step === 'register'">
          <label class="block mb-1.5" style="font-size: 14px; font-weight: 400; letter-spacing: -0.224px; line-height: 1.43; color: rgba(0, 0, 0, 0.8);">
            确认密码
          </label>
          <input
            type="password"
            v-model="confirmPassword"
            placeholder="再次输入密码"
            class="w-full px-3 py-2.5 text-base outline-none transition-colors"
            style="background-color: #fafafc; border-radius: 8px; border: 1px solid rgba(0, 0, 0, 0.04); font-size: 17px; letter-spacing: -0.374px; color: #1d1d1f;"
          />
        </div>

        <p v-if="errorMessage" class="text-sm" style="color: #ff3b30;">
          {{ errorMessage }}
        </p>

        <div class="flex items-center justify-center gap-3 pt-2">
          <button
            type="button"
            @click="goToLogin"
            class="px-4 py-2 text-white transition-colors"
            style="background-color: #1d1d1f; border-radius: 8px; font-size: 17px; font-weight: 400; line-height: 2.41; letter-spacing: normal;"
          >
            返回登录
          </button>
          <button
            v-if="step === 'register'"
            type="button"
            @click="handleRegister"
            :disabled="isLoading"
            class="px-4 py-2 text-white transition-colors disabled:opacity-50"
            style="background-color: #0071e3; border-radius: 8px; font-size: 17px; font-weight: 400; line-height: 2.41; letter-spacing: normal;"
          >
            {{ isLoading ? '注册中...' : '注册' }}
          </button>
        </div>

        <div v-if="step === 'email'" class="text-center pt-2">
          <span style="font-size: 14px; color: #999;">已有账号？</span>
          <button type="button" @click="goToLogin" style="font-size: 14px; color: #0071e3; background: none; border: none; cursor: pointer; text-decoration: underline;">
            去登录
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
