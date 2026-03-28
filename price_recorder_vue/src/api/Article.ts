import instance from "@/utils/request.ts";
export async function getAllArticle() {
  return await instance.get("/post/page/v1")
    .then(res=>{
      return res.data
    })
    .catch((err)=>{
      console.log(err)
    })
}
