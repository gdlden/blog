# Codebase Concerns

**Analysis Date:** 2026-03-25

## Tech Debt

**Debt detail CRUD is only partially implemented:**
- Issue: update/delete methods are stubbed in repo and service layers return empty success replies without calling usecase logic.
- Files: `blog/internal/data/debtDetail.go`, `blog/internal/service/debtdetail.go`, `blog/internal/biz/debtDetail.go`
- Impact: clients can believe update/delete succeeded while data never changes; later implementation can break API expectations because behavior is already inconsistent.
- Fix approach: implement end-to-end update/delete paths in `biz` and `data`, return explicit unsupported errors until complete, and add API-level tests for create/get/list/update/delete.

**Backend build depends on generated protobuf files that are not committed:**
- Issue: `blog/api/**` contains only `.proto` files while services import generated Go packages such as `blog/api/debt/v1`; `go test ./...` fails in a clean checkout before code-level tests run.
- Files: `blog/api/debt/v1/debt.proto`, `blog/api/user/v1/user.proto`, `blog/internal/service/debt.go`, `blog/internal/service/user.go`, `blog/internal/server/http.go`, `blog/Makefile`
- Impact: CI/bootstrap is fragile, local verification is blocked, and any contributor must remember to run codegen before basic compilation.
- Fix approach: either commit generated `*.pb.go` / `*_http.pb.go` outputs or enforce generation in bootstrap/CI with a documented pre-test step and validation.

**Generated schema migration runs on every startup with no failure handling:**
- Issue: database initialization calls `AutoMigrate` during `NewDb` and ignores connection/setup failure after logging.
- Files: `blog/internal/data/data.go`
- Impact: service startup can continue with a nil or unusable DB handle path, schema drift is uncontrolled, and production startup performs implicit writes.
- Fix approach: fail fast on DB open/migration errors and move schema evolution to explicit migration tooling.

## Known Bugs

**Router auth state becomes stale after login/logout:**
- Symptoms: route guard computes `isAuthenticated` once at module load from `localStorage` and never recomputes.
- Files: `price_recorder_vue/src/router/index.ts`, `price_recorder_vue/src/stores/userStore.ts`, `price_recorder_vue/src/view/Login.vue`
- Trigger: change login state after app startup, or clear storage in another tab/session.
- Workaround: refresh the page after login/logout so the module-level boolean is recalculated.

**Frontend login error path can silently resolve as success:**
- Symptoms: `login()` catches request errors, logs them, and resolves `undefined`; callers then treat the promise as success-only flow.
- Files: `price_recorder_vue/src/api/Login.ts`, `price_recorder_vue/src/view/Login.vue`, `price_recorder_vue/src/utils/request.ts`
- Trigger: backend rejects credentials or network request fails.
- Workaround: none in-app; users rely on alert side effects from the axios interceptor.

**User query path can dereference nil results:**
- Symptoms: repository uses `Find(&user)` into a pointer and then unconditionally reads fields from `user`.
- Files: `blog/internal/data/user.go`
- Trigger: login with unknown username or any query returning no rows.
- Workaround: none; behavior depends on GORM result shape and can panic or return zero-value data.

## Security Considerations

**Authentication secrets and trust boundaries are hard-coded:**
- Risk: JWT signing/verification uses embedded secrets in code, making rotation difficult and leaking trust configuration through source access.
- Files: `blog/internal/service/user.go`, `blog/internal/server/http.go`
- Current mitigation: token-based auth middleware exists for non-whitelisted routes.
- Recommendations: load signing secrets from config, centralize token creation/validation, add expiry/issuer/audience claims, and reject missing/invalid claims consistently.

**Sensitive data is exposed and weakly controlled in user APIs:**
- Risk: create/get/list responses include password fields, and update/delete/list user flows show no authorization checks beyond route-level middleware.
- Files: `blog/internal/service/user.go`, `blog/internal/biz/user.go`, `blog/internal/data/user.go`
- Current mitigation: passwords are hashed during create.
- Recommendations: never return password hashes, scope user reads/writes to current principal, and add authorization tests around user endpoints.

**Repository tests contain hard-coded external credentials/tokens and call live services:**
- Risk: test code embeds database DSNs, bearer tokens, local file paths, and HTTP targets.
- Files: `blog/internal/data/geo_test.go`, `blog/internal/data/hzszgx_file_ocr_test.go`, `blog/internal/data/wjln_db_data_test.go`
- Current mitigation: none detected.
- Recommendations: remove secrets from source, replace with fixtures/env-driven integration tests, and guard live tests behind build tags or `testing.Short()`.

## Performance Bottlenecks

**Debt listing/counting paths rely on broad scans with limited indexing clues:**
- Problem: list endpoints filter by `user_id`, optional `LIKE` clauses, and sort by `created_at DESC`; detail listing joins `debts` and `debt_details` then orders by creation time.
- Files: `blog/internal/data/Debt.go`, `blog/internal/data/debtDetail.go`
- Cause: no repository-level evidence of supporting indexes or query-specific optimization.
- Improvement path: add DB indexes for `(user_id, created_at)` and join/filter columns, then benchmark paginated list endpoints with realistic data volumes.

