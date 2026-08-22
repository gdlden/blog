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
      actualAmount: 200,
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
      actualAmount: '200',
      station: '',
      isFull: true,
      remark: '',
    })
  })

  it('serializes attachments as-is when posting refuel records', async () => {
    const { createRefuelRecord } = await import('@/api/fuel')

    await createRefuelRecord({
      vehicleId: '1',
      refuelTime: '2026-05-03 00:00:00',
      odometer: '1000',
      volume: '30',
      unitPrice: '7.5',
      amount: '225',
      actualAmount: '200',
      station: '',
      isFull: true,
      remark: '',
      attachments: [
        { fileId: '9', attachType: 'station_screen', sort: 0 },
        { fileId: '10', attachType: 'dashboard', sort: 1 },
      ],
    })

    expect(postMock).toHaveBeenCalledWith('/fuel/refuel/save/v1', {
      vehicleId: '1',
      refuelTime: '2026-05-03 00:00:00',
      odometer: '1000',
      volume: '30',
      unitPrice: '7.5',
      amount: '225',
      actualAmount: '200',
      station: '',
      isFull: true,
      remark: '',
      attachments: [
        { fileId: '9', attachType: 'station_screen', sort: 0 },
        { fileId: '10', attachType: 'dashboard', sort: 1 },
      ],
    })
  })

  it('passes keyword as name and plateNo to vehicle list', async () => {
    const { getVehicles } = await import('@/api/fuel')

    await getVehicles('1', '12', '沪A')

    expect(getMock).toHaveBeenCalledWith('/fuel/vehicle/page/v1', {
      params: { page: '1', pageSize: '12', name: '沪A', plateNo: '沪A' },
    })
  })

  it('omits empty keyword from vehicle list params', async () => {
    const { getVehicles } = await import('@/api/fuel')

    await getVehicles('1', '12')

    expect(getMock).toHaveBeenCalledWith('/fuel/vehicle/page/v1', {
      params: { page: '1', pageSize: '12' },
    })
  })

  it('passes time range to stats params', async () => {
    const { getFuelStats } = await import('@/api/fuel')

    await getFuelStats('1', '2026-01-01', '2026-01-31')

    expect(getMock).toHaveBeenCalledWith('/fuel/stats/v1', {
      params: { vehicleId: '1', startTime: '2026-01-01', endTime: '2026-01-31' },
    })
  })

  it('omits empty time range from stats params', async () => {
    const { getFuelStats } = await import('@/api/fuel')

    await getFuelStats('1')

    expect(getMock).toHaveBeenCalledWith('/fuel/stats/v1', { params: { vehicleId: '1' } })
  })

  it('posts multipart form data for fuel OCR', async () => {
    const { recognizeFuelOcr } = await import('@/api/fuel')
    const file = new File(['fake-image'], 'screen.jpg', { type: 'image/jpeg' })

    await recognizeFuelOcr(file, 'station_screen')

    expect(postMock).toHaveBeenCalledTimes(1)
    const [url, body, config] = postMock.mock.calls[0]
    expect(url).toBe('/fuel/refuel/ocr/v1')
    expect(body).toBeInstanceOf(FormData)
    expect((body as FormData).get('file')).toBe(file)
    expect((body as FormData).get('attachType')).toBe('station_screen')
    expect(config).toEqual({ headers: { 'Content-Type': 'multipart/form-data' } })
  })
})
