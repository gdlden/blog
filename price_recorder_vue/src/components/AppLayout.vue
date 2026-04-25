<script setup lang="ts">
import { ref } from 'vue'
import { useUserStore } from '@/stores/userStore'
import { useRouter, useRoute } from 'vue-router'
import { computed } from 'vue'

const userStore = useUserStore()
const router = useRouter()
const route = useRoute()
const mobileMenuOpen = ref(false)

const navItems = [
  { name: 'blog', path: '/blog', label: '博文' },
  { name: 'debt', path: '/debt', label: '债务' },
]

const currentRouteName = computed(() => route.name as string)

function handleLogout() {
  userStore.clearUserInfo()
  router.push('/login')
  mobileMenuOpen.value = false
}

function navigateTo(path: string) {
  router.push(path)
  mobileMenuOpen.value = false
}
</script>

<template>
  <div class="min-h-screen" style="background-color: #f5f5f7;">
    <!-- Navigation -->
    <nav class="sticky top-0 z-50 h-14 flex items-center px-5 md:px-8" style="background: rgba(0,0,0,0.85); backdrop-filter: saturate(180%) blur(20px); -webkit-backdrop-filter: saturate(180%) blur(20px);">
      <div class="max-w-[1100px] w-full mx-auto flex items-center justify-between">
        <!-- Logo -->
        <router-link
          to="/blog"
          class="text-white text-[15px] font-semibold tracking-tight hover:opacity-90 transition-opacity"
        >
          Blog Debt Hub
        </router-link>

        <!-- Desktop Nav -->
        <div class="hidden md:flex items-center gap-1">
          <router-link
            v-for="item in navItems"
            :key="item.name"
            :to="item.path"
            class="relative px-4 py-1.5 text-[13px] font-medium rounded-lg transition-all duration-200"
            :class="currentRouteName === item.name
              ? 'text-white bg-white/15'
              : 'text-white/70 hover:text-white hover:bg-white/10'"
          >
            {{ item.label }}
          </router-link>
        </div>

        <!-- Desktop User -->
        <div class="hidden md:flex items-center gap-3">
          <span
            v-if="userStore.userInfo?.username"
            class="text-white/50 text-[13px]"
          >
            {{ userStore.userInfo.username }}
          </span>
          <button
            @click="handleLogout"
            class="text-white/60 hover:text-white text-[13px] font-medium transition-colors"
          >
            退出
          </button>
        </div>

        <!-- Mobile Menu Button -->
        <button
          class="md:hidden text-white/80 hover:text-white p-1"
          @click="mobileMenuOpen = !mobileMenuOpen"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path v-if="!mobileMenuOpen" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"/>
            <path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
          </svg>
        </button>
      </div>
    </nav>

    <!-- Mobile Menu -->
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0 -translate-y-2"
      enter-to-class="opacity-100 translate-y-0"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100 translate-y-0"
      leave-to-class="opacity-0 -translate-y-2"
    >
      <div
        v-if="mobileMenuOpen"
        class="md:hidden fixed inset-x-0 top-14 z-40 px-4 pb-4 pt-2"
        style="background: rgba(0,0,0,0.9); backdrop-filter: saturate(180%) blur(20px);"
      >
        <div class="max-w-[1100px] mx-auto space-y-1">
          <button
            v-for="item in navItems"
            :key="item.name"
            @click="navigateTo(item.path)"
            class="w-full text-left px-4 py-2.5 rounded-xl text-[15px] font-medium transition-colors"
            :class="currentRouteName === item.name
              ? 'text-white bg-white/15'
              : 'text-white/70 hover:text-white hover:bg-white/10'"
          >
            {{ item.label }}
          </button>
          <div class="border-t border-white/10 pt-2 mt-2 flex items-center justify-between px-4">
            <span v-if="userStore.userInfo?.username" class="text-white/50 text-[13px]">
              {{ userStore.userInfo.username }}
            </span>
            <button @click="handleLogout" class="text-[#ff3b30] text-[13px] font-medium">
              退出
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Main Content -->
    <main class="min-h-[calc(100vh-56px)]">
      <router-view />
    </main>
  </div>
</template>
