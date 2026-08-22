<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useLedgerStore } from '@/stores/ledgerStore'
import * as ledgerApi from '@/api/ledger'
import { TRANSACTION_TYPE_LABELS, ledgerErrorMessage } from '@/api/ledger'
import type {
  LedgerPosting,
  LedgerTransaction,
  LedgerTransactionType,
  PostingDraft,
  TransactionLeg,
} from '@/api/ledger'
import LedgerNav from '@/components/LedgerNav.vue'

const ledgerStore = useLedgerStore()
const {
  transactions,
  txLoading,
  txPage,
  txTotalPages,
  filterMonth,
  filterAccountId,
  filterCategoryId,
  filterType,
  activeAccounts,
  expenseCategories,
  incomeCategories,
} = storeToRefs(ledgerStore)

const formatAmount = ledgerApi.formatAmount

onMounted(() => {
  ledgerStore.fetchAccounts()
  ledgerStore.fetchCategories()
  ledgerStore.fetchTransactions(1)
  autoApplyRecurring()
})

// 静默触发周期账单生成：不阻塞首屏、不打扰用户；有新生成则刷新列表与统计
async function autoApplyRecurring() {
  const created = await ledgerStore.applyRecurring()
  if (created > 0) {
    ledgerStore.fetchTransactions()
    ledgerStore.fetchMonthlyStats(filterMonth.value)
  }
}

// 月份/筛选变化都回到第一页重新查询
watch([filterMonth, filterAccountId, filterCategoryId, filterType], () => {
  ledgerStore.fetchTransactions(1)
})

function prevMonth() {
  filterMonth.value = ledgerApi.shiftMonth(filterMonth.value, -1)
}

function nextMonth() {
  filterMonth.value = ledgerApi.shiftMonth(filterMonth.value, 1)
}

function changeTxPage(page: number) {
  if (page >= 1 && page <= txTotalPages.value) ledgerStore.fetchTransactions(page)
}

/* ---------- 列表渲染辅助 ---------- */

// 用户侧腿按金额符号识别：支出为负腿，收入为正腿，与系统账户 id 无关
function txAmount(tx: LedgerTransaction): number {
  if (tx.type === 'expense') {
    return tx.postings
      .filter((p) => Number(p.amount) < 0)
      .reduce((sum, p) => sum + Math.abs(Number(p.amount)), 0)
  }
  if (tx.type === 'income') {
    return tx.postings
      .filter((p) => Number(p.amount) > 0)
      .reduce((sum, p) => sum + Number(p.amount), 0)
  }
  return Math.abs(Number(tx.postings[0]?.amount || 0))
}

function txAmountText(tx: LedgerTransaction): string {
  const sign = tx.type === 'expense' ? '-' : tx.type === 'income' ? '+' : ''
  return `${sign}¥${formatAmount(txAmount(tx))}`
}

function amountColor(type: LedgerTransactionType): string {
  if (type === 'expense') return 'text-[#ff3b30]'
  if (type === 'income') return 'text-[#34c759]'
  return 'text-[#86868b]'
}

function typeIconClass(type: LedgerTransactionType): string {
  if (type === 'expense') return 'bg-[#ff3b30]/10 text-[#ff3b30]'
  if (type === 'income') return 'bg-[#34c759]/10 text-[#34c759]'
  return 'bg-[#86868b]/10 text-[#86868b]'
}

function txCategoryText(tx: LedgerTransaction): string {
  const names = tx.postings
    .filter((p) => p.categoryId)
    .map((p) => ledgerStore.categoryMap.get(String(p.categoryId))?.name || '未分类')
  const unique = [...new Set(names)]
  if (unique.length === 0) return tx.type === 'transfer' ? '转账' : '未分类'
  return unique.join(' + ')
}

function accountName(accountId?: string): string {
  if (!accountId) return '未知账户'
  return ledgerStore.accountMap.get(String(accountId))?.name || '未知账户'
}

