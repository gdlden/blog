import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises, RouterLinkStub } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import * as ledgerApi from '@/api/ledger'
import { currentMonth } from '@/api/ledger'

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => ({ name: 'ledgerReport' }),
    useRouter: () => ({ push: vi.fn() }),
  }
})

vi.mock('vue-toastification', () => ({
  useToast: () => ({
    success: vi.fn(),
    error: vi.fn(),
  }),
}))

function mockStats() {
  return vi.spyOn(ledgerApi, 'getMonthlyStats').mockResolvedValue({
    month: '2026-08',
    totalExpense: '300',
    totalIncome: '10000',
    expenseByCategory: [
      { categoryId: '12', categoryName: '交通', amount: '100' },
      { categoryId: '10', categoryName: '餐饮', amount: '200' },
    ],
    incomeByCategory: [{ categoryId: '20', categoryName: '工资', amount: '10000' }],
  })
}

// 组件挂载时会拉取账户/分类/预算/走势，统一打桩避免真实请求
function mockP2Apis() {
  const budgetListSpy = vi.spyOn(ledgerApi, 'getBudgetList').mockResolvedValue({
    list: [
      { id: '1', categoryId: '10', categoryName: '餐饮', amount: '500', used: '600' },
      { id: '2', categoryId: '12', categoryName: '交通', amount: '200', used: '100' },
    ],
  })
  const trendSpy = vi.spyOn(ledgerApi, 'getBalanceTrend').mockResolvedValue({
    points: [
      { date: '2026-08-01', balance: '100' },
      { date: '2026-08-02', balance: '300' },
      { date: '2026-08-03', balance: '200' },
    ],
  })
  vi.spyOn(ledgerApi, 'getAccountList').mockResolvedValue({
    list: [
      { id: '1', name: '现金', type: 'asset', subtype: 'cash', balance: '100', archived: false },
    ],
  })
  vi.spyOn(ledgerApi, 'getCategoryList').mockResolvedValue({
    list: [
      { id: '10', parentId: '0', name: '餐饮', direction: 'expense', isSystem: true },
      { id: '11', parentId: '10', name: '午餐', direction: 'expense' },
      { id: '20', parentId: '0', name: '工资', direction: 'income', isSystem: true },
    ],
  })
  return { budgetListSpy, trendSpy }
}

async function mountReport() {
  const statsSpy = mockStats()
  const p2 = mockP2Apis()
  const { default: LedgerReport } = await import('@/view/LedgerReport.vue')
  const wrapper = mount(LedgerReport, {
    global: { stubs: { RouterLink: RouterLinkStub } },
  })
  await flushPromises()
  return { wrapper, statsSpy, ...p2 }
}

