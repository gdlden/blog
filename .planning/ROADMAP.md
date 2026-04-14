# Roadmap: Blog Debt Hub

**Created:** 2026-03-26
**Mode:** yolo
**Granularity:** standard
**Coverage:** 13 / 13 v1.1 requirements mapped

## Overview

| Phase | Name | Goal | Requirements |
|-------|------|------|--------------|
| 8 | API Response Standardization | Unify all backend HTTP responses under a consistent `{code, message, data}` wrapper via Kratos response encoding, and align the frontend Axios interceptor | API-01, API-02, API-03, API-04, API-05 |
| 9 | Frontend Automated Tests | Add Vitest coverage for critical auth, navigation, blog CRUD, and debt CRUD flows | TEST-01, TEST-02, TEST-03, TEST-04 |
| 10 | Contributor Documentation | Document the generated-code workflow and verification commands so contributors can set up and test reliably | DOC-01, DOC-02, DOC-03, DOC-04 |

## Phase Details

### Phase 8: API Response Standardization
**Plans:** 2 plans

Plans:
- [ ] 08-01-PLAN.md — Backend Kratos transport-layer response wrapper
- [ ] 08-02-PLAN.md — Frontend Axios interceptor and API client alignment

**Goal:** Unify all backend HTTP responses under a consistent `{code, message, data}` wrapper via Kratos response encoding, and align the frontend Axios interceptor.

**Requirements:** API-01, API-02, API-03, API-04, API-05

**Success Criteria:**
1. All endpoints (post, debt, user, etc.) return `{code, message, data}`
2. Success responses have `code: "200"` and payload in `data`
3. Error responses include non-success code and human-readable message
4. Wrapper is applied at transport layer, not in individual handlers
5. Frontend interceptor correctly parses the unified wrapper

**UI hint:** no

### Phase 9: Frontend Automated Tests

**Goal:** Add Vitest coverage for critical auth, navigation, blog CRUD, and debt CRUD flows.

**Requirements:** TEST-01, TEST-02, TEST-03, TEST-04

**Success Criteria:**
1. Login/logout and route-guard flows have tests
2. Blog create/read/update/delete flows have tests
3. Debt create/read/update/delete flows have tests
4. AppLayout and router navigation have tests
5. All new tests pass in CI/local `npm run test:unit`

**UI hint:** no

### Phase 10: Contributor Documentation

**Goal:** Document the generated-code workflow and verification commands so contributors can set up and test reliably.

**Requirements:** DOC-01, DOC-02, DOC-03, DOC-04

**Success Criteria:**
1. README/CONTRIBUTING explains `make init`, `make api`, `make config`
2. README/CONTRIBUTING explains `cd cmd/blog && wire`
3. README/CONTRIBUTING explains `go test ./...`
4. README/CONTRIBUTING explains `npm run test:unit`
5. New contributor can follow docs to run full verification from clean clone

**UI hint:** no

---
*Roadmap created: 2026-03-26*
*Updated: 2026-04-14 - Added v1.1 phases 8-10*
