import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { useToast } from 'vue-toastification'
import * as mapApi from '@/api/map'
import type { SpotEntity, CreateSpotRequest, UpdateSpotRequest } from '@/api/map'

export const useMapStore = defineStore('map', () => {
  const toast = useToast()

  const spots = ref<SpotEntity[]>([])
  const selectedSpot = ref<SpotEntity | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const uploadingPhotos = ref<Array<{ url: string; progress: number; name: string }>>([])
  const selectedTags = ref<string[]>([])
  const searchQuery = ref<string>('')

  const spotCount = computed(() => spots.value.length)
  const allTags = computed(() => {
    const tagSet = new Set<string>()
    for (const spot of spots.value) {
      const list = (spot.tags || '')
        .split(',')
        .map((t) => t.trim())
        .filter(Boolean)
      for (const t of list) tagSet.add(t)
    }
    return Array.from(tagSet)
  })
  const filteredSpots = computed(() => {
    const query = searchQuery.value.trim().toLowerCase()
    return spots.value.filter((spot) => {
      if (selectedTags.value.length > 0) {
        const spotTagList = (spot.tags || '')
          .split(',')
          .map((t) => t.trim())
          .filter(Boolean)
        const matchesTag = selectedTags.value.some((t) => spotTagList.includes(t))
        if (!matchesTag) return false
      }
      if (query && !String(spot.name || '').toLowerCase().includes(query)) return false
      return true
    })
  })

  async function fetchSpots(): Promise<void> {
    loading.value = true
    error.value = null
    try {
      const reply = await mapApi.listSpots()
      spots.value = reply.spots || []
    } catch (err: any) {
      error.value = err?.message || '获取钓点列表失败'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function saveSpot(data: CreateSpotRequest): Promise<void> {
    loading.value = true
    try {
      await mapApi.saveSpot(data)
      toast.success('位置已保存')
      await fetchSpots()
    } catch (err: any) {
      toast.error(err?.message || '保存失败')
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateSpot(data: UpdateSpotRequest): Promise<void> {
    loading.value = true
    try {
      const reply = await mapApi.updateSpot(data)
      toast.success('位置已更新')
      await fetchSpots()
      if (reply.spot && selectedSpot.value?.id === data.id) selectedSpot.value = reply.spot
    } catch (err: any) {
      toast.error(err?.message || '更新失败')
      throw err
    } finally {
      loading.value = false
    }
  }

  // CR-3: mapApi.deleteSpot returns Promise<boolean> (DeleteSpotReply is
  // { success: bool }, not fuel's { flag }). Re-fetch is replaced by a local
  // filter for snappy UX, mirroring fuelStore.deleteVehicle.
  async function deleteSpot(id: string): Promise<void> {
    loading.value = true
    try {
      const ok = await mapApi.deleteSpot(id)
      if (!ok) throw new Error('删除失败')
      toast.success('位置已删除')
      spots.value = spots.value.filter((s) => s.id !== id)
      if (selectedSpot.value?.id === id) selectedSpot.value = null
    } catch (err: any) {
      toast.error(err?.message || '删除失败')
      throw err
    } finally {
      loading.value = false
    }
  }

  async function getSpot(id: string): Promise<SpotEntity> {
    return (await mapApi.getSpot(id)).spot
  }

  // CR-2: caller (Map.vue) MUST pass already-converted GCJ-02 coords; the
  // store is coordinate-agnostic. Backend biz/map.go has no WGS-84→GCJ-02.
  async function reverseGeocode(input: {
    latitude: number
    longitude: number
  }): Promise<string> {
    try {
      return (await mapApi.reverseGeocode(input)).address
    } catch (err: any) {
      toast.error(err?.message || '逆地理失败')
      throw err
    }
  }

  // D-16: triggered on file selection; the returned URL is appended into the
  // spot draft's photos[] by Map.vue at submit time, not here.
  async function uploadSpotPhoto(file: File): Promise<{ url: string; name: string }> {
    const entry = { name: file.name, progress: 0, url: '' }
    uploadingPhotos.value.push(entry)
    try {
      const reply = await mapApi.uploadSpotPhoto(file, (e: any) => {
        if (e && typeof e.total === 'number' && e.total > 0) {
          entry.progress = Math.round((e.loaded / e.total) * 100)
        }
      })
      entry.url = reply.url
      return { url: reply.url, name: file.name }
    } catch (err: any) {
      const idx = uploadingPhotos.value.indexOf(entry)
      if (idx >= 0) uploadingPhotos.value.splice(idx, 1)
      toast.error(err?.message || `上传失败：${file.name}`)
      throw err
    }
  }

  function removeUploadingPhoto(entry: { url: string; progress: number; name: string }) {
    const idx = uploadingPhotos.value.indexOf(entry)
    if (idx >= 0) uploadingPhotos.value.splice(idx, 1)
  }

  function clearUploadingPhotos() {
    uploadingPhotos.value = []
  }

  function setSelectedSpot(spot: SpotEntity | null): void {
    selectedSpot.value = spot
  }

  function setFilter(payload: { tags?: string[]; query?: string }): void {
    if (payload.tags !== undefined) selectedTags.value = payload.tags
    if (payload.query !== undefined) searchQuery.value = payload.query
  }

  return {
    spots,
    selectedSpot,
    loading,
    error,
    uploadingPhotos,
    selectedTags,
    searchQuery,
    spotCount,
    allTags,
    filteredSpots,
    fetchSpots,
    saveSpot,
    updateSpot,
    deleteSpot,
    getSpot,
    reverseGeocode,
    uploadSpotPhoto,
    removeUploadingPhoto,
    clearUploadingPhotos,
    setSelectedSpot,
    setFilter,
  }
})