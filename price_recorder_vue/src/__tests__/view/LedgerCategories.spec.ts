import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises, RouterLinkStub } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import * as ledgerApi from '@/api/ledger'
import type { LedgerCategory } from '@/api/ledger'

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => ({ name: 'ledgerCategories' }),
    useRouter: () => ({ push: vi.fn() }),
  }
})

vi.mock('vue-toastification', () => ({
  useToast: () => ({
    success: vi.fn(),
    error: vi.fn(),
  }),
}))

const defaultCategories: Array<Partial<LedgerCategory>> = [
  { id: '10', parentId: '0', name: '餐饮', direction: 'expense', isSystem: true },
  { id: '11', parentId: '10', name: '午餐', direction: 'expense', isSystem: false },
  { id: '12', parentId: '0', name: '交通', direction: 'expense', isSystem: false },
  { id: '20', parentId: '0', name: '工资', direction: 'income', isSystem: true },
]

async function mountCategories(categories: Array<Partial<LedgerCategory>> = defaultCategories) {
  vi.spyOn(ledgerApi, 'getCategoryList').mockResolvedValue({
    list: categories as LedgerCategory[],
  })
  const { default: LedgerCategories } = await import('@/view/LedgerCategories.vue')
  const wrapper = mount(LedgerCategories, {
    global: { stubs: { RouterLink: RouterLinkStub } },
  })
  await flushPromises()
  return wrapper
}

describe('LedgerCategories.vue', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.restoreAllMocks()
  })

  it('renders expense and income groups with two-level tree', async () => {
    const wrapper = await mountCategories()

    const expenseGroup = wrapper.find('.category-group[data-direction="expense"]')
    const incomeGroup = wrapper.find('.category-group[data-direction="income"]')
    expect(expenseGroup.text()).toContain('餐饮')
    expect(expenseGroup.text()).toContain('午餐')
    expect(expenseGroup.text()).toContain('交通')
    expect(incomeGroup.text()).toContain('工资')
    expect(incomeGroup.text()).not.toContain('餐饮')

    // 子分类行缩进在父分类之下
    const children = expenseGroup.findAll('.category-child')
    expect(children).toHaveLength(1)
    expect(children[0].text()).toContain('午餐')
  })

  it('hides delete button for system categories', async () => {
    const wrapper = await mountCategories()

    const expenseGroup = wrapper.find('.category-group[data-direction="expense"]')
    const items = expenseGroup.findAll('.category-item')
    const systemItem = items.find((item) => item.text().includes('餐饮'))!
    const normalItem = items.find((item) => item.text().includes('交通'))!

    expect(systemItem.text()).toContain('内置')
    expect(systemItem.find('.category-delete').exists()).toBe(false)
    expect(normalItem.find('.category-delete').exists()).toBe(true)
  })

  it('confirms before deleting a category and calls api on confirm', async () => {
    const deleteSpy = vi.spyOn(ledgerApi, 'deleteCategory').mockResolvedValue(true)
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true)
    const wrapper = await mountCategories()

    const expenseGroup = wrapper.find('.category-group[data-direction="expense"]')
    const trafficItem = expenseGroup
      .findAll('.category-item')
      .find((item) => item.text().includes('交通'))!
    await trafficItem.find('.category-delete').trigger('click')
    await flushPromises()

    expect(confirmSpy).toHaveBeenCalledWith('确定要删除分类「交通」吗？')
    expect(deleteSpy).toHaveBeenCalledWith('12')
  })

  it('blocks deleting a category that still has children', async () => {
    const deleteSpy = vi.spyOn(ledgerApi, 'deleteCategory')
    const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => {})
    // 餐饮虽是内置分类没有删除按钮，这里构造非内置父分类验证子分类拦截
    const wrapper = await mountCategories([
      { id: '30', parentId: '0', name: '购物', direction: 'expense', isSystem: false },
      { id: '31', parentId: '30', name: '衣服', direction: 'expense', isSystem: false },
    ])

    const expenseGroup = wrapper.find('.category-group[data-direction="expense"]')
    const parentItem = expenseGroup
      .findAll('.category-item')
      .find((item) => item.text().includes('购物'))!
    await parentItem.find('.category-delete').trigger('click')
    await flushPromises()

    expect(alertSpy).toHaveBeenCalledWith('请先删除该分类下的子分类')
    expect(deleteSpy).not.toHaveBeenCalled()
  })

  it('creates a child category with parentId preset from the row action', async () => {
    const saveSpy = vi.spyOn(ledgerApi, 'saveCategory').mockResolvedValue({ id: '40', message: 'ok' })
    const wrapper = await mountCategories()

    const expenseGroup = wrapper.find('.category-group[data-direction="expense"]')
    const parentItem = expenseGroup
      .findAll('.category-item')
      .find((item) => item.text().includes('交通'))!
    await parentItem.find('.category-add-child').trigger('click')
    await flushPromises()

    const modal = wrapper.find('.ledger-category-modal')
    expect((modal.find('select.category-parent').element as HTMLSelectElement).value).toBe('12')
    await modal.find('input.category-name').setValue('地铁')
    await wrapper
      .findAll('button')
      .find((button) => button.text() === '保存')!
      .trigger('click')
    await flushPromises()

    expect(saveSpy).toHaveBeenCalledWith(
      expect.objectContaining({ name: '地铁', direction: 'expense', parentId: '12' }),
    )
  })
})
