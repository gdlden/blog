import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import * as fuelApi from '@/api/fuel'

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => ({
      params: { vehicleId: '1' },
    }),
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

function mockDashboard(records: Array<Partial<fuelApi.RefuelRecord>>) {
  vi.spyOn(fuelApi, 'getVehicleById').mockResolvedValue({
    id: '1',
    name: 'Test Car',
    plateNo: '沪A12345',
    brand: '',
    model: '',
    tankCapacity: '50',
    remark: '',
  })
  vi.spyOn(fuelApi, 'getRefuelRecords').mockResolvedValue({
    page: '1',
    total: String(records.length),
    list: records as fuelApi.RefuelRecord[],
  })
  vi.spyOn(fuelApi, 'getFuelStats').mockResolvedValue({
    vehicleId: '1',
    totalDistance: '600',
    totalVolume: '45',
    totalAmount: '315',
    averageConsumption: '7.50',
    latestConsumption: '7.50',
    costPerKm: '0.53',
    trend: [],
  })
}

describe('FuelDetail.vue', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.restoreAllMocks()
  })

  it('warns when new record odometer is lower than the latest record', async () => {
    mockDashboard([{ id: '2', odometer: '1500', isFull: true }])
    const { default: FuelDetail } = await import('@/view/FuelDetail.vue')
    const wrapper = mount(FuelDetail)
    await flushPromises()

    await wrapper.findAll('button').find((button) => button.text().includes('新增加油'))!.trigger('click')
    await flushPromises()

    const odometerInput = wrapper.findAll('input[type="number"]')[0]
    await odometerInput.setValue('1000')
    await flushPromises()

    expect(wrapper.text()).toContain('里程回拨')
  })

  it('does not warn when new record odometer is higher than the latest record', async () => {
    mockDashboard([{ id: '2', odometer: '1500', isFull: true }])
    const { default: FuelDetail } = await import('@/view/FuelDetail.vue')
    const wrapper = mount(FuelDetail)
    await flushPromises()

    await wrapper.findAll('button').find((button) => button.text().includes('新增加油'))!.trigger('click')
    await flushPromises()

    const odometerInput = wrapper.findAll('input[type="number"]')[0]
    await odometerInput.setValue('1600')
    await flushPromises()

    expect(wrapper.text()).not.toContain('里程回拨')
  })

  it('confirms before editing a full-tank record', async () => {
    mockDashboard([{ id: '2', odometer: '1500', isFull: true }])
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true)
    const { default: FuelDetail } = await import('@/view/FuelDetail.vue')
    const wrapper = mount(FuelDetail)
    await flushPromises()

    await wrapper.findAll('button').find((button) => button.attributes('title') === '编辑')!.trigger('click')

    expect(confirmSpy).toHaveBeenCalledWith('修改加满记录将重新计算后续油耗统计，确定继续吗？')
  })

  it('uses plain confirm text for non-full records when deleting', async () => {
    mockDashboard([{ id: '2', odometer: '1500', isFull: false }])
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(false)
    const { default: FuelDetail } = await import('@/view/FuelDetail.vue')
    const wrapper = mount(FuelDetail)
    await flushPromises()

    await wrapper.findAll('button').find((button) => button.attributes('title') === '删除')!.trigger('click')

    expect(confirmSpy).toHaveBeenCalledWith('确定要删除这条加油记录吗？')
  })

  it('refetches stats with time range when start date changes', async () => {
    mockDashboard([])
    const { default: FuelDetail } = await import('@/view/FuelDetail.vue')
    const wrapper = mount(FuelDetail)
    await flushPromises()

    const statsSpy = vi.spyOn(fuelApi, 'getFuelStats')
    const dateInputs = wrapper.findAll('input[type="date"]')
    await dateInputs[0].setValue('2026-01-01')
    await flushPromises()

    expect(statsSpy).toHaveBeenCalledWith('1', '2026-01-01', undefined)
  })
})
