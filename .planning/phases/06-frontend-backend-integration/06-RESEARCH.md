# Phase 6: Frontend-Backend Integration - Research

**Researched:** 2026-04-05
**Domain:** Vue 3 + Pinia + Axios CRUD, Tailwind CSS, Modal/Toast patterns
**Confidence:** HIGH

## Summary

This phase integrates the Vue 3 frontend with the Go/Kratos backend to enable full CRUD operations for blog posts and debt records. The frontend uses Vue 3 Composition API with Pinia for state management, Axios for API calls, and Tailwind CSS v4 for styling.

Key findings:
1. **Backend Gap Confirmed**: Post service lacks Update and Delete implementations (proto methods are commented, biz layer has empty Update, no Delete usecase)
2. **Pinia Setup Store Pattern**: The project already uses Composition API style stores (`userStore.ts`) - continue this pattern for `blogStore` and `debtStore`
3. **Toast Library**: `vue-toastification@1.7.14` is the stable Vue 3 version, install with `next` tag
4. **Modal Pattern**: Use Vue 3 `<Teleport to="body">` with `<Transition>` for proper z-index and positioning
5. **Card Grid**: Tailwind v4 supports `grid-cols-1 md:grid-cols-2 lg:grid-cols-3` for responsive layouts

**Primary recommendation:** Implement backend Post Update/Delete first, then build frontend API layer, stores, and UI components following established project patterns.

## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01 through D-04**: Enable full CRUD for blog posts - uncomment and implement EditPostById and DeletePostById in proto, add Update/Delete to service and usecase
- **D-05 through D-07**: API client organization by domain (`blog.ts`, `debt.ts` in `src/api/`)
- **D-08 through D-10**: Pinia stores (`useBlogStore`, `useDebtStore`) managing list data, current detail, loading states, errors
- **D-11 through D-13**: Card grid layout for lists with pagination
- **D-14 through D-16**: Modal dialogs for create/edit forms (no page navigation)
- **D-17 through D-19**: Toast notifications for feedback (auto-dismiss 3-5s), inline validation errors

### Claude's Discretion
- Exact card styling and spacing
- Toast library choice (vue-toastification vs custom)
- Modal component implementation details
- Form field validation rules

### Deferred Ideas (OUT OF SCOPE)
- Image upload for blog posts
- Advanced filtering/search
- Bulk operations (delete multiple)
- Real-time updates via WebSocket

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| BLOG-01 | User can view list of blog posts | GET /post/page/v1 exists, needs pagination UI |
| BLOG-02 | User can view full content of selected post | GET /post/get/{id} exists |
| BLOG-03 | User can create blog post | POST /post/add/v1 exists, needs modal form |
| BLOG-04 | User can edit existing post | Requires backend: uncomment proto, add Update usecase |
| BLOG-05 | User can delete existing post | Requires backend: uncomment proto, add Delete usecase |
| BLOG-06 | Manage from authenticated interface | AppLayout exists, needs store integration |
| DEBT-01 | Create debt record | POST /debt/save/v1 exists |
| DEBT-02 | View paginated debt list | GET /debt/page/v1 exists |
| DEBT-03 | View debt details | GET /debt/get/v1 exists |
| DEBT-04 | Update debt record | POST /debt/update/v1 exists |
| DEBT-05 | Delete debt record | POST /debt/delete/v1 exists |
| DEBT-06 | View debt summaries | Aggregate from list data in store |
| DEBT-07 | Manage debt-detail records | DebtDetail API exists, deferred per CONTEXT.md |

## Standard Stack

### Core (Already Installed)
| Library | Version | Purpose | Status |
|---------|---------|---------|--------|
| vue | ^3.5.26 | Framework | Installed |
| pinia | ^3.0.4 | State management | Installed |
| axios | ^1.13.3 | HTTP client | Installed |
| vue-router | ^4.6.4 | Routing | Installed |
| tailwindcss | ^4.1.18 | Styling | Installed |

