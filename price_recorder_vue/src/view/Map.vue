<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useMapStore } from '@/stores/mapStore'
import { useToast } from 'vue-toastification'
import type { CreateSpotRequest, SpotEntity, UpdateSpotRequest } from '@/api/map'

// PATTERNS risk-flag #5: window._AMapSecurityConfig MUST be set BEFORE
// AMapLoader.load. Set at the top of module scope (NOT inside onMounted)
// so the security check runs against the configured code. Guard with
// non-empty check so unit tests (no env var) skip assignment without error.
const securityCode = import.meta.env.VITE_GAODE_JS_SECURITY_CODE
if (securityCode && typeof window !== 'undefined') {
  ;(window as any)._AMapSecurityConfig = { securityJsCode: securityCode }
}

const mapStore = useMapStore()
const toast = useToast()

const mapInstance = ref<any>(null)
const markers = ref<Map<string, any>>(new Map())
let AMapNS: any = null

const captureOpen = ref(false)
const editingSpot = ref<SpotEntity | null>(null)
const isSubmitting = ref(false)
const showExpandedForm = ref(false)
const filterOpen = ref(false)
const capturedAddress = ref('')
const capturedWgs84 = ref<{ lng: number; lat: number } | null>(null)
const capturedGcj02 = ref<{ lng: number; lat: number } | null>(null)
const showEmptyTip = ref(true)

const name = ref('')
const notes = ref('')
const tags = ref('')
const photoPreviews = ref<Array<{ url: string; name: string; progress: number }>>([])

const mapContainerId = 'map-container'
const GUANGZHOU_GCJ02: [number, number] = [113.264385, 23.129112]

const selectedSpot = computed(() => mapStore.selectedSpot)
const searchQuery = computed({
  get: () => mapStore.searchQuery,
  set: (v: string) => mapStore.setFilter({ query: v }),
})
const selectedTags = computed({
  get: () => mapStore.selectedTags,
  set: (v: string[]) => mapStore.setFilter({ tags: v }),
})

// CR-2 / RESEARCH Pattern 5: WGS-84 → GCJ-02 conversion goes through
// AMap.convertFrom; reject on status !== 'complete' or info !== 'ok'.
// Every marker creation path AND every reverse-geocode-display call MUST
// go through this. SpotEntity persists RAW WGS-84 (display concern only).
function convertWgs84ToGcj02(lng: number, lat: number): Promise<{ lng: number; lat: number }> {
  return new Promise((resolve, reject) => {
    if (!AMapNS || typeof AMapNS.convertFrom !== 'function') {
      reject(new Error('AMap.convertFrom unavailable'))
      return
    }
    AMapNS.convertFrom([lng, lat], 'gps', (status: string, result: any) => {
      if (status === 'complete' && result?.info === 'ok') {
        const loc = result.locations[0]
        resolve({ lng: loc.lng, lat: loc.lat })
      } else {
        reject(new Error('Coordinate conversion failed'))
      }
    })
  })
}

// Markers are placed at GCJ-02 (CR-2); spot data stores RAW WGS-84.
async function addSpotMarker(spot: SpotEntity): Promise<void> {
  if (!mapInstance.value || !AMapNS) return
  const gcj = await convertWgs84ToGcj02(spot.longitude, spot.latitude)
  const marker = new AMapNS.Marker({
    position: [gcj.lng, gcj.lat],
    title: spot.name,
  })
  marker.on('click', () => {
    mapStore.setSelectedSpot(spot)
  })
  mapInstance.value.add(marker)
  markers.value.set(spot.id, marker)
}

function applyFilterVisibility(): void {
  if (!mapInstance.value) return
  const visibleIds = new Set(mapStore.filteredSpots.map((s) => s.id))
  markers.value.forEach((marker, id) => {
    if (typeof marker.hide === 'function') {
      marker.hide(!visibleIds.has(id))
    }
  })
}

function fitView(): void {
  if (mapInstance.value && typeof mapInstance.value.setFitView === 'function') {
    mapInstance.value.setFitView()
  }
}

