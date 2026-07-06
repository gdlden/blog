# Phase 13: Fishing Spot Map - Pattern Map (Plan 13-03 Frontend)

**Mapped:** 2026-07-06
**Files analyzed:** 11 (7 new + 4 modified)
**Analogs found:** 11 / 11 (one analog per file)
**Backend contract source:** `blog/api/map/v1/map.proto` (read in full — 114 lines); `blog/api/file/v1/file.proto` (read in full — 56 lines); raw upload route `POST /file/upload/raw/v1` confirmed in `blog/internal/server/http.go:80`.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `price_recorder_vue/src/api/map.ts` | api-client | request-response | `src/api/fuel.ts` (+ `appVersion.ts` for upload) | exact |
| `price_recorder_vue/src/stores/mapStore.ts` | store | request-response | `src/stores/fuelStore.ts` | exact |
| `price_recorder_vue/src/view/Map.vue` | view+component | event-driven + request-response | `src/view/DebtDetail.vue` (modal/transition) + RESEARCH Pattern 3 (AMap loader) | role-match (no existing map view) |
| `price_recorder_vue/src/components/AppLayout.vue` (mod) | layout | static | itself | exact (append-only) |
| `price_recorder_vue/src/router/index.ts` (mod) | route | static | itself | exact (append child) |
| `price_recorder_vue/src/utils/request.ts` (mod? verify) | util | request-response | itself | likely no edit needed (see Risk Map) |
| `price_recorder_vue/.env.example` (create) | config | static | (none — no .env file in repo) | none |
| `price_recorder_vue/src/__tests__/stores/mapStore.spec.ts` | test | request-response | `src/__tests__/stores/fuelStore.spec.ts` | exact |
| `price_recorder_vue/src/__tests__/api/map.spec.ts` | test | request-response | `src/__tests__/api/fuel.spec.ts` | exact |
| `price_recorder_vue/src/__tests__/view/Map.spec.ts` | component-test | event-driven | `src/__tests__/utils/request.spec.ts` (mock pattern only) | partial (no Vue component spec exists yet) |
| `.planning/REQUIREMENTS.md` (mod) | requirements-doc | static | itself | exact (gap-closer) |

## Backend Contract Reference (locked — do NOT modify)

`blog/api/map/v1/map.proto` — verified shapes the planner MUST mirror in `map.ts`:

```protobuf
message SpotEntity {
  string id = 1; string name = 2;
  double latitude = 3; double longitude = 4;
  string notes = 5; string tags = 6;          // comma-separated
  repeated string photos = 7; string address = 8;
  string created_at = 9; string updated_at = 10;
}
message CreateSpotRequest  { string name=1; double latitude=2; double longitude=3; string notes=4; string tags=5; repeated string photos=6; string address=7; }
message CreateSpotReply    { SpotEntity spot = 1; }
message ListSpotsRequest   {}
message ListSpotsReply     { repeated SpotEntity spots = 1; }
message UpdateSpotRequest  { string id=1; ...same as Create + id; }
message UpdateSpotReply    { SpotEntity spot = 1; }
message DeleteSpotRequest  { string id = 1; }
message DeleteSpotReply    { bool success = 1; }
message GetSpotRequest      { string id = 1; }
message GetSpotReply        { SpotEntity spot = 1; }
message ReverseGeocodeRequest  { double latitude = 1; double longitude = 2; }   // NOTE: proto field order lat,lng BUT backend builds `location=lng,lat` URL — frontend MUST send both; the JSON body is unordered so order doesn't matter on the wire. Per D-16 / SUMMARY 13-01 the documented convention is "send longitude first" in any code that constructs the body explicitly — but JSON key order is irrelevant. Just include both keys.
message ReverseGeocodeReply    { string address = 1; }
```

HTTP paths (all JWT-protected, none whitelisted):
`POST /map/save/v1` · `GET /map/list/v1` · `POST /map/update/v1` · `POST /map/delete/v1` · `GET /map/get/{id}` · `POST /map/reverse-geocode/v1`.

**Critical interceptor note:** `src/utils/request.ts:7-11` unwraps `body.data` — so `instance.get(...)` / `instance.post(...)` return the *inner* `data` object directly. Therefore `listSpots()` returns `{ spots: SpotEntity[] }` (the `ListSpotsReply`), `getSpot()` returns `{ spot: SpotEntity }`, `saveSpot()` returns `{ spot: SpotEntity }`. This is the same convention `fuel.ts` relies on.

