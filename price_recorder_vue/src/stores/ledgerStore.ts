import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { useToast } from 'vue-toastification'
import * as ledgerApi from '@/api/ledger'
import { ledgerErrorMessage } from '@/api/ledger'
import type {
  LedgerAccount,
  LedgerAccountRequest,
  LedgerBalancePoint,
  LedgerBalanceTrendParams,
  LedgerBudget,
  LedgerBudgetRequest,
  LedgerCategory,
  LedgerCategoryRequest,
  LedgerMonthlyStats,
  LedgerRecurring,
  LedgerRecurringRequest,
  LedgerTransaction,
  LedgerTransactionRequest,
} from '@/api/ledger'
import {
  computeCreditReminder,
  creditReminderQueryRange,
  type CreditReminder,
} from '@/utils/creditReminder'

export const useLedgerStore = defineStore('ledger', () => {
  const toast = useToast()
  const accounts = ref<LedgerAccount[]>([])
  const categories = ref<LedgerCategory[]>([])
  const transactions = ref<LedgerTransaction[]>([])
  const monthlyStats = ref<LedgerMonthlyStats | null>(null)
  const budgets = ref<LedgerBudget[]>([])
  const balanceTrend = ref<LedgerBalancePoint[]>([])
  const recurringList = ref<LedgerRecurring[]>([])
  const creditReminders = ref<CreditReminder[]>([])
  const accountsLoading = ref(false)
  const categoriesLoading = ref(false)
  const txLoading = ref(false)
  const statsLoading = ref(false)
  const budgetsLoading = ref(false)
  const trendLoading = ref(false)
  const recurringLoading = ref(false)
  const creditRemindersLoading = ref(false)
  const txTotal = ref(0)
  const txPage = ref(1)
  const txPageSize = ref(10)

  // 列表页筛选状态：月份 + 账户/分类/类型
  const filterMonth = ref(ledgerApi.currentMonth())
  const filterAccountId = ref('')
  const filterCategoryId = ref('')
  const filterType = ref('')

  const txTotalPages = computed(() => Math.max(1, Math.ceil(txTotal.value / txPageSize.value)))
  const activeAccounts = computed(() => accounts.value.filter((account) => !account.archived))
  const expenseCategories = computed(() =>
    categories.value.filter((category) => category.direction === 'expense'),
  )
  const incomeCategories = computed(() =>
    categories.value.filter((category) => category.direction === 'income'),
  )
  const accountMap = computed(() => {
    const map = new Map<string, LedgerAccount>()
    accounts.value.forEach((account) => map.set(account.id, account))
    return map
  })
  const categoryMap = computed(() => {
    const map = new Map<string, LedgerCategory>()
    categories.value.forEach((category) => map.set(category.id, category))
    return map
  })

  async function fetchAccounts(): Promise<void> {
    accountsLoading.value = true
    try {
      const response = await ledgerApi.getAccountList()
      accounts.value = response.list || []
    } catch (err) {
      toast.error(ledgerErrorMessage(err, '获取账户列表失败'))
      throw err
    } finally {
      accountsLoading.value = false
    }
  }

  async function saveAccount(data: LedgerAccountRequest, isEdit: boolean): Promise<void> {
    accountsLoading.value = true
    try {
      if (isEdit) {
        await ledgerApi.updateAccount(data)
        toast.success('账户更新成功')
      } else {
        await ledgerApi.saveAccount(data)
        toast.success('账户创建成功')
      }
      await fetchAccounts()
    } catch (err) {
      toast.error(ledgerErrorMessage(err, '保存账户失败'))
      throw err
    } finally {
      accountsLoading.value = false
    }
  }

  async function removeAccount(id: string): Promise<void> {
    accountsLoading.value = true
    try {
      await ledgerApi.deleteAccount(id)
      toast.success('账户删除成功')
      await fetchAccounts()
    } catch (err) {
      // 后端拒绝（如账户下仍有交易）时 toast 后端错误信息
      toast.error(ledgerErrorMessage(err, '删除账户失败'))
      throw err
    } finally {
      accountsLoading.value = false
    }
  }

  async function fetchCategories(): Promise<void> {
    categoriesLoading.value = true
    try {
      const response = await ledgerApi.getCategoryList()
      categories.value = response.list || []
    } catch (err) {
      toast.error(ledgerErrorMessage(err, '获取分类列表失败'))
      throw err
    } finally {
      categoriesLoading.value = false
    }
  }

  async function saveCategory(data: LedgerCategoryRequest, isEdit: boolean): Promise<void> {
    categoriesLoading.value = true
    try {
      if (isEdit) {
        await ledgerApi.updateCategory(data)
        toast.success('分类更新成功')
      } else {
        await ledgerApi.saveCategory(data)
        toast.success('分类创建成功')
      }
      await fetchCategories()
    } catch (err) {
      toast.error(ledgerErrorMessage(err, '保存分类失败'))
      throw err
    } finally {
      categoriesLoading.value = false
    }
  }

  async function removeCategory(id: string): Promise<void> {
    categoriesLoading.value = true
    try {
      await ledgerApi.deleteCategory(id)
      toast.success('分类删除成功')
      await fetchCategories()
    } catch (err) {
      toast.error(ledgerErrorMessage(err, '删除分类失败'))
      throw err
    } finally {
      categoriesLoading.value = false
    }
  }

  // 月份切换转化为 startTime/endTime 查询当月交易
  async function fetchTransactions(page?: number): Promise<void> {
    txLoading.value = true
    try {
      const pageNum = page ?? txPage.value
      const { startTime, endTime } = ledgerApi.monthRange(filterMonth.value)
      const response = await ledgerApi.getTransactionPage({
        page: String(pageNum),
        pageSize: String(txPageSize.value),
        accountId: filterAccountId.value || undefined,
        categoryId: filterCategoryId.value || undefined,
        type: filterType.value || undefined,
        startTime,
        endTime,
      })
      transactions.value = response.list || []
      txTotal.value = Number(response.total || 0)
      txPage.value = pageNum
    } catch (err) {
      toast.error(ledgerErrorMessage(err, '获取交易列表失败'))
      throw err
    } finally {
      txLoading.value = false
    }
  }

  async function createTransaction(data: LedgerTransactionRequest): Promise<void> {
    txLoading.value = true
    try {
      await ledgerApi.createTransaction(data)
      toast.success('记账成功')
      await fetchTransactions()
    } catch (err) {
      toast.error(ledgerErrorMessage(err, '记账失败'))
      throw err
    } finally {
      txLoading.value = false
    }
  }

  async function updateTransaction(data: LedgerTransactionRequest): Promise<void> {
    txLoading.value = true
    try {
      await ledgerApi.updateTransaction(data)
      toast.success('交易更新成功')
      await fetchTransactions()
    } catch (err) {
      toast.error(ledgerErrorMessage(err, '更新失败'))
      throw err
    } finally {
      txLoading.value = false
    }
  }

  async function deleteTransaction(id: string): Promise<void> {
    txLoading.value = true
    try {
      await ledgerApi.deleteTransaction(id)
      toast.success('交易删除成功')
      await fetchTransactions()
    } catch (err) {
      toast.error(ledgerErrorMessage(err, '删除失败'))
      throw err
    } finally {
      txLoading.value = false
    }
  }

  async function fetchMonthlyStats(month: string): Promise<void> {
    statsLoading.value = true
    try {
      monthlyStats.value = await ledgerApi.getMonthlyStats(month)
    } catch (err) {
      toast.error(ledgerErrorMessage(err, '获取月度统计失败'))
      throw err
    } finally {
      statsLoading.value = false
    }
  }

  async function fetchBudgets(month: string): Promise<void> {
    budgetsLoading.value = true
    try {
      const response = await ledgerApi.getBudgetList(month)
      budgets.value = response.list || []
    } catch (err) {
      toast.error(ledgerErrorMessage(err, '获取预算列表失败'))
      throw err
    } finally {
      budgetsLoading.value = false
    }
  }

  // upsert 语义：保存后按预算所属月份刷新，保证报表页切换到对应月时数据一致
  async function saveBudget(data: LedgerBudgetRequest): Promise<void> {
    budgetsLoading.value = true
    try {
      await ledgerApi.saveBudget(data)
      toast.success('预算保存成功')
      await fetchBudgets(data.month)
    } catch (err) {
      toast.error(ledgerErrorMessage(err, '保存预算失败'))
      throw err
    } finally {
      budgetsLoading.value = false
    }
  }

  async function removeBudget(id: string, month: string): Promise<void> {
    budgetsLoading.value = true
    try {
      await ledgerApi.deleteBudget(id)
      toast.success('预算删除成功')
      await fetchBudgets(month)
    } catch (err) {
      toast.error(ledgerErrorMessage(err, '删除预算失败'))
      throw err
    } finally {
      budgetsLoading.value = false
    }
  }

  async function fetchBalanceTrend(params: LedgerBalanceTrendParams): Promise<void> {
    trendLoading.value = true
    try {
      const response = await ledgerApi.getBalanceTrend(params)
      balanceTrend.value = response.points || []
    } catch (err) {
      toast.error(ledgerErrorMessage(err, '获取余额走势失败'))
      throw err
    } finally {
      trendLoading.value = false
    }
  }

  async function fetchRecurring(): Promise<void> {
    recurringLoading.value = true
    try {
      const response = await ledgerApi.getRecurringList()
      recurringList.value = response.list || []
    } catch (err) {
      toast.error(ledgerErrorMessage(err, '获取周期账单失败'))
      throw err
    } finally {
      recurringLoading.value = false
    }
  }

  // save 幂等：无 id 创建、带 id 更新，保存后刷新列表
  async function saveRecurring(data: LedgerRecurringRequest): Promise<void> {
    recurringLoading.value = true
    try {
      await ledgerApi.saveRecurring(data)
      toast.success(data.id ? '周期账单更新成功' : '周期账单创建成功')
      await fetchRecurring()
    } catch (err) {
      toast.error(ledgerErrorMessage(err, '保存周期账单失败'))
      throw err
    } finally {
      recurringLoading.value = false
    }
  }

  async function removeRecurring(id: string): Promise<void> {
    recurringLoading.value = true
    try {
      await ledgerApi.deleteRecurring(id)
      toast.success('周期账单删除成功')
      await fetchRecurring()
    } catch (err) {
      toast.error(ledgerErrorMessage(err, '删除周期账单失败'))
      throw err
    } finally {
      recurringLoading.value = false
    }
  }

  // 静默生成到期账单：失败不打断用户操作，返回 created 供页面决定是否刷新
  async function applyRecurring(): Promise<number> {
    try {
      const response = await ledgerApi.applyRecurring()
      return Number(response.created || 0)
    } catch (err) {
      console.warn('周期账单自动生成失败', err)
      return 0
    }
  }

  // 纯前端提醒：按账户账单区间拉交易再本地计算；辅助功能失败仅 console.warn
  async function fetchCreditReminders(): Promise<void> {
    const targets = accounts.value.filter(
      (account) => account.billingDay && account.paymentDueDay,
    )
    if (targets.length === 0) {
      creditReminders.value = []
      return
    }
    creditRemindersLoading.value = true
    try {
      const today = new Date()
      const reminders: CreditReminder[] = []
      await Promise.all(
        targets.map(async (account) => {
          const range = creditReminderQueryRange(account, today)
          if (!range) return
          const response = await ledgerApi.getTransactionPage({
            page: '1',
            pageSize: '500',
            accountId: account.id,
            startTime: range.startTime,
            endTime: range.endTime,
          })
          const reminder = computeCreditReminder(account, response.list || [], today)
          if (reminder) reminders.push(reminder)
        }),
      )
      creditReminders.value = reminders.sort((a, b) => a.daysUntilDue - b.daysUntilDue)
    } catch (err) {
      console.warn('信用卡还款提醒加载失败', err)
    } finally {
      creditRemindersLoading.value = false
    }
  }

  return {
    accounts,
    categories,
    transactions,
    monthlyStats,
    budgets,
    balanceTrend,
    recurringList,
    creditReminders,
    accountsLoading,
    categoriesLoading,
    txLoading,
    statsLoading,
    budgetsLoading,
    trendLoading,
    recurringLoading,
    creditRemindersLoading,
    txTotal,
    txPage,
    txPageSize,
    filterMonth,
    filterAccountId,
    filterCategoryId,
    filterType,
    txTotalPages,
    activeAccounts,
    expenseCategories,
    incomeCategories,
    accountMap,
    categoryMap,
    fetchAccounts,
    saveAccount,
    removeAccount,
    fetchCategories,
    saveCategory,
    removeCategory,
    fetchTransactions,
    createTransaction,
    updateTransaction,
    deleteTransaction,
    fetchMonthlyStats,
    fetchBudgets,
    saveBudget,
    removeBudget,
    fetchBalanceTrend,
    fetchRecurring,
    saveRecurring,
    removeRecurring,
    applyRecurring,
    fetchCreditReminders,
  }
})
