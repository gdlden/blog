<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useLedgerStore } from '@/stores/ledgerStore'
import { DIRECTION_LABELS, ledgerErrorMessage } from '@/api/ledger'
import type { LedgerCategory, LedgerDirection } from '@/api/ledger'
import LedgerNav from '@/components/LedgerNav.vue'

const ledgerStore = useLedgerStore()
const { categories, categoriesLoading } = storeToRefs(ledgerStore)

onMounted(() => ledgerStore.fetchCategories())

const directions: LedgerDirection[] = ['expense', 'income']

function topOf(direction: LedgerDirection) {
  return categories.value.filter(
    (c) => c.direction === direction && (!c.parentId || c.parentId === '0'),
  )
}

function childrenOf(direction: LedgerDirection, parentId: string) {
  return categories.value.filter((c) => c.direction === direction && c.parentId === parentId)
}

const showModal = ref(false)
const isEditing = ref(false)
const isSubmitting = ref(false)
const formData = ref({
  id: '',
  name: '',
  direction: 'expense' as LedgerDirection,
  parentId: '0',
})

// 弹窗内父分类候选：当前方向下的一级分类
const parentOptions = computed(() => topOf(formData.value.direction))
const isChildCategory = computed(() => formData.value.parentId !== '0' && formData.value.parentId !== '')

const canSubmit = computed(() => !!formData.value.name.trim())

function openCreateModal(direction: LedgerDirection, parentId = '0') {
  isEditing.value = false
  formData.value = { id: '', name: '', direction, parentId }
  showModal.value = true
}

function openEditModal(category: LedgerCategory) {
  isEditing.value = true
  formData.value = {
    id: category.id,
    name: category.name,
    direction: category.direction,
    parentId: category.parentId || '0',
  }
  showModal.value = true
}

async function handleSubmit() {
  if (!canSubmit.value) return
  isSubmitting.value = true
  try {
    await ledgerStore.saveCategory(
      {
        id: formData.value.id || undefined,
        name: formData.value.name,
        direction: formData.value.direction,
        parentId: formData.value.parentId,
      },
      isEditing.value,
    )
    showModal.value = false
  } catch (err) {
    alert(ledgerErrorMessage(err, '操作失败'))
  } finally {
    isSubmitting.value = false
  }
}

async function handleDelete(category: LedgerCategory) {
  if (childrenOf(category.direction, category.id).length > 0) {
    alert('请先删除该分类下的子分类')
    return
  }
  if (!confirm(`确定要删除分类「${category.name}」吗？`)) return
  try {
    await ledgerStore.removeCategory(category.id)
  } catch (err) {
    alert(ledgerErrorMessage(err, '删除失败'))
  }
}

function closeModal() {
  showModal.value = false
}
</script>

