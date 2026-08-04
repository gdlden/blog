import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useFuelStore } from '@/stores/fuelStore'
import * as fuelApi from '@/api/fuel'

vi.mock('vue-toastification', () => ({
  useToast: () => ({
    success: vi.fn(),
    error: vi.fn(),
  }),
}))

vi.mock('@/api/fuel', () => ({
  getVehicles: vi.fn(),
  createVehicle: vi.fn(),
  getFuelStats: vi.fn(),
  getRefuelRecords: vi.fn(),
}))

describe('fuelStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('loads vehicle list and pagination state', async () => {
    vi.mocked(fuelApi.getVehicles).mockResolvedValue({
      page: '1',
      total: '1',
      list: [
        {
          id: '1',
          name: 'Civic',
          plateNo: '沪A12345',
          brand: 'Honda',
          model: 'Civic',
          tankCapacity: '47',
          remark: '',
        },
      ],
    })

    const store = useFuelStore()
    await store.fetchVehicles(1, 12)

    expect(store.vehicles).toHaveLength(1)
    expect(store.total).toBe(1)
    expect(store.loading).toBe(false)
    expect(fuelApi.getVehicles).toHaveBeenCalledWith('1', '12', '')
  })

  it('creates vehicle then refreshes list', async () => {
    vi.mocked(fuelApi.createVehicle).mockResolvedValue({ id: '2', message: 'save success' })
    vi.mocked(fuelApi.getVehicles).mockResolvedValue({ page: '1', total: '0', list: [] })

    const store = useFuelStore()
    await store.createVehicle({
      name: 'Fit',
      plateNo: '',
      brand: '',
      model: '',
      tankCapacity: '',
      remark: '',
    })

    expect(fuelApi.createVehicle).toHaveBeenCalled()
    expect(fuelApi.getVehicles).toHaveBeenCalled()
  })

  it('loads stats and records for a vehicle', async () => {
    vi.mocked(fuelApi.getFuelStats).mockResolvedValue({
      vehicleId: '1',
      totalDistance: '600',
      totalVolume: '45',
      totalAmount: '315',
      averageConsumption: '7.50',
      latestConsumption: '7.50',
      costPerKm: '0.53',
      trend: [],
    })
    vi.mocked(fuelApi.getRefuelRecords).mockResolvedValue({ page: '1', total: '0', list: [] })

    const store = useFuelStore()
    await store.fetchVehicleDashboard('1')

    expect(store.stats?.averageConsumption).toBe('7.50')
    expect(store.records).toEqual([])
  })

  it('passes search keyword to vehicle list', async () => {
    vi.mocked(fuelApi.getVehicles).mockResolvedValue({ page: '1', total: '0', list: [] })

    const store = useFuelStore()
    await store.fetchVehicles(1, 12, '沪A')

    expect(fuelApi.getVehicles).toHaveBeenCalledWith('1', '12', '沪A')
  })

  it('passes time range to stats', async () => {
    vi.mocked(fuelApi.getFuelStats).mockResolvedValue({
      vehicleId: '1',
      totalDistance: '600',
      totalVolume: '45',
      totalAmount: '315',
      averageConsumption: '7.50',
      latestConsumption: '7.50',
      costPerKm: '0.53',
      trend: [],
    })

    const store = useFuelStore()
    await store.fetchStats('1', '2026-01-01', '2026-01-31')

    expect(fuelApi.getFuelStats).toHaveBeenCalledWith('1', '2026-01-01', '2026-01-31')
  })

  it('keeps last search keyword when paginating', async () => {
    vi.mocked(fuelApi.getVehicles).mockResolvedValue({ page: '1', total: '0', list: [] })

    const store = useFuelStore()
    await store.fetchVehicles(1, 12, '本田')
    await store.fetchVehicles(2)

    // 第二次不带 keyword，应沿用上次搜索词
    expect(fuelApi.getVehicles).toHaveBeenLastCalledWith('2', '12', '本田')
  })
})
