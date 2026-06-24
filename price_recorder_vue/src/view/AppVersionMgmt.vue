<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import {
  getAppVersions,
  createAppVersion,
  updateAppVersion,
  deleteAppVersion,
  uploadFile,
  type AppVersionEntity,
} from '@/api/appVersion'

const versions = ref<AppVersionEntity[]>([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))

const visiblePages = computed<Array<number | '...'>>(() => {
  const totalP = totalPages.value
  const cur = currentPage.value
  if (totalP <= 7) return Array.from({ length: totalP }, (_, i) => i + 1)
  if (cur <= 4) return [1, 2, 3, 4, 5, '...', totalP]
  if (cur >= totalP - 3) return [1, '...', totalP - 4, totalP - 3, totalP - 2, totalP - 1, totalP]
  return [1, '...', cur - 1, cur, cur + 1, '...', totalP]
})

const showModal = ref(false)
const isEditing = ref(false)
const isSubmitting = ref(false)
const uploadingIOS = ref(false)
const uploadingAndroid = ref(false)
const formData = ref({
  id: 0,
  version: '',
  info: [''],
  iosUrl: '',
  androidUrl: '',
  isActive: false,
})

onMounted(() => fetchVersions())

async function fetchVersions(page?: number) {
  loading.value = true
  try {
    const res = await getAppVersions(page ?? currentPage.value, pageSize.value)
    versions.value = res.list
    total.value = res.total
    if (page) currentPage.value = page
  } catch {
    // handled by interceptor
  } finally {
    loading.value = false
  }
}

function openCreateModal() {
  isEditing.value = false
  formData.value = { id: 0, version: '', info: [''], iosUrl: '', androidUrl: '', isActive: false }
  showModal.value = true
}

function openEditModal(v: AppVersionEntity) {
  isEditing.value = true
  formData.value = {
    id: v.id,
    version: v.version,
    info: v.info.length ? [...v.info] : [''],
    iosUrl: v.iosUrl || '',
    androidUrl: v.androidUrl || '',
    isActive: v.isActive,
  }
  showModal.value = true
}

function addInfoLine() {
  formData.value.info.push('')
}

function removeInfoLine(index: number) {
  if (formData.value.info.length > 1) {
    formData.value.info.splice(index, 1)
  }
}

async function handleSubmit() {
  if (!formData.value.version.trim()) return
  isSubmitting.value = true
  try {
    const payload = {
      version: formData.value.version.trim(),
      info: formData.value.info.filter((s) => s.trim()),
      iosUrl: formData.value.iosUrl.trim(),
      androidUrl: formData.value.androidUrl.trim(),
      isActive: formData.value.isActive,
    }
    if (isEditing.value) {
      await updateAppVersion({ id: formData.value.id, ...payload })
    } else {
      await createAppVersion(payload)
    }
    showModal.value = false
    await fetchVersions(currentPage.value)
  } catch {
    // handled by interceptor
  } finally {
    isSubmitting.value = false
  }
}

async function handleDelete(id: number) {
  if (!confirm('确定要删除这个版本记录吗？')) return
  try {
    await deleteAppVersion(id)
    await fetchVersions()
  } catch {
    // handled by interceptor
  }
}

function changePage(page: number) {
  if (page >= 1 && page <= totalPages.value) fetchVersions(page)
}

function closeModal() {
  showModal.value = false
}

async function handleUploadIOS(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  uploadingIOS.value = true
  try {
    const res = await uploadFile(file)
    formData.value.iosUrl = res.url
  } catch {
    // handled by interceptor
  } finally {
    uploadingIOS.value = false
    input.value = '' // allow re-upload same file
  }
}

async function handleUploadAndroid(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  uploadingAndroid.value = true
  try {
    const res = await uploadFile(file)
    formData.value.androidUrl = res.url
  } catch {
    // handled by interceptor
  } finally {
    uploadingAndroid.value = false
    input.value = ''
  }
}
</script>

