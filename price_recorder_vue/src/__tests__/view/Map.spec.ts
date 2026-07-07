/* eslint-disable @typescript-eslint/no-explicit-any */
// D-22 skip-list — behaviors jsdom cannot instantiate. Each `vi.skipIf(...)` test
// below is re-routed to .planning/phases/13-fishing-spot-map/13-VALIDATION.md
// ## Manual-Only Verifications. Do NOT silently lose them — they MUST be run
// by a human on a real device before Phase 13 is declared complete.
//
// SKIP Tests (from 13-03-PLAN.md frontmatter `skipped_tests`):
//   SKIP-pin-clustering      -> 13-VALIDATION Manual-Only — Pin clustering under 50+ spots
//   SKIP-touch-pinch-zoom   -> 13-VALIDATION Manual-Only — Mobile map interaction (pinch-zoom, pan, tap)
//   SKIP-geolocation-denied -> 13-VALIDATION Manual-Only — Geolocation-denied fallback flow on a real phone
//   SKIP-real-gps-accuracy  -> 13-VALIDATION Manual-Only — GPS capture in field accuracy
//   SKIP-gaode-deep-link    -> 13-VALIDATION Manual-Only — Gaode deep-link navigation opens Gaode App
//   SKIP-real-tile-render   -> 13-VALIDATION Manual-Only — Real map tile rendering
//   SKIP-https-geolocation  -> 13-VALIDATION Manual-Only — Geolocation over HTTPS in production

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { reactive } from 'vue'
import type { SpotEntity } from '@/api/map'

// Stubs for vue-router + toast; the Pinia store mock below is constructed
// per-test so selectedSpot is reactive.
const pushMock = vi.fn()
const setSelectedSpotMock = vi.fn()
const fetchSpotsMock = vi.fn().mockResolvedValue(undefined)
const saveSpotMock = vi.fn().mockResolvedValue(undefined)
const reverseGeocodeMock = vi.fn().mockResolvedValue('测试地址')
const uploadSpotPhotoMock = vi.fn().mockResolvedValue({ url: 'https://cdn/x.png', name: 'p.png' })

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<any>('vue-router')
  return {
    ...actual,
    useRoute: () => ({ name: 'map' }),
    useRouter: () => ({ push: pushMock }),
  }
})

vi.mock('vue-toastification', () => ({
  useToast: () => ({
    success: vi.fn(),
    error: vi.fn(),
  }),
}))

const stubState = reactive({ selectedSpot: null as SpotEntity | null })

vi.mock('@/stores/mapStore', () => ({
  useMapStore: () => ({
    spots: [] as SpotEntity[],
    get selectedSpot() {
      return stubState.selectedSpot
    },
    loading: false,
    uploadingPhotos: [] as any[],
    selectedTags: [] as string[],
    searchQuery: '',
    spotCount: 0,
    allTags: [] as string[],
    filteredSpots: [] as SpotEntity[],
    fetchSpots: fetchSpotsMock,
    saveSpot: saveSpotMock,
    updateSpot: vi.fn(),
    deleteSpot: vi.fn(),
    getSpot: vi.fn(),
    reverseGeocode: reverseGeocodeMock,
    uploadSpotPhoto: uploadSpotPhotoMock,
    removeUploadingPhoto: vi.fn(),
    clearUploadingPhotos: vi.fn(),
    setSelectedSpot: (s: SpotEntity | null) => {
      stubState.selectedSpot = s
      setSelectedSpotMock(s)
    },
    setFilter: vi.fn(),
  }),
}))

function stubAmapGlobals(): void {
  // Use class declarations so `new AMapNS.Map(...)` and `new AMapNS.Marker(...)`
  // work in jsdom — vi.fn().mockImplementation(() => ({...})) loses
  // [[Construct]] semantics under vitest 4.x spy.
  class StubMap {
    add = vi.fn()
    remove = vi.fn()
    destroy = vi.fn()
    setFitView = vi.fn()
    setCenter = vi.fn()
    constructor(_id: string, _opts: any) {}
  }
  class StubMarker {
    on = vi.fn()
    hide = vi.fn()
    getPosition = vi.fn()
    constructor(_opts: any) {}
  }
  const convertFromImpl = vi.fn(
    (coords: number[], _type: string, cb: any) => {
      cb('complete', {
        info: 'ok',
        locations: [{ lng: coords[0] + 0.001, lat: coords[1] + 0.001 }],
      })
    },
  )
  ;(window as any).AMapLoader = {
    load: vi.fn().mockResolvedValue({
      Map: StubMap,
      Marker: StubMarker,
      Icon: vi.fn(),
      Size: vi.fn(),
      convertFrom: convertFromImpl,
      Geolocation: vi.fn(),
      plugin: vi.fn((_name: string, cb: any) => cb && cb()),
    }),
  }
  ;(window as any).AMap = {
    Map: StubMap,
    Marker: StubMarker,
    convertFrom: convertFromImpl,
  }
}