function txAccountText(tx: LedgerTransaction): string {
  if (tx.type === 'transfer') {
    const from = tx.postings.find((p) => Number(p.amount) < 0)
    const to = tx.postings.find((p) => Number(p.amount) > 0)
    return `${accountName(from?.accountId)} → ${accountName(to?.accountId)}`
  }
  const userLeg = tx.postings.find((p) =>
    tx.type === 'expense' ? Number(p.amount) < 0 : Number(p.amount) > 0,
  )
  return accountName(userLeg?.accountId)
}

/* ---------- 记一笔弹窗 ---------- */

const showTxModal = ref(false)
const isEditingTx = ref(false)
const isSubmittingTx = ref(false)
const editingTxId = ref('')
const txForm = ref({
  type: 'expense' as LedgerTransactionType,
  accountId: '',
  toAccountId: '',
  total: '' as string | number,
  bookedAt: '',
  remark: '',
  categoryId: '',
  subCategoryId: '',
})
const splitEnabled = ref(false)
const txLegs = ref<Array<{ amount: string | number; categoryId: string; subCategoryId: string }>>(
  [],
)

const txTypeTabs: LedgerTransactionType[] = ['expense', 'income', 'transfer']

function nowLocalDateTime(): string {
  const date = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

// 当前类型方向的分类树（转账无分类）
const directionCategories = computed(() =>
  txForm.value.type === 'income' ? incomeCategories.value : expenseCategories.value,
)
const topCategories = computed(() =>
  directionCategories.value.filter((c) => !c.parentId || c.parentId === '0'),
)

function childrenOf(parentId: string) {
  if (!parentId) return []
  return directionCategories.value.filter((c) => c.parentId === parentId)
}

function tabActiveClass(tab: LedgerTransactionType): string {
  if (txForm.value.type !== tab) return 'text-[#86868b] bg-[#f5f5f7] hover:text-[#1d1d1f]'
  if (tab === 'expense') return 'text-white bg-[#ff3b30]'
  if (tab === 'income') return 'text-white bg-[#34c759]'
  return 'text-white bg-[#0071e3]'
}

// 不拆分时由外层分类组成单腿；拆分时取各腿，二级分类优先于一级
const effectiveLegs = computed<TransactionLeg[]>(() => {
  if (!splitEnabled.value) {
    return [
      {
        amount: txForm.value.total,
        categoryId: txForm.value.subCategoryId || txForm.value.categoryId,
      },
    ]
  }
  return txLegs.value.map((leg) => ({
    amount: leg.amount,
    categoryId: leg.subCategoryId || leg.categoryId,
  }))
})

// 金额按"分"比较，规避浮点误差
const totalCents = computed(() => Math.round(Number(txForm.value.total || 0) * 100))
const legSumCents = computed(() =>
  Math.round(txLegs.value.reduce((sum, leg) => sum + Number(leg.amount || 0), 0) * 100),
)
const legSumText = computed(() => formatAmount(legSumCents.value / 100))
const splitDiffText = computed(() => formatAmount(Math.abs(totalCents.value - legSumCents.value) / 100))
const splitBalanced = computed(() => legSumCents.value === totalCents.value)

const canSubmitTx = computed(() => {
  const total = Number(txForm.value.total)
  if (!txForm.value.accountId || !txForm.value.bookedAt || !total || total <= 0) return false
  if (txForm.value.type === 'transfer') {
    return !!txForm.value.toAccountId && txForm.value.toAccountId !== txForm.value.accountId
  }
  const legs = effectiveLegs.value
  if (legs.some((leg) => !leg.categoryId)) return false
  if (splitEnabled.value) {
    return legs.length > 0 && legs.every((leg) => Number(leg.amount) > 0) && splitBalanced.value
  }
  return true
})

function resetTxForm(type: LedgerTransactionType = 'expense') {
  txForm.value = {
    type,
    accountId: '',
    toAccountId: '',
    total: '',
    bookedAt: nowLocalDateTime(),
    remark: '',
    categoryId: '',
    subCategoryId: '',
  }
  splitEnabled.value = false
  txLegs.value = []
}

function openCreateTxModal() {
  isEditingTx.value = false
  editingTxId.value = ''
  resetTxForm()
  showTxModal.value = true
}

function switchTxType(type: LedgerTransactionType) {
  if (isEditingTx.value) return
  txForm.value.type = type
  txForm.value.categoryId = ''
  txForm.value.subCategoryId = ''
  txLegs.value = []
}

function toggleSplit() {
  splitEnabled.value = !splitEnabled.value
  if (splitEnabled.value && txLegs.value.length === 0) {
    txLegs.value = [
      { amount: '', categoryId: '', subCategoryId: '' },
      { amount: '', categoryId: '', subCategoryId: '' },
    ]
  }
}

function addLeg() {
  txLegs.value.push({ amount: '', categoryId: '', subCategoryId: '' })
}

function removeLeg(index: number) {
  txLegs.value.splice(index, 1)
}

// 分类腿回显：单腿不拆分；多腿进入拆分模式。二级分类拆回父/子两个 select
function fillLegs(categoryLegs: LedgerPosting[]) {
  const legs = categoryLegs.filter((p) => p.categoryId)
  if (legs.length <= 1) {
    splitEnabled.value = false
    txLegs.value = []
    const categoryId = String(legs[0]?.categoryId || '')
    const parentId = ledgerStore.categoryMap.get(categoryId)?.parentId
    if (parentId && parentId !== '0') {
      txForm.value.categoryId = parentId
      txForm.value.subCategoryId = categoryId
    } else {
      txForm.value.categoryId = categoryId
      txForm.value.subCategoryId = ''
    }
    return
  }
  splitEnabled.value = true
  txForm.value.categoryId = ''
  txForm.value.subCategoryId = ''
  txLegs.value = legs.map((p) => {
    const categoryId = String(p.categoryId)
    const parentId = ledgerStore.categoryMap.get(categoryId)?.parentId
    const amount = String(Math.abs(Number(p.amount)))
    if (parentId && parentId !== '0') {
      return { amount, categoryId: parentId, subCategoryId: categoryId }
    }
    return { amount, categoryId, subCategoryId: '' }
  })
}

async function openEditTxModal(tx: LedgerTransaction) {
  try {
    // 编辑统一走 get/v1 取整组详情，避免依赖列表项字段
    const detail = await ledgerApi.getTransactionById(tx.id)
    isEditingTx.value = true
    editingTxId.value = detail.id
    txForm.value.type = detail.type
    txForm.value.bookedAt = detail.bookedAt
      ? detail.bookedAt.replace(' ', 'T').slice(0, 16)
      : nowLocalDateTime()
    txForm.value.remark = detail.remark || ''
    const negatives = detail.postings.filter((p) => Number(p.amount) < 0)
    const positives = detail.postings.filter((p) => Number(p.amount) > 0)
    if (detail.type === 'transfer') {
      txForm.value.accountId = String(negatives[0]?.accountId || '')
      txForm.value.toAccountId = String(positives[0]?.accountId || '')
      txForm.value.total = String(Math.abs(Number(negatives[0]?.amount || 0)))
      splitEnabled.value = false
      txLegs.value = []
    } else if (detail.type === 'expense') {
      txForm.value.accountId = String(negatives[0]?.accountId || '')
      txForm.value.total = String(Math.abs(Number(negatives[0]?.amount || 0)))
      fillLegs(positives)
    } else {
      txForm.value.accountId = String(positives[0]?.accountId || '')
      txForm.value.total = String(Math.abs(Number(positives[0]?.amount || 0)))
      fillLegs(negatives)
    }
    showTxModal.value = true
  } catch (err) {
    alert(ledgerErrorMessage(err, '获取交易详情失败'))
  }
}

async function handleTxSubmit() {
  if (!canSubmitTx.value) return
  isSubmittingTx.value = true
  // datetime-local 值 'YYYY-MM-DDTHH:mm'，契约时间格式 'YYYY-MM-DD HH:mm:ss'
  const bookedAt =
    txForm.value.bookedAt.length === 16
      ? `${txForm.value.bookedAt.replace('T', ' ')}:00`
      : txForm.value.bookedAt.replace('T', ' ')
  let postings: PostingDraft[]
  if (txForm.value.type === 'transfer') {
    postings = ledgerApi.buildTransferPostings(
      txForm.value.accountId,
      txForm.value.toAccountId,
      txForm.value.total,
    )
  } else if (txForm.value.type === 'income') {
    postings = ledgerApi.buildIncomePostings(
      txForm.value.accountId,
      txForm.value.total,
      effectiveLegs.value,
    )
  } else {
    postings = ledgerApi.buildExpensePostings(
      txForm.value.accountId,
      txForm.value.total,
      effectiveLegs.value,
    )
  }
  const payload = {
    type: txForm.value.type,
    bookedAt,
    remark: txForm.value.remark,
    postings,
  }
  try {
    if (isEditingTx.value) {
      await ledgerStore.updateTransaction({ id: editingTxId.value, ...payload })
    } else {
      await ledgerStore.createTransaction(payload)
    }
    showTxModal.value = false
  } catch (err) {
    alert(ledgerErrorMessage(err, '操作失败'))
  } finally {
    isSubmittingTx.value = false
  }
}

async function handleDeleteTx(tx: LedgerTransaction) {
  if (!confirm('确定要删除这条交易吗？')) return
  try {
    await ledgerStore.deleteTransaction(tx.id)
  } catch (err) {
    alert(ledgerErrorMessage(err, '删除失败'))
  }
}

function closeTxModal() {
  showTxModal.value = false
}
</script>

<template>
  <div class="max-w-[1100px] mx-auto px-5 py-10">
    <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between mb-6">
      <div>
        <h1 class="text-[32px] font-semibold tracking-tight text-[#1d1d1f]">记账</h1>
        <p class="mt-1 text-sm text-[#86868b]">记录每一笔收支与转账</p>
      </div>
      <div class="flex flex-wrap items-center gap-3">
        <LedgerNav />
        <button
          @click="openCreateTxModal"
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
          记一笔
        </button>
      </div>
    </div>

    <div class="flex flex-wrap items-center gap-3 mb-6">
      <div
        class="flex items-center gap-1 bg-white border border-[#e8e8ed] rounded-xl px-1.5 py-1"
      >
        <button
          @click="prevMonth"
          class="p-1.5 rounded-lg text-[#86868b] hover:text-[#1d1d1f] hover:bg-[#f5f5f7] transition-all"
          title="上月"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M15 19l-7-7 7-7"
            />
          </svg>
        </button>
        <input
          v-model="filterMonth"
          type="month"
          class="ledger-filter-month px-2 py-1 bg-transparent text-[14px] text-[#1d1d1f] outline-none"
        />
        <button
          @click="nextMonth"
          class="p-1.5 rounded-lg text-[#86868b] hover:text-[#1d1d1f] hover:bg-[#f5f5f7] transition-all"
          title="下月"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </button>
      </div>
      <select
        v-model="filterAccountId"
        class="ledger-filter-account px-4 py-2.5 bg-white border border-[#e8e8ed] rounded-xl text-[14px] text-[#1d1d1f] outline-none transition-all focus:border-[#0071e3]"
      >
        <option value="">全部账户</option>
        <option v-for="account in activeAccounts" :key="account.id" :value="account.id">
          {{ account.name }}
        </option>
      </select>
      <select
        v-model="filterCategoryId"
        class="ledger-filter-category px-4 py-2.5 bg-white border border-[#e8e8ed] rounded-xl text-[14px] text-[#1d1d1f] outline-none transition-all focus:border-[#0071e3]"
      >
        <option value="">全部分类</option>
        <optgroup label="支出">
          <option v-for="cat in expenseCategories" :key="cat.id" :value="cat.id">
            {{ cat.parentId && cat.parentId !== '0' ? '└ ' + cat.name : cat.name }}
          </option>
        </optgroup>
        <optgroup label="收入">
          <option v-for="cat in incomeCategories" :key="cat.id" :value="cat.id">
            {{ cat.parentId && cat.parentId !== '0' ? '└ ' + cat.name : cat.name }}
          </option>
        </optgroup>
      </select>
      <select
        v-model="filterType"
        class="ledger-filter-type px-4 py-2.5 bg-white border border-[#e8e8ed] rounded-xl text-[14px] text-[#1d1d1f] outline-none transition-all focus:border-[#0071e3]"
      >
        <option value="">全部类型</option>
        <option value="expense">支出</option>
        <option value="income">收入</option>
        <option value="transfer">转账</option>
      </select>
    </div>

    <div v-if="txLoading" class="space-y-3">
      <div v-for="n in 5" :key="n" class="bg-white rounded-2xl p-4 border border-[#f0f0f0]">
        <div class="h-4 bg-[#f5f5f7] rounded w-1/3 mb-2 animate-pulse" />
        <div class="h-3 bg-[#f5f5f7] rounded w-1/4 animate-pulse" />
      </div>
    </div>

    <div
      v-else-if="transactions.length === 0"
      class="text-center py-24 bg-white rounded-2xl border border-[#f0f0f0]"
    >
      <h3 class="text-xl font-semibold text-[#1d1d1f] mb-2">本月暂无交易</h3>
      <p class="text-sm text-[#86868b] mb-6">点击"记一笔"开始记录</p>
      <button
        @click="openCreateTxModal"
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
        记一笔
      </button>
    </div>

    <div
      v-else
      class="bg-white rounded-2xl border border-[#f0f0f0] divide-y divide-[#f5f5f7] overflow-hidden"
    >
      <div
        v-for="tx in transactions"
        :key="tx.id"
        class="ledger-tx-item flex items-center gap-4 px-5 py-4 hover:bg-[#fafafc] transition-colors"
      >
        <div
          class="w-9 h-9 rounded-xl flex items-center justify-center shrink-0"
          :class="typeIconClass(tx.type)"
        >
          <svg
            v-if="tx.type === 'expense'"
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
          <svg
            v-else-if="tx.type === 'income'"
            class="w-4 h-4"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 19V5m0 0l-5 5m5-5l5 5"
            />
          </svg>
          <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M8 7h13m0 0l-3-3m3 3l-3 3M16 17H3m0 0l3 3m-3-3l3-3"
            />
          </svg>
        </div>
        <div class="flex-1 min-w-0">
          <p class="text-[15px] font-medium text-[#1d1d1f] truncate">{{ txCategoryText(tx) }}</p>
          <p class="text-xs text-[#86868b] truncate">
            {{ txAccountText(tx) }} · {{ tx.bookedAt?.slice(0, 16)
            }}<span v-if="tx.remark"> · {{ tx.remark }}</span>
          </p>
        </div>
        <p class="text-[16px] font-semibold shrink-0" :class="amountColor(tx.type)">
          {{ txAmountText(tx) }}
        </p>
        <div class="inline-flex gap-1 shrink-0">
          <button
            @click="openEditTxModal(tx)"
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
            @click="handleDeleteTx(tx)"
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

    <div v-if="txTotalPages > 1 && !txLoading" class="flex justify-center items-center gap-1.5 mt-8">
      <button
        @click="changeTxPage(txPage - 1)"
        :disabled="txPage <= 1"
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
      <span class="text-sm text-[#86868b] px-3">{{ txPage }} / {{ txTotalPages }}</span>
      <button
        @click="changeTxPage(txPage + 1)"
        :disabled="txPage >= txTotalPages"
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
      v-if="showTxModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4"
      style="background-color: rgba(0, 0, 0, 0.35)"
      @click.self="closeTxModal"
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
          v-if="showTxModal"
          class="ledger-tx-modal bg-white w-full max-w-2xl rounded-2xl shadow-2xl overflow-hidden max-h-[90vh] overflow-y-auto"
        >
          <div class="flex justify-between items-center px-6 py-4 border-b border-[#f0f0f0]">
            <h3 class="text-lg font-semibold text-[#1d1d1f]">
              {{ isEditingTx ? '编辑交易' : '记一笔' }}
            </h3>
            <button
              @click="closeTxModal"
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
              v-for="tab in txTypeTabs"
              :key="tab"
              @click="switchTxType(tab)"
              :disabled="isEditingTx"
              class="tx-tab flex-1 py-2 text-sm font-medium rounded-xl transition-all disabled:cursor-not-allowed disabled:opacity-80"
              :class="tabActiveClass(tab)"
            >
              {{ TRANSACTION_TYPE_LABELS[tab] }}
            </button>
          </div>

          <div class="px-6 py-5 space-y-4">
            <template v-if="txForm.type === 'transfer'">
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">转出账户</label>
                  <select
                    v-model="txForm.accountId"
                    class="tx-from-account w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
                  >
                    <option value="" disabled>请选择转出账户</option>
                    <option v-for="account in activeAccounts" :key="account.id" :value="account.id">
                      {{ account.name }}
                    </option>
                  </select>
                </div>
                <div>
                  <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">转入账户</label>
                  <select
                    v-model="txForm.toAccountId"
                    class="tx-to-account w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
                  >
                    <option value="" disabled>请选择转入账户</option>
                    <option v-for="account in activeAccounts" :key="account.id" :value="account.id">
                      {{ account.name }}
                    </option>
                  </select>
                </div>
                <div class="md:col-span-2">
                  <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">金额</label>
                  <input
                    v-model="txForm.total"
                    type="number"
                    step="0.01"
                    min="0"
                    placeholder="请输入金额"
                    class="tx-total w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
                  />
                </div>
              </div>
            </template>

            <template v-else>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">
                    {{ txForm.type === 'expense' ? '付款账户' : '收款账户' }}
                  </label>
                  <select
                    v-model="txForm.accountId"
                    class="tx-account w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
                  >
                    <option value="" disabled>请选择账户</option>
                    <option v-for="account in activeAccounts" :key="account.id" :value="account.id">
                      {{ account.name }}
                    </option>
                  </select>
                </div>
                <div>
                  <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">总金额</label>
                  <input
                    v-model="txForm.total"
                    type="number"
                    step="0.01"
                    min="0"
                    placeholder="请输入总金额"
                    class="tx-total w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
                  />
                </div>
              </div>

              <div v-if="!splitEnabled" class="tx-single-category grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">分类</label>
                  <select
                    v-model="txForm.categoryId"
                    @change="txForm.subCategoryId = ''"
                    class="tx-category w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
                  >
                    <option value="" disabled>请选择分类</option>
                    <option v-for="cat in topCategories" :key="cat.id" :value="cat.id">
                      {{ cat.name }}
                    </option>
                  </select>
                </div>
                <div v-if="childrenOf(txForm.categoryId).length > 0">
                  <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">子分类</label>
                  <select
                    v-model="txForm.subCategoryId"
                    class="tx-sub-category w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
                  >
                    <option value="">不选子分类</option>
                    <option
                      v-for="cat in childrenOf(txForm.categoryId)"
                      :key="cat.id"
                      :value="cat.id"
                    >
                      {{ cat.name }}
                    </option>
                  </select>
                </div>
              </div>

              <div v-else class="tx-legs space-y-2">
                <div
                  v-for="(leg, index) in txLegs"
                  :key="index"
                  class="tx-leg flex flex-wrap items-center gap-2"
                >
                  <input
                    v-model="leg.amount"
                    type="number"
                    step="0.01"
                    min="0"
                    placeholder="金额"
                    class="tx-leg-amount w-28 px-3 py-2 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[14px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
                  />
                  <select
                    v-model="leg.categoryId"
                    @change="leg.subCategoryId = ''"
                    class="tx-leg-category flex-1 min-w-[110px] px-3 py-2 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[14px] text-[#1d1d1f] outline-none transition-all focus:border-[#0071e3] focus:bg-white"
                  >
                    <option value="" disabled>分类</option>
                    <option v-for="cat in topCategories" :key="cat.id" :value="cat.id">
                      {{ cat.name }}
                    </option>
                  </select>
                  <select
                    v-if="childrenOf(leg.categoryId).length > 0"
                    v-model="leg.subCategoryId"
                    class="tx-leg-sub-category flex-1 min-w-[110px] px-3 py-2 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[14px] text-[#1d1d1f] outline-none transition-all focus:border-[#0071e3] focus:bg-white"
                  >
                    <option value="">子分类</option>
                    <option
                      v-for="cat in childrenOf(leg.categoryId)"
                      :key="cat.id"
                      :value="cat.id"
                    >
                      {{ cat.name }}
                    </option>
                  </select>
                  <button
                    @click="removeLeg(index)"
                    title="删除该条"
                    class="tx-leg-remove p-1.5 rounded-lg text-[#ff3b30] hover:bg-[#ff3b30]/10 transition-colors"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M6 18L18 6M6 6l12 12"
                      />
                    </svg>
                  </button>
                </div>
                <button
                  @click="addLeg"
                  class="tx-add-leg inline-flex items-center gap-1 text-sm font-medium text-[#0071e3] hover:underline"
                >
                  + 添加一条
                </button>
                <p class="text-xs" :class="splitBalanced ? 'text-[#86868b]' : 'text-[#ff3b30]'">
                  已拆分 ¥{{ legSumText }} / 总额 ¥{{ formatAmount(txForm.total || 0) }}
                  <span v-if="!splitBalanced">（差额 ¥{{ splitDiffText }}）</span>
                </p>
              </div>

              <div>
                <button
                  @click="toggleSplit"
                  class="tx-split-toggle text-sm font-medium text-[#0071e3] hover:underline"
                >
                  {{ splitEnabled ? '取消拆分' : '拆分' }}
                </button>
              </div>
            </template>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">时间</label>
                <input
                  v-model="txForm.bookedAt"
                  type="datetime-local"
                  class="tx-booked-at w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
                />
              </div>
              <div>
                <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">备注</label>
                <input
                  v-model="txForm.remark"
                  type="text"
                  placeholder="请输入备注"
                  class="tx-remark w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
                />
              </div>
            </div>
          </div>

          <div class="flex justify-end gap-2.5 px-6 py-4 border-t border-[#f0f0f0] bg-[#fafafc]/50">
            <button
              @click="closeTxModal"
              class="px-5 py-2 text-sm font-medium text-[#1d1d1f] bg-white border border-[#e8e8ed] rounded-xl hover:bg-[#f5f5f7] transition-colors"
            >
              取消
            </button>
            <button
              @click="handleTxSubmit"
              :disabled="!canSubmitTx || isSubmittingTx"
              class="px-5 py-2 text-sm font-medium text-white rounded-xl transition-all hover:scale-[1.02] active:scale-[0.98] disabled:opacity-40 disabled:scale-100 disabled:cursor-not-allowed"
              style="background: linear-gradient(135deg, #0071e3, #0063c7)"
            >
              {{ isSubmittingTx ? '保存中...' : '保存' }}
            </button>
          </div>
        </div>
      </Transition>
    </div>
  </Transition>
</template>
