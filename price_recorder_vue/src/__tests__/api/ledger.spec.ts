import { beforeEach, describe, expect, it, vi } from 'vitest'

const getMock = vi.fn()
const postMock = vi.fn()

vi.mock('@/utils/request.ts', () => ({
  default: {
    get: getMock,
    post: postMock,
  },
}))

describe('ledger api', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    postMock.mockResolvedValue({})
  })

  /* ---------- 账户 ---------- */

  it('serializes account fields before posting save', async () => {
    const { saveAccount } = await import('@/api/ledger')

    await saveAccount({
      name: '招行信用卡',
      type: 'liability',
      subtype: 'credit_card',
      creditLimit: 50000,
      billingDay: '5',
      paymentDueDay: '25',
      remark: '',
      openingBalance: -1200.5,
    })

    expect(postMock).toHaveBeenCalledWith('/ledger/account/save/v1', {
      name: '招行信用卡',
      type: 'liability',
      subtype: 'credit_card',
      creditLimit: '50000',
      billingDay: 5,
      paymentDueDay: 25,
      remark: '',
      openingBalance: '-1200.5',
    })
  })

  it('omits empty optional account fields', async () => {
    const { saveAccount } = await import('@/api/ledger')

    await saveAccount({ name: '现金', type: 'asset', subtype: 'cash', remark: '' })

    expect(postMock).toHaveBeenCalledWith('/ledger/account/save/v1', {
      name: '现金',
      type: 'asset',
      subtype: 'cash',
      remark: '',
    })
  })

  it('posts account update with numeric id', async () => {
    const { updateAccount } = await import('@/api/ledger')

    await updateAccount({
      id: '3',
      name: '现金',
      type: 'asset',
      subtype: 'cash',
      remark: '',
      archived: true,
    })

    expect(postMock).toHaveBeenCalledWith('/ledger/account/update/v1', {
      id: 3,
      name: '现金',
      type: 'asset',
      subtype: 'cash',
      remark: '',
      archived: true,
    })
  })

  it('deletes account with numeric id and returns flag', async () => {
    postMock.mockResolvedValue({ flag: true })
    const { deleteAccount } = await import('@/api/ledger')

    const result = await deleteAccount('3')

    expect(postMock).toHaveBeenCalledWith('/ledger/account/delete/v1', { id: 3 })
    expect(result).toBe(true)
  })

  it('fetches account list', async () => {
    const { getAccountList } = await import('@/api/ledger')

    await getAccountList()

    expect(getMock).toHaveBeenCalledWith('/ledger/account/list/v1')
  })

  /* ---------- 分类 ---------- */

  it('posts category save with numeric parentId; top level defaults to 0', async () => {
    const { saveCategory } = await import('@/api/ledger')

    await saveCategory({ name: '餐饮', direction: 'expense' })
    expect(postMock).toHaveBeenCalledWith('/ledger/category/save/v1', {
      name: '餐饮',
      direction: 'expense',
      parentId: 0,
    })

    await saveCategory({ parentId: '10', name: '午餐', direction: 'expense' })
    expect(postMock).toHaveBeenCalledWith('/ledger/category/save/v1', {
      name: '午餐',
      direction: 'expense',
      parentId: 10,
    })
  })

  it('posts category update with numeric id and parentId', async () => {
    const { updateCategory } = await import('@/api/ledger')

    await updateCategory({ id: '11', parentId: '10', name: '午餐', direction: 'expense' })

    expect(postMock).toHaveBeenCalledWith('/ledger/category/update/v1', {
      id: 11,
      name: '午餐',
      direction: 'expense',
      parentId: 10,
    })
  })

  it('deletes category with numeric id', async () => {
    postMock.mockResolvedValue({ flag: true })
    const { deleteCategory } = await import('@/api/ledger')

    await deleteCategory('11')

    expect(postMock).toHaveBeenCalledWith('/ledger/category/delete/v1', { id: 11 })
  })

  /* ---------- 交易 ---------- */

  it('serializes transaction postings with numeric ids and string amounts', async () => {
    const { createTransaction } = await import('@/api/ledger')

    await createTransaction({
      type: 'expense',
      bookedAt: '2026-08-22 12:00:00',
      remark: '午饭',
      postings: [
        { accountId: '1', amount: '-100', sort: 0 },
        { accountId: 0, amount: '60', categoryId: '10', sort: 1 },
        { accountId: 0, amount: '40', categoryId: '11', sort: 2 },
      ],
    })

    expect(postMock).toHaveBeenCalledWith('/ledger/transaction/save/v1', {
      type: 'expense',
      bookedAt: '2026-08-22 12:00:00',
      remark: '午饭',
      postings: [
        { accountId: 1, amount: '-100', sort: 0 },
        { accountId: 0, amount: '60', categoryId: 10, sort: 1 },
        { accountId: 0, amount: '40', categoryId: 11, sort: 2 },
      ],
    })
  })

  it('omits categoryId from postings without one', async () => {
    const { createTransaction } = await import('@/api/ledger')

    await createTransaction({
      type: 'transfer',
      bookedAt: '2026-08-22 12:00:00',
      remark: '',
      postings: [
        { accountId: '1', amount: '-500', sort: 0 },
        { accountId: '2', amount: '500', sort: 1 },
      ],
    })

    expect(postMock).toHaveBeenCalledWith('/ledger/transaction/save/v1', {
      type: 'transfer',
      bookedAt: '2026-08-22 12:00:00',
      remark: '',
      postings: [
        { accountId: 1, amount: '-500', sort: 0 },
        { accountId: 2, amount: '500', sort: 1 },
      ],
    })
  })

  it('posts transaction update with numeric id', async () => {
    const { updateTransaction } = await import('@/api/ledger')

    await updateTransaction({
      id: '7',
      type: 'transfer',
      bookedAt: '2026-08-22 12:00:00',
      remark: '',
      postings: [
        { accountId: '1', amount: '-500', sort: 0 },
        { accountId: '2', amount: '500', sort: 1 },
      ],
    })

    expect(postMock).toHaveBeenCalledWith('/ledger/transaction/update/v1', {
      id: 7,
      type: 'transfer',
      bookedAt: '2026-08-22 12:00:00',
      remark: '',
      postings: [
        { accountId: 1, amount: '-500', sort: 0 },
        { accountId: 2, amount: '500', sort: 1 },
      ],
    })
  })

  it('deletes transaction with numeric id', async () => {
    postMock.mockResolvedValue({ flag: true })
    const { deleteTransaction } = await import('@/api/ledger')

    await deleteTransaction('7')

    expect(postMock).toHaveBeenCalledWith('/ledger/transaction/delete/v1', { id: 7 })
  })

  it('builds transaction page params with filters and time range', async () => {
    const { getTransactionPage } = await import('@/api/ledger')

    await getTransactionPage({
      page: '2',
      pageSize: '10',
      accountId: '1',
      categoryId: '10',
      type: 'expense',
      startTime: '2026-08-01 00:00:00',
      endTime: '2026-08-31 23:59:59',
    })

    expect(getMock).toHaveBeenCalledWith('/ledger/transaction/page/v1', {
      params: {
        page: '2',
        pageSize: '10',
        accountId: '1',
        categoryId: '10',
        type: 'expense',
        startTime: '2026-08-01 00:00:00',
        endTime: '2026-08-31 23:59:59',
      },
    })
  })

  it('omits empty filters from transaction page params', async () => {
    const { getTransactionPage } = await import('@/api/ledger')

    await getTransactionPage({ page: '1', pageSize: '10' })

    expect(getMock).toHaveBeenCalledWith('/ledger/transaction/page/v1', {
      params: { page: '1', pageSize: '10' },
    })
  })

  it('fetches transaction detail by id', async () => {
    const { getTransactionById } = await import('@/api/ledger')

    await getTransactionById('7')

    expect(getMock).toHaveBeenCalledWith('/ledger/transaction/get/v1', { params: { id: '7' } })
  })

  it('fetches monthly stats by month', async () => {
    const { getMonthlyStats } = await import('@/api/ledger')

    await getMonthlyStats('2026-08')

    expect(getMock).toHaveBeenCalledWith('/ledger/stats/monthly/v1', {
      params: { month: '2026-08' },
    })
  })

  /* ---------- 预算 ---------- */

  it('posts budget save with numeric categoryId and string amount', async () => {
    const { saveBudget } = await import('@/api/ledger')

    await saveBudget({ categoryId: '10', month: '2026-08', amount: 500 })

    expect(postMock).toHaveBeenCalledWith('/ledger/budget/save/v1', {
      categoryId: 10,
      month: '2026-08',
      amount: '500',
    })
  })

  it('deletes budget with numeric id and returns flag', async () => {
    postMock.mockResolvedValue({ flag: true })
    const { deleteBudget } = await import('@/api/ledger')

    const result = await deleteBudget('3')

    expect(postMock).toHaveBeenCalledWith('/ledger/budget/delete/v1', { id: 3 })
    expect(result).toBe(true)
  })

  it('fetches budget list by month', async () => {
    const { getBudgetList } = await import('@/api/ledger')

    await getBudgetList('2026-08')

    expect(getMock).toHaveBeenCalledWith('/ledger/budget/list/v1', {
      params: { month: '2026-08' },
    })
  })

  /* ---------- 余额走势 ---------- */

  it('fetches balance trend with account and date range', async () => {
    const { getBalanceTrend } = await import('@/api/ledger')

    await getBalanceTrend({
      accountId: '1',
      startTime: '2026-03-01',
      endTime: '2026-08-22',
    })

    expect(getMock).toHaveBeenCalledWith('/ledger/stats/balance-trend/v1', {
      params: { accountId: '1', startTime: '2026-03-01', endTime: '2026-08-22' },
    })
  })

  it('omits accountId from balance trend params when empty (net asset)', async () => {
    const { getBalanceTrend } = await import('@/api/ledger')

    await getBalanceTrend({ accountId: '', startTime: '2026-03-01', endTime: '2026-08-22' })

    expect(getMock).toHaveBeenCalledWith('/ledger/stats/balance-trend/v1', {
      params: { startTime: '2026-03-01', endTime: '2026-08-22' },
    })
  })

  /* ---------- 周期账单 ---------- */

  it('posts recurring save with numeric ids, numeric dayOfMonth and string amount', async () => {
    const { saveRecurring } = await import('@/api/ledger')

    await saveRecurring({
      accountId: '1',
      categoryId: '10',
      type: 'expense',
      amount: 3500,
      remark: '房租',
      dayOfMonth: '5',
      startMonth: '2026-08',
      enabled: true,
    })

    expect(postMock).toHaveBeenCalledWith('/ledger/recurring/save/v1', {
      accountId: 1,
      categoryId: 10,
      type: 'expense',
      amount: '3500',
      remark: '房租',
      dayOfMonth: 5,
      startMonth: '2026-08',
      enabled: true,
    })
  })

  it('includes numeric id when updating an existing recurring', async () => {
    const { saveRecurring } = await import('@/api/ledger')

    await saveRecurring({
      id: '7',
      accountId: '2',
      categoryId: '20',
      type: 'income',
      amount: '12000',
      remark: '工资',
      dayOfMonth: 10,
      startMonth: '2026-01',
      enabled: false,
    })

    expect(postMock).toHaveBeenCalledWith('/ledger/recurring/save/v1', {
      id: 7,
      accountId: 2,
      categoryId: 20,
      type: 'income',
      amount: '12000',
      remark: '工资',
      dayOfMonth: 10,
      startMonth: '2026-01',
      enabled: false,
    })
  })

  it('deletes recurring with numeric id and returns flag', async () => {
    postMock.mockResolvedValue({ flag: true })
    const { deleteRecurring } = await import('@/api/ledger')

    const result = await deleteRecurring('7')

    expect(postMock).toHaveBeenCalledWith('/ledger/recurring/delete/v1', { id: 7 })
    expect(result).toBe(true)
  })

  it('fetches recurring list', async () => {
    const { getRecurringList } = await import('@/api/ledger')

    await getRecurringList()

    expect(getMock).toHaveBeenCalledWith('/ledger/recurring/list/v1')
  })

  it('applies recurring with an empty body', async () => {
    postMock.mockResolvedValue({ created: 2 })
    const { applyRecurring } = await import('@/api/ledger')

    const result = await applyRecurring()

    expect(postMock).toHaveBeenCalledWith('/ledger/recurring/apply/v1', {})
    expect(result.created).toBe(2)
  })

  /* ---------- postings 构造 ---------- */

  it('builds expense postings with split legs on the system account (accountId=0)', async () => {
    const { buildExpensePostings } = await import('@/api/ledger')

    const postings = buildExpensePostings('1', '100', [
      { amount: '60', categoryId: '10' },
      { amount: 40, categoryId: '11' },
    ])

    expect(postings).toEqual([
      { accountId: '1', amount: '-100', sort: 0 },
      { accountId: 0, amount: '60', categoryId: '10', sort: 1 },
      { accountId: 0, amount: '40', categoryId: '11', sort: 2 },
    ])
    // 自传腿合计为 0
    const sum = postings.reduce((acc, p) => acc + Number(p.amount), 0)
    expect(sum).toBe(0)
  })

  it('builds single-leg expense postings when not split', async () => {
    const { buildExpensePostings } = await import('@/api/ledger')

    const postings = buildExpensePostings('1', '88.5', [{ amount: '88.5', categoryId: '10' }])

    expect(postings).toEqual([
      { accountId: '1', amount: '-88.5', sort: 0 },
      { accountId: 0, amount: '88.5', categoryId: '10', sort: 1 },
    ])
  })

  it('builds income postings with negative system legs', async () => {
    const { buildIncomePostings } = await import('@/api/ledger')

    const postings = buildIncomePostings('2', '1000', [
      { amount: '700', categoryId: '20' },
      { amount: '300', categoryId: '21' },
    ])

    expect(postings).toEqual([
      { accountId: '2', amount: '1000', sort: 0 },
      { accountId: 0, amount: '-700', categoryId: '20', sort: 1 },
      { accountId: 0, amount: '-300', categoryId: '21', sort: 2 },
    ])
    const sum = postings.reduce((acc, p) => acc + Number(p.amount), 0)
    expect(sum).toBe(0)
  })

  it('builds transfer postings as two user-account legs', async () => {
    const { buildTransferPostings } = await import('@/api/ledger')

    const postings = buildTransferPostings('1', '2', '500')

    expect(postings).toEqual([
      { accountId: '1', amount: '-500', sort: 0 },
      { accountId: '2', amount: '500', sort: 1 },
    ])
  })

  it('strips redundant negative sign when building postings', async () => {
    const { buildExpensePostings } = await import('@/api/ledger')

    const postings = buildExpensePostings('1', '-100', [{ amount: '-100', categoryId: '10' }])

    expect(postings[0].amount).toBe('-100')
    expect(postings[1].amount).toBe('100')
  })

  /* ---------- 月份工具 ---------- */

  it('computes month range covering the whole month', async () => {
    const { monthRange } = await import('@/api/ledger')

    expect(monthRange('2026-08')).toEqual({
      startTime: '2026-08-01 00:00:00',
      endTime: '2026-08-31 23:59:59',
    })
    expect(monthRange('2026-02')).toEqual({
      startTime: '2026-02-01 00:00:00',
      endTime: '2026-02-28 23:59:59',
    })
  })

  it('shifts month across year boundary', async () => {
    const { shiftMonth } = await import('@/api/ledger')

    expect(shiftMonth('2026-01', -1)).toBe('2025-12')
    expect(shiftMonth('2025-12', 1)).toBe('2026-01')
  })
})