### To Install
| Library | Version | Purpose | Install Command |
|---------|---------|---------|-----------------|
| vue-toastification | 1.7.14 | Toast notifications | `pnpm add vue-toastification@next` |

**Version verification:**
```bash
npm view vue-toastification version  # 1.7.14 (stable for Vue 3)
```

## Architecture Patterns

### 1. Pinia Store Pattern (Setup Store)

Following the existing `userStore.ts` pattern, use Setup Stores with Composition API:

```typescript
// stores/blogStore.ts
import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import * as blogApi from '@/api/blog'

export interface Post {
  id: string
  title: string
  content: string
}

export const useBlogStore = defineStore('blog', () => {
  // State
  const posts = ref<Post[]>([])
  const currentPost = ref<Post | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const total = ref(0)

  // Getters
  const postCount = computed(() => posts.value.length)
  const getPostById = computed(() => (id: string) => 
    posts.value.find(p => p.id === id)
  )

  // Actions
  async function fetchPosts(page = 1, size = 10) {
    loading.value = true
    error.value = null
    try {
      const res = await blogApi.getPosts(page, size)
      posts.value = res.data || []
      total.value = parseInt(res.total || '0')
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch posts'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function createPost(data: { title: string; content: string }) {
    loading.value = true
    try {
      const res = await blogApi.createPost(data)
      await fetchPosts() // Refresh list
      return res
    } finally {
      loading.value = false
    }
  }

  async function updatePost(id: string, data: { title: string; content: string }) {
    loading.value = true
    try {
      const res = await blogApi.updatePost(id, data)
      await fetchPosts() // Refresh list
      return res
    } finally {
      loading.value = false
    }
  }

  async function deletePost(id: string) {
    loading.value = true
    try {
      await blogApi.deletePost(id)
      posts.value = posts.value.filter(p => p.id !== id)
    } finally {
      loading.value = false
    }
  }

  return {
    posts,
    currentPost,
    loading,
    error,
    total,
    postCount,
    getPostById,
    fetchPosts,
    createPost,
    updatePost,
    deletePost,
  }
})
```

### 2. API Client Organization

Domain-based organization as decided in CONTEXT.md:

```typescript
// api/blog.ts
import instance from '@/utils/request.ts'

export interface Post {
  id: string
  title: string
  content: string
}

export interface PostPageResponse {
  current: string
  size: string
  total: string
  data: Post[]
}

export async function getPosts(current = '1', size = '10'): Promise<PostPageResponse> {
  const res = await instance.get('/post/page/v1', {
    params: { current, size }
  })
  return res.data
}

export async function getPostById(id: string): Promise<Post> {
  const res = await instance.get(`/post/get/${id}`)
  return res.data
}

export async function createPost(data: { title: string; content: string }): Promise<Post> {
  const res = await instance.post('/post/add/v1', data)
  return res.data
}

export async function updatePost(id: string, data: { title: string; content: string }): Promise<Post> {
  const res = await instance.post('/post/edit/v1', { id, ...data })
  return res.data
}

export async function deletePost(id: string): Promise<void> {
  await instance.post('/post/delete/v1', { id })
}
```

### 3. Modal Dialog Pattern

Use Vue 3 `<Teleport>` to avoid z-index issues:

```vue
<!-- components/Modal.vue -->
<script setup lang="ts">
const props = defineProps<{
  show: boolean
  title: string
}>()

const emit = defineEmits<{
  close: []
  confirm: []
}>()

function handleBackdropClick(e: MouseEvent) {
  if (e.target === e.currentTarget) {
    emit('close')
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-if="show"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
        @click="handleBackdropClick"
      >
        <div class="bg-white rounded-lg shadow-xl w-full max-w-lg mx-4">
          <div class="flex items-center justify-between p-4 border-b">
            <h3 class="text-lg font-semibold">{{ title }}</h3>
            <button @click="emit('close')" class="text-gray-400 hover:text-gray-600">
              <span class="sr-only">Close</span>
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          <div class="p-4">
            <slot />
          </div>
          <div class="flex justify-end gap-2 p-4 border-t bg-gray-50 rounded-b-lg">
            <button
              @click="emit('close')"
              class="px-4 py-2 rounded border border-gray-300 text-gray-700 hover:bg-gray-100"
            >
              取消
            </button>
            <button
              @click="emit('confirm')"
              class="px-4 py-2 rounded bg-blue-600 text-white hover:bg-blue-700"
            >
              保存
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
</style>
```

