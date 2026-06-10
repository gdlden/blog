<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { usePriceStore } from '@/stores/priceStore'
import type { PriceItem } from '@/api/price'

const priceStore = usePriceStore()
const { items, loading, totalPages, currentPage, itemCount, total } = storeToRefs(priceStore)

const showModal = ref(false)
const isEditing = ref(false)
const isSubmitting = ref(false)
const editingId = ref('')
const formData = ref({
  productName: '',
  weight: '',
  unitPrice: '',
  priceDate: '',
})

onMounted(() => priceStore.fetchItems())

function openCreateModal() {
  isEditing.value = false
  formData.value = {
    productName: '',
    weight: '',
    unitPrice: '',
    priceDate: new Date().toISOString().slice(0, 10),
  }
  showModal.value = true
}

function openEditModal(item: PriceItem) {
  isEditing.value = true
  editingId.value = item.id
  formData.value = {
    productName: item.productName,
    weight: item.weight,
    unitPrice: item.unitPrice,
    priceDate: item.priceDate,
  }
  showModal.value = true
}

async function handleSubmit() {
  if (!formData.value.productName.trim()) return
  isSubmitting.value = true
  try {
    if (isEditing.value) {
      await priceStore.updateItem({
        id: editingId.value,
        productName: formData.value.productName,
        weight: formData.value.weight || '0',
        unitPrice: formData.value.unitPrice || '0',
        totalPrice: '',
        priceDate: formData.value.priceDate,
      })
    } else {
      await priceStore.createItem({
        productName: formData.value.productName,
        weight: formData.value.weight || '0',
        unitPrice: formData.value.unitPrice || '0',
        priceDate: formData.value.priceDate,
      })
    }
    showModal.value = false
  } catch (err: any) {
    alert(err.message || '操作失败')
  } finally {
    isSubmitting.value = false
  }
}

async function handleDelete(id: string) {
  if (!confirm('确定要删除这条价格记录吗？')) return
  try {
    await priceStore.deleteItem(id)
  } catch (err: any) {
    alert(err.message || '删除失败')
  }
}

function changePage(page: number) {
  if (page >= 1 && page <= totalPages.value) priceStore.fetchItems(page)
}

function closeModal() {
  showModal.value = false
}

function getToday() {
  return new Date().toISOString().slice(0, 10)
}
</script>