function toggleTag(tag: string): void {
  const list = selectedTags.value.slice()
  const idx = list.indexOf(tag)
  if (idx >= 0) list.splice(idx, 1)
  else list.push(tag)
  selectedTags.value = list
  applyFilterVisibility()
}

function onSearchInput(): void {
  applyFilterVisibility()
}

// D-20: empty-state fallback chain on mount — AMap.Geolocation plugin
// first; on failure, fall back to last-known spot's GCJ-02 position or
// the static Guangzhou GCJ-02 default (no conversion for this fallback).
async function resolveInitialCenter(): Promise<[number, number]> {
  if (AMapNS && typeof AMapNS.Geolocation === 'function') {
    try {
      const geo = await new Promise<[number, number]>((resolve, reject) => {
        const Geolocation = AMapNS.Geolocation
        const inst = new Geolocation({
          enableHighAccuracy: true,
          timeout: 10000,
        })
        inst.getCurrentPosition((status: string, result: any) => {
          if (status === 'complete' && result?.position) {
            convertWgs84ToGcj02(result.position.lng, result.position.lat)
              .then((gcj) => resolve([gcj.lng, gcj.lat]))
              .catch(() => reject(new Error('convert failed')))
          } else {
            reject(new Error('geo failed'))
          }
        })
      })
      return geo
    } catch {
      // fall through to last-known / default
    }
  }
  if (mapStore.spots.length > 0) {
    const last = mapStore.spots[mapStore.spots.length - 1]
    if (last) {
      try {
        const gcj = await convertWgs84ToGcj02(last.longitude, last.latitude)
        return [gcj.lng, gcj.lat]
      } catch {
        // fall through
      }
    }
  }
  return GUANGZHOU_GCJ02
}

onMounted(async () => {
  const loader = (window as any).AMapLoader
  if (!loader || typeof loader.load !== 'function') {
    toast.error('AMap 未加载')
    showEmptyTip.value = mapStore.spots.length === 0
    return
  }
  try {
    AMapNS = await loader.load({
      key: import.meta.env.VITE_GAODE_JS_API_KEY,
      version: '2.0',
      plugins: ['AMap.Geolocation', 'AMap.MarkerClusterer'],
    })
  } catch (err: any) {
    toast.error(err?.message || 'AMap 加载失败')
    showEmptyTip.value = mapStore.spots.length === 0
    return
  }
  if (!AMapNS || typeof AMapNS.Map !== 'function') {
    showEmptyTip.value = mapStore.spots.length === 0
    return
  }
  const center = await resolveInitialCenter()
  mapInstance.value = new AMapNS.Map(mapContainerId, {
    zoom: 13,
    center,
  })
  await mapStore.fetchSpots()
  for (const spot of mapStore.spots) {
    try {
      await addSpotMarker(spot)
    } catch {
      // skip a single bad marker; conversion may fail in test env
    }
  }
  showEmptyTip.value = mapStore.spots.length === 0
})

onUnmounted(() => {
  mapInstance.value?.destroy?.()
  mapInstance.value = null
  AMapNS = null
})

// Capture FAB handler (D-07, D-15) — getCurrentPosition captures RAW WGS-84,
// which we convert to GCJ-02 BEFORE calling reverseGeocode (CR-2: backend
// biz/map.go has NO conversion). Spot is persisted with RAW WGS-84 lat/lng.
function handleCaptureFab(): void {
  showEmptyTip.value = false
  if (!navigator?.geolocation?.getCurrentPosition) {
    toast.error('无法获取位置')
    openCaptureSheet(null)
    return
  }
  navigator.geolocation.getCurrentPosition(
    async (pos) => {
      capturedWgs84.value = {
        lng: pos.coords.longitude,
        lat: pos.coords.latitude,
      }
      try {
        capturedGcj02.value = await convertWgs84ToGcj02(
          pos.coords.longitude,
          pos.coords.latitude,
        )
        const address = await mapStore.reverseGeocode({
          latitude: capturedGcj02.value.lat,
          longitude: capturedGcj02.value.lng,
        })
        capturedAddress.value = address || ''
      } catch {
        capturedAddress.value = ''
      }
      openCaptureSheet(null)
    },
    () => {
      toast.error('无法获取位置')
      // D-20 geolocation-denied: surface manual UAT path; center remains.
    },
    { enableHighAccuracy: true, timeout: 10000 },
  )
}

