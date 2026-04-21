import instance from "@/utils/request.ts";

export interface userReq {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  userId?: number | string
  username?: string
  [key: string]: any
}

export let user = <userReq>{
  username: "",
  password: "",
}

export async function login(data: userReq): Promise<LoginResponse | undefined> {
  return await instance.post("/user/login/v1", data)
    .then(res => {
      return res.data as LoginResponse
    })
    .catch((err) => {
      console.log(err)
      return undefined
    })
}
