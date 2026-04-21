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
  <div class="max-w-[980px] mx-auto px-5 py-12">
    <!-- Header -->
    <div class="flex justify-between items-center mb-10">
      <h2 style="font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', 'SF Pro Icons', 'Helvetica Neue', Helvetica, Arial, sans-serif; font-size: 40px; font-weight: 600; line-height: 1.10; letter-spacing: normal; color: #1d1d1f;">
        博文
      </h2>
      <button
        @click="openCreateModal"
        class="text-white transition-colors"
        style="background-color: #0071e3; border-radius: 8px; padding: 8px 15px; font-size: 17px; font-weight: 400; line-height: 2.41;"
      >
        新建博文
      </button>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="text-center py-20">
      <p style="font-size: 17px; letter-spacing: -0.374px; line-height: 1.47; color: rgba(0, 0, 0, 0.48);">
        加载中...
      </p>
    </div>

    <!-- Empty State -->
    <div v-else-if="posts.length === 0" class="text-center py-20">
      <p style="font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', 'SF Pro Icons', 'Helvetica Neue', Helvetica, Arial, sans-serif; font-size: 28px; font-weight: 400; line-height: 1.14; letter-spacing: 0.196px; color: #1d1d1f; margin-bottom: 8px;">
        暂无内容
      </p>
      <p style="font-size: 14px; letter-spacing: -0.224px; line-height: 1.29; color: rgba(0, 0, 0, 0.48);">
        点击"新建博文"创建第一篇博文
      </p>
    </div>

    <!-- Card Grid -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div
        v-for="post in posts"
        :key="post.id"
        class="bg-white p-5 flex flex-col"
        style="border-radius: 8px; box-shadow: rgba(0, 0, 0, 0.22) 3px 5px 30px 0px;"
      >
        <h3
          class="truncate mb-2"
          style="font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', 'SF Pro Icons', 'Helvetica Neue', Helvetica, Arial, sans-serif; font-size: 21px; font-weight: 700; line-height: 1.19; letter-spacing: 0.231px; color: #1d1d1f;"
          :title="post.title"
        >
          {{ post.title }}
        </h3>
        <p
          class="line-clamp-3 flex-1 mb-5"
          style="font-size: 14px; font-weight: 400; line-height: 1.29; letter-spacing: -0.224px; color: rgba(0, 0, 0, 0.8);"
        >
          {{ post.content }}
        </p>
        <div class="flex justify-end gap-4 pt-4" style="border-top: 1px solid rgba(0, 0, 0, 0.04);">
          <button
            @click="openEditModal(post)"
            class="transition-colors"
            style="font-size: 14px; font-weight: 400; line-height: 1.43; letter-spacing: -0.224px; color: #0066cc;"
          >
            编辑
          </button>
          <button
            @click="handleDelete(post.id)"
            class="transition-colors"
            style="font-size: 14px; font-weight: 400; line-height: 1.43; letter-spacing: -0.224px; color: #0066cc;"
          >
            删除
          </button>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1 && !loading" class="flex justify-center gap-2 mt-10">
      <button
        @click="changePage(currentPage - 1)"
        :disabled="currentPage <= 1"
        class="px-3 py-1.5 text-sm font-medium transition-colors"
        :class="currentPage <= 1 ? 'cursor-not-allowed' : 'hover:bg-white'"
        :style="currentPage <= 1
          ? 'background-color: transparent; color: rgba(0, 0, 0, 0.48); border-radius: 8px;'
          : 'background-color: #ffffff; color: #1d1d1f; border-radius: 8px;'"
      >
        上一页
      </button>
      <button
        v-for="page in totalPages"
        :key="page"
        @click="changePage(page)"
        class="px-3 py-1.5 text-sm font-medium transition-colors"
        :style="page === currentPage
          ? 'background-color: #0071e3; color: #ffffff; border-radius: 8px;'
          : 'background-color: #ffffff; color: #1d1d1f; border-radius: 8px;'"
      >
        {{ page }}
      </button>
      <button
        @click="changePage(currentPage + 1)"
        :disabled="currentPage >= totalPages"
        class="px-3 py-1.5 text-sm font-medium transition-colors"
        :class="currentPage >= totalPages ? 'cursor-not-allowed' : 'hover:bg-white'"
        :style="currentPage >= totalPages
          ? 'background-color: transparent; color: rgba(0, 0, 0, 0.48); border-radius: 8px;'
          : 'background-color: #ffffff; color: #1d1d1f; border-radius: 8px;'"
      >
        下一页
      </button>
    </div>

    <!-- Modal -->
    <div
      v-if="showModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4"
      style="background-color: rgba(0, 0, 0, 0.4);"
      @click.self="closeModal"
    >
      <div class="bg-white w-full max-w-lg" style="border-radius: 12px;">
        <!-- Modal Header -->
        <div class="flex justify-between items-center px-6 py-4" style="border-bottom: 1px solid rgba(0, 0, 0, 0.04);">
          <h3 style="font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', 'SF Pro Icons', 'Helvetica Neue', Helvetica, Arial, sans-serif; font-size: 21px; font-weight: 700; line-height: 1.19; letter-spacing: 0.231px; color: #1d1d1f;">
            {{ isEditing ? '编辑博文' : '新建博文' }}
          </h3>
          <button
            @click="closeModal"
            class="text-2xl leading-none transition-colors"
            style="color: rgba(0, 0, 0, 0.48);"
          >
            &times;
          </button>
        </div>

        <!-- Modal Body -->
        <div class="px-6 py-5 space-y-4">
          <div>
            <label class="block mb-1.5" style="font-size: 14px; font-weight: 400; letter-spacing: -0.224px; line-height: 1.43; color: rgba(0, 0, 0, 0.8);">
              标题
            </label>
            <input
              v-model="formData.title"
              type="text"
              placeholder="请输入标题"
              class="w-full px-3 py-2.5 outline-none"
              style="background-color: #fafafc; border-radius: 8px; border: 1px solid rgba(0, 0, 0, 0.04); font-size: 17px; letter-spacing: -0.374px; color: #1d1d1f;"
            />
          </div>
          <div>
            <label class="block mb-1.5" style="font-size: 14px; font-weight: 400; letter-spacing: -0.224px; line-height: 1.43; color: rgba(0, 0, 0, 0.8);">
              内容
            </label>
            <textarea
              v-model="formData.content"
              rows="6"
              placeholder="请输入内容"
              class="w-full px-3 py-2.5 outline-none resize-none"
              style="background-color: #fafafc; border-radius: 8px; border: 1px solid rgba(0, 0, 0, 0.04); font-size: 17px; letter-spacing: -0.374px; color: #1d1d1f;"
            ></textarea>
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="flex justify-end gap-3 px-6 py-4" style="border-top: 1px solid rgba(0, 0, 0, 0.04);">
          <button
            @click="closeModal"
            class="px-4 py-2 transition-colors"
            style="background-color: #1d1d1f; color: #ffffff; border-radius: 8px; font-size: 17px; font-weight: 400; line-height: 2.41;"
          >
            取消
          </button>
          <button
            @click="handleSubmit"
            :disabled="!formData.title.trim() || loading"
            class="px-4 py-2 text-white transition-colors disabled:opacity-50"
            style="background-color: #0071e3; border-radius: 8px; font-size: 17px; font-weight: 400; line-height: 2.41;"
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
