# Architecture

**Analysis Date:** 2026-03-26

## Pattern Overview

**Overall:** Polyrepo-style application with a Kratos backend in `blog/` and a Vue 3 SPA frontend in `price_recorder_vue/`, connected over HTTP JSON endpoints generated from protobuf contracts.

**Key Characteristics:**
- Backend follows a strict Kratos layering pattern: transport in `blog/internal/server`, RPC handlers in `blog/internal/service`, use cases in `blog/internal/biz`, persistence in `blog/internal/data`.
- Backend dependencies are assembled with Google Wire from `blog/cmd/blog/wire.go`, so outer layers depend inward and concrete implementations are injected at startup.
- Frontend is a client-only Vue app with route-level views in `price_recorder_vue/src/view`, thin API wrappers in `price_recorder_vue/src/api`, shared Axios transport in `price_recorder_vue/src/utils/request.ts`, and lightweight Pinia state in `price_recorder_vue/src/stores/userStore.ts`.

## Backend and Frontend Boundary

**Backend API surface:**
- Contracts live under `blog/api/**`, including `blog/api/post/v1/post.proto`, `blog/api/user/v1/user.proto`, `blog/api/price/v1/price.proto`, `blog/api/debt/v1/debt.proto`, and `blog/api/ocr/v1/aiocr.proto`.
- HTTP transport registers generated handlers in `blog/internal/server/http.go`.
- gRPC transport is present in `blog/internal/server/grpc.go`, but currently only registers the Greeter service.

**Frontend boundary:**
- The SPA mounts from `price_recorder_vue/src/main.ts` and renders routes through `price_recorder_vue/src/App.vue`.
- API calls use relative `/api` paths from `price_recorder_vue/src/utils/request.ts`, so the frontend expects a reverse proxy or dev proxy in front of the Kratos HTTP server.
- Authentication state is stored in browser `localStorage` and propagated into request headers by the Axios request interceptor in `price_recorder_vue/src/utils/request.ts`.

## Layers

**Application Bootstrap:**
- Purpose: Load config, build providers, start transports.
- Location: `blog/cmd/blog/main.go`, `blog/cmd/blog/wire.go`
- Contains: config loading, logger setup, `kratos.App` creation, Wire assembly.
- Depends on: `blog/internal/conf`, `blog/internal/server`, `blog/internal/service`, `blog/internal/biz`, `blog/internal/data`.
- Used by: backend process startup.

**Transport Layer:**
- Purpose: Expose HTTP/gRPC endpoints and apply middleware.
- Location: `blog/internal/server/http.go`, `blog/internal/server/grpc.go`, `blog/internal/server/server.go`
- Contains: server constructors, middleware chains, handler registration, whitelist-based JWT bypass.
- Depends on: generated API packages under `blog/api/**`, service implementations under `blog/internal/service`, config under `blog/internal/conf`.
- Used by: `kratos.App` in `blog/cmd/blog/main.go`.

**Service Layer:**
- Purpose: Translate protobuf requests into use-case calls and map use-case results back into protobuf replies.
- Location: `blog/internal/service/*.go`
- Contains: one service per API domain, such as `blog/internal/service/post.go`, `blog/internal/service/user.go`, `blog/internal/service/price.go`, `blog/internal/service/debt.go`.
- Depends on: use cases in `blog/internal/biz`, protobuf request/response types in `blog/api/**`.
- Used by: transport registration in `blog/internal/server/http.go` and `blog/internal/server/grpc.go`.

**Business Layer:**
- Purpose: Hold domain models, repository interfaces, and business rules.
- Location: `blog/internal/biz/*.go`
- Contains: domain entities like `Post` in `blog/internal/biz/post.go`, `User` in `blog/internal/biz/user.go`, use cases, and repo interfaces.
- Depends on: logging and utility helpers such as `blog/internal/utils/userutil.go`.
- Used by: services in `blog/internal/service` and repositories implemented in `blog/internal/data`.

**Data Layer:**
- Purpose: Create infrastructure clients and implement repository interfaces.
- Location: `blog/internal/data/*.go`
- Contains: GORM database bootstrap in `blog/internal/data/data.go`, repo implementations like `blog/internal/data/post.go` and `blog/internal/data/user.go`, persistence models such as `blog/internal/data/Debt.go`.
- Depends on: `gorm`, PostgreSQL driver, config from `blog/internal/conf`, shared models from `blog/internal/model`.
- Used by: use cases through interfaces declared in `blog/internal/biz`.

**Shared Model Layer:**
- Purpose: Hold reusable persistence models not owned by a single repo implementation.
- Location: `blog/internal/model/user.go`
- Contains: the shared `User` GORM model.
- Depends on: `gorm`.
- Used by: `blog/internal/data/user.go` and DB auto-migration in `blog/internal/data/data.go`.

**Frontend Presentation Layer:**
- Purpose: Render pages and trigger API calls.
- Location: `price_recorder_vue/src/view/*.vue`, `price_recorder_vue/src/router/index.ts`
- Contains: route views such as `price_recorder_vue/src/view/Login.vue`, `price_recorder_vue/src/view/Article.vue`, `price_recorder_vue/src/view/Api.vue`.
- Depends on: API wrappers in `price_recorder_vue/src/api`, Pinia stores in `price_recorder_vue/src/stores`, Vue Router.
- Used by: `price_recorder_vue/src/App.vue`.

**Frontend Client/State Layer:**
- Purpose: Centralize HTTP behavior and minimal user session persistence.
- Location: `price_recorder_vue/src/utils/request.ts`, `price_recorder_vue/src/api/*.ts`, `price_recorder_vue/src/stores/userStore.ts`
- Contains: Axios instance, request/response interceptors, per-domain API calls, browser-backed session state.
- Depends on: `axios`, `pinia`, browser `localStorage`.
- Used by: route views in `price_recorder_vue/src/view`.

