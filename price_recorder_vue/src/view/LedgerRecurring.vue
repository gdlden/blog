<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useLedgerStore } from '@/stores/ledgerStore'
import { DIRECTION_LABELS, currentMonth, formatAmount, ledgerErrorMessage } from '@/api/ledger'
import type { LedgerDirection, LedgerRecurring } from '@/api/ledger'
import LedgerNav from '@/components/LedgerNav.vue'

const ledgerStore = useLedgerStore()
const { recurringList, recurringLoading, activeAccounts, expenseCategories, incomeCategories } =
  storeToRefs(ledgerStore)

onMounted(() => {
  ledgerStore.fetchAccounts()
  ledgerStore.fetchCategories()
  ledgerStore.fetchRecurring()
})

const showModal = ref(false)
const isEditing = ref(false)
const isSubmitting = ref(false)
const formData = ref({
  id: '',
  type: 'expense' as LedgerDirection,
  accountId: '',
  categoryId: '',
  subCategoryId: '',
  amount: '' as string | number,
  dayOfMonth: '' as string | number,
  startMonth: '',
  remark: '',
  enabled: true,
})

const typeTabs: LedgerDirection[] = ['expense', 'income']

// 与记账弹窗同一写法：分类树按类型方向过滤，二级联动
const directionCategories = computed(() =>
  formData.value.type === 'income' ? incomeCategories.value : expenseCategories.value,
)
const topCategories = computed(() =>
  directionCategories.value.filter((c) => !c.parentId || c.parentId === '0'),
)

function childrenOf(parentId: string) {
  if (!parentId) return []
  return directionCategories.value.filter((c) => c.parentId === parentId)
}

function tabActiveClass(tab: LedgerDirection): string {
  if (formData.value.type !== tab) return 'text-[#86868b] bg-[#f5f5f7] hover:text-[#1d1d1f]'
  return tab === 'expense' ? 'text-white bg-[#ff3b30]' : 'text-white bg-[#34c759]'
}

const canSubmit = computed(() => {
  const day = Number(formData.value.dayOfMonth)
  return (
    !!formData.value.accountId &&
    !!(formData.value.subCategoryId || formData.value.categoryId) &&
    Number(formData.value.amount) > 0 &&
    Number.isInteger(day) &&
    day >= 1 &&
    day <= 31 &&
    !!formData.value.startMonth
  )
})

function openCreateModal() {
  isEditing.value = false
  formData.value = {
    id: '',
    type: 'expense',
    accountId: '',
    categoryId: '',
    subCategoryId: '',
    amount: '',
    dayOfMonth: '',
    startMonth: currentMonth(),
    remark: '',
    enabled: true,
  }
  showModal.value = true
}

// 编辑回显：二级分类拆回父/子两个 select
function openEditModal(item: LedgerRecurring) {
  isEditing.value = true
  const categoryId = String(item.categoryId)
  const parentId = ledgerStore.categoryMap.get(categoryId)?.parentId
  const hasParent = !!parentId && parentId !== '0'
  formData.value = {
    id: item.id,
    type: item.type,
    accountId: String(item.accountId),
    categoryId: hasParent ? parentId : categoryId,
    subCategoryId: hasParent ? categoryId : '',
    amount: item.amount,
    dayOfMonth: item.dayOfMonth,
    startMonth: item.startMonth,
    remark: item.remark || '',
    enabled: item.enabled,
  }
  showModal.value = true
}

function switchType(tab: LedgerDirection) {
  if (isEditing.value) return
  formData.value.type = tab
  formData.value.categoryId = ''
  formData.value.subCategoryId = ''
}

async function handleSubmit() {
  if (!canSubmit.value) return
  isSubmitting.value = true
  try {
    await ledgerStore.saveRecurring({
      id: formData.value.id || undefined,
      accountId: formData.value.accountId,
      categoryId: formData.value.subCategoryId || formData.value.categoryId,
      type: formData.value.type,
      amount: formData.value.amount,
      remark: formData.value.remark,
      dayOfMonth: formData.value.dayOfMonth,
      startMonth: formData.value.startMonth,
      enabled: formData.value.enabled,
    })
    showModal.value = false
  } catch (err) {
    alert(ledgerErrorMessage(err, '操作失败'))
  } finally {
    isSubmitting.value = false
  }
}

