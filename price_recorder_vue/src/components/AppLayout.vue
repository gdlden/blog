<script setup lang="ts">
import { useUserStore } from '@/stores/userStore'
import { useRouter, useRoute } from 'vue-router'
import { computed } from 'vue'

const userStore = useUserStore()
const router = useRouter()
const route = useRoute()

const navItems = [
  { name: 'blog', path: '/blog', label: '博文' },
  { name: 'debt', path: '/debt', label: '债务' },
]

const currentRouteName = computed(() => route.name as string)

function handleLogout() {
  userStore.clearUserInfo()
  router.push('/login')
}
</script>

<template>
  <div class="min-h-screen bg-apple-gray">
    <!-- Apple Glass Navigation -->
    <nav class="sticky top-0 z-50 h-12 flex items-center px-6" style="background: rgba(0,0,0,0.8); backdrop-filter: saturate(180%) blur(20px); -webkit-backdrop-filter: saturate(180%) blur(20px);">
      <div class="max-w-[980px] w-full mx-auto flex items-center justify-between">
        <!-- Logo / Brand -->
        <router-link
          to="/blog"
          class="text-white text-sm font-semibold tracking-tight"
          style="font-size: 14px; letter-spacing: -0.224px;"
        >
          Blog Debt Hub
        </router-link>

        <!-- Nav Links -->
        <div class="flex items-center gap-8">
          <router-link
            v-for="item in navItems"
            :key="item.name"
            :to="item.path"
            class="text-white/80 hover:text-white transition-colors"
            style="font-size: 12px; font-weight: 400;"
            :class="{ 'text-white': currentRouteName === item.name }"
          >
            {{ item.label }}
          </router-link>
        </div>

        <!-- User & Logout -->
        <div class="flex items-center gap-4">
          <span
            v-if="userStore.userInfo"
            class="text-white/60 truncate max-w-[120px]"
            style="font-size: 12px;"
          >
            {{ userStore.userInfo.username || '' }}
          </span>
          <button
            type="button"
            @click="handleLogout"
            class="text-white/80 hover:text-white transition-colors"
            style="font-size: 12px;"
          >
            退出
          </button>
        </div>
      </div>
    </nav>

    <!-- Main Content -->
    <main class="min-h-[calc(100vh-48px)]">
      <router-view />
    </main>
  </div>
</template>
