import instance from "@/utils/request.ts";

export interface Post {
  id: string;
  title: string;
  content: string;
}

export interface PostPageResponse {
  current: string;
  size: string;
  total: string;
  data: Post[];
}

export async function getPosts(
  current?: string,
  size?: string
): Promise<PostPageResponse> {
  const params: Record<string, string> = {};
  if (current) params.current = current;
  if (size) params.size = size;
  return await instance.get("/post/page/v1", { params }).then((res) => res.data);
}

export async function getPostById(id: string): Promise<Post> {
  return await instance.get(`/post/get/${id}`).then((res) => res.data);
}

export async function createPost(data: {
  title: string;
  content: string;
}): Promise<Post> {
  return await instance.post("/post/add/v1", data).then((res) => res.data);
}

export async function updatePost(
  id: string,
  data: { title: string; content: string }
): Promise<Post> {
  return await instance
    .post("/post/edit/v1", { id, ...data })
    .then((res) => res.data);
}

export async function deletePost(id: string): Promise<void> {
  await instance.post("/post/delete/v1", { id });
}
