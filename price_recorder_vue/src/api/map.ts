import instance from '@/utils/request.ts'

// SpotEntity mirrors the proto3 SpotEntity in blog/api/map/v1/map.proto.
// Lat/lng are double on the wire (numbers); we serialize as strings via
// decimalToString for parity with fuel.ts and the backend's strconv round-trip.
export interface SpotEntity {
  id: string
  name: string
  latitude: number
  longitude: number
  notes: string
  tags: string
  photos: string[]
  address: string
  createdAt: string
  updatedAt: string
}

// CreateSpotRequest = SpotEntity minus id/timestamps (see proto CreateSpotRequest).
export type CreateSpotRequest = Omit<SpotEntity, 'id' | 'createdAt' | 'updatedAt'>

// UpdateSpotRequest keeps id (see proto UpdateSpotRequest).
export type UpdateSpotRequest = Omit<SpotEntity, 'createdAt' | 'updatedAt'>

function decimalToString(value: string | number | undefined | null): string {
  if (value === undefined || value === null) return ''
  return String(value)
}

// Only lat/lng become strings; other fields keep their TypeScript-native shape
// per the proto contract (tags=string, photos=string[]).
function serializeSpot(data: CreateSpotRequest | UpdateSpotRequest) {
  return {
    ...data,
    latitude: decimalToString((data as any).latitude),
    longitude: decimalToString((data as any).longitude),
  }
}

// listSpots() — ListSpotsRequest is empty; reply is { spots: SpotEntity[] }.
// request.ts interceptor unwraps body.data, so we get the proto reply directly.
export async function listSpots(): Promise<{ spots: SpotEntity[] }> {
  return await instance.get('/map/list/v1')
}

// getSpot(id) — CR-1: PATH-TEMPLATE, NOT query form.
// map_http.pb.go registers r.GET("/map/get/{id}", ...) — using ?id= would 404.
// fuel.ts uses the query form because fuel.proto declares get: /fuel/vehicle/get/v1
// (no path-template) — that does NOT apply here.
export async function getSpot(id: string): Promise<{ spot: SpotEntity }> {
  return await instance.get('/map/get/' + encodeURIComponent(id))
}

// saveSpot(data) — CR-3: reply is { spot: SpotEntity }, NOT SaveFuelReply { id, message }.
export async function saveSpot(data: CreateSpotRequest): Promise<{ spot: SpotEntity }> {
  return await instance.post('/map/save/v1', serializeSpot(data))
}

// updateSpot(data) — CR-3: reply is { spot: SpotEntity }.
export async function updateSpot(data: UpdateSpotRequest): Promise<{ spot: SpotEntity }> {
  return await instance.post('/map/update/v1', serializeSpot(data))
}

// deleteSpot(id) — CR-3: DeleteSpotReply is { success: bool }, NOT fuel's { flag }.
// Mirror the appVersion.ts shape conceptually (cast-and-pick).
export async function deleteSpot(id: string): Promise<boolean> {
  const reply = (await instance.post('/map/delete/v1', { id })) as { success: boolean }
  return reply.success
}

// reverseGeocode(input) — caller (Map.vue) MUST pass already-converted GCJ-02
// coords per CR-2: backend biz/map.go builds Gaode regeo with NO WGS-84→GCJ-02
// conversion. JSON keys are unordered on the wire — include both.
export async function reverseGeocode(input: {
  latitude: number
  longitude: number
}): Promise<{ address: string }> {
  return await instance.post('/map/reverse-geocode/v1', {
    latitude: input.latitude,
    longitude: input.longitude,
  })
}

// uploadSpotPhoto(file, onUploadProgress?) — mirror appVersion.ts:51-55 with
// per-file progress. request.ts forwards the third axios arg verbatim (no
// interceptor strips it), so { onUploadProgress } works without modifying it.
export async function uploadSpotPhoto(
  file: File,
  onUploadProgress?: (event: any) => void,
): Promise<{ id: string; url: string }> {
  const formData = new FormData()
  formData.append('file', file)
  const config = onUploadProgress ? { onUploadProgress } : undefined
  return await instance.post('/file/upload/raw/v1', formData, config)
}