import axios from "axios";
import {useUserStore} from "@/stores/userStore.ts";

const baseURL = "/api";
const instance = axios.create({ baseURL })
instance.interceptors.response.use(
  res => {
    // console.log(res.data)
    if(res.data.code === "200") {
      return res.data;
    }
    // console.log(JSON.stringify(res.data))
    return Promise.reject(res.data)
  },
  err =>{

    if(err.response) {
      console.log(err.response.status)
      console.log(err.response.data)
      // console.log(err.response.headers)
      alert(err.response.data.message)
    } else {
      console.log(err+"w22")
    }
    return Promise.reject(err);
  }
)
instance.interceptors.request.use(
  req=>{
    const userInfo = localStorage.getItem("user")
    if(userInfo != null && userInfo !== "") {
      console.log(userInfo)
      const user = JSON.parse(userInfo)
      req.headers.Authorization = "Bearer " + user.token
    }
    return req;
  }
)
export default instance;
