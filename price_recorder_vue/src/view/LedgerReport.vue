<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useLedgerStore } from '@/stores/ledgerStore'
import * as ledgerApi from '@/api/ledger'
import { formatAmount, ledgerErrorMessage } from '@/api/ledger'
import type { LedgerBudget } from '@/api/ledger'
import { reminderBadgeText } from '@/utils/creditReminder'
import LedgerNav from '@/components/LedgerNav.vue'

const ledgerStore = useLedgerStore()
const {
  monthlyStats,
  statsLoading,
  budgets,
  budgetsLoading,
  balanceTrend,
  trendLoading,
  activeAccounts,
  expenseCategories,
  creditReminders,
} = storeToRefs(ledgerStore)

const reportMonth = ref(ledgerApi.currentMonth())

onMounted(() => {
  ledgerStore.fetchMonthlyStats(reportMonth.value)
  ledgerStore.fetchBudgets(reportMonth.value)
  loadAccountsAndReminders()
  ledgerStore.fetchCategories()
  fetchTrend()
})

async function loadAccountsAndReminders() {
  try {
    await ledgerStore.fetchAccounts()
  } catch {
    return // 账户加载失败已 toast，跳过提醒计算
  }
  ledgerStore.fetchCreditReminders()
}
watch(reportMonth, (month) => {
  if (month) {
    ledgerStore.fetchMonthlyStats(month)
    ledgerStore.fetchBudgets(month)
  }
})

function prevMonth() {
  reportMonth.value = ledgerApi.shiftMonth(reportMonth.value, -1)
}

function nextMonth() {
  reportMonth.value = ledgerApi.shiftMonth(reportMonth.value, 1)
}

interface CategoryRow {
  categoryId: string
  categoryName: string
  amount: string
  widthPercent: number
  ratioPercent: number
}

// 横向条形排行：按 amount 降序；条宽相对最大值，百分比相对合计
function buildRows(
  list: { categoryId: string; categoryName: string; amount: string }[] | undefined,
  total: string | undefined,
): CategoryRow[] {
  const items = [...(list || [])].sort((a, b) => Number(b.amount || 0) - Number(a.amount || 0))
  const totalNum = Number(total || 0)
  const max = Math.max(...items.map((item) => Number(item.amount || 0)), 0)
  return items.map((item) => {
    const amount = Number(item.amount || 0)
    return {
      ...item,
      widthPercent: max > 0 ? (amount / max) * 100 : 0,
      ratioPercent: totalNum > 0 ? (amount / totalNum) * 100 : 0,
    }
  })
}

const expenseRows = computed(() =>
  buildRows(monthlyStats.value?.expenseByCategory, monthlyStats.value?.totalExpense),
)
const incomeRows = computed(() =>
  buildRows(monthlyStats.value?.incomeByCategory, monthlyStats.value?.totalIncome),
)

/* ---------- 分类预算 ---------- */

interface BudgetRow {
  id: string
  categoryId: string
  categoryName: string
  amount: string
  used: string
  usedPercent: number
  widthPercent: number
  over: boolean
}

// 条宽封顶 100%，超支（used > amount）标红
const budgetRows = computed<BudgetRow[]>(() =>
  budgets.value.map((budget) => {
    const amount = Number(budget.amount || 0)
    const used = Number(budget.used || 0)
    const ratio = amount > 0 ? used / amount : 0
    return {
      ...budget,
      usedPercent: ratio * 100,
      widthPercent: Math.min(ratio, 1) * 100,
      over: used > amount,
    }
  }),
)

const showBudgetModal = ref(false)
const isEditingBudget = ref(false)
const isSubmittingBudget = ref(false)
const budgetForm = ref({
  categoryId: '',
  subCategoryId: '',
  amount: '' as string | number,
  month: '',
})

// 仅 expense 方向分类可设预算；两级联动与记账弹窗同一写法
const budgetTopCategories = computed(() =>
  expenseCategories.value.filter((c) => !c.parentId || c.parentId === '0'),
)

function budgetChildrenOf(parentId: string) {
  if (!parentId) return []
  return expenseCategories.value.filter((c) => c.parentId === parentId)
}

const canSubmitBudget = computed(() => {
  const categoryId = budgetForm.value.subCategoryId || budgetForm.value.categoryId
  return !!categoryId && Number(budgetForm.value.amount) > 0 && !!budgetForm.value.month
})

function openCreateBudgetModal() {
  isEditingBudget.value = false
  budgetForm.value = { categoryId: '', subCategoryId: '', amount: '', month: reportMonth.value }
  showBudgetModal.value = true
}

