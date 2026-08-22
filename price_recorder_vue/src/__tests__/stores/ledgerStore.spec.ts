import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useLedgerStore } from '@/stores/ledgerStore'
import * as ledgerApi from '@/api/ledger'

const toastSuccess = vi.fn()
const toastError = vi.fn()

vi.mock('vue-toastification', () => ({
  useToast: () => ({
    success: toastSuccess,
    error: toastError,
  }),
}))

vi.mock('@/api/ledger', async () => {
  const actual = await vi.importActual<typeof import('@/api/ledger')>('@/api/ledger')
  return {
    ...actual,
    getAccountList: vi.fn(),
    saveAccount: vi.fn(),
    updateAccount: vi.fn(),
    deleteAccount: vi.fn(),
    getCategoryList: vi.fn(),
    saveCategory: vi.fn(),
    updateCategory: vi.fn(),
    deleteCategory: vi.fn(),
    getTransactionPage: vi.fn(),
    getTransactionById: vi.fn(),
    createTransaction: vi.fn(),
    updateTransaction: vi.fn(),
    deleteTransaction: vi.fn(),
    getMonthlyStats: vi.fn(),
    getRecurringList: vi.fn(),
    saveRecurring: vi.fn(),
    deleteRecurring: vi.fn(),
    applyRecurring: vi.fn(),
  }
})