function openCaptureSheet(spot: SpotEntity | null): void {
  editingSpot.value = spot
  if (spot) {
    name.value = spot.name
    notes.value = spot.notes || ''
    tags.value = spot.tags || ''
    photoPreviews.value = (spot.photos || []).map((url) => ({
      url,
      name: url.split('/').pop() || url,
      progress: 100,
    }))
    capturedAddress.value = spot.address || ''
    if (spot.latitude && spot.longitude) {
      capturedWgs84.value = { lng: spot.longitude, lat: spot.latitude }
    } else {
      capturedWgs84.value = null
    }
    capturedGcj02.value = null
    showExpandedForm.value = true
  } else {
    name.value = ''
    notes.value = ''
    tags.value = ''
    photoPreviews.value = []
    capturedAddress.value = ''
    showExpandedForm.value = false
  }
  captureOpen.value = true
}

function closeCaptureSheet(): void {
  captureOpen.value = false
  editingSpot.value = null
}

// D-16: photo picker uploads in parallel on selection (NOT on submit);
// URL list is read into CreateSpotRequest.photos at submit time.
async function handlePhotoChange(event: Event): Promise<void> {
  const input = event.target as HTMLInputElement
  if (!input.files || input.files.length === 0) return
  const files = Array.from(input.files)
  try {
    await Promise.allSettled(
      files.map(async (file) => {
        const preview = { url: '', name: file.name, progress: 0 }
        photoPreviews.value.push(preview)
        try {
          const result = await mapStore.uploadSpotPhoto(file)
          preview.url = result.url
          preview.progress = 100
        } catch {
          const idx = photoPreviews.value.indexOf(preview)
          if (idx >= 0) photoPreviews.value.splice(idx, 1)
        }
      }),
    )
  } finally {
    input.value = ''
  }
}

function removePhotoAt(index: number): void {
  photoPreviews.value.splice(index, 1)
}

// D-09/D-17: only name is required; CR-2: persist RAW WGS-84 lat/lng.
async function handleSubmitCaptureForm(): Promise<void> {
  if (!name.value.trim()) return
  if (!capturedWgs84.value && !editingSpot.value) return
  isSubmitting.value = true
  try {
    const photos = photoPreviews.value.map((p) => p.url).filter(Boolean)
    if (editingSpot.value) {
      const payload: UpdateSpotRequest = {
        id: editingSpot.value.id,
        name: name.value.trim(),
        latitude: editingSpot.value.latitude,
        longitude: editingSpot.value.longitude,
        notes: notes.value,
        tags: tags.value,
        photos,
        address: capturedAddress.value,
      }
      await mapStore.updateSpot(payload)
    } else {
      const lat = capturedWgs84.value!.lat
      const lng = capturedWgs84.value!.lng
      const payload: CreateSpotRequest = {
        name: name.value.trim(),
        latitude: lat,
        longitude: lng,
        notes: notes.value,
        tags: tags.value,
        photos,
        address: capturedAddress.value,
      }
      await mapStore.saveSpot(payload)
      // refresh markers — re-fetch already happened in the store; re-render
      const last = mapStore.spots[mapStore.spots.length - 1]
      if (last && last.id && !markers.value.has(last.id)) {
        await addSpotMarker(last)
      }
      showEmptyTip.value = mapStore.spots.length === 0
    }
    captureOpen.value = false
    editingSpot.value = null
  } catch (err: any) {
    // toast already fired by the store
  } finally {
    isSubmitting.value = false
  }
}

