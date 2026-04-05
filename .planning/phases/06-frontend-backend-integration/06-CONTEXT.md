# Phase 6: Frontend-Backend Integration - Context

**Gathered:** 2026-04-05
**Status:** Ready for planning

<domain>
## Phase Boundary

Complete frontend integration with backend blog and debt APIs, enabling full CRUD operations for both modules through the Vue frontend. This phase bridges the gap between backend API implementations and frontend user interface.

Scope includes:
- Blog module: List, view, create, edit, delete posts
- Debt module: List, view, create, edit, delete debt records
- API client layer, state management, UI components, error handling

</domain>

<decisions>
## Implementation Decisions

### Blog Edit/Delete API
- **D-01:** Enable full CRUD for blog posts (not read-only)
- **D-02:** Uncomment and implement EditPostById and DeletePostById in proto
- **D-03:** Add Update and Delete methods to PostService and PostUsecase
- **D-04:** Ensure backend has complete CRUD before frontend integration

### API Client Organization
- **D-05:** Organize by domain: `blog.ts` and `debt.ts` in `src/api/`
- **D-06:** Each file exports functions for its domain's CRUD operations
- **D-07:** Use existing Axios instance from `src/utils/request.ts`

### State Management Strategy
- **D-08:** Use Pinia stores: `useBlogStore` and `useDebtStore`
- **D-09:** Stores manage: list data, current detail, loading states, errors
- **D-10:** Components subscribe to stores for reactive updates

### List Page Design
- **D-11:** Use card grid layout for both blog and debt lists
- **D-12:** Cards display key information with action buttons
- **D-13:** Include pagination controls at bottom of list

### Form Interaction
- **D-14:** Use Modal dialogs for create/edit forms
- **D-15:** Forms open without leaving list page (context preserved)
- **D-16:** Modal has close button, cancel button, and save/submit button

### Error Handling & Feedback
- **D-17:** Use Toast notifications for success/error feedback
- **D-18:** Toast auto-dismisses after 3-5 seconds
- **D-19:** Form validation errors show inline near fields

### Claude's Discretion
- Exact card styling and spacing
- Toast library choice (vue-toastification vs custom)
- Modal component implementation details
- Form field validation rules

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Backend APIs
- `blog/api/post/v1/post.proto` — Post service definitions (Edit/Delete commented)
- `blog/api/debt/v1/debt.proto` — Debt service definitions (complete CRUD)
- `blog/internal/service/post.go` — Post service implementation
- `blog/internal/biz/post.go` — Post usecase (missing Update/Delete)

### Frontend Structure
- `price_recorder_vue/src/view/BlogList.vue` — Current placeholder
- `price_recorder_vue/src/view/DebtList.vue` — Current placeholder
- `price_recorder_vue/src/api/Article.ts` — Existing API pattern
- `price_recorder_vue/src/utils/request.ts` — Axios instance
- `price_recorder_vue/src/stores/` — Pinia stores directory

### Project Conventions
- `CLAUDE.md` — Frontend/backend development guidelines

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- **Axios instance** (`request.ts`): Already configured with auth interceptor
- **AppLayout** (`AppLayout.vue`): Navigation shell ready
- **Login store** (`userStore.ts`): Pattern for Pinia store implementation
- **API pattern** (`Article.ts`): Template for new API clients

### Established Patterns
- **Kratos HTTP APIs**: RESTful endpoints generated from proto
- **Vue 3 Composition API**: `<script setup>` pattern
- **Tailwind CSS**: Utility-first styling already in use
- **Pinia**: State management already configured in main.ts

### Integration Points
- `BlogList.vue` and `DebtList.vue` need full implementation
- New API files: `src/api/blog.ts`, `src/api/debt.ts`
- New stores: `src/stores/blogStore.ts`, `src/stores/debtStore.ts`
- Backend needs: Post Update/Delete in service, biz, data layers

</code_context>

<specifics>
## Specific Ideas

- Blog cards show: title (truncated), content preview, created date, edit/delete buttons
- Debt cards show: name, bank, amount, status, apply date, action buttons
- List pages have "Create" button in header
- Modals have form fields matching backend proto messages
- Toast shows: "Created successfully", "Updated successfully", "Deleted successfully", or error message

</specifics>

<deferred>
## Deferred Ideas

- Image upload for blog posts — future enhancement
- Advanced filtering/search — future enhancement
- Bulk operations (delete multiple) — future enhancement
- Real-time updates via WebSocket — out of scope

</deferred>

---

*Phase: 06-frontend-backend-integration*
*Context gathered: 2026-04-05*