describe('ledgerStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('loads accounts and categories', async () => {
    vi.mocked(ledgerApi.getAccountList).mockResolvedValue({
      list: [
        {
          id: '1',
          name: '现金',
          type: 'asset',
          subtype: 'cash',
          balance: '100',
          archived: false,
        },
      ],
    })
    vi.mocked(ledgerApi.getCategoryList).mockResolvedValue({
      list: [{ id: '10', parentId: '0', name: '餐饮', direction: 'expense', isSystem: true }],
    })

    const store = useLedgerStore()
    await store.fetchAccounts()
    await store.fetchCategories()

    expect(store.accounts).toHaveLength(1)
    expect(store.categories).toHaveLength(1)
    expect(store.expenseCategories).toHaveLength(1)
    expect(store.incomeCategories).toHaveLength(0)
  })

  it('saves account then refreshes list', async () => {
    vi.mocked(ledgerApi.saveAccount).mockResolvedValue({ id: '2', message: 'ok' })
    vi.mocked(ledgerApi.getAccountList).mockResolvedValue({ list: [] })

    const store = useLedgerStore()
    await store.saveAccount({ name: '现金', type: 'asset', subtype: 'cash', remark: '' }, false)

    expect(ledgerApi.saveAccount).toHaveBeenCalled()
    expect(ledgerApi.getAccountList).toHaveBeenCalled()
    expect(toastSuccess).toHaveBeenCalledWith('账户创建成功')
  })

  it('toasts backend message when account deletion is rejected', async () => {
    vi.mocked(ledgerApi.deleteAccount).mockRejectedValue(new Error('账户存在交易，无法删除'))

    const store = useLedgerStore()
    await expect(store.removeAccount('1')).rejects.toThrow('账户存在交易，无法删除')

    expect(toastError).toHaveBeenCalledWith('账户存在交易，无法删除')
  })

  it('fetches transactions with month range and filters', async () => {
    vi.mocked(ledgerApi.getTransactionPage).mockResolvedValue({
      page: '1',
      total: '21',
      list: [],
    })

    const store = useLedgerStore()
    store.filterMonth = '2026-08'
    store.filterAccountId = '1'
    store.filterType = 'expense'
    await store.fetchTransactions(2)

    expect(ledgerApi.getTransactionPage).toHaveBeenCalledWith({
      page: '2',
      pageSize: '10',
      accountId: '1',
      categoryId: undefined,
      type: 'expense',
      startTime: '2026-08-01 00:00:00',
      endTime: '2026-08-31 23:59:59',
    })
    expect(store.txTotal).toBe(21)
    expect(store.txTotalPages).toBe(3)
  })

  it('creates transaction then refreshes current page', async () => {
    vi.mocked(ledgerApi.createTransaction).mockResolvedValue({ id: '9', message: 'ok' })
    vi.mocked(ledgerApi.getTransactionPage).mockResolvedValue({ page: '1', total: '0', list: [] })

    const store = useLedgerStore()
    await store.createTransaction({
      type: 'transfer',
      bookedAt: '2026-08-22 12:00:00',
      remark: '',
      postings: [
        { accountId: '1', amount: '-100', sort: 0 },
        { accountId: '2', amount: '100', sort: 1 },
      ],
    })

    expect(ledgerApi.createTransaction).toHaveBeenCalled()
    expect(ledgerApi.getTransactionPage).toHaveBeenCalled()
    expect(toastSuccess).toHaveBeenCalledWith('记账成功')
  })

  it('loads monthly stats', async () => {
    vi.mocked(ledgerApi.getMonthlyStats).mockResolvedValue({
      month: '2026-08',
      totalExpense: '300',
      totalIncome: '10000',
      expenseByCategory: [{ categoryId: '10', categoryName: '餐饮', amount: '300' }],
      incomeByCategory: [{ categoryId: '20', categoryName: '工资', amount: '10000' }],
    })

    const store = useLedgerStore()
    await store.fetchMonthlyStats('2026-08')

    expect(ledgerApi.getMonthlyStats).toHaveBeenCalledWith('2026-08')
    expect(store.monthlyStats?.totalExpense).toBe('300')
  })

  it('loads recurring list', async () => {
    vi.mocked(ledgerApi.getRecurringList).mockResolvedValue({
      list: [
        {
          id: '1',
          accountId: '1',
          accountName: '招行储蓄卡',
          categoryId: '10',
          categoryName: '住房',
          type: 'expense',
          amount: '3500',
          remark: '房租',
          dayOfMonth: 5,
          startMonth: '2026-01',
          lastGeneratedMonth: '2026-08',
          enabled: true,
          nextDate: '2026-09-05',
        },
      ],
    })

    const store = useLedgerStore()
    await store.fetchRecurring()

    expect(store.recurringList).toHaveLength(1)
    expect(store.recurringList[0].nextDate).toBe('2026-09-05')
  })

  it('saves recurring then refreshes list', async () => {
    vi.mocked(ledgerApi.saveRecurring).mockResolvedValue({ id: '2', message: 'ok' })
    vi.mocked(ledgerApi.getRecurringList).mockResolvedValue({ list: [] })

    const store = useLedgerStore()
    await store.saveRecurring({
      accountId: '1',
      categoryId: '10',
      type: 'expense',
      amount: '3500',
      remark: '房租',
      dayOfMonth: 5,
      startMonth: '2026-01',
      enabled: true,
    })

    expect(ledgerApi.saveRecurring).toHaveBeenCalled()
    expect(ledgerApi.getRecurringList).toHaveBeenCalled()
    expect(toastSuccess).toHaveBeenCalledWith('周期账单创建成功')
  })

  it('toasts update message when saving recurring with id', async () => {
    vi.mocked(ledgerApi.saveRecurring).mockResolvedValue({ id: '2', message: 'ok' })
    vi.mocked(ledgerApi.getRecurringList).mockResolvedValue({ list: [] })

    const store = useLedgerStore()
    await store.saveRecurring({
      id: '2',
      accountId: '1',
      categoryId: '10',
      type: 'expense',
      amount: '3500',
      remark: '房租',
      dayOfMonth: 5,
      startMonth: '2026-01',
      enabled: false,
    })

    expect(toastSuccess).toHaveBeenCalledWith('周期账单更新成功')
  })

  it('removes recurring then refreshes list', async () => {
    vi.mocked(ledgerApi.deleteRecurring).mockResolvedValue(true)
    vi.mocked(ledgerApi.getRecurringList).mockResolvedValue({ list: [] })

    const store = useLedgerStore()
    await store.removeRecurring('2')

    expect(ledgerApi.deleteRecurring).toHaveBeenCalledWith('2')
    expect(toastSuccess).toHaveBeenCalledWith('周期账单删除成功')
  })

  it('applies recurring silently and returns the created count', async () => {
    vi.mocked(ledgerApi.applyRecurring).mockResolvedValue({ created: 3 })

    const store = useLedgerStore()
    const created = await store.applyRecurring()

    expect(created).toBe(3)
    expect(toastSuccess).not.toHaveBeenCalled()
    expect(toastError).not.toHaveBeenCalled()
  })

  it('returns 0 and only warns when apply fails', async () => {
    vi.mocked(ledgerApi.applyRecurring).mockRejectedValue(new Error('网络错误'))
    const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {})

    const store = useLedgerStore()
    const created = await store.applyRecurring()

    expect(created).toBe(0)
    expect(warnSpy).toHaveBeenCalled()
    expect(toastError).not.toHaveBeenCalled()
    warnSpy.mockRestore()
  })

  it('computes credit reminders from account billing fields and period transactions', async () => {
    // 固定今天 2026-08-22：账单日 5 → 本期 (07-05, 08-05]，还款日 08-25 还有 3 天
    vi.useFakeTimers({ now: new Date(2026, 7, 22), toFake: ['Date'] })
    try {
      const store = useLedgerStore()
      store.accounts = [
        {
          id: '1',
          name: '招行信用卡',
          type: 'liability',
          subtype: 'credit_card',
          billingDay: 5,
          paymentDueDay: 25,
          archived: false,
        },
        { id: '2', name: '现金', type: 'asset', subtype: 'cash', archived: false },
      ]
      vi.mocked(ledgerApi.getTransactionPage).mockResolvedValue({
        page: '1',
        total: '2',
        list: [
          {
            id: '100',
            type: 'expense',
            bookedAt: '2026-08-01 12:00:00',
            remark: '',
            postings: [
              { accountId: '1', amount: '-100', sort: 0 },
              { accountId: '999', amount: '100', categoryId: '10', sort: 1 },
            ],
          },
          {
            id: '101',
            type: 'expense',
            bookedAt: '2026-07-20 12:00:00',
            remark: '',
            postings: [
              { accountId: '1', amount: '-50.5', sort: 0 },
              { accountId: '999', amount: '50.5', categoryId: '10', sort: 1 },
            ],
          },
        ],
      })

      await store.fetchCreditReminders()

      // 现金账户无账单字段，不参与；交易按 accountId + 账单区间查询
      expect(ledgerApi.getTransactionPage).toHaveBeenCalledTimes(1)
      expect(ledgerApi.getTransactionPage).toHaveBeenCalledWith({
        page: '1',
        pageSize: '500',
        accountId: '1',
        startTime: '2026-07-05 00:00:00',
        endTime: '2026-08-05 23:59:59',
      })
      expect(store.creditReminders).toHaveLength(1)
      expect(store.creditReminders[0]).toMatchObject({
        accountId: '1',
        accountName: '招行信用卡',
        dueDate: '2026-08-25',
        daysUntilDue: 3,
        level: 'warning',
      })
      expect(store.creditReminders[0].amountDue).toBeCloseTo(150.5)
    } finally {
      vi.useRealTimers()
    }
  })

  it('clears credit reminders when no account has billing fields', async () => {
    const store = useLedgerStore()
    store.accounts = [{ id: '2', name: '现金', type: 'asset', subtype: 'cash', archived: false }]

    await store.fetchCreditReminders()

    expect(store.creditReminders).toEqual([])
    expect(ledgerApi.getTransactionPage).not.toHaveBeenCalled()
  })

  it('keeps existing reminders and only warns when reminder loading fails', async () => {
    const store = useLedgerStore()
    store.accounts = [
      {
        id: '1',
        name: '招行信用卡',
        type: 'liability',
        subtype: 'credit_card',
        billingDay: 5,
        paymentDueDay: 25,
        archived: false,
      },
    ]
    vi.mocked(ledgerApi.getTransactionPage).mockRejectedValue(new Error('网络错误'))
    const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {})

    await store.fetchCreditReminders()

    expect(warnSpy).toHaveBeenCalled()
    expect(toastError).not.toHaveBeenCalled()
    warnSpy.mockRestore()
  })
})