describe('LedgerReport.vue', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.restoreAllMocks()
  })

  it('renders monthly totals and sorted category breakdown with percents', async () => {
    const { wrapper } = await mountReport()

    expect(wrapper.find('.total-expense').text()).toBe('¥300.00')
    expect(wrapper.find('.total-income').text()).toBe('¥10,000.00')

    // 排行按 amount 降序：餐饮(200) 在交通(100) 之前，尽管接口返回顺序相反
    const rows = wrapper.findAll('.expense-rank .category-row')
    expect(rows).toHaveLength(2)
    expect(rows[0].text()).toContain('餐饮')
    expect(rows[0].text()).toContain('¥200.00')
    expect(rows[0].text()).toContain('66.7%')
    expect(rows[1].text()).toContain('交通')
    expect(rows[1].text()).toContain('33.3%')

    // 条形宽度相对最大值：第一名 100%，第二名 50%
    const bars = wrapper.findAll('.expense-rank .category-bar')
    expect(bars[0].attributes('style')).toContain('width: 100%')
    expect(bars[1].attributes('style')).toContain('width: 50%')

    const incomeRows = wrapper.findAll('.income-rank .category-row')
    expect(incomeRows).toHaveLength(1)
    expect(incomeRows[0].text()).toContain('工资')
    expect(incomeRows[0].text()).toContain('100.0%')
  })

  it('fetches stats for the current month on mount', async () => {
    const { statsSpy } = await mountReport()

    expect(statsSpy).toHaveBeenCalledWith(currentMonth())
  })

  it('refetches stats when the month changes', async () => {
    const { wrapper, statsSpy } = await mountReport()

    await wrapper.find('.report-month').setValue('2026-07')
    await flushPromises()

    expect(statsSpy).toHaveBeenCalledWith('2026-07')
  })

  it('shows empty hints when no data for the month', async () => {
    vi.spyOn(ledgerApi, 'getMonthlyStats').mockResolvedValue({
      month: '2026-08',
      totalExpense: '0',
      totalIncome: '0',
      expenseByCategory: [],
      incomeByCategory: [],
    })
    mockP2Apis()
    const { default: LedgerReport } = await import('@/view/LedgerReport.vue')
    const wrapper = mount(LedgerReport, {
      global: { stubs: { RouterLink: RouterLinkStub } },
    })
    await flushPromises()

    expect(wrapper.text()).toContain('本月暂无支出')
    expect(wrapper.text()).toContain('本月暂无收入')
  })

  it('renders budget progress bars, red when over budget', async () => {
    const { wrapper, budgetListSpy } = await mountReport()

    expect(budgetListSpy).toHaveBeenCalledWith(currentMonth())

    const rows = wrapper.findAll('.budget-row')
    expect(rows).toHaveLength(2)
    expect(rows[0].text()).toContain('餐饮')
    expect(rows[0].text()).toContain('已用 ¥600.00 / 预算 ¥500.00')
    expect(rows[1].text()).toContain('交通')

    const bars = wrapper.findAll('.budget-bar')
    // 餐饮 600/500 超支：红色、条宽封顶 100%；交通 100/200：主色、50%
    expect(bars[0].classes()).toContain('bg-[#ff3b30]')
    expect(bars[0].attributes('style')).toContain('width: 100%')
    expect(bars[1].classes()).toContain('bg-[#0071e3]')
    expect(bars[1].attributes('style')).toContain('width: 50%')
  })

  it('refetches budgets when the month changes', async () => {
    const { wrapper, budgetListSpy } = await mountReport()

    await wrapper.find('.report-month').setValue('2026-07')
    await flushPromises()

    expect(budgetListSpy).toHaveBeenCalledWith('2026-07')
  })

  it('draws the trend polyline with one coordinate per mock point', async () => {
    const { wrapper, trendSpy } = await mountReport()

    // 默认净资产：accountId 省略
    expect(trendSpy).toHaveBeenCalledWith(
      expect.objectContaining({ accountId: undefined }),
    )

    // 640x220 画布、padding 8/16：三点 100/300/200 → x 均分、y 按 min/max 归一化
    const points = wrapper.find('.trend-line').attributes('points')
    expect(points).toBe('8,204 320,16 632,110')
    expect(wrapper.find('.trend-area').attributes('points')).toBe(
      '8,204 8,204 320,16 632,110 632,204',
    )

    expect(wrapper.find('.trend-max').text()).toBe('最高 ¥300.00')
    expect(wrapper.find('.trend-min').text()).toBe('最低 ¥100.00')
    expect(wrapper.find('.trend-start').text()).toBe('2026-08-01')
    expect(wrapper.find('.trend-end').text()).toBe('2026-08-03')
  })

  it('submits budget modal with two-level expense category and default month', async () => {
    const saveBudgetSpy = vi.spyOn(ledgerApi, 'saveBudget').mockResolvedValue({
      id: '3',
      message: 'ok',
    })
    const { wrapper } = await mountReport()

    await wrapper.find('.budget-add').trigger('click')
    await flushPromises()

    const modal = wrapper.find('.budget-modal')
    expect(modal.exists()).toBe(true)

    // 两级联动：选一级"餐饮"后出现子分类，再选"午餐"
    await modal.find('.budget-category').setValue('10')
    await flushPromises()
    await modal.find('.budget-sub-category').setValue('11')
    await modal.find('.budget-amount').setValue('500')

    // 月份默认报表当前月
    expect((modal.find('.budget-month').element as HTMLInputElement).value).toBe(currentMonth())

    await modal.find('.budget-submit').trigger('click')
    await flushPromises()

    expect(saveBudgetSpy).toHaveBeenCalledWith({
      categoryId: '11',
      month: currentMonth(),
      amount: 500,
    })
  })

  it('hides the reminder banner when no account has a pending repayment', async () => {
    const { wrapper } = await mountReport()

    expect(wrapper.find('.credit-reminder-banner').exists()).toBe(false)
  })

  it('renders the reminder banner with red rows for accounts due within 7 days', async () => {
    // 固定今天 2026-08-22：账单日 5 → 本期 (07-05, 08-05]，还款日 08-25 还有 3 天
    vi.useFakeTimers({ now: new Date(2026, 7, 22), toFake: ['Date'] })
    try {
      mockStats()
      vi.spyOn(ledgerApi, 'getBudgetList').mockResolvedValue({ list: [] })
      vi.spyOn(ledgerApi, 'getBalanceTrend').mockResolvedValue({ points: [] })
      vi.spyOn(ledgerApi, 'getAccountList').mockResolvedValue({
        list: [
          {
            id: '1',
            name: '招行信用卡',
            type: 'liability',
            subtype: 'credit_card',
            billingDay: 5,
            paymentDueDay: 25,
            balance: '-150.5',
            archived: false,
          },
        ],
      })
      vi.spyOn(ledgerApi, 'getCategoryList').mockResolvedValue({ list: [] })
      vi.spyOn(ledgerApi, 'getTransactionPage').mockResolvedValue({
        page: '1',
        total: '1',
        list: [
          {
            id: '100',
            type: 'expense',
            bookedAt: '2026-08-01 12:00:00',
            remark: '',
            postings: [
              { accountId: '1', amount: '-150.5', sort: 0 },
              { accountId: '999', amount: '150.5', categoryId: '10', sort: 1 },
            ],
          },
        ],
      })

      const { default: LedgerReport } = await import('@/view/LedgerReport.vue')
      const wrapper = mount(LedgerReport, {
        global: { stubs: { RouterLink: RouterLinkStub } },
      })
      await flushPromises()

      const banner = wrapper.find('.credit-reminder-banner')
      expect(banner.exists()).toBe(true)

      const rows = banner.findAll('.credit-reminder-row')
      expect(rows).toHaveLength(1)
      expect(rows[0].text()).toContain('招行信用卡')
      expect(rows[0].text()).toContain('本期应还 ¥150.50')
      expect(rows[0].text()).toContain('还有 3 天还款')
      expect(rows[0].text()).toContain('2026-08-25')
      // <=7 天：红色边框与文字
      expect(rows[0].classes().join(' ')).toContain('border-[#ff3b30]')
      expect(rows[0].find('.text-\\[\\#ff3b30\\]').exists()).toBe(true)
    } finally {
      vi.useRealTimers()
    }
  })
})
