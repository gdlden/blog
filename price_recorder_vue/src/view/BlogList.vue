<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useBlogStore } from '@/stores/blogStore'

const blogStore = useBlogStore()
const { posts, loading, totalPages, currentPage } = storeToRefs(blogStore)

const showModal = ref(false)
const isEditing = ref(false)
const formData = ref({ id: '', title: '', content: '' })
const isSubmitting = ref(false)

onMounted(() => {
  blogStore.fetchPosts()
})

function openCreateModal() {
  isEditing.value = false
  formData.value = { id: '', title: '', content: '' }
  showModal.value = true
}

function openEditModal(post: { id: string; title: string; content: string }) {
  isEditing.value = true
  formData.value = { ...post }
  showModal.value = true
}

async function handleSubmit() {
  if (!formData.value.title.trim()) return
  isSubmitting.value = true
  try {
    if (isEditing.value) {
      await blogStore.updatePost(formData.value.id, {
        title: formData.value.title,
        content: formData.value.content
      })
    } else {
      await blogStore.createPost({
        title: formData.value.title,
        content: formData.value.content
      })
    }
    showModal.value = false
    formData.value = { id: '', title: '', content: '' }
  } catch (err: any) {
    alert(err.message || '操作失败')
  } finally {
    isSubmitting.value = false
  }
}

async function handleDelete(id: string) {
  if (confirm('确定要删除这篇博文吗？')) {
    try {
      await blogStore.deletePost(id)
    } catch (err: any) {
      alert(err.message || '删除失败')
    }
  }
}

function changePage(page: number) {
  if (page >= 1 && page <= totalPages.value) {
    blogStore.fetchPosts(page)
  }
}

function closeModal() {
  showModal.value = false
  formData.value = { id: '', title: '', content: '' }
}

function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}
</script>

