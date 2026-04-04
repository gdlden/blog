# Phase 1: Shared Foundation - Research

**Researched:** 2026-04-04
**Domain:** Vue 3 + Pinia authentication and session management with Vue Router navigation
**Confidence:** HIGH

## Summary

Phase 1 focuses on stabilizing the authentication system, session state management, and creating a unified navigation shell for both blog and debt areas. The current implementation has a basic login flow but suffers from stale router-level authentication state, incomplete session persistence, and lacks a unified app shell. The research identifies Vue 3 Composition API patterns with Pinia 3.0 stores, Vue Router 4.6.4 navigation guards, and localStorage persistence as the standard stack for implementing robust session management.

**Primary recommendation:** Implement a reactive Pinia store with localStorage synchronization and Vue Router navigation guards that derive authentication state from the store rather than maintaining separate state.

## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** Frontend authentication state should use `Pinia + localStorage`, where `localStorage` provides persistence and the live app state is driven by the store rather than a router-level cached boolean.
- **D-02:** On page refresh, the app should automatically restore authenticated state if valid login information exists in `localStorage`.
- **D-03:** When an unauthenticated user or expired session reaches a protected page, the app should redirect to the login page.
- **D-04:** The shared site shell should include a clear logout entry that clears both the Pinia store and `localStorage`, then redirects to the login page.

### Claude's Discretion
- Exact store shape for persisted user/session data
- Whether auth restoration happens from a store init helper, router bootstrap, or app bootstrap
- Exact presentation of invalid-session handling as long as protected routes redirect correctly

### Deferred Ideas (OUT OF SCOPE)
None - discussion stayed within phase scope.

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| FND-01 | User can log in and remain authenticated across normal page navigation without stale route-guard state. | Vue Router navigation guards with reactive Pinia store eliminate stale state |
| FND-02 | User can log out or otherwise clear session state cleanly and return to the login page. | Pinia store logout action clears localStorage and redirects to login route |
| FND-03 | User can navigate between the blog area and debt area within one authenticated site shell. | Vue Router nested routes with shared AppLayout component provides unified navigation |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Vue | 3.5.26 | Progressive framework for building user interfaces | Official stable release, composition API standard |
| Vue Router | 4.6.4 | Official router for Vue.js | Industry standard, supports navigation guards |
| Pinia | 3.0.4 | State management library | Official Vue state management, better than Vuex |
| Axios | 1.13.3 | HTTP client for API requests | Industry standard, interceptors for auth headers |
| TypeScript | 5.9.3 | JavaScript with syntax for types | Type safety, Vue 3 official support |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| @vue/test-utils | 2.4.6 | Unit testing utilities for Vue.js | Testing Vue components and stores |
| Vitest | 4.0.18 | Unit testing framework | Frontend test execution |
| Tailwind CSS | 4.1.18 | Utility-first CSS framework | Styling layout and navigation components |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Manual localStorage sync | pinia-plugin-persistedstate | Plugin adds dependency but provides cleaner API; manual sync is sufficient for current needs |
| Vue Router 4.6.4 | Vue Router 5.0.4 | Vue Router 5 has breaking changes (removed `next()`); staying on 4.6.4 for stability |

**Installation:**
```bash
cd price_recorder_vue
pnpm install
```

**Version verification:**
```bash
npm view vue version              # 3.5.32
npm view vue-router version       # 5.0.4 (current) - project uses 4.6.4
npm view pinia version            # 3.0.4
npm view @vue/test-utils version  # 2.4.6
npm view vitest version           # 4.0.18
```

## Architecture Patterns

### Recommended Project Structure
```
src/
├── components/
│   └── AppLayout.vue          # Shared navigation shell with logout
├── composables/
│   └── useAuth.ts              # Authentication composable (optional)
├── router/
│   └── index.ts                # Route definitions with auth guards
├── stores/
│   └── userStore.ts            # Enhanced auth state management
├── view/
│   ├── Login.vue               # Login page (enhanced)
│   ├── Blog/
│   │   └── [blog views].vue    # Blog area routes
│   └── Debt/
│       └── [debt views].vue    # Debt area routes
├── api/
│   └── Login.ts               # Login API client
└── utils/
    └── request.ts              # Axios instance (existing)
```

