# Codebase Structure

**Analysis Date:** 2026-03-26

## Directory Layout

```text
[project-root]/
├── blog/                    # Kratos backend service, protobuf contracts, config, and persistence code
│   ├── api/                 # Versioned protobuf API contracts
│   ├── cmd/blog/            # Backend executable entry point and Wire assembly
│   ├── configs/             # Backend runtime configuration
│   ├── internal/            # Service, business, data, server, model, and utility layers
│   ├── third_party/         # Vendored protobuf dependencies
│   └── openapi.yaml         # Generated/maintained OpenAPI surface for backend APIs
├── price_recorder_vue/      # Vue 3 + Vite frontend application
│   ├── public/              # Static assets
│   └── src/                 # App entry, router, views, API clients, stores, utilities, tests
└── .planning/codebase/      # Architecture/stack/convention mapping docs for future agents
```

## Directory Purposes

**`blog/api/`:**
- Purpose: Store the external service contracts.
- Contains: versioned `.proto` files grouped by domain, such as `blog/api/post/v1/post.proto` and `blog/api/user/v1/user.proto`.
- Key files: `blog/api/post/v1/post.proto`, `blog/api/user/v1/user.proto`, `blog/api/debt/v1/debt.proto`, `blog/api/price/v1/price.proto`

**`blog/cmd/blog/`:**
- Purpose: Own the backend binary entry point.
- Contains: `main.go` bootstrap code and `wire.go` provider assembly.
- Key files: `blog/cmd/blog/main.go`, `blog/cmd/blog/wire.go`

**`blog/configs/`:**
- Purpose: Hold backend runtime config files.
- Contains: bootstrap configuration consumed by `blog/internal/conf`.
- Key files: `blog/configs/config.yaml`

**`blog/internal/server/`:**
- Purpose: Define transport servers and middleware.
- Contains: HTTP/gRPC server constructors and provider set.
- Key files: `blog/internal/server/http.go`, `blog/internal/server/grpc.go`, `blog/internal/server/server.go`

**`blog/internal/service/`:**
- Purpose: Implement API handlers.
- Contains: one service file per domain, plus the service provider set.
- Key files: `blog/internal/service/post.go`, `blog/internal/service/user.go`, `blog/internal/service/price.go`, `blog/internal/service/debt.go`, `blog/internal/service/service.go`

**`blog/internal/biz/`:**
- Purpose: Hold domain models, business rules, repo interfaces, and use cases.
- Contains: domain-specific use case files and layer README guidance.
- Key files: `blog/internal/biz/post.go`, `blog/internal/biz/user.go`, `blog/internal/biz/price.go`, `blog/internal/biz/debt.go`, `blog/internal/biz/debtDetail.go`, `blog/internal/biz/biz.go`

**`blog/internal/data/`:**
- Purpose: Own infrastructure clients and persistence implementations.
- Contains: GORM DB bootstrap, repo implementations, persistence structs, and backend data tests.
- Key files: `blog/internal/data/data.go`, `blog/internal/data/post.go`, `blog/internal/data/user.go`, `blog/internal/data/price.go`, `blog/internal/data/Debt.go`, `blog/internal/data/debtDetail.go`

**`blog/internal/model/`:**
- Purpose: Hold shared persistence models reused across repos.
- Contains: GORM model definitions.
- Key files: `blog/internal/model/user.go`

**`blog/internal/utils/`:**
- Purpose: Hold helpers shared across backend layers.
- Contains: request-context utilities.
- Key files: `blog/internal/utils/userutil.go`

**`blog/third_party/`:**
- Purpose: Vendor protobuf dependency definitions used by code generation.
- Contains: Google API, protobuf, validation, and OpenAPI proto files.
- Key files: `blog/third_party/google/api/annotations.proto`, `blog/third_party/openapi/v3/annotations.proto`, `blog/third_party/validate/validate.proto`

**`price_recorder_vue/src/view/`:**
- Purpose: Hold route-level pages.
- Contains: `.vue` view components rendered directly by Vue Router.
- Key files: `price_recorder_vue/src/view/Login.vue`, `price_recorder_vue/src/view/Article.vue`, `price_recorder_vue/src/view/Api.vue`

**`price_recorder_vue/src/api/`:**
- Purpose: Hold thin request wrappers by domain/use case.
- Contains: functions that import the shared Axios instance and call backend endpoints.
- Key files: `price_recorder_vue/src/api/Login.ts`, `price_recorder_vue/src/api/Article.ts`

**`price_recorder_vue/src/stores/`:**
- Purpose: Hold Pinia stores.
- Contains: browser-backed user session state.
- Key files: `price_recorder_vue/src/stores/userStore.ts`

**`price_recorder_vue/src/utils/`:**
- Purpose: Hold shared frontend utilities.
- Contains: the central Axios instance and interceptors.
- Key files: `price_recorder_vue/src/utils/request.ts`