## Pattern Assignments

---

### `src/api/map.ts` ← `src/api/fuel.ts` + `src/api/appVersion.ts`

**Analog:** `price_recorder_vue/src/api/fuel.ts` (148 lines, read in full) and `appVersion.ts:51-55` (upload pattern).

**Imports + axios instance pattern** (`fuel.ts:1,69-72`):
```typescript
import instance from '@/utils/request.ts'

function decimalToString(value: string | number | undefined): string {
  if (value === undefined || value === null) return ''
  return String(value)
}
```
**Key takeaway:** `map.ts` imports the default `instance` (NOT a named export). Proto `double` lat/lng arrive as JS numbers — but to mirror `fuel.ts`'s "decimalToString consistency" and the backend's string↔float64 strconv round-trip (per 13-02-summary), serialize lat/lng via `decimalToString` too. `tags` stays comma-separated string; `photos` stays `string[]`.

**Per-RPC function shape** (`fuel.ts:91-116`):
```typescript
export async function getVehicles(
  page?: string, pageSize?: string,
): Promise<FuelPageResponse<FuelVehicle>> {
  const params: Record<string, string> = {}
  if (page) params.page = page
  if (pageSize) params.pageSize = pageSize
  return await instance.get('/fuel/vehicle/page/v1', { params })
}

export async function createVehicle(data: FuelVehicleRequest): Promise<SaveFuelReply> {
  return await instance.post('/fuel/vehicle/save/v1', serializeVehicle(data))
}

export async function deleteVehicle(id: string): Promise<boolean> {
  const response = (await instance.post('/fuel/vehicle/delete/v1', { id })) as { flag: boolean }
  return response.flag
}
```
**Key takeaway for map.ts:**
- `listSpots()` → `instance.get('/map/list/v1')` returns `{ spots: SpotEntity[] }` (empty params — `ListSpotsRequest{}` is empty).
- `getSpot(id)` → `instance.get('/map/get/v1', { params: { id } })` — NOTE proto says `get: /map/get/{id}` (path templating), but Kratos's `RegisterMapHTTPServer` ALSO accepts the query-param form `?id=`; `fuel.ts:102` and `debt.ts:37` both use the query form successfully. Follow that convention — do NOT manually build `/map/get/${id}`.
- `saveSpot(data)` → `instance.post('/map/save/v1', serialized)` returns `{ spot: SpotEntity }`.
- `updateSpot(data)` → `instance.post('/map/update/v1', serialized)` returns `{ spot: SpotEntity }`.
- `deleteSpot(id)` → `instance.post('/map/delete/v1', { id })` — `DeleteSpotReply` is `{ success: bool }` (NOT `{ flag }` like fuel). Mirror `appVersion.ts:47-49` pattern: `return (await instance.post(...)).success` OR return the raw reply.
- `reverseGeocode({latitude, longitude})` → `instance.post('/map/reverse-geocode/v1', { latitude, longitude })` returns `{ address: string }`.

**Omit-type pattern** (`fuel.ts:58-67`):
```typescript
type FuelVehicleRequest = Omit<FuelVehicle, 'id'> & {
  tankCapacity: string | number
}
```
**Takeaway for map.ts:** `CreateSpotRequest = Omit<SpotEntity, 'id' | 'createdAt' | 'updatedAt'>`. `UpdateSpotRequest = Omit<SpotEntity, 'createdAt' | 'updatedAt'>` (keeps `id`). Exactly the `appVersion.ts:36,42` shape.

**Upload pattern** (`appVersion.ts:51-55` — the existing `POST /file/upload/raw/v1` client):
```typescript
export async function uploadFile(file: File): Promise<{ id: string; url: string }> {
  const formData = new FormData()
  formData.append('file', file)
  return await instance.post('/file/upload/raw/v1', formData)
}
```
**Takeaway for D-16 multi-photo upload:** copy this exactly, but add `onUploadProgress` to the axios config to wire per-file progress (per D-16). `request.ts`'s `instance` is a plain axios instance — it forwards unknown config to axios, so `{ onUploadProgress: (e) => ... }` works with zero changes to `request.ts`. **Conclusion: `request.ts` does NOT need modification.** (See Risk Map.)

---

### `src/stores/mapStore.ts` ← `src/stores/fuelStore.ts`

