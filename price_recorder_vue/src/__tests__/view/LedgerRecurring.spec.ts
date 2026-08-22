import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises, RouterLinkStub } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import * as ledgerApi from '@/api/ledger'
import type { LedgerRecurring } from '@/api/ledger'

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => ({ name: 'ledgerRecurring' }),
    useRouter: () => ({ push: vi.fn() }),
  }
})

vi.mock('vue-toastification', () => ({
  useToast: () => ({
    success: vi.fn(),
    error: vi.fn(),
  }),
}))

function mockBase(recurring: Array<Partial<LedgerRecurring>> = []) {
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
      { id: '10', parentId: '0', name: '居住', direction: 'expense', isSystem: false },
      { id: '11', parentId: '10', name: '房租', direction: 'expense', isSystem: false },
      { id: '20', parentId: '0', name: '工资', direction: 'income', isSystem: true },
    ],
  })
  vi.spyOn(ledgerApi, 'getRecurringList').mockResolvedValue({
    list: recurring as LedgerRecurring[],
  })
}

async function mountRecurring(recurring: Array<Partial<LedgerRecurring>> = []) {
  mockBase(recurring)
  const { default: LedgerRecurringView } = await import('@/view/LedgerRecurring.vue')
  const wrapper = mount(LedgerRecurringView, {
    global: { stubs: { RouterLink: RouterLinkStub } },
  })
  await flushPromises()
  return wrapper
}

const sampleItem: LedgerRecurring = {
  id: '1',
  accountId: '2',
  accountName: '招行储蓄卡',
  categoryId: '11',
  categoryName: '房租',
  type: 'expense',
  amount: '3500',
  remark: '每月房租',
  dayOfMonth: 5,
  startMonth: '2026-01',
  lastGeneratedMonth: '2026-08',
  enabled: true,
  nextDate: '2026-09-05',
}