### Pattern 1: Pinia Store with Reactive Authentication
**What:** Store uses Composition API with reactive state and computed getters for authentication status.
**When to use:** All authentication-related state and actions should live in the Pinia store.
**Example:**
```typescript
// Source: Pinia 3.0 official docs
import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'

export const useUserStore = defineStore('user', () => {
  // State
  const userInfo = ref(null)
  const token = ref('')

  // Getters
  const isAuthenticated = computed(() => {
    return !!userInfo.value && !!token.value
  })

  const userId = computed(() => {
    return userInfo.value?.userId
  })

  // Actions
  function setUserInfo(data) {
    userInfo.value = data
    token.value = data.token
    // Persist to localStorage
    localStorage.setItem('user', JSON.stringify(data))
  }

  function clearUserInfo() {
    userInfo.value = null
    token.value = ''
    localStorage.removeItem('user')
  }

  // Initialize from localStorage on store creation
  function initializeFromStorage() {
    const stored = localStorage.getItem('user')
    if (stored) {
      try {
        const parsed = JSON.parse(stored)
        userInfo.value = parsed
        token.value = parsed.token
      } catch (e) {
        console.error('Failed to parse stored user data', e)
        localStorage.removeItem('user')
      }
    }
  }

  // Watch for state changes to sync with localStorage
  watch(userInfo, (newVal) => {
    if (newVal) {
      localStorage.setItem('user', JSON.stringify(newVal))
    }
  }, { deep: true })

  return {
    userInfo,
    token,
    isAuthenticated,
    userId,
    setUserInfo,
    clearUserInfo,
    initializeFromStorage
  }
})
```

### Pattern 2: Vue Router Navigation Guard with Store Integration
**What:** Router guards check authentication status from Pinia store and redirect accordingly.
**When to use:** Protect routes that require authentication.
**Example:**
```typescript
// Source: Vue Router 4.6.4 documentation
import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/userStore'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/view/Login.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/',
      component: () => import('@/components/AppLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'home',
          component: () => import('@/view/Home.vue')
        },
        {
          path: 'blog',
          name: 'blog',
          component: () => import('@/view/Blog/BlogList.vue')
        },
        {
          path: 'debt',
          name: 'debt',
          component: () => import('@/view/Debt/DebtList.vue')
        }
      ]
    }
  ]
})

router.beforeEach((to, from, next) => {
  const userStore = useUserStore()

  // Initialize store from localStorage if not already done
  if (!userStore.userInfo && !userStore.token) {
    userStore.initializeFromStorage()
  }

  if (to.meta.requiresAuth && !userStore.isAuthenticated) {
    // Redirect to login with redirect query
    return {
      name: 'login',
      query: { redirect: to.fullPath }
    }
  } else if (to.name === 'login' && userStore.isAuthenticated) {
    // Already logged in, redirect to home
    return { name: 'home' }
  }

  return true
})

export default router
```

### Pattern 3: AppLayout Component with Logout
**What:** Shared layout component provides navigation menu and logout functionality.
**When to use:** Wrap all authenticated routes to provide consistent UI shell.
**Example:**
```vue
<template>
  <div class="app-layout">
    <nav class="sidebar">
      <div class="nav-header">
        <h1>Blog Debt Hub</h1>
      </div>
      <div class="nav-links">
        <router-link to="/" exact-active-class="active">Home</router-link>
        <router-link to="/blog" exact-active-class="active">Blog</router-link>
        <router-link to="/debt" exact-active-class="active">Debt</router-link>
      </div>
      <div class="nav-footer">
        <div class="user-info">
          <span>{{ userStore.userInfo?.username }}</span>
        </div>
        <button @click="handleLogout" class="logout-btn">Logout</button>
      </div>
    </nav>
    <main class="main-content">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import { useUserStore } from '@/stores/userStore'
import { useRouter } from 'vue-router'

const userStore = useUserStore()
const router = useRouter()

function handleLogout() {
  userStore.clearUserInfo()
  router.push('/login')
}
</script>
```

### Pattern 4: Login Page with Redirect Handling
**What:** Login page handles successful authentication and redirects to intended destination.
**When to use:** After successful login, redirect user to their original destination.
**Example:**
```vue
<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/userStore'
import { login, user } from '@/api/Login'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const isLoading = ref(false)
const errorMessage = ref('')

async function loginAction() {
  isLoading.value = true
  errorMessage.value = ''

  try {
    const response = await login(user)
    if (response && response.token) {
      userStore.setUserInfo(response)
      alert('登录成功！')

      // Redirect to intended destination or home
      const redirect = route.query.redirect as string
      router.push(redirect || '/')
    } else {
      errorMessage.value = '登录失败，请检查用户名和密码'
    }
  } catch (error) {
    console.error('Login error:', error)
    errorMessage.value = '登录失败，请稍后重试'
  } finally {
    isLoading.value = false
  }
}

function cancel() {
  user.username = ''
  user.password = ''
}
</script>
```

