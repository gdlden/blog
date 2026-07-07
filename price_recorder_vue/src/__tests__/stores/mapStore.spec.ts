import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useMapStore } from '@/stores/mapStore'
import * as mapApi from '@/api/map'

vi.mock('vue-toastification', () => ({
  useToast: () => ({
    success: vi.fn(),
    error: vi.fn(),
  }),
}))

vi.mock('@/api/map', () => ({
  listSpots: vi.fn(),
  getSpot: vi.fn(),
  saveSpot: vi.fn(),
  updateSpot: vi.fn(),
  deleteSpot: vi.fn(),
  reverseGeocode: vi.fn(),
  uploadSpotPhoto: vi.fn(),
}))

describe('mapStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('fetchSpots populates spots from listSpots() reply', async () => {
    vi.mocked(mapApi.listSpots).mockResolvedValue({
      spots: [
        {
          id: '1',
          name: 'A',
          latitude: 23.1,
          longitude: 113.2,
          notes: '',
          tags: 'river,carp',
          photos: [],
          address: '广东省XX',
          createdAt: '',
          updatedAt: '',
        },
      ],
    })

    const store = useMapStore()
    await store.fetchSpots()

    expect(mapApi.listSpots).toHaveBeenCalledTimes(1)
    expect(store.spots).toHaveLength(1)
    expect(store.spots[0]?.name).toBe('A')
    expect(store.loading).toBe(false)
  })

  it('saveSpot calls api then re-fetches the spot list', async () => {
    vi.mocked(mapApi.saveSpot).mockResolvedValue({
      spot: {
        id: '2',
        name: 'B',
        latitude: 23.5,
        longitude: 113.5,
        notes: '',
        tags: '',
        photos: [],
        address: '',
        createdAt: '',
        updatedAt: '',
      },
    })
    vi.mocked(mapApi.listSpots).mockResolvedValue({ spots: [] })

    const store = useMapStore()
    await store.saveSpot({
      name: 'B',
      latitude: 23.5,
      longitude: 113.5,
      notes: '',
      tags: '',
      photos: [],
      address: '',
    })

    expect(mapApi.saveSpot).toHaveBeenCalledTimes(1)
    expect(mapApi.listSpots).toHaveBeenCalledTimes(1)
  })

  // CR-3 regression guard: deleteSpot replies with { success: bool }, not { flag }.
  // The store must accept the boolean from the api (the api layer casts it)
  // and local-mutates spots instead of re-fetching for snappy UX (mirror
  // fuelStore.deleteVehicle).
  it('deleteSpot filters spots locally when api returns success', async () => {
    vi.mocked(mapApi.deleteSpot).mockResolvedValue(true)
    vi.mocked(mapApi.listSpots).mockResolvedValue({ spots: [] })

    const store = useMapStore()
    store.$patch({
      spots: [
        {
          id: '5',
          name: 'old',
          latitude: 0,
          longitude: 0,
          notes: '',
          tags: '',
          photos: [],
          address: '',
          createdAt: '',
          updatedAt: '',
        },
      ],
    })

    await store.deleteSpot('5')

    expect(mapApi.deleteSpot).toHaveBeenCalledWith('5')
    // No re-fetch — local mutate mirror
    expect(mapApi.listSpots).not.toHaveBeenCalled()
    expect(store.spots.find((s) => s.id === '5')).toBeUndefined()
  })

  it('deleteSpot rethrows when api reports failure', async () => {
    vi.mocked(mapApi.deleteSpot).mockResolvedValue(false)
    const store = useMapStore()
    store.$patch({
      spots: [
        {
          id: '7',
          name: 'doomed',
          latitude: 0,
          longitude: 0,
          notes: '',
          tags: '',
          photos: [],
          address: '',
          createdAt: '',
          updatedAt: '',
        },
      ],
    })

    await expect(store.deleteSpot('7')).rejects.toBeTruthy()
    // spot must NOT be removed on failure
    expect(store.spots.find((s) => s.id === '7')).toBeDefined()
  })

  it('uploadSpotPhoto pushes an entry and replaces url on api success', async () => {
    vi.mocked(mapApi.uploadSpotPhoto).mockImplementation(async (_file, onProgress) => {
      // simulate a tiny progress tick
      onProgress?.({ loaded: 50, total: 100 })
      return { id: 'f1', url: 'https://cdn/x.png' }
    })

    const store = useMapStore()
    const file = new File(['data'], 'photo.png', { type: 'image/png' })

    const result = await store.uploadSpotPhoto(file)

    expect(result.url).toBe('https://cdn/x.png')
    expect(store.uploadingPhotos.length).toBe(1)
    expect(store.uploadingPhotos[0]?.url).toBe('https://cdn/x.png')
    expect(store.uploadingPhotos[0]?.progress).toBeGreaterThanOrEqual(0)
  })

  it('setFilter updates selectedTags and searchQuery; filteredSpots respects both', () => {
    const store = useMapStore()
    store.$patch({
      spots: [
        {
          id: '1',
          name: 'Lake Blue',
          latitude: 0,
          longitude: 0,
          notes: '',
          tags: 'river,carp',
          photos: [],
          address: '',
          createdAt: '',
          updatedAt: '',
        },
        {
          id: '2',
          name: 'Reed Pond',
          latitude: 0,
          longitude: 0,
          notes: '',
          tags: 'pond',
          photos: [],
          address: '',
          createdAt: '',
          updatedAt: '',
        },
      ],
    })

    store.setFilter({ tags: [], query: 'lake' })
    expect(store.filteredSpots.map((s) => s.id)).toEqual(['1'])

    store.setFilter({ tags: ['pond'], query: '' })
    expect(store.selectedTags).toEqual(['pond'])
    expect(store.filteredSpots.map((s) => s.id)).toEqual(['2'])

    store.setFilter({ tags: [], query: '' })
    expect(store.filteredSpots.length).toBe(2)
  })
})