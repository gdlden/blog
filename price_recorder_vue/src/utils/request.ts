import axios from "axios";
import { useUserStore } from "@/stores/userStore";

const baseURL = "/api";
const instance = axios.create({ baseURL })

/**
 * 后端 proto 响应的 int64 字段（id、vehicleId、fileId 等）经自定义 encoder
 * 序列化为 JSON 数字，而前端统一按字符串契约使用 id。这里在响应边界把
 * 所有 id 类字段的数字值递归转为字符串，保证全前端 id 一律为 string。
 */
function normalizeIdStrings(value: unknown): unknown {
  if (Array.isArray(value)) {
    return value.map(normalizeIdStrings)
  }
  if (value && typeof value === "object") {
    const result: Record<string, unknown> = {}
    for (const [key, item] of Object.entries(value)) {
      const isIdKey = key === "id" || /Id$/.test(key)
      result[key] = isIdKey && typeof item === "number" ? String(item) : normalizeIdStrings(item)
    }
    return result
  }
  return value
}

instance.interceptors.response.use(
  (res) => {
    const body = res.data as { code?: number; message?: string; data?: any } | undefined;
    if (!body || body.code === 200) {
      return normalizeIdStrings(body?.data) as any;
    }
    return Promise.reject(new Error(body.message || "请求失败"));
  },
  (err) => {
    const msg = err.response?.data?.message || err.message || "网络错误";
    alert(msg);
    return Promise.reject(new Error(msg));
  }
)
instance.interceptors.request.use(
  req=>{
    const userStore = useUserStore()
    if (userStore.token) {
      req.headers.Authorization = "Bearer " + userStore.token
    }
    return req;
  }
)
export default instance;