### Anti-Patterns to Avoid
- **Storing auth state in router:** Router guards should derive state from Pinia store, not maintain separate state.
- **Synchronous localStorage reads in render:** Initialize from localStorage in store creation, not during render.
- **Using `next()` in Vue Router 4.6.4:** Use return values (true/false/route object) instead.
- **Not clearing localStorage on logout:** Must clear both store and localStorage to prevent stale sessions.
- **Mixing auth concerns:** Keep auth logic in store, navigation logic in router, UI logic in components.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| State management | Custom reactive state with manual persistence | Pinia store with localStorage sync | Pinia provides reactivity, devtools, TypeScript support |
| Route protection | Manual route checking in components | Vue Router navigation guards | Centralized protection, cleaner code, handles edge cases |
| HTTP requests | Custom fetch wrapper | Axios with interceptors | Built-in interceptors, error handling, request/response transformation |
| Authentication | Custom token validation | Backend JWT middleware (already exists) | Backend already implements JWT validation; frontend just needs to manage token storage |

**Key insight:** The backend already has JWT middleware with whitelisted public routes. The frontend just needs to store the token, attach it to requests, and manage session state. No need to implement authentication validation logic on the frontend.

## Runtime State Inventory

> Include this section for rename/refactor/migration phases only. Omit entirely for greenfield phases.

This is a greenfield phase (new features), so runtime state inventory is not applicable.

## Common Pitfalls

### Pitfall 1: Stale Router Authentication State
**What goes wrong:** Router stores authentication state in a boolean variable (`isAuthenticated`) that is set during route guard initialization but never updates when the user logs in or out.
**Why it happens:** The current implementation in `src/router/index.ts` reads localStorage once during module initialization and caches the result in a variable.
**How to avoid:** Use reactive Pinia store state in navigation guards. The store provides computed `isAuthenticated` getter that updates when state changes.
**Warning signs:** User logs out but can still access protected routes, or user logs in but is redirected to login page.

### Pitfall 2: localStorage Synchronization Issues
**What goes wrong:** localStorage and Pinia store get out of sync, leading to inconsistent authentication state.
**Why it happens:** Manual localStorage writes/reads in multiple places without proper coordination.
**How to avoid:** Centralize localStorage operations in the Pinia store. Use `watch` to sync state changes to localStorage, and provide a single `initializeFromStorage()` method for restoration.
**Warning signs:** Page refresh shows different auth state than current session, or logout doesn't clear persisted data.

### Pitfall 3: Missing Redirect Query Parameter
**What goes wrong:** User navigates to `/blog/post/1`, gets redirected to `/login`, logs in, but ends up at `/` instead of `/blog/post/1`.
**Why it happens:** Login page doesn't check for `redirect` query parameter after successful authentication.
**How to avoid:** Store intended destination in redirect query parameter when redirecting unauthenticated users to login. Read and use it after successful login.
**Warning signs:** Users report losing their place after login.

### Pitfall 4: Incorrect Router Guard Return Values
**What goes wrong:** Navigation guard doesn't properly handle redirect logic or causes infinite loops.
**Why it happens:** Using deprecated `next()` callback or incorrect return values in Vue Router 4.6.4.
**How to avoid:** Use return values: `true` to proceed, `false` to cancel, route object/string to redirect. Don't use `next()`.
**Warning signs:** Browser hangs, console shows "Maximum call stack size exceeded", or navigation doesn't work.

### Pitfall 5: Axios Interceptor Reading Stale Data
**What goes wrong:** Request interceptor reads from localStorage directly, which may be stale or missing data.
**Why it happens:** Current implementation in `src/utils/request.ts` reads localStorage for every request without validation.
**How to avoid:** Interceptor should read from Pinia store, which provides reactive state with validation. Or enhance localStorage reading with proper error handling and fallback.
**Warning signs:** Requests fail with 401 errors despite user being logged in.

## Code Examples

Verified patterns from official sources:

### Pinia Store with Authentication State
```typescript
// Source: Pinia 3.0 official documentation
import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'

export const useUserStore = defineStore('user', () => {
  const userInfo = ref(null)
  const token = ref('')

  const isAuthenticated = computed(() => !!userInfo.value && !!token.value)

  function setUserInfo(data) {
    userInfo.value = data
    token.value = data.token
    localStorage.setItem('user', JSON.stringify(data))
  }

  function clearUserInfo() {
    userInfo.value = null
    token.value = ''
    localStorage.removeItem('user')
  }

  function initializeFromStorage() {
    const stored = localStorage.getItem('user')
    if (stored) {
      try {
        const parsed = JSON.parse(stored)
        userInfo.value = parsed
        token.value = parsed.token
      } catch (e) {
        localStorage.removeItem('user')
      }
    }
  }

  return { userInfo, token, isAuthenticated, setUserInfo, clearUserInfo, initializeFromStorage }
})
```

