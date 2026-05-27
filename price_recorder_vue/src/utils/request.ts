import axios from "axios";
import { useUserStore } from "@/stores/userStore";

const baseURL = "/api";
const instance = axios.create({ baseURL })
instance.interceptors.response.use(
  (res) => {
    const body = res.data as { code?: number; message?: string; data?: any } | undefined;
    if (!body || body.code === 200) {
      return body?.data;
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