### 4. Card Grid Layout (Tailwind v4)

```vue
<!-- BlogList.vue card grid -->
<template>
  <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
    <div
      v-for="post in blogStore.posts"
      :key="post.id"
      class="bg-white rounded-lg shadow border border-gray-200 p-4 flex flex-col"
    >
      <h3 class="font-semibold text-lg mb-2 truncate">{{ post.title }}</h3>
      <p class="text-gray-600 text-sm line-clamp-3 flex-1">{{ post.content }}</p>
      <div class="flex justify-end gap-2 mt-4 pt-4 border-t">
        <button @click="editPost(post)" class="text-blue-600 hover:text-blue-800 text-sm">
          编辑
        </button>
        <button @click="deletePost(post.id)" class="text-red-600 hover:text-red-800 text-sm">
          删除
        </button>
      </div>
    </div>
  </div>
</template>
```

### 5. Toast Notification Pattern

```typescript
// main.ts
import { createApp } from 'vue'
import Toast from 'vue-toastification'
import 'vue-toastification/dist/index.css'
import App from './App.vue'

const app = createApp(App)
app.use(Toast, {
  position: 'top-right',
  timeout: 3000,
  closeOnClick: true,
  pauseOnFocusLoss: true,
  pauseOnHover: true,
  draggable: true,
  draggablePercent: 0.6,
  showCloseButtonOnHover: false,
  hideProgressBar: false,
  closeButton: 'button',
  icon: true,
  rtl: false,
})
app.mount('#app')
```

```typescript
// In store actions
import { useToast } from 'vue-toastification'

const toast = useToast()

async function createPost(data) {
  try {
    const res = await blogApi.createPost(data)
    toast.success('创建成功')
    return res
  } catch (err) {
    toast.error(err.message || '创建失败')
    throw err
  }
}
```

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Toast notifications | Custom toast component | vue-toastification | Global positioning, auto-dismiss, queue management, accessibility |
| Modal z-index handling | CSS-only modal | `<Teleport to="body">` | Avoids stacking context issues with fixed positioning |
| State management | Global reactive objects | Pinia stores | DevTools integration, SSR safety, TypeScript inference |
| HTTP client | Fetch wrapper | Existing Axios instance | Already configured with auth interceptor |
| Form validation | Manual validation | HTML5 + inline errors | Sufficient for simple forms, add VeeValidate later if needed |

## Backend Gaps to Fill

### Post Service (CRITICAL)

Current state in `blog/api/post/v1/post.proto`:
```protobuf
// rpc EditPostById (HelloRequest) returns (HelloReply) { ... }  // COMMENTED
// rpc DeletePostById (HelloRequest) returns (HelloReply) { ... }  // COMMENTED
```

Required additions:

1. **Proto file** - Uncomment and fix method signatures:
```protobuf
rpc UpdatePost (UpdatePostRequest) returns (UpdatePostReply) {
  option (google.api.http) = {
    post: "/post/edit/v1"
    body: "*"
  };
}
rpc DeletePost (DeletePostRequest) returns (DeletePostReply) {
  option (google.api.http) = {
    post: "/post/delete/v1"
    body: "*"
  };
}

message UpdatePostRequest {
  string id = 1;
  string title = 2;
  string content = 3;
}
message UpdatePostReply {
  string id = 1;
  string title = 2;
  string content = 3;
}
message DeletePostRequest {
  string id = 1;
}
message DeletePostReply {
  bool success = 1;
}
```

2. **biz/post.go** - Add usecase methods:
```go
func (uc *PostUsecase) UpdatePost(ctx context.Context, id int64, g *Post) (*Post, error) {
  uc.log.WithContext(ctx).Infof("UpdatePost: %v", id)
  return uc.repo.Update(ctx, g)
}

func (uc *PostUsecase) DeletePost(ctx context.Context, id int64) error {
  uc.log.WithContext(ctx).Infof("DeletePost: %v", id)
  return uc.repo.Delete(ctx, id)
}
```