describe('LedgerRecurring.vue', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.restoreAllMocks()
  })

  it('renders recurring rows with amount, account, category, day and next date', async () => {
    const wrapper = await mountRecurring([
      sampleItem,
      {
        ...sampleItem,
        id: '2',
        categoryId: '20',
        categoryName: '工资',
        type: 'income',
        amount: '12000',
        remark: '',
        dayOfMonth: 10,
        enabled: false,
        nextDate: '-',
      },
    ])

    const items = wrapper.findAll('.recurring-item')
    expect(items).toHaveLength(2)

    expect(items[0].text()).toContain('房租')
    expect(items[0].text()).toContain('招行储蓄卡')
    expect(items[0].text()).toContain('每月 5 号')
    expect(items[0].text()).toContain('2026-09-05')
    expect(items[0].text()).toContain('每月房租')
    expect(items[0].text()).toContain('-¥3,500.00')

    expect(items[1].text()).toContain('+¥12,000.00')
    // 停用行降透明度
    expect(items[1].classes()).toContain('opacity-60')
  })

  it('creates recurring with two-level category cascade filtered by type', async () => {
    const saveSpy = vi.spyOn(ledgerApi, 'saveRecurring').mockResolvedValue({ id: '3', message: 'ok' })
    const wrapper = await mountRecurring()

    await wrapper
      .findAll('button')
      .find((button) => button.text().includes('新增周期账单'))!
      .trigger('click')
    await flushPromises()

    const modal = wrapper.find('.recurring-modal')
    await modal.find('select.recurring-account').setValue('2')
    await modal.find('input.recurring-amount').setValue('3500')
    // 支出方向只有居住分类；选一级后出现子分类
    await modal.find('select.recurring-category').setValue('10')
    await flushPromises()
    await modal.find('select.recurring-sub-category').setValue('11')
    await modal.find('input.recurring-day').setValue('5')
    await modal.find('input.recurring-remark').setValue('每月房租')

    // 开始月份默认当前月
    expect((modal.find('input.recurring-month').element as HTMLInputElement).value).toMatch(
      /^\d{4}-\d{2}$/,
    )

    await modal.find('.recurring-submit').trigger('click')
    await flushPromises()

    expect(saveSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        id: undefined,
        accountId: '2',
        categoryId: '11',
        type: 'expense',
        amount: 3500,
        remark: '每月房租',
        dayOfMonth: 5,
        enabled: true,
      }),
    )
  })

  it('filters category options by the type tab', async () => {
    const wrapper = await mountRecurring()
    await wrapper
      .findAll('button')
      .find((button) => button.text().includes('新增周期账单'))!
      .trigger('click')
    await flushPromises()

    const modal = wrapper.find('.recurring-modal')
    // 切到收入 Tab：一级分类只剩工资
    await wrapper
      .findAll('.recurring-tab')
      .find((tab) => tab.text() === '收入')!
      .trigger('click')
    await flushPromises()

    const options = modal.find('select.recurring-category').findAll('option')
    expect(options.map((option) => option.text())).toEqual(['请选择分类', '工资'])
  })

  it('blocks submit when required fields are missing', async () => {
    const saveSpy = vi.spyOn(ledgerApi, 'saveRecurring')
    const wrapper = await mountRecurring()

    await wrapper
      .findAll('button')
      .find((button) => button.text().includes('新增周期账单'))!
      .trigger('click')
    await flushPromises()

    const submit = wrapper.find('.recurring-submit')
    expect(submit.attributes('disabled')).toBeDefined()

    await submit.trigger('click')
    await flushPromises()
    expect(saveSpy).not.toHaveBeenCalled()
  })

  it('fills the edit modal with the item and submits with id', async () => {
    const saveSpy = vi.spyOn(ledgerApi, 'saveRecurring').mockResolvedValue({ id: '1', message: 'ok' })
    const wrapper = await mountRecurring([sampleItem])

    await wrapper
      .findAll('button')
      .find((button) => button.attributes('title') === '编辑')!
      .trigger('click')
    await flushPromises()

    const modal = wrapper.find('.recurring-modal')
    expect(modal.find('h3').text()).toBe('编辑周期账单')
    expect((modal.find('select.recurring-account').element as HTMLSelectElement).value).toBe('2')
    // 二级分类拆回父/子：居住 + 房租
    expect((modal.find('select.recurring-category').element as HTMLSelectElement).value).toBe('10')
    expect((modal.find('select.recurring-sub-category').element as HTMLSelectElement).value).toBe(
      '11',
    )
    expect((modal.find('input.recurring-day').element as HTMLInputElement).value).toBe('5')
    // 编辑时类型 Tab 禁用
    expect(modal.find('.recurring-tab').attributes('disabled')).toBeDefined()

    await modal.find('.recurring-submit').trigger('click')
    await flushPromises()

    expect(saveSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        id: '1',
        accountId: '2',
        categoryId: '11',
        type: 'expense',
        amount: '3500',
        dayOfMonth: 5,
        startMonth: '2026-01',
        enabled: true,
      }),
    )
  })

  it('toggles enabled via save with the full object', async () => {
    const saveSpy = vi.spyOn(ledgerApi, 'saveRecurring').mockResolvedValue({ id: '1', message: 'ok' })
    const wrapper = await mountRecurring([sampleItem])

    await wrapper.find('.recurring-toggle').trigger('click')
    await flushPromises()

    expect(saveSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        id: '1',
        accountId: '2',
        categoryId: '11',
        enabled: false,
      }),
    )
  })

  it('confirms before deleting a recurring item', async () => {
    const deleteSpy = vi.spyOn(ledgerApi, 'deleteRecurring').mockResolvedValue(true)
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(false)
    const wrapper = await mountRecurring([sampleItem])

    await wrapper.find('.recurring-delete').trigger('click')
    expect(confirmSpy).toHaveBeenCalledWith('确定要删除周期账单「每月房租」吗？')
    expect(deleteSpy).not.toHaveBeenCalled()

    confirmSpy.mockReturnValue(true)
    await wrapper.find('.recurring-delete').trigger('click')
    await flushPromises()
    expect(deleteSpy).toHaveBeenCalledWith('1')
  })
})
