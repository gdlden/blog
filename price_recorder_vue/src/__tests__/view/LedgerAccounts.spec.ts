import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises, RouterLinkStub } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import * as ledgerApi from '@/api/ledger'
import type { LedgerAccount } from '@/api/ledger'

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => ({ name: 'ledgerAccounts' }),
    useRouter: () => ({ push: vi.fn() }),
  }
})

const toastSuccess = vi.fn()
const toastError = vi.fn()
vi.mock('vue-toastification', () => ({
  useToast: () => ({
    success: toastSuccess,
    error: toastError,
  }),
}))

function mockAccounts(accounts: Array<Partial<LedgerAccount>> = []) {
  vi.spyOn(ledgerApi, 'getAccountList').mockResolvedValue({
    list: accounts as LedgerAccount[],
  })
}

async function mountAccounts(accounts: Array<Partial<LedgerAccount>> = []) {
  mockAccounts(accounts)
  const { default: LedgerAccounts } = await import('@/view/LedgerAccounts.vue')
  const wrapper = mount(LedgerAccounts, {
    global: { stubs: { RouterLink: RouterLinkStub } },
  })
  await flushPromises()
  return wrapper
}

async function openCreateModal(wrapper: ReturnType<typeof mount>) {
  await wrapper
    .findAll('button')
    .find((button) => button.text().includes('新增账户'))!
    .trigger('click')
  await flushPromises()
}

describe('LedgerAccounts.vue', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.restoreAllMocks()
    toastSuccess.mockClear()
    toastError.mockClear()
  })

  it('shows credit card fields only when subtype is credit_card', async () => {
    const wrapper = await mountAccounts()
    await openCreateModal(wrapper)

    const modal = wrapper.find('.ledger-account-modal')
    // 默认 asset/cash：无信用卡字段
    expect(modal.find('.account-credit-limit').exists()).toBe(false)

    // 切到负债 → subtype 联动重置为 credit_card，信用卡字段出现
    await modal.find('select.account-type').setValue('liability')
    await flushPromises()
    expect((modal.find('select.account-subtype').element as HTMLSelectElement).value).toBe(
      'credit_card',
    )
    expect(modal.find('.account-credit-limit').exists()).toBe(true)
    expect(modal.find('.account-billing-day').exists()).toBe(true)
    expect(modal.find('.account-payment-due-day').exists()).toBe(true)

    // 切到其他负债 subtype → 字段消失
    await modal.find('select.account-subtype').setValue('loan_payable')
    await flushPromises()
    expect(modal.find('.account-credit-limit').exists()).toBe(false)
  })

  it('creates account with serialized payload including openingBalance', async () => {
    const saveSpy = vi
      .spyOn(ledgerApi, 'saveAccount')
      .mockResolvedValue({ id: '3', message: 'ok' })
    const wrapper = await mountAccounts()
    await openCreateModal(wrapper)

    const modal = wrapper.find('.ledger-account-modal')
    await modal.find('input.account-name').setValue('招行信用卡')
    await modal.find('select.account-type').setValue('liability')
    await flushPromises()
    await modal.find('input.account-opening-balance').setValue('-1200.5')
    await modal.find('input.account-credit-limit').setValue('50000')
    await modal.find('input.account-billing-day').setValue('5')
    await modal.find('input.account-payment-due-day').setValue('25')

    await wrapper
      .findAll('button')
      .find((button) => button.text() === '保存')!
      .trigger('click')
    await flushPromises()

    expect(saveSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        name: '招行信用卡',
        type: 'liability',
        subtype: 'credit_card',
        openingBalance: -1200.5,
        creditLimit: 50000,
        billingDay: 5,
        paymentDueDay: 25,
      }),
    )
  })

  it('toasts backend message when account deletion is rejected', async () => {
    vi.spyOn(ledgerApi, 'deleteAccount').mockRejectedValue(new Error('账户存在交易，无法删除'))
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    vi.spyOn(window, 'alert').mockImplementation(() => {})
    const wrapper = await mountAccounts([
      { id: '1', name: '现金', type: 'asset', subtype: 'cash', balance: '100' },
    ])

    await wrapper
      .findAll('button')
      .find((button) => button.attributes('title') === '删除')!
      .trigger('click')
    await flushPromises()

    expect(toastError).toHaveBeenCalledWith('账户存在交易，无法删除')
  })

  it('toggles archive via update', async () => {
    const updateSpy = vi
      .spyOn(ledgerApi, 'updateAccount')
      .mockResolvedValue({ id: '1', message: 'ok' })
    const wrapper = await mountAccounts([
      { id: '1', name: '旧卡', type: 'asset', subtype: 'debit_card', balance: '0', archived: false },
    ])

    await wrapper
      .findAll('button')
      .find((button) => button.text() === '归档')!
      .trigger('click')
    await flushPromises()

    expect(updateSpy).toHaveBeenCalledWith(expect.objectContaining({ id: '1', archived: true }))
  })

  it('shows a red reminder badge on credit cards due within 7 days', async () => {
    // 固定今天 2026-08-22：账单日 5 → 本期 (07-05, 08-05]，还款日 08-25 还有 3 天
    vi.useFakeTimers({ now: new Date(2026, 7, 22), toFake: ['Date'] })
    try {
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
              { accountId: '1', amount: '-1200', sort: 0 },
              { accountId: '999', amount: '1200', categoryId: '10', sort: 1 },
            ],
          },
        ],
      })
      const wrapper = await mountAccounts([
        {
          id: '1',
          name: '招行信用卡',
          type: 'liability',
          subtype: 'credit_card',
          creditLimit: '50000',
          billingDay: 5,
          paymentDueDay: 25,
          balance: '-1200',
          archived: false,
        },
        { id: '2', name: '现金', type: 'asset', subtype: 'cash', balance: '100', archived: false },
      ])

      const badges = wrapper.findAll('.credit-reminder-badge')
      expect(badges).toHaveLength(1)
      expect(badges[0].text()).toBe('还有 3 天还款')
      expect(badges[0].classes()).toContain('text-[#ff3b30]')

      // 无账单字段的账户不出角标
      const cashCard = wrapper
        .findAll('.ledger-account-card')
        .find((card) => card.text().includes('现金'))!
      expect(cashCard.find('.credit-reminder-badge').exists()).toBe(false)
    } finally {
      vi.useRealTimers()
    }
  })

  it('shows a neutral badge when the due date is more than 7 days away', async () => {
    // 固定今天 2026-08-10：账单日 5 → 本期 (07-05, 08-05]，还款日 08-25 还有 15 天
    vi.useFakeTimers({ now: new Date(2026, 7, 10), toFake: ['Date'] })
    try {
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
              { accountId: '1', amount: '-1200', sort: 0 },
              { accountId: '999', amount: '1200', categoryId: '10', sort: 1 },
            ],
          },
        ],
      })
      const wrapper = await mountAccounts([
        {
          id: '1',
          name: '招行信用卡',
          type: 'liability',
          subtype: 'credit_card',
          billingDay: 5,
          paymentDueDay: 25,
          balance: '-1200',
          archived: false,
        },
      ])

      const badge = wrapper.find('.credit-reminder-badge')
      expect(badge.exists()).toBe(true)
      expect(badge.text()).toBe('还有 15 天还款')
      expect(badge.classes()).toContain('text-[#0071e3]')
    } finally {
      vi.useRealTimers()
    }
  })
})