**`price_recorder_vue/src/router/`:**
- Purpose: Define SPA routes and route guards.
- Contains: Vue Router setup.
- Key files: `price_recorder_vue/src/router/index.ts`

**`price_recorder_vue/src/__tests__/`:**
- Purpose: Hold frontend unit tests.
- Contains: Vitest specs.
- Key files: `price_recorder_vue/src/__tests__/App.spec.ts`

## Key File Locations

**Entry Points:**
- `blog/cmd/blog/main.go`: backend process entry and Kratos app startup.
- `price_recorder_vue/src/main.ts`: frontend app bootstrap.
- `price_recorder_vue/src/router/index.ts`: frontend route table and auth guard.

**Configuration:**
- `blog/configs/config.yaml`: backend runtime settings for server and data infrastructure.
- `blog/internal/conf/conf.proto`: typed config schema consumed by backend bootstrap.
- `price_recorder_vue/package.json`: frontend scripts and package boundaries.

**Core Logic:**
- `blog/internal/service/*.go`: protobuf handler implementations.
- `blog/internal/biz/*.go`: use cases and repo interfaces.
- `blog/internal/data/*.go`: DB access and repo implementations.
- `price_recorder_vue/src/api/*.ts`: frontend endpoint wrappers.
- `price_recorder_vue/src/view/*.vue`: frontend route-level UI.

**Testing:**
- `blog/internal/data/*_test.go`: backend Go tests currently live beside data-layer code.
- `price_recorder_vue/src/__tests__/*.spec.ts`: frontend unit tests.

## Naming Conventions

**Files:**
- Backend domain files use lower-case Go filenames by layer, such as `blog/internal/service/post.go` and `blog/internal/biz/user.go`.
- Backend protobuf files use versioned paths and domain names, such as `blog/api/post/v1/post.proto`.
- Frontend views use PascalCase component filenames, such as `price_recorder_vue/src/view/Login.vue` and `price_recorder_vue/src/view/Article.vue`.
- Frontend API and store modules use descriptive TypeScript filenames, such as `price_recorder_vue/src/api/Login.ts` and `price_recorder_vue/src/stores/userStore.ts`.

**Directories:**
- Backend directories map directly to architectural layers: `server`, `service`, `biz`, `data`, `model`, `utils`.
- Frontend directories group by responsibility: `view`, `api`, `router`, `stores`, `utils`, `__tests__`.

## Where to Add New Code

**New backend API feature:**
- Contract first: add or update the versioned protobuf file under `blog/api/<domain>/v1/`.
- Transport/service implementation: add handler logic in `blog/internal/service/<domain>.go`.
- Business logic: add use case methods and repo interface changes in `blog/internal/biz/<domain>.go`.
- Persistence: add repo implementation changes and DB structs in `blog/internal/data/<domain>.go` or `blog/internal/model/` if the model is shared.
- Wiring: update provider sets only if a new service/use case/repo constructor is introduced in `blog/internal/service/service.go`, `blog/internal/biz/biz.go`, or `blog/internal/data/data.go`.

**New frontend page or flow:**
- Route-level UI: add the page under `price_recorder_vue/src/view/`.
- Route registration: register it in `price_recorder_vue/src/router/index.ts`.
- Backend calls: add a thin wrapper in `price_recorder_vue/src/api/`.
- Shared transport logic: touch `price_recorder_vue/src/utils/request.ts` only for behavior needed across multiple API modules.
- Session/shared state: use `price_recorder_vue/src/stores/` when data must persist across views.

**New utilities:**
- Backend helpers shared across domains belong in `blog/internal/utils/`.
- Frontend helpers shared across pages belong in `price_recorder_vue/src/utils/`.

**Tests:**
- Backend tests should stay adjacent to the layer under test, following the existing `*_test.go` pattern in `blog/internal/data/`.
- Frontend unit tests should go under `price_recorder_vue/src/__tests__/` or next to modules only if the existing pattern changes intentionally.

## Special Directories

**`blog/third_party/`:**
- Purpose: Store vendored protobuf definitions required for generation.
- Generated: No.
- Committed: Yes.

**`blog/internal/conf/`:**
- Purpose: Store generated config bindings and their source proto schema.
- Generated: `blog/internal/conf/conf.pb.go` is generated; `blog/internal/conf/conf.proto` is source.
- Committed: Yes.

**`blog/bin/`:**
- Purpose: Build output target referenced by backend `make build`.
- Generated: Yes.
- Committed: Not detected in the current tree.

**`.planning/codebase/`:**
- Purpose: Store machine-readable architecture and quality mapping docs for future planning/execution agents.
- Generated: Yes, by mapper workflows.
- Committed: Yes, when the mapping is intentionally updated.

---

*Structure analysis: 2026-03-26*
