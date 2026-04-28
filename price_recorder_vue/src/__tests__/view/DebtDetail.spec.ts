import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import * as debtApi from '@/api/debt'
import * as debtDetailApi from '@/api/debtDetail'

const pushMock = vi.fn()

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => ({
      params: { id: '42' },
    }),
    useRouter: () => ({
      push: pushMock,
    }),
  }
})

vi.mock('vue-toastification', () => ({
  useToast: () => ({
    success: vi.fn(),
    error: vi.fn(),
  }),
}))

describe('DebtDetail.vue', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    pushMock.mockClear()
    vi.restoreAllMocks()
    vi.spyOn(debtApi, 'getDebtById').mockResolvedValue({
      id: '42',
      name: '房贷',
      bankName: '测试银行',
      bankAccount: '',
      applyTime: '',
      endTime: '',
      amount: '100000',
      status: '进行中',
      remark: '',
      apr: '3.5',
      fee: '',
      tenor: '360',
    })
    vi.spyOn(debtDetailApi, 'getDebtDetails').mockResolvedValue([])
  })

  it('fills the create detail form from the first OCR item', async () => {
    const { default: DebtDetail } = await import('@/view/DebtDetail.vue')
    const recognizeSpy = vi.spyOn(debtDetailApi, 'recognizeDebtDetailOcr').mockResolvedValue({
      rawText: '第1期 本金: 1000.00 利息: 12.34 入账日: 2026-03-25',
      items: [
        {
          debtId: '42',
          postingDate: '2026-03-25 00:00:00',
          principal: '1000.00',
          interest: '12.34',
          period: '1',
        },
      ],
    })

    const wrapper = mount(DebtDetail)
    await flushPromises()

    await wrapper.findAll('button').find((button) => button.text() === '新增明细')!.trigger('click')
    const fileInput = wrapper.find('[data-testid="debt-detail-ocr-file"]')
    const file = new File(['image'], 'repayment.png', { type: 'image/png' })
    Object.defineProperty(fileInput.element, 'files', { value: [file] })

    await fileInput.trigger('change')
    await flushPromises()

    expect(recognizeSpy).toHaveBeenCalledWith(file, '42')
    expect((wrapper.find('input[type="date"]').element as HTMLInputElement).value).toBe('2026-03-25')
    const numberInputs = wrapper.findAll('input[type="number"]')
    expect((numberInputs[0].element as HTMLInputElement).value).toBe('1')
    expect((numberInputs[1].element as HTMLInputElement).value).toBe('1000.00')
    expect((numberInputs[2].element as HTMLInputElement).value).toBe('12.34')
  })
})