describe('Map.vue', () => {
  let originalLocation: any
  let geoMock: ReturnType<typeof vi.fn>

  beforeEach(() => {
    setActivePinia(createPinia())
    pushMock.mockClear()
    setSelectedSpotMock.mockClear()
    fetchSpotsMock.mockClear().mockResolvedValue(undefined)
    saveSpotMock.mockClear().mockResolvedValue(undefined)
    reverseGeocodeMock.mockClear().mockResolvedValue('测试地址')
    uploadSpotPhotoMock.mockClear().mockResolvedValue({ url: 'https://cdn/x.png', name: 'p.png' })
    stubState.selectedSpot = null
    geoMock = vi.fn((success: any) =>
      success({ coords: { longitude: 113.265, latitude: 23.129, accuracy: 10 } }),
    )
    originalLocation = window.location
    stubAmapGlobals()
    vi.stubGlobal('navigator', {
      ...navigator,
      geolocation: {
        getCurrentPosition: geoMock,
      },
    })
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    ;(window as any).location = originalLocation
    delete (window as any).AMapLoader
    delete (window as any).AMap
    delete (window as any)._AMapSecurityConfig
  })

  it('calls navigator.geolocation.getCurrentPosition when the capture FAB is tapped', async () => {
    const { default: MapComponent } = await import('@/view/Map.vue')
    const wrapper = mount(MapComponent, { global: { plugins: [createPinia()] } })
    await flushPromises()

    const fab = wrapper.find('[data-testid="capture-fab"]')
    expect(fab.exists()).toBe(true)
    await fab.trigger('click')

    expect(geoMock).toHaveBeenCalledTimes(1)
    wrapper.unmount()
  })

  // CR-2 order guard for T-13-FE-05: AMap.convertFrom MUST be invoked BEFORE
  // mapStore.reverseGeocode in the capture flow. A future refactor that
  // reorders or skips the conversion re-introduces the 300-500m NE offset
  // (Pitfall 1) and an inaccurate reverse-geocoded address.
  it('invokes AMap.convertFrom BEFORE mapStore.reverseGeocode in the capture flow (T-13-FE-05 guard)', async () => {
    const { default: MapComponent } = await import('@/view/Map.vue')
    const wrapper = mount(MapComponent, { global: { plugins: [createPinia()] } })
    await flushPromises()

    const convertFromSpy = vi.mocked((window as any).AMap.convertFrom)
    await wrapper.find('[data-testid="capture-fab"]').trigger('click')
    await flushPromises()

    expect(convertFromSpy).toHaveBeenCalled()
    expect(reverseGeocodeMock).toHaveBeenCalled()
    expect(reverseGeocodeMock).toHaveBeenCalledAfter(convertFromSpy)
    wrapper.unmount()
  })

  it('renders the address chip with the reverse-geocoded address after a successful capture flow', async () => {
    const { default: MapComponent } = await import('@/view/Map.vue')
    const wrapper = mount(MapComponent, { global: { plugins: [createPinia()] } })
    await flushPromises()

    await wrapper.find('[data-testid="capture-fab"]').trigger('click')
    await flushPromises()
    await wrapper.vm.$nextTick()
    await flushPromises()
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('测试地址')
    wrapper.unmount()
  })

  it('calls mapStore.setSelectedSpot(null) when closing an open spot detail bottom-sheet', async () => {
    stubState.selectedSpot = {
      id: '1',
      name: 'A',
      latitude: 23,
      longitude: 113,
      notes: '',
      tags: '',
      photos: [],
      address: '',
      createdAt: '',
      updatedAt: '',
    }
    const { default: MapComponent } = await import('@/view/Map.vue')
    const wrapper = mount(MapComponent, { global: { plugins: [createPinia()] } })
    await flushPromises()
    await wrapper.vm.$nextTick()

    expect(wrapper.find('[data-testid="spot-navigate-btn"]').exists()).toBe(true)
    const closeBtn = wrapper.findAll('button').find((b) => b.text() === '关闭')
    expect(closeBtn).toBeDefined()
    await closeBtn!.trigger('click')
    await wrapper.vm.$nextTick()

    expect(setSelectedSpotMock).toHaveBeenCalledWith(null)
    wrapper.unmount()
  })

  it('sets window.location.href to an androidamap:// scheme with dev=0 when the Navigate button is pressed', async () => {
    let capturedHref = ''
    Object.defineProperty(window, 'location', {
      configurable: true,
      writable: true,
      value: {
        ...originalLocation,
        get href() {
          return capturedHref
        },
        set href(v: string) {
          capturedHref = v
        },
        assign: vi.fn(),
        replace: vi.fn(),
        reload: vi.fn(),
      },
    })

    stubState.selectedSpot = {
      id: '1',
      name: 'A',
      latitude: 23.129,
      longitude: 113.265,
      notes: '',
      tags: '',
      photos: [],
      address: '',
      createdAt: '',
      updatedAt: '',
    }
    const { default: MapComponent } = await import('@/view/Map.vue')
    const wrapper = mount(MapComponent, { global: { plugins: [createPinia()] } })
    await flushPromises()
    await wrapper.vm.$nextTick()

    const navBtn = wrapper.find('[data-testid="spot-navigate-btn"]')
    expect(navBtn.exists()).toBe(true)
    await navBtn.trigger('click')
    await flushPromises()

    expect(capturedHref.startsWith('androidamap://navi') || capturedHref.startsWith('iosamap://navi')).toBe(true)
    expect(capturedHref).toContain('dev=0')
    wrapper.unmount()
  })

  it('handlePhotoChange with two File inputs calls mapStore.uploadSpotPhoto twice in parallel (D-16)', async () => {
    const { default: MapComponent } = await import('@/view/Map.vue')
    const wrapper = mount(MapComponent, { global: { plugins: [createPinia()] } })
    await flushPromises()

    await wrapper.find('[data-testid="capture-fab"]').trigger('click')
    await flushPromises()

    const expand = wrapper.findAll('button').find((b) => b.text() === '展开更多')
    if (expand) {
      await expand.trigger('click')
      await wrapper.vm.$nextTick()
    }

    const fileInput = wrapper.find('[data-testid="capture-photo-input"]')
    expect(fileInput.exists()).toBe(true)
    const f1 = new File(['a'], 'a.png', { type: 'image/png' })
    const f2 = new File(['b'], 'b.png', { type: 'image/png' })
    Object.defineProperty(fileInput.element, 'files', { value: [f1, f2] })

    await fileInput.trigger('change')
    await flushPromises()

    expect(uploadSpotPhotoMock).toHaveBeenCalledTimes(2)
    wrapper.unmount()
  })

  it('shows the empty-state tip when spots.length === 0 on mount', async () => {
    const { default: MapComponent } = await import('@/view/Map.vue')
    const wrapper = mount(MapComponent, { global: { plugins: [createPinia()] } })
    await flushPromises()

    expect(wrapper.text()).toContain('点右上按钮保存当前钓点')
    wrapper.unmount()
  })

  // ---- D-22 skip-list (7 tests re-routed to 13-VALIDATION.md Manual-Only) ----

  it.skipIf(
    true,
    'SKIP-pin-clustering — AMap.MarkerClusterer is an async plugin; cluster behavior is deferred (Manual-Only)',
    () => {
      throw new Error('MarkerClusterer not constructable in jsdom — manual UAT required')
    },
  )

  it.skipIf(
    true,
    'SKIP-touch-pinch-zoom — touch gestures need a real device (Manual-Only)',
    () => {
      throw new Error('Touch gestures not testable in jsdom — manual UAT required')
    },
  )

  it.skipIf(
    typeof navigator?.geolocation?.getCurrentPosition !== 'function' || !('permissions' in navigator),
    'SKIP-geolocation-denied — real denial flow needs an OS permission prompt + rejection (Manual-Only)',
    () => {
      throw new Error('Geolocation-denied fallback needs real device — manual UAT required')
    },
  )

  it.skipIf(
    true,
    'SKIP-real-gps-accuracy — real 10–30m GPS accuracy in the field (Manual-Only)',
    () => {
      throw new Error('Real GPS accuracy needs outdoor device — manual UAT required')
    },
  )

  it.skipIf(
    typeof window !== 'undefined' && !/iPhone|iPad|iPod|Android/.test(navigator.userAgent),
    'SKIP-gaode-deep-link — deep-link opens Gaode App only on a real device (Manual-Only)',
    () => {
      throw new Error('Gaode deep-link hand-off needs mobile device — manual UAT required')
    },
  )

  it.skipIf(
    true,
    'SKIP-real-tile-render — real AMap.Map DOM attachment needs a network + key (Manual-Only)',
    () => {
      throw new Error('Real tile render not testable in jsdom — manual UAT required')
    },
  )

  it.skipIf(
    typeof window !== 'undefined' && window.location?.protocol !== 'https:' && location.hostname !== 'localhost',
    'SKIP-https-geolocation — Geolocation API requires HTTPS in production (Manual-Only)',
    () => {
      throw new Error('Geolocation over HTTPS in production is network/origin-level only')
    },
  )
})