// 列表项无 month 字段，编辑时月份取报表当前月；二级分类拆回父/子两个 select
function openEditBudgetModal(budget: LedgerBudget) {
  isEditingBudget.value = true
  const categoryId = String(budget.categoryId)
  const parentId = ledgerStore.categoryMap.get(categoryId)?.parentId
  const hasParent = !!parentId && parentId !== '0'
  budgetForm.value = {
    categoryId: hasParent ? parentId : categoryId,
    subCategoryId: hasParent ? categoryId : '',
    amount: budget.amount || '',
    month: reportMonth.value,
  }
  showBudgetModal.value = true
}

async function handleBudgetSubmit() {
  if (!canSubmitBudget.value) return
  isSubmittingBudget.value = true
  try {
    await ledgerStore.saveBudget({
      categoryId: budgetForm.value.subCategoryId || budgetForm.value.categoryId,
      month: budgetForm.value.month,
      amount: budgetForm.value.amount,
    })
    showBudgetModal.value = false
  } catch (err) {
    alert(ledgerErrorMessage(err, '操作失败'))
  } finally {
    isSubmittingBudget.value = false
  }
}

async function handleDeleteBudget(budget: LedgerBudget) {
  if (!confirm(`确定要删除「${budget.categoryName}」的预算吗？`)) return
  try {
    await ledgerStore.removeBudget(budget.id, reportMonth.value)
  } catch (err) {
    alert(ledgerErrorMessage(err, '删除失败'))
  }
}

function closeBudgetModal() {
  showBudgetModal.value = false
}

/* ---------- 余额走势 ---------- */

const trendAccountId = ref('')
const trendRangeMonths = ref(6)
const trendRangeOptions = [
  { label: '近 3 月', months: 3 },
  { label: '近 6 月', months: 6 },
  { label: '近 1 年', months: 12 },
]

function fetchTrend() {
  const { startTime, endTime } = ledgerApi.trendDateRange(trendRangeMonths.value)
  ledgerStore.fetchBalanceTrend({
    accountId: trendAccountId.value || undefined,
    startTime,
    endTime,
  })
}

watch([trendAccountId, trendRangeMonths], fetchTrend)

// 手绘 SVG 折线：viewBox 固定 640x220，坐标按 min/max 归一化，x 按日序均分
const TREND_WIDTH = 640
const TREND_HEIGHT = 220
const TREND_PAD_X = 8
const TREND_PAD_Y = 16

interface TrendChartPoint {
  x: number
  y: number
}

const trendChartPoints = computed<TrendChartPoint[]>(() => {
  const points = balanceTrend.value
  if (points.length === 0) return []
  const values = points.map((p) => Number(p.balance || 0))
  const min = Math.min(...values)
  const max = Math.max(...values)
  const spanX = TREND_WIDTH - TREND_PAD_X * 2
  const spanY = TREND_HEIGHT - TREND_PAD_Y * 2
  const stepX = points.length > 1 ? spanX / (points.length - 1) : 0
  return points.map((p, index) => {
    const ratio = max > min ? (Number(p.balance || 0) - min) / (max - min) : 0.5
    return {
      x: Number((TREND_PAD_X + stepX * index).toFixed(2)),
      y: Number((TREND_PAD_Y + (1 - ratio) * spanY).toFixed(2)),
    }
  })
})

const trendPointsAttr = computed(() =>
  trendChartPoints.value.map((p) => `${p.x},${p.y}`).join(' '),
)

// 面积 polygon：折线点 + 底部两角闭合
const trendAreaAttr = computed(() => {
  const points = trendChartPoints.value
  const first = points[0]
  const last = points[points.length - 1]
  if (!first || !last) return ''
  const baseline = TREND_HEIGHT - TREND_PAD_Y
  return `${first.x},${baseline} ${trendPointsAttr.value} ${last.x},${baseline}`
})

const trendBalances = computed(() => balanceTrend.value.map((p) => Number(p.balance || 0)))
const trendMin = computed(() =>
  trendBalances.value.length ? Math.min(...trendBalances.value) : 0,
)
const trendMax = computed(() =>
  trendBalances.value.length ? Math.max(...trendBalances.value) : 0,
)
const trendStartDate = computed(() => balanceTrend.value[0]?.date || '')
const trendEndDate = computed(() => balanceTrend.value[balanceTrend.value.length - 1]?.date || '')
</script>