**Analog:** `price_recorder_vue/src/stores/fuelStore.ts` (210 lines, read in full).

**Store skeleton + state** (`fuelStore.ts:7-21`):
```typescript
import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { useToast } from 'vue-toastification'
import * as fuelApi from '@/api/fuel'
import type { FuelVehicle, RefuelRecord } from '@/api/fuel'

export const useFuelStore = defineStore('fuel', () => {
  const toast = useToast()
  const vehicles = ref<FuelVehicle[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const total = ref(0)
  const currentPage = ref(1)
  const pageSize = ref(12)
```
**Takeaway for mapStore:** `defineStore('map', () => { ... })` setup-syntax store. State: `spots: ref<SpotEntity[]>([])`, `selectedSpot: ref<SpotEntity | null>(null)` (for bottom sheet D-10), `loading: ref(false)`, `error: ref<string | null>(null)`, plus D-16 local `uploadingPhotos: ref<Array<{ url: string; progress: number }>>([])` (or similar) — these are the URLs that get attached to the in-progress spot draft. Optionally computed getters: `spotCount`, `filteredSpots` (filter by selected tags/search — D-12, D-19).

**Async action + toast mutation pattern** (`fuelStore.ts:60-72, 91-104`):
```typescript
async function createVehicle(data: Omit<FuelVehicle, 'id'>): Promise<void> {
  loading.value = true
  try {
    await fuelApi.createVehicle(data)
    toast.success('车辆创建成功')
    await fetchVehicles()                              // re-fetch list after create
  } catch (err: any) {
    toast.error(err.message || '创建失败')
    throw err
  } finally {
    loading.value = false
  }
}

async function deleteVehicle(id: string): Promise<void> {
  loading.value = true
  try {
    await fuelApi.deleteVehicle(id)
    toast.success('车辆删除成功')
    vehicles.value = vehicles.value.filter((v) => v.id !== id)   // local mutation — no re-fetch
    total.value = Math.max(0, total.value - 1)
  } catch (err: any) {
    toast.error(err.message || '删除失败')
    throw err
  } finally {
    loading.value = false
  }
}
```
**Takeaway for mapStore:** Toast notifications live IN the store (not the component) — copy this exact try/catch/finally shape. Local-mutation-after-delete (filter the `spots` array) is preferred over re-fetching for snappy UI — mirror `deleteVehicle`. After `saveSpot`/`updateSpot`/`reverseGeocode` either re-fetch the list or update the relevant entry in-place; `saveSpot` should `await listSpots()` (mirror `createVehicle`'s `await fetchVehicles()`).

**Return block** (`fuelStore.ts:181-209`): return all state refs + action functions as a flat object. Computed getters also returned.

---

### `src/view/Map.vue` ← `src/view/DebtDetail.vue` (modal/transition) + RESEARCH Pattern 3 (AMap)

**Analog A — bottom-sheet modal + transition:** `price_recorder_vue/src/view/DebtDetail.vue:280-326`.

```html
<Transition enter-active-class="transition duration-200 ease-out" enter-from-class="opacity-0" enter-to-class="opacity-100" leave-active-class="transition duration-150 ease-in" leave-from-class="opacity-100" leave-to-class="opacity-0">
  <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4" style="background-color: rgba(0,0,0,0.35);" @click.self="closeModal">
    <Transition enter-active-class="transition duration-300 ease-out" enter-from-class="opacity-0 scale-95 translate-y-2" enter-to-class="opacity-100 scale-100 translate-y-0" leave-active-class="transition duration-200 ease-in" leave-from-class="opacity-100 scale-100 translate-y-0" leave-to-class="opacity-0 scale-95 translate-y-2">
      <div v-if="showModal" class="bg-white w-full max-w-lg rounded-2xl shadow-2xl overflow-hidden">
        <!-- header / form / footer -->
      </div>
    </Transition>
  </div>
</Transition>
```
**Takeaway for Map.vue bottom sheet (D-15):** reuse the nested-`<Transition>` pattern, but change the inner panel: replace `items-center justify-center` with `items-end justify-center` (bottom-anchored) and swap the inner transition's `translate-y-2`→`translate-y-full` enter-from (slide up from bottom). The overlay `@click.self="closeModal"` and the close button stay. Per D-15 prefer custom Tailwind over a library.

**Form + submit + loading state** (`DebtDetail.vue:17-19, 56-72`):
```typescript
const showModal = ref(false)
const isEditing = ref(false)
const isSubmitting = ref(false)
const formData = ref({ id: '', debtId: '', postingDate: '', principal: '', interest: '', period: '' })

async function handleSubmit() {
  if (!formData.value.postingDate.trim()) return
  isSubmitting.value = true
  try {
    const payload = { ...formData.value, postingDate: formData.value.postingDate + ' 00:00:00', principal: String(formData.value.principal ?? '') }
    if (isEditing.value) await detailStore.updateDetail(payload)
    else { const { id, ...rest } = payload; await detailStore.createDetail(rest) }
    showModal.value = false
  } catch (err: any) { alert(err.message || '操作失败') }
  finally { isSubmitting.value = false }
}
```
**Takeaway for Map.vue capture form (D-17):** copy `isSubmitting` + try/finally + `showModal.value = false` on success. Note `DebtDetail.vue:70` uses `alert` for errors — **prefer the store's toast** (fuelStore pattern) instead; DebtDetail is the older style. Required-field gate (`if (!formData.value.postingDate.trim()) return`) is the validation analog for "name required" (D-09).

**File-input change handler** (`DebtDetail.vue:101-131`) — analog for the multi-photo picker (D-16):
```typescript
async function handleOcrFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  clearOcrState()
  try {
    const reply = await detailStore.recognizeOcr(file, debtId)
    // ... populate form fields from reply
  } catch (err: any) {
    ocrError.value = err.message || 'OCR识别失败'
  } finally {
    input.value = ''                                              // reset input so same file re-triggers
  }
}
```
**Takeaway for D-16:** loop over `input.files` (multiple), call `uploadFile(file, onProgress)` per file in parallel (`Promise.allSettled`), accumulate `{ url, progress }` into a reactive array, and `input.value = ''` in finally so re-selecting the same file re-fires the change event. The OCR handler is the closest existing example of `<input type="file">` + store interaction.

**Analog B — AMap loader + map init (RESEARCH Pattern 3, lines 321-353 + Code Examples 503-535):**

```typescript
// RESEARCH.md:336-353 — Map.vue setup
import { ref, onMounted, onUnmounted } from 'vue'

const map = ref<any>(null)
let AMapInstance: any = null

onMounted(async () => {
  const AMap = await (window as any).AMapLoader.load({
    key: '「你的Key」',                 // → import.meta.env.VITE_GAODE_JS_API_KEY
    version: '2.0',
  })
  AMapInstance = AMap
  map.value = new AMap.Map('map-container', {
    zoom: 13,
    center: [113.5, 23.0],            // default center (Guangzhou area)
  })
})
```
```typescript
// RESEARCH.md:376-388 — WGS-84 → GCJ-02 (Pattern 5)
function convertWgs84ToGcj02(lng: number, lat: number): Promise<[number, number]> {
  return new Promise((resolve, reject) => {
    AMap.convertFrom([lng, lat], 'gps', (status: string, result: any) => {
      if (status === 'complete' && result.info === 'ok') {
        const gcjLoc = result.locations[0]
        resolve([gcjLoc.lng, gcjLoc.lat])
      } else { reject(new Error('Coordinate conversion failed')) }
    })
  })
}
```
```typescript
// RESEARCH.md:541-560 — Add marker from spot data
async function addSpotMarker(spot: Spot) {
  const gcjCoords = await convertWgs84ToGcj02(spot.longitude, spot.latitude)
  const marker = new AMap.Marker({ position: gcjCoords, title: spot.name,
    icon: new AMap.Icon({ size: new AMap.Size(32, 32), image: '/marker-fish.png', imageSize: new AMap.Size(32, 32) }) })
  marker.on('click', () => openBottomSheet(spot))
  mapInstance.value.add(marker)
  markers.value.set(spot.id, marker)
}
```
```typescript
// RESEARCH.md:532-534 — cleanup
onUnmounted(() => { mapInstance.value?.destroy() })
```
**Takeaway:** these four blocks are the verbatim skeleton for Map.vue's `<script setup>`. Plugs: (1) inject keys via `import.meta.env.VITE_GAODE_JS_API_KEY` / `VITE_GAODE_JS_SECURITY_CODE`; (2) set `window._AMapSecurityConfig` BEFORE AMapLoader.load — put it in `index.html` or at the top of `onMounted`; (3) default center fallback per D-20: `[113.264385, 23.129112]` (Guangzhou, GCJ-02 — no conversion needed for this static fallback).

**Navigation deep-link (RESEARCH Pattern 6, lines 393-405 + Code 619-640):**
```typescript
function navigateToSpot(name: string, gcjLng: number, gcjLat: number) {
  const isIOS = /iPhone|iPad|iPod/.test(navigator.userAgent)
  const scheme = isIOS ? 'iosamap://' : 'androidamap://'
  const uri = `${scheme}navi?sourceApplication=blog&poiname=${encodeURIComponent(name)}&lat=${gcjLat}&lon=${gcjLng}&dev=0&style=2`
  window.location.href = uri
  setTimeout(() => {
    if (document.hasFocus()) {
      window.open(`https://uri.amap.com/navigation?to=${gcjLng},${gcjLat},${encodeURIComponent(name)}`, '_blank')
    }
  }, 2000)
}
```
**Takeaway:** `dev=0` because frontend already converted to GCJ-02 (RESEARCH Pattern 6). Use this verbatim for D-03 / D-10 Navigate button.

---

### `src/components/AppLayout.vue` (modify) ← itself

**Analog:** `price_recorder_vue/src/components/AppLayout.vue:12-18` (the `navItems` array).

```typescript
const navItems = [
  { name: 'blog', path: '/blog', label: '博文' },
  { name: 'debt', path: '/debt', label: '债务' },
  { name: 'fuel', path: '/fuel', label: '油耗' },
  { name: 'price', path: '/price', label: '价格' },
  { name: 'appVersion', path: '/app-version', label: '版本' },
]
```
**Takeaway (D-14):** append ONE line as the 6th entry — `{ name: 'map', path: '/map', label: '地图' }`. The `currentRouteName` computed (`AppLayout.vue:20`) and the `v-for="item in navItems"` loops at lines 57 and 125 already generically render ANY entry — zero other edits. Do NOT collapse other tabs, do NOT replace 版本. Mobile hamburger menu auto-includes the new entry.

---

### `src/router/index.ts` (modify) ← itself

**Analog:** `price_recorder_vue/src/router/index.ts:26-69` (the `children` array under `/` AppLayout).

```typescript
children: [
  { name: 'blog', path: 'blog', component: () => import('@/view/BlogList.vue'), meta: { requiresAuth: true } },
  // ... debt, debtDetail, fuel, fuelDetail, price, appVersion ...
]
```
**Takeaway (D-21):** add as a sibling child:
```typescript
{ name: 'map', path: 'map', component: () => import('@/view/Map.vue'), meta: { requiresAuth: true } },
```
Auth guard at `router/index.ts:74-90` already protects any child of the `/` route (which itself has `meta: { requiresAuth: true }` at line 24) — no guard changes needed.

---

### `src/utils/request.ts` — likely NO modification (verify only)

**Analog:** itself, `price_recorder_vue/src/utils/request.ts` (29 lines, read in full).

```typescript
const instance = axios.create({ baseURL })
// response interceptor: unwraps body.data (lines 7-13)
// request interceptor: injects Authorization Bearer token (lines 20-28)
export default instance;
```
**Takeaway:** `instance.post(url, formData, { onUploadProgress })` forwards the third arg to axios verbatim — `onUploadProgress` works with zero edits. **Risk Map flag:** verify by writing one test in `map.spec.ts` that asserts `postMock` was called with a config object containing `onUploadProgress` — do NOT modify `request.ts` unless proven broken.

---

### `src/__tests__/stores/mapStore.spec.ts` ← `src/__tests__/stores/fuelStore.spec.ts`

**Analog:** `price_recorder_vue/src/__tests__/stores/fuelStore.spec.ts` (89 lines, read in full).

```typescript
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useFuelStore } from '@/stores/fuelStore'
import * as fuelApi from '@/api/fuel'