<template>
  <div class="max-w-[1100px] mx-auto px-5 py-10">
    <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between mb-6">
      <div>
        <h1 class="text-[32px] font-semibold tracking-tight text-[#1d1d1f]">分类管理</h1>
        <p class="mt-1 text-sm text-[#86868b]">支出/收入两组，各支持两级分类</p>
      </div>
      <LedgerNav />
    </div>

    <div v-if="categoriesLoading" class="grid grid-cols-1 md:grid-cols-2 gap-5">
      <div v-for="n in 2" :key="n" class="bg-white rounded-2xl p-5 border border-[#f0f0f0]">
        <div class="h-5 bg-[#f5f5f7] rounded-lg w-1/3 mb-3 animate-pulse" />
        <div class="h-3 bg-[#f5f5f7] rounded w-full mb-2 animate-pulse" />
        <div class="h-3 bg-[#f5f5f7] rounded w-2/3 animate-pulse" />
      </div>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-5">
      <div
        v-for="direction in directions"
        :key="direction"
        class="category-group bg-white rounded-2xl border border-[#f0f0f0] p-5"
        :data-direction="direction"
      >
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-lg font-semibold text-[#1d1d1f]">
            {{ DIRECTION_LABELS[direction] }}分类
          </h2>
          <button
            @click="openCreateModal(direction)"
            class="category-add-top inline-flex items-center gap-1 px-3 py-1.5 text-sm font-medium text-[#0071e3] bg-[#0071e3]/5 rounded-xl hover:bg-[#0071e3]/10 transition-colors"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 4v16m8-8H4"
              />
            </svg>
            新增一级分类
          </button>
        </div>

        <p v-if="topOf(direction).length === 0" class="text-sm text-[#86868b] py-6 text-center">
          暂无分类
        </p>

        <div v-else class="space-y-1">
          <template v-for="parent in topOf(direction)" :key="parent.id">
            <div
              class="category-item flex items-center gap-2 px-3 py-2 rounded-xl hover:bg-[#fafafc] transition-colors"
            >
              <span class="flex-1 text-[15px] font-medium text-[#1d1d1f]">{{ parent.name }}</span>
              <span
                v-if="parent.isSystem"
                class="text-xs px-2 py-0.5 rounded-full border font-medium bg-gray-50 text-gray-500 border-gray-200"
                >内置</span
              >
              <button
                @click="openCreateModal(direction, parent.id)"
                class="category-add-child px-2 py-1 rounded-lg text-xs font-medium text-[#0071e3] hover:bg-[#0071e3]/10 transition-colors"
                title="添加子分类"
              >
                加子类
              </button>
              <button
                @click="openEditModal(parent)"
                class="p-1.5 rounded-lg text-[#0071e3] hover:bg-[#0071e3]/10 transition-colors"
                title="编辑"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                  />
                </svg>
              </button>
              <button
                v-if="!parent.isSystem"
                @click="handleDelete(parent)"
                class="category-delete p-1.5 rounded-lg text-[#ff3b30] hover:bg-[#ff3b30]/10 transition-colors"
                title="删除"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                  />
                </svg>
              </button>
            </div>
            <div
              v-for="child in childrenOf(direction, parent.id)"
              :key="child.id"
              class="category-item category-child flex items-center gap-2 pl-9 pr-3 py-1.5 rounded-xl hover:bg-[#fafafc] transition-colors"
            >
              <span class="text-[#c7c7cc] text-xs">└</span>
              <span class="flex-1 text-sm text-[#1d1d1f]">{{ child.name }}</span>
              <span
                v-if="child.isSystem"
                class="text-xs px-2 py-0.5 rounded-full border font-medium bg-gray-50 text-gray-500 border-gray-200"
                >内置</span
              >
              <button
                @click="openEditModal(child)"
                class="p-1.5 rounded-lg text-[#0071e3] hover:bg-[#0071e3]/10 transition-colors"
                title="编辑"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                  />
                </svg>
              </button>
              <button
                v-if="!child.isSystem"
                @click="handleDelete(child)"
                class="category-delete p-1.5 rounded-lg text-[#ff3b30] hover:bg-[#ff3b30]/10 transition-colors"
                title="删除"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                  />
                </svg>
              </button>
            </div>
          </template>
        </div>
      </div>
    </div>
  </div>

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
      style="background-color: rgba(0, 0, 0, 0.35)"
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
          class="ledger-category-modal bg-white w-full max-w-md rounded-2xl shadow-2xl overflow-hidden max-h-[90vh] overflow-y-auto"
        >
          <div class="flex justify-between items-center px-6 py-4 border-b border-[#f0f0f0]">
            <h3 class="text-lg font-semibold text-[#1d1d1f]">
              {{ isEditing ? '编辑分类' : '新增分类' }}
            </h3>
            <button
              @click="closeModal"
              class="p-1 rounded-lg text-[#86868b] hover:text-[#1d1d1f] hover:bg-[#f5f5f7] transition-colors"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>
          <div class="px-6 py-5 space-y-4">
            <div>
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">方向</label>
              <select
                v-model="formData.direction"
                :disabled="isEditing"
                class="category-direction w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10 disabled:opacity-60"
              >
                <option value="expense">支出</option>
                <option value="income">收入</option>
              </select>
            </div>
            <div>
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">层级</label>
              <select
                v-model="formData.parentId"
                :disabled="isEditing"
                class="category-parent w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10 disabled:opacity-60"
              >
                <option value="0">作为一级分类</option>
                <option v-for="parent in parentOptions" :key="parent.id" :value="parent.id">
                  属于：{{ parent.name }}
                </option>
              </select>
              <p v-if="isChildCategory && !isEditing" class="mt-1 text-xs text-[#86868b]">
                将作为所选一级分类的子分类
              </p>
            </div>
            <div>
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">分类名称</label>
              <input
                v-model="formData.name"
                type="text"
                placeholder="请输入分类名称"
                class="category-name w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              />
            </div>
          </div>
          <div class="flex justify-end gap-2.5 px-6 py-4 border-t border-[#f0f0f0] bg-[#fafafc]/50">
            <button
              @click="closeModal"
              class="px-5 py-2 text-sm font-medium text-[#1d1d1f] bg-white border border-[#e8e8ed] rounded-xl hover:bg-[#f5f5f7] transition-colors"
            >
              取消
            </button>
            <button
              @click="handleSubmit"
              :disabled="!canSubmit || isSubmitting"
              class="px-5 py-2 text-sm font-medium text-white rounded-xl transition-all hover:scale-[1.02] active:scale-[0.98] disabled:opacity-40 disabled:scale-100 disabled:cursor-not-allowed"
              style="background: linear-gradient(135deg, #0071e3, #0063c7)"
            >
              {{ isSubmitting ? '保存中...' : '保存' }}
            </button>
          </div>
        </div>
      </Transition>
    </div>
  </Transition>
</template>