// 启停切换走 save 整对象更新，仅翻转 enabled
async function toggleEnabled(item: LedgerRecurring) {
  try {
    await ledgerStore.saveRecurring({
      id: item.id,
      accountId: item.accountId,
      categoryId: item.categoryId,
      type: item.type,
      amount: item.amount,
      remark: item.remark,
      dayOfMonth: item.dayOfMonth,
      startMonth: item.startMonth,
      enabled: !item.enabled,
    })
  } catch (err) {
    alert(ledgerErrorMessage(err, '操作失败'))
  }
}

async function handleDelete(item: LedgerRecurring) {
  if (!confirm(`确定要删除周期账单「${item.remark || item.categoryName}」吗？`)) return
  try {
    await ledgerStore.removeRecurring(item.id)
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
        <h1 class="text-[32px] font-semibold tracking-tight text-[#1d1d1f]">周期账单</h1>
        <p class="mt-1 text-sm text-[#86868b]">每月固定收支自动生成交易</p>
      </div>
      <div class="flex flex-wrap items-center gap-3">
        <LedgerNav />
        <button
          @click="openCreateModal"
          class="recurring-add inline-flex items-center justify-center gap-2 px-5 py-2.5 text-white text-[15px] font-medium rounded-xl transition-all hover:scale-[1.02] active:scale-[0.98] shadow-sm"
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
          新增周期账单
        </button>
      </div>
    </div>

    <div v-if="recurringLoading" class="space-y-3">
      <div v-for="n in 3" :key="n" class="bg-white rounded-2xl p-4 border border-[#f0f0f0]">
        <div class="h-4 bg-[#f5f5f7] rounded w-1/3 mb-2 animate-pulse" />
        <div class="h-3 bg-[#f5f5f7] rounded w-1/4 animate-pulse" />
      </div>
    </div>

    <div
      v-else-if="recurringList.length === 0"
      class="text-center py-24 bg-white rounded-2xl border border-[#f0f0f0]"
    >
      <h3 class="text-xl font-semibold text-[#1d1d1f] mb-2">暂无周期账单</h3>
      <p class="text-sm text-[#86868b] mb-6">添加房租、工资等每月固定收支，到期自动记账</p>
      <button
        @click="openCreateModal"
        class="inline-flex items-center gap-2 px-5 py-2.5 text-white text-[15px] font-medium rounded-xl transition-all hover:scale-[1.02]"
        style="background: linear-gradient(135deg, #0071e3, #0063c7)"
      >
        新增周期账单
      </button>
    </div>

    <div
      v-else
      class="bg-white rounded-2xl border border-[#f0f0f0] divide-y divide-[#f5f5f7] overflow-hidden"
    >
      <div
        v-for="item in recurringList"
        :key="item.id"
        class="recurring-item flex items-center gap-4 px-5 py-4 hover:bg-[#fafafc] transition-colors"
        :class="item.enabled ? '' : 'opacity-60'"
      >
        <div
          class="w-9 h-9 rounded-xl flex items-center justify-center shrink-0"
          :class="
            item.type === 'expense' ? 'bg-[#ff3b30]/10 text-[#ff3b30]' : 'bg-[#34c759]/10 text-[#34c759]'
          "
        >
          <svg
            v-if="item.type === 'expense'"
            class="w-4 h-4"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 5v14m0 0l-5-5m5 5l5-5"
            />
          </svg>
          <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 19V5m0 0l-5 5m5-5l5 5"
            />
          </svg>
        </div>
        <div class="flex-1 min-w-0">
          <p class="text-[15px] font-medium text-[#1d1d1f] truncate">
            {{ item.categoryName || '未分类' }}
          </p>
          <p class="text-xs text-[#86868b] truncate">
            {{ item.accountName || '未知账户' }} · 每月 {{ item.dayOfMonth }} 号 · 下次
            {{ item.nextDate || '-' }}
            <span v-if="item.remark"> · {{ item.remark }}</span>
          </p>
        </div>
        <p
          class="text-[16px] font-semibold shrink-0"
          :class="item.type === 'expense' ? 'text-[#ff3b30]' : 'text-[#34c759]'"
        >
          {{ item.type === 'expense' ? '-' : '+' }}¥{{ formatAmount(item.amount) }}
        </p>
        <button
          @click="toggleEnabled(item)"
          class="recurring-toggle relative w-9 h-5 rounded-full transition-colors shrink-0"
          :class="item.enabled ? 'bg-[#34c759]' : 'bg-[#d2d2d7]'"
          :title="item.enabled ? '点击停用' : '点击启用'"
        >
          <span
            class="absolute top-0.5 left-0.5 w-4 h-4 bg-white rounded-full shadow transform transition-transform"
            :class="item.enabled ? 'translate-x-4' : ''"
          />
        </button>
        <div class="inline-flex gap-1 shrink-0">
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
            @click="handleDelete(item)"
            class="recurring-delete p-1.5 rounded-lg text-[#ff3b30] hover:bg-[#ff3b30]/10 transition-colors"
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
          class="recurring-modal bg-white w-full max-w-2xl rounded-2xl shadow-2xl overflow-hidden max-h-[90vh] overflow-y-auto"
        >
          <div class="flex justify-between items-center px-6 py-4 border-b border-[#f0f0f0]">
            <h3 class="text-lg font-semibold text-[#1d1d1f]">
              {{ isEditing ? '编辑周期账单' : '新增周期账单' }}
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

          <div class="flex gap-2 px-6 pt-5">
            <button
              v-for="tab in typeTabs"
              :key="tab"
              @click="switchType(tab)"
              :disabled="isEditing"
              class="recurring-tab flex-1 py-2 text-sm font-medium rounded-xl transition-all disabled:cursor-not-allowed disabled:opacity-80"
              :class="tabActiveClass(tab)"
            >
              {{ DIRECTION_LABELS[tab] }}
            </button>
          </div>

          <div class="px-6 py-5 grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">
                {{ formData.type === 'expense' ? '付款账户' : '收款账户' }}
              </label>
              <select
                v-model="formData.accountId"
                class="recurring-account w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              >
                <option value="" disabled>请选择账户</option>
                <option v-for="account in activeAccounts" :key="account.id" :value="account.id">
                  {{ account.name }}
                </option>
              </select>
            </div>
            <div>
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">金额</label>
              <input
                v-model="formData.amount"
                type="number"
                step="0.01"
                min="0"
                placeholder="请输入金额"
                class="recurring-amount w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              />
            </div>
            <div>
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">分类</label>
              <select
                v-model="formData.categoryId"
                @change="formData.subCategoryId = ''"
                class="recurring-category w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              >
                <option value="" disabled>请选择分类</option>
                <option v-for="cat in topCategories" :key="cat.id" :value="cat.id">
                  {{ cat.name }}
                </option>
              </select>
            </div>
            <div v-if="childrenOf(formData.categoryId).length > 0">
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">子分类</label>
              <select
                v-model="formData.subCategoryId"
                class="recurring-sub-category w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              >
                <option value="">不选子分类</option>
                <option
                  v-for="cat in childrenOf(formData.categoryId)"
                  :key="cat.id"
                  :value="cat.id"
                >
                  {{ cat.name }}
                </option>
              </select>
            </div>
            <div>
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">每月几号</label>
              <input
                v-model="formData.dayOfMonth"
                type="number"
                min="1"
                max="31"
                placeholder="1-31"
                class="recurring-day w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              />
            </div>
            <div>
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">开始月份</label>
              <input
                v-model="formData.startMonth"
                type="month"
                class="recurring-month w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              />
            </div>
            <div class="md:col-span-2">
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">备注</label>
              <input
                v-model="formData.remark"
                type="text"
                placeholder="例如：房租、工资"
                class="recurring-remark w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              />
            </div>
            <label class="md:col-span-2 flex items-center gap-2 text-sm font-medium text-[#1d1d1f]">
              <input
                v-model="formData.enabled"
                type="checkbox"
                class="recurring-enabled w-4 h-4 rounded border-[#d2d2d7]"
              />
              启用（每月到期自动生成交易）
            </label>
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
              class="recurring-submit px-5 py-2 text-sm font-medium text-white rounded-xl transition-all hover:scale-[1.02] active:scale-[0.98] disabled:opacity-40 disabled:scale-100 disabled:cursor-not-allowed"
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
