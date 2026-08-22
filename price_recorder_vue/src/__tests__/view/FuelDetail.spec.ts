import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import * as fuelApi from '@/api/fuel'
import { uploadImage } from '@/api/file'

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

vi.mock('@/api/file', () => ({
  uploadImage: vi.fn(),
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

  it('shows existing attachments with type badges when editing a record', async () => {
    mockDashboard([
      {
        id: '2',
        odometer: '1500',
        isFull: false,
        attachmentInfos: [
          {
            id: '5',
            fileId: '9',
            url: '/file/download/v1/9',
            fileName: 'screen.jpg',
            attachType: 'station_screen',
            sort: 0,
          },
          {
            id: '6',
            fileId: '10',
            url: '/file/download/v1/10',
            fileName: 'legacy.jpg',
            attachType: 'receipt',
            sort: 1,
          },
        ],
      },
    ])
    const { default: FuelDetail } = await import('@/view/FuelDetail.vue')
    const wrapper = mount(FuelDetail)
    await flushPromises()

    await wrapper
      .findAll('button')
      .find((button) => button.attributes('title') === '编辑')!
      .trigger('click')
    await flushPromises()

    const thumb = wrapper.find('.attachment-picker img')
    expect(thumb.exists()).toBe(true)
    expect(thumb.attributes('src')).toBe('/api/file/download/v1/9')

    const badges = wrapper.findAll('.attachment-picker .attach-type-badge')
    expect(badges).toHaveLength(2)
    expect(badges[0].text()).toBe('加油站屏幕')
    // 旧数据的未知类型归入"其他"展示，不崩溃
    expect(badges[1].text()).toBe('其他')
    expect(wrapper.find('.attachment-picker select').exists()).toBe(false)
  })

  it('uploads images and submits attachment references on save', async () => {
    vi.mocked(uploadImage).mockResolvedValue({ id: '9', url: '/file/download/v1/9' })
    // 从加油站屏幕入口上传，上传成功后会自动触发 OCR
    vi.spyOn(fuelApi, 'recognizeFuelOcr').mockResolvedValue({
      rawText: '',
      amount: '225',
      volume: '30',
      unitPrice: '7.5',
      odometer: '',
    })
    mockDashboard([])
    const createSpy = vi.spyOn(fuelApi, 'createRefuelRecord')
    const { default: FuelDetail } = await import('@/view/FuelDetail.vue')
    const wrapper = mount(FuelDetail)
    await flushPromises()

    await wrapper
      .findAll('button')
      .find((button) => button.text().includes('新增加油'))!
      .trigger('click')
    await flushPromises()

    const fileInput = wrapper.find('.attachment-picker input[data-attach-type="station_screen"]')
    const file = new File(['fake-image'], 'screen.jpg', { type: 'image/jpeg' })
    Object.defineProperty(fileInput.element, 'files', { value: [file] })
    await fileInput.trigger('change')
    await flushPromises()

    expect(uploadImage).toHaveBeenCalledWith(file)
    expect(wrapper.find('.attachment-picker img').attributes('src')).toBe(
      '/api/file/download/v1/9',
    )

    const modalDate = wrapper.findAll('input[type="date"]').at(-1)!
    await modalDate.setValue('2026-05-03')
    await flushPromises()
    await wrapper
      .findAll('button')
      .find((button) => button.text().includes('保存'))!
      .trigger('click')
    await flushPromises()

    expect(createSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        attachments: [{ fileId: '9', attachType: 'station_screen', sort: 0 }],
      }),
    )
  })

  it('syncs actual amount with amount until manually edited', async () => {
    mockDashboard([])
    const { default: FuelDetail } = await import('@/view/FuelDetail.vue')
    const wrapper = mount(FuelDetail)
    await flushPromises()

    await wrapper
      .findAll('button')
      .find((button) => button.text().includes('新增加油'))!
      .trigger('click')
    await flushPromises()

    const amountInput = wrapper.find('input[placeholder="留空自动计算，手输须与油量×单价一致"]')
    const actualInput = wrapper.find('input[placeholder="默认与金额一致，可修改"]')

    await amountInput.setValue('225')
    expect((actualInput.element as HTMLInputElement).value).toBe('225')

    await amountInput.setValue('300')
    expect((actualInput.element as HTMLInputElement).value).toBe('300')

    await actualInput.setValue('280')
    await amountInput.setValue('350')
    expect((actualInput.element as HTMLInputElement).value).toBe('280')
  })

  it('submits actual amount with the record payload', async () => {
    mockDashboard([])
    const createSpy = vi.spyOn(fuelApi, 'createRefuelRecord')
    const { default: FuelDetail } = await import('@/view/FuelDetail.vue')
    const wrapper = mount(FuelDetail)
    await flushPromises()

    await wrapper
      .findAll('button')
      .find((button) => button.text().includes('新增加油'))!
      .trigger('click')
    await flushPromises()

    const amountInput = wrapper.find('input[placeholder="留空自动计算，手输须与油量×单价一致"]')
    const actualInput = wrapper.find('input[placeholder="默认与金额一致，可修改"]')
    await amountInput.setValue('225')
    await actualInput.setValue('200')

    const modalDate = wrapper.findAll('input[type="date"]').at(-1)!
    await modalDate.setValue('2026-05-03')
    await flushPromises()
    await wrapper
      .findAll('button')
      .find((button) => button.text().includes('保存'))!
      .trigger('click')
    await flushPromises()

    // type="number" 输入框的 v-model 自动转数字，字符串序列化由 api 层负责（fuel.spec.ts 覆盖）
    expect(createSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        amount: 225,
        actualAmount: 200,
      }),
    )
  })

  it('fills actual amount from record and does not relink when editing', async () => {
    mockDashboard([
      {
        id: '2',
        odometer: '1500',
        amount: '225',
        actualAmount: '200',
        isFull: false,
        refuelTime: '2026-05-03 00:00:00',
      },
    ])
    const { default: FuelDetail } = await import('@/view/FuelDetail.vue')
    const wrapper = mount(FuelDetail)
    await flushPromises()

    await wrapper
      .findAll('button')
      .find((button) => button.attributes('title') === '编辑')!
      .trigger('click')
    await flushPromises()

    const amountInput = wrapper.find('input[placeholder="留空自动计算，手输须与油量×单价一致"]')
    const actualInput = wrapper.find('input[placeholder="默认与金额一致，可修改"]')
    expect((actualInput.element as HTMLInputElement).value).toBe('200')

    await amountInput.setValue('300')
    expect((actualInput.element as HTMLInputElement).value).toBe('200')
  })

  it('shows actual amount with original amount reference when they differ', async () => {
    mockDashboard([{ id: '2', odometer: '1500', amount: '225', actualAmount: '200', isFull: true }])
    const { default: FuelDetail } = await import('@/view/FuelDetail.vue')
    const wrapper = mount(FuelDetail)
    await flushPromises()

    const amountCell = wrapper.find('tbody tr').findAll('td')[3]
    expect(amountCell.text()).toContain('¥200')
    expect(amountCell.text()).toContain('¥225')
  })

  it('omits original amount reference when amounts are equal', async () => {
    mockDashboard([{ id: '2', odometer: '1500', amount: '225', actualAmount: '225', isFull: true }])
    const { default: FuelDetail } = await import('@/view/FuelDetail.vue')
    const wrapper = mount(FuelDetail)
    await flushPromises()

    const amountCell = wrapper.find('tbody tr').findAll('td')[3]
    expect(amountCell.text()).toBe('¥225')
  })

  async function uploadViaEntry(
    wrapper: ReturnType<typeof mount>,
    file: File,
    attachType = 'station_screen',
  ) {
    const fileInput = wrapper.find(`.attachment-picker input[data-attach-type="${attachType}"]`)
    // configurable 允许同一入口连续传多张（other 场景会重复 define 同一元素的 files）
    Object.defineProperty(fileInput.element, 'files', { value: [file], configurable: true })
    await fileInput.trigger('change')
  }

  async function openCreateAndUpload(
    wrapper: ReturnType<typeof mount>,
    file: File,
    attachType = 'station_screen',
  ) {
    await wrapper
      .findAll('button')
      .find((button) => button.text().includes('新增加油'))!
      .trigger('click')
    await flushPromises()

    await uploadViaEntry(wrapper, file, attachType)
  }

  it('fills amount, volume and unit price from station screen OCR', async () => {
    vi.mocked(uploadImage).mockResolvedValue({ id: '9', url: '/file/download/v1/9' })
    const ocrSpy = vi.spyOn(fuelApi, 'recognizeFuelOcr').mockResolvedValue({
      rawText: 'raw',
      amount: '225',
      volume: '30',
      unitPrice: '7.5',
      odometer: '',
    })
    mockDashboard([])
    const { default: FuelDetail } = await import('@/view/FuelDetail.vue')
    const wrapper = mount(FuelDetail)
    await flushPromises()

    const file = new File(['fake-image'], 'screen.jpg', { type: 'image/jpeg' })
    await openCreateAndUpload(wrapper, file)
    await flushPromises()

    expect(ocrSpy).toHaveBeenCalledWith(file, 'station_screen')
    const amountInput = wrapper.find('input[placeholder="留空自动计算，手输须与油量×单价一致"]')
    const volumeInput = wrapper.find('input[placeholder="请输入油量"]')
    const unitPriceInput = wrapper.find('input[placeholder="请输入单价"]')
    expect((amountInput.element as HTMLInputElement).value).toBe('225')
    expect((volumeInput.element as HTMLInputElement).value).toBe('30')
    expect((unitPriceInput.element as HTMLInputElement).value).toBe('7.5')
  })

  it('does not overwrite fields the user already filled', async () => {
    vi.mocked(uploadImage).mockResolvedValue({ id: '9', url: '/file/download/v1/9' })
    vi.spyOn(fuelApi, 'recognizeFuelOcr').mockResolvedValue({
      rawText: 'raw',
      amount: '225',
      volume: '30',
      unitPrice: '7.5',
      odometer: '',
    })
    mockDashboard([])
    const { default: FuelDetail } = await import('@/view/FuelDetail.vue')
    const wrapper = mount(FuelDetail)
    await flushPromises()

    await wrapper
      .findAll('button')
      .find((button) => button.text().includes('新增加油'))!
      .trigger('click')
    await flushPromises()

    // 先填金额，再上传：OCR 回填只能补空字段
    const amountInput = wrapper.find('input[placeholder="留空自动计算，手输须与油量×单价一致"]')
    await amountInput.setValue('500')

    const file = new File(['fake-image'], 'screen.jpg', { type: 'image/jpeg' })
    const fileInput = wrapper.find('.attachment-picker input[data-attach-type="station_screen"]')
    Object.defineProperty(fileInput.element, 'files', { value: [file] })
    await fileInput.trigger('change')
    await flushPromises()

    expect((amountInput.element as HTMLInputElement).value).toBe('500')
    const volumeInput = wrapper.find('input[placeholder="请输入油量"]')
    const unitPriceInput = wrapper.find('input[placeholder="请输入单价"]')
    expect((volumeInput.element as HTMLInputElement).value).toBe('30')
    expect((unitPriceInput.element as HTMLInputElement).value).toBe('7.5')
  })

  it('fills odometer from dashboard OCR when uploading via the dashboard entry', async () => {
    vi.mocked(uploadImage).mockResolvedValue({ id: '9', url: '/file/download/v1/9' })
    const ocrSpy = vi.spyOn(fuelApi, 'recognizeFuelOcr').mockResolvedValue({
      rawText: 'raw',
      amount: '',
      volume: '',
      unitPrice: '',
      odometer: '12345',
    })
    mockDashboard([])
    const { default: FuelDetail } = await import('@/view/FuelDetail.vue')
    const wrapper = mount(FuelDetail)
    await flushPromises()

    const file = new File(['fake-image'], 'dashboard.jpg', { type: 'image/jpeg' })
    await openCreateAndUpload(wrapper, file, 'dashboard')
    await flushPromises()

    expect(ocrSpy).toHaveBeenCalledWith(file, 'dashboard')
    const odometerInput = wrapper.find('input[placeholder="请输入总里程"]')
    expect((odometerInput.element as HTMLInputElement).value).toBe('12345')
    expect(wrapper.find('.attachment-picker .attach-type-badge').text()).toBe('车辆仪表')
  })

  it('skips OCR when uploading via the other entry', async () => {
    vi.mocked(uploadImage).mockResolvedValue({ id: '9', url: '/file/download/v1/9' })
    const ocrSpy = vi.spyOn(fuelApi, 'recognizeFuelOcr')
    mockDashboard([])
    const { default: FuelDetail } = await import('@/view/FuelDetail.vue')
    const wrapper = mount(FuelDetail)
    await flushPromises()

    const file = new File(['fake-image'], 'other.jpg', { type: 'image/jpeg' })
    await openCreateAndUpload(wrapper, file, 'other')
    await flushPromises()

    expect(ocrSpy).not.toHaveBeenCalled()
    expect(wrapper.find('.attachment-picker img').exists()).toBe(true)
    expect(wrapper.find('.attachment-picker .attach-type-badge').text()).toBe('其他')
  })

  it('alerts on OCR failure without removing the attachment', async () => {
    vi.mocked(uploadImage).mockResolvedValue({ id: '9', url: '/file/download/v1/9' })
    vi.spyOn(fuelApi, 'recognizeFuelOcr').mockRejectedValue(new Error('ocr boom'))
    const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => {})
    mockDashboard([])
    const { default: FuelDetail } = await import('@/view/FuelDetail.vue')
    const wrapper = mount(FuelDetail)
    await flushPromises()

    const file = new File(['fake-image'], 'screen.jpg', { type: 'image/jpeg' })
    await openCreateAndUpload(wrapper, file)
    await flushPromises()

    expect(alertSpy).toHaveBeenCalledWith('识别失败，请手动填写')
    expect(wrapper.find('.attachment-picker img').attributes('src')).toBe(
      '/api/file/download/v1/9',
    )
  })

  it('hides the station screen entry after one upload while keeping dashboard and other', async () => {
    vi.mocked(uploadImage).mockResolvedValue({ id: '9', url: '/file/download/v1/9' })
    vi.spyOn(fuelApi, 'recognizeFuelOcr').mockResolvedValue({
      rawText: '',
      amount: '225',
      volume: '30',
      unitPrice: '7.5',
      odometer: '',
    })
    mockDashboard([])
    const { default: FuelDetail } = await import('@/view/FuelDetail.vue')
    const wrapper = mount(FuelDetail)
    await flushPromises()

    const file = new File(['fake-image'], 'screen.jpg', { type: 'image/jpeg' })
    await openCreateAndUpload(wrapper, file)
    await flushPromises()

    expect(
      wrapper.find('.attachment-picker input[data-attach-type="station_screen"]').exists(),
    ).toBe(false)
    expect(wrapper.find('.attachment-picker input[data-attach-type="dashboard"]').exists()).toBe(
      true,
    )
    expect(wrapper.find('.attachment-picker input[data-attach-type="other"]').exists()).toBe(true)
  })

  it('allows uploading multiple other attachments within the total limit', async () => {
    vi.mocked(uploadImage)
      .mockResolvedValueOnce({ id: '9', url: '/file/download/v1/9' })
      .mockResolvedValueOnce({ id: '10', url: '/file/download/v1/10' })
      .mockResolvedValueOnce({ id: '11', url: '/file/download/v1/11' })
    mockDashboard([])
    const { default: FuelDetail } = await import('@/view/FuelDetail.vue')
    const wrapper = mount(FuelDetail)
    await flushPromises()

    await wrapper
      .findAll('button')
      .find((button) => button.text().includes('新增加油'))!
      .trigger('click')
    await flushPromises()

    await uploadViaEntry(wrapper, new File(['a'], 'a.jpg', { type: 'image/jpeg' }), 'other')
    await flushPromises()
    await uploadViaEntry(wrapper, new File(['b'], 'b.jpg', { type: 'image/jpeg' }), 'other')
    await flushPromises()
    await uploadViaEntry(wrapper, new File(['c'], 'c.jpg', { type: 'image/jpeg' }), 'other')
    await flushPromises()

    expect(wrapper.findAll('.attachment-picker .attach-type-badge')).toHaveLength(3)
    expect(wrapper.find('.attachment-picker input[data-attach-type="other"]').exists()).toBe(true)
  })

  it('hides the station screen entry when editing a record that already has one', async () => {
    mockDashboard([
      {
        id: '2',
        odometer: '1500',
        isFull: false,
        attachmentInfos: [
          {
            id: '5',
            fileId: '9',
            url: '/file/download/v1/9',
            fileName: 'screen.jpg',
            attachType: 'station_screen',
            sort: 0,
          },
        ],
      },
    ])
    const { default: FuelDetail } = await import('@/view/FuelDetail.vue')
    const wrapper = mount(FuelDetail)
    await flushPromises()

    await wrapper
      .findAll('button')
      .find((button) => button.attributes('title') === '编辑')!
      .trigger('click')
    await flushPromises()

    expect(
      wrapper.find('.attachment-picker input[data-attach-type="station_screen"]').exists(),
    ).toBe(false)
    expect(wrapper.find('.attachment-picker input[data-attach-type="dashboard"]').exists()).toBe(
      true,
    )
    expect(wrapper.find('.attachment-picker input[data-attach-type="other"]').exists()).toBe(true)
  })
})
