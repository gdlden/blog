<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useBlogStore } from '@/stores/blogStore'

const blogStore = useBlogStore()
const { posts, loading, totalPages, currentPage } = storeToRefs(blogStore)

const showModal = ref(false)
const isEditing = ref(false)
const formData = ref({ id: '', title: '', content: '' })

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
</script>

<template>
  <div class="max-w-6xl mx-auto px-4 py-6">
    <!-- Header -->
    <div class="flex justify-between items-center mb-6">
      <h2 class="text-2xl font-semibold text-gray-800">博文</h2>
      <button
        @click="openCreateModal"
        class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg font-medium transition-colors"
      >
        新建博文
      </button>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="text-center py-12">
      <p class="text-gray-600 text-lg">加载中...</p>
    </div>

    <!-- Empty State -->
    <div v-else-if="posts.length === 0" class="bg-gray-50 rounded-lg p-12 text-center border border-gray-200">
      <p class="text-xl font-semibold text-gray-700 mb-2">暂无内容</p>
      <p class="text-gray-500 mb-4">点击"新建博文"创建第一篇博文</p>
    </div>

    <!-- Card Grid -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div
        v-for="post in posts"
        :key="post.id"
        class="bg-white rounded-lg shadow border border-gray-200 p-4 flex flex-col"
      >
        <h3 class="text-lg font-semibold text-gray-800 truncate mb-2" :title="post.title">
          {{ post.title }}
        </h3>
        <p class="text-gray-600 text-sm line-clamp-3 flex-1 mb-4">
          {{ post.content }}
        </p>
        <div class="flex justify-end gap-2 pt-4 border-t border-gray-100">
          <button
            @click="openEditModal(post)"
            class="text-blue-600 hover:text-blue-800 text-sm font-medium px-3 py-1 rounded hover:bg-blue-50 transition-colors"
          >
            编辑
          </button>
          <button
            @click="handleDelete(post.id)"
            class="text-red-600 hover:text-red-800 text-sm font-medium px-3 py-1 rounded hover:bg-red-50 transition-colors"
          >
            删除
          </button>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1 && !loading" class="flex justify-center gap-2 mt-8">
      <button
        @click="changePage(currentPage - 1)"
        :disabled="currentPage <= 1"
        class="px-3 py-1 rounded border border-gray-300 text-sm font-medium transition-colors"
        :class="currentPage <= 1 ? 'bg-gray-100 text-gray-400 cursor-not-allowed' : 'bg-white text-gray-700 hover:bg-gray-50'"
      >
        上一页
      </button>
      <button
        v-for="page in totalPages"
        :key="page"
        @click="changePage(page)"
        class="px-3 py-1 rounded border text-sm font-medium transition-colors"
        :class="page === currentPage ? 'bg-blue-600 text-white border-blue-600' : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50'"
      >
        {{ page }}
      </button>
      <button
        @click="changePage(currentPage + 1)"
        :disabled="currentPage >= totalPages"
        class="px-3 py-1 rounded border border-gray-300 text-sm font-medium transition-colors"
        :class="currentPage >= totalPages ? 'bg-gray-100 text-gray-400 cursor-not-allowed' : 'bg-white text-gray-700 hover:bg-gray-50'"
      >
        下一页
      </button>
    </div>

    <!-- Modal -->
    <div
      v-if="showModal"
      class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4"
      @click.self="closeModal"
    >
      <div class="bg-white rounded-lg shadow-xl w-full max-w-lg">
        <!-- Modal Header -->
        <div class="flex justify-between items-center px-6 py-4 border-b border-gray-200">
          <h3 class="text-lg font-semibold text-gray-800">
            {{ isEditing ? '编辑博文' : '新建博文' }}
          </h3>
          <button
            @click="closeModal"
            class="text-gray-400 hover:text-gray-600 text-xl leading-none"
          >
            ×
          </button>
        </div>

        <!-- Modal Body -->
        <div class="px-6 py-4 space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">标题</label>
            <input
              v-model="formData.title"
              type="text"
              placeholder="请输入标题"
              class="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:border-transparent"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">内容</label>
            <textarea
              v-model="formData.content"
              rows="6"
              placeholder="请输入内容"
              class="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:border-transparent resize-none"
            ></textarea>
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="flex justify-end gap-3 px-6 py-4 border-t border-gray-200">
          <button
            @click="closeModal"
            class="px-4 py-2 text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg font-medium transition-colors"
          >
            取消
          </button>
          <button
            @click="handleSubmit"
            :disabled="!formData.title.trim() || loading"
            class="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-300 text-white rounded-lg font-medium transition-colors"
          >
            {{ loading ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.line-clamp-3 {
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