**Request logging is verbose at DB and HTTP layers:**
- Problem: GORM runs with `logger.Info` and frontend/backend code logs request/user payload details directly.
- Files: `blog/internal/data/data.go`, `price_recorder_vue/src/utils/request.ts`, `price_recorder_vue/src/router/index.ts`, `price_recorder_vue/src/view/Login.vue`, `price_recorder_vue/src/view/Article.vue`
- Cause: development logging defaults are left in runtime paths.
- Improvement path: lower DB log level outside local debug, remove token/user payload logging, and standardize structured logs for errors only.

## Fragile Areas

**Auth context extraction assumes claim type safety:**
- Files: `blog/internal/utils/userutil.go`
- Why fragile: `token.(jwtv5.MapClaims)` and `userId.(string)` are unchecked assertions; malformed or differently shaped claims will panic.
- Safe modification: convert with type checks and explicit error returns before adding more claim usage.
- Test coverage: no tests detected for auth middleware or `CurrentUserId`.

**Frontend import resolution mixes TypeScript sources with `.js` imports:**
- Files: `price_recorder_vue/src/view/Login.vue`, `price_recorder_vue/src/view/Article.vue`, `price_recorder_vue/src/api/Login.ts`, `price_recorder_vue/src/stores/userStore.ts`
- Why fragile: Vue files import `@/api/Login.js`, `@/stores/userStore.js`, and `@/router/index.js` even though the source files are `.ts`; behavior depends on tooling resolution and can break editors/build tools.
- Safe modification: normalize imports to extensionless or `.ts` references consistently and verify Vite/Vitest path alias handling.
- Test coverage: no tests cover these imports or login navigation.

**Backend debt domain logic is concentrated in a few large files:**
- Files: `blog/internal/data/Debt.go`, `blog/internal/service/debt.go`, `blog/internal/service/user.go`
- Why fragile: parsing, validation, mapping, and persistence rules are hand-coded per endpoint with duplicated decimal/time conversion logic.
- Safe modification: extract validation/parsing helpers and mapper functions before adding new fields or filters.
- Test coverage: no unit tests detected for these services/usecases.

## Scaling Limits

**Current test strategy does not scale to CI or team onboarding:**
- Current capacity: one frontend example test and three backend data tests tied to external systems.
- Limit: `pnpm test:unit --run` fails without installing dependencies; `go test ./...` fails in a clean checkout because generated API code is missing.
- Scaling path: add deterministic unit tests for service/biz layers, commit or generate protobuf outputs in CI, and isolate optional integration tests behind explicit commands.

## Dependencies at Risk

**Vite is pinned to a beta release for the main frontend build tool:**
- Risk: toolchain behavior may shift unexpectedly and plugin compatibility can lag.
- Impact: frontend build/dev server stability depends on unreleased semantics.
- Migration plan: move `price_recorder_vue/package.json` to a stable Vite release once current plugin set is validated.

## Missing Critical Features

**No production-grade authorization model is visible for multi-user data ownership beyond debt reads/writes:**
- Problem: debt paths scope by current user, but user CRUD and other routes do not show consistent ownership/admin rules.
- Blocks: safe expansion of account management and admin/user separation.

**No explicit logout/session-expiry handling in frontend auth flow:**
- Problem: token storage is persistent but there is no logout action, token expiry handling, or refresh flow.
- Blocks: predictable authentication UX and secure session invalidation.

## Test Coverage Gaps

**Debt and debt-detail service behavior is untested:**
- What's not tested: decimal parsing, paging/filter validation, ownership checks, and stubbed debt-detail update/delete behavior.
- Files: `blog/internal/service/debt.go`, `blog/internal/service/debtdetail.go`, `blog/internal/biz/debt.go`, `blog/internal/biz/debtDetail.go`
- Risk: regressions in request validation and authorization will reach runtime unnoticed.
- Priority: High

**Authentication and user management flows are untested end to end:**
- What's not tested: password hashing/compare, JWT creation/claim parsing, route whitelist behavior, and frontend login/navigation.
- Files: `blog/internal/service/user.go`, `blog/internal/server/http.go`, `blog/internal/utils/userutil.go`, `price_recorder_vue/src/router/index.ts`, `price_recorder_vue/src/view/Login.vue`
- Risk: auth bugs can produce security holes or lock users out without quick detection.
- Priority: High

**Frontend test suite does not exercise current app behavior:**
- What's not tested: router rendering, API error handling, store persistence, and actual pages.
- Files: `price_recorder_vue/src/__tests__/App.spec.ts`, `price_recorder_vue/src/App.vue`, `price_recorder_vue/src/view/Api.vue`, `price_recorder_vue/src/view/Article.vue`
- Risk: the only test still expects template text not present in `App.vue`, so it does not provide meaningful regression protection even after dependencies are installed.
- Priority: Medium

---

*Concerns audit: 2026-03-25*
