import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises, RouterLinkStub } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import * as ledgerApi from '@/api/ledger'
import type { LedgerTransaction } from '@/api/ledger'

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => ({ name: 'ledger' }),
    useRouter: () => ({ push: vi.fn() }),
  }
})

vi.mock('vue-toastification', () => ({
  useToast: () => ({
    success: vi.fn(),
    error: vi.fn(),
  }),
}))

function mockLedger(transactions: Array<Partial<LedgerTransaction>> = [], applyCreated = 0) {
  vi.spyOn(ledgerApi, 'getAccountList').mockResolvedValue({
    list: [
      { id: '1', name: '现金', type: 'asset', subtype: 'cash', balance: '100', archived: false },
      {
        id: '2',
        name: '招行储蓄卡',
        type: 'asset',
        subtype: 'debit_card',
        balance: '2000',
        archived: false,
      },
    ],
  })
  vi.spyOn(ledgerApi, 'getCategoryList').mockResolvedValue({
    list: [
      { id: '10', parentId: '0', name: '餐饮', direction: 'expense', isSystem: false },
      { id: '11', parentId: '10', name: '午餐', direction: 'expense', isSystem: false },
      { id: '12', parentId: '0', name: '交通', direction: 'expense', isSystem: false },
      { id: '20', parentId: '0', name: '工资', direction: 'income', isSystem: true },
    ],
  })
  vi.spyOn(ledgerApi, 'getTransactionPage').mockResolvedValue({
    page: '1',
    total: String(transactions.length),
    list: transactions as LedgerTransaction[],
  })
  vi.spyOn(ledgerApi, 'applyRecurring').mockResolvedValue({ created: applyCreated })
}

async function mountList(transactions: Array<Partial<LedgerTransaction>> = [], applyCreated = 0) {
  mockLedger(transactions, applyCreated)
  const { default: LedgerList } = await import('@/view/LedgerList.vue')
  const wrapper = mount(LedgerList, {
    global: { stubs: { RouterLink: RouterLinkStub } },
  })
  await flushPromises()
  return wrapper
}

async function openCreateModal(wrapper: ReturnType<typeof mount>) {
  await wrapper
    .findAll('button')
    .find((button) => button.text().includes('记一笔'))!
    .trigger('click')
  await flushPromises()
}

async function submitModal(wrapper: ReturnType<typeof mount>) {
  await wrapper
    .findAll('button')
    .find((button) => button.text() === '保存')!
    .trigger('click')
  await flushPromises()
}

