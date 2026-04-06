# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Structure

This is a monorepo with two projects:
- `blog/` - Go backend service using Kratos framework
- `price_recorder_vue/` - Vue 3 + Vite frontend SPA

The backend follows a strict Kratos layering pattern: `server` (transport) → `service` → `biz` (business logic) → `data` (persistence). Dependencies are wired using Google Wire from `blog/cmd/blog/wire.go`.

## Backend Development (blog/)

### Common Commands
```bash
cd blog

# Setup: install proto/codegen tools
make init

# Generate code from proto files
make api          # Generate API protobuf (includes OpenAPI to openapi.yaml)
make config       # Generate internal config protobuf
make all          # Run api + config + generate

# Run and build
make run          # Run Kratos service (reads from configs/config.yaml)
make build        # Build binary to blog/bin/
go test ./...     # Run Go tests

# Dependency injection
cd cmd/blog && wire  # Generate Wire assembly code
```

### Architecture Layers

**Transport (`blog/internal/server/`)**: HTTP/gRPC server setup, middleware, handler registration. Auth middleware is configured with a whitelist for public routes.

**Service (`blog/internal/service/*.go`)**: Implements protobuf handlers, maps requests/responses, calls use cases. One service file per domain (post.go, user.go, debt.go, price.go, etc.).

**Business (`blog/internal/biz/*.go`)**: Domain models, use cases, repository interfaces. Use cases contain business rules and call repo interfaces.

**Data (`blog/internal/data/*.go`)**: Repository implementations, GORM DB bootstrap, persistence models. Also contains backend tests in `*_test.go` files.

**Model (`blog/internal/model/`)**: Shared GORM models used across repos (e.g., User).

### Adding Backend Features

1. Define or update the protobuf contract in `blog/api/<domain>/v1/<domain>.proto`
2. Run `make api` to generate API bindings and update `blog/openapi.yaml`
3. Implement service handlers in `blog/internal/service/<domain>.go`
4. Add use cases and repo interfaces in `blog/internal/biz/<domain>.go`
5. Implement repo in `blog/internal/data/<domain>.go` or `blog/internal/model/` for shared models
6. Update Wire providers in `blog/internal/{data,biz,service}/*.go` if adding new constructors

### Function Reuse Rule

Before adding a new backend function in `blog/internal`, check `blog/FUNCTION_INDEX.md` for existing reusable functions. When adding/removing/changing functions, update the index in the same change. PR descriptions should confirm this was done.

## Frontend Development (price_recorder_vue/)

### Common Commands
```bash
cd price_recorder_vue

# Setup
pnpm install

# Development
pnpm dev          # Start Vite dev server

# Build and test
pnpm build        # Type-check and build production assets
pnpm test:unit    # Run Vitest unit tests
pnpm lint         # Run ESLint with auto-fix
pnpm format       # Run Prettier
```

### Architecture

**Entry point**: `src/main.ts` mounts Vue app with Pinia and Vue Router.

**Router (`src/router/index.ts`)**: Route definitions and auth guards. Unauthenticated users are redirected to `/login`.

**Views (`src/view/*.vue`)**: Route-level pages. Components use PascalCase filenames (e.g., Article.vue, Login.vue).

**API clients (`src/api/*.ts`)**: Thin wrappers that import the shared Axios instance and call backend endpoints.

**Shared transport (`src/utils/request.ts`)**: Central Axios instance with request/response interceptors. Adds `Authorization: Bearer <token>` header from localStorage.

**Stores (`src/stores/*.ts`)**: Pinia stores for persistent state (e.g., userStore.ts mirrors localStorage for auth).

### Adding Frontend Features

1. Add route-level UI in `src/view/<FeatureName>.vue`
2. Register route in `src/router/index.ts`
3. Add API wrapper in `src/api/<FeatureName>.ts` if needed
4. Use `src/stores/` for state shared across views
5. Modify `src/utils/request.ts` only for cross-module transport changes

## Code Conventions

**Backend**: Go files use lowercase names (post.go, user.go). Proto files use versioned paths (api/post/v1/post.proto). Format with `go fmt ./...` before committing.

**Frontend**: Vue components use PascalCase filenames. Use ESLint + Prettier configuration in `eslint.config.ts` and `.prettierrc.json`.

**API contracts**: Maintain protobuf files under `blog/api/`. Run `make api` after changes to regenerate bindings and `blog/openapi.yaml`.

## Cross-Cutting Concerns

**Authentication**: Backend JWT middleware in `blog/internal/server/http.go` with route whitelisting. Frontend stores token in localStorage, attaches via Axios interceptor in `src/utils/request.ts`. User identity extracted via `blog/internal/utils/userutil.go`.

**Database**: GORM auto-migration enabled in `blog/internal/data/data.go`. Uses PostgreSQL driver. Decimal fields use `github.com/shopspring/decimal`.

**Configuration**: Backend reads from `blog/configs/config.yaml` using Kratos file source. Schema defined in `blog/internal/conf/conf.proto`.

## Testing

- Backend tests: Place in `blog/internal/**/*_test.go` adjacent to implementation. Run with `go test ./...`.
- Frontend tests: Place in `src/__tests__/*.spec.ts`. Run with `pnpm test:unit`.

## Important References

- `blog/FUNCTION_INDEX.md` - Backend function index for reuse
- `blog/openapi.yaml` - Generated OpenAPI spec for HTTP endpoints
- `AGENTS.md` - Detailed repository specifications (Chinese)
- `.planning/codebase/` - Architecture and structure documentation