3. **data/post.go** - Implement repository methods:
```go
func (r *postRepo) Update(ctx context.Context, g *biz.Post) (*biz.Post, error) {
  var post Post
  id, _ := strconv.ParseUint(g.Id, 10, 64)
  err := r.data.db.Model(&Post{}).Where("id = ?", id).Updates(map[string]interface{}{
    "title":   g.Title,
    "content": g.Content,
  }).First(&post).Error
  if err != nil {
    return nil, err
  }
  return g, nil
}

func (r *postRepo) Delete(ctx context.Context, id int64) error {
  return r.data.db.Delete(&Post{}, id).Error
}
```

4. **service/post.go** - Add service handlers:
```go
func (s *PostService) UpdatePost(ctx context.Context, req *pb.UpdatePostRequest) (*pb.UpdatePostReply, error) {
  id, _ := strconv.ParseInt(req.Id, 10, 64)
  post, _ := s.pu.UpdatePost(ctx, id, &biz.Post{
    Id:      req.Id,
    Title:   req.Title,
    Content: req.Content,
  })
  return &pb.UpdatePostReply{
    Id:      post.Id,
    Title:   post.Title,
    Content: post.Content,
  }, nil
}

func (s *PostService) DeletePost(ctx context.Context, req *pb.DeletePostRequest) (*pb.DeletePostReply, error) {
  id, _ := strconv.ParseInt(req.Id, 10, 64)
  err := s.pu.DeletePost(ctx, id)
  return &pb.DeletePostReply{Success: err == nil}, nil
}
```

## Common Pitfalls

### Pitfall 1: Pinia Store Destructuring
**What goes wrong:** Destructuring store state loses reactivity
```typescript
// WRONG - count won't update in template
const { count, posts } = useBlogStore()
```
**How to avoid:** Use `storeToRefs()` for state, direct access for actions
```typescript
import { storeToRefs } from 'pinia'
const blogStore = useBlogStore()
const { posts, loading } = storeToRefs(blogStore)  // reactive
const { fetchPosts } = blogStore  // actions can be destructured
```

### Pitfall 2: Modal Without Teleport
**What goes wrong:** Modal appears behind other content or gets clipped by parent containers with `overflow: hidden`
**How to avoid:** Always wrap modals in `<Teleport to="body">`

### Pitfall 3: Missing Error State Reset
**What goes wrong:** Error messages persist between form submissions
**How to avoid:** Clear error state at the start of each action:
```typescript
async function submitForm() {
  error.value = null  // Clear previous error
  try {
    await apiCall()
  } catch (e) {
    error.value = e.message
  }
}
```

### Pitfall 4: Backend ID Type Mismatch
**What goes wrong:** Backend uses `int64` for IDs, frontend sends strings, causing parse errors
**How to avoid:** Frontend stores IDs as strings (from JSON), backend parses with `strconv.ParseInt`

### Pitfall 5: Pagination String vs Number
**What goes wrong:** Kratos proto generates string parameters for pagination, but components expect numbers
**How to avoid:** Convert in API layer:
```typescript
export async function getPosts(current: number, size: number) {
  return instance.get('/post/page/v1', {
    params: { 
      current: String(current), 
      size: String(size) 
    }
  })
}
```

## Code Examples

### Complete Blog Store with Toast Integration

