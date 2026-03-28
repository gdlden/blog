# Conventions

**Analysis Date:** 2026-03-26

## Repository-Level Rules

**Monorepo split:**
- `blog/` contains the Go backend service.
- `price_recorder_vue/` contains the Vue 3 frontend app.
- Shared runtime behavior happens over HTTP API contracts rather than shared source modules.

**Primary guidance source:**
- Repository instructions are defined in `AGENTS.md` at the repo root.
- Backend reuse policy is additionally documented in `blog/FUNCTION_INDEX.md`.

## Backend Conventions

**Layering:**
- Code follows a Kratos-style inward dependency flow: `server -> service -> biz -> data`.
- Transport registration lives in `blog/internal/server`.
- Request-to-domain mapping lives in `blog/internal/service`.
- Domain logic and repository interfaces live in `blog/internal/biz`.
- GORM-backed repository implementations live in `blog/internal/data`.

**Naming:**
- Package names are lowercase, matching Go conventions, such as `biz`, `data`, `server`, `service`, and `utils`.
- API contracts use versioned paths such as `blog/api/post/v1/post.proto`.
- Constructors generally use `NewXxx` naming, such as `NewUserService`, `NewPostUsecase`, and `NewDebtRepo`.

**Formatting and style:**
- Go code is expected to be formatted with `gofmt` or `go fmt ./...`.
- Files use exported camel-case names for public symbols and short receiver names for methods.
- Error returns are propagated directly in most service and data methods.

**Function reuse policy:**
- Before adding a backend function, contributors are expected to search existing code and `blog/FUNCTION_INDEX.md` for reusable implementations.
- Any backend function add/remove/signature change/behavior change must update `blog/FUNCTION_INDEX.md` in the same change.
- PRs are expected to explicitly confirm that reuse was checked and `blog/FUNCTION_INDEX.md` was updated.

**Dependency injection and providers:**
- Provider aggregation follows `ProviderSet` patterns in `blog/internal/{server,service,biz,data}`.
- Application assembly is generated through Wire from `blog/cmd/blog/wire.go`.

## API and Schema Conventions

**Proto-first contracts:**
- API definitions live under `blog/api/**`.
- Versioning is part of the path and package naming, for example `v1`.
- Backend service implementations import generated Go code from those proto packages.

**Generated artifacts:**
- Proto and config schemas are sources of truth.
- `make api && make config` is the documented regeneration path.
- `blog/third_party/` vendors protobuf dependencies used by code generation.

**Current caveat:**
- The codebase expects generated Go artifacts for API packages, but they are not fully present in the checkout. That is a convention in practice, even though the generated outputs are not committed.

## Frontend Conventions

**Framework structure:**
- Vue single-file components live in `price_recorder_vue/src/view`.
- API wrappers live in `price_recorder_vue/src/api`.
- Shared client transport lives in `price_recorder_vue/src/utils/request.ts`.
- State management lives in `price_recorder_vue/src/stores`.
- Router setup lives in `price_recorder_vue/src/router/index.ts`.

**Formatting and style:**
- Frontend code is expected to follow ESLint + Prettier from `price_recorder_vue/eslint.config.ts`.
- Repository instructions specify 2-space indentation and PascalCase component file names, such as `Article.vue`.
- The checked-in frontend code is partially inconsistent with that target style, especially around spacing and semicolon usage in files like `price_recorder_vue/src/router/index.ts` and `price_recorder_vue/src/utils/request.ts`.

**Imports and aliases:**
- The app uses `@` as an alias for `price_recorder_vue/src` from `price_recorder_vue/vite.config.ts`.
- Imports commonly reference source modules directly, for example `@/api/Login.ts`.
- Some Vue files appear to import `.js` paths for TypeScript sources, which is a fragile but recurring local pattern.

**State and persistence:**
- Session state is persisted to `localStorage`.
- The shared Axios instance injects bearer tokens from browser storage into outgoing requests.

## Testing Conventions

**Backend:**
- Tests live adjacent to Go implementation code and use `*_test.go` naming.
- Existing backend tests are concentrated under `blog/internal/data`.

**Frontend:**
- Unit tests live under `price_recorder_vue/src/__tests__`.
- Vitest is configured with a `jsdom` environment in `price_recorder_vue/vitest.config.ts`.

## Documentation and Operational Conventions

**Build commands:**
- Backend uses `make run`, `make build`, `make api`, and `make config`.
- Frontend uses `pnpm dev`, `pnpm build`, `pnpm test:unit`, `pnpm lint`, and `pnpm format`.

**Config handling:**
- Backend runtime config is file-based and loaded from `blog/configs/config.yaml`.
- Auth secrets and some external endpoints are currently embedded in code/config rather than centralized in environment variables.

## Consistency Assessment

**Conventions that are followed consistently:**
- Backend directory layering is clear and easy to trace.
- Proto path versioning is applied across API domains.
- Frontend tooling config exists for linting, formatting, type-checking, and tests.

**Conventions that are only partially enforced:**
- Frontend style consistency varies by file.
- Generated-code expectations are implicit rather than enforced by CI/bootstrap.
- Security-sensitive configuration is not consistently externalized.
- Test placement conventions exist, but test coverage is sparse.

---

*Convention analysis: 2026-03-26*
