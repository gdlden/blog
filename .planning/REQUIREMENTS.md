# Milestone v1.1 Requirements

## API Standardization

- [ ] **API-01**: All backend HTTP endpoints return a unified response structure containing `code`, `message`, and `data` fields.
- [ ] **API-02**: Successful responses use `code: "200"` (or numeric equivalent) with the payload in `data`.
- [ ] **API-03**: Error responses include a non-success `code` and a human-readable `message`.
- [ ] **API-04**: The response wrapper is applied at the Kratos transport layer so individual handlers do not manually construct it.
- [ ] **API-05**: Frontend Axios interceptor is updated to expect and correctly parse the unified wrapper across all modules (blog, debt, user).

## Frontend Testing

- [ ] **TEST-01**: Critical auth flows (login, logout, route guards) have automated Vitest coverage.
- [ ] **TEST-02**: Blog CRUD flows (create, read, update, delete) have automated Vitest coverage.
- [ ] **TEST-03**: Debt CRUD flows (create, read, update, delete) have automated Vitest coverage.
- [ ] **TEST-04**: Navigation and layout components (AppLayout, router) have automated Vitest coverage.

## Contributor Documentation

- [ ] **DOC-01**: README or CONTRIBUTING doc explains how to run `make api` and `make config` in `blog/`.
- [ ] **DOC-02**: README or CONTRIBUTING doc explains how to run backend tests (`go test ./...`).
- [ ] **DOC-03**: README or CONTRIBUTING doc explains how to run frontend tests (`npm run test:unit`).
- [ ] **DOC-04**: Document covers installing proto tools (`make init`) and regenerating Wire (`cd cmd/blog && wire`).

---

## Traceability

| REQ-ID | Phase |
|--------|-------|
| API-01 | 8 |
| API-02 | 8 |
| API-03 | 8 |
| API-04 | 8 |
| API-05 | 8 |
| TEST-01 | 9 |
| TEST-02 | 9 |
| TEST-03 | 9 |
| TEST-04 | 9 |
| DOC-01 | 10 |
| DOC-02 | 10 |
| DOC-03 | 10 |
| DOC-04 | 10 |