### Vue Router Navigation Guard
```typescript
// Source: Vue Router 4.6.4 official documentation
router.beforeEach((to, from) => {
  const userStore = useUserStore()

  if (to.meta.requiresAuth && !userStore.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  } else if (to.name === 'login' && userStore.isAuthenticated) {
    return { name: 'home' }
  }

  return true
})
```

### Axios Request Interceptor
```typescript
// Source: Axios official documentation
instance.interceptors.request.use(
  config => {
    const userStore = useUserStore()
    if (userStore.token) {
      config.headers.Authorization = `Bearer ${userStore.token}`
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Vuex | Pinia | 2022 | Simpler API, better TypeScript support, no mutations |
| Vue Router Options API | Vue Router Composition API | 2021 | Better composability, easier to share logic |
| Manual state management | Pinia with localStorage sync | 2022 | Automatic persistence, cleaner code |

**Deprecated/outdated:**
- **Vuex:** Replaced by Pinia as official Vue state management solution
- **Vue Router `next()` callback:** Deprecated in Vue Router 4, use return values instead
- **Options API in stores:** Composition API with setup syntax is now recommended

## Open Questions

None - all required research completed with HIGH confidence.

## Environment Availability

> Skip this section if the phase has no external dependencies (code/config-only changes).

This phase has external dependencies:

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Node.js | Frontend runtime | ✓ | v20.19.0+ | — |
| pnpm | Package manager | ✓ | 10.28.0 | npm (project uses pnpm) |
| PostgreSQL | Backend database | ? | — | Backend not modified in this phase |
| Kratos backend | Authentication API | ? | — | Backend already provides login API |

**Missing dependencies with no fallback:**
- None detected

**Missing dependencies with fallback:**
- PostgreSQL: Not required for frontend-only changes in this phase
- Kratos backend: Not modified in this phase; existing API is sufficient

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Vitest 4.0.18 |
| Config file | vitest.config.ts |
| Quick run command | `cd price_recorder_vue && pnpm test:unit` |
| Full suite command | `cd price_recorder_vue && pnpm test:unit` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| FND-01 | Login maintains auth state across navigation | integration | `pnpm test:unit` | ❌ Wave 0 |
| FND-01 | Auth state updates reactively on login/logout | unit | `pnpm test:unit` | ❌ Wave 0 |
| FND-02 | Logout clears store and localStorage | unit | `pnpm test:unit` | ❌ Wave 0 |
| FND-02 | Logout redirects to login page | unit | `pnpm test:unit` | ❌ Wave 0 |
| FND-03 | Navigation menu provides access to blog and debt | component | `pnpm test:unit` | ❌ Wave 0 |
| FND-03 | Router guards protect authenticated routes | unit | `pnpm test:unit` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `cd price_recorder_vue && pnpm test:unit`
- **Per wave merge:** `cd price_recorder_vue && pnpm test:unit && pnpm lint && pnpm type-check`
- **Phase gate:** Full suite green, lint passes, type-check passes before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `src/__tests__/stores/userStore.spec.ts` - covers FND-01, FND-02 store behavior
- [ ] `src/__tests__/router/index.spec.ts` - covers FND-01, FND-02, FND-03 navigation guard behavior
- [ ] `src/__tests__/components/AppLayout.spec.ts` - covers FND-03 navigation menu and logout
- [ ] `src/__tests__/view/Login.spec.ts` - covers FND-01 login flow with redirect handling
- [ ] Test utilities for mocking localStorage and Vue Router

## Sources

### Primary (HIGH confidence)
- Pinia 3.0 official documentation - https://pinia.vuejs.org/core-concepts/state.html
- Vue Router 4.6.4 official documentation - https://router.vuejs.org/guide/advanced/navigation-guards.html
- Vue 3.5 official documentation - https://vuejs.org/guide/introduction.html
- Axios official documentation - https://axios-http.com/docs/interceptors
- @vue/test-utils documentation - https://test-utils.vuejs.org/
- Vitest documentation - https://vitest.dev/

### Secondary (MEDIUM confidence)
- Web search results for Vue 3 authentication patterns (2025)
- Web search results for Pinia localStorage persistence (2025)
- Web search results for Vue Router navigation guard best practices (2025)
- Web search results for Vitest testing Vue Router and Pinia (2025)

### Tertiary (LOW confidence)
- None - all findings verified with official documentation or multiple sources

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All versions verified via npm registry, official documentation confirms current best practices
- Architecture: HIGH - All patterns verified with official Vue/Pinia/Router documentation
- Pitfalls: HIGH - Identified through code review of current implementation and verified against official best practices
- Testing: MEDIUM - Testing infrastructure exists, but test patterns for Vue Router guards need verification during implementation

**Research date:** 2026-04-04
**Valid until:** 2026-05-04 (30 days for stable ecosystem, Vue/Pinia/Router are mature technologies)
