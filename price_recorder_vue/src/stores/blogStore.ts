import { ref, computed } from "vue";
import { defineStore } from "pinia";
import * as blogApi from "@/api/blog";
import type { Post } from "@/api/blog";

export const useBlogStore = defineStore("blog", () => {
  // State
  const posts = ref<Post[]>([]);
  const currentPost = ref<Post | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const total = ref(0);
  const currentPage = ref(1);
  const pageSize = ref(10);

  // Getters
  const postCount = computed(() => posts.value.length);
  const totalPages = computed(() => Math.ceil(total.value / pageSize.value));

  // Actions
  async function fetchPosts(page?: number, size?: number): Promise<void> {
    loading.value = true;
    error.value = null;
    try {
      const pageNum = page ?? currentPage.value;
      const pageSizeNum = size ?? pageSize.value;
      const response = await blogApi.getPosts(
        String(pageNum),
        String(pageSizeNum)
      );
      posts.value = response.data || [];
      total.value = parseInt(response.total || "0", 10);
      currentPage.value = pageNum;
      pageSize.value = pageSizeNum;
    } catch (err: any) {
      error.value = err.message || "获取博文列表失败";
      throw err;
    } finally {
      loading.value = false;
    }
  }

  async function createPost(data: {
    title: string;
    content: string;
  }): Promise<void> {
    loading.value = true;
    try {
      await blogApi.createPost(data);
      await fetchPosts();
    } catch (err: any) {
      throw err;
    } finally {
      loading.value = false;
    }
  }

  async function updatePost(
    id: string,
    data: { title: string; content: string }
  ): Promise<void> {
    loading.value = true;
    try {
      await blogApi.updatePost(id, data);
      await fetchPosts();
    } catch (err: any) {
      throw err;
    } finally {
      loading.value = false;
    }
  }

  async function deletePost(id: string): Promise<void> {
    loading.value = true;
    try {
      await blogApi.deletePost(id);
      posts.value = posts.value.filter((post) => post.id !== id);
      total.value = Math.max(0, total.value - 1);
    } catch (err: any) {
      throw err;
    } finally {
      loading.value = false;
    }
  }

  function setCurrentPost(post: Post | null): void {
    currentPost.value = post;
  }

  return {
    // State
    posts,
    currentPost,
    loading,
    error,
    total,
    currentPage,
    pageSize,
    // Getters
    postCount,
    totalPages,
    // Actions
    fetchPosts,
    createPost,
    updatePost,
    deletePost,
    setCurrentPost,
  };
});
