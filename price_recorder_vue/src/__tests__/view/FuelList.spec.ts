import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import * as fuelApi from '@/api/fuel'

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({
      push: vi.fn(),
    }),
  }
})

vi.mock('vue-toastification', () => ({
  useToast: () => ({
    success: vi.fn(),
    error: vi.fn(),
  }),
}))

describe('FuelList.vue', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.restoreAllMocks()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('filters vehicles by search keyword with debounce', async () => {
    vi.useFakeTimers()
    const getVehiclesSpy = vi.spyOn(fuelApi, 'getVehicles').mockResolvedValue({
      page: '1',
      total: '0',
      list: [],
    })
    const { default: FuelList } = await import('@/view/FuelList.vue')
    const wrapper = mount(FuelList)
    await flushPromises()

    const input = wrapper.find('input[type="text"]')
    await input.setValue('沪A')

    vi.advanceTimersByTime(300)
    vi.useRealTimers()
    await flushPromises()

    expect(getVehiclesSpy).toHaveBeenLastCalledWith('1', '12', '沪A')
  })
})