vi.mock('vue-toastification', () => ({
  useToast: () => ({ success: vi.fn(), error: vi.fn() }),
}))
vi.mock('@/api/fuel', () => ({
  getVehicles: vi.fn(), createVehicle: vi.fn(), getFuelStats: vi.fn(), getRefuelRecords: vi.fn(),
}))

describe('fuelStore', () => {
  beforeEach(() => { setActivePinia(createPinia()); vi.clearAllMocks() })

  it('loads vehicle list and pagination state', async () => {
    vi.mocked(fuelApi.getVehicles).mockResolvedValue({ page: '1', total: '1', list: [ {...} ] })
    const store = useFuelStore()
    await store.fetchVehicles(1, 12)
    expect(store.vehicles).toHaveLength(1)
    expect(store.total).toBe(1)
    expect(fuelApi.getVehicles).toHaveBeenCalledWith('1', '12')
  })
})
```
**Takeaway (D-21):** mirror exactly — `vi.mock('vue-toastification')` + `vi.mock('@/api/map')` listing every map.ts export as a `vi.fn()`. `beforeEach` resets pinia + mocks. Test: `listSpots` populates `store.spots`, `saveSpot` calls api then re-lists, `deleteSpot` filters locally + decrements count. No axios/network — pure store mutation assertions.

---

### `src/__tests__/api/map.spec.ts` ← `src/__tests__/api/fuel.spec.ts`

**Analog:** `price_recorder_vue/src/__tests__/api/fuel.spec.ts` (68 lines, read in full).

```typescript
import { beforeEach, describe, expect, it, vi } from 'vitest'