<template>
  <div class="max-w-[1100px] mx-auto px-5 py-10">
    <!-- Header -->
    <div class="flex justify-between items-center mb-8">
      <div>
        <h1 class="text-[32px] font-semibold tracking-tight text-[#1d1d1f]">
          博文
        </h1>
        <p class="mt-1 text-sm text-[#86868b]">
          管理和分享你的文章
        </p>
      </div>
      <button
        @click="openCreateModal"
        class="inline-flex items-center gap-2 px-5 py-2.5 text-white text-[15px] font-medium rounded-xl transition-all duration-200 hover:scale-[1.02] active:scale-[0.98] shadow-sm"
        style="background: linear-gradient(135deg, #0071e3, #0063c7);"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
        </svg>
        新建博文
      </button>
    </div>

    <!-- Skeleton Loading -->
    <div v-if="loading" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
      <div v-for="n in 6" :key="n" class="bg-white rounded-2xl p-5 border border-[#f0f0f0]">
        <div class="h-5 bg-[#f5f5f7] rounded-lg w-3/4 mb-3 animate-pulse"/>
        <div class="h-3 bg-[#f5f5f7] rounded w-full mb-2 animate-pulse"/>
        <div class="h-3 bg-[#f5f5f7] rounded w-2/3 mb-4 animate-pulse"/>
        <div class="flex justify-between items-center pt-3 border-t border-[#f5f5f7]">
          <div class="h-3 bg-[#f5f5f7] rounded w-20 animate-pulse"/>
          <div class="flex gap-3">
            <div class="h-3 bg-[#f5f5f7] rounded w-8 animate-pulse"/>
            <div class="h-3 bg-[#f5f5f7] rounded w-8 animate-pulse"/>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else-if="posts.length === 0" class="text-center py-24">
      <div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-[#f5f5f7] mb-5">
        <svg class="w-8 h-8 text-[#86868b]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10a2 2 0 012 2v1m2 13a2 2 0 01-2-2V7m2 13a2 2 0 002-2V9a2 2 0 00-2-2h-2m-4-3H9M7 16h6M7 8h6v4H7V8z"/>
        </svg>
      </div>
      <h3 class="text-xl font-semibold text-[#1d1d1f] mb-2">
        暂无博文
      </h3>
      <p class="text-sm text-[#86868b] mb-6">
        创建你的第一篇文章，开始记录和分享
      </p>
      <button
        @click="openCreateModal"
        class="inline-flex items-center gap-2 px-5 py-2.5 text-white text-[15px] font-medium rounded-xl transition-all duration-200 hover:scale-[1.02] active:scale-[0.98]"
        style="background: linear-gradient(135deg, #0071e3, #0063c7);"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
        </svg>
        新建博文
      </button>
    </div>

    <!-- Card Grid -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
      <div
        v-for="post in posts"
        :key="post.id"
        class="group bg-white rounded-2xl p-5 flex flex-col border border-[#f0f0f0] transition-all duration-300 hover:shadow-lg hover:border-[#e8e8ed] hover:-translate-y-0.5"
      >
        <h3
          class="text-[17px] font-semibold text-[#1d1d1f] leading-snug mb-2 line-clamp-2"
          :title="post.title"
        >
          {{ post.title }}
        </h3>
        <p class="text-sm text-[#86868b] leading-relaxed line-clamp-3 flex-1 mb-4">
          {{ post.content }}
        </p>
        <div class="flex justify-between items-center pt-3 border-t border-[#f5f5f7]">
          <span v-if="(post as any).createdAt" class="text-xs text-[#86868b]">
            {{ formatDate((post as any).createdAt) }}
          </span>
          <span v-else class="text-xs text-[#86868b]">博文</span>
          <div class="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity duration-200">
            <button
              @click="openEditModal(post)"
              class="p-1.5 rounded-lg text-[#0071e3] hover:bg-[#0071e3]/10 transition-colors"
              title="编辑"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
              </svg>
            </button>
            <button
              @click="handleDelete(post.id)"
              class="p-1.5 rounded-lg text-[#ff3b30] hover:bg-[#ff3b30]/10 transition-colors"
              title="删除"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
              </svg>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1 && !loading" class="flex justify-center items-center gap-1.5 mt-10">
      <button
        @click="changePage(currentPage - 1)"
        :disabled="currentPage <= 1"
        class="p-2 rounded-lg text-[#86868b] hover:text-[#1d1d1f] hover:bg-[#f5f5f7] disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:bg-transparent transition-all duration-150"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/>
        </svg>
      </button>
      <button
        v-for="page in totalPages"
        :key="page"
        @click="changePage(page)"
        class="min-w-[36px] h-9 px-2.5 rounded-lg text-sm font-medium transition-all duration-150"
        :class="page === currentPage
          ? 'bg-[#1d1d1f] text-white'
          : 'text-[#86868b] hover:text-[#1d1d1f] hover:bg-[#f5f5f7]'"
      >
        {{ page }}
      </button>
      <button
        @click="changePage(currentPage + 1)"
        :disabled="currentPage >= totalPages"
        class="p-2 rounded-lg text-[#86868b] hover:text-[#1d1d1f] hover:bg-[#f5f5f7] disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:bg-transparent transition-all duration-150"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
        </svg>
      </button>
    </div>
  </div>

  <!-- Modal -->
  <Transition
    enter-active-class="transition duration-200 ease-out"
    enter-from-class="opacity-0"
    enter-to-class="opacity-100"
    leave-active-class="transition duration-150 ease-in"
    leave-from-class="opacity-100"
    leave-to-class="opacity-0"
  >
    <div
      v-if="showModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4"
      style="background-color: rgba(0, 0, 0, 0.35);"
      @click.self="closeModal"
    >
      <Transition
        enter-active-class="transition duration-300 ease-out"
        enter-from-class="opacity-0 scale-95 translate-y-2"
        enter-to-class="opacity-100 scale-100 translate-y-0"
        leave-active-class="transition duration-200 ease-in"
        leave-from-class="opacity-100 scale-100 translate-y-0"
        leave-to-class="opacity-0 scale-95 translate-y-2"
      >
        <div
          v-if="showModal"
          class="bg-white w-full max-w-lg rounded-2xl shadow-2xl overflow-hidden"
        >
          <!-- Modal Header -->
          <div class="flex justify-between items-center px-6 py-4 border-b border-[#f0f0f0]">
            <h3 class="text-lg font-semibold text-[#1d1d1f]">
              {{ isEditing ? '编辑博文' : '新建博文' }}
            </h3>
            <button
              @click="closeModal"
              class="p-1 rounded-lg text-[#86868b] hover:text-[#1d1d1f] hover:bg-[#f5f5f7] transition-colors"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
              </svg>
            </button>
          </div>

          <!-- Modal Body -->
          <div class="px-6 py-5 space-y-4">
            <div>
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">
                标题
              </label>
              <input
                v-model="formData.title"
                type="text"
                placeholder="请输入标题"
                class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all duration-200 placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              />
            </div>
            <div>
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">
                内容
              </label>
              <textarea
                v-model="formData.content"
                rows="6"
                placeholder="请输入内容"
                class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none resize-none transition-all duration-200 placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              ></textarea>
            </div>
          </div>

          <!-- Modal Footer -->
          <div class="flex justify-end gap-2.5 px-6 py-4 border-t border-[#f0f0f0] bg-[#fafafc]/50">
            <button
              @click="closeModal"
              class="px-5 py-2 text-sm font-medium text-[#1d1d1f] bg-white border border-[#e8e8ed] rounded-xl hover:bg-[#f5f5f7] transition-colors"
            >
              取消
            </button>
            <button
              @click="handleSubmit"
              :disabled="!formData.title.trim() || isSubmitting"
              class="px-5 py-2 text-sm font-medium text-white rounded-xl transition-all duration-200 hover:scale-[1.02] active:scale-[0.98] disabled:opacity-40 disabled:scale-100 disabled:cursor-not-allowed"
              style="background: linear-gradient(135deg, #0071e3, #0063c7);"
            >
              {{ isSubmitting ? '保存中...' : '保存' }}
            </button>
          </div>
        </div>
      </Transition>
    </div>
  </Transition>
</template>

<style scoped>
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.line-clamp-3 {
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