```typescript
// stores/blogStore.ts
import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { useToast } from 'vue-toastification'
import * as blogApi from '@/api/blog'

export interface Post {
  id: string
  title: string
  content: string
}

export const useBlogStore = defineStore('blog', () => {
  const toast = useToast()
  
  // State
  const posts = ref<Post[]>([])
  const currentPost = ref<Post | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const total = ref(0)
  const currentPage = ref(1)
  const pageSize = ref(10)

  // Getters
  const postCount = computed(() => posts.value.length)
  const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

  // Actions
  async function fetchPosts(page = currentPage.value, size = pageSize.value) {
    loading.value = true
    error.value = null
    try {
      const res = await blogApi.getPosts(String(page), String(size))
      posts.value = res.data || []
      total.value = parseInt(res.total || '0')
      currentPage.value = page
      pageSize.value = size
    } catch (err: any) {
      error.value = err.message || '获取博文列表失败'
      toast.error(error.value)
    } finally {
      loading.value = false
    }
  }

  async function createPost(data: { title: string; content: string }) {
    loading.value = true
    try {
      await blogApi.createPost(data)
      toast.success('博文创建成功')
      await fetchPosts()
    } catch (err: any) {
      toast.error(err.message || '创建失败')
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updatePost(id: string, data: { title: string; content: string }) {
    loading.value = true
    try {
      await blogApi.updatePost(id, data)
      toast.success('博文更新成功')
      await fetchPosts()
    } catch (err: any) {
      toast.error(err.message || '更新失败')
      throw err
    } finally {
      loading.value = false
    }
  }

  async function deletePost(id: string) {
    loading.value = true
    try {
      await blogApi.deletePost(id)
      posts.value = posts.value.filter(p => p.id !== id)
      toast.success('博文删除成功')
    } catch (err: any) {
      toast.error(err.message || '删除失败')
      throw err
    } finally {
      loading.value = false
    }
  }

  function setCurrentPost(post: Post | null) {
    currentPost.value = post
  }

  return {
    posts,
    currentPost,
    loading,
    error,
    total,
    currentPage,
    pageSize,
    postCount,
    totalPages,
    fetchPosts,
    createPost,
    updatePost,
    deletePost,
    setCurrentPost,
  }
})
```

### BlogList.vue Component