async function handleDeleteSpot(): Promise<void> {
  const spot = mapStore.selectedSpot
  if (!spot) return
  if (!window.confirm('确定删除此钓点？')) return
  try {
    await mapStore.deleteSpot(spot.id)
    const marker = markers.value.get(spot.id)
    if (marker && mapInstance.value) {
      mapInstance.value.remove(marker)
    }
    markers.value.delete(spot.id)
    mapStore.setSelectedSpot(null)
  } catch {
    // toast already fired
  }
}

function startEditSpot(): void {
  const spot = mapStore.selectedSpot
  if (!spot) return
  openCaptureSheet(spot)
  mapStore.setSelectedSpot(null)
}

// CR-2 / RESEARCH Pattern 6 — frontend already converted to GCJ-02, so
// dev=0 in the deep-link. iosamap:// on iOS UA, androidamap:// otherwise,
// with a 2s setTimeout fallback to the public uri.amap.com navigation page.
async function navigateToSpot(spot: SpotEntity): Promise<void> {
  try {
    const gcj = await convertWgs84ToGcj02(spot.longitude, spot.latitude)
    const isIOS = /iPhone|iPad|iPod/.test(navigator.userAgent)
    const scheme = isIOS ? 'iosamap://' : 'androidamap://'
    const uri = `${scheme}navi?sourceApplication=blog&poiname=${encodeURIComponent(
      spot.name,
    )}&lat=${gcj.lat}&lon=${gcj.lng}&dev=0&style=2`
    window.location.href = uri
    setTimeout(() => {
      if (document.hasFocus()) {
        window.open(
          `https://uri.amap.com/navigation?to=${gcj.lng},${gcj.lat},${encodeURIComponent(
            spot.name,
          )}`,
          '_blank',
        )
      }
    }, 2000)
  } catch {
    toast.error('导航失败')
  }
}

function closeSpotDetail(): void {
  mapStore.setSelectedSpot(null)
}

function toggleFilter(): void {
  filterOpen.value = !filterOpen.value
  if (filterOpen.value) showEmptyTip.value = false
}
</script>

