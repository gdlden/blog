# Roadmap: Blog Debt Hub

**Created:** 2026-03-26
**Mode:** yolo
**Granularity:** standard
**Coverage:** 19 / 19 v1 requirements mapped

## Overview

| Phase | Name | Goal | Requirements |
|-------|------|------|--------------|
| 1 | Shared Foundation | Stabilize authentication, session handling, and the unified app shell for both functional areas | FND-01, FND-02, FND-03 |
| 2 | Blog Core Workflow | Make the personal blog experience complete and dependable end to end | BLOG-01, BLOG-02, BLOG-03, BLOG-04, BLOG-05, BLOG-06 |
| 3 | Debt Management and Statistics | Make debt records, debt details, and debt summaries dependable for daily personal tracking | DEBT-01, DEBT-02, DEBT-03, DEBT-04, DEBT-05, DEBT-06, DEBT-07 |
| 4 | Verification and Delivery Hardening | Add enough backend/frontend verification and workflow clarity to support safe iteration | QUAL-01, QUAL-02, QUAL-03 |
| 5 | Database Configuration Enhancement | Enable database connection source to be injected via environment variables for flexible deployment | CFG-01 |
| 6 | Frontend-Backend Integration | Complete frontend integration with backend blog and debt APIs for full CRUD operations | INT-01, INT-02 |

## Phase Details

### Phase 1: Shared Foundation

**Goal:** Stabilize authentication, session state, and unified navigation so the same account can reliably access both the blog and debt areas.

**Requirements:** FND-01, FND-02, FND-03

**Plans:** 3/3 plans complete

Plans:
- [x] `01-01-PLAN.md` — Refactor Pinia auth store and align Axios interceptor with reactive store state
- [x] `01-02-PLAN.md` — Replace stale router auth boolean with reactive guards and update Login.vue for redirects/errors
- [x] `01-03-PLAN.md` — Create shared AppLayout with blog/debt navigation and wire authenticated routes

**Success Criteria:**
1. Login success and failure states are handled cleanly in the frontend.
2. Auth state is not cached incorrectly by the router after login/logout.
3. The user can reach both the blog and debt sections from one authenticated shell.
4. Shared request/session behavior is explicit enough to support later blog and debt phases.

**UI hint:** yes

### Phase 2: Blog Core Workflow

**Goal:** Turn the blog area into the strongest part of the product by completing the user-facing write, browse, and manage flows.

**Requirements:** BLOG-01, BLOG-02, BLOG-03, BLOG-04, BLOG-05, BLOG-06

**Success Criteria:**
1. The frontend presents a usable blog list and detail flow backed by current APIs.
2. The user can create, edit, and delete posts through the application interface.
3. Blog management works from an authenticated area rather than ad hoc testing pages.
4. Backend and frontend behavior for blog flows is consistent and verified.

**UI hint:** yes

### Phase 3: Debt Management and Statistics

**Goal:** Make the debt area useful for everyday personal tracking by completing CRUD reliability and exposing core summary metrics.

**Requirements:** DEBT-01, DEBT-02, DEBT-03, DEBT-04, DEBT-05, DEBT-06, DEBT-07

**Success Criteria:**
1. Debt record CRUD works end to end without stubbed or misleading success paths.
2. Debt-detail behavior supports accurate record history and summary calculations.
3. The user can view total debt, repaid amount, outstanding amount, and per-record breakdowns.
4. Debt flows remain aligned with the single-user product scope.

**UI hint:** yes

### Phase 4: Verification and Delivery Hardening

**Goal:** Reduce change risk by adding targeted tests and making project verification/codegen steps reliable.

**Requirements:** QUAL-01, QUAL-02, QUAL-03

**Success Criteria:**
1. Backend tests cover critical blog or debt service/domain behavior.
2. Frontend tests cover auth/navigation and at least one real product workflow.
3. The generated-code or setup workflow is documented clearly enough for repeatable verification.
4. Core validation commands can be run intentionally as part of ongoing work.

**UI hint:** no

## Notes

- Blog is intentionally scheduled ahead of debt because it is the declared v1 priority.
- Shared auth stability comes first because both product areas depend on it.
- Debt visualization, reminders, payments, and collaboration stay out of the current roadmap.

### Phase 5: Database Configuration Enhancement

**Goal:** Enable database connection source to be injected via environment variables for flexible deployment and configuration management.

**Requirements:** CFG-01

**Success Criteria:**
1. Database source can be configured via environment variable (e.g., `DATABASE_URL` or `DB_SOURCE`).
2. Environment variable takes precedence over YAML config file when present.
3. Existing config.yaml behavior remains backward compatible.
4. Clear documentation on supported environment variables and their format.

**UI hint:** no

### Phase 6: Frontend-Backend Integration

**Goal:** Complete frontend integration with backend blog and debt APIs, enabling full CRUD operations for both modules through the Vue frontend.

**Requirements:** INT-01, INT-02

**Plans:** 4/4 plans complete

Plans:
- [x] `06-01-PLAN.md` — Backend: Enable Post Update/Delete APIs (uncomment proto, implement service/biz/data layers)
- [x] `06-02-PLAN.md` — Frontend: Blog module (API client, Pinia store, BlogList view with CRUD)
- [x] `06-03-PLAN.md` — Frontend: Debt module (API client, Pinia store, DebtList view with CRUD and summary)
- [x] `06-04-PLAN.md` — Shared: Toast notifications integration for user feedback

**Success Criteria:**
1. Blog module: Frontend can list, view, create, edit, and delete blog posts via backend APIs.
2. Debt module: Frontend can list, view, create, edit, and delete debt records via backend APIs.
3. Proper error handling and loading states for all API operations.
4. Consistent data flow between frontend stores and backend services.

**UI hint:** yes

---
*Roadmap created: 2026-03-26*
