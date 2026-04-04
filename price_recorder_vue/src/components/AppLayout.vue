<script setup lang="ts">
import { useUserStore } from '@/stores/userStore'
import { useRouter } from 'vue-router'

const userStore = useUserStore()
const router = useRouter()

function handleLogout() {
  userStore.clearUserInfo()
  router.push('/login')
}
</script>

<template>
  <div class="flex h-screen bg-white">
    <aside class="w-56 bg-gray-100 flex flex-col">
      <div class="p-4 border-b border-gray-200">
        <h1 class="text-heading font-semibold">Blog Debt Hub</h1>
      </div>
      <nav class="flex-1 p-2 space-y-1">
        <router-link
          to="/blog"
          class="block px-3 py-2 rounded text-body hover:bg-gray-200"
          active-class="bg-blue-600 text-white hover:bg-blue-600"
        >
          博文
        </router-link>
        <router-link
          to="/debt"
          class="block px-3 py-2 rounded text-body hover:bg-gray-200"
          active-class="bg-blue-600 text-white hover:bg-blue-600"
        >
          债务
        </router-link>
      </nav>
      <div class="p-4 border-t border-gray-200">
        <div v-if="userStore.userInfo" class="text-label mb-2 truncate">
          {{ userStore.userInfo.username || '' }}
        </div>
        <button
          type="button"
          @click="handleLogout"
          class="w-full px-3 py-2 rounded bg-red-600 text-white hover:bg-red-700 text-label"
        >
          退出登录
        </button>
      </div>
    </aside>
    <main class="flex-1 p-6 overflow-auto">
      <router-view />
    </main>
  </div>
</template>

<style scoped>
.text-heading {
  font-size: 20px;
  line-height: 1.2;
}
.text-body {
  font-size: 16px;
  line-height: 1.5;
}
.text-label {
  font-size: 14px;
  line-height: 1.4;
}
</style>