<template>
  <div class="max-w-[1100px] mx-auto px-5 py-10">
    <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between mb-6">
      <div>
        <h1 class="text-[32px] font-semibold tracking-tight text-[#1d1d1f]">月度报表</h1>
        <p class="mt-1 text-sm text-[#86868b]">按月汇总收支与分类占比</p>
      </div>
      <div class="flex flex-wrap items-center gap-3">
        <LedgerNav />
        <div class="flex items-center gap-1 bg-white border border-[#e8e8ed] rounded-xl px-1.5 py-1">
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
            v-model="reportMonth"
            type="month"
            class="report-month px-2 py-1 bg-transparent text-[14px] text-[#1d1d1f] outline-none"
          />
          <button
            @click="nextMonth"
            class="p-1.5 rounded-lg text-[#86868b] hover:text-[#1d1d1f] hover:bg-[#f5f5f7] transition-all"
            title="下月"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M9 5l7 7-7 7"
              />
            </svg>
          </button>
        </div>
      </div>
    </div>

    <div
      v-if="creditReminders.length > 0"
      class="credit-reminder-banner bg-white rounded-2xl border border-[#f0f0f0] p-5 mb-6"
    >
      <h2 class="text-lg font-semibold text-[#1d1d1f] mb-3">还款提醒</h2>
      <div class="space-y-2">
        <div
          v-for="reminder in creditReminders"
          :key="reminder.accountId"
          class="credit-reminder-row flex flex-wrap items-center justify-between gap-1 rounded-xl border px-4 py-3 text-sm"
          :class="
            reminder.level === 'warning'
              ? 'border-[#ff3b30]/40 bg-[#ff3b30]/5'
              : 'border-[#f0f0f0] bg-[#fafafc]'
          "
        >
          <span
            class="font-medium"
            :class="reminder.level === 'warning' ? 'text-[#ff3b30]' : 'text-[#1d1d1f]'"
          >
            {{ reminder.accountName }}
          </span>
          <span :class="reminder.level === 'warning' ? 'text-[#ff3b30]' : 'text-[#86868b]'">
            本期应还 ¥{{ formatAmount(reminder.amountDue) }} ·
            {{ reminderBadgeText(reminder) }}（{{ reminder.dueDate }}）
          </span>
        </div>
      </div>
    </div>

    <div v-if="statsLoading" class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
      <div v-for="n in 2" :key="n" class="bg-white rounded-2xl p-5 border border-[#f0f0f0]">
        <div class="h-4 bg-[#f5f5f7] rounded w-1/3 mb-3 animate-pulse" />
        <div class="h-6 bg-[#f5f5f7] rounded w-1/2 animate-pulse" />
      </div>
    </div>

    <template v-else>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-8">
        <div class="bg-white rounded-2xl p-5 border border-[#f0f0f0]">
          <p class="text-sm text-[#86868b] mb-2">{{ reportMonth }} 总支出</p>
          <p class="total-expense text-[26px] font-semibold text-[#ff3b30]">
            ¥{{ formatAmount(monthlyStats?.totalExpense) }}
          </p>
        </div>
        <div class="bg-white rounded-2xl p-5 border border-[#f0f0f0]">
          <p class="text-sm text-[#86868b] mb-2">{{ reportMonth }} 总收入</p>
          <p class="total-income text-[26px] font-semibold text-[#34c759]">
            ¥{{ formatAmount(monthlyStats?.totalIncome) }}
          </p>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-5">
        <div class="expense-rank bg-white rounded-2xl p-5 border border-[#f0f0f0]">
          <h2 class="text-lg font-semibold text-[#1d1d1f] mb-4">支出分类排行</h2>
          <p v-if="expenseRows.length === 0" class="text-sm text-[#86868b] py-6 text-center">
            本月暂无支出
          </p>
          <div v-else class="space-y-3">
            <div v-for="row in expenseRows" :key="row.categoryId" class="category-row">
              <div class="flex justify-between items-baseline text-sm mb-1">
                <span class="text-[#1d1d1f] font-medium">{{ row.categoryName }}</span>
                <span class="text-[#86868b]">
                  ¥{{ formatAmount(row.amount) }} · {{ row.ratioPercent.toFixed(1) }}%
                </span>
              </div>
              <div class="h-2 bg-[#f5f5f7] rounded-full overflow-hidden">
                <div
                  class="category-bar h-2 rounded-full bg-[#ff3b30] transition-all duration-500"
                  :style="{ width: row.widthPercent + '%' }"
                />
              </div>
            </div>
          </div>
        </div>

        <div class="income-rank bg-white rounded-2xl p-5 border border-[#f0f0f0]">
          <h2 class="text-lg font-semibold text-[#1d1d1f] mb-4">收入分类排行</h2>
          <p v-if="incomeRows.length === 0" class="text-sm text-[#86868b] py-6 text-center">
            本月暂无收入
          </p>
          <div v-else class="space-y-3">
            <div v-for="row in incomeRows" :key="row.categoryId" class="category-row">
              <div class="flex justify-between items-baseline text-sm mb-1">
                <span class="text-[#1d1d1f] font-medium">{{ row.categoryName }}</span>
                <span class="text-[#86868b]">
                  ¥{{ formatAmount(row.amount) }} · {{ row.ratioPercent.toFixed(1) }}%
                </span>
              </div>
              <div class="h-2 bg-[#f5f5f7] rounded-full overflow-hidden">
                <div
                  class="category-bar h-2 rounded-full bg-[#34c759] transition-all duration-500"
                  :style="{ width: row.widthPercent + '%' }"
                />
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="budget-section bg-white rounded-2xl p-5 border border-[#f0f0f0] mt-5">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-lg font-semibold text-[#1d1d1f]">分类预算</h2>
          <button
            @click="openCreateBudgetModal"
            class="budget-add inline-flex items-center gap-1.5 px-4 py-2 text-white text-sm font-medium rounded-xl transition-all hover:scale-[1.02] active:scale-[0.98] shadow-sm"
            style="background: linear-gradient(135deg, #0071e3, #0063c7)"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 4v16m8-8H4"
              />
            </svg>
            新增预算
          </button>
        </div>
        <div v-if="budgetsLoading" class="space-y-3">
          <div v-for="n in 2" :key="n">
            <div class="h-4 bg-[#f5f5f7] rounded w-1/3 mb-2 animate-pulse" />
            <div class="h-2 bg-[#f5f5f7] rounded-full animate-pulse" />
          </div>
        </div>
        <p v-else-if="budgetRows.length === 0" class="text-sm text-[#86868b] py-6 text-center">
          本月暂无预算，点击"新增预算"设置
        </p>
        <div v-else class="space-y-3">
          <div v-for="row in budgetRows" :key="row.id" class="budget-row">
            <div class="flex justify-between items-baseline text-sm mb-1">
              <span class="text-[#1d1d1f] font-medium">{{ row.categoryName }}</span>
              <span class="budget-usage text-[#86868b]">
                已用 ¥{{ formatAmount(row.used) }} / 预算 ¥{{ formatAmount(row.amount) }}
              </span>
            </div>
            <div class="flex items-center gap-2">
              <div class="h-2 flex-1 bg-[#f5f5f7] rounded-full overflow-hidden">
                <div
                  class="budget-bar h-2 rounded-full transition-all duration-500"
                  :class="row.over ? 'bg-[#ff3b30]' : 'bg-[#0071e3]'"
                  :style="{ width: row.widthPercent + '%' }"
                />
              </div>
              <span
                class="text-xs w-12 text-right shrink-0"
                :class="row.over ? 'text-[#ff3b30] font-medium' : 'text-[#86868b]'"
              >
                {{ row.usedPercent.toFixed(1) }}%
              </span>
              <div class="inline-flex gap-1 shrink-0">
                <button
                  @click="openEditBudgetModal(row)"
                  class="budget-edit p-1.5 rounded-lg text-[#0071e3] hover:bg-[#0071e3]/10 transition-colors"
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
                  @click="handleDeleteBudget(row)"
                  class="budget-delete p-1.5 rounded-lg text-[#ff3b30] hover:bg-[#ff3b30]/10 transition-colors"
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
      </div>
    </template>

    <div class="trend-section bg-white rounded-2xl p-5 border border-[#f0f0f0]">
      <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between mb-4">
        <h2 class="text-lg font-semibold text-[#1d1d1f]">余额走势</h2>
        <div class="flex flex-wrap items-center gap-2">
          <select
            v-model="trendAccountId"
            class="trend-account px-3 py-1.5 bg-white border border-[#e8e8ed] rounded-xl text-[13px] text-[#1d1d1f] outline-none transition-all focus:border-[#0071e3]"
          >
            <option value="">净资产</option>
            <option v-for="account in activeAccounts" :key="account.id" :value="account.id">
              {{ account.name }}
            </option>
          </select>
          <div class="flex items-center gap-0.5 bg-[#f5f5f7] rounded-lg p-0.5">
            <button
              v-for="opt in trendRangeOptions"
              :key="opt.months"
              @click="trendRangeMonths = opt.months"
              class="trend-range px-3 py-1 text-xs font-medium rounded-md transition-all"
              :class="
                trendRangeMonths === opt.months
                  ? 'bg-white text-[#0071e3] shadow-sm'
                  : 'text-[#86868b] hover:text-[#1d1d1f]'
              "
            >
              {{ opt.label }}
            </button>
          </div>
        </div>
      </div>
      <div v-if="trendLoading" class="h-[220px] flex items-center justify-center">
        <div class="h-4 bg-[#f5f5f7] rounded w-1/4 animate-pulse" />
      </div>
      <p v-else-if="trendChartPoints.length === 0" class="text-sm text-[#86868b] py-10 text-center">
        暂无余额数据
      </p>
      <template v-else>
        <div class="flex justify-between text-xs text-[#86868b] mb-2">
          <span class="trend-max">最高 ¥{{ formatAmount(trendMax) }}</span>
          <span class="trend-min">最低 ¥{{ formatAmount(trendMin) }}</span>
        </div>
        <svg
          class="trend-chart w-full h-auto"
          :viewBox="`0 0 ${TREND_WIDTH} ${TREND_HEIGHT}`"
          role="img"
        >
          <polygon class="trend-area" :points="trendAreaAttr" fill="#0071e3" fill-opacity="0.08" />
          <polyline
            class="trend-line"
            :points="trendPointsAttr"
            fill="none"
            stroke="#0071e3"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
        <div class="flex justify-between text-xs text-[#86868b] mt-2">
          <span class="trend-start">{{ trendStartDate }}</span>
          <span class="trend-end">{{ trendEndDate }}</span>
        </div>
      </template>
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
      v-if="showBudgetModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4"
      style="background-color: rgba(0, 0, 0, 0.35)"
      @click.self="closeBudgetModal"
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
          v-if="showBudgetModal"
          class="budget-modal bg-white w-full max-w-lg rounded-2xl shadow-2xl overflow-hidden max-h-[90vh] overflow-y-auto"
        >
          <div class="flex justify-between items-center px-6 py-4 border-b border-[#f0f0f0]">
            <h3 class="text-lg font-semibold text-[#1d1d1f]">
              {{ isEditingBudget ? '编辑预算' : '新增预算' }}
            </h3>
            <button
              @click="closeBudgetModal"
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
            <div>
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">分类</label>
              <select
                v-model="budgetForm.categoryId"
                @change="budgetForm.subCategoryId = ''"
                class="budget-category w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              >
                <option value="" disabled>请选择支出分类</option>
                <option v-for="cat in budgetTopCategories" :key="cat.id" :value="cat.id">
                  {{ cat.name }}
                </option>
              </select>
            </div>
            <div v-if="budgetChildrenOf(budgetForm.categoryId).length > 0">
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">子分类</label>
              <select
                v-model="budgetForm.subCategoryId"
                class="budget-sub-category w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              >
                <option value="">不选子分类</option>
                <option
                  v-for="cat in budgetChildrenOf(budgetForm.categoryId)"
                  :key="cat.id"
                  :value="cat.id"
                >
                  {{ cat.name }}
                </option>
              </select>
            </div>
            <div>
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">月份</label>
              <input
                v-model="budgetForm.month"
                type="month"
                class="budget-month w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              />
            </div>
            <div>
              <label class="block mb-1.5 text-sm font-medium text-[#1d1d1f]">预算金额</label>
              <input
                v-model="budgetForm.amount"
                type="number"
                step="0.01"
                min="0"
                placeholder="请输入预算金额"
                class="budget-amount w-full px-4 py-2.5 bg-[#fafafc] border border-[#e8e8ed] rounded-xl text-[15px] text-[#1d1d1f] outline-none transition-all placeholder:text-[#c7c7cc] focus:border-[#0071e3] focus:bg-white focus:ring-2 focus:ring-[#0071e3]/10"
              />
            </div>
          </div>
          <div class="flex justify-end gap-2.5 px-6 py-4 border-t border-[#f0f0f0] bg-[#fafafc]/50">
            <button
              @click="closeBudgetModal"
              class="px-5 py-2 text-sm font-medium text-[#1d1d1f] bg-white border border-[#e8e8ed] rounded-xl hover:bg-[#f5f5f7] transition-colors"
            >
              取消
            </button>
            <button
              @click="handleBudgetSubmit"
              :disabled="!canSubmitBudget || isSubmittingBudget"
              class="budget-submit px-5 py-2 text-sm font-medium text-white rounded-xl transition-all hover:scale-[1.02] active:scale-[0.98] disabled:opacity-40 disabled:scale-100 disabled:cursor-not-allowed"
              style="background: linear-gradient(135deg, #0071e3, #0063c7)"
            >
              {{ isSubmittingBudget ? '保存中...' : '保存' }}
            </button>
          </div>
        </div>
      </Transition>
    </div>
  </Transition>
</template>
