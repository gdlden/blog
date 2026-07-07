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
| 11 | OCR Debt Detail Recognition | Allow users to upload a screenshot of debt installment details and automatically extract period, principal, interest, and posting date to pre-fill the debt detail form | OCR-01, OCR-02, OCR-03, OCR-04 |
| 13 | 3/3 | Complete   | 2026-07-07 |

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
### Phase 11: OCR Debt Detail Recognition

**Goal:** Allow users to upload a screenshot of debt installment details and automatically extract period, principal, interest, and posting date to pre-fill the debt detail form.

**Requirements:** OCR-01, OCR-02, OCR-03, OCR-04

**Success Criteria:**

1. User can upload/select an image in the debt detail page
2. OCR extracts text from the image accurately enough to parse structured data
3. Parsed data (period, principal, interest, posting date) pre-fills the create detail form
4. User can review and edit before saving
5. Works with common Chinese bank/loan app installment detail screenshots

**UI hint:** yes

### Phase 12: ocr识别新增个deepseek的

**Goal:** Replace Kimi with DeepSeek as the OCR provider for debt detail screenshot recognition
**Requirements**: TBD
**Depends on:** Phase 11
**Plans:** 1 plan

Plans:

- [x] 12-01: Replace Kimi with DeepSeek VisionTextRecognizer

### Phase 13: Fishing Spot Map

**Goal:** Save and manage GPS-tagged favorite locations — capture current position, name spots, and find them later. Primary use case: fishing spots.

**Requirements:** MAP-01, MAP-02, MAP-03

**Plans:** 3/3 plans complete

Plans:

- [x] 13-01-PLAN.md — Map proto contract + GORM data layer + MapUsecase with Gaode regeo (shipped)
- [x] 13-02-PLAN.md — MapService proto handler + Wire DI + JWT-protected HTTP routes (shipped)
- [x] 13-03-PLAN.md — Frontend resumption: Gaode map view + capture bottom-sheet + spot CRUD UI + nav deep-link + Vitest coverage (closes 4 VERIFICATION gaps + registers MAP reqs)

**Success Criteria:**

1. User can capture and save current GPS coordinates with a name/note
2. User can view, edit, and delete saved spots
3. User can view spots on a map and navigate back to them
4. Data persists reliably and works on mobile

**UI hint:** yes

---

*Roadmap created: 2026-03-26*
*Updated: 2026-04-26 - Added phase 11 OCR Debt Detail Recognition*
*Updated: 2026-07-05 - Added phase 13 Fishing Spot Map*
*Updated: 2026-07-07 - Phase 13 plan 13-03 complete; 3/3 plans shipped*