<template>
  <div class="max-w-[1100px] mx-auto px-5 py-10">
    <!-- Header -->
    <div class="flex justify-between items-center mb-8">
      <div>
        <h1 class="text-[32px] font-semibold tracking-tight text-[#1d1d1f]">版本管理</h1>
        <p class="mt-1 text-sm text-[#86868b]">管理 App 版本信息与更新说明</p>
      </div>
      <button
        @click="openCreateModal"
        class="bg-[#0071e3] text-white text-[15px] font-medium px-4 py-2 rounded-lg hover:brightness-110 transition-all focus:outline-none focus:ring-2 focus:ring-[#0071e3] focus:ring-offset-2"
      >
        + 新增版本
      </button>
    </div>

    <!-- Loading state -->
    <div v-if="loading && !versions.length" class="text-center py-20 text-[#86868b] text-sm">
      加载中...
    </div>

    <!-- Empty state -->
    <div v-else-if="!versions.length" class="text-center py-20">
      <p class="text-[#86868b] text-sm">暂无版本记录</p>
    </div>

    <!-- Table -->
    <div v-else class="bg-white rounded-xl shadow-sm overflow-hidden">
      <table class="w-full text-left">
        <thead>
          <tr class="border-b border-gray-100 text-[13px] text-[#86868b] font-medium">
            <th class="px-5 py-3 w-16">ID</th>
            <th class="px-5 py-3">版本号</th>
            <th class="px-5 py-3">更新内容</th>
            <th class="px-5 py-3 hidden md:table-cell">iOS 链接</th>
            <th class="px-5 py-3 hidden md:table-cell">Android 链接</th>
            <th class="px-5 py-3 w-20">状态</th>
            <th class="px-5 py-3 w-24 text-right">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="v in versions" :key="v.id" class="border-b border-gray-50 hover:bg-[#f5f5f7] transition-colors text-[15px]">
            <td class="px-5 py-4 text-[#86868b] text-sm">{{ v.id }}</td>
            <td class="px-5 py-4 font-medium text-[#1d1d1f]">{{ v.version }}</td>
            <td class="px-5 py-4 max-w-[200px]">
              <div class="flex flex-wrap gap-1">
                <span
                  v-for="(item, i) in v.info.slice(0, 2)"
                  :key="i"
                  class="inline-block bg-[#f5f5f7] text-[#1d1d1f] text-[12px] px-2 py-0.5 rounded-md truncate max-w-[120px]"
                >
                  {{ item }}
                </span>
                <span v-if="v.info.length > 2" class="text-[#86868b] text-[12px]">
                  +{{ v.info.length - 2 }}
                </span>
              </div>
            </td>
            <td class="px-5 py-4 hidden md:table-cell text-sm text-[#86868b] max-w-[150px] truncate">
              {{ v.iosUrl || '-' }}
            </td>
            <td class="px-5 py-4 hidden md:table-cell text-sm text-[#86868b] max-w-[150px] truncate">
              {{ v.androidUrl || '-' }}
            </td>
            <td class="px-5 py-4">
              <span
                class="inline-block text-[12px] font-medium px-2 py-0.5 rounded-full"
                :class="v.isActive ? 'bg-green-50 text-green-700 border border-green-200' : 'bg-gray-50 text-[#86868b] border border-gray-200'"
              >
                {{ v.isActive ? '激活' : '未激活' }}
              </span>
            </td>
            <td class="px-5 py-4 text-right">
              <button
                @click="openEditModal(v)"
                class="text-[#0071e3] text-[14px] font-medium hover:underline mr-3"
              >
                编辑
              </button>
              <button
                @click="handleDelete(v.id)"
                class="text-[#ff3b30] text-[14px] font-medium hover:underline"
              >
                删除
              </button>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Pagination -->
      <div v-if="totalPages > 1" class="flex items-center justify-center gap-1 px-5 py-4 border-t border-gray-100">
        <button
          :disabled="currentPage <= 1"
          @click="changePage(currentPage - 1)"
          class="px-3 py-1.5 text-[13px] font-medium rounded-lg transition-colors disabled:opacity-30 disabled:cursor-not-allowed hover:bg-gray-100"
        >
          上一页
        </button>
        <button
          v-for="(p, i) in visiblePages"
          :key="i"
          @click="typeof p === 'number' && changePage(p)"
          class="min-w-[32px] px-2 py-1.5 text-[13px] font-medium rounded-lg transition-colors"
          :class="p === currentPage ? 'bg-[#0071e3] text-white' : 'hover:bg-gray-100 text-[#1d1d1f]'"
          :disabled="typeof p !== 'number'"
        >
          {{ p }}
        </button>
        <button
          :disabled="currentPage >= totalPages"
          @click="changePage(currentPage + 1)"
          class="px-3 py-1.5 text-[13px] font-medium rounded-lg transition-colors disabled:opacity-30 disabled:cursor-not-allowed hover:bg-gray-100"
        >
          下一页
        </button>
      </div>
    </div>

    <!-- Modal Overlay -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition duration-200 ease-out"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition duration-150 ease-in"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
      >
        <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm" @click.self="closeModal">
          <div class="bg-white rounded-2xl shadow-lg w-full max-w-lg mx-4 p-6">
            <h2 class="text-[21px] font-semibold text-[#1d1d1f] mb-1">
              {{ isEditing ? '编辑版本' : '新增版本' }}
            </h2>
            <p class="text-sm text-[#86868b] mb-6">请填写版本信息，带 * 为必填</p>

            <form @submit.prevent="handleSubmit" class="space-y-4">
              <!-- Version -->
              <div>
                <label class="block text-[13px] font-medium text-[#1d1d1f] mb-1">版本号 *</label>
                <input
                  v-model="formData.version"
                  placeholder="例如 1.0.0"
                  class="w-full px-3 py-2 border border-gray-200 rounded-lg text-[15px] text-[#1d1d1f] placeholder:text-[#86868b] focus:outline-none focus:ring-2 focus:ring-[#0071e3] focus:border-transparent"
                  required
                />
              </div>

              <!-- Info (update notes) -->
              <div>
                <label class="block text-[13px] font-medium text-[#1d1d1f] mb-1">更新内容</label>
                <div class="space-y-2">
                  <div v-for="(_, i) in formData.info" :key="i" class="flex items-center gap-2">
                    <input
                      v-model="formData.info[i]"
                      placeholder="输入一条更新说明"
                      class="flex-1 px-3 py-2 border border-gray-200 rounded-lg text-[15px] text-[#1d1d1f] placeholder:text-[#86868b] focus:outline-none focus:ring-2 focus:ring-[#0071e3] focus:border-transparent"
                    />
                    <button
                      v-if="formData.info.length > 1"
                      type="button"
                      @click="removeInfoLine(i)"
                      class="text-[#86868b] hover:text-[#ff3b30] p-1"
                    >
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
                    </button>
                  </div>
                </div>
                <button
                  type="button"
                  @click="addInfoLine"
                  class="mt-2 text-[#0071e3] text-[13px] font-medium hover:underline"
                >
                  + 添加更新说明
                </button>
              </div>

              <!-- iOS URL -->
              <div>
                <label class="block text-[13px] font-medium text-[#1d1d1f] mb-1">iOS 下载地址</label>
                <div class="flex gap-2">
                  <input
                    v-model="formData.iosUrl"
                    placeholder="itms-apps://..."
                    class="flex-1 px-3 py-2 border border-gray-200 rounded-lg text-[15px] text-[#1d1d1f] placeholder:text-[#86868b] focus:outline-none focus:ring-2 focus:ring-[#0071e3] focus:border-transparent"
                  />
                  <label
                    class="inline-flex items-center gap-1.5 px-3 py-2 bg-[#f5f5f7] text-[#1d1d1f] text-[13px] font-medium rounded-lg cursor-pointer hover:bg-gray-200 transition-colors whitespace-nowrap"
                    :class="{ 'opacity-50 pointer-events-none': uploadingIOS }"
                  >
                    <svg v-if="!uploadingIOS" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"/></svg>
                    <svg v-else class="w-4 h-4 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/></svg>
                    {{ uploadingIOS ? '上传中...' : '上传 IPA' }}
                    <input type="file" accept=".ipa,.dmg,.app" hidden @change="handleUploadIOS" />
                  </label>
                </div>
              </div>

              <!-- Android URL -->
              <div>
                <label class="block text-[13px] font-medium text-[#1d1d1f] mb-1">Android 下载地址</label>
                <div class="flex gap-2">
                  <input
                    v-model="formData.androidUrl"
                    placeholder="https://..."
                    class="flex-1 px-3 py-2 border border-gray-200 rounded-lg text-[15px] text-[#1d1d1f] placeholder:text-[#86868b] focus:outline-none focus:ring-2 focus:ring-[#0071e3] focus:border-transparent"
                  />
                  <label
                    class="inline-flex items-center gap-1.5 px-3 py-2 bg-[#f5f5f7] text-[#1d1d1f] text-[13px] font-medium rounded-lg cursor-pointer hover:bg-gray-200 transition-colors whitespace-nowrap"
                    :class="{ 'opacity-50 pointer-events-none': uploadingAndroid }"
                  >
                    <svg v-if="!uploadingAndroid" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"/></svg>
                    <svg v-else class="w-4 h-4 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/></svg>
                    {{ uploadingAndroid ? '上传中...' : '上传 APK' }}
                    <input type="file" accept=".apk,.aab" hidden @change="handleUploadAndroid" />
                  </label>
                </div>
              </div>

              <!-- Active toggle -->
              <div class="flex items-center gap-2">
                <input
                  id="isActive"
                  type="checkbox"
                  v-model="formData.isActive"
                  class="w-4 h-4 rounded border-gray-300 text-[#0071e3] focus:ring-[#0071e3]"
                />
                <label for="isActive" class="text-[15px] text-[#1d1d1f]">设为当前激活版本（对外发布）</label>
              </div>

              <!-- Actions -->
              <div class="flex justify-end gap-3 pt-2">
                <button
                  type="button"
                  @click="closeModal"
                  class="px-4 py-2 text-[15px] font-medium text-[#1d1d1f] bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
                >
                  取消
                </button>
                <button
                  type="submit"
                  :disabled="isSubmitting"
                  class="px-4 py-2 text-[15px] font-medium text-white bg-[#0071e3] rounded-lg hover:brightness-110 transition-all disabled:opacity-50"
                >
                  {{ isSubmitting ? '提交中...' : '保存' }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
