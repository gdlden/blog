import { describe, it, expect } from 'vitest'
import { resolveFileUrl } from '@/utils/fileUrl'

describe('resolveFileUrl', () => {
  it('prepends /api to relative download paths', () => {
    expect(resolveFileUrl('/file/download/v1/9')).toBe('/api/file/download/v1/9')
  })

  it('keeps paths already prefixed with /api', () => {
    expect(resolveFileUrl('/api/file/download/v1/9')).toBe('/api/file/download/v1/9')
  })

  it('keeps full http(s) URLs untouched', () => {
    expect(resolveFileUrl('https://file.hukss.cn/blog-files/uploads/a.jpg')).toBe(
      'https://file.hukss.cn/blog-files/uploads/a.jpg',
    )
    expect(resolveFileUrl('http://127.0.0.1:9000/bucket/a.jpg')).toBe(
      'http://127.0.0.1:9000/bucket/a.jpg',
    )
  })

  it('returns empty string for empty input', () => {
    expect(resolveFileUrl('')).toBe('')
    expect(resolveFileUrl(undefined)).toBe('')
  })
})