const getMock = vi.fn()
const postMock = vi.fn()

vi.mock('@/utils/request.ts', () => ({
  default: { get: getMock, post: postMock },
}))

describe('fuel api', () => {
  beforeEach(() => { vi.clearAllMocks(); postMock.mockResolvedValue({}) })

  it('serializes decimal fields as strings before posting vehicles', async () => {
    const { createVehicle } = await import('@/api/fuel')
    await createVehicle({ name: 'Car', plateNo: '', brand: '', model: '', tankCapacity: 7 as any, remark: '' })
    expect(postMock).toHaveBeenCalledWith('/fuel/vehicle/save/v1', {
      name: 'Car', plateNo: '', brand: '', model: '', tankCapacity: '7', remark: '',
    })
  })
})
```
**Takeaway (D-21):** copy the `getMock`/`postMock` + `vi.mock('@/utils/request.ts', ...)` scaffolding. Map tests: assert `listSpots()` calls `getMock` with `/map/list/v1` (no params), `saveSpot(...)` calls `postMock` with `/map/save/v1` and serialized payload (lat/lng as strings via decimalToString), `deleteSpot('5')` posts `{ id: '5' }` to `/map/delete/v1`, `reverseGeocode({lat,lng})` posts to `/map/reverse-geocode/v1`. Add ONE test for `uploadFile(file, onProgress)` asserting `postMock` was called with `formData` AND a config object containing `onUploadProgress`.

---

### `src/__tests__/view/Map.spec.ts` — partial analog (compose from three)

**Analogs:**
- `src/__tests__/utils/request.spec.ts` (55 lines, read in full) — for the `setActivePinia` + interceptor-mock scaffolding.
- `fuelStore.spec.ts` (above) — for the store-mock pattern.
- RESEARCH anti-patterns + D-22 for `vi.skipIf` strategy.

```typescript
// request.spec.ts scaffolding to reuse:
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

