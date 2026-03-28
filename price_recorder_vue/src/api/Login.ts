import instance from "@/utils/request.ts";


export interface userReq {
  username: string
  password: string
}
export  let user = <userReq>{
  username: "",
  password: "",
}
export async function login(data:userReq) {
  return await instance.post("/user/login/v1",data)
    .then(res=>{
      return res
    })
    .catch((err)=>{
      console.log(err)
    })
}
