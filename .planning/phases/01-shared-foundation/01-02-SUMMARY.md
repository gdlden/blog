---
phase: 01-shared-foundation
plan: "02"
subsystem: frontend
tags: [auth, router, login, pinia, tests]
dependency:
  requires: ["01-01"]
  provides: ["reactive-router-guards", "redirect-aware-login"]
  affects: ["01-03"]
tech-stack:
  added: []
  patterns:
    - "Vue Router beforeEach guards derive auth state from Pinia store"
    - "Login page reads route.query.redirect and pushes on success"
    - "Unit tests using Vitest + @vue/test-utils with mocked vue-router"
key-files:
  created:
    - price_recorder_vue/src/__tests__/router/index.spec.ts
    - price_recorder_vue/src/__tests__/view/Login.spec.ts
  modified:
    - price_recorder_vue/src/router/index.ts
    - price_recorder_vue/src/view/Login.vue
decisions:
  - "Router guards must initialize the store from localStorage inside beforeEach when isAuthenticated is false"
  - "Router uses route meta.requiresAuth instead of a cached boolean"
  - "Login page uses TypeScript (lang=ts) and imports from @/api/Login (not .js)"
metrics:
  duration-minutes: 12
  completed-date: "2026-04-04"
---

# Phase 01 Plan 02: Reactive Router Guards and Login Redirect Summary

## One-liner
Eliminated the stale router-level authentication boolean and rewrote both router guards and the login page so auth state is reactive via Pinia, redirect queries work correctly, and both have passing unit tests.

## What Changed

### Router (`price_recorder_vue/src/router/index.ts`)
- Removed `let isAuthenticated = false` and localStorage parsing at module load time.
- Added `meta: { requiresAuth: true }` to protected routes (`/` and `/test`).
- `beforeEach` now uses `useUserStore()`:
  - Calls `userStore.initializeFromStorage()` when `!userStore.isAuthenticated`.
  - Redirects unauthenticated users to `/login` with `?redirect=` query.
  - Redirects authenticated users away from `/login` to `/`.

### Login Page (`price_recorder_vue/src/view/Login.vue`)
- Converted to `<script setup lang="ts">` and imports from `@/api/Login` (was `.js`).
- Added `useRoute` to read `route.query.redirect` and `useRouter` for navigation.
- Added `isLoading` and `errorMessage` reactive state.
- On successful login: calls `userStore.setUserInfo(response)` then `router.push(redirect || '/')`.
- On failure: shows the exact UI-SPEC error copy: "登录失败，请检查用户名或密码后重试。".
- Applied UI-SPEC styling: `bg-gray-100` outer, `bg-white` card, `bg-blue-600` primary button, focus rings.
- Removed `animate-pulse` classes.

### Tests
- `src/__tests__/router/index.spec.ts`: 4 tests covering redirect-with-query, login-redirect-away, protected access, and localStorage initialization.
- `src/__tests__/view/Login.spec.ts`: 4 tests covering render, API call, success redirect, and failure error message.

## Commits

| Hash | Message |
|------|---------|
| `58fb6b3` | test(01-shared-foundation-02): add failing tests for router guards and Login.vue |
| `bdfaa4b` | feat(01-shared-foundation-02): rewrite router guards to use reactive Pinia auth state |
| `cd2b097` | feat(01-shared-foundation-02): rewrite Login.vue with redirect handling, error state, and UI-SPEC styling |

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

None.

## Self-Check: PASSED

- All created files exist: `price_recorder_vue/src/__tests__/router/index.spec.ts`, `price_recorder_vue/src/__tests__/view/Login.spec.ts`
- All modified files exist: `price_recorder_vue/src/router/index.ts`, `price_recorder_vue/src/view/Login.vue`
- All 8 unit tests pass.