<template>
  <div class="relative w-full" style="height: calc(100vh - 56px)">
    <div :id="mapContainerId" class="absolute inset-0 w-full h-full" />

    <!-- Filter FAB (top-right) — D-18 -->
    <button
      data-testid="filter-fab"
      class="absolute top-4 right-4 z-30 w-11 h-11 rounded-full shadow-lg bg-white/90 hover:bg-white text-gray-700 flex items-center justify-center"
      @click="toggleFilter"
      aria-label="筛选"
    >
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M3 4h18M6 12h12M10 20h4"
        />
      </svg>
    </button>

    <!-- Capture FAB (bottom edge) — D-07, D-15. Separate from filter FAB. -->
    <button
      data-testid="capture-fab"
      class="absolute bottom-6 right-4 z-30 w-14 h-14 rounded-full shadow-xl bg-blue-500 hover:bg-blue-600 text-white flex items-center justify-center"
      @click="handleCaptureFab"
      aria-label="保存当前钓点"
    >
      <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2.5"
          d="M12 4v16M4 12h16"
        />
      </svg>
    </button>

    <!-- Empty-state tooltip — D-20 -->
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0 -translate-y-1"
      enter-to-class="opacity-100 translate-y-0"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100 translate-y-0"
      leave-to-class="opacity-0 -translate-y-1"
    >
      <div
        v-if="showEmptyTip && mapStore.spots.length === 0"
        class="absolute top-20 right-4 z-20 bg-white/95 shadow-lg rounded-xl px-3 py-2 text-[12px] text-gray-700 max-w-[200px]"
      >
        点右上按钮保存当前钓点
      </div>
    </Transition>

    <!-- Filter & search overlay — D-18 slide-down -->
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0 -translate-y-full"
      enter-to-class="opacity-100 translate-y-0"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100 translate-y-0"
      leave-to-class="opacity-0 -translate-y-full"
    >
      <div
        v-if="filterOpen"
        class="absolute top-16 right-4 left-4 z-40 bg-white rounded-2xl shadow-2xl p-4 space-y-3"
      >
        <input
          v-model="searchQuery"
          data-testid="filter-search"
          type="text"
          placeholder="按名称搜索"
          class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-400"
          @input="onSearchInput"
        />
        <div v-if="mapStore.allTags.length > 0" class="flex flex-wrap gap-2">
          <button
            v-for="tag in mapStore.allTags"
            :key="tag"
            data-testid="filter-tag-chip"
            class="px-3 py-1 rounded-full text-xs border transition-colors"
            :class="
              selectedTags.includes(tag)
                ? 'bg-blue-500 text-white border-blue-500'
                : 'bg-white text-gray-700 border-gray-300 hover:border-blue-400'
            "
            @click="toggleTag(tag)"
          >
            {{ tag }}
          </button>
        </div>
        <div class="flex justify-between items-center pt-2 border-t border-gray-100">
          <button
            class="text-xs text-gray-500 hover:text-gray-700"
            @click="filterOpen = false"
          >
            关闭
          </button>
          <button
            data-testid="fit-view-btn"
            class="text-xs px-3 py-1 rounded bg-gray-100 hover:bg-gray-200 text-gray-700"
            @click="fitView"
          >
            适配视图
          </button>
        </div>
      </div>
    </Transition>

    <!-- Spot detail bottom-sheet (D-10) -->
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="selectedSpot"
        class="fixed inset-0 z-50 flex items-end justify-center p-0"
        style="background-color: rgba(0, 0, 0, 0.35)"
        @click.self="closeSpotDetail"
      >
        <Transition
          enter-active-class="transition duration-300 ease-out"
          enter-from-class="opacity-0 translate-y-full"
          enter-to-class="opacity-100 translate-y-0"
          leave-active-class="transition duration-200 ease-in"
          leave-from-class="opacity-100 translate-y-0"
          leave-to-class="opacity-0 translate-y-full"
          appear
        >
          <div
            v-if="selectedSpot"
            class="w-full max-w-lg bg-white rounded-t-2xl shadow-2xl overflow-hidden p-4 space-y-3"
          >
            <div class="flex items-start justify-between">
              <div>
                <h3 class="text-base font-semibold text-gray-900">
                  {{ selectedSpot.name }}
                </h3>
                <p v-if="selectedSpot.address" class="text-xs text-gray-500 mt-1">
                  {{ selectedSpot.address }}
                </p>
              </div>
              <button
                class="text-gray-400 hover:text-gray-600 text-sm"
                @click="closeSpotDetail"
              >
                关闭
              </button>
            </div>

            <div v-if="selectedSpot.tags" class="flex flex-wrap gap-1.5">
              <span
                v-for="tag in selectedSpot.tags.split(',').map((t) => t.trim()).filter(Boolean)"
                :key="tag"
                class="px-2 py-0.5 rounded bg-blue-50 text-blue-700 text-[11px]"
              >
                {{ tag }}
              </span>
            </div>

            <p v-if="selectedSpot.notes" class="text-sm text-gray-700 whitespace-pre-wrap">
              {{ selectedSpot.notes }}
            </p>

            <div v-if="(selectedSpot.photos || []).length > 0" class="flex gap-2 overflow-x-auto">
              <img
                v-for="(url, idx) in selectedSpot.photos"
                :key="idx"
                :src="url"
                :alt="`photo-${idx}`"
                class="w-16 h-16 object-cover rounded-lg flex-shrink-0"
              />
            </div>

            <div class="flex gap-2 pt-2 border-t border-gray-100">
              <button
                data-testid="spot-edit-btn"
                class="flex-1 px-3 py-2 rounded-lg bg-gray-100 hover:bg-gray-200 text-gray-700 text-sm font-medium"
                @click="startEditSpot"
              >
                编辑
              </button>
              <button
                data-testid="spot-delete-btn"
                class="flex-1 px-3 py-2 rounded-lg bg-red-50 hover:bg-red-100 text-red-600 text-sm font-medium"
                @click="handleDeleteSpot"
              >
                删除
              </button>
              <button
                data-testid="spot-navigate-btn"
                class="flex-1 px-3 py-2 rounded-lg bg-blue-500 hover:bg-blue-600 text-white text-sm font-medium"
                @click="navigateToSpot(selectedSpot)"
              >
                导航
              </button>
            </div>
          </div>
        </Transition>
      </div>
    </Transition>

    <!-- Capture bottom-sheet (D-15, D-17) -->
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="captureOpen"
        class="fixed inset-0 z-50 flex items-end justify-center p-0"
        style="background-color: rgba(0, 0, 0, 0.35)"
        @click.self="closeCaptureSheet"
      >
        <Transition
          enter-active-class="transition duration-300 ease-out"
          enter-from-class="opacity-0 translate-y-full"
          enter-to-class="opacity-100 translate-y-0"
          leave-active-class="transition duration-200 ease-in"
          leave-from-class="opacity-100 translate-y-0"
          leave-to-class="opacity-0 translate-y-full"
          appear
        >
          <div
            v-if="captureOpen"
            class="w-full max-w-lg bg-white rounded-t-2xl shadow-2xl overflow-hidden p-4 space-y-3"
          >
            <div class="flex items-center justify-between">
              <h3 class="text-base font-semibold text-gray-900">
                {{ editingSpot ? '编辑钓点' : '保存当前钓点' }}
              </h3>
              <button
                class="text-gray-400 hover:text-gray-600 text-sm"
                @click="closeCaptureSheet"
              >
                关闭
              </button>
            </div>

            <input
              v-model="name"
              data-testid="capture-name-input"
              type="text"
              placeholder="名称（必填）"
              class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-400"
            />

            <div v-if="capturedAddress" class="text-xs text-gray-500">
              <span
                class="inline-block px-2 py-1 rounded-full bg-gray-100 text-gray-700 text-[11px]"
              >
                {{ capturedAddress }}
              </span>
            </div>

            <button
              type="button"
              class="text-xs text-blue-600 hover:text-blue-800"
              @click="showExpandedForm = !showExpandedForm"
            >
              {{ showExpandedForm ? '收起' : '展开更多' }}
            </button>

            <div v-if="showExpandedForm" class="space-y-3">
              <textarea
                v-model="notes"
                data-testid="capture-notes-input"
                placeholder="备注（可选）"
                rows="2"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-400"
              ></textarea>

              <input
                v-model="tags"
                data-testid="capture-tags-input"
                type="text"
                placeholder="标签（逗号分隔，可选）"
                class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-400"
              />

              <div class="space-y-2">
                <label class="text-xs text-gray-600">
                  照片
                  <input
                    type="file"
                    multiple
                    accept="image/*"
                    data-testid="capture-photo-input"
                    class="block mt-1 text-xs"
                    @change="handlePhotoChange"
                  />
                </label>
                <div v-if="photoPreviews.length > 0" class="flex gap-2 flex-wrap">
                  <div
                    v-for="(p, idx) in photoPreviews"
                    :key="idx"
                    class="relative w-16 h-16"
                  >
                    <img
                      v-if="p.url"
                      :src="p.url"
                      :alt="p.name"
                      class="w-16 h-16 object-cover rounded-lg"
                    />
                    <div
                      v-else
                      class="w-16 h-16 rounded-lg bg-gray-100 flex items-center justify-center text-[10px] text-gray-500"
                    >
                      {{ p.progress }}%
                    </div>
                    <button
                      class="absolute -top-1 -right-1 w-4 h-4 rounded-full bg-red-500 text-white text-[10px] flex items-center justify-center"
                      @click="removePhotoAt(idx)"
                    >
                      x
                    </button>
                  </div>
                </div>
              </div>
            </div>

            <button
              data-testid="capture-submit-btn"
              :disabled="isSubmitting || !name.trim()"
              class="w-full px-3 py-2.5 rounded-lg bg-blue-500 text-white text-sm font-medium hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
              @click="handleSubmitCaptureForm"
            >
              {{ isSubmitting ? '保存中…' : '保存' }}
            </button>
          </div>
        </Transition>
      </div>
    </Transition>
  </div>
</template>