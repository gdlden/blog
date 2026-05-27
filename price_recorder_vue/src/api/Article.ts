import instance from "@/utils/request.ts";
export async function getAllArticle() {
  return await instance.get("/post/page/v1");
}
