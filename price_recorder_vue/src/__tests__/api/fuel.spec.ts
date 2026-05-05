import { beforeEach, describe, expect, it, vi } from 'vitest'

const getMock = vi.fn()
const postMock = vi.fn()

vi.mock('@/utils/request.ts', () => ({
  default: {
    get: getMock,
    post: postMock,
  },
}))

describe('fuel api', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    postMock.mockResolvedValue({})
  })

  it('serializes decimal fields as strings before posting vehicles', async () => {
    const { createVehicle } = await import('@/api/fuel')

    await createVehicle({
      name: 'Car',
      plateNo: '',
      brand: '',
      model: '',
      tankCapacity: 7 as any,
      remark: '',
    })

    expect(postMock).toHaveBeenCalledWith('/fuel/vehicle/save/v1', {
      name: 'Car',
      plateNo: '',
      brand: '',
      model: '',
      tankCapacity: '7',
      remark: '',
    })
  })

  it('serializes decimal fields as strings before posting refuel records', async () => {
    const { createRefuelRecord } = await import('@/api/fuel')

    await createRefuelRecord({
      vehicleId: '1',
      refuelTime: '2026-05-03 00:00:00',
      odometer: 1000 as any,
      volume: 30 as any,
      unitPrice: 7.5 as any,
      amount: 225 as any,
      station: '',
      isFull: true,
      remark: '',
    })

    expect(postMock).toHaveBeenCalledWith('/fuel/refuel/save/v1', {
      vehicleId: '1',
      refuelTime: '2026-05-03 00:00:00',
      odometer: '1000',
      volume: '30',
      unitPrice: '7.5',
      amount: '225',
      station: '',
      isFull: true,
      remark: '',
    })
  })
})