describe('LedgerList.vue', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.restoreAllMocks()
  })

  it('builds expense postings with split legs (accountId=0 system legs)', async () => {
    const createSpy = vi
      .spyOn(ledgerApi, 'createTransaction')
      .mockResolvedValue({ id: '1', message: 'ok' })
    const wrapper = await mountList()

    await openCreateModal(wrapper)
    // 默认支出 Tab：付款账户 + 总金额
    await wrapper.find('.ledger-tx-modal select.tx-account').setValue('1')
    await wrapper.find('.ledger-tx-modal input.tx-total').setValue('100')
    // 开启拆分，填两条腿
    await wrapper
      .findAll('button')
      .find((button) => button.text() === '拆分')!
      .trigger('click')
    await flushPromises()

    const legs = wrapper.findAll('.tx-leg')
    expect(legs).toHaveLength(2)
    await legs[0].find('.tx-leg-amount').setValue('60')
    await legs[0].find('.tx-leg-category').setValue('10')
    await legs[1].find('.tx-leg-amount').setValue('40')
    await legs[1].find('.tx-leg-category').setValue('10')
    await flushPromises()
    // 第二腿选二级分类
    await legs[1].find('.tx-leg-sub-category').setValue('11')

    await submitModal(wrapper)

    expect(createSpy).toHaveBeenCalledWith({
      type: 'expense',
      bookedAt: expect.stringMatching(/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:00$/),
      remark: '',
      postings: [
        { accountId: '1', amount: '-100', sort: 0 },
        { accountId: 0, amount: '60', categoryId: '10', sort: 1 },
        { accountId: 0, amount: '40', categoryId: '11', sort: 2 },
      ],
    })
  })

  it('builds single-leg expense postings when not split', async () => {
    const createSpy = vi
      .spyOn(ledgerApi, 'createTransaction')
      .mockResolvedValue({ id: '1', message: 'ok' })
    const wrapper = await mountList()

    await openCreateModal(wrapper)
    await wrapper.find('.ledger-tx-modal select.tx-account').setValue('2')
    await wrapper.find('.ledger-tx-modal input.tx-total').setValue('88.5')
    await wrapper.find('.ledger-tx-modal select.tx-category').setValue('12')
    await wrapper.find('.ledger-tx-modal input.tx-remark').setValue('打车')

    await submitModal(wrapper)

    expect(createSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        type: 'expense',
        remark: '打车',
        postings: [
          { accountId: '2', amount: '-88.5', sort: 0 },
          { accountId: 0, amount: '88.5', categoryId: '12', sort: 1 },
        ],
      }),
    )
  })

  it('builds income postings with negative system legs', async () => {
    const createSpy = vi
      .spyOn(ledgerApi, 'createTransaction')
      .mockResolvedValue({ id: '1', message: 'ok' })
    const wrapper = await mountList()

    await openCreateModal(wrapper)
    await wrapper
      .findAll('button')
      .find((button) => button.text() === '收入')!
      .trigger('click')
    await flushPromises()

    await wrapper.find('.ledger-tx-modal select.tx-account').setValue('2')
    await wrapper.find('.ledger-tx-modal input.tx-total').setValue('1000')
    await wrapper.find('.ledger-tx-modal select.tx-category').setValue('20')

    await submitModal(wrapper)

    expect(createSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        type: 'income',
        postings: [
          { accountId: '2', amount: '1000', sort: 0 },
          { accountId: 0, amount: '-1000', categoryId: '20', sort: 1 },
        ],
      }),
    )
  })

  it('builds transfer postings as two user-account legs', async () => {
    const createSpy = vi
      .spyOn(ledgerApi, 'createTransaction')
      .mockResolvedValue({ id: '1', message: 'ok' })
    const wrapper = await mountList()

    await openCreateModal(wrapper)
    await wrapper
      .findAll('button')
      .find((button) => button.text() === '转账')!
      .trigger('click')
    await flushPromises()

    await wrapper.find('.ledger-tx-modal select.tx-from-account').setValue('1')
    await wrapper.find('.ledger-tx-modal select.tx-to-account').setValue('2')
    await wrapper.find('.ledger-tx-modal input.tx-total').setValue('500')

    await submitModal(wrapper)

    expect(createSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        type: 'transfer',
        postings: [
          { accountId: '1', amount: '-500', sort: 0 },
          { accountId: '2', amount: '500', sort: 1 },
        ],
      }),
    )
  })

  it('blocks submit when split legs do not sum to the total', async () => {
    const createSpy = vi.spyOn(ledgerApi, 'createTransaction')
    const wrapper = await mountList()

    await openCreateModal(wrapper)
    await wrapper.find('.ledger-tx-modal select.tx-account').setValue('1')
    await wrapper.find('.ledger-tx-modal input.tx-total').setValue('100')
    await wrapper
      .findAll('button')
      .find((button) => button.text() === '拆分')!
      .trigger('click')
    await flushPromises()

    const legs = wrapper.findAll('.tx-leg')
    await legs[0].find('.tx-leg-amount').setValue('60')
    await legs[0].find('.tx-leg-category').setValue('10')
    await legs[1].find('.tx-leg-amount').setValue('30')
    await legs[1].find('.tx-leg-category').setValue('12')
    await flushPromises()

    expect(wrapper.text()).toContain('差额')
    const saveButton = wrapper.findAll('button').find((button) => button.text() === '保存')!
    expect(saveButton.attributes('disabled')).toBeDefined()

    await saveButton.trigger('click')
    await flushPromises()
    expect(createSpy).not.toHaveBeenCalled()
  })

  it('renders transaction rows with colored amounts, category and account names', async () => {
    const wrapper = await mountList([
      {
        id: '100',
        type: 'expense',
        bookedAt: '2026-08-20 12:30:00',
        remark: '聚餐',
        postings: [
          { accountId: '1', amount: '-100', sort: 0 },
          { accountId: '999', amount: '100', categoryId: '10', sort: 1 },
        ],
      },
      {
        id: '101',
        type: 'income',
        bookedAt: '2026-08-01 09:00:00',
        remark: '',
        postings: [
          { accountId: '2', amount: '8000', sort: 0 },
          { accountId: '999', amount: '-8000', categoryId: '20', sort: 1 },
        ],
      },
      {
        id: '102',
        type: 'transfer',
        bookedAt: '2026-08-02 10:00:00',
        remark: '',
        postings: [
          { accountId: '1', amount: '-500', sort: 0 },
          { accountId: '2', amount: '500', sort: 1 },
        ],
      },
    ])

    const items = wrapper.findAll('.ledger-tx-item')
    expect(items).toHaveLength(3)

    // 支出红：-¥100.00，分类名与账户名
    expect(items[0].text()).toContain('-¥100.00')
    expect(items[0].text()).toContain('餐饮')
    expect(items[0].text()).toContain('现金')
    expect(items[0].text()).toContain('聚餐')
    expect(items[0].find('.text-\\[\\#ff3b30\\]').exists()).toBe(true)

    // 收入绿：+¥8,000.00
    expect(items[1].text()).toContain('+¥8,000.00')
    expect(items[1].text()).toContain('工资')

    // 转账灰：无符号，显示 转出 → 转入
    expect(items[2].text()).toContain('¥500.00')
    expect(items[2].text()).not.toContain('-¥500.00')
    expect(items[2].text()).toContain('现金 → 招行储蓄卡')
  })

  it('confirms before deleting a transaction', async () => {
    const deleteSpy = vi.spyOn(ledgerApi, 'deleteTransaction').mockResolvedValue(true)
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(false)
    const wrapper = await mountList([
      {
        id: '100',
        type: 'expense',
        bookedAt: '2026-08-20 12:30:00',
        remark: '',
        postings: [
          { accountId: '1', amount: '-100', sort: 0 },
          { accountId: '999', amount: '100', categoryId: '10', sort: 1 },
        ],
      },
    ])

    await wrapper
      .findAll('button')
      .find((button) => button.attributes('title') === '删除')!
      .trigger('click')

    expect(confirmSpy).toHaveBeenCalledWith('确定要删除这条交易吗？')
    expect(deleteSpy).not.toHaveBeenCalled()

    confirmSpy.mockReturnValue(true)
    await wrapper
      .findAll('button')
      .find((button) => button.attributes('title') === '删除')!
      .trigger('click')
    await flushPromises()

    expect(deleteSpy).toHaveBeenCalledWith('100')
  })

  it('rebuilds split legs when editing an expense with multiple category legs', async () => {
    vi.spyOn(ledgerApi, 'getTransactionById').mockResolvedValue({
      id: '100',
      type: 'expense',
      bookedAt: '2026-08-20 12:30:00',
      remark: '聚餐',
      postings: [
        { accountId: '1', amount: '-100', sort: 0 },
        { accountId: '999', amount: '60', categoryId: '10', sort: 1 },
        { accountId: '999', amount: '40', categoryId: '11', sort: 2 },
      ],
    })
    const updateSpy = vi
      .spyOn(ledgerApi, 'updateTransaction')
      .mockResolvedValue({ id: '100', message: 'ok' })
    const wrapper = await mountList([
      {
        id: '100',
        type: 'expense',
        bookedAt: '2026-08-20 12:30:00',
        remark: '聚餐',
        postings: [
          { accountId: '1', amount: '-100', sort: 0 },
          { accountId: '999', amount: '60', categoryId: '10', sort: 1 },
          { accountId: '999', amount: '40', categoryId: '11', sort: 2 },
        ],
      },
    ])

    await wrapper
      .findAll('button')
      .find((button) => button.attributes('title') === '编辑')!
      .trigger('click')
    await flushPromises()

    // 回显：付款账户、拆分两腿（第二腿为二级分类午餐）
    const modal = wrapper.find('.ledger-tx-modal')
    expect((modal.find('select.tx-account').element as HTMLSelectElement).value).toBe('1')
    const legs = modal.findAll('.tx-leg')
    expect(legs).toHaveLength(2)
    expect((legs[0].find('.tx-leg-amount').element as HTMLInputElement).value).toBe('60')
    expect((legs[1].find('.tx-leg-amount').element as HTMLInputElement).value).toBe('40')
    expect((legs[1].find('.tx-leg-category').element as HTMLSelectElement).value).toBe('10')
    expect((legs[1].find('.tx-leg-sub-category').element as HTMLSelectElement).value).toBe('11')

    await submitModal(wrapper)

    // 整组替换：带 id，postings 结构不变
    expect(updateSpy).toHaveBeenCalledWith({
      id: '100',
      type: 'expense',
      bookedAt: '2026-08-20 12:30:00',
      remark: '聚餐',
      postings: [
        { accountId: '1', amount: '-100', sort: 0 },
        { accountId: 0, amount: '60', categoryId: '10', sort: 1 },
        { accountId: 0, amount: '40', categoryId: '11', sort: 2 },
      ],
    })
  })

  it('refetches transactions when the month filter changes', async () => {
    const wrapper = await mountList()
    const pageSpy = vi.spyOn(ledgerApi, 'getTransactionPage')
    pageSpy.mockClear()

    await wrapper.find('.ledger-filter-month').setValue('2026-07')
    await flushPromises()

    expect(pageSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        startTime: '2026-07-01 00:00:00',
        endTime: '2026-07-31 23:59:59',
      }),
    )
  })

  it('applies account/category/type filters on change', async () => {
    const wrapper = await mountList()
    const pageSpy = vi.spyOn(ledgerApi, 'getTransactionPage')
    pageSpy.mockClear()

    await wrapper.find('.ledger-filter-account').setValue('1')
    await flushPromises()
    expect(pageSpy).toHaveBeenLastCalledWith(expect.objectContaining({ accountId: '1' }))

    await wrapper.find('.ledger-filter-category').setValue('10')
    await flushPromises()
    expect(pageSpy).toHaveBeenLastCalledWith(
      expect.objectContaining({ accountId: '1', categoryId: '10' }),
    )

    await wrapper.find('.ledger-filter-type').setValue('expense')
    await flushPromises()
    expect(pageSpy).toHaveBeenLastCalledWith(
      expect.objectContaining({ accountId: '1', categoryId: '10', type: 'expense' }),
    )
  })

  it('silently applies recurring on mount and refreshes list and stats when created > 0', async () => {
    const statsSpy = vi.spyOn(ledgerApi, 'getMonthlyStats').mockResolvedValue({
      month: '2026-08',
      totalExpense: '0',
      totalIncome: '0',
      expenseByCategory: [],
      incomeByCategory: [],
    })
    const wrapper = await mountList([], 2)
    const applySpy = vi.spyOn(ledgerApi, 'applyRecurring')
    const pageSpy = vi.spyOn(ledgerApi, 'getTransactionPage')

    expect(applySpy).toHaveBeenCalledTimes(1)
    // 初次加载 + created>0 触发的刷新
    expect(pageSpy).toHaveBeenCalledTimes(2)
    expect(statsSpy).toHaveBeenCalledTimes(1)
    expect(wrapper.find('.ledger-tx-item').exists()).toBe(false)
  })

  it('does not refresh list or stats when apply creates nothing', async () => {
    const statsSpy = vi.spyOn(ledgerApi, 'getMonthlyStats')
    await mountList([], 0)
    const pageSpy = vi.spyOn(ledgerApi, 'getTransactionPage')

    expect(pageSpy).toHaveBeenCalledTimes(1)
    expect(statsSpy).not.toHaveBeenCalled()
  })

  it('stays usable and silent when apply fails', async () => {
    mockLedger([])
    vi.spyOn(ledgerApi, 'applyRecurring').mockRejectedValue(new Error('网络错误'))
    const pageSpy = vi.spyOn(ledgerApi, 'getTransactionPage')
    const { default: LedgerList } = await import('@/view/LedgerList.vue')
    const wrapper = mount(LedgerList, {
      global: { stubs: { RouterLink: RouterLinkStub } },
    })
    await flushPromises()

    // 列表照常加载，页面无报错打扰
    expect(wrapper.text()).toContain('本月暂无交易')
    expect(pageSpy).toHaveBeenCalledTimes(1)
  })
})
