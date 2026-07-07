import { beforeEach, describe, expect, it, vi } from 'vitest'

const getMock = vi.fn()
const postMock = vi.fn()

vi.mock('@/utils/request.ts', () => ({
  default: {
    get: getMock,
    post: postMock,
  },
}))

describe('map api', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    postMock.mockResolvedValue({})
    getMock.mockResolvedValue({})
  })

  it('listSpots() calls GET /map/list/v1 with no params', async () => {
    getMock.mockResolvedValue({ spots: [] })
    const { listSpots } = await import('@/api/map')

    const result = await listSpots()

    expect(getMock).toHaveBeenCalledWith('/map/list/v1')
    expect(result).toEqual({ spots: [] })
  })

  it('saveSpot serializes lat/lng as strings and posts to /map/save/v1 (CR-3 reply { spot })', async () => {
    postMock.mockResolvedValue({ spot: { id: '1' } })
    const { saveSpot } = await import('@/api/map')

    const result = await saveSpot({
      name: 'X',
      latitude: 23.1,
      longitude: 113.2,
      notes: '',
      tags: '',
      photos: [],
      address: '',
    })

    expect(postMock).toHaveBeenCalledWith('/map/save/v1', {
      name: 'X',
      latitude: '23.1',
      longitude: '113.2',
      notes: '',
      tags: '',
      photos: [],
      address: '',
    })
    expect(result).toEqual({ spot: { id: '1' } })
  })

  // CR-1 regression guard: PATTERNS.md risk-flag #2 says to use the query form
  // (like fuel.ts) — VERIFIED FALSE for the map domain. map_http.pb.go registers
  // /map/get/{id} as a path-template route; /map/get/v1?id=5 would 404.
  it('getSpot uses the path-template form /map/get/{id} (CR-1 regression guard)', async () => {
    getMock.mockResolvedValue({ spot: { id: '5' } })
    const { getSpot } = await import('@/api/map')

    await getSpot('5')

    expect(getMock).toHaveBeenCalledWith('/map/get/5')
  })

  it('getSpot URL-encodes the id segment', async () => {
    getMock.mockResolvedValue({ spot: {} })
    const { getSpot } = await import('@/api/map')

    await getSpot('a b/1')

    expect(getMock).toHaveBeenCalledWith('/map/get/' + encodeURIComponent('a b/1'))
  })

  // CR-3 regression guard: DeleteSpotReply is { success: bool }, NOT { flag }.
  // Copy-pasting fuel's delete return path would break the call site.
  it('deleteSpot posts { id } to /map/delete/v1 and returns reply.success (CR-3)', async () => {
    postMock.mockResolvedValue({ success: true })
    const { deleteSpot } = await import('@/api/map')

    const ok = await deleteSpot('5')

    expect(postMock).toHaveBeenCalledWith('/map/delete/v1', { id: '5' })
    expect(ok).toBe(true)
  })

  it('reverseGeocode posts both latitude and longitude to /map/reverse-geocode/v1', async () => {
    postMock.mockResolvedValue({ address: '测试地址' })
    const { reverseGeocode } = await import('@/api/map')

    const result = await reverseGeocode({ latitude: 23.1, longitude: 113.2 })

    expect(postMock).toHaveBeenCalledWith('/map/reverse-geocode/v1', {
      latitude: 23.1,
      longitude: 113.2,
    })
    expect(result).toEqual({ address: '测试地址' })
  })

  // PATTERNS risk-flag #7 regression guard: request.ts forwards the third
  // axios arg verbatim, but no existing call exercises it (appVersion.ts:54
  // passes only two args). Assert uploadSpotPhoto passes onUploadProgress in
  // the config so a future refactor of request.ts can't silently strip it.
  it('uploadSpotPhoto forwards onUploadProgress in the third axios config arg (risk-flag #7)', async () => {
    postMock.mockResolvedValue({ id: 'f1', url: 'https://cdn/x.png' })
    const { uploadSpotPhoto } = await import('@/api/map')

    const onProgress = (e: any) => {
      void e
    }
    const file = new File(['data'], 'photo.png', { type: 'image/png' })
    await uploadSpotPhoto(file, onProgress)

    expect(postMock).toHaveBeenCalledTimes(1)
    const [, body, config] = postMock.mock.calls[0]
    expect(postMock.mock.calls[0][0]).toBe('/file/upload/raw/v1')
    expect(body).toBeInstanceOf(FormData)
    expect(config).toEqual({ onUploadProgress: onProgress })
  })

  it('uploadSpotPhoto omits the third arg when no progress callback is provided', async () => {
    postMock.mockResolvedValue({ id: 'f2', url: 'https://cdn/y.png' })
    const { uploadSpotPhoto } = await import('@/api/map')

    await uploadSpotPhoto(new File(['data'], 'p.png'))

    expect(postMock.mock.calls[0].length).toBe(2)
  })
})