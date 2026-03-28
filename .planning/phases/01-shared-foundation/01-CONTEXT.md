# Phase 1: Shared Foundation - Context

**Gathered:** 2026-03-26
**Status:** Ready for planning

<domain>
## Phase Boundary

This phase stabilizes authentication, session handling, and the unified app shell so the same personal account can reliably access both the blog area and the debt area. It does not add new blog or debt capabilities beyond what is necessary to make the shared foundation dependable.

</domain>

<decisions>
## Implementation Decisions

### Authentication and session state
- **D-01:** Frontend authentication state should use `Pinia + localStorage`, where `localStorage` provides persistence and the live app state is driven by the store rather than a router-level cached boolean.
- **D-02:** On page refresh, the app should automatically restore authenticated state if valid login information exists in `localStorage`.
- **D-03:** When an unauthenticated user or expired session reaches a protected page, the app should redirect to the login page.
- **D-04:** The shared site shell should include a clear logout entry that clears both the Pinia store and `localStorage`, then redirects to the login page.

### the agent's Discretion
- Exact store shape for persisted user/session data
- Whether auth restoration happens from a store init helper, router bootstrap, or app bootstrap
- Exact presentation of invalid-session handling as long as protected routes redirect correctly

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project and phase scope
- `.planning/PROJECT.md` - overall product goals, single-user scope, and blog-first priority
- `.planning/REQUIREMENTS.md` - Phase 1 requirements `FND-01`, `FND-02`, and `FND-03`
- `.planning/ROADMAP.md` - Phase 1 goal and success criteria for the shared foundation
- `.planning/STATE.md` - current workflow state and next-step context

### Existing implementation references
- `price_recorder_vue/src/router/index.ts` - current route table and stale auth-guard implementation
- `price_recorder_vue/src/stores/userStore.ts` - current Pinia store and browser persistence behavior
- `price_recorder_vue/src/utils/request.ts` - current request auth-header injection from persisted user info
- `price_recorder_vue/src/view/Login.vue` - current login page and post-login redirect behavior
- `blog/internal/service/user.go` - backend login reply shape and token issuance behavior

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `price_recorder_vue/src/stores/userStore.ts`: existing Pinia store can be evolved into the canonical session source for the SPA.
- `price_recorder_vue/src/utils/request.ts`: existing Axios request interceptor already reads persisted user data and can be aligned with the finalized auth state approach.
- `price_recorder_vue/src/view/Login.vue`: existing login screen and redirect flow provide the starting point for stable session entry behavior.

### Established Patterns
- Frontend session persistence currently relies on `localStorage`.
- Frontend routing is centralized in `price_recorder_vue/src/router/index.ts` with a global `beforeEach` guard.
- The backend returns a token plus user payload from `UserLogin`, so the frontend already has a workable login response contract.

### Integration Points
- Phase 1 work will connect through the Vue router, Pinia user store, login view, and shared Axios client.
- Shared navigation introduced in this phase should become the shell used later by the blog and debt feature pages.
- Logout behavior will need to coordinate store clearing, browser persistence clearing, and route navigation.

</code_context>

<specifics>
## Specific Ideas

- The user explicitly wants auth state restored after refresh when persisted login data is present.
- The user explicitly prefers redirecting unauthenticated access to the login page.
- The user explicitly wants a visible logout entry rather than relying only on implicit session loss.

</specifics>

<deferred>
## Deferred Ideas

None - discussion stayed within phase scope.

</deferred>

---

*Phase: 01-shared-foundation*
*Context gathered: 2026-03-26*
