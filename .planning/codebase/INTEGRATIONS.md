# External Integrations

**Analysis Date:** 2026-03-26

## APIs & External Services

**Internal application APIs:**
- Blog backend HTTP API - frontend consumes backend routes through Axios with `/api` base URL in `price_recorder_vue/src/utils/request.ts`
  - SDK/Client: `axios`
  - Auth: `Authorization: Bearer <token>` header set in `price_recorder_vue/src/utils/request.ts`
- Blog backend gRPC API - backend exposes gRPC transport for registered services in `blog/internal/server/grpc.go`
  - SDK/Client: Kratos gRPC transport
  - Auth: Not detected on active gRPC registrations

**AI/OCR:**
- Volcengine Ark OpenAI-compatible chat completion API - OCR endpoint forwards image+text prompt to `https://ark.cn-beijing.volces.com/api/v3` in `blog/internal/service/aiocr.go`
  - SDK/Client: `github.com/sashabaranov/go-openai`
  - Auth: API key literal passed to `ark.DefaultConfig(...)` in `blog/internal/service/aiocr.go`

## Data Storage

**Databases:**
- PostgreSQL
  - Connection: file-configured `data.database.source` from `blog/internal/conf/conf.proto`
  - Client: `gorm.io/gorm` with `gorm.io/driver/postgres` in `blog/internal/data/data.go`
  - Usage: auto-migrates `Debt`, `DebtDetail`, `Post`, `Price`, and `User` models in `blog/internal/data/data.go`

**File Storage:**
- Not active in production code
  - File API contract exists in `blog/api/file/v1/file.proto`
  - Stub service exists in `blog/internal/service/file.go`
  - Service is not in `blog/internal/service/service.go` and is not registered in `blog/internal/server/http.go`

**Caching:**
- Redis schema is defined in `blog/internal/conf/conf.proto`
- No Redis client or cache integration is instantiated in `blog/internal/data/` or `blog/internal/service/`

## Authentication & Identity

**Auth Provider:**
- Custom JWT auth
  - Implementation: login signs HS256 JWT in `blog/internal/service/user.go`; HTTP middleware validates JWT for non-whitelisted routes in `blog/internal/server/http.go`
  - Secret handling: signing/verification keys are hardcoded in `blog/internal/service/user.go` and `blog/internal/server/http.go`
- Password verification uses bcrypt in `blog/internal/service/user.go`
- Frontend persists user session in `localStorage` and injects bearer tokens from `price_recorder_vue/src/utils/request.ts` and `price_recorder_vue/src/router/index.ts`

## Monitoring & Observability

**Error Tracking:**
- None detected

**Logs:**
- Kratos structured logger writes stdout logs with trace/span metadata in `blog/cmd/blog/main.go`
- GORM SQL logging is enabled at `logger.Info` in `blog/internal/data/data.go`
- Frontend uses browser `console.log`/`console.error` in `price_recorder_vue/src/{router/index.ts,utils/request.ts,view/Login.vue}`

## CI/CD & Deployment

**Hosting:**
- Backend container image defined in `blog/Dockerfile`
- Frontend hosting target is not explicitly defined; Vite build output is the deployable artifact from `price_recorder_vue/package.json`

**CI Pipeline:**
- Not detected

## Environment Configuration

**Required env vars:**
- None detected as required by application code

**Secrets location:**
- Backend runtime config file at `blog/configs/config.yaml`
- Hardcoded application secrets exist in `blog/internal/service/aiocr.go`, `blog/internal/service/user.go`, and `blog/internal/server/http.go`
- `.env` files are not detected from the repository files examined

## Webhooks & Callbacks

**Incoming:**
- None detected

**Outgoing:**
- OCR requests to Volcengine Ark from `blog/internal/service/aiocr.go`

## Service Boundaries

**Frontend to backend boundary:**
- Vite dev server proxies `/api/*` to `http://localhost:8000` in `price_recorder_vue/vite.config.ts`
- Frontend currently calls `/user/login/v1` in `price_recorder_vue/src/api/Login.ts` and `/post/page/v1` in `price_recorder_vue/src/api/Article.ts`

**Backend transport boundary:**
- HTTP server registers `Greeter`, `Post`, `User`, `Aiocr`, `Price`, `Debt`, and `DebtDetail` services in `blog/internal/server/http.go`
- gRPC server currently registers only `Greeter` in `blog/internal/server/grpc.go`
- API contracts live under `blog/api/**`; generated OpenAPI output is `blog/openapi.yaml`

**Unwired boundary:**
- File upload contract/service is defined in `blog/api/file/v1/file.proto` and `blog/internal/service/file.go` but is not wired into provider sets or transport registration

---

*Integration audit: 2026-03-26*