## Data Flow

**Backend HTTP request flow:**

1. `blog/cmd/blog/main.go` loads bootstrap config from `blog/configs/config.yaml`, builds providers with `wireApp`, and starts HTTP and gRPC servers.
2. `blog/internal/server/http.go` receives the request, runs recovery/tracing/logging/JWT middleware, and dispatches to the generated handler for a service implementation.
3. A service such as `blog/internal/service/post.go` maps protobuf fields into a domain input for a use case like `blog/internal/biz/post.go`.
4. The use case invokes a repo interface implemented in `blog/internal/data/post.go` or `blog/internal/data/user.go`.
5. The data layer uses the shared `*gorm.DB` from `blog/internal/data/data.go`, persists or queries PostgreSQL, and returns domain models upward.
6. The service maps the result back into a protobuf reply, and Kratos writes the HTTP JSON response.

**Frontend login flow:**

1. `price_recorder_vue/src/router/index.ts` checks `localStorage` before navigation and redirects unauthenticated users to `/login`.
2. `price_recorder_vue/src/view/Login.vue` calls `login()` from `price_recorder_vue/src/api/Login.ts`.
3. `price_recorder_vue/src/api/Login.ts` posts through the shared Axios instance in `price_recorder_vue/src/utils/request.ts`.
4. `price_recorder_vue/src/stores/userStore.ts` watches the store state and writes the login payload into `localStorage`.
5. Subsequent requests attach `Authorization: Bearer <token>` in `price_recorder_vue/src/utils/request.ts`.

**Frontend article list flow:**

1. `price_recorder_vue/src/view/Article.vue` fetches data in `onMounted`.
2. `price_recorder_vue/src/api/Article.ts` calls `GET /post/page/v1` through the shared Axios instance.
3. The backend `PostService` in `blog/internal/service/post.go` calls `PostUsecase` in `blog/internal/biz/post.go`.
4. `blog/internal/data/post.go` queries the database with GORM paging and returns post rows to the frontend.

**State Management:**
- Backend request state is mostly per-request via `context.Context`; user identity is extracted from context in `blog/internal/utils/userutil.go`.
- Backend long-lived state is the shared GORM connection in `blog/internal/data.Data`.
- Frontend state is intentionally thin: route state lives in Vue Router, persistent auth state lives in `localStorage`, and `price_recorder_vue/src/stores/userStore.ts` mirrors that browser state.

## Key Abstractions

**Proto-first service contract:**
- Purpose: Define transport shape independently from implementation.
- Examples: `blog/api/post/v1/post.proto`, `blog/api/user/v1/user.proto`, `blog/openapi.yaml`
- Pattern: contracts are versioned under `api/<domain>/v1`, then registered by generated Kratos handlers.

**Use case + repo pair:**
- Purpose: Keep business rules and persistence separate.
- Examples: `blog/internal/biz/post.go` with `blog/internal/data/post.go`, `blog/internal/biz/user.go` with `blog/internal/data/user.go`
- Pattern: `biz` owns interfaces and domain structs; `data` owns GORM-backed implementations.

**Shared HTTP client wrapper:**
- Purpose: Make frontend requests consistent around auth and error handling.
- Examples: `price_recorder_vue/src/utils/request.ts`, `price_recorder_vue/src/api/Article.ts`, `price_recorder_vue/src/api/Login.ts`
- Pattern: one Axios instance is imported into per-domain API modules.

## Entry Points

**Backend process entry:**
- Location: `blog/cmd/blog/main.go`
- Triggers: `make run`, `make build`, direct `go run`.
- Responsibilities: parse `-conf`, load bootstrap config, create logger, assemble the app, run transports.

**Dependency injection assembly:**
- Location: `blog/cmd/blog/wire.go`
- Triggers: Wire code generation and backend startup.
- Responsibilities: combine provider sets from `server`, `data`, `biz`, and `service`.

**Frontend application entry:**
- Location: `price_recorder_vue/src/main.ts`
- Triggers: Vite dev server or production bundle load.
- Responsibilities: create Vue app, install Pinia and Vue Router, mount `App.vue`.

**Frontend router entry:**
- Location: `price_recorder_vue/src/router/index.ts`
- Triggers: client-side navigation.
- Responsibilities: define route table and enforce login redirect behavior.

## Error Handling

**Strategy:** Middleware-based recovery on the backend, ad hoc error propagation in service/data methods, and Axios interceptor handling on the frontend.

**Patterns:**
- Backend transport uses `recovery.Recovery()` in both `blog/internal/server/http.go` and `blog/internal/server/grpc.go`.
- Backend services often return raw errors from the use case/data stack, as in `blog/internal/service/user.go`.
- Frontend responses are normalized in `price_recorder_vue/src/utils/request.ts`; non-`200` payload codes and HTTP errors are rejected and surfaced with `alert`.

## Cross-Cutting Concerns

**Logging:** Backend logging is configured in `blog/cmd/blog/main.go` and passed through Kratos logging middleware in `blog/internal/server/http.go`. Use cases also create `log.Helper` instances in files like `blog/internal/biz/post.go` and `blog/internal/biz/user.go`.

**Validation:** API contracts and field shape are primarily expressed through protobuf files under `blog/api/**`. There is no separate frontend form validation layer; current checks happen inline in view logic.

**Authentication:** HTTP auth is token-based. `blog/internal/server/http.go` attaches JWT middleware and skips auth for whitelisted routes. `blog/internal/service/user.go` issues JWTs during login. The frontend stores the returned token and injects it into the `Authorization` header via `price_recorder_vue/src/utils/request.ts`.

---

*Architecture analysis: 2026-03-26*