<template>
  <div class="max-w-[1100px] mx-auto px-5 py-10">
    <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between mb-8">
      <div>
        <h1 class="text-[32px] font-semibold tracking-tight text-[#1d1d1f]">价格记录</h1>
        <p class="mt-1 text-sm text-[#86868b]">记录每次买菜的价格信息</p>
      </div>
      <button
        @click="openCreateModal"
        class="inline-flex items-center justify-center gap-2 px-5 py-2.5 text-white text-[15px] font-medium rounded-xl transition-all hover:scale-[1.02] active:scale-[0.98] shadow-sm"
        style="background: linear-gradient(135deg, #0071e3, #0063c7)"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 4v16m8-8H4"
          />
        </svg>
        新建记录
      </button>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
      <div class="bg-white rounded-2xl p-5 border border-[#f0f0f0]">
        <p class="text-sm text-[#86868b] mb-2">记录数</p>
        <p class="text-[24px] font-semibold text-[#1d1d1f]">{{ itemCount }}</p>
      </div>
      <div class="bg-white rounded-2xl p-5 border border-[#f0f0f0]">
        <p class="text-sm text-[#86868b] mb-2">当前页</p>
        <p class="text-[24px] font-semibold text-[#1d1d1f]">{{ currentPage }}</p>
      </div>
      <div class="bg-white rounded-2xl p-5 border border-[#f0f0f0]">
        <p class="text-sm text-[#86868b] mb-2">共 {{ total }} 条</p>
        <p class="text-[18px] font-semibold text-[#1d1d1f]">菜市场小票</p>
      </div>
    </div>

    <div v-if="loading" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
      <div v-for="n in 6" :key="n" class="bg-white rounded-2xl p-5 border border-[#f0f0f0]">
        <div class="h-5 bg-[#f5f5f7] rounded-lg w-3/4 mb-3 animate-pulse" />
        <div class="h-3 bg-[#f5f5f7] rounded w-full mb-2 animate-pulse" />
        <div class="h-3 bg-[#f5f5f7] rounded w-2/3 animate-pulse" />
      </div>
    </div>

    <div v-else-if="items.length === 0" class="text-center py-24">
      <div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-[#f5f5f7] mb-5">
        <svg class="w-8 h-8 text-[#86868b]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.5"
            d="M3 10h18M3 14h18m-9-4v8m-7 0h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"
          />
        </svg>
      </div>
      <h3 class="text-xl font-semibold text-[#1d1d1f] mb-2">暂无记录</h3>
      <p class="text-sm text-[#86868b] mb-6">点击上方按钮添加第一条价格记录</p>
      <button
        @click="openCreateModal"
        class="inline-flex items-center gap-2 px-5 py-2.5 text-white text-[15px] font-medium rounded-xl transition-all hover:scale-[1.02]"
        style="background: linear-gradient(135deg, #0071e3, #0063c7)"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 4v16m8-8H4"
          />
        </svg>
        新建记录
      </button>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
      <div
        v-for="item in items"
        :key="item.id"
        class="group bg-white rounded-2xl p-5 flex flex-col border border-[#f0f0f0] transition-all duration-300 hover:shadow-lg hover:border-[#e8e8ed] hover:-translate-y-0.5"
      >
        <div class="flex justify-between items-start mb-3">
          <h3
            class="text-[17px] font-semibold text-[#1d1d1f] leading-snug line-clamp-1 flex-1"
            :title="item.productName"
          >
            {{ item.productName }}
          </h3>
          <span
            class="text-xs px-2 py-0.5 rounded-full border font-medium ml-2 bg-green-50 text-green-700 border-green-200"
          >
            {{ item.priceDate }}
          </span>
        </div>

        <div class="grid grid-cols-3 gap-3 mb-3">
          <div class="bg-[#f5f5f7] rounded-xl p-3 text-center">
            <p class="text-xs text-[#86868b] mb-0.5">重量</p>
            <p class="text-[15px] font-semibold text-[#1d1d1f]">{{ item.weight }}</p>
          </div>
          <div class="bg-[#f5f5f7] rounded-xl p-3 text-center">
            <p class="text-xs text-[#86868b] mb-0.5">单价</p>
            <p class="text-[15px] font-semibold text-[#1d1d1f]">{{ item.unitPrice }}</p>
          </div>
          <div class="bg-[#f5f5f7] rounded-xl p-3 text-center">
            <p class="text-xs text-[#86868b] mb-0.5">总价</p>
            <p class="text-[15px] font-semibold text-[#0071e3]">{{ item.totalPrice }}</p>
          </div>
        </div>

        <div
          class="flex justify-end gap-1 opacity-0 group-hover:opacity-100 transition-opacity pt-3 border-t border-[#f5f5f7]"
        >
          <button
            @click="openEditModal(item)"
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
            @click="handleDelete(item.id)"
            class="p-1.5 rounded-lg text-[#ff3b30] hover:bg-[#ff3b30]/10 transition-colors"
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
      </div>
    </div>

    <div v-if="totalPages > 1 && !loading" class="flex justify-center items-center gap-1.5 mt-10">
      <button
        @click="changePage(currentPage - 1)"
        :disabled="currentPage <= 1"
        class="p-2 rounded-lg text-[#86868b] hover:text-[#1d1d1f] hover:bg-[#f5f5f7] disabled:opacity-30 disabled:cursor-not-allowed transition-all"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M15 19l-7-7 7-7"
          />
        </svg>
      </button>
      <span class="text-sm text-[#86868b] px-3">{{ currentPage }} / {{ totalPages }}</span>
      <button
        @click="changePage(currentPage + 1)"
        :disabled="currentPage >= totalPages"
        class="p-2 rounded-lg text-[#86868b] hover:text-[#1d1d1f] hover:bg-[#f5f5f7] disabled:opacity-30 disabled:cursor-not-allowed transition-all"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
      </button>
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
          class="bg-white w-full max-w-2xl rounded-2xl shadow-2xl overflow-hidden max-h-[90vh] overflow-y-auto"
        >
          <div class="flex justify-between items-center px-6 py-4 border-b border-[#f0f0f0]">
            <h3 class="text-lg font-semibold text-[#1d1d1f]">
              {{ isEditing ? '编辑记录' : '新建记录' }}
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
          <div class="px-6 py-5 grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="md:col-span-2">
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">商品名称</label>
              <input
                v-model="formData.productName"
                type="text"
                placeholder="例如：白菜、猪肉"
                class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              />
            </div>
            <div>
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">重量 (kg)</label>
              <input
                v-model="formData.weight"
                type="number"
                step="0.01"
                placeholder="0.00"
                class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              />
            </div>
            <div>
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">单价</label>
              <input
                v-model="formData.unitPrice"
                type="number"
                step="0.01"
                placeholder="0.00"
                class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              />
            </div>
            <div>
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">日期</label>
              <input
                v-model="formData.priceDate"
                type="date"
                class="w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              />
            </div>
            <div v-if="!isEditing && formData.weight && formData.unitPrice" class="md:col-span-2">
              <div class="bg-[#f0f7ff] rounded-xl p-4 text-center border border-[#d0e5ff]">
                <p class="text-sm text-[#86868b] mb-1">自动计算总价</p>
                <p class="text-[20px] font-semibold text-[#0071e3]">
                  {{ (parseFloat(formData.weight || '0') * parseFloat(formData.unitPrice || '0')).toFixed(2) }}
                </p>
              </div>
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
              :disabled="!formData.productName.trim() || isSubmitting"
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

<style scoped>
.line-clamp-1 {
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