describe('request interceptor', () => {
  beforeEach(() => { setActivePinia(createPinia()); localStorage.clear(); vi.resetModules() })
  // ...tests...
})
```
**Takeaway (D-21 + D-22):** For Map.spec.ts:
1. `vi.mock('@/stores/mapStore')` returning a stub store with `vi.fn()` actions.
2. `vi.mock` the AMap globals: `window.AMapLoader = { load: vi.fn().mockResolvedValue({...stub AMap namespace...}) }`, `window.AMap = { Map: vi.fn(), Marker: vi.fn(), convertFrom: vi.fn() }`, `window._AMapSecurityConfig` allowed to be undefined.
3. `vi.stubGlobal('navigator', { geolocation: { getCurrentPosition: vi.fn() } })` for the capture FAB.
4. `vi.stubGlobal('URL', { createObjectURL: vi.fn().mockReturnValue('blob:stub') })` for photo previews.
5. Mount with `@vue/test-utils` `mount(Map, { global: { plugins: [pinia] } })` and assert: FAB click → `navigator.geolocation.getCurrentPosition` called; pin click → `selectedSpot` mutation; nav button click → `window.location.href` set to `androidamap://...`.
6. Real-map behaviors gated per D-22:
```typescript
it.skipIf(typeof (window as any).AMap?.Map !== 'function',
  'renders real map tiles (skipped in happy-dom — manual UAT)', () => { /* ... */ })
```
Skip list MUST be enumerated in a comment block at the top of the file and re-routed to `13-VALIDATION.md` (per D-22): pin clustering under 50+ pins, touch-pinch zoom, real tile rendering, geolocation-denied fallback on a real device, real `AMap.Map` DOM attachment.

