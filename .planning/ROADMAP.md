# Roadmap: Blog Debt Hub

**Mode:** yolo | **Granularity:** standard

## Milestones

- ✅ **v1.0 MVP** — Phases 1, 5-7 (shipped 2026-04-13) — [Archive](milestones/v1.0-ROADMAP.md)
- **v1.1 Polish & Standardization** — Phases 8-10 (in progress)

## Phases

<details>
<summary>✅ v1.0 MVP — SHIPPED 2026-04-13</summary>

- [x] **Phase 1: Shared Foundation** (3/3 plans) — Auth state, router guards, app shell
- [x] **Phase 5: Database Configuration Enhancement** (1/1 plan) — Env-var database source override
- [x] **Phase 6: Frontend-Backend Integration** (4/4 plans) — Blog and debt CRUD with toast notifications
- [x] **Phase 7: Backend Unit Tests** (3/3 plans) — Data, biz, and service layer unit tests

*Full details archived in [milestones/v1.0-ROADMAP.md](milestones/v1.0-ROADMAP.md)*
</details>

### Planned

- [ ] **Phase 8: API Response Standardization** — Unify all backend HTTP responses under a consistent `{code, message, data}` wrapper via Kratos response encoding, and align the frontend Axios interceptor.
  - Requirements: API-01, API-02, API-03, API-04, API-05
  - Success criteria:
    1. All endpoints (post, debt, user, etc.) return `{code, message, data}`
    2. Success responses have `code: "200"` and payload in `data`
    3. Error responses include non-success code and human-readable message
    4. Wrapper is applied at transport layer, not in individual handlers
    5. Frontend interceptor correctly parses the unified wrapper

- [ ] **Phase 9: Frontend Automated Tests** — Add Vitest coverage for critical auth, navigation, blog CRUD, and debt CRUD flows.
  - Requirements: TEST-01, TEST-02, TEST-03, TEST-04
  - Success criteria:
    1. Login/logout and route-guard flows have tests
    2. Blog create/read/update/delete flows have tests
    3. Debt create/read/update/delete flows have tests
    4. AppLayout and router navigation have tests
    5. All new tests pass in CI/local `npm run test:unit`

- [ ] **Phase 10: Contributor Documentation** — Document the generated-code workflow and verification commands so contributors can set up and test reliably.
  - Requirements: DOC-01, DOC-02, DOC-03, DOC-04
  - Success criteria:
    1. README/CONTRIBUTING explains `make init`, `make api`, `make config`
    2. README/CONTRIBUTING explains `cd cmd/blog && wire`
    3. README/CONTRIBUTING explains `go test ./...`
    4. README/CONTRIBUTING explains `npm run test:unit`
    5. New contributor can follow docs to run full verification from clean clone

---
*Updated: 2026-04-14 after v1.1 milestone start*
