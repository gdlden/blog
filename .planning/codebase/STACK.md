# Technology Stack

**Analysis Date:** 2026-03-26

## Languages

**Primary:**
- Go 1.24.0 - backend service in `blog/`
- TypeScript 5.9.x - frontend app in `price_recorder_vue/src/`

**Secondary:**
- Vue SFC templates/styles - UI components in `price_recorder_vue/src/**/*.vue`
- Protocol Buffers - API contracts in `blog/api/**` and config schema in `blog/internal/conf/conf.proto`
- YAML - backend runtime config in `blog/configs/config.yaml` and generated HTTP contract in `blog/openapi.yaml`

## Runtime

**Environment:**
- Go runtime/toolchain 1.24.0 for `blog/` from `blog/go.mod`
- Node.js `^20.19.0 || >=22.12.0` for `price_recorder_vue/` from `price_recorder_vue/package.json`

**Package Manager:**
- Go modules - lockfile present at `blog/go.sum`
- pnpm - lockfile present at `price_recorder_vue/pnpm-lock.yaml`

## Frameworks

**Core:**
- Kratos v2.8.0 - backend application/runtime, HTTP and gRPC transport in `blog/cmd/blog/main.go`, `blog/internal/server/http.go`, `blog/internal/server/grpc.go`
- GORM v1.25.12 - ORM and schema migration in `blog/internal/data/data.go`
- Vue 3.5.x - SPA frontend entry in `price_recorder_vue/src/main.ts`
- Vue Router 4.6.x - client routing in `price_recorder_vue/src/router/index.ts`
- Pinia 3.0.x - client state store in `price_recorder_vue/src/stores/userStore.ts`

**Testing:**
- Go `testing` package - backend tests in `blog/internal/data/*_test.go`
- Vitest 4.0.x with jsdom - frontend unit tests in `price_recorder_vue/vitest.config.ts` and `price_recorder_vue/src/__tests__/App.spec.ts`
- Vue Test Utils 2.4.x - Vue component test support from `price_recorder_vue/package.json`

**Build/Dev:**
- Wire 0.7.0 - dependency injection wiring in `blog/cmd/blog/wire.go`
- protoc + Kratos/OpenAPI generators - API/config codegen in `blog/Makefile`
- Vite 6 beta - frontend dev/build server in `price_recorder_vue/vite.config.ts`
- Tailwind CSS 4.1.x via Vite plugin - styling pipeline in `price_recorder_vue/package.json` and `price_recorder_vue/vite.config.ts`
- ESLint 9 + Prettier 3 - frontend lint/format tooling in `price_recorder_vue/eslint.config.ts`
- vue-tsc - frontend type checking in `price_recorder_vue/package.json`
- Docker - backend container build/runtime in `blog/Dockerfile`

## Key Dependencies

**Critical:**
- `github.com/go-kratos/kratos/v2` v2.8.0 - service lifecycle, config loading, middleware, HTTP/gRPC transport in `blog/cmd/blog/main.go`
- `gorm.io/gorm` v1.25.12 - persistence and auto-migration in `blog/internal/data/data.go`
- `gorm.io/driver/postgres` v1.5.11 - active database driver in `blog/internal/data/data.go`
- `github.com/golang-jwt/jwt/v5` v5.1.0 - token signing and auth middleware in `blog/internal/service/user.go` and `blog/internal/server/http.go`
- `axios` v1.13.x - frontend HTTP client in `price_recorder_vue/src/utils/request.ts`

**Infrastructure:**
- `github.com/google/wire` v0.7.0 - provider graph assembly in `blog/internal/{data,service,server}/*.go`
- `github.com/google/uuid` v1.6.0 - backend ID generation in `blog/internal/service/user.go`
- `golang.org/x/crypto` v0.46.0 - password hashing support via bcrypt in `blog/internal/service/user.go`
- `github.com/shopspring/decimal` v1.4.0 - money/decimal fields in `blog/internal/data/{Debt.go,debtDetail.go}`
- `@vitejs/plugin-vue` 6.0.x - Vue SFC build integration in `price_recorder_vue/vite.config.ts`
- `vite-plugin-vue-devtools` 8.0.x - Vue devtools plugin in `price_recorder_vue/vite.config.ts`

## Configuration

**Environment:**
- Backend reads file-based config from `blog/configs/` using Kratos file source in `blog/cmd/blog/main.go`
- Config schema defines `server.http`, `server.grpc`, `data.database`, and `data.redis` in `blog/internal/conf/conf.proto`
- No `.env`-based runtime wiring is detected in application code

**Build:**
- Backend build/codegen config lives in `blog/Makefile`, `blog/Dockerfile`, and `blog/cmd/blog/wire.go`
- Frontend build/lint/test config lives in `price_recorder_vue/{vite.config.ts,vitest.config.ts,eslint.config.ts,tsconfig.json}`
- HTTP contract is generated to `blog/openapi.yaml` from `blog/api/**`

## Platform Requirements

**Development:**
- Go toolchain compatible with `go 1.24.0`
- Node.js compatible with `^20.19.0 || >=22.12.0`
- `pnpm` for frontend dependency management
- `protoc` plus Kratos/OpenAPI generator binaries for backend code generation from `blog/Makefile`

**Production:**
- Backend targets a Linux container image built from `blog/Dockerfile`
- Container exposes HTTP `8000` and gRPC `9000` per `blog/Dockerfile`
- Frontend builds to static assets through Vite in `price_recorder_vue/package.json`

---

*Stack analysis: 2026-03-26*