---

### `.env.example` (NEW) — no analog

No `.env*` file exists in `price_recorder_vue/` (verified via glob). Create `.env.example` documenting the two `VITE_*` vars (D-frontend specifics):
```
VITE_GAODE_JS_API_KEY=your_web_js_api_key
VITE_GAODE_JS_SECURITY_CODE=your_js_security_code
```
Crucial note for the planner: **`VITE_` prefix is REQUIRED** — Vite only exposes env vars prefixed with `VITE_` to client code via `import.meta.env`. Document in `.env.example` that these are distinct from the backend `GAODE_WEB_API_KEY` (server-side, NEVER in frontend).

### `.planning/REQUIREMENTS.md` (modify) — gap-closer

Per CONTEXT specifics (lines 149): Register `MAP-01`, `MAP-02`, `MAP-03` rows in the ACTIVE section pointing to Phase 13. No analog code needed — pure doc edit closing the traceability hop VERIFICATION flagged.

## Shared Patterns

### Toast + Loading + Error wrapper
**Source:** `fuelStore.ts:60-72` (excerpt above)
**Apply to:** every `mapStore` async action (saveSpot, updateSpot, deleteSpot, listSpots, getSpot, reverseGeocode, uploadPhoto).
Pattern: `loading=true → try { await api; toast.success; await refetch OR local mutate } catch { toast.error; throw } finally { loading=false }`.

### Axios instance + response unwrap
**Source:** `src/utils/request.ts:6-13`
```typescript
instance.interceptors.response.use(
  (res) => { const body = res.data as { code?; message?; data? } | undefined
             if (!body || body.code === 200) return body?.data
             return Promise.reject(new Error(body.message || "请求失败")) },
  (err) => { const msg = err.response?.data?.message || err.message || "网络错误"
             alert(msg); return Promise.reject(new Error(msg)) }
)
```
**Apply to:** every `map.ts` function — they return `body.data` directly (the proto reply). Toast injection happens in the STORE layer, not here. The interceptor's `alert(msg)` for network errors is acceptable but store-level `toast.error` overlays it for richer UX — this layered feedback is the established `fuel.ts`/`fuelStore.ts` convention.

### JWT auth header
**Source:** `src/utils/request.ts:20-28`
```typescript
instance.interceptors.request.use(req => {
  const userStore = useUserStore()
  if (userStore.token) req.headers.Authorization = "Bearer " + userStore.token
  return req
})
```
**Apply to:** all `/map/*` calls inherit auth automatically — no per-call work. Confirmed all 6 map endpoints are JWT-protected (no whitelist, per 13-02-summary).

## No Analog Found

| File | Role | Reason | Planner fallback |
|------|------|--------|------------------|
| `src/view/Map.vue` (Gaode-specific parts) | view | No existing map view in repo; AMap.* APIs are novel | Use RESEARCH.md Pattern 3 + Pattern 5 + Code Examples verbatim (quoted in assignment above) |
| `src/__tests__/view/Map.spec.ts` | component-spec | No existing `.vue` component spec in `src/__tests__/` — only store/api/util specs exist | Compose from `request.spec.ts` (scaffolding) + `fuelStore.spec.ts` (mock pattern) + RESEARCH D-22 (`vi.skipIf` strategy) |
| `.env.example` | config | No `.env*` exists in repo | Create from scratch with the two `VITE_*` vars |

## Pattern Risk Map

Things that look similar but aren't — flag in 13-03-PLAN.md `risks`:

