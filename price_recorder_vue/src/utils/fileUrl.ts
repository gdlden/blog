/**
 * 把后端返回的文件访问路径解析成可直接用于 <img src> / <a href> 的 URL。
 *
 * 后端上传/查询接口统一返回相对下载路径（/file/download/v1/{id}），
 * dev 下需加 /api 前缀经 Vite 代理转发；完整 URL（http/https）与已带
 * /api 前缀的路径原样返回。
 */
export function resolveFileUrl(url?: string): string {
  if (!url) return ''
  if (url.startsWith('http://') || url.startsWith('https://')) return url
  if (url.startsWith('/api')) return url
  return `/api${url}`
}
