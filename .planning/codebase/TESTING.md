# Testing

**Analysis Date:** 2026-03-26

## Test Tooling

**Backend:**
- Go uses the standard `testing` package.
- Repository guidance says backend verification should use `go test ./...` from `blog/`.

**Frontend:**
- Vitest is configured in `price_recorder_vue/vitest.config.ts`.
- The frontend test environment is `jsdom`.
- Vue Test Utils is present for component mounting and assertions.

## Test Commands

**Backend commands:**
- `cd blog && go test ./...`
- `cd blog && make build`
- `cd blog && make api && make config`

**Frontend commands:**
- `cd price_recorder_vue && pnpm test:unit`
- `cd price_recorder_vue && pnpm build`
- `cd price_recorder_vue && pnpm lint`

**Cross-project expectation:**
- When a change spans backend and frontend, repository guidance expects both `go test ./...` and `pnpm test:unit` to run.

## Test Layout

**Backend test files detected:**
- `blog/internal/data/geo_test.go`
- `blog/internal/data/hzszgx_file_ocr_test.go`
- `blog/internal/data/wjln_db_data_test.go`

**Frontend test files detected:**
- `price_recorder_vue/src/__tests__/App.spec.ts`

**Observations:**
- Backend tests are concentrated in the data layer.
- No service-layer or biz-layer unit tests were detected.
- Frontend has a single example-style component test.

## Current Coverage

**Covered areas:**
- Some backend data/integration behavior appears to be exercised through the `blog/internal/data/*_test.go` files.
- Frontend component mounting is minimally exercised by `price_recorder_vue/src/__tests__/App.spec.ts`.

**Thin or missing coverage:**
- Backend service handlers in `blog/internal/service/*.go`
- Backend use cases in `blog/internal/biz/*.go`
- HTTP middleware behavior in `blog/internal/server/http.go`
- Auth context extraction in `blog/internal/utils/userutil.go`
- Frontend router guards in `price_recorder_vue/src/router/index.ts`
- Frontend stores in `price_recorder_vue/src/stores/userStore.ts`
- Frontend API wrappers and Axios interceptor behavior in `price_recorder_vue/src/api/*.ts` and `price_recorder_vue/src/utils/request.ts`

## Reliability of Existing Tests

**Backend reliability concerns:**
- Several backend tests appear integration-heavy and reference live credentials, tokens, local paths, or external services.
- This makes them poor candidates for deterministic CI without isolation or fixtures.
- The backend also depends on generated protobuf outputs that are not fully committed, which can block `go test ./...` in a clean checkout before tests even run.

**Frontend reliability concerns:**
- `price_recorder_vue/src/__tests__/App.spec.ts` still expects the text `You did it!`, but `price_recorder_vue/src/App.vue` only renders `<router-view>`.
- That means the current frontend test is stale and not aligned with real app behavior.

## Suggested Test Pyramid

**Backend priority:**
- Add unit tests for service request parsing and reply mapping.
- Add use-case tests for ownership checks and domain validation.
- Add repo tests that use isolated test databases or fixtures rather than live endpoints.
- Keep external integration tests behind explicit opt-in commands or build tags.

**Frontend priority:**
- Replace the stale `App.spec.ts` example test with router-aware tests for the actual app shell.
- Add tests for login flow, route guard behavior, and token persistence.
- Add tests around the Axios response/request interceptors.

## Verification Readiness

**What is ready now:**
- Tooling for frontend lint/type/test exists.
- A documented command path exists for backend and frontend verification.

**What is not ready yet:**
- Deterministic backend CI coverage is weak.
- Frontend regression coverage is too small to protect active behavior.
- Cross-project end-to-end verification is not present.

## Overall Assessment

**Testing maturity:** Low

**Reasoning:**
- The repo has test tooling configured, but not a dependable automated safety net.
- Most critical business, auth, and routing behavior remains untested.
- Existing backend tests lean toward environment-dependent integration work rather than fast repeatable unit coverage.

---

*Testing analysis: 2026-03-26*