1. **`DeleteSpotReply` is `{ success: bool }`, NOT `{ flag: bool }`.** `fuel.ts:114` casts to `{ flag: boolean }` — that's the *fuel* backend reply shape. Map's proto says `bool success = 1`. `map.ts` must cast to `{ success: boolean }`. Copy-pasting fuel's delete return pattern blindly will break.
2. **`GetSpot` HTTP path-template vs query-param.** Proto says `get: /map/get/{id}` (path templating), but `fuel.ts:102`/`debt.ts:37` use the query form `instance.get(url, { params: { id } })`. Kratos's generated HTTP handler accepts BOTH (`{id}` path AND `?id=` query) — the existing two callers prove the query form works. **Follow the established convention (query form).** DO NOT manually build `/map/get/${id}` — it skips the `/api` baseURL consistency and may bypass the JWT interceptor's path normalization.
3. **`ReverseGeocodeRequest` proto field order is `latitude=1, longitude=2`.** RESEARCH/SUMMARY repeatedly says "send longitude first". This is a *code-style* convention for readability — JSON body keys are unordered on the wire, so the Go handler reads `req.Longitude` and `req.Latitude` correctly regardless of JSON key order. Just include both keys; the backend builds `location=lng,lat` URL internally (verified in 13-01-summary lines 38-40).
4. **`AMap.MarkerClusterer` is a PLUGIN, not core.** Requires `AMap.plugin('AMap.MarkerClusterer', cb)` async load before use (RESEARCH Open Question 2 mentions clustering at 10+ spots). Do NOT call `new AMap.MarkerClusterer(...)` directly — it will be `undefined`. Either gate clustering behind `AMap.plugin(...)` callback or defer to a follow-up.
5. **`window._AMapSecurityConfig` must be set BEFORE `AMapLoader.load`.** Per RESEARCH Pattern 3 lines 329-333. If set inside `onMounted` after the loader script tag has already executed, the security check may run with the old config. Set it in `index.html` `<head>` or at the very top of `Map.vue`'s `<script setup>` module scope (not inside `onMounted`).
6. **`request.ts`'s response interceptor calls `alert(msg)` on network error (line 16).** This is the established behavior — do NOT "fix" it in Plan 13-03 (out of scope). Map.vue's toast error overlay supplements it; both popups on a real network failure is acceptable per existing fuel/debt UX.
7. **`onUploadProgress` config forwarding is UNTESTED in this repo.** `appVersion.ts:54` calls `instance.post(url, formData)` with NO third arg. D-16's per-file progress is the first use of the third-arg config. The axios `instance` should forward it (axios's `post(url, data, config?)` signature), but `request.ts`'s interceptor pipeline does not strip the config — verify with a `map.spec.ts` assertion (per the api-spec takeaway above). If it breaks, the fix is a one-line pass-through in `request.ts` (NOT a wrapper) — but try without modifying first.
8. **RESEARCH Pitfall 1 — GCJ-02 offset.** Storing WGS-84 but NOT calling `AMap.convertFrom` before `new AMap.Marker({ position })` produces 300-500m NE offset. Every marker creation path MUST go through `convertWgs84ToGcj02` (RESEARCH Pattern 5). The reverse-geocode backend endpoint takes RAW WGS-84 (per 13-01-summary line 38-40 — backend builds `location=lng,lat` for the Gaode regeo which itself accepts raw GPS coords? **NO** — Gaode regeo expects GCJ-02. Verify: 13-02-summary line 23 says "ReverseGeocode handler passes (lng, lat) to MatchUsecase — Gaode API location=lng,lat ordering preserved end-to-end" but does NOT say it converts WGS→GCJ first. **FLAG for the planner:** the ReverseGeocode backend endpoint may be receiving WGS-84 from the frontend and passing it raw to Gaode regeo — which would return an offset address. Check `blog/internal/biz/map.go` to confirm whether a conversion step exists. If not, the frontend should either (a) call `convertWgs84ToGcj02` BEFORE calling `reverseGeocode`, or (b) the planner raises this as an open question for the user. This is a potential regeo-correctness bug.)
9. **RESEARCH Pitfall 3 — HTTPS for Geolocation.** `navigator.geolocation.getCurrentPosition` silently fails on HTTP in production (Chrome 50+, Safari 10+). Localhost is whitelisted. For mobile LAN testing, use ngrok HTTPS tunnel. Document in `.env.example`/README and in `13-VALIDATION.md` UAT.
10. **RESEARCH Pitfall 4 — Gaode JS API key domain whitelist.** Key must whitelist `localhost` AND the production origin in the Gaode console, else blank tiles. Document in `.env.example`.

## Metadata

**Analog search scope:** `price_recorder_vue/src/{api,stores,view,components,utils,__tests__}/` + `blog/api/{map,file}/v1/*.proto` + `blog/internal/server/http.go` + `blog/internal/service/file.go` (grep only).
**Files scanned (Read in full):** 13 (4 planning docs + 9 source files + 2 proto files).
**Files scanned (Grep only):** 4 (FormData, onUploadProgress, HandleRawUpload, VITE_).
**Pattern extraction date:** 2026-07-06.