```vue
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useBlogStore } from '@/stores/blogStore'
import Modal from '@/components/Modal.vue'

const blogStore = useBlogStore()
const { posts, loading, totalPages, currentPage } = storeToRefs(blogStore)

const showModal = ref(false)
const isEditing = ref(false)
const formData = ref({ id: '', title: '', content: '' })

onMounted(() => {
  blogStore.fetchPosts()
})

function openCreateModal() {
  isEditing.value = false
  formData.value = { id: '', title: '', content: '' }
  showModal.value = true
}

function openEditModal(post: any) {
  isEditing.value = true
  formData.value = { ...post }
  showModal.value = true
}

async function handleSubmit() {
  if (isEditing.value) {
    await blogStore.updatePost(formData.value.id, {
      title: formData.value.title,
      content: formData.value.content,
    })
  } else {
    await blogStore.createPost({
      title: formData.value.title,
      content: formData.value.content,
    })
  }
  showModal.value = false
}

async function handleDelete(id: string) {
  if (confirm('确定要删除这篇博文吗？')) {
    await blogStore.deletePost(id)
  }
}

function changePage(page: number) {
  blogStore.fetchPosts(page)
}
</script>

<template>
  <div class="max-w-6xl mx-auto">
    <div class="flex justify-between items-center mb-6">
      <h2 class="text-2xl font-semibold">博文</h2>
      <button
        @click="openCreateModal"
        class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
      >
        新建博文
      </button>
    </div>

    <div v-if="loading" class="text-center py-12">
      <p class="text-gray-600">加载中...</p>
    </div>

    <div v-else-if="posts.length === 0" class="bg-gray-100 rounded p-8 text-center">
      <p class="text-lg font-semibold mb-2">暂无内容</p>
      <p class="text-gray-600">点击"新建博文"创建第一篇博文</p>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div
        v-for="post in posts"
        :key="post.id"
        class="bg-white rounded-lg shadow border border-gray-200 p-4 flex flex-col"
      >
        <h3 class="font-semibold text-lg mb-2 truncate">{{ post.title }}</h3>
        <p class="text-gray-600 text-sm line-clamp-3 flex-1">{{ post.content }}</p>
        <div class="flex justify-end gap-2 mt-4 pt-4 border-t">
          <button
            @click="openEditModal(post)"
            class="text-blue-600 hover:text-blue-800 text-sm"
          >
            编辑
          </button>
          <button
            @click="handleDelete(post.id)"
            class="text-red-600 hover:text-red-800 text-sm"
          >
            删除
          </button>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="flex justify-center gap-2 mt-8">
      <button
        v-for="page in totalPages"
        :key="page"
        @click="changePage(page)"
        :class="[
          'px-3 py-1 rounded',
          page === currentPage
            ? 'bg-blue-600 text-white'
            : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
        ]"
      >
        {{ page }}
      </button>
    </div>

    <!-- Create/Edit Modal -->
    <Modal
      :show="showModal"
      :title="isEditing ? '编辑博文' : '新建博文'"
      @close="showModal = false"
      @confirm="handleSubmit"
    >
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium mb-1">标题</label>
          <input
            v-model="formData.title"
            type="text"
            class="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-600"
          />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">内容</label>
          <textarea
            v-model="formData.content"
            rows="6"
            class="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-600"
          ></textarea>
        </div>
      </div>
    </Modal>
  </div>
</template>
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Vuex | Pinia | Vue 3 release (2020) | Simpler API, better TypeScript, Composition API style |
| Options API | Composition API | Vue 3 (2020) | Better code organization, TypeScript support |
| Vue 2 global events | `useToast()` composable | vue-toastification v2 | Works anywhere, including outside components |
| CSS frameworks | Tailwind CSS v4 | 2025 | Native CSS imports, no config needed |
| Manual z-index | `<Teleport>` | Vue 3.0 | Solves stacking context issues properly |

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Vitest 4.0.17 |
| Config file | `vitest.config.ts` (inferred) |
| Quick run command | `pnpm test:unit` |
| Full suite command | `pnpm test:unit` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| BLOG-01 | View blog list | integration | Manual verification | ❌ Wave 0 |
| BLOG-03 | Create blog post | integration | Manual verification | ❌ Wave 0 |
| BLOG-04 | Edit blog post | integration | Manual verification | ❌ Wave 0 |
| BLOG-05 | Delete blog post | integration | Manual verification | ❌ Wave 0 |
| DEBT-01 | Create debt record | integration | Manual verification | ❌ Wave 0 |
| DEBT-04 | Update debt record | integration | Manual verification | ❌ Wave 0 |
| DEBT-05 | Delete debt record | integration | Manual verification | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** Manual verification via UI
- **Per wave merge:** Full API testing via UI
- **Phase gate:** All CRUD operations verified working end-to-end

### Wave 0 Gaps
- [ ] `src/stores/blogStore.ts` - Pinia store for blog state
- [ ] `src/stores/debtStore.ts` - Pinia store for debt state
- [ ] `src/api/blog.ts` - Blog API client
- [ ] `src/api/debt.ts` - Debt API client
- [ ] `src/components/Modal.vue` - Reusable modal component
- [ ] `vue-toastification` package installation

## Sources

### Primary (HIGH confidence)
- `price_recorder_vue/src/stores/userStore.ts` - Existing Pinia pattern
- `price_recorder_vue/src/api/Article.ts` - Existing API pattern
- `price_recorder_vue/src/utils/request.ts` - Axios instance configuration
- `blog/api/post/v1/post.proto` - Backend API contract
- `blog/api/debt/v1/debt.proto` - Backend API contract
- Pinia documentation (pinia.vuejs.org) - Setup Store pattern
- Vue 3 documentation (vuejs.org) - Teleport usage

### Secondary (MEDIUM confidence)
- Tailwind CSS v4 documentation - Grid patterns
- vue-toastification GitHub README - Installation and usage

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All libraries verified, versions confirmed
- Architecture: HIGH - Patterns established in existing codebase
- Pitfalls: MEDIUM - Based on common Vue 3/Pinia issues and project-specific findings
- Backend gaps: HIGH - Direct code inspection confirms missing implementations

**Research date:** 2026-04-05
**Valid until:** 2026-05-05 (30 days for stable stack)
