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
    expect(fuelApi.getVehicles).toHaveBeenCalledWith('1', '12')
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
